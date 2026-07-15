package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ogarridojimenez/securevault/internal/model"
	"github.com/ogarridojimenez/securevault/internal/repository"
	"golang.org/x/crypto/argon2"
)

type AuthService struct {
	userRepo       *repository.UserRepository
	refreshRepo    *repository.RefreshTokenRepository
	jwtSecret      []byte
	accessTokenTTL time.Duration
	refreshTTL     time.Duration
}

type AuthConfig struct {
	JWTSecret      string
	AccessTokenTTL time.Duration
	RefreshTTL     time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, refreshRepo *repository.RefreshTokenRepository, cfg AuthConfig) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		refreshRepo:    refreshRepo,
		jwtSecret:      []byte(cfg.JWTSecret),
		accessTokenTTL: cfg.AccessTokenTTL,
		refreshTTL:     cfg.RefreshTTL,
	}
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*model.User, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("%w: email and password are required", model.ErrInvalidInput)
	}
	if len(password) < 8 {
		return nil, fmt.Errorf("%w: password must be at least 8 characters", model.ErrInvalidInput)
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	encoded := fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=4$%s$%s",
		hex.EncodeToString(salt), hex.EncodeToString(hash))

	return s.userRepo.Create(ctx, email, encoded)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*model.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid credentials", model.ErrUnauthorized)
	}

	if !s.verifyPassword(user.PasswordHash, password) {
		return nil, fmt.Errorf("%w: invalid credentials", model.ErrUnauthorized)
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       user.ID,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshTokenStr string) (*model.AuthResponse, error) {
	tokenHash := s.hashToken(refreshTokenStr)

	rt, err := s.refreshRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid refresh token", model.ErrUnauthorized)
	}
	if rt.Revoked {
		return nil, fmt.Errorf("%w: refresh token revoked", model.ErrUnauthorized)
	}
	if time.Now().After(rt.ExpiresAt) {
		return nil, fmt.Errorf("%w: refresh token expired", model.ErrTokenExpired)
	}

	// Revoke old refresh token (rotation)
	if err := s.refreshRepo.Revoke(ctx, rt.ID); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.generateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserID:       user.ID,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshTokenStr string) error {
	tokenHash := s.hashToken(refreshTokenStr)
	rt, err := s.refreshRepo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil // Silently succeed even if token not found
	}
	return s.refreshRepo.Revoke(ctx, rt.ID)
}

func (s *AuthService) ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", model.ErrUnauthorized, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%w: invalid token claims", model.ErrUnauthorized)
	}

	return claims, nil
}

func (s *AuthService) verifyPassword(hash, password string) bool {
	var salt, key []byte
	n, err := fmt.Sscanf(hash, "$argon2id$v=19$m=65536,t=3,p=4$%x$%x", &salt, &key)
	if err != nil || n != 2 {
		return false
	}
	computed := argon2.IDKey([]byte(password), salt, 3, 64*1024, 4, 32)
	return subtle.ConstantTimeCompare(key, computed) == 1
}

func (s *AuthService) generateAccessToken(userID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "securevault",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) generateRefreshToken(ctx context.Context, userID string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate refresh token bytes: %w", err)
	}
	refreshToken := hex.EncodeToString(tokenBytes)
	tokenHash := s.hashToken(refreshToken)

	_, err := s.refreshRepo.Create(ctx, userID, tokenHash, time.Now().Add(s.refreshTTL))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *AuthService) hashToken(token string) string {
	h := sha256Hash([]byte(token))
	return hex.EncodeToString(h)
}
