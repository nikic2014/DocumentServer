package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"TestTask/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthService struct {
	users     domain.UserRepository
	tokens    domain.TokenStore
	jwtSecret []byte
	tokenTTL  time.Duration
}

func NewAuthService(users domain.UserRepository, tokens domain.TokenStore, jwtSecret string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		tokens:    tokens,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  tokenTTL,
	}
}

// ParseToken - Проверяем валидность токена
func (s *AuthService) ParseToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("Неизвестный метод подписания")
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, errors.New("Токен недействителен или просрочен")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("Невалидные поля токена")
	}

	active, err := s.tokens.Exists(ctx, claims.ID)
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, errors.New("Токен был отозван")
	}

	return claims, nil
}

func (s *AuthService) Register(ctx context.Context, login, password string) (*domain.User, error) {
	login = strings.TrimSpace(login)
	if login == "" || password == "" {
		return nil, errors.New("Логин и пароль не могут быть пустыми")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Login:        login,
		PasswordHash: string(hash),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.users.GetUserByLogin(ctx, &login)

	if err != nil {
		return "", domain.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidPassword
	}

	tokenID, err := domain.GenerateTokenID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	expiresAt := now.Add(s.tokenTTL)

	claims := Claims{
		UserID:   user.ID,
		Username: user.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	if err := s.tokens.Store(ctx, tokenID, expiresAt); err != nil {
		return "", err
	}

	return signed, nil
}

func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	claims, err := s.ParseToken(ctx, tokenString)
	if err != nil {
		return err
	}
	return s.tokens.Delete(ctx, claims.ID)
}
