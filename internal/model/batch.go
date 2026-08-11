package model

import "time"

// BatchJobStatus denotes the processing state of a batch job.
type BatchJobStatus string

const (
	StatusPending   BatchJobStatus = "PENDING"
	StatusRunning   BatchJobStatus = "RUNNING"
	StatusCompleted BatchJobStatus = "COMPLETED"
	StatusFailed    BatchJobStatus = "FAILED"
)

// BatchJob represents a collection of tasks to process concurrently.
type BatchJob struct {
	ID         string         `json:"id" db:"id"`
	Name       string         `json:"name" db:"name"`
	Status     BatchJobStatus `json:"status" db:"status"`
	TotalItems int            `json:"total_items" db:"total_items"`
	Processed  int            `json:"processed" db:"processed"`
	Failed     int            `json:"failed" db:"failed"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at" db:"updated_at"`
}

// BatchTask represents an individual unit of work inside a batch job.
type BatchTask struct {
	ID        int64     `json:"id" db:"id"`
	JobID     string    `json:"job_id" db:"job_id"`
	Payload   string    `json:"payload" db:"payload"`
	Status    string    `json:"status" db:"status"`
	Retries   int       `json:"retries" db:"retries"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TaskResult captures output of an executed batch task.
type TaskResult struct {
	TaskID    int64  `json:"task_id"`
	JobID     string `json:"job_id"`
	Success   bool   `json:"success"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	ExecTimeMs int64 `json:"exec_time_ms"`
}

// BatchSummary holds aggregated statistics for batch operations.
type BatchSummary struct {
	JobID           string         `json:"job_id"`
	Status          BatchJobStatus `json:"status"`
	ProcessedCount  int            `json:"processed_count"`
	FailedCount     int            `json:"failed_count"`
	CategoryCounts  map[string]int `json:"category_counts"`
	ExecutionSec    float64        `json:"execution_sec"`
}
