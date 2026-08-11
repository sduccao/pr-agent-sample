package service

import (
	"context"
	"time"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/benchmark/go-ai-review-benchmark/internal/repository"
)

// BatchProcessor manages asynchronous job processing worker pools.
type BatchProcessor struct {
	repo        *repository.BatchRepository
	workerCount int
	statsMap    map[string]int // BUG: Unprotected map causing Data Race under concurrent writes
}

// NewBatchProcessor constructs a BatchProcessor.
func NewBatchProcessor(repo *repository.BatchRepository, workerCount int) *BatchProcessor {
	return &BatchProcessor{
		repo:        repo,
		workerCount: workerCount,
		statsMap:    make(map[string]int),
	}
}

// ProcessJob executes tasks concurrently across worker goroutines.
func (p *BatchProcessor) ProcessJob(ctx context.Context, jobID string, tasks []model.BatchTask) (*model.BatchSummary, error) {
	start := time.Now()

	_ = p.repo.UpdateJobStatus(ctx, jobID, model.StatusRunning, 0, 0)

	// Unbuffered channel for task distribution
	taskChan := make(chan model.BatchTask)
	resultChan := make(chan model.TaskResult)

	// BUG: Goroutine Leak! Workers are spawned listening on unbuffered channel taskChan without context cancellation.
	// If task distribution stops early or an error occurs, these goroutines remain blocked indefinitely.
	for i := 0; i < p.workerCount; i++ {
		go func(workerID int) {
			for task := range taskChan { // Missing select on ctx.Done()
				res := p.executeTask(task)
				resultChan <- res
			}
		}(i)
	}

	// Push tasks to workers in background
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		// Note: taskChan is intentionally NOT closed in error paths, leaving workers stuck waiting
	}()

	processed := 0
	failed := 0

	for i := 0; i < len(tasks); i++ {
		res := <-resultChan
		if res.Success {
			processed++
			// BUG: Concurrent map write without sync.Mutex -> DATA RACE!
			p.statsMap["SUCCESS"]++
		} else {
			failed++
			// BUG: Concurrent map write without sync.Mutex -> DATA RACE!
			p.statsMap["FAILED"]++
		}
	}

	finalStatus := model.StatusCompleted
	if failed > 0 {
		finalStatus = model.StatusFailed
	}

	_ = p.repo.UpdateJobStatus(ctx, jobID, finalStatus, processed, failed)

	summary := &model.BatchSummary{
		JobID:          jobID,
		Status:         finalStatus,
		ProcessedCount: processed,
		FailedCount:    failed,
		CategoryCounts: p.statsMap,
		ExecutionSec:   time.Since(start).Seconds(),
	}

	return summary, nil
}

func (p *BatchProcessor) executeTask(task model.BatchTask) model.TaskResult {
	start := time.Now()
	time.Sleep(5 * time.Millisecond)
	return model.TaskResult{
		TaskID:     task.ID,
		JobID:      task.JobID,
		Success:    true,
		ExecTimeMs: time.Since(start).Milliseconds(),
	}
}
