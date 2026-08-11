package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/benchmark/go-ai-review-benchmark/internal/repository"
	"github.com/benchmark/go-ai-review-benchmark/internal/service"
)

// BatchHandler handles HTTP routes for async batch job triggers and status.
type BatchHandler struct {
	processor *service.BatchProcessor
	repo      *repository.BatchRepository
}

// NewBatchHandler constructs a BatchHandler.
func NewBatchHandler(processor *service.BatchProcessor, repo *repository.BatchRepository) *BatchHandler {
	return &BatchHandler{
		processor: processor,
		repo:      repo,
	}
}

// RegisterRoutes sets up HTTP routes for batch operations.
func (h *BatchHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/batch/submit", h.handleSubmit)
	mux.HandleFunc("/api/batch/status", h.handleStatus)
}

type submitBatchRequest struct {
	Name  string   `json:"name"`
	Items []string `json:"items"`
}

func (h *BatchHandler) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req submitBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	jobID := fmt.Sprintf("job-%d", time.Now().UnixNano())
	job := &model.BatchJob{
		ID:         jobID,
		Name:       req.Name,
		Status:     model.StatusPending,
		TotalItems: len(req.Items),
	}

	if err := h.repo.CreateJob(r.Context(), job); err != nil {
		http.Error(w, "Failed to create batch job", http.StatusInternalServerError)
		return
	}

	tasks := make([]model.BatchTask, len(req.Items))
	for i, item := range req.Items {
		tasks[i] = model.BatchTask{
			JobID:   jobID,
			Payload: item,
			Status:  "PENDING",
		}
	}

	if err := h.repo.SaveTasks(r.Context(), tasks); err != nil {
		http.Error(w, "Failed to persist batch tasks", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	summary, err := h.processor.ProcessJob(ctx, jobID, tasks)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing batch: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(summary)
}

func (h *BatchHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(w, "Missing job id parameter", http.StatusBadRequest)
		return
	}

	job, err := h.repo.GetJobByID(r.Context(), jobID)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}
