package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
)

type MatchHandler struct {
	svc          *service.MatchService
	wsHub        *service.WSHub
	userRepo     *repository.UserRepo
	zodiacSvc    *service.ZodiacService
	icebreakerSvc *service.IcebreakerService
	interestRepo  *repository.InterestRepo
	chatRepo      *repository.ChatRepo
}

func NewMatchHandler(svc *service.MatchService, wsHub *service.WSHub, ur *repository.UserRepo, zs *service.ZodiacService, is *service.IcebreakerService, ir *repository.InterestRepo, cr *repository.ChatRepo) *MatchHandler {
	return &MatchHandler{svc: svc, wsHub: wsHub, userRepo: ur, zodiacSvc: zs, icebreakerSvc: is, interestRepo: ir, chatRepo: cr}
}

func (h *MatchHandler) GetCandidates(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	city := c.DefaultQuery("city", "")
	genderStr := c.DefaultQuery("gender", "0")
	minAgeStr := c.DefaultQuery("min_age", "0")
	maxAgeStr := c.DefaultQuery("max_age", "0")

	var gender int8
	if g, err := parseInt(genderStr); err == nil && g >= 0 && g <= 2 {
		gender = int8(g)
	}
	var minAge, maxAge int
	if a, err := parseInt(minAgeStr); err == nil && a >= 18 && a <= 100 {
		minAge = a
	}
	if a, err := parseInt(maxAgeStr); err == nil && a >= 18 && a <= 100 {
		maxAge = a
	}

	candidates, err := h.svc.GetCandidates(c.Request.Context(), uid, city, gender, minAge, maxAge)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if candidates == nil {
		candidates = []model.User{}
	}
	c.JSON(http.StatusOK, candidates)
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func (h *MatchHandler) Swipe(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	var req model.MatchSwipeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	match, err := h.svc.Swipe(c.Request.Context(), uid, req.TargetUserID, req.Direction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := gin.H{"matched": false}
	if match != nil && match.Status == model.MatchStatusMatched {
		result["matched"] = true
		result["match_id"] = match.ID

		// Auto-create chat room on match
		room, _ := h.chatRepo.FindOrCreateRoom(c.Request.Context(), uid, req.TargetUserID)
		if room != nil {
			result["room_id"] = room.ID
		}

		// Generate icebreaker messages
		user, _ := h.userRepo.FindByID(c.Request.Context(), uid)
		target, _ := h.userRepo.FindByID(c.Request.Context(), req.TargetUserID)
		if user != nil && target != nil {
			userPersonality, _ := h.interestRepo.GetUserPersonality(c.Request.Context(), uid)
			icebreakers := h.icebreakerSvc.Generate(
				zodiacFromBirth(user.BirthDate),
				zodiacFromBirth(target.BirthDate),
				user.Interests,
				target.Interests,
				userPersonality,
			)
			result["icebreakers"] = icebreakers
		}

		matchData, _ := json.Marshal(gin.H{
			"match_id":            match.ID,
			"user_id_1":           match.UserID1,
			"user_id_2":           match.UserID2,
			"compatibility_score": match.Score,
		})
		msg := &service.WSMessage{
			Type:      "match.new",
			Data:      matchData,
			Timestamp: time.Now().Unix(),
		}
		_ = h.wsHub.SendToUser(match.UserID1, msg)
		_ = h.wsHub.SendToUser(match.UserID2, msg)
	}
	c.JSON(http.StatusOK, result)
}

func (h *MatchHandler) ZodiacCompat(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	targetID, err := uuid.Parse(c.Param("target_user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	target, err := h.userRepo.FindByID(c.Request.Context(), targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}

	userZodiac := zodiacFromBirth(user.BirthDate)
	targetZodiac := zodiacFromBirth(target.BirthDate)
	if userZodiac == "" || targetZodiac == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "birth date required for both users"})
		return
	}

	score := h.zodiacSvc.Compatibility(userZodiac, targetZodiac)
	report := h.zodiacSvc.Report(userZodiac, targetZodiac)

	c.JSON(http.StatusOK, gin.H{
		"user_zodiac":   userZodiac,
		"target_zodiac": targetZodiac,
		"score":         score,
		"report":        report,
	})
}

func (h *MatchHandler) LikesCount(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	count, err := h.svc.CountLikesReceived(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *MatchHandler) GetLikers(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	matches, err := h.svc.GetLikers(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type LikerItem struct {
		MatchID   string     `json:"match_id"`
		User      model.User `json:"user"`
		CreatedAt time.Time  `json:"created_at"`
	}
	result := make([]LikerItem, 0, len(matches))
	for _, m := range matches {
		if m.User1 != nil {
			result = append(result, LikerItem{
				MatchID:   m.ID.String(),
				User:      *m.User1,
				CreatedAt: m.CreatedAt,
			})
		}
	}
	c.JSON(http.StatusOK, result)
}

func (h *MatchHandler) RemainingSwipes(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	remaining := h.svc.RemainingSwipes(c.Request.Context(), uid)
	c.JSON(http.StatusOK, gin.H{"remaining": remaining})
}

func (h *MatchHandler) GetBlocked(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	matches, err := h.svc.GetBlocked(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type BlockedItem struct {
		MatchID   string     `json:"match_id"`
		User      model.User `json:"user"`
		CreatedAt time.Time  `json:"created_at"`
	}
	result := make([]BlockedItem, 0, len(matches))
	for _, m := range matches {
		if m.User2 != nil {
			result = append(result, BlockedItem{
				MatchID:   m.ID.String(),
				User:      *m.User2,
				CreatedAt: m.CreatedAt,
			})
		}
	}
	c.JSON(http.StatusOK, result)
}

func (h *MatchHandler) Unblock(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	var req struct {
		TargetUserID string `json:"target_user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id required"})
		return
	}
	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}
	if err := h.svc.Unblock(c.Request.Context(), uid, targetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *MatchHandler) GetMatches(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	matches, err := h.svc.GetMatches(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, matches)
}
