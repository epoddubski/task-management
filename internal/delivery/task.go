package delivery

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"task-management/internal/domain"
)

type TaskService interface {
	Create(ctx context.Context, creatorID int64, task domain.CreateTaskInput) (*domain.Task, error)
	Get(ctx context.Context, id int64) (*domain.Task, error)
	List(ctx context.Context, filter domain.ListTaskInput) (*domain.TaskPage, error)
	Update(ctx context.Context, id, userID int64, task domain.UpdateTaskInput) (*domain.Task, error)
	Delete(ctx context.Context, id, userID int64) error
}

type TaskHandler struct {
	tasks TaskService
}

func NewTaskHandler(task TaskService) *TaskHandler {
	return &TaskHandler{tasks: task}
}
func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	var req createTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	task, err := h.tasks.Create(r.Context(), userID, domain.CreateTaskInput{
		Title:       req.Title,
		Description: req.Description,
		Deadline:    req.Deadline,
		AssigneeID:  req.AssigneeID,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTaskResponse(task))
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, err)
		return
	}

	task, err := h.tasks.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	input, err := parseQueryFilter(r)
	if err != nil {
		writeError(w, fmt.Errorf("%w: %v", ErrQueryParameter, err))
		return
	}

	result, err := h.tasks.List(r.Context(), *input)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := pageResponse{
		Items:    make([]any, len(result.Tasks)),
		Page:     result.Page,
		PageSize: result.PageSize,
		Total:    result.Total,
	}

	for i, t := range result.Tasks {
		resp.Items[i] = toTaskResponse(&t)
	}

	writeJSON(w, http.StatusOK, resp)
}

func parseQueryFilter(r *http.Request) (*domain.ListTaskInput, error) {
	var in domain.ListTaskInput

	q := r.URL.Query()

	if s := q.Get("status"); s != "" {
		status := domain.TaskStatus(s)
		in.Status = &status
	}

	if id := q.Get("assignee_id"); id != "" {
		assigneeID, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("assignee_id must be a valid integer")
		}
		in.AssigneeID = &assigneeID
	}

	if from := q.Get("deadline_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			return nil, fmt.Errorf("deadline_from must be a valid RFC3339 timestamp")
		}
		in.DeadlineFrom = &t
	}

	if to := q.Get("deadline_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			return nil, fmt.Errorf("deadline_to must be a valid RFC3339 timestamp")
		}
		in.DeadlineTo = &t
	}

	if page := q.Get("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			return nil, fmt.Errorf("page must be a valid integer")
		}
		in.Page = &p
	}

	if size := q.Get("page_size"); size != "" {
		s, err := strconv.Atoi(size)
		if err != nil {
			return nil, fmt.Errorf("page_size must be a valid integer")
		}
		in.PageSize = &s
	}

	return &in, nil
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, err)
		return
	}

	var req updateTaskRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	task, err := h.tasks.Update(r.Context(), id, userID, domain.UpdateTaskInput{
		AssigneeID:  req.AssigneeID,
		Title:       req.Title,
		Description: req.Description,
		Deadline:    req.Deadline,
		Status:      req.Status,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := userIDFromContext(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	id, err := parseIDParam(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if err := h.tasks.Delete(r.Context(), id, userID); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
