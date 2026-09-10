package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")

	ErrInvalidCreds = errors.New("invalid email or password")
	ErrInvalidToken = errors.New("invalid or expired token")
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}
