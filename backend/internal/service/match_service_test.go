package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
)

func TestJaccard(t *testing.T) {
	svc := &MatchService{}

	a := []model.InterestTag{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}, {ID: 5}}
	b := []model.InterestTag{{ID: 3}, {ID: 4}, {ID: 5}, {ID: 6}, {ID: 7}}

	score := svc.jaccard(a, b)
	// intersection = {3,4,5} = 3, union = {1,2,3,4,5,6,7} = 7
	expected := 3.0 / 7.0
	if score != expected {
		t.Errorf("jaccard = %v, want %v", score, expected)
	}

	// Empty sets
	if svc.jaccard(nil, b) != 0 {
		t.Error("jaccard with empty should be 0")
	}
	if svc.jaccard(a, nil) != 0 {
		t.Error("jaccard with empty should be 0")
	}

	// Identical sets
	if svc.jaccard(a, a) != 1.0 {
		t.Error("jaccard of identical sets should be 1")
	}

	// Disjoint sets
	c := []model.InterestTag{{ID: 10}, {ID: 11}}
	if svc.jaccard(a, c) != 0 {
		t.Error("jaccard of disjoint sets should be 0")
	}
}

func TestPersonalityDistance(t *testing.T) {
	svc := &MatchService{}

	self := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 4.0},
		{Dimension: "openness", Score: 3.0},
	}
	other := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 4.0},
		{Dimension: "openness", Score: 3.0},
	}

	// Identical personalities -> distance 0 -> score 1.0
	score := svc.personalityDistance(self, other)
	if score != 1.0 {
		t.Errorf("identical personality should be 1.0, got %v", score)
	}

	// Max difference: scores 1 and 5, diff=4, normalized: 1-4/4=0
	other2 := []model.PersonalityDimension{
		{Dimension: "extraversion", Score: 1.0},
		{Dimension: "openness", Score: 5.0},
	}
	score2 := svc.personalityDistance(self, other2)
	// extraversion: diff=3, openness: diff=2, avg squares=(9+4)/2=6.5, sqrt≈2.55, 1-2.55/4≈0.3625
	if score2 < 0.35 || score2 > 0.38 {
		t.Errorf("max diff personality score ~0.3625, got %v", score2)
	}

	// Empty -> neutral 0.5
	if svc.personalityDistance(nil, other) != 0.5 {
		t.Error("empty personality should return neutral 0.5")
	}
}

func TestRecencyBoost(t *testing.T) {
	svc := &MatchService{}

	if svc.recencyBoost(time.Now()) != 1.0 {
		t.Error("just now should be 1.0")
	}
	if svc.recencyBoost(time.Now().Add(-2*time.Hour)) != 0.8 {
		t.Error("2h ago should be 0.8")
	}
	if svc.recencyBoost(time.Now().Add(-48*time.Hour)) != 0.5 {
		t.Error("48h ago should be 0.5")
	}
	if svc.recencyBoost(time.Now().Add(-100*time.Hour)) != 0.3 {
		t.Error("100h ago should be 0.3")
	}
	if svc.recencyBoost(time.Now().Add(-200*time.Hour)) != 0.1 {
		t.Error("200h ago should be 0.1")
	}
}

func TestDiversityPenalty(t *testing.T) {
	svc := &MatchService{}

	a := []model.InterestTag{{ID: 1}, {ID: 2}, {ID: 3}}
	b := []model.InterestTag{{ID: 3}, {ID: 4}, {ID: 5}}

	// uniqueInB = {4,5} = 2, len(b)=3, 2/3 ≈ 0.667
	score := svc.diversityBonus(a, b)
	if score != 2.0/3.0 {
		t.Errorf("diversity = %v, want %v", score, 2.0/3.0)
	}

	// All overlap -> 0 unique
	c := []model.InterestTag{{ID: 1}, {ID: 2}}
	if svc.diversityBonus(a, c) != 0 {
		t.Error("all overlapping should be 0")
	}

	// All unique -> 1.0
	d := []model.InterestTag{{ID: 10}, {ID: 11}}
	if svc.diversityBonus(a, d) != 1.0 {
		t.Error("all unique should be 1.0")
	}
}

func TestDailyKey(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	key := dailyKey(uid)
	if key == "" {
		t.Error("dailyKey should not be empty")
	}
}

func TestEndOfDay(t *testing.T) {
	eod := endOfDay()
	now := time.Now()
	if eod.Before(now) {
		t.Error("endOfDay should be in the future")
	}
	if eod.Sub(now) > 24*time.Hour {
		t.Error("endOfDay should be within 24 hours")
	}
}
