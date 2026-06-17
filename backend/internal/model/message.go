package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatRoom struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MatchID       *uuid.UUID `gorm:"type:uuid" json:"match_id,omitempty"`
	UserID1       uuid.UUID `gorm:"not null;index" json:"user_id_1"`
	UserID2       uuid.UUID `gorm:"not null;index" json:"user_id_2"`
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type Message struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RoomID      uuid.UUID `gorm:"not null;index" json:"room_id"`
	SenderID    uuid.UUID `gorm:"not null" json:"sender_id"`
	ClientMsgID string    `gorm:"uniqueIndex;size:100;not null" json:"client_msg_id"`
	ContentType int8      `gorm:"default:1" json:"content_type"` // 1:text, 2:image, 3:sticker
	Content     string    `gorm:"type:text;not null" json:"content"`
	IsRead      bool      `gorm:"default:false" json:"is_read"`
	SentAt      time.Time `json:"sent_at"`
	CreatedAt   time.Time `json:"created_at"`

	Sender *User `gorm:"foreignKey:SenderID" json:"sender,omitempty"`
}

const (
	ContentTypeText    = 1
	ContentTypeImage   = 2
	ContentTypeSticker = 3
)
