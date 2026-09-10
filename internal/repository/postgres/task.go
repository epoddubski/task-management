package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"task-management/internal/domain"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) error {
	const q = `
		INSERT INTO tasks (title, description, deadline, status, creator_id, assignee_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, q,
		t.Title,
		t.Description,
		t.Deadline,
		t.Status,
		t.CreatorID,
		t.AssigneeID,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	const q = `
		SELECT id, title, description, deadline, status, creator_id, assignee_id, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	t := &domain.Task{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Deadline,
		&t.Status,
		&t.CreatorID,
		&t.AssigneeID,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTaskNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select task by id: %w", err)
	}
	return t, nil
}

func (r *TaskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, int, error) {
	var (
		conditions []string
		args       []any
	)

	add := func(cond string, val any) {
		args = append(args, val)
		conditions = append(conditions, fmt.Sprintf(cond, len(args)))
	}

	if filter.Status != nil {
		add("status = $%d", *filter.Status)
	}
	if filter.AssigneeID != nil {
		add("assignee_id = $%d", *filter.AssigneeID)
	}
	if filter.DeadlineFrom != nil {
		add("deadline >= $%d", *filter.DeadlineFrom)
	}
	if filter.DeadlineTo != nil {
		add("deadline <= $%d", *filter.DeadlineTo)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQ := "SELECT count(*) FROM tasks " + where

	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	args = append(args, filter.Limit)
	limitPlaceholder := fmt.Sprintf("$%d", len(args))

	args = append(args, filter.Offset)
	offsetPlaceholder := fmt.Sprintf("$%d", len(args))

	q := fmt.Sprintf(`
		SELECT id, title, description, deadline, status, creator_id, assignee_id, created_at, updated_at
		FROM tasks
		%s
		ORDER BY created_at DESC
		LIMIT %s OFFSET %s`, where, limitPlaceholder, offsetPlaceholder)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		t := domain.Task{}
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Deadline,
			&t.Status,
			&t.CreatorID,
			&t.AssigneeID,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan task row: %w", err)
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate task rows: %w", err)
	}

	return tasks, total, nil
}

func (r *TaskRepository) Update(ctx context.Context, t *domain.Task) error {
	const q = `
		UPDATE tasks
		SET title = $1, description = $2, deadline = $3, status = $4, assignee_id = $5, updated_at = now()
		WHERE id = $6
		RETURNING updated_at`

	err := r.db.QueryRowContext(ctx, q,
		t.Title,
		t.Description,
		t.Deadline,
		t.Status,
		t.AssigneeID,
		t.ID,
	).Scan(&t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepository) MarkOverdue(ctx context.Context, now time.Time) (int64, error) {
	const q = `
		UPDATE tasks
		SET status = 'overdue', updated_at = now()
		WHERE deadline IS NOT NULL
		  AND deadline < $1
		  AND status NOT IN ($2, $3)`

	res, err := r.db.ExecContext(ctx, q, now, domain.StatusCompleted, domain.StatusOverdue)
	if err != nil {
		return 0, fmt.Errorf("mark overdue: %w", err)
	}
	return res.RowsAffected()
}
