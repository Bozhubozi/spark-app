package model

import (
	"time"

	"github.com/google/uuid"
)

type Match struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID1   uuid.UUID  `gorm:"column:user_id_1;not null;index" json:"user_id_1"`
	UserID2   uuid.UUID  `gorm:"column:user_id_2;not null;index" json:"user_id_2"`
	Score     float64    `gorm:"not null;default:0" json:"score"`
	Status    int8       `gorm:"default:0" json:"status"` // 0:pending, 1:matched, 2:rejected
	CreatedAt time.Time  `json:"created_at"`
	MatchedAt *time.Time `json:"matched_at,omitempty"`

	User1 *User `gorm:"foreignKey:UserID1" json:"user1,omitempty"`
	User2 *User `gorm:"foreignKey:UserID2" json:"user2,omitempty"`
}

const (
	MatchStatusPending  = 0
	MatchStatusMatched  = 1
	MatchStatusRejected = 2
)

type MatchSwipeReq struct {
	TargetUserID uuid.UUID `json:"target_user_id" binding:"required"`
	Direction    string    `json:"direction" binding:"required,oneof=like pass"`
}
