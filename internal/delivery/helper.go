package delivery

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

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
