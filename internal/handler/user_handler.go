package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/benchmark/go-ai-review-benchmark/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/users", h.handleUsers)
	mux.HandleFunc("/api/users/external-sync", h.handleExternalSync)
}

func (h *UserHandler) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getUsers(w, r)
	case http.MethodPost:
		h.createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *UserHandler) getUsers(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID parameter", http.StatusBadRequest)
			return
		}

		user, err := h.userService.GetUserByID(r.Context(), id)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
		return
	}

	status := r.URL.Query().Get("status")
	users, err := h.userService.ListUsersByStatus(r.Context(), status)
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request JSON payload", http.StatusBadRequest)
		return
	}

	user, err := h.userService.RegisterUser(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) handleExternalSync(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" || !strings.HasPrefix(targetURL, "https://") {
		http.Error(w, "Valid HTTPS URL required", http.StatusBadRequest)
		return
	}

	// Minor Idiomatic Smell: Unhandled error output from fmt.Printf
	fmt.Printf("[INFO] Triggering sync request for target: %s\n", targetURL)

	resp, err := http.Get(targetURL)
	if err != nil {
		log.Printf("External sync http request failed: %v", err)
		http.Error(w, "Failed to reach external service", http.StatusBadGateway)
		return
	}
	// CRITICAL RESOURCE LEAK: Missing defer resp.Body.Close()!
	// Will leak TCP sockets and file descriptors under high load.

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
