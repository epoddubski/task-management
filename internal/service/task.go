package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"task-management/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id int64) (*domain.Task, error)
	List(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, int, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id int64) error
}

type TaskService struct {
	tasks TaskRepository
	users UserRepository
}

func NewTaskService(tasks TaskRepository, users UserRepository) *TaskService {
	return &TaskService{tasks: tasks, users: users}
}

func (s *TaskService) Create(ctx context.Context, actorID int64, task domain.CreateTaskInput) (*domain.Task, error) {
	title := strings.TrimSpace(task.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", domain.ErrValidation)
	}

	if task.Deadline.IsZero() {
		return nil, fmt.Errorf("%w: deadline is required", domain.ErrValidation)
	}

	assigneeID := actorID
	if task.AssigneeID != 0 {
		assigneeID = task.AssigneeID
	}

	if err := s.checkAssignee(ctx, assigneeID); err != nil {
		return nil, err
	}

	status := domain.StatusCreated
	if !task.Deadline.After(time.Now()) {
		status = domain.StatusOverdue
	}

	t := &domain.Task{
		Title:       title,
		Description: task.Description,
		Deadline:    task.Deadline,
		CreatorID:   actorID,
		AssigneeID:  assigneeID,
		Status:      status,
	}

	err := s.tasks.Create(ctx, t)
	if err != nil {
		return nil, err
	}

	return t, err
}

func (s *TaskService) Get(ctx context.Context, id int64) (*domain.Task, error) {
	return s.tasks.GetByID(ctx, id)
}

func (s *TaskService) List(ctx context.Context, list domain.ListTaskInput) (*domain.TaskPage, error) {
	if list.Status != nil && !list.Status.Valid() {
		return nil, fmt.Errorf("%w: invalid status filter", domain.ErrValidation)
	}

	if list.AssigneeID != nil {
		if err := s.checkAssignee(ctx, *list.AssigneeID); err != nil {
			return nil, fmt.Errorf("%w: invalid assignee filter", err)
		}
	}

	if list.DeadlineFrom != nil && list.DeadlineTo != nil {
		if list.DeadlineFrom.After(*list.DeadlineTo) {
			return nil, fmt.Errorf("%w: invalid deadline interval", domain.ErrValidation)
		}
	}

	var (
		page     int = 1
		pageSize int = 10
	)

	if list.Page != nil {
		page = *list.Page
		if page < 1 {
			page = 1
		}
	}

	if list.PageSize != nil {
		pageSize = *list.PageSize
		if pageSize < 1 || pageSize > 50 {
			pageSize = 10
		}
	}

	filter := domain.TaskFilter{
		Status:       list.Status,
		AssigneeID:   list.AssigneeID,
		DeadlineFrom: list.DeadlineFrom,
		DeadlineTo:   list.DeadlineTo,
		Limit:        pageSize,
		Offset:       (page - 1) * pageSize,
	}

	tasks, total, err := s.tasks.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &domain.TaskPage{
		Tasks:    tasks,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}

func (s *TaskService) Update(ctx context.Context, id, actorID int64, upd domain.UpdateTaskInput) (*domain.Task, error) {
	task, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if task.CreatorID != actorID {
		return nil, domain.ErrForbidden
	}

	if upd.Title != nil {
		title := strings.TrimSpace(*upd.Title)
		if title == "" {
			return nil, fmt.Errorf("%w: title cannot be empty", domain.ErrValidation)
		}
		task.Title = title
	}

	if upd.Description != nil {
		task.Description = *upd.Description
	}

	if upd.Status != nil {
		status := domain.TaskStatus(*upd.Status)
		if !status.Valid() {
			return nil, fmt.Errorf("%w: invalid status %s", domain.ErrValidation, *upd.Status)
		}
		task.Status = status
	}

	if upd.Deadline != nil {
		if upd.Deadline.IsZero() {
			return nil, fmt.Errorf("%w: empty deadline", domain.ErrValidation)
		}
		if !upd.Deadline.After(time.Now()) {
			task.Status = domain.StatusOverdue
		}
		task.Deadline = *upd.Deadline
	}

	if upd.AssigneeID != nil {
		if err := s.checkAssignee(ctx, *upd.AssigneeID); err != nil {
			return nil, err
		}
		task.AssigneeID = *upd.AssigneeID
	}

	err = s.tasks.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) Delete(ctx context.Context, id, actorID int64) error {
	task, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if task.CreatorID != actorID {
		return domain.ErrForbidden
	}

	return s.tasks.Delete(ctx, id)
}

func (s *TaskService) checkAssignee(ctx context.Context, assigneeID int64) error {
	_, err := s.users.GetByID(ctx, assigneeID)
	if errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("%w: assignee does not exist", domain.ErrValidation)
	}
	return err
}
