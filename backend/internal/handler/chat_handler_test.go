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
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupChatHandlerRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	dsn := "host=localhost port=5432 user=spark password=spark123 dbname=spark sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	userRepo := repository.NewUserRepo(db)
	chatRepo := repository.NewChatRepo(db)
	chatSvc := service.NewChatService(chatRepo, userRepo)
	wsHub := service.NewWSHub(rdb)
	chatH := NewChatHandler(chatSvc, chatRepo, wsHub)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	api.Use(func(c *gin.Context) {
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
	api.POST("/chat/rooms", chatH.GetOrCreateRoom)
	api.GET("/chat/rooms", chatH.GetRooms)
	api.GET("/chat/rooms/:room_id/messages", chatH.GetMessages)
	api.POST("/chat/rooms/:room_id/read", chatH.MarkRead)

	return r, db
}

func makeChatToken(uid uuid.UUID) string {
	claims := jwt.RegisteredClaims{Subject: uid.String()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("test-secret"))
	return s
}

func TestGetOrCreateRoom(t *testing.T) {
	r, db := setupChatHandlerRouter(t)

	userA := createMatchTestUser(t, db)
	userB := createMatchTestUser(t, db)

	// Create room for the first time
	body, _ := json.Marshal(gin.H{"target_user_id": userB.String()})
	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeChatToken(userA))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("create room: status %d, body: %s", w.Code, w.Body.String())
	}

	var room model.ChatRoom
	json.Unmarshal(w.Body.Bytes(), &room)
	if room.ID == uuid.Nil {
		t.Fatal("expected room ID")
	}

	// Second call should return the same room
	req2 := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+makeChatToken(userA))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var room2 model.ChatRoom
	json.Unmarshal(w2.Body.Bytes(), &room2)
	if room2.ID != room.ID {
		t.Error("second call should return same room")
	}
}

func TestGetRooms(t *testing.T) {
	r, db := setupChatHandlerRouter(t)

	userA := createMatchTestUser(t, db)
	userB := createMatchTestUser(t, db)
	userC := createMatchTestUser(t, db)

	// Create 2 rooms
	for _, target := range []uuid.UUID{userB, userC} {
		body, _ := json.Marshal(gin.H{"target_user_id": target.String()})
		req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+makeChatToken(userA))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("create room: %d", w.Code)
		}
	}

	// Get rooms for user A
	req := httptest.NewRequest("GET", "/api/v1/chat/rooms", nil)
	req.Header.Set("Authorization", "Bearer "+makeChatToken(userA))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get rooms: %d", w.Code)
	}
	var rooms []model.ChatRoom
	json.Unmarshal(w.Body.Bytes(), &rooms)
	if len(rooms) < 2 {
		t.Errorf("expected at least 2 rooms, got %d", len(rooms))
	}
}

func TestMarkRead(t *testing.T) {
	r, db := setupChatHandlerRouter(t)

	userA := createMatchTestUser(t, db)
	userB := createMatchTestUser(t, db)

	// Create room
	body, _ := json.Marshal(gin.H{"target_user_id": userB.String()})
	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+makeChatToken(userA))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var room model.ChatRoom
	json.Unmarshal(w.Body.Bytes(), &room)

	// Mark read should not error even with no messages
	req2 := httptest.NewRequest("POST", "/api/v1/chat/rooms/"+room.ID.String()+"/read", nil)
	req2.Header.Set("Authorization", "Bearer "+makeChatToken(userA))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("mark read: %d, body: %s", w2.Code, w2.Body.String())
	}
}
