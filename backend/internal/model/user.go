package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Phone         *string    `gorm:"uniqueIndex;size:20" json:"phone,omitempty"`
	Email         *string    `gorm:"uniqueIndex;size:255" json:"email,omitempty"`
	WechatOpenID  *string    `gorm:"uniqueIndex;size:128" json:"-"`
	WechatUnionID *string    `gorm:"index;size:128" json:"-"`
	PasswordHash  string     `gorm:"not null" json:"-"`
	Nickname      string     `gorm:"uniqueIndex;size:50;not null" json:"nickname"`
	AvatarURL     *string    `gorm:"size:500" json:"avatar_url,omitempty"`
	Gender        int8       `gorm:"default:0" json:"gender"`
	BirthDate     *time.Time `json:"birth_date,omitempty"`
	Bio           *string    `gorm:"size:500" json:"bio,omitempty"`
	City          *string    `gorm:"size:100" json:"city,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastActiveAt  time.Time  `json:"last_active_at"`
	IsActive      bool       `gorm:"default:true" json:"is_active"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`

	Interests   []InterestTag          `gorm:"many2many:user_interests;" json:"interests,omitempty"`
	Personality []PersonalityDimension `gorm:"-" json:"personality,omitempty"`
}

type UserRegisterReq struct {
	Phone    string `json:"phone" binding:"required_without=Email"`
	Email    string `json:"email" binding:"required_without=Phone"`
	Password string `json:"password" binding:"required,min=6,max=50"`
	Nickname string `json:"nickname" binding:"required,min=1,max=50"`
}

type UserLoginReq struct {
	Account  string `json:"account" binding:"required"` // phone or email
	Password string `json:"password" binding:"required"`
}

type UserLoginResp struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type WechatLoginReq struct {
	Code string `json:"code" binding:"required"`
}

type DeleteAccountReq struct {
	Reason string `json:"reason"`
}

type UserReport struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ReporterID   uuid.UUID `gorm:"not null;index" json:"reporter_id"`
	TargetUserID uuid.UUID `gorm:"not null;index" json:"target_user_id"`
	Reason       string    `gorm:"size:500" json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}

// WechatTokenResp is the response from WeChat /sns/oauth2/access_token
type WechatTokenResp struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid"`
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
}
