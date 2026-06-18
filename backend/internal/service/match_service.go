package service

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
)

const (
	weightJaccard       = 0.40
	weightPersonality   = 0.30
	weightRecency       = 0.20
	weightDiversity     = 0.10
	maxCandidates       = 100
	dailyCandidateLimit = 50 // page views per day (not candidates)
	minCandidates       = 10
)

type MatchService struct {
	matchRepo    *repository.MatchRepo
	interestRepo *repository.InterestRepo
	userRepo     *repository.UserRepo
	rdb          *redis.Client
}

func NewMatchService(mr *repository.MatchRepo, ir *repository.InterestRepo, ur *repository.UserRepo, rdb *redis.Client) *MatchService {
	return &MatchService{matchRepo: mr, interestRepo: ir, userRepo: ur, rdb: rdb}
}

func (s *MatchService) GetCandidates(ctx context.Context, userID uuid.UUID, city string, gender int8, minAge, maxAge int) ([]model.User, error) {
	// Check daily page-view limit
	if s.dailyRemaining(ctx, userID) <= 0 {
		return nil, nil
	}

	// Progressive relaxation
	candidates, err := s.fetchWithRelaxation(ctx, userID, city, gender, minAge, maxAge, maxCandidates)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// Track daily page views (1 per call, not per candidate)
	_ = s.rdb.IncrBy(ctx, dailyKey(userID), 1).Err()
	_ = s.rdb.ExpireAt(ctx, dailyKey(userID), endOfDay()).Err()

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	userPersonality, _ := s.interestRepo.GetUserPersonality(ctx, userID)

	type scored struct {
		user  model.User
		score float64
	}
	var scoredList []scored

	for _, c := range candidates {
		score := s.computeScore(user, &c, userPersonality)
		scoredList = append(scoredList, scored{user: c, score: score})
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].score > scoredList[j].score
	})

	result := make([]model.User, len(scoredList))
	for i, sc := range scoredList {
		result[i] = sc.user
	}
	return result, nil
}

func (s *MatchService) Swipe(ctx context.Context, userID, targetID uuid.UUID, direction string) (*model.Match, error) {
	existing, err := s.matchRepo.FindExisting(ctx, userID, targetID)

	if direction == "pass" {
		if err == nil && existing.Status == model.MatchStatusPending && existing.UserID2 == userID {
			_ = s.matchRepo.UpdateStatus(ctx, existing.ID, model.MatchStatusRejected)
		}
		if err != nil {
			m := &model.Match{
				UserID1: userID,
				UserID2: targetID,
				Status:  model.MatchStatusRejected,
			}
			_ = s.matchRepo.Create(ctx, m)
		}
		return nil, nil
	}

	// direction == "like"
	if err == nil {
		if existing.Status == model.MatchStatusPending && existing.UserID2 == userID {
			now := time.Now()
			existing.Status = model.MatchStatusMatched
			existing.MatchedAt = &now
			_ = s.matchRepo.UpdateStatus(ctx, existing.ID, model.MatchStatusMatched)
			return existing, nil
		}
		return nil, nil
	}

	// No existing match record, create pending
	m := &model.Match{
		UserID1: userID,
		UserID2: targetID,
		Status:  model.MatchStatusPending,
	}
	if err := s.matchRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *MatchService) CountLikesReceived(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.matchRepo.CountLikesReceived(ctx, userID)
}

func (s *MatchService) GetLikers(ctx context.Context, userID uuid.UUID) ([]model.Match, error) {
	return s.matchRepo.FindLikers(ctx, userID)
}

func (s *MatchService) GetMatches(ctx context.Context, userID uuid.UUID) ([]model.Match, error) {
	return s.matchRepo.FindMatches(ctx, userID)
}

func (s *MatchService) RemainingSwipes(ctx context.Context, userID uuid.UUID) int {
	return s.dailyRemaining(ctx, userID)
}

func (s *MatchService) GetBlocked(ctx context.Context, userID uuid.UUID) ([]model.Match, error) {
	return s.matchRepo.FindBlocked(ctx, userID)
}

func (s *MatchService) Unblock(ctx context.Context, userID, targetID uuid.UUID) error {
	return s.matchRepo.Unblock(ctx, userID, targetID)
}

func (s *MatchService) computeScore(user, candidate *model.User, userPersonality []model.PersonalityDimension) float64 {
	jaccardScore := s.jaccard(user.Interests, candidate.Interests)
	personalityScore := s.personalityDistance(userPersonality, candidate.Personality)
	recencyScore := s.recencyBoost(candidate.LastActiveAt)
	diversityScore := s.diversityBonus(user.Interests, candidate.Interests)

	return weightJaccard*jaccardScore +
		weightPersonality*personalityScore +
		weightRecency*recencyScore +
		weightDiversity*diversityScore
}

func (s *MatchService) jaccard(a, b []model.InterestTag) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	setA := make(map[int]bool, len(a))
	for _, t := range a {
		setA[t.ID] = true
	}
	intersection := 0
	for _, t := range b {
		if setA[t.ID] {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	return float64(intersection) / float64(union)
}

func (s *MatchService) personalityDistance(self, other []model.PersonalityDimension) float64 {
	if len(self) == 0 || len(other) == 0 {
		return 0.5 // neutral
	}
	selfMap := make(map[string]float64, len(self))
	for _, d := range self {
		selfMap[d.Dimension] = d.Score
	}
	sumSquares := 0.0
	count := 0
	for _, d := range other {
		if sv, ok := selfMap[d.Dimension]; ok {
			diff := sv - d.Score
			sumSquares += diff * diff
			count++
		}
	}
	if count == 0 {
		return 0.5
	}
	dist := math.Sqrt(sumSquares / float64(count))
	// Normalize: max possible distance is 4 (scores 1-5), convert to 0-1 where 1=most similar
	return 1.0 - dist/4.0
}

func (s *MatchService) recencyBoost(lastActive time.Time) float64 {
	hours := time.Since(lastActive).Hours()
	if hours < 1 {
		return 1.0
	}
	if hours < 24 {
		return 0.8
	}
	if hours < 72 {
		return 0.5
	}
	if hours < 168 { // 7 days
		return 0.3
	}
	return 0.1
}

func (s *MatchService) diversityBonus(a, b []model.InterestTag) float64 {
	setA := make(map[int]bool, len(a))
	for _, t := range a {
		setA[t.ID] = true
	}
	uniqueInB := 0
	for _, t := range b {
		if !setA[t.ID] {
			uniqueInB++
		}
	}
	// Reward discovering new interests (diversity bonus, not penalty)
	return math.Min(1.0, float64(uniqueInB)/float64(max(len(b), 1)))
}

// Progressive relaxation tiers for candidate pool.
func (s *MatchService) fetchWithRelaxation(ctx context.Context, userID uuid.UUID, city string, gender int8, minAge, maxAge int, limit int) ([]model.User, error) {
	tiers := []repository.CandidateFilters{
		{City: city, Gender: gender, MinAge: minAge, MaxAge: maxAge, Limit: limit, MinLastActiveHours: 72},
		{City: city, Gender: gender, MinAge: minAge, MaxAge: maxAge, Limit: limit, MinLastActiveHours: 168},
		{City: "", Gender: gender, MinAge: minAge, MaxAge: maxAge, Limit: limit, MinLastActiveHours: 168},
		{City: "", Gender: gender, MinAge: minAge, MaxAge: maxAge, Limit: limit, MinLastActiveHours: 720},
	}

	var result []model.User
	for _, f := range tiers {
		candidates, err := s.userRepo.FindCandidates(ctx, userID, f)
		if err != nil {
			return nil, err
		}
		result = append(result, candidates...)
		if len(result) >= minCandidates {
			break
		}
	}

	seen := map[uuid.UUID]bool{}
	var unique []model.User
	for _, u := range result {
		if !seen[u.ID] {
			seen[u.ID] = true
			unique = append(unique, u)
		}
	}
	if len(unique) > limit {
		unique = unique[:limit]
	}
	return unique, nil
}

func (s *MatchService) dailyRemaining(ctx context.Context, userID uuid.UUID) int {
	val, err := s.rdb.Get(ctx, dailyKey(userID)).Int()
	if err != nil {
		return dailyCandidateLimit
	}
	return max(0, dailyCandidateLimit-val)
}

func dailyKey(userID uuid.UUID) string {
	return fmt.Sprintf("candidates:seen:%s:%s", time.Now().Format("2006-01-02"), userID.String())
}

func endOfDay() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
}
