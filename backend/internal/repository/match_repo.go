package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"gorm.io/gorm"
)

type MatchRepo struct {
	db *gorm.DB
}

func NewMatchRepo(db *gorm.DB) *MatchRepo { return &MatchRepo{db: db} }

func (r *MatchRepo) Create(ctx context.Context, match *model.Match) error {
	return r.db.WithContext(ctx).Create(match).Error
}

func (r *MatchRepo) FindExisting(ctx context.Context, user1, user2 uuid.UUID) (*model.Match, error) {
	var m model.Match
	err := r.db.WithContext(ctx).
		Where("(user_id_1 = ? AND user_id_2 = ?) OR (user_id_1 = ? AND user_id_2 = ?)",
			user1, user2, user2, user1).
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *MatchRepo) UpdateStatus(ctx context.Context, matchID uuid.UUID, status int8) error {
	return r.db.WithContext(ctx).
		Model(&model.Match{}).
		Where("id = ?", matchID).
		Update("status", status).Error
}

func (r *MatchRepo) CountLikesReceived(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Match{}).
		Where("user_id_2 = ? AND status = ?", userID, model.MatchStatusPending).
		Count(&count).Error
	return count, err
}

func (r *MatchRepo) FindLikers(ctx context.Context, userID uuid.UUID) ([]model.Match, error) {
	var matches []model.Match
	err := r.db.WithContext(ctx).
		Preload("User1").
		Where("user_id_2 = ? AND status = ?", userID, model.MatchStatusPending).
		Order("created_at DESC").
		Find(&matches).Error
	return matches, err
}

func (r *MatchRepo) FindBlocked(ctx context.Context, userID uuid.UUID) ([]model.Match, error) {
	var matches []model.Match
	err := r.db.WithContext(ctx).
		Preload("User2").
		Where("user_id_1 = ? AND status = ?", userID, model.MatchStatusRejected).
		Order("created_at DESC").
		Find(&matches).Error
	return matches, err
}

func (r *MatchRepo) Unblock(ctx context.Context, userID, targetID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id_1 = ? AND user_id_2 = ? AND status = ?", userID, targetID, model.MatchStatusRejected).
		Delete(&model.Match{}).Error
}

func (r *MatchRepo) FindMatches(ctx context.Context, userID uuid.UUID) ([]model.Match, error) {
	var matches []model.Match
	err := r.db.WithContext(ctx).
		Preload("User1").Preload("User2").
		Where("(user_id_1 = ? OR user_id_2 = ?) AND status = ?", userID, userID, model.MatchStatusMatched).
		Order("matched_at DESC").
		Find(&matches).Error
	return matches, err
}
