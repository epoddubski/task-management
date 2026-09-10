package domain

import (
	"time"
)

type TaskStatus string

const (
	StatusCreated    TaskStatus = "created"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
	StatusOverdue    TaskStatus = "overdue"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case StatusCreated, StatusInProgress, StatusCompleted, StatusOverdue:
		return true
	default:
		return false
	}
}

type Task struct {
	ID         int64
	CreatorID  int64
	AssigneeID int64

	Title       string
	Description string
	Status      TaskStatus
	Deadline    time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateTaskInput struct {
	AssigneeID  int64
	Title       string
	Description string
	Deadline    time.Time
}

type UpdateTaskInput struct {
	AssigneeID  *int64
	Title       *string
	Description *string
	Status      *string
	Deadline    *time.Time
}

type ListTaskInput struct {
	Status     *TaskStatus
	AssigneeID *int64

	DeadlineFrom *time.Time
	DeadlineTo   *time.Time

	Page     *int
	PageSize *int
}

type TaskFilter struct {
	Status     *TaskStatus
	AssigneeID *int64

	DeadlineFrom *time.Time
	DeadlineTo   *time.Time

	Limit  int
	Offset int
}

type TaskPage struct {
	Tasks    []Task
	Page     int
	PageSize int
	Total    int
}
