package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spark-app/backend/internal/config"
	"github.com/spark-app/backend/internal/model"
	"github.com/spark-app/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	cfg      *config.Config
	userRepo *repository.UserRepo
}

func NewAuthService(cfg *config.Config, userRepo *repository.UserRepo) *AuthService {
	return &AuthService{cfg: cfg, userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, req *model.UserRegisterReq) (*model.UserLoginResp, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &model.User{
		PasswordHash: string(hash),
		Nickname:     req.Nickname,
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}
	if req.Email != "" {
		user.Email = &req.Email
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.UserLoginResp{Token: token, User: *user}, nil
}

func (s *AuthService) Login(ctx context.Context, req *model.UserLoginReq) (*model.UserLoginResp, error) {
	user, err := s.userRepo.FindByAccount(ctx, req.Account)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("account not found")
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid password")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.UserLoginResp{Token: token, User: *user}, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	idStr, ok := claims["sub"].(string)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}
	return uuid.Parse(idStr)
}

func (s *AuthService) WechatLogin(ctx context.Context, code string) (*model.UserLoginResp, error) {
	openID, unionID, err := s.exchangeWechatCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("wechat code exchange: %w", err)
	}

	user, err := s.userRepo.FindByWechatOpenID(ctx, openID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find user by wechat: %w", err)
	}

	if user == nil {
		user = &model.User{
			WechatOpenID:  &openID,
			WechatUnionID: &unionID,
			Nickname:      fmt.Sprintf("用户%s", openID[:8]),
			PasswordHash:  "",
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create wechat user: %w", err)
		}
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &model.UserLoginResp{Token: token, User: *user}, nil
}

func (s *AuthService) exchangeWechatCode(ctx context.Context, code string) (string, string, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		s.cfg.WechatAppID, s.cfg.WechatSecret, code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("wechat api call: %w", err)
	}
	defer resp.Body.Close()

	var result model.WechatTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("decode wechat resp: %w", err)
	}

	if result.ErrCode != 0 {
		return "", "", fmt.Errorf("wechat error: %s (code=%d)", result.ErrMsg, result.ErrCode)
	}

	return result.OpenID, result.UnionID, nil
}

func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(72 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
