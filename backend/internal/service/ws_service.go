package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// WSMessage is the WebSocket protocol envelope.
type WSMessage struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp int64           `json:"ts"`
}

// ChatMessageData is the payload for chat.message.send events.
type ChatMessageData struct {
	RoomID      uuid.UUID `json:"room_id"`
	SenderID    uuid.UUID `json:"sender_id"`
	ClientMsgID string    `json:"client_msg_id"`
	ContentType int8      `json:"content_type"`
	Content     string    `json:"content"`
}

type WSClient interface {
	SendJSON(v interface{}) error
	UserID() uuid.UUID
}

type WSHub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[WSClient]bool // userID -> set of connections
	redis   *redis.Client
}

func NewWSHub(redis *redis.Client) *WSHub {
	return &WSHub{
		clients: make(map[uuid.UUID]map[WSClient]bool),
		redis:   redis,
	}
}

func (h *WSHub) Register(client WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	uid := client.UserID()
	if h.clients[uid] == nil {
		h.clients[uid] = make(map[WSClient]bool)
	}
	h.clients[uid][client] = true
	log.Printf("[WS] user %s connected (%d connections)", uid, len(h.clients[uid]))
}

func (h *WSHub) Unregister(client WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	uid := client.UserID()
	if clients, ok := h.clients[uid]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, uid)
		}
	}
	log.Printf("[WS] user %s disconnected", uid)
}

func (h *WSHub) SendToUser(userID uuid.UUID, msg *WSMessage) error {
	h.mu.RLock()
	clients := h.clients[userID]
	h.mu.RUnlock()

	if len(clients) == 0 {
		// User offline, queue message to Redis
		return h.queueOffline(userID, msg)
	}

	var lastErr error
	for c := range clients {
		if err := c.SendJSON(msg); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (h *WSHub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

func (h *WSHub) BroadcastToRoom(roomID uuid.UUID, senderID uuid.UUID, msg *WSMessage) {
	// This would need a room->users lookup. For now, SendToUser is used directly.
}

func (h *WSHub) queueOffline(userID uuid.UUID, msg *WSMessage) error {
	key := fmt.Sprintf("offline:msgs:%s", userID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := h.redis.RPush(ctx, key, data).Err(); err != nil {
		return err
	}
	return h.redis.Expire(ctx, key, 7*24*time.Hour).Err()
}

func (h *WSHub) FetchOfflineMessages(userID uuid.UUID) ([]json.RawMessage, error) {
	key := fmt.Sprintf("offline:msgs:%s", userID)
	ctx := context.Background()
	// LPopCount atomically fetches and removes up to 100 messages (Redis 6.2+).
	// Messages are deleted immediately — if the client disconnects mid-send,
	// unsent messages are lost. Post-MVP: ACK-based deletion.
	vals, err := h.redis.LPopCount(ctx, key, 100).Result()
	if err != nil || len(vals) == 0 {
		return nil, err
	}
	var msgs []json.RawMessage
	for _, v := range vals {
		msgs = append(msgs, json.RawMessage(v))
	}
	return msgs, nil
}
