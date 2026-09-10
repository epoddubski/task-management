package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"task-management/internal/domain"
)

func parseIDParam(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid id %q", ErrPathParameter, idStr)
	}
	return id, nil
}

func userIDFromContext(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(userIDContextKey{}).(int64)
	if !ok {
		return 0, ErrUnauthorized
	}
	return id, nil
}

func decodeJSON(r *http.Request, req any) error {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return fmt.Errorf("%w: %v", ErrRequestBody, err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	status := mapErrorToStatus(err)
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func mapErrorToStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrTaskNotFound),
		errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound

	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden

	case errors.Is(err, domain.ErrInvalidCreds),
		errors.Is(err, domain.ErrInvalidToken),
		errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized

	case errors.Is(err, domain.ErrUserExists):
		return http.StatusConflict

	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity

	case errors.Is(err, ErrRequestBody),
		errors.Is(err, ErrQueryParameter),
		errors.Is(err, ErrPathParameter):
		return http.StatusBadRequest

	default:
		return http.StatusInternalServerError
	}
}
