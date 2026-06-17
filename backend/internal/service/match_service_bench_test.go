package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
)

func BenchmarkJaccard100(b *testing.B) {
	svc := &MatchService{}
	a := make([]model.InterestTag, 100)
	b_ := make([]model.InterestTag, 100)
	for i := range 100 {
		a[i] = model.InterestTag{ID: i}
		b_[i] = model.InterestTag{ID: i + 50}
	}
	b.ResetTimer()
	for range b.N {
		svc.jaccard(a, b_)
	}
}

func BenchmarkJaccard10(b *testing.B) {
	svc := &MatchService{}
	a := make([]model.InterestTag, 10)
	b_ := make([]model.InterestTag, 10)
	for i := range 10 {
		a[i] = model.InterestTag{ID: i}
		b_[i] = model.InterestTag{ID: i + 3}
	}
	b.ResetTimer()
	for range b.N {
		svc.jaccard(a, b_)
	}
}

func BenchmarkPersonalityDistance(b *testing.B) {
	svc := &MatchService{}
	self := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 4.0},
		{Dimension: "agreeableness", Score: 3.0},
		{Dimension: "conscientiousness", Score: 4.0},
		{Dimension: "neuroticism", Score: 2.0},
		{Dimension: "openness", Score: 4.5},
	}
	other := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 3.0},
		{Dimension: "agreeableness", Score: 4.0},
		{Dimension: "conscientiousness", Score: 3.0},
		{Dimension: "neuroticism", Score: 3.0},
		{Dimension: "openness", Score: 3.5},
	}
	b.ResetTimer()
	for range b.N {
		svc.personalityDistance(self, other)
	}
}

func BenchmarkRecencyBoost(b *testing.B) {
	svc := &MatchService{}
	now := time.Now()
	fewHoursAgo := now.Add(-3 * time.Hour)
	fewDaysAgo := now.Add(-48 * time.Hour)
	weekAgo := now.Add(-100 * time.Hour)
	monthAgo := now.Add(-200 * time.Hour)

	b.Run("JustNow", func(b *testing.B) {
		for range b.N {
			svc.recencyBoost(now)
		}
	})
	b.Run("FewHours", func(b *testing.B) {
		for range b.N {
			svc.recencyBoost(fewHoursAgo)
		}
	})
	b.Run("FewDays", func(b *testing.B) {
		for range b.N {
			svc.recencyBoost(fewDaysAgo)
		}
	})
	b.Run("WeekAgo", func(b *testing.B) {
		for range b.N {
			svc.recencyBoost(weekAgo)
		}
	})
	b.Run("MonthAgo", func(b *testing.B) {
		for range b.N {
			svc.recencyBoost(monthAgo)
		}
	})
}

func BenchmarkComputeScore(b *testing.B) {
	svc := &MatchService{}
	uid := uuid.New()
	user := &model.User{
		ID:           uid,
		Nickname:     "test",
		LastActiveAt: time.Now().Add(-2 * time.Hour),
		Interests: []model.InterestTag{
			{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5},
			{ID: 6}, {ID: 7}, {ID: 8}, {ID: 9}, {ID: 10},
		},
	}
	candidate := &model.User{
		ID:       uuid.New(),
		Nickname: "candidate",
		Interests: []model.InterestTag{
			{ID: 3}, {ID: 4}, {ID: 5}, {ID: 10}, {ID: 11},
			{ID: 12}, {ID: 13}, {ID: 14}, {ID: 15}, {ID: 16},
		},
		LastActiveAt: time.Now(),
		Personality: []model.PersonalityDimension{
			{Dimension: "extraversion", Score: 3.5},
			{Dimension: "agreeableness", Score: 4.0},
		},
	}
	userPersonality := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 4.0},
		{Dimension: "agreeableness", Score: 3.0},
	}

	b.ResetTimer()
	for range b.N {
		svc.computeScore(user, candidate, userPersonality)
	}
}
