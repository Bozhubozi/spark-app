package service

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type ABService struct {
	rdb *redis.Client
}

func NewABService(rdb *redis.Client) *ABService {
	return &ABService{rdb: rdb}
}

// Experiment defines an A/B test.
type Experiment struct {
	ID        string            // e.g. "matching_algorithm_v2"
	Variants  []string          // e.g. ["control", "jaccard_only", "personality_boost"]
	Weights   []int             // e.g. [50, 25, 25] = 50/25/25 split
	Enabled   bool
}

// Assign returns the variant name for a given user in an experiment.
// Assignment is deterministic per user+experiment (consistent bucketing).
func (s *ABService) Assign(ctx context.Context, userID, experimentID string, variants []string, weights []int) string {
	// Check Redis override first
	cacheKey := fmt.Sprintf("ab:%s:%s", experimentID, userID)
	val, err := s.rdb.Get(ctx, cacheKey).Result()
	if err == nil && val != "" {
		return val
	}

	// Deterministic assignment based on hash of user+experiment
	variant := deterministicAssign(userID+experimentID, variants, weights)

	// Cache for consistency
	_ = s.rdb.Set(ctx, cacheKey, variant, 0).Err()
	return variant
}

// GetExperiment fetches experiment config from Redis.
func (s *ABService) GetExperiment(ctx context.Context, id string) (*Experiment, error) {
	key := fmt.Sprintf("ab:config:%s", id)
	data, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil || len(data) == 0 {
		return nil, fmt.Errorf("experiment not found: %s", id)
	}
	enabled := data["enabled"] == "true"
	variants := []string{}
	weights := []int{}
	// Parse variants: v1,w1|v2,w2|v3,w3
	raw := data["variants"]
	for _, part := range splitStr(raw, "|") {
		parts := splitStr(part, ",")
		if len(parts) >= 2 {
			variants = append(variants, parts[0])
			w, _ := strconv.Atoi(parts[1])
			weights = append(weights, w)
		}
	}
	return &Experiment{ID: id, Variants: variants, Weights: weights, Enabled: enabled}, nil
}

// SetExperiment stores experiment config in Redis.
func (s *ABService) SetExperiment(ctx context.Context, e *Experiment) error {
	key := fmt.Sprintf("ab:config:%s", e.ID)
	variants := ""
	for i, v := range e.Variants {
		if i > 0 {
			variants += "|"
		}
		variants += fmt.Sprintf("%s,%d", v, e.Weights[i])
	}
	return s.rdb.HSet(ctx, key,
		"enabled", fmt.Sprint(e.Enabled),
		"variants", variants,
	).Err()
}

func deterministicAssign(key string, variants []string, weights []int) string {
	if len(variants) == 0 {
		return "default"
	}
	if len(variants) == 1 {
		return variants[0]
	}

	hash := md5.Sum([]byte(key))
	hashInt := int(hash[0])<<16 | int(hash[1])<<8 | int(hash[2])

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}
	if totalWeight == 0 {
		return variants[rand.Intn(len(variants))]
	}

	bucket := hashInt % totalWeight
	cumulative := 0
	for i, w := range weights {
		cumulative += w
		if bucket < cumulative {
			return variants[i]
		}
	}
	return variants[len(variants)-1]
}

func splitStr(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}
