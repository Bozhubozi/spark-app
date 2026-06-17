package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"gorm.io/gorm"
)

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo { return &ChatRepo{db: db} }

func (r *ChatRepo) FindOrCreateRoom(ctx context.Context, user1, user2 uuid.UUID) (*model.ChatRoom, error) {
	var room model.ChatRoom
	err := r.db.WithContext(ctx).
		Where("(user_id_1 = ? AND user_id_2 = ?) OR (user_id_1 = ? AND user_id_2 = ?)",
			user1, user2, user2, user1).
		First(&room).Error
	if err == nil {
		return &room, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	room = model.ChatRoom{
		UserID1: user1,
		UserID2: user2,
	}
	if err := r.db.WithContext(ctx).Create(&room).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *ChatRepo) SaveMessage(ctx context.Context, msg *model.Message) error {
	msg.SentAt = time.Now()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}
		return tx.Model(&model.ChatRoom{}).
			Where("id = ?", msg.RoomID).
			Update("last_message_at", msg.SentAt).Error
	})
}

func (r *ChatRepo) FindMessages(ctx context.Context, roomID uuid.UUID, before time.Time, limit int) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.WithContext(ctx).
		Preload("Sender").
		Where("room_id = ? AND sent_at < ?", roomID, before).
		Order("sent_at DESC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

type RoomWithLastMsg struct {
	model.ChatRoom
	LastContent    *string    `json:"last_content"`
	LastSentAt     *time.Time `json:"last_sent_at"`
	UnreadCount    int        `json:"unread_count"`
	OtherNickname  string     `json:"other_nickname"`
	OtherAvatarURL *string    `json:"other_avatar_url"`
}

func (r *ChatRepo) FindRooms(ctx context.Context, userID uuid.UUID) ([]model.ChatRoom, error) {
	var rooms []model.ChatRoom
	err := r.db.WithContext(ctx).
		Where("user_id_1 = ? OR user_id_2 = ?", userID, userID).
		Order("last_message_at DESC").
		Find(&rooms).Error
	return rooms, err
}

func (r *ChatRepo) FindRoomsWithDetails(ctx context.Context, userID uuid.UUID) ([]RoomWithLastMsg, error) {
	var results []RoomWithLastMsg
	err := r.db.WithContext(ctx).Raw(`
		SELECT c.*,
			m.content AS last_content,
			m.sent_at AS last_sent_at,
			(SELECT COUNT(*) FROM messages WHERE room_id = c.id AND sender_id != ? AND is_read = false) AS unread_count,
			CASE WHEN c.user_id_1 = ? THEN u2.nickname ELSE u1.nickname END AS other_nickname,
			CASE WHEN c.user_id_1 = ? THEN u2.avatar_url ELSE u1.avatar_url END AS other_avatar_url
		FROM chat_rooms c
		LEFT JOIN users u1 ON c.user_id_1 = u1.id
		LEFT JOIN users u2 ON c.user_id_2 = u2.id
		LEFT JOIN LATERAL (
			SELECT content, sent_at FROM messages
			WHERE room_id = c.id
			ORDER BY sent_at DESC LIMIT 1
		) m ON true
		WHERE c.user_id_1 = ? OR c.user_id_2 = ?
		ORDER BY COALESCE(m.sent_at, c.last_message_at) DESC
	`, userID, userID, userID, userID, userID).Scan(&results).Error
	return results, err
}

func (r *ChatRepo) FindByID(ctx context.Context, roomID uuid.UUID) (*model.ChatRoom, error) {
	var room model.ChatRoom
	err := r.db.WithContext(ctx).First(&room, "id = ?", roomID).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *ChatRepo) MarkRead(ctx context.Context, roomID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.Message{}).
		Where("room_id = ? AND sender_id != ? AND is_read = false", roomID, userID).
		Update("is_read", true).Error
}
