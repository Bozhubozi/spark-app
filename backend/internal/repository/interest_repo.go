package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"gorm.io/gorm"
)

type InterestRepo struct {
	db *gorm.DB
}

func NewInterestRepo(db *gorm.DB) *InterestRepo { return &InterestRepo{db: db} }

func (r *InterestRepo) AllTags(ctx context.Context) ([]model.InterestTag, error) {
	var tags []model.InterestTag
	err := r.db.WithContext(ctx).Order("category, name").Find(&tags).Error
	return tags, err
}

func (r *InterestRepo) SaveUserInterests(ctx context.Context, userID uuid.UUID, tagIDs []int) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserInterest{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, tagID := range tagIDs {
		if err := tx.Create(&model.UserInterest{UserID: userID, TagID: tagID}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *InterestRepo) QuestionsWithOptions(ctx context.Context) ([]model.PersonalityQuestion, error) {
	var questions []model.PersonalityQuestion
	err := r.db.WithContext(ctx).
		Preload("Options").
		Order("sort_order").
		Find(&questions).Error
	return questions, err
}

func (r *InterestRepo) SavePersonalityAnswers(ctx context.Context, userID uuid.UUID, answers []model.PersonalityAnswerItem) error {
	tx := r.db.WithContext(ctx).Begin()
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserPersonalityAnswer{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	for _, a := range answers {
		rec := model.UserPersonalityAnswer{
			UserID:     userID,
			QuestionID: a.QuestionID,
			OptionID:   a.OptionID,
		}
		if err := tx.Create(&rec).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *InterestRepo) GetUserPersonality(ctx context.Context, userID uuid.UUID) ([]model.PersonalityDimension, error) {
	type row struct {
		Dimension string
		AvgScore  float64
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT q.dimension, AVG(o.score)::float as avg_score
		FROM user_personality_answers a
		JOIN personality_questions q ON q.id = a.question_id
		JOIN personality_options o ON o.id = a.option_id
		WHERE a.user_id = ?
		GROUP BY q.dimension`, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	var dims []model.PersonalityDimension
	for _, r := range rows {
		dims = append(dims, model.PersonalityDimension{Dimension: r.Dimension, Score: r.AvgScore})
	}
	return dims, nil
}

func (r *InterestRepo) AvatarComponents(ctx context.Context) ([]model.AvatarComponent, error) {
	var comps []model.AvatarComponent
	err := r.db.WithContext(ctx).Order("category, rarity DESC").Find(&comps).Error
	return comps, err
}
