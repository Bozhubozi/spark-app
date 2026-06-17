package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupAuthRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	dsn := "host=localhost port=5432 user=spark password=spark123 dbname=spark sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	cfg := &config.Config{JWTSecret: "test-secret"}
	userRepo := repository.NewUserRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo)
	authH := NewAuthHandler(authSvc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/auth/register", authH.Register)
	r.POST("/api/v1/auth/login", authH.Login)
	return r, db
}

func TestRegisterAndLogin(t *testing.T) {
	r, db := setupAuthRouter(t)
	phone := "19900199000"
	nickname := "htest_" + uuid.New().String()[:6]

	t.Cleanup(func() {
		db.Where("phone = ?", phone).Delete(&model.User{})
	})

	// Register
	body, _ := json.Marshal(gin.H{
		"phone":    phone,
		"password": "testpass123",
		"nickname": nickname,
	})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code < 200 || w.Code >= 300 {
		t.Fatalf("register: status %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil {
		t.Fatal("register should return token")
	}

	// Login (uses "account" field)
	body2, _ := json.Marshal(gin.H{"account": phone, "password": "testpass123"})
	req2 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("login: status %d, body: %s", w2.Code, w2.Body.String())
	}

	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	if resp2["token"] == nil {
		t.Fatal("login should return token")
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	r, db := setupAuthRouter(t)
	phone := "19900199001"

	t.Cleanup(func() {
		db.Where("phone = ?", phone).Delete(&model.User{})
	})

	// Register
	body, _ := json.Marshal(gin.H{"phone": phone, "password": "correct", "nickname": "test"})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Login with wrong password
	body2, _ := json.Marshal(gin.H{"phone": phone, "password": "wrongpass"})
	req2 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code == http.StatusOK {
		t.Fatal("login with wrong password should fail")
	}
}

func TestDuplicateRegister(t *testing.T) {
	r, db := setupAuthRouter(t)
	phone := "19900199002"

	t.Cleanup(func() {
		db.Where("phone = ?", phone).Delete(&model.User{})
	})

	body, _ := json.Marshal(gin.H{"phone": phone, "password": "pass123", "nickname": "user1"})
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// First register
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req)
	if w1.Code < 200 || w1.Code >= 300 {
		t.Fatalf("first register failed: %d %s", w1.Code, w1.Body.String())
	}

	// Second register with same phone
	body2, _ := json.Marshal(gin.H{"phone": phone, "password": "pass2", "nickname": "user2"})
	req2 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code < 400 {
		t.Fatalf("duplicate register should fail, got %d: %s", w2.Code, w2.Body.String())
	}
}
