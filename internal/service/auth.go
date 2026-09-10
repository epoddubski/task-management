package service

import (
	"context"
	"fmt"
	"task-management/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type AuthService struct {
	users     UserRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAuthService(users UserRepository, secret string, ttl time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: secret,
		tokenTTL:  ttl,
	}
}

type claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

func (u *AuthService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	if email == "" || password == "" {
		return nil, domain.ErrInvalidCreds
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
	}

	err = u.users.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:        user.ID,
		Email:     email,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (u *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.users.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCreds
	}

	now := time.Now()
	c := claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(u.tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString([]byte(u.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (u *AuthService) ParseToken(tokenString string) (int64, error) {
	c := &claims{}

	token, err := jwt.ParseWithClaims(tokenString, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.ErrInvalidToken
		}
		return []byte(u.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return 0, domain.ErrInvalidToken
	}

	return c.UserID, nil
}
