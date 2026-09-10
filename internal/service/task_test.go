package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"task-management/internal/domain"
)

func ptr[T any](v T) *T { return &v }

func TestTaskServiceCreate(t *testing.T) {
	t.Parallel()

	uRepo := newFakeUserRepo()

	_ = uRepo.Create(context.Background(), &domain.User{Email: "creator@example.com"})
	_ = uRepo.Create(context.Background(), &domain.User{Email: "assignee@example.com"})

	futureDeadline := time.Now().Add(12 * time.Hour)
	pastDeadline := time.Now().Add(-12 * time.Hour)

	cases := []struct {
		name         string
		actorID      int64
		input        domain.CreateTaskInput
		wantAssignee int64
		wantStatus   domain.TaskStatus
		wantErr      error
	}{
		{
			name:    "success",
			actorID: 1,
			input: domain.CreateTaskInput{
				Title:       "Valid Task",
				Description: "description",
				Deadline:    futureDeadline,
			},
			wantAssignee: 1,
			wantStatus:   domain.StatusCreated,
		},
		{
			name:    "success with assignee",
			actorID: 1,
			input: domain.CreateTaskInput{
				Title:      "Assigned Task",
				Deadline:   futureDeadline,
				AssigneeID: 2,
			},
			wantAssignee: 2,
			wantStatus:   domain.StatusCreated,
		},
		{
			name:    "success with past deadline",
			actorID: 1,
			input: domain.CreateTaskInput{
				Title:    "Expired Task",
				Deadline: pastDeadline,
			},
			wantAssignee: 1,
			wantStatus:   domain.StatusOverdue,
		},
		{
			name:    "empty title",
			actorID: 1,
			input: domain.CreateTaskInput{
				Title:    "   ",
				Deadline: futureDeadline,
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "zero deadline",
			actorID: 1,
			input: domain.CreateTaskInput{
				Title:    "Task without deadline",
				Deadline: time.Time{},
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "assignee does not exist",
			actorID: 1,
			input: domain.CreateTaskInput{
				Title:      "Task for ghost",
				Deadline:   futureDeadline,
				AssigneeID: 999,
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewTaskService(newFakeTaskRepo(), uRepo)
			task, err := s.Create(context.Background(), c.actorID, c.input)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error wrapper %v, got %v", c.wantErr, err)
			}

			if c.wantErr != nil {
				return
			}

			if task.CreatorID != c.actorID {
				t.Errorf("expected creator %d, got %d", c.actorID, task.CreatorID)
			}

			if task.AssigneeID != c.wantAssignee {
				t.Errorf("expected assignee %d, got %d", c.wantAssignee, task.AssigneeID)
			}

			if task.Status != c.wantStatus {
				t.Errorf("expected status %v, got %v", c.wantStatus, task.Status)
			}
		})
	}
}

func TestTaskServiceGet(t *testing.T) {
	t.Parallel()

	tRepo := newFakeTaskRepo()
	_ = tRepo.Create(context.Background(), &domain.Task{
		Title:       "Test Task",
		Description: "Some description",
		CreatorID:   1,
		Status:      domain.StatusCreated,
	})

	cases := []struct {
		name    string
		taskID  int64
		wantErr error
	}{
		{
			name:    "success",
			taskID:  1,
			wantErr: nil,
		},
		{
			name:    "task not found",
			taskID:  999,
			wantErr: domain.ErrTaskNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			s := NewTaskService(tRepo, newFakeUserRepo())
			task, err := s.Get(context.Background(), c.taskID)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error %v, got %v", c.wantErr, err)
			}

			if c.wantErr != nil {
				return
			}

			if task == nil {
				t.Fatalf("expected task to be not nil")
			}

			if task.ID != c.taskID {
				t.Errorf("expected task ID %d, got %d", c.taskID, task.ID)
			}

			if task.Title != "Test Task" {
				t.Errorf("expected title %q, got %q", "Test Task", task.Title)
			}
		})
	}
}

func TestTaskService_List(t *testing.T) {
	t.Parallel()

	uRepo := newFakeUserRepo()
	_ = uRepo.Create(context.Background(), &domain.User{Email: "user1@test.com"})
	_ = uRepo.Create(context.Background(), &domain.User{Email: "user2@test.com"})

	tRepo := newFakeTaskRepo()
	now := time.Now()

	tasksToCreate := []domain.Task{
		{Title: "Task 1", AssigneeID: 1, Status: domain.StatusCreated, Deadline: now.Add(2 * time.Hour)},
		{Title: "Task 2", AssigneeID: 1, Status: domain.StatusCompleted, Deadline: now.Add(5 * time.Hour)},
		{Title: "Task 3", AssigneeID: 2, Status: domain.StatusCreated, Deadline: now.Add(10 * time.Hour)},
		{Title: "Task 4", AssigneeID: 2, Status: domain.StatusOverdue, Deadline: now.Add(-2 * time.Hour)},
	}
	for i := range tasksToCreate {
		_ = tRepo.Create(context.Background(), &tasksToCreate[i])
	}

	s := NewTaskService(tRepo, uRepo)

	cases := []struct {
		name         string
		input        domain.ListTaskInput
		wantTasksLen int
		wantTotal    int
		wantPage     int
		wantPageSize int
		wantErr      error
	}{
		{
			name:         "all tasks with default pagination",
			input:        domain.ListTaskInput{},
			wantTasksLen: 4,
			wantTotal:    4,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name: "filter by status",
			input: domain.ListTaskInput{
				Status: ptr(domain.StatusCreated),
			},
			wantTasksLen: 2, // Task 1, Task 3
			wantTotal:    2,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name: "filter by assignee",
			input: domain.ListTaskInput{
				AssigneeID: ptr(int64(1)),
			},
			wantTasksLen: 2, // Task 1, Task 2
			wantTotal:    2,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name: "filter by deadline interval",
			input: domain.ListTaskInput{
				DeadlineFrom: ptr(now.Add(1 * time.Hour)),
				DeadlineTo:   ptr(now.Add(6 * time.Hour)),
			},
			wantTasksLen: 2, // Task 1 (2h), Task 2 (5h)
			wantTotal:    2,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name: "pagination limits",
			input: domain.ListTaskInput{
				PageSize: ptr(2),
			},
			wantTasksLen: 2, // len(items)
			wantTotal:    4,
			wantPage:     1,
			wantPageSize: 2,
		},
		{
			name: "pagination max",
			input: domain.ListTaskInput{
				PageSize: ptr(100), // > 50
			},
			wantTasksLen: 4,
			wantTotal:    4,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name: "pagination negative",
			input: domain.ListTaskInput{
				Page: ptr(-5),
			},
			wantTasksLen: 4,
			wantTotal:    4,
			wantPage:     1,
			wantPageSize: 10,
		},
		{
			name: "wrong status",
			input: domain.ListTaskInput{
				Status: ptr(domain.TaskStatus("INVALID_STATUS")),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "wrong assignee",
			input: domain.ListTaskInput{
				AssigneeID: ptr(int64(999)),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name: "wrong deadline interval",
			input: domain.ListTaskInput{
				DeadlineFrom: ptr(now.Add(10 * time.Hour)),
				DeadlineTo:   ptr(now.Add(2 * time.Hour)),
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			page, err := s.List(context.Background(), c.input)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error %v, got %v", c.wantErr, err)
			}
			if c.wantErr != nil {
				return
			}

			if page == nil {
				t.Fatalf("expected page to be not nil")
			}

			if len(page.Tasks) != c.wantTasksLen {
				t.Errorf("expected %d tasks on page, got %d", c.wantTasksLen, len(page.Tasks))
			}

			if page.Total != c.wantTotal {
				t.Errorf("expected total count %d, got %d", c.wantTotal, page.Total)
			}

			if page.Page != c.wantPage {
				t.Errorf("expected page number %d, got %d", c.wantPage, page.Page)
			}

			if page.PageSize != c.wantPageSize {
				t.Errorf("expected page size %d, got %d", c.wantPageSize, page.PageSize)
			}
		})
	}
}

func TestTaskServiceUpdate(t *testing.T) {
	t.Parallel()

	uRepo := newFakeUserRepo()
	_ = uRepo.Create(context.Background(), &domain.User{Email: "creator@test.com"})
	_ = uRepo.Create(context.Background(), &domain.User{Email: "assignee@test.com"})

	cases := []struct {
		name       string
		taskID     int64
		actorID    int64
		input      domain.UpdateTaskInput
		wantStatus domain.TaskStatus
		wantErr    error
	}{
		{
			name:    "success title",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				Title:       ptr("New Dynamic Title"),
				Description: ptr("description"),
			},
			wantStatus: domain.StatusCreated,
		},
		{
			name:    "forbidden",
			taskID:  1,
			actorID: 2,
			input: domain.UpdateTaskInput{
				Title: ptr("Malicious change"),
			},
			wantErr: domain.ErrForbidden,
		},
		{
			name:    "empty title",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				Title: ptr("   "),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "success status",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				Status: ptr(string(domain.StatusCompleted)),
			},
			wantStatus: domain.StatusCompleted,
		},
		{
			name:    "wrong status",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				Status: ptr("wrong status"),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "success deadline",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				Deadline: ptr(time.Now().Add(2 * time.Hour)),
			},
			wantStatus: domain.StatusCreated,
		},
		{
			name:    "empty deadline",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				Deadline: ptr(time.Time{}),
			},
			wantErr: domain.ErrValidation,
		},
		{
			name:    "overdue deadline status",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				Deadline: ptr(time.Now().Add(-2 * time.Hour)),
			},
			wantStatus: domain.StatusOverdue,
		},
		{
			name:    "success assignee",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				AssigneeID: ptr(int64(2)),
			},
			wantStatus: domain.StatusCreated,
		},
		{
			name:    "wrong assignee",
			taskID:  1,
			actorID: 1,
			input: domain.UpdateTaskInput{
				AssigneeID: ptr(int64(3)),
			},
			wantErr: domain.ErrValidation,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			tRepo := newFakeTaskRepo()
			_ = tRepo.Create(context.Background(), &domain.Task{
				Title:     "Old Title",
				CreatorID: 1,
				Status:    domain.StatusCreated,
				Deadline:  time.Now().Add(10 * time.Hour),
			})

			s := NewTaskService(tRepo, uRepo)
			updated, err := s.Update(context.Background(), c.taskID, c.actorID, c.input)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error %v, got %v", c.wantErr, err)
			}

			if c.wantErr != nil {
				return
			}

			if updated == nil {
				t.Fatalf("expected task to be not nil")
			}

			if updated.ID != c.taskID {
				t.Errorf("expected task ID %d, got %d", c.taskID, updated.ID)
			}

			if updated.CreatorID != 1 {
				t.Errorf("expected creator ID to remain 1, got %d", updated.CreatorID)
			}

			if c.input.Title != nil && updated.Title != *c.input.Title {
				t.Errorf("expected title to be updated to %s, got %s", *c.input.Title, updated.Title)
			}

			if c.input.Description != nil && updated.Description != *c.input.Description {
				t.Errorf("expected description %s, got %s", *c.input.Description, updated.Description)
			}

			if c.input.Deadline != nil && !updated.Deadline.Equal(*c.input.Deadline) {
				t.Errorf("expected deadline %v, got %v", *c.input.Deadline, updated.Deadline)
			}

			if c.input.AssigneeID != nil && updated.AssigneeID != *c.input.AssigneeID {
				t.Errorf("expected assignee %d, got %d", *c.input.AssigneeID, updated.AssigneeID)
			}

			if updated.Status != c.wantStatus {
				t.Errorf("expected status %v, got %v", c.wantStatus, updated.Status)
			}
		})
	}
}

func TestTaskServiceDelete(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		taskID  int64
		actorID int64
		wantErr error
	}{
		{
			name:    "success",
			taskID:  1,
			actorID: 1,
		},
		{
			name:    "forbidden",
			taskID:  1,
			actorID: 2,
			wantErr: domain.ErrForbidden,
		},
		{
			name:    "task not found",
			taskID:  999,
			actorID: 1,
			wantErr: domain.ErrTaskNotFound,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			tRepo := newFakeTaskRepo()
			_ = tRepo.Create(context.Background(), &domain.Task{
				Title:     "Task to be deleted",
				CreatorID: 1,
				Status:    domain.StatusCreated,
			})

			s := NewTaskService(tRepo, newFakeUserRepo())
			err := s.Delete(context.Background(), c.taskID, c.actorID)

			if !errors.Is(err, c.wantErr) {
				t.Fatalf("expected error %v, got %v", c.wantErr, err)
			}

			if c.wantErr != nil {
				return
			}

			_, errGet := tRepo.GetByID(context.Background(), c.taskID)
			if !errors.Is(errGet, domain.ErrTaskNotFound) {
				t.Errorf("expected task %d to be deleted from repository, but it still exists", c.taskID)
			}
		})
	}
}
