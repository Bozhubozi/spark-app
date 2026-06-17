package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
)

type UserHandler struct {
	userRepo       *repository.UserRepo
	interestRepo   *repository.InterestRepo
	personalitySvc *service.PersonalityReportService
	zodiacSvc      *service.ZodiacService
	horoscopeSvc   *service.HoroscopeService
}

func NewUserHandler(ur *repository.UserRepo, ir *repository.InterestRepo, prs *service.PersonalityReportService, zs *service.ZodiacService, hs *service.HoroscopeService) *UserHandler {
	return &UserHandler{userRepo: ur, interestRepo: ir, personalitySvc: prs, zodiacSvc: zs, horoscopeSvc: hs}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	user, err := h.userRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	user, err := h.userRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if v, ok := updates["nickname"]; ok {
		user.Nickname = v.(string)
	}
	if v, ok := updates["bio"]; ok {
		bio := v.(string)
		user.Bio = &bio
	}
	if v, ok := updates["city"]; ok {
		city := v.(string)
		user.City = &city
	}
	if v, ok := updates["gender"]; ok {
		user.Gender = int8(v.(float64))
	}
	if v, ok := updates["birth_date"]; ok {
		if birthStr, ok := v.(string); ok {
			bd, err := time.Parse("2006-01-02", birthStr)
			if err == nil {
				user.BirthDate = &bd
			}
		}
	}
	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetTags(c *gin.Context) {
	tags, err := h.interestRepo.AllTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func (h *UserHandler) SaveInterests(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	var req struct {
		TagIDs []int `json:"tag_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.interestRepo.SaveUserInterests(c.Request.Context(), uid, req.TagIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) GetPersonalityQuestions(c *gin.Context) {
	qs, err := h.interestRepo.QuestionsWithOptions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, qs)
}

func (h *UserHandler) SubmitPersonality(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	var req model.PersonalitySubmitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.interestRepo.SavePersonalityAnswers(c.Request.Context(), uid, req.Answers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) GetPersonality(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	dims, err := h.interestRepo.GetUserPersonality(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dims)
}

func (h *UserHandler) GetAvatarComponents(c *gin.Context) {
	comps, err := h.interestRepo.AvatarComponents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comps)
}

func (h *UserHandler) GetPersonalityReport(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	dims, err := h.interestRepo.GetUserPersonality(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(dims) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "personality quiz not completed"})
		return
	}
	report := h.personalitySvc.Generate(dims)
	c.JSON(http.StatusOK, report)
}

func (h *UserHandler) GetHoroscope(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	user, err := h.userRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	zodiac := zodiacFromBirth(user.BirthDate)
	if zodiac == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "birth date not set"})
		return
	}
	dims, _ := h.interestRepo.GetUserPersonality(c.Request.Context(), uid)
	c.JSON(http.StatusOK, gin.H{
		"zodiac":    zodiac,
		"horoscope": h.horoscopeSvc.Daily(zodiac, dims),
	})
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	var req model.DeleteAccountReq
	_ = c.ShouldBindJSON(&req)
	if err := h.userRepo.RequestDelete(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "cooldown_days": 7})
}

func (h *UserHandler) RestoreAccount(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	if err := h.userRepo.RestoreAccount(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type reportUserReq struct {
	TargetUserID string `json:"target_user_id" binding:"required"`
	Reason       string `json:"reason"`
}

func (h *UserHandler) ReportUser(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)

	var req reportUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_user_id required"})
		return
	}
	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_user_id"})
		return
	}

	report := &model.UserReport{
		ReporterID:   uid,
		TargetUserID: targetID,
		Reason:       req.Reason,
	}
	if err := h.userRepo.SaveReport(c.Request.Context(), report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *UserHandler) SaveDeviceToken(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := uuid.Parse(userID)
	var req model.DeviceToken
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.UserID = uid
	if err := h.userRepo.SaveDeviceToken(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func zodiacFromBirth(birth *time.Time) string {
	if birth == nil {
		return ""
	}
	m := birth.Month()
	d := birth.Day()
	switch {
	case (m == 1 && d >= 20) || (m == 2 && d <= 18):
		return "水瓶座"
	case (m == 2 && d >= 19) || (m == 3 && d <= 20):
		return "双鱼座"
	case (m == 3 && d >= 21) || (m == 4 && d <= 19):
		return "白羊座"
	case (m == 4 && d >= 20) || (m == 5 && d <= 20):
		return "金牛座"
	case (m == 5 && d >= 21) || (m == 6 && d <= 21):
		return "双子座"
	case (m == 6 && d >= 22) || (m == 7 && d <= 22):
		return "巨蟹座"
	case (m == 7 && d >= 23) || (m == 8 && d <= 22):
		return "狮子座"
	case (m == 8 && d >= 23) || (m == 9 && d <= 22):
		return "处女座"
	case (m == 9 && d >= 23) || (m == 10 && d <= 23):
		return "天秤座"
	case (m == 10 && d >= 24) || (m == 11 && d <= 22):
		return "天蝎座"
	case (m == 11 && d >= 23) || (m == 12 && d <= 21):
		return "射手座"
	default:
		return "摩羯座"
	}
}
