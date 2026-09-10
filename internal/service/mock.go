package service

import (
	"context"
	"sync"
	"time"

	"task-management/internal/domain"
)

type fakeUserRepo struct {
	mu      sync.Mutex
	counter int64

	byID  map[int64]*domain.User
	email map[string]int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{byID: map[int64]*domain.User{}, email: map[string]int64{}}
}

func (r *fakeUserRepo) Create(_ context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++

	user.ID = r.counter
	user.CreatedAt = time.Now()

	if _, ok := r.email[user.Email]; ok {
		return domain.ErrUserExists
	}

	cp := *user

	r.byID[user.ID] = &cp
	r.email[user.Email] = user.ID

	return nil
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id, ok := r.email[email]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	cp := *r.byID[id]
	return &cp, nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}

	cp := *u
	return &cp, nil
}

type fakeTaskRepo struct {
	mu      sync.Mutex
	counter int64
	byID    map[int64]*domain.Task
}

func newFakeTaskRepo() *fakeTaskRepo {
	return &fakeTaskRepo{
		byID: map[int64]*domain.Task{},
	}
}

func (r *fakeTaskRepo) Create(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	task.ID = r.counter

	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	cp := *task
	r.byID[task.ID] = &cp
	return nil
}

func (r *fakeTaskRepo) GetByID(_ context.Context, id int64) (*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.byID[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}

	cp := *t
	return &cp, nil
}

func (r *fakeTaskRepo) List(_ context.Context, filter domain.TaskFilter) ([]domain.Task, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var matched []domain.Task

	for _, t := range r.byID {
		if filter.Status != nil && t.Status != *filter.Status {
			continue
		}
		if filter.AssigneeID != nil && t.AssigneeID != *filter.AssigneeID {
			continue
		}
		if filter.DeadlineFrom != nil && t.Deadline.Before(*filter.DeadlineFrom) {
			continue
		}
		if filter.DeadlineTo != nil && t.Deadline.After(*filter.DeadlineTo) {
			continue
		}
		matched = append(matched, *t)
	}

	total := len(matched)

	if filter.Limit >= total {
		return matched, total, nil
	}

	end := filter.Offset + filter.Limit
	if end > total {
		end = total
	}

	return matched[filter.Offset:end], total, nil
}

func (r *fakeTaskRepo) Update(_ context.Context, task *domain.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[task.ID]; !ok {
		return domain.ErrTaskNotFound
	}

	task.UpdatedAt = time.Now()

	cp := *task
	r.byID[task.ID] = &cp
	return nil
}

func (r *fakeTaskRepo) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.byID[id]; !ok {
		return domain.ErrTaskNotFound
	}

	delete(r.byID, id)
	return nil
}
