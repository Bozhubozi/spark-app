package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
	"github.com/spark-app/backend/internal/util"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub      *service.WSHub
	chatSvc  *service.ChatService
	authSvc  *service.AuthService
	chatRepo *repository.ChatRepo
	dfa      *util.DFAFilter
}

func NewWSHandler(hub *service.WSHub, chatSvc *service.ChatService, authSvc *service.AuthService, cr *repository.ChatRepo, dfa *util.DFAFilter) *WSHandler {
	return &WSHandler{hub: hub, chatSvc: chatSvc, authSvc: authSvc, chatRepo: cr, dfa: dfa}
}

func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
		return
	}
	userID, err := h.authSvc.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] upgrade error: %v", err)
		return
	}

	client := &wsConn{
		userID:  userID,
		conn:    conn,
		send:    make(chan []byte, 64),
		handler: h,
	}

	h.hub.Register(client)
	go client.writePump()
	go client.readPump()
	go h.sendOfflineMessages(userID, client)
}

func (h *WSHandler) sendOfflineMessages(userID uuid.UUID, client *wsConn) {
	msgs, err := h.hub.FetchOfflineMessages(userID)
	if err != nil || len(msgs) == 0 {
		return
	}
	for _, raw := range msgs {
		client.send <- raw
	}
}

type wsConn struct {
	userID  uuid.UUID
	conn    *websocket.Conn
	send    chan []byte
	handler *WSHandler
}

func (c *wsConn) UserID() uuid.UUID            { return c.userID }
func (c *wsConn) SendJSON(v interface{}) error { return c.conn.WriteJSON(v) }

func (c *wsConn) readPump() {
	defer func() {
		c.handler.hub.Unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg service.WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		c.handleMessage(&msg)
	}
}

func (c *wsConn) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case data, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *wsConn) handleMessage(msg *service.WSMessage) {
	ctx := context.Background()
	msg.Timestamp = time.Now().UnixMilli()

	switch msg.Type {
	case "chat.message.send":
		var data service.ChatMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return
		}
		data.SenderID = c.userID
		if data.ClientMsgID == "" {
			data.ClientMsgID = uuid.New().String()
		}

		// Sensitive word filtering
		if c.handler.dfa != nil && c.handler.dfa.Contains(data.Content) {
			c.SendJSON(&service.WSMessage{
				Type: "system.error",
				Data: json.RawMessage(`{"client_msg_id":"` + data.ClientMsgID + `","error":"message contains sensitive content"}`),
			})
			return
		}

		// Verify sender belongs to this chat room
		room, err := c.handler.chatRepo.FindByID(ctx, data.RoomID)
		if err != nil || room == nil || (room.UserID1 != c.userID && room.UserID2 != c.userID) {
			return // silently drop — unauthorized sender
		}

		dbMsg := &model.Message{
			RoomID:      data.RoomID,
			SenderID:    data.SenderID,
			ClientMsgID: data.ClientMsgID,
			ContentType: data.ContentType,
			Content:     data.Content,
		}

		if err := c.handler.chatSvc.SaveMessage(ctx, dbMsg); err != nil {
			log.Printf("[WS] save message error: %v", err)
			c.SendJSON(&service.WSMessage{
				Type: "system.error",
				Data: json.RawMessage(`{"client_msg_id":"` + data.ClientMsgID + `","error":"failed to save"}`),
			})
			return
		}

		// Ack to sender
		ackData, _ := json.Marshal(map[string]string{
			"client_msg_id": data.ClientMsgID,
			"server_msg_id": dbMsg.ID.String(),
		})
		c.SendJSON(&service.WSMessage{Type: "chat.message.ack", Data: ackData})

		// Forward to receiver
		receiverID := c.getReceiver(data.RoomID)
		if receiverID != uuid.Nil {
			newMsgData, _ := json.Marshal(map[string]interface{}{
				"room_id":       data.RoomID,
				"sender_id":     data.SenderID,
				"client_msg_id": data.ClientMsgID,
				"server_msg_id": dbMsg.ID.String(),
				"content_type":  data.ContentType,
				"content":       data.Content,
				"sent_at":       dbMsg.SentAt,
			})
			c.handler.hub.SendToUser(receiverID, &service.WSMessage{
				Type: "chat.message.new",
				Data: newMsgData,
			})
		}

	case "chat.message.read":
		var data struct {
			RoomID   string `json:"room_id"`
			ReaderID string `json:"reader_id"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return
		}
		roomID, _ := uuid.Parse(data.RoomID)
		if roomID == uuid.Nil {
			return
		}
		_ = c.handler.chatSvc.MarkRead(ctx, roomID, c.userID)

		// Broadcast read status to the partner
		receiverID := c.getReceiver(roomID)
		if receiverID != uuid.Nil {
			readData, _ := json.Marshal(map[string]string{"room_id": data.RoomID, "reader_id": c.userID.String()})
			c.handler.hub.SendToUser(receiverID, &service.WSMessage{
				Type: "chat.message.read",
				Data: readData,
			})
		}

	case "chat.typing":
		var data struct {
			RoomID   string `json:"room_id"`
			IsTyping bool   `json:"is_typing"`
		}
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			return
		}
		roomID, _ := uuid.Parse(data.RoomID)
		receiverID := c.getReceiver(roomID)
		if receiverID != uuid.Nil {
			typingData, _ := json.Marshal(map[string]interface{}{
				"room_id":   data.RoomID,
				"user_id":   c.userID.String(),
				"is_typing": data.IsTyping,
			})
			c.handler.hub.SendToUser(receiverID, &service.WSMessage{
				Type: "chat.typing",
				Data: typingData,
			})
		}

	case "system.heartbeat":
		c.SendJSON(&service.WSMessage{Type: "system.heartbeat", Data: json.RawMessage(`{}`)})
	}
}

func (c *wsConn) getReceiver(roomID uuid.UUID) uuid.UUID {
	if roomID == uuid.Nil {
		return uuid.Nil
	}
	room, err := c.handler.chatRepo.FindByID(context.Background(), roomID)
	if err != nil || room == nil {
		return uuid.Nil
	}
	if room.UserID1 == c.userID {
		return room.UserID2
	}
	return room.UserID1
}
