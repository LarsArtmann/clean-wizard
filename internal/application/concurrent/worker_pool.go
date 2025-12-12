package concurrent

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/shared/errors"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// WorkerPool represents a concurrent worker pool
type WorkerPool struct {
	maxWorkers int
	workerChan chan func()
	errorChan  chan error
	resultChan chan any
	waitGroup  sync.WaitGroup
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	logger     errors.Logger
	stats      *PoolStatistics
}

// PoolStatistics holds worker pool statistics
type PoolStatistics struct {
	TotalTasks     uint64    `json:"total_tasks"`
	CompletedTasks uint64    `json:"completed_tasks"`
	FailedTasks    uint64    `json:"failed_tasks"`
	ActiveWorkers  uint64    `json:"active_workers"`
	PendingTasks   uint64    `json:"pending_tasks"`
	LastTaskTime   time.Time `json:"last_task_time"`
	PoolStartTime  time.Time `json:"pool_start_time"`
}

// NewWorkerPool creates new worker pool
func NewWorkerPool(maxWorkers int, logger errors.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	wp := &WorkerPool{
		maxWorkers: maxWorkers,
		workerChan: make(chan func(), maxWorkers*2),
		errorChan:  make(chan error, maxWorkers),
		resultChan: make(chan any, maxWorkers),
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
		stats: &PoolStatistics{
			PoolStartTime: time.Now(),
		},
	}

	return wp
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.logger.Info("Starting worker pool", "max_workers", wp.maxWorkers)

	for i := 0; i < wp.maxWorkers; i++ {
		wp.waitGroup.Add(1)
		go wp.worker(i)
	}

	wp.waitGroup.Add(1)
	go wp.errorHandler()
}

// Stop stops the worker pool gracefully
func (wp *WorkerPool) Stop(timeout time.Duration) error {
	wp.logger.Info("Stopping worker pool", "timeout", timeout.String())

	wp.cancel()

	done := make(chan struct{})
	go func() {
		wp.waitGroup.Wait()
		close(done)
	}()

	select {
	case <-done:
		wp.logger.Info("Worker pool stopped gracefully")
		return nil
	case <-time.After(timeout):
		wp.logger.Warn("Worker pool stop timeout")
		return errors.NewErrorFactory("worker_pool").
			Timeout("worker pool stop timeout").
			ToResult()
	}
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task func()) error {
	select {
	case wp.workerChan <- task:
		wp.mutex.Lock()
		wp.stats.TotalTasks++
		wp.stats.PendingTasks++
		wp.stats.LastTaskTime = time.Now()
		wp.mutex.Unlock()
		return nil
	case <-wp.ctx.Done():
		return errors.NewErrorFactory("worker_pool").
			Cancelled("worker pool is shutting down").
			ToResult()
	default:
		return errors.NewErrorFactory("worker_pool").
			Execution("worker pool is full").
			ToResult()
	}
}

// SubmitWithTimeout submits a task with timeout
func (wp *WorkerPool) SubmitWithTimeout(task func(), timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case wp.workerChan <- task:
		wp.mutex.Lock()
		wp.stats.TotalTasks++
		wp.stats.PendingTasks++
		wp.stats.LastTaskTime = time.Now()
		wp.mutex.Unlock()
		return nil
	case <-timer.C:
		return errors.NewErrorFactory("worker_pool").
			Timeout("task submission timeout").
			ToResult()
	case <-wp.ctx.Done():
		return errors.NewErrorFactory("worker_pool").
			Cancelled("worker pool is shutting down").
			ToResult()
	}
}

// GetStats returns worker pool statistics
func (wp *WorkerPool) GetStats() PoolStatistics {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	return *wp.stats
}

// worker processes tasks from the worker channel
func (wp *WorkerPool) worker(id int) {
	defer wp.waitGroup.Done()

	wp.logger.Info("Worker started", "worker_id", id)

	wp.mutex.Lock()
	wp.stats.ActiveWorkers++
	wp.mutex.Unlock()

	defer func() {
		wp.mutex.Lock()
		wp.stats.ActiveWorkers--
		wp.mutex.Unlock()
		wp.logger.Info("Worker stopped", "worker_id", id)
	}()

	for {
		select {
		case task := <-wp.workerChan:
			wp.executeTask(task, id)
		case <-wp.ctx.Done():
			return
		}
	}
}

// executeTask executes a single task
func (wp *WorkerPool) executeTask(task func(), workerID int) {
	wp.mutex.Lock()
	wp.stats.PendingTasks--
	wp.mutex.Unlock()

	defer func() {
		wp.mutex.Lock()
		wp.stats.CompletedTasks++
		wp.mutex.Unlock()
	}()

	if err := wp.executeSafely(task, workerID); err != nil {
		wp.mutex.Lock()
		wp.stats.FailedTasks++
		wp.mutex.Unlock()

		select {
		case wp.errorChan <- err:
		case <-wp.ctx.Done():
		}
	}
}

// executeSafely executes task safely with panic recovery
func (wp *WorkerPool) executeSafely(task func(), workerID int) (err error) {
	defer func() {
		if r := recover(); r != nil {
			factory := errors.NewErrorFactory("worker_panic")
			err = factory.Execution(fmt.Sprintf("worker %d panicked: %v", workerID, r)).
				WithDetail("worker_id", workerID).
				WithDetail("panic", fmt.Sprintf("%v", r))
		}
	}()

	task()
	return nil
}

// errorHandler handles errors from workers
func (wp *WorkerPool) errorHandler() {
	defer wp.waitGroup.Done()

	for {
		select {
		case err := <-wp.errorChan:
			wp.logger.Error("Worker task failed", "error", err.Error())
		case <-wp.ctx.Done():
			return
		}
	}
}

// ConcurrentExecutor manages concurrent operation execution
type ConcurrentExecutor struct {
	workerPool   *WorkerPool
	rollbackMgr  *RollbackManager
	operationMgr *OperationManager
	logger       errors.Logger
}

// NewConcurrentExecutor creates new concurrent executor
func NewConcurrentExecutor(
	maxWorkers int,
	rollbackMgr *RollbackManager,
	operationMgr *OperationManager,
	logger errors.Logger,
) *ConcurrentExecutor {
	return &ConcurrentExecutor{
		workerPool:   NewWorkerPool(maxWorkers, logger),
		rollbackMgr:  rollbackMgr,
		operationMgr: operationMgr,
		logger:       logger,
	}
}

// Start starts the concurrent executor
func (ce *ConcurrentExecutor) Start() {
	ce.workerPool.Start()
}

// Stop stops the concurrent executor
func (ce *ConcurrentExecutor) Stop(timeout time.Duration) error {
	return ce.workerPool.Stop(timeout)
}

// ExecuteConcurrent executes operations concurrently
func (ce *ConcurrentExecutor) ExecuteConcurrent(
	ctx context.Context,
	operations []*ConcurrentOperation,
) result.Result[*ConcurrentExecutionResult] {
	if len(operations) == 0 {
		return result.Err[*ConcurrentExecutionResult](
			errors.NewErrorFactory("concurrent_executor").
				InvalidInput("no operations provided").
				ToResult())
	}

	ce.logger.Info("Starting concurrent execution",
		"operation_count", len(operations))

	result := &ConcurrentExecutionResult{
		Operations: make([]*OperationResult, len(operations)),
		StartTime:  time.Now(),
		State:      ExecutionStateRunning,
	}

	// Submit all operations
	var waitGroup sync.WaitGroup
	resultMutex := sync.Mutex{}

	for i, op := range operations {
		waitGroup.Add(1)

		task := func(index int, operation *ConcurrentOperation) func() {
			return func() {
				defer waitGroup.Done()
				opResult := ce.executeOperation(ctx, operation)

				resultMutex.Lock()
				result.Operations[index] = opResult
				resultMutex.Unlock()
			}
		}(i, op)

		// Submit to worker pool
		if err := ce.workerPool.Submit(task); err != nil {
			waitGroup.Done()

			resultMutex.Lock()
			result.Operations[i] = &OperationResult{
				OperationID: operation.ID,
				Name:        operation.Name,
				State:       OperationStateFailed,
				Error:       errors.AddContext(err, "concurrent_executor"),
				StartTime:   time.Now(),
			}
			resultMutex.Unlock()
		}
	}

	// Wait for all operations to complete
	waitGroup.Wait()

	resultMutex.Lock()
	result.EndTime = time.Now()
	result.Duration = &result.EndTime.Sub(result.StartTime)

	// Calculate final state
	completedCount := 0
	failedCount := 0

	for _, opResult := range result.Operations {
		if opResult.State == OperationStateCompleted {
			completedCount++
		} else if opResult.State == OperationStateFailed {
			failedCount++
		}
	}

	if failedCount == 0 {
		result.State = ExecutionStateCompleted
	} else if completedCount > 0 {
		result.State = ExecutionStatePartial
	} else {
		result.State = ExecutionStateFailed
	}

	result.TotalOperations = len(operations)
	result.CompletedOperations = completedCount
	result.FailedOperations = failedCount
	resultMutex.Unlock()

	ce.logger.Info("Concurrent execution completed",
		"total_operations", result.TotalOperations,
		"completed_operations", result.CompletedOperations,
		"failed_operations", result.FailedOperations,
		"state", result.State.String(),
		"duration", result.Duration.String())

	return result.Ok(result)
}

// executeOperation executes a single operation
func (ce *ConcurrentExecutor) executeOperation(
	ctx context.Context,
	operation *ConcurrentOperation,
) *OperationResult {
	result := &OperationResult{
		OperationID: operation.ID,
		Name:        operation.Name,
		State:       OperationStatePending,
		StartTime:   time.Now(),
	}

	// Start operation
	if err := ce.operationMgr.StartOperation(operation.ID); err != nil {
		result.State = OperationStateFailed
		result.Error = err
		return result
	}

	result.State = OperationStateRunning

	// Execute operation with rollback function
	opResult := operation.Execute(ctx)

	if opResult.IsOk() {
		// Operation completed successfully
		ce.operationMgr.CompleteOperation(operation.ID, opResult.Value())
		result.State = OperationStateCompleted
		result.Result = &opResult.Value()
	} else {
		// Operation failed
		ce.operationMgr.FailOperation(operation.ID, opResult.Error())
		result.State = OperationStateFailed
		result.Error = opResult.Error()
	}

	result.EndTime = time.Now()
	result.Duration = &result.EndTime.Sub(result.StartTime)

	return result
}

// ConcurrentOperation represents an operation to be executed concurrently
type ConcurrentOperation struct {
	ID      OperationID
	Name    string
	Execute func(ctx context.Context) result.Result[shared.CleanResult]
}

// OperationResult represents the result of a concurrent operation
type OperationResult struct {
	OperationID OperationID         `json:"operation_id"`
	Name        string              `json:"name"`
	State       OperationState      `json:"state"`
	StartTime   time.Time           `json:"start_time"`
	EndTime     *time.Time          `json:"end_time,omitempty"`
	Duration    *time.Duration      `json:"duration,omitempty"`
	Result      *shared.CleanResult `json:"result,omitempty"`
	Error       *errors.DomainError `json:"error,omitempty"`
}

// ConcurrentExecutionResult represents the result of concurrent execution
type ConcurrentExecutionResult struct {
	Operations          []*OperationResult `json:"operations"`
	TotalOperations     int                `json:"total_operations"`
	CompletedOperations int                `json:"completed_operations"`
	FailedOperations    int                `json:"failed_operations"`
	StartTime           time.Time          `json:"start_time"`
	EndTime             *time.Time         `json:"end_time,omitempty"`
	Duration            *time.Duration     `json:"duration,omitempty"`
	State               ExecutionState     `json:"state"`
}

// ExecutionState represents execution state
type ExecutionState int

const (
	ExecutionStatePending ExecutionState = iota
	ExecutionStateRunning
	ExecutionStateCompleted
	ExecutionStatePartial
	ExecutionStateFailed
)

// String returns string representation of execution state
func (es ExecutionState) String() string {
	switch es {
	case ExecutionStatePending:
		return "PENDING"
	case ExecutionStateRunning:
		return "RUNNING"
	case ExecutionStateCompleted:
		return "COMPLETED"
	case ExecutionStatePartial:
		return "PARTIAL"
	case ExecutionStateFailed:
		return "FAILED"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(es))
	}
}

// IsValid checks if execution state is valid
func (es ExecutionState) IsValid() bool {
	return es >= ExecutionStatePending && es <= ExecutionStateFailed
}

// OK returns result as Result[T]
func (cer *ConcurrentExecutionResult) OK() result.Result[*ConcurrentExecutionResult] {
	return result.Ok(cer)
}

// GetSuccessfulOperations returns successful operations
func (cer *ConcurrentExecutionResult) GetSuccessfulOperations() []*OperationResult {
	successful := make([]*OperationResult, 0)
	for _, opResult := range cer.Operations {
		if opResult.State == OperationStateCompleted {
			successful = append(successful, opResult)
		}
	}
	return successful
}

// GetFailedOperations returns failed operations
func (cer *ConcurrentExecutionResult) GetFailedOperations() []*OperationResult {
	failed := make([]*OperationResult, 0)
	for _, opResult := range cer.Operations {
		if opResult.State == OperationStateFailed {
			failed = append(failed, opResult)
		}
	}
	return failed
}

// GetSuccessRate returns success rate as percentage
func (cer *ConcurrentExecutionResult) GetSuccessRate() float64 {
	if cer.TotalOperations == 0 {
		return 0.0
	}
	return float64(cer.CompletedOperations) / float64(cer.TotalOperations) * 100.0
}

// OptimalWorkerCount calculates optimal worker count based on CPU and I/O
func OptimalWorkerCount(cpuIntensive, ioIntensive bool) int {
	cpuCount := runtime.NumCPU()

	if cpuIntensive && !ioIntensive {
		// CPU-bound operations: use CPU count
		return cpuCount
	} else if !cpuIntensive && ioIntensive {
		// I/O-bound operations: use more workers
		return cpuCount * 4
	} else {
		// Mixed operations: moderate worker count
		return cpuCount * 2
	}
}
