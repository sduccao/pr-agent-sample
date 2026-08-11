package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/jmoiron/sqlx"
)

// BatchRepository manages batch job and task persistence.
type BatchRepository struct {
	db *sqlx.DB
}

// NewBatchRepository constructs a BatchRepository.
func NewBatchRepository(db *sqlx.DB) *BatchRepository {
	return &BatchRepository{db: db}
}

// InitSchema creates tables for batch processing.
func (r *BatchRepository) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS batch_jobs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL,
		total_items INTEGER DEFAULT 0,
		processed INTEGER DEFAULT 0,
		failed INTEGER DEFAULT 0,
		priority INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS batch_tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		job_id TEXT NOT NULL,
		payload TEXT NOT NULL,
		status TEXT NOT NULL,
		retries INTEGER DEFAULT 0,
		metadata TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (job_id) REFERENCES batch_jobs(id)
	);`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// CreateJob inserts a new BatchJob entry into database.
func (r *BatchRepository) CreateJob(ctx context.Context, job *model.BatchJob) error {
	query := `INSERT INTO batch_jobs (id, name, status, total_items, processed, failed, priority, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, job.ID, job.Name, string(job.Status), job.TotalItems, job.Processed, job.Failed, job.Priority, now, now)
	if err != nil {
		return fmt.Errorf("batch repository CreateJob: %w", err)
	}
	return nil
}

// GetJobByID retrieves a BatchJob by its ID.
func (r *BatchRepository) GetJobByID(ctx context.Context, jobID string) (*model.BatchJob, error) {
	var job model.BatchJob
	query := `SELECT id, name, status, total_items, processed, failed, priority, created_at, updated_at FROM batch_jobs WHERE id = ?`
	err := r.db.GetContext(ctx, &job, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("batch repository GetJobByID: %w", err)
	}
	return &job, nil
}

// SaveTasks batch inserts tasks for a job.
func (r *BatchRepository) SaveTasks(ctx context.Context, tasks []model.BatchTask) error {
	query := `INSERT INTO batch_tasks (job_id, payload, status, retries, metadata, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	now := time.Now()
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("batch repository BeginTx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PreparexContext(ctx, query)
	if err != nil {
		return fmt.Errorf("batch repository Prepare: %w", err)
	}
	defer stmt.Close()

	for _, task := range tasks {
		_, err := stmt.ExecContext(ctx, task.JobID, task.Payload, task.Status, task.Retries, task.Metadata, now)
		if err != nil {
			return fmt.Errorf("batch repository InsertTask: %w", err)
		}
	}

	return tx.Commit()
}

// UpdateJobStatus updates metrics and status for a batch job.
func (r *BatchRepository) UpdateJobStatus(ctx context.Context, jobID string, status model.BatchJobStatus, processed, failed int) error {
	query := `UPDATE batch_jobs SET status = ?, processed = ?, failed = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, string(status), processed, failed, time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("batch repository UpdateJobStatus: %w", err)
	}
	return nil
}
