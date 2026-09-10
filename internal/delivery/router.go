package delivery

import (
	"net/http"
)

func NewRouter(auth AuthService, task TaskService) http.Handler {
	authHandler := NewAuthHandler(auth)
	taskHandler := NewTaskHandler(task)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	requireAuth := AuthMiddleware(auth)

	mux.Handle("POST /tasks", requireAuth(http.HandlerFunc(taskHandler.Create)))
	mux.Handle("GET /tasks", requireAuth(http.HandlerFunc(taskHandler.List)))
	mux.Handle("GET /tasks/{id}", requireAuth(http.HandlerFunc(taskHandler.Get)))
	mux.Handle("PATCH /tasks/{id}", requireAuth(http.HandlerFunc(taskHandler.Update)))
	mux.Handle("DELETE /tasks/{id}", requireAuth(http.HandlerFunc(taskHandler.Delete)))

	return RecoverMiddleware(LoggingMiddleware(mux))
}
