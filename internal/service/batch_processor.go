package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/benchmark/go-ai-review-benchmark/internal/model"
	"github.com/benchmark/go-ai-review-benchmark/internal/repository"
)

// BatchProcessor manages asynchronous job processing worker pools.
type BatchProcessor struct {
	repo         *repository.BatchRepository
	workerCount  int
	mu           sync.RWMutex
	statsMap     map[string]int
}

// NewBatchProcessor constructs a thread-safe BatchProcessor.
func NewBatchProcessor(repo *repository.BatchRepository, workerCount int) *BatchProcessor {
	return &BatchProcessor{
		repo:        repo,
		workerCount: workerCount,
		statsMap:    make(map[string]int),
	}
}

// ProcessJob executes tasks concurrently with context cancellation handling and thread-safe stats updates.
func (p *BatchProcessor) ProcessJob(ctx context.Context, jobID string, tasks []model.BatchTask) (*model.BatchSummary, error) {
	start := time.Now()

	err := p.repo.UpdateJobStatus(ctx, jobID, model.StatusRunning, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to mark job running: %w", err)
	}

	taskChan := make(chan model.BatchTask, len(tasks))
	resultChan := make(chan model.TaskResult, len(tasks))

	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	var wg sync.WaitGroup
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-taskChan:
					if !ok {
						return
					}
					res := p.executeTask(ctx, task)
					select {
					case <-ctx.Done():
						return
					case resultChan <- res:
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(resultChan)

	processed := 0
	failed := 0

	for res := range resultChan {
		if res.Success {
			processed++
			p.recordCategoryStat("SUCCESS")
		} else {
			failed++
			p.recordCategoryStat("FAILED")
		}
	}

	finalStatus := model.StatusCompleted
	if failed > 0 {
		finalStatus = model.StatusFailed
	}

	err = p.repo.UpdateJobStatus(ctx, jobID, finalStatus, processed, failed)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize job status: %w", err)
	}

	summary := &model.BatchSummary{
		JobID:          jobID,
		Status:         finalStatus,
		ProcessedCount: processed,
		FailedCount:    failed,
		CategoryCounts: p.getStatsSnapshot(),
		ExecutionSec:   time.Since(start).Seconds(),
	}

	return summary, nil
}

func (p *BatchProcessor) executeTask(ctx context.Context, task model.BatchTask) model.TaskResult {
	start := time.Now()
	select {
	case <-ctx.Done():
		return model.TaskResult{
			TaskID:    task.ID,
			JobID:     task.JobID,
			Success:   false,
			ErrorMsg:  ctx.Err().Error(),
			ExecTimeMs: time.Since(start).Milliseconds(),
		}
	default:
		time.Sleep(10 * time.Millisecond)
		return model.TaskResult{
			TaskID:    task.ID,
			JobID:     task.JobID,
			Success:   true,
			ExecTimeMs: time.Since(start).Milliseconds(),
		}
	}
}

func (p *BatchProcessor) recordCategoryStat(category string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.statsMap[category]++
}

func (p *BatchProcessor) getStatsSnapshot() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	snapshot := make(map[string]int, len(p.statsMap))
	for k, v := range p.statsMap {
		snapshot[k] = v
	}
	return snapshot
}
