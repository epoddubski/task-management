package delivery

import (
	"errors"
	"time"

	"task-management/internal/domain"
)

var (
	ErrRequestBody    = errors.New("invalid request body")
	ErrQueryParameter = errors.New("invalid query parameter")
	ErrPathParameter  = errors.New("invalid path parameter")

	ErrUnauthorized = errors.New("invalid authorization token")
)

type errorResponse struct {
	Error string `json:"error"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type createTaskRequest struct {
	AssigneeID  int64     `json:"assignee_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
}

type updateTaskRequest struct {
	AssigneeID  *int64     `json:"assignee_id"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	Deadline    *time.Time `json:"deadline"`
}

type pageResponse struct {
	Items    []any `json:"items"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int   `json:"total"`
}

type taskResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status"`
	CreatorID   int64     `json:"creator_id"`
	AssigneeID  int64     `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toTaskResponse(t *domain.Task) taskResponse {
	return taskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Deadline:    t.Deadline,
		Status:      string(t.Status),
		CreatorID:   t.CreatorID,
		AssigneeID:  t.AssigneeID,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}
