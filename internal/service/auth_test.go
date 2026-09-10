package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"task-management/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

func hash(pass string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	return string(hash)
}

func TestAuthServiceRegister(t *testing.T) {
	t.Parallel()

	repoWithUser := newFakeUserRepo()
	repoWithUser.Create(context.Background(), &domain.User{Email: "duplicate@example.com", PasswordHash: "hash"})

	cases := []struct {
		name     string
		repo     *fakeUserRepo
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "success",
			repo:     newFakeUserRepo(),
			email:    "user@example.com",
			password: "password123",
		},
		{
			name:     "empty email",
			repo:     newFakeUserRepo(),
			email:    "",
			password: "password123",
			wantErr:  domain.ErrInvalidCreds,
		},
		{
			name:     "empty password",
			repo:     newFakeUserRepo(),
			email:    "user@example.com",
			password: "",
			wantErr:  domain.ErrInvalidCreds,
		},
		{
			name:     "password too long",
			repo:     newFakeUserRepo(),
			email:    "user@example.com",
			password: string(make([]byte, 73)),
			wantErr:  bcrypt.ErrPasswordTooLong,
		},
		{
			name:     "rejects a duplicate email",
			repo:     repoWithUser,
			email:    "duplicate@example.com",
			password: "password123",
			wantErr:  domain.ErrUserExists,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := NewAuthService(c.repo, "secret", 10*time.Minute)

			u, err := s.Register(context.Background(), c.email, c.password)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error %v, got %v", c.wantErr, err)
			}

			if c.wantErr != nil {
				return
			}

			if u == nil {
				t.Fatal("expected user to be not nil")
			}

			if u.ID == 0 {
				t.Errorf("expected user ID to be assigned, got 0")
			}

			if u.Email != c.email {
				t.Errorf("expected normalized email, got %q", u.Email)
			}

			if u.PasswordHash == c.password {
				t.Errorf("password should be hashed, not stored raw")
			}
		})
	}
}

func TestAuthServiceLogin(t *testing.T) {
	t.Parallel()

	repo := newFakeUserRepo()

	_ = repo.Create(context.Background(), &domain.User{
		Email:        "user@example.com",
		PasswordHash: hash("password123"),
	})

	cases := []struct {
		name     string
		email    string
		password string
		wantErr  error
	}{
		{
			name:     "success",
			email:    "user@example.com",
			password: "password123",
		},
		{
			name:     "unknown email",
			email:    "ghost@example.com",
			password: "password123",
			wantErr:  domain.ErrInvalidCreds,
		},
		{
			name:     "wrong password",
			email:    "user@example.com",
			password: "wrongpassword",
			wantErr:  domain.ErrInvalidCreds,
		},
		{
			name:     "empty password",
			email:    "user@example.com",
			password: "",
			wantErr:  domain.ErrInvalidCreds,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := NewAuthService(repo, "secret", 10*time.Minute)

			token, err := s.Login(context.Background(), c.email, c.password)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error %v, got %v", c.wantErr, err)
			}

			if c.wantErr != nil {
				return
			}

			if token == "" {
				t.Fatal("expected a non-empty token")
			}
		})
	}
}

func TestAuthServiceParseToken(t *testing.T) {
	t.Parallel()

	repo := newFakeUserRepo()
	s := NewAuthService(repo, "secret", 10*time.Minute)

	user, err := s.Register(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("test setup: register failed: %v", err)
	}

	token, err := s.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("test setup: login failed: %v", err)
	}

	superService := NewAuthService(repo, "super-secret", 10*time.Minute)
	superToken, err := superService.Login(context.Background(), "user@example.com", "password123")
	if err != nil {
		t.Fatalf("test setup: super login failed: %v", err)
	}

	cases := []struct {
		name    string
		token   string
		wantID  int64
		wantErr error
	}{
		{
			name:   "valid token",
			token:  token,
			wantID: user.ID,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: domain.ErrInvalidToken,
		},
		{
			name:    "wrong token",
			token:   "wrong-jwt",
			wantErr: domain.ErrInvalidToken,
		},
		{
			name:    "different secret service",
			token:   superToken,
			wantErr: domain.ErrInvalidToken,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			id, err := s.ParseToken(c.token)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error %v, got %v", c.wantErr, err)
			}

			if c.wantErr != nil {
				if id != 0 {
					t.Errorf("expected id 0 on error, got %d", id)
				}
				return
			}

			if id != c.wantID {
				t.Errorf("expected user id %d, got %d", c.wantID, id)
			}
		})
	}
}
