package delivery

import (
	"context"
	"net/http"

	"task-management/internal/domain"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
	ParseToken(token string) (int64, error)
}

type AuthHandler struct {
	auth AuthService
}

func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{ID: user.ID, Email: user.Email})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	token, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{AccessToken: token})
}
