package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupAuthService(t *testing.T) (*AuthService, *gorm.DB) {
	t.Helper()
	dsn := "host=localhost port=5432 user=spark password=spark123 dbname=spark sslmode=disable TimeZone=Asia/Shanghai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("DB not available: %v", err)
	}
	cfg := &config.Config{JWTSecret: "test-secret", ServerPort: "8080"}
	userRepo := repository.NewUserRepo(db)
	return NewAuthService(cfg, userRepo), db
}

func TestAuthRegister(t *testing.T) {
	svc, db := setupAuthService(t)
	ctx := t.Context()

	req := &model.UserRegisterReq{
		Phone:    "auth-test-001",
		Password: "testpass123",
		Nickname: "authtest1",
	}

	t.Cleanup(func() {
		db.Where("phone = ?", req.Phone).Delete(&model.User{})
	})

	resp, err := svc.Register(ctx, req)
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("token should not be empty")
	}
	if resp.User.ID == uuid.Nil {
		t.Fatal("user ID should not be nil")
	}
	if resp.User.Nickname != req.Nickname {
		t.Errorf("nickname: got %q, want %q", resp.User.Nickname, req.Nickname)
	}

	// Duplicate registration should fail
	_, err = svc.Register(ctx, req)
	if err == nil {
		t.Fatal("duplicate register should fail")
	}

	// Password should be hashed
	var user model.User
	db.Where("phone = ?", req.Phone).First(&user)
	if user.PasswordHash == req.Password {
		t.Fatal("password should be hashed")
	}
}

func TestAuthLogin(t *testing.T) {
	svc, db := setupAuthService(t)
	ctx := t.Context()

	phone := "auth-test-002"
	password := "correct-password"

	t.Cleanup(func() {
		db.Where("phone = ?", phone).Delete(&model.User{})
	})

	// Register
	_, err := svc.Register(ctx, &model.UserRegisterReq{
		Phone: phone, Password: password, Nickname: "logintest",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Login with phone
	resp, err := svc.Login(ctx, &model.UserLoginReq{
		Account: phone, Password: password,
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("token should not be empty")
	}

	// Wrong password
	_, err = svc.Login(ctx, &model.UserLoginReq{
		Account: phone, Password: "wrong",
	})
	if err == nil {
		t.Fatal("wrong password should fail")
	}

	// Non-existent account
	_, err = svc.Login(ctx, &model.UserLoginReq{
		Account: "nonexistent-55555", Password: "anything",
	})
	if err == nil {
		t.Fatal("non-existent account should fail")
	}
}

func TestAuthValidateToken(t *testing.T) {
	svc, db := setupAuthService(t)
	ctx := t.Context()

	phone := "auth-test-003"
	t.Cleanup(func() {
		db.Where("phone = ?", phone).Delete(&model.User{})
	})

	resp, err := svc.Register(ctx, &model.UserRegisterReq{
		Phone: phone, Password: "pass123", Nickname: "tokentest",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Valid token
	uid, err := svc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if uid != resp.User.ID {
		t.Errorf("token user ID mismatch: %s vs %s", uid, resp.User.ID)
	}

	// Invalid token
	_, err = svc.ValidateToken("not.a.valid.token")
	if err == nil {
		t.Fatal("invalid token should fail")
	}

	// Expired token
	expiredClaims := jwt.RegisteredClaims{
		Subject:   resp.User.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredStr, _ := expiredToken.SignedString([]byte("test-secret"))
	_, err = svc.ValidateToken(expiredStr)
	if err == nil {
		t.Fatal("expired token should fail")
	}
}

func TestAuthPasswordHashing(t *testing.T) {
	svc, db := setupAuthService(t)
	ctx := t.Context()

	phone := "auth-test-004"
	t.Cleanup(func() {
		db.Where("phone = ?", phone).Delete(&model.User{})
	})

	_, err := svc.Register(ctx, &model.UserRegisterReq{
		Phone: phone, Password: "securepassword", Nickname: "hashtest",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	var user model.User
	db.Where("phone = ?", phone).First(&user)

	// Verify hash works
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("securepassword")); err != nil {
		t.Fatal("password hash mismatch")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("wrongpassword")) == nil {
		t.Fatal("wrong password should not match hash")
	}
}
