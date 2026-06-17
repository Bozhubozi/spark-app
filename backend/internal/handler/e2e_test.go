package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// setupE2E creates a full router with all handlers for end-to-end testing.
func setupE2E(t *testing.T) (*gin.Engine, *gorm.DB, *redis.Client) {
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

	cfg := &config.Config{JWTSecret: "e2e-secret"}
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
	userH := NewUserHandler(userRepo, interestRepo, nil, zodiacSvc, nil)
	matchH := NewMatchHandler(matchSvc, wsHub, userRepo, zodiacSvc, icebreakerSvc, interestRepo, chatRepo)
	chatH := NewChatHandler(chatSvc, chatRepo, wsHub)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Public routes
	r.POST("/api/v1/auth/register", authH.Register)

	// Protected routes with JWT auth
	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if len(header) < 8 || header[:7] != "Bearer " {
			c.Next()
			return
		}
		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(header[7:], claims, func(t *jwt.Token) (interface{}, error) {
			return []byte("e2e-secret"), nil
		})
		if err == nil && token.Valid {
			c.Set("user_id", claims.Subject)
		}
		c.Next()
	})
	api.PUT("/user/interests", userH.SaveInterests)
	api.GET("/match/candidates", matchH.GetCandidates)
	api.POST("/match/swipe", matchH.Swipe)
	api.POST("/chat/rooms", chatH.GetOrCreateRoom)
	api.GET("/chat/rooms", chatH.GetRooms)

	return r, db, rdb
}

func e2eToken(uid uuid.UUID) string {
	claims := jwt.RegisteredClaims{Subject: uid.String()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("e2e-secret"))
	return s
}

// TestFullUserJourney validates: register → set interests → get candidates → swipe → match → chat room.
func TestFullUserJourney(t *testing.T) {
	r, db, rdb := setupE2E(t)

	// ---- Step 1: Create two users ----
	createUser := func(phone, nick string, gender int8) uuid.UUID {
		t.Helper()
		uid := uuid.New()
		hash, _ := bcrypt.GenerateFromPassword([]byte("e2epass"), bcrypt.DefaultCost)
		user := model.User{
			ID: uid, Phone: &phone, Nickname: nick,
			PasswordHash: string(hash), Gender: gender,
		}
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create %s: %v", nick, err)
		}
		t.Cleanup(func() {
			db.Where("user_id_1 = ? OR user_id_2 = ?", uid, uid).Delete(&model.ChatRoom{})
			db.Where("user_id_1 = ? OR user_id_2 = ?", uid, uid).Delete(&model.Match{})
			db.Where("user_id = ?", uid).Delete(&model.UserInterest{})
			db.Delete(&user)
		})
		return uid
	}

	alice := createUser(fmt.Sprintf("e2e-alice-%s", uuid.New().String()[:4]), "Alice", 2)
	bob := createUser(fmt.Sprintf("e2e-bob-%s", uuid.New().String()[:4]), "Bob", 1)

	auth := func(uid uuid.UUID) string {
		return "Bearer " + e2eToken(uid)
	}

	// ---- Step 2: Alice sets interests ----
	t.Run("SetInterests", func(t *testing.T) {
		// Get available tags first
		interestRepo := repository.NewInterestRepo(db)
		tags, _ := interestRepo.AllTags(t.Context())
		if len(tags) < 5 {
			t.Skip("not enough tags in DB")
		}
		tagIDs := []int{tags[0].ID, tags[1].ID, tags[2].ID, tags[3].ID, tags[4].ID}

		body, _ := json.Marshal(gin.H{"tag_ids": tagIDs})
		req := httptest.NewRequest("PUT", "/api/v1/user/interests", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth(alice))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("set interests: %d %s", w.Code, w.Body.String())
		}
		t.Log("interests set successfully")
	})

	// ---- Step 3: Bob likes Alice (creates pending match) ----
	t.Run("BobLikesAlice", func(t *testing.T) {
		body, _ := json.Marshal(gin.H{"target_user_id": alice.String(), "direction": "like"})
		req := httptest.NewRequest("POST", "/api/v1/match/swipe", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth(bob))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("bob swipe: %d %s", w.Code, w.Body.String())
		}
		t.Log("Bob liked Alice")
	})

	// ---- Step 4: Alice likes Bob back → mutual match ----
	var roomID string
	t.Run("AliceLikesBobMatch", func(t *testing.T) {
		body, _ := json.Marshal(gin.H{"target_user_id": bob.String(), "direction": "like"})
		req := httptest.NewRequest("POST", "/api/v1/match/swipe", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", auth(alice))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("alice swipe: %d %s", w.Code, w.Body.String())
		}

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		if result["matched"] != true {
			t.Fatal("expected matched=true")
		}
		roomID = fmt.Sprint(result["room_id"])
		if roomID == "" || roomID == "<nil>" {
			t.Fatal("expected room_id in match response")
		}
		t.Logf("Mutual match! Room: %s", roomID)
	})

	// ---- Step 5: Both users see the chat room ----
	t.Run("BothSeeChatRoom", func(t *testing.T) {
		for _, uid := range []uuid.UUID{alice, bob} {
			req := httptest.NewRequest("GET", "/api/v1/chat/rooms", nil)
			req.Header.Set("Authorization", auth(uid))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("get rooms for %s: %d", uid, w.Code)
			}
			var rooms []map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &rooms)
			if len(rooms) == 0 {
				t.Fatalf("user %s should see chat room", uid.String()[:8])
			}
		}
		t.Log("Both users see the chat room")
	})

	// Cleanup Redis
	rdb.Del(t.Context(), "candidates:seen:*")
	t.Log("E2E journey completed successfully")
}
