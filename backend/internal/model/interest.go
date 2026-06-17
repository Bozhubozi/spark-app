package model

import (
	"time"

	"github.com/google/uuid"
)

type InterestTag struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Category string `gorm:"size:50;not null" json:"category"`
	Icon     string `gorm:"size:100" json:"icon,omitempty"`
}

type UserInterest struct {
	UserID uuid.UUID `gorm:"primaryKey" json:"user_id"`
	TagID  int       `gorm:"primaryKey" json:"tag_id"`
	Weight int       `gorm:"default:1" json:"weight"`
}

type PersonalityQuestion struct {
	ID           int                 `gorm:"primaryKey" json:"id"`
	Dimension    string              `gorm:"size:50;not null" json:"dimension"`
	QuestionText string              `gorm:"size:500;not null" json:"question_text"`
	SortOrder    int                 `json:"sort_order"`
	Options      []PersonalityOption `gorm:"foreignKey:QuestionID" json:"options"`
}

type PersonalityOption struct {
	ID         int    `gorm:"primaryKey" json:"id"`
	QuestionID int    `json:"question_id"`
	OptionText string `gorm:"size:200;not null" json:"option_text"`
	Score      int    `gorm:"not null" json:"score"`
	SortOrder  int    `json:"sort_order"`
}

type UserPersonalityAnswer struct {
	UserID     uuid.UUID `gorm:"primaryKey" json:"user_id"`
	QuestionID int       `gorm:"primaryKey" json:"question_id"`
	OptionID   int       `json:"option_id"`
}

type PersonalityDimension struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"`
}

type PersonalitySubmitReq struct {
	Answers []PersonalityAnswerItem `json:"answers" binding:"required,min=1"`
}

type PersonalityAnswerItem struct {
	QuestionID int `json:"question_id" binding:"required"`
	OptionID   int `json:"option_id" binding:"required"`
}

type AvatarComponent struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	Category string `gorm:"size:50;not null" json:"category"`
	Name     string `gorm:"size:100;not null" json:"name"`
	ImageURL string `gorm:"size:500;not null" json:"image_url"`
	Rarity   int8   `gorm:"default:1" json:"rarity"`
}

type DeviceToken struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"size:500;not null" json:"token"`
	Platform  string    `gorm:"size:20;not null" json:"platform"`
	CreatedAt time.Time `json:"created_at"`
}

type PersonalityReport struct {
	Title              string   `json:"title"`
	Summary            string   `json:"summary"`
	Traits             []string `json:"traits"`
	Advice             string   `json:"advice"`
	ExtraversionDetail string   `json:"extraversion_detail"`
}
