package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/repository"
	"github.com/spark-app/backend/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupAuthMiddleware(t *testing.T) (*gin.Engine, *service.AuthService) {
	t.Helper()
	dsn := "host=localhost port=5432 user=spark password=spark123 dbname=spark sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	cfg := &config.Config{JWTSecret: "test-secret"}
	userRepo := repository.NewUserRepo(db)
	authSvc := service.NewAuthService(cfg, userRepo)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(AuthRequired(authSvc))
	r.GET("/api/v1/test", func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		c.JSON(200, gin.H{"user_id": uid})
	})
	return r, authSvc
}

func makeValidToken(uid uuid.UUID) string {
	claims := jwt.RegisteredClaims{
		Subject:   uid.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte("test-secret"))
	return s
}

func TestAuthRequiredValidToken(t *testing.T) {
	r, _ := setupAuthMiddleware(t)
	uid := uuid.New()

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+makeValidToken(uid))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthRequiredNoToken(t *testing.T) {
	r, _ := setupAuthMiddleware(t)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthRequiredInvalidToken(t *testing.T) {
	r, _ := setupAuthMiddleware(t)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthRequiredMalformedHeader(t *testing.T) {
	r, _ := setupAuthMiddleware(t)

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	req.Header.Set("Authorization", "NoBearerPrefix")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
