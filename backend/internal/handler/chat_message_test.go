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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMessageRouter(t *testing.T) (*gin.Engine, *gorm.DB, *redis.Client) {
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

	cfg := &config.Config{JWTSecret: "msg-test-secret"}
	userRepo := repository.NewUserRepo(db)
	chatRepo := repository.NewChatRepo(db)

	authSvc := service.NewAuthService(cfg, userRepo)
	chatSvc := service.NewChatService(chatRepo, userRepo)
	wsHub := service.NewWSHub(rdb)

	chatH := NewChatHandler(chatSvc, chatRepo, wsHub)
	wsH := NewWSHandler(wsHub, chatSvc, authSvc, chatRepo, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	authMiddleware := func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if len(header) > 7 && header[:7] == "Bearer " {
			claims := &jwt.RegisteredClaims{}
			token, _ := jwt.ParseWithClaims(header[7:], claims, func(t *jwt.Token) (interface{}, error) {
				return []byte("msg-test-secret"), nil
			})
			if token != nil && token.Valid {
				c.Set("user_id", claims.Subject)
			}
		}
		c.Next()
	}

	api := r.Group("/api/v1")
	api.Use(authMiddleware)
	api.POST("/chat/rooms", chatH.GetOrCreateRoom)
	api.GET("/chat/rooms", chatH.GetRooms)
	api.GET("/chat/rooms/:room_id/messages", chatH.GetMessages)
	api.POST("/chat/rooms/:room_id/read", chatH.MarkRead)

	_ = wsH // WebSocket handler tested separately

	return r, db, rdb
}

func msgToken(uid uuid.UUID) string {
	claims := jwt.RegisteredClaims{Subject: uid.String()}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("msg-test-secret"))
	return s
}

func TestSendAndRetrieveMessages(t *testing.T) {
	r, db, _ := setupMessageRouter(t)

	alice := createMatchTestUser(t, db)
	bob := createMatchTestUser(t, db)

	// Create room
	body, _ := json.Marshal(gin.H{"target_user_id": bob.String()})
	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+msgToken(alice))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var room model.ChatRoom
	json.Unmarshal(w.Body.Bytes(), &room)
	if room.ID == uuid.Nil {
		t.Fatal("failed to create room")
	}

	// Save messages directly via repo (simulating received messages)
	chatRepo := repository.NewChatRepo(db)
	contents := []string{"Hello Bob!", "Hi Alice!", "How are you?", "I'm great!"}
	for i, content := range contents {
		senderID := alice
		if i%2 == 1 {
			senderID = bob
		}
		msg := &model.Message{
			ID:          uuid.New(),
			RoomID:      room.ID,
			SenderID:    senderID,
			ClientMsgID: uuid.New().String(),
			ContentType: model.ContentTypeText,
			Content:     content,
		}
		if err := chatRepo.SaveMessage(t.Context(), msg); err != nil {
			t.Fatalf("save message %d: %v", i, err)
		}
	}

	// Retrieve messages
	req2 := httptest.NewRequest("GET", "/api/v1/chat/rooms/"+room.ID.String()+"/messages", nil)
	req2.Header.Set("Authorization", "Bearer "+msgToken(alice))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("get messages: %d %s", w2.Code, w2.Body.String())
	}

	var msgs []model.Message
	json.Unmarshal(w2.Body.Bytes(), &msgs)
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}

	// Messages should be in chronological order
	for i, msg := range msgs {
		if msg.Content != contents[i] {
			t.Errorf("msg[%d]: got %q, want %q", i, msg.Content, contents[i])
		}
	}
	t.Logf("Retrieved %d messages in order", len(msgs))
}

func TestMarkReadMessages(t *testing.T) {
	r, db, _ := setupMessageRouter(t)

	alice := createMatchTestUser(t, db)
	bob := createMatchTestUser(t, db)

	// Create room
	body, _ := json.Marshal(gin.H{"target_user_id": bob.String()})
	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+msgToken(alice))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var room model.ChatRoom
	json.Unmarshal(w.Body.Bytes(), &room)

	// Bob sends a message to Alice
	chatRepo := repository.NewChatRepo(db)
	msg := &model.Message{
		ID:          uuid.New(),
		RoomID:      room.ID,
		SenderID:    bob,
		ClientMsgID: uuid.New().String(),
		ContentType: model.ContentTypeText,
		Content:     "Hey Alice!",
	}
	chatRepo.SaveMessage(t.Context(), msg)

	// Alice marks as read
	req2 := httptest.NewRequest("POST", "/api/v1/chat/rooms/"+room.ID.String()+"/read", nil)
	req2.Header.Set("Authorization", "Bearer "+msgToken(alice))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("mark read: %d %s", w2.Code, w2.Body.String())
	}

	// Verify message is marked as read
	var updated model.Message
	db.First(&updated, "id = ?", msg.ID)
	if !updated.IsRead {
		t.Error("message should be marked as read")
	}
}

func TestEmptyMessages(t *testing.T) {
	r, db, _ := setupMessageRouter(t)

	alice := createMatchTestUser(t, db)
	bob := createMatchTestUser(t, db)

	body, _ := json.Marshal(gin.H{"target_user_id": bob.String()})
	req := httptest.NewRequest("POST", "/api/v1/chat/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+msgToken(alice))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var room model.ChatRoom
	json.Unmarshal(w.Body.Bytes(), &room)

	// Get messages for empty room
	req2 := httptest.NewRequest("GET", "/api/v1/chat/rooms/"+room.ID.String()+"/messages", nil)
	req2.Header.Set("Authorization", "Bearer "+msgToken(alice))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("get empty messages: %d", w2.Code)
	}

	var msgs []model.Message
	json.Unmarshal(w2.Body.Bytes(), &msgs)
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}
