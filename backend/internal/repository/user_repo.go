package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) FindByAccount(ctx context.Context, account string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("phone = ? OR email = ?", account, account).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByWechatOpenID(ctx context.Context, openID string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("wechat_open_id = ?", openID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("Interests").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepo) FindCandidates(ctx context.Context, userID uuid.UUID, filters CandidateFilters) ([]model.User, error) {
	var users []model.User

	// Exclude all users the current user has already matched or rejected (both directions)
	matchSub := r.db.Model(&model.Match{}).
		Select("CASE WHEN user_id_1 = ? THEN user_id_2 ELSE user_id_1 END", userID).
		Where("(user_id_1 = ? OR user_id_2 = ?)", userID, userID)

	query := r.db.WithContext(ctx).
		Preload("Interests").
		Where("id != ? AND is_active = true AND deleted_at IS NULL", userID).
		Where("id NOT IN (?)", matchSub)

	if filters.City != "" {
		query = query.Where("city = ?", filters.City)
	}
	if filters.MinLastActiveHours > 0 {
		query = query.Where("last_active_at > NOW() - interval '1 hour' * ?", filters.MinLastActiveHours)
	}
	if filters.Gender > 0 {
		query = query.Where("gender = ?", filters.Gender)
	}
	if filters.MinAge > 0 {
		query = query.Where("birth_date <= NOW() - interval '1 year' * ?", filters.MinAge)
	}
	if filters.MaxAge > 0 {
		query = query.Where("birth_date >= NOW() - interval '1 year' * ?", filters.MaxAge+1)
	}

	err := query.Order("last_active_at DESC").
		Limit(filters.Limit).
		Find(&users).Error
	return users, err
}

type CandidateFilters struct {
	City              string
	Limit             int
	MinLastActiveHours int // 0 = no filter
	Gender            int8 // 0 = any
	MinAge            int  // 0 = no filter
	MaxAge            int  // 0 = no filter
}

func (r *UserRepo) SaveDeviceToken(ctx context.Context, dt *model.DeviceToken) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND token = ?", dt.UserID, dt.Token).
		FirstOrCreate(dt).Error
}

func (r *UserRepo) RequestDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": now,
		"is_active":  false,
	}).Error
}

func (r *UserRepo) SaveReport(ctx context.Context, report *model.UserReport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *UserRepo) RestoreAccount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"deleted_at": nil,
		"is_active":  true,
	}).Error
}
