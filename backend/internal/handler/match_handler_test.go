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
	"github.com/redis/go-redis/v9"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMatchRouter(t *testing.T) (*gin.Engine, *gorm.DB, *redis.Client) {
	t.Helper()
	dsn := "host=localhost port=5432 user=spark password=spark123 dbname=spark sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if err := rdb.Ping(t.Context()).Err(); err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	cfg := &config.Config{JWTSecret: "test-secret"}
	userRepo := repository.NewUserRepo(db)
	matchRepo := repository.NewMatchRepo(db)
	chatRepo := repository.NewChatRepo(db)
	interestRepo := repository.NewInterestRepo(db)

	authSvc := service.NewAuthService(cfg, userRepo)
	matchSvc := service.NewMatchService(matchRepo, interestRepo, userRepo, rdb)
	chatSvc := service.NewChatService(chatRepo, userRepo)
	wsHub := service.NewWSHub(rdb)
	zodiacSvc := service.NewZodiacService()
	icebreakerSvc := service.NewIcebreakerService(zodiacSvc)

	authH := NewAuthHandler(authSvc)
	matchH := NewMatchHandler(matchSvc, wsHub, userRepo, zodiacSvc, icebreakerSvc, interestRepo, chatRepo)
	chatH := NewChatHandler(chatSvc, chatRepo, wsHub)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Auth routes (public)
	r.POST("/api/v1/auth/register", authH.Register)

	// Protected routes
	auth := r.Group("/api/v1")
	auth.Use(func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:]
		}
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		if err == nil && token.Valid {
			c.Set("user_id", claims.Subject)
		}
		c.Next()
	})
	auth.GET("/match/candidates", matchH.GetCandidates)
	auth.POST("/match/swipe", matchH.Swipe)
	auth.GET("/chat/rooms", chatH.GetRooms)

	return r, db, rdb
}

func makeToken(uid uuid.UUID) string {
	claims := jwt.RegisteredClaims{Subject: uid.String()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("test-secret"))
	return s
}

func createMatchTestUser(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	uid := uuid.New()
	hash, _ := bcrypt.GenerateFromPassword([]byte("test"), bcrypt.DefaultCost)
	user := model.User{
		ID: uid, Nickname: "mtest_" + uid.String()[:6],
		PasswordHash: string(hash), Gender: 1,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}
	t.Cleanup(func() {
		db.Where("user_id_1 = ? OR user_id_2 = ?", uid, uid).Delete(&model.ChatRoom{})
		db.Where("user_id_1 = ? OR user_id_2 = ?", uid, uid).Delete(&model.Match{})
		db.Delete(&user)
	})
	return uid
}

func TestSwipeMatchCreatesRoom(t *testing.T) {
	r, db, rdb := setupMatchRouter(t)

	userA := createMatchTestUser(t, db)
	userB := createMatchTestUser(t, db)

	// User B likes User A first (creates pending match)
	matchRepo := repository.NewMatchRepo(db)
	m := &model.Match{
		ID:      uuid.New(),
		UserID1: userB,
		UserID2: userA,
		Status:  model.MatchStatusPending,
	}
	if err := matchRepo.Create(t.Context(), m); err != nil {
		t.Fatalf("create pending match: %v", err)
	}

	// User A swipes right on User B → mutual match
	body, _ := json.Marshal(gin.H{"target_user_id": userB.String(), "direction": "like"})
	req := httptest.NewRequest("POST", "/api/v1/match/swipe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeToken(userA))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("swipe: status %d, body: %s", w.Code, w.Body.String())
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["matched"] != true {
		t.Fatal("expected matched=true")
	}
	if result["room_id"] == nil {
		t.Fatal("expected room_id in match response")
	}

	// Verify User A can see the chat room
	req2 := httptest.NewRequest("GET", "/api/v1/chat/rooms", nil)
	req2.Header.Set("Authorization", "Bearer "+makeToken(userA))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("get rooms: %d", w2.Code)
	}
	var rooms []map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &rooms)
	if len(rooms) == 0 {
		t.Fatal("expected at least 1 chat room after match")
	}

	// Verify User B can also see it
	req3 := httptest.NewRequest("GET", "/api/v1/chat/rooms", nil)
	req3.Header.Set("Authorization", "Bearer "+makeToken(userB))
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)
	var roomsB []map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &roomsB)
	if len(roomsB) == 0 {
		t.Fatal("User B should also see the chat room")
	}

	// Cleanup
	rdb.Del(t.Context(), "candidates:seen:*")
	t.Cleanup(func() { db.Where("id = ?", m.ID).Delete(&model.Match{}) })
}

func TestSwipePassNoMatch(t *testing.T) {
	r, db, _ := setupMatchRouter(t)

	userA := createMatchTestUser(t, db)
	userB := createMatchTestUser(t, db)

	body, _ := json.Marshal(gin.H{"target_user_id": userB.String(), "direction": "pass"})
	req := httptest.NewRequest("POST", "/api/v1/match/swipe", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeToken(userA))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["matched"] == true {
		t.Fatal("pass should not trigger a match")
	}
}
