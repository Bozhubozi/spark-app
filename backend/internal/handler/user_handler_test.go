package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupUserRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	dsn := "host=localhost port=5432 user=spark password=spark123 dbname=spark sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}

	userRepo := repository.NewUserRepo(db)
	interestRepo := repository.NewInterestRepo(db)
	prs := service.NewPersonalityReportService(nil)
	zodiacSvc := service.NewZodiacService()
	horoscopeSvc := service.NewHoroscopeService()

	userH := NewUserHandler(userRepo, interestRepo, prs, zodiacSvc, horoscopeSvc)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if len(header) > 7 && header[:7] == "Bearer " {
			claims := &jwt.RegisteredClaims{}
			token, _ := jwt.ParseWithClaims(header[7:], claims, func(t *jwt.Token) (interface{}, error) {
				return []byte("user-test-secret"), nil
			})
			if token != nil && token.Valid {
				c.Set("user_id", claims.Subject)
			}
		}
		c.Next()
	})

	api.GET("/user/profile", userH.GetProfile)
	api.PUT("/user/profile", userH.UpdateProfile)
	api.GET("/user/tags", userH.GetTags)
	api.GET("/user/personality/questions", userH.GetPersonalityQuestions)
	api.GET("/user/horoscope", userH.GetHoroscope)
	api.GET("/user/avatars", userH.GetAvatarComponents)

	return r, db
}

func userToken(uid uuid.UUID) string {
	claims := jwt.RegisteredClaims{Subject: uid.String()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("user-test-secret"))
	return s
}

func createUserTestUser(t *testing.T, db *gorm.DB, nick string) uuid.UUID {
	t.Helper()
	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	phone := "utest-" + uid.String()[:6]
	user := model.User{
		ID: uid, Nickname: nick, Phone: &phone,
		PasswordHash: string(hash), Gender: 1,
	}
	db.Create(&user)
	t.Cleanup(func() {
		db.Where("user_id = ?", uid).Delete(&model.UserInterest{})
		db.Where("user_id = ?", uid).Delete(&model.UserPersonalityAnswer{})
		db.Delete(&user)
	})
	return uid
}

func TestGetProfile(t *testing.T) {
	r, db := setupUserRouter(t)
	uid := createUserTestUser(t, db, "profiletest")

	req := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(uid))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetProfile: %d %s", w.Code, w.Body.String())
	}

	var user model.User
	json.Unmarshal(w.Body.Bytes(), &user)
	if user.Nickname != "profiletest" {
		t.Errorf("nickname: %q", user.Nickname)
	}
}

func TestUpdateProfile(t *testing.T) {
	r, db := setupUserRouter(t)
	uid := createUserTestUser(t, db, "updatetest")

	bio := "hello world"
	city := "深圳"
	body, _ := json.Marshal(gin.H{"bio": bio, "city": city})
	req := httptest.NewRequest("PUT", "/api/v1/user/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken(uid))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("UpdateProfile: %d %s", w.Code, w.Body.String())
	}

	// Verify update persisted
	req2 := httptest.NewRequest("GET", "/api/v1/user/profile", nil)
	req2.Header.Set("Authorization", "Bearer "+userToken(uid))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var user model.User
	json.Unmarshal(w2.Body.Bytes(), &user)
	if user.Bio == nil || *user.Bio != bio {
		t.Errorf("bio: got %v, want %q", user.Bio, bio)
	}
	if user.City == nil || *user.City != city {
		t.Errorf("city: got %v, want %q", user.City, city)
	}
}

func TestGetPersonalityQuestions(t *testing.T) {
	r, db := setupUserRouter(t)
	uid := createUserTestUser(t, db, "pqtest")

	req := httptest.NewRequest("GET", "/api/v1/user/personality/questions", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(uid))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetQuestions: %d %s", w.Code, w.Body.String())
	}

	var questions []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &questions)
	if len(questions) != 10 {
		t.Errorf("expected 10 questions, got %d", len(questions))
	}
}

func TestGetHoroscope(t *testing.T) {
	r, db := setupUserRouter(t)
	uid := createUserTestUser(t, db, "horotest")

	req := httptest.NewRequest("GET", "/api/v1/user/horoscope", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(uid))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Horoscope needs birth date — without it, may return error or fallback
	// Just verify it doesn't crash
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("unexpected status: %d", w.Code)
	}
}

func TestGetAvatarComponents(t *testing.T) {
	r, db := setupUserRouter(t)
	uid := createUserTestUser(t, db, "avatartest")

	req := httptest.NewRequest("GET", "/api/v1/user/avatars", nil)
	req.Header.Set("Authorization", "Bearer "+userToken(uid))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GetAvatars: %d %s", w.Code, w.Body.String())
	}
	var avatars []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &avatars)
	if len(avatars) != 16 {
		t.Errorf("expected 16 avatar components, got %d", len(avatars))
	}
}
