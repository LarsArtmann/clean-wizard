package errorrecovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LarsArtmann/clean-wizard/internal/domain/shared"
	"github.com/LarsArtmann/clean-wizard/internal/shared/errors"
	"github.com/LarsArtmann/clean-wizard/internal/shared/result"
)

// OperationID represents unique operation identifier
type OperationID string

// OperationState represents operation state with type safety
type OperationState int

const (
	OperationStatePending OperationState = iota
	OperationStateRunning
	OperationStateCompleted
	OperationStateFailed
	OperationStateRollback
	OperationStateRolledBack
)

// String returns string representation of operation state
func (os OperationState) String() string {
	switch os {
	case OperationStatePending:
		return "PENDING"
	case OperationStateRunning:
		return "RUNNING"
	case OperationStateCompleted:
		return "COMPLETED"
	case OperationStateFailed:
		return "FAILED"
	case OperationStateRollback:
		return "ROLLBACK"
	case OperationStateRolledBack:
		return "ROLLED_BACK"
	default:
		return fmt.Sprintf("UNKNOWN_%d", int(os))
	}
}

// IsValid checks if operation state is valid
func (os OperationState) IsValid() bool {
	return os >= OperationStatePending && os <= OperationStateRolledBack
}

// OperationRecord represents a recorded operation
type OperationRecord struct {
	ID           OperationID                              `json:"id"`
	Name         string                                   `json:"name"`
	State        OperationState                           `json:"state"`
	StartTime    time.Time                                `json:"start_time"`
	EndTime      *time.Time                               `json:"end_time,omitempty"`
	Duration     *time.Duration                           `json:"duration,omitempty"`
	Result       *shared.CleanResult                      `json:"result,omitempty"`
	Error        *errors.DomainError                      `json:"error,omitempty"`
	RollbackFunc func() result.Result[shared.CleanResult] `json:"-"`
	Context      map[string]interface{}                   `json:"context"`
	Retryable    bool                                     `json:"retryable"`
	CreatedAt    time.Time                                `json:"created_at"`
}

// IsCompleted returns true if operation is completed (either success or failure)
func (or *OperationRecord) IsCompleted() bool {
	return or.State == OperationStateCompleted ||
		or.State == OperationStateFailed ||
		or.State == OperationStateRolledBack
}

// IsSuccess returns true if operation completed successfully
func (or *OperationRecord) IsSuccess() bool {
	return or.State == OperationStateCompleted
}

// IsFailed returns true if operation failed
func (or *OperationRecord) IsFailed() bool {
	return or.State == OperationStateFailed || or.State == OperationStateRolledBack
}

// CanRetry returns true if operation can be retried
func (or *OperationRecord) CanRetry() bool {
	return or.IsFailed() && or.Retryable
}

// GetDuration returns operation duration
func (or *OperationRecord) GetDuration() time.Duration {
	if or.Duration != nil {
		return *or.Duration
	}

	if or.EndTime != nil {
		return or.EndTime.Sub(or.StartTime)
	}

	return time.Since(or.StartTime)
}

// RollbackManager manages operation rollbacks
type RollbackManager struct {
	operations map[OperationID]*OperationRecord
	mutex      sync.RWMutex
	maxHistory int
	logger     errors.Logger
}

// NewRollbackManager creates new rollback manager
func NewRollbackManager(maxHistory int, logger errors.Logger) *RollbackManager {
	return &RollbackManager{
		operations: make(map[OperationID]*OperationRecord),
		maxHistory: maxHistory,
		logger:     logger,
	}
}

// RecordOperation records a new operation
func (rm *RollbackManager) RecordOperation(
	id OperationID,
	name string,
	rollbackFunc func() result.Result[shared.CleanResult],
	context map[string]interface{},
	retryable bool,
) *OperationRecord {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	record := &OperationRecord{
		ID:           id,
		Name:         name,
		State:        OperationStatePending,
		StartTime:    time.Now(),
		RollbackFunc: rollbackFunc,
		Context:      context,
		Retryable:    retryable,
		CreatedAt:    time.Now(),
	}

	rm.operations[id] = record
	rm.cleanupHistory()

	rm.logger.Info("Operation recorded",
		"operation_id", id,
		"operation_name", name,
		"state", record.State.String())

	return record
}

// StartOperation marks operation as started
func (rm *RollbackManager) StartOperation(id OperationID) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	record := rm.operations[id]
	if record == nil {
		return errors.NewErrorFactory("rollback_manager").
			NotFound("operation not found").
			WithDetail("operation_id", id).
			ToResult()
	}

	if record.State != OperationStatePending {
		return errors.NewErrorFactory("rollback_manager").
			InvalidState(fmt.Sprintf("operation is %s, cannot start", record.State.String())).
			WithDetail("operation_id", id).
			WithDetail("current_state", record.State.String()).
			ToResult()
	}

	record.State = OperationStateRunning

	rm.logger.Info("Operation started",
		"operation_id", id,
		"operation_name", record.Name,
		"state", record.State.String())

	return nil
}

// CompleteOperation marks operation as completed successfully
func (rm *RollbackManager) CompleteOperation(
	id OperationID,
	result shared.CleanResult,
) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	record := rm.operations[id]
	if record == nil {
		return errors.NewErrorFactory("rollback_manager").
			NotFound("operation not found").
			WithDetail("operation_id", id).
			ToResult()
	}

	if record.State != OperationStateRunning {
		return errors.NewErrorFactory("rollback_manager").
			InvalidState(fmt.Sprintf("operation is %s, cannot complete", record.State.String())).
			WithDetail("operation_id", id).
			WithDetail("current_state", record.State.String()).
			ToResult()
	}

	endTime := time.Now()
	duration := endTime.Sub(record.StartTime)

	record.State = OperationStateCompleted
	record.EndTime = &endTime
	record.Duration = &duration
	record.Result = &result

	rm.logger.Info("Operation completed successfully",
		"operation_id", id,
		"operation_name", record.Name,
		"state", record.State.String(),
		"duration", duration.String(),
		"items_removed", result.ItemsRemoved,
		"freed_bytes", result.FreedBytes)

	return nil
}

// FailOperation marks operation as failed
func (rm *RollbackManager) FailOperation(
	id OperationID,
	err error,
) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	record := rm.operations[id]
	if record == nil {
		return errors.NewErrorFactory("rollback_manager").
			NotFound("operation not found").
			WithDetail("operation_id", id).
			ToResult()
	}

	if record.State != OperationStateRunning {
		return errors.NewErrorFactory("rollback_manager").
			InvalidState(fmt.Sprintf("operation is %s, cannot fail", record.State.String())).
			WithDetail("operation_id", id).
			WithDetail("current_state", record.State.String()).
			ToResult()
	}

	endTime := time.Now()
	duration := endTime.Sub(record.StartTime)

	record.State = OperationStateFailed
	record.EndTime = &endTime
	record.Duration = &duration

	if domainErr, ok := err.(*errors.DomainError); ok {
		record.Error = domainErr
	} else {
		factory := errors.NewErrorFactory("operation_failure")
		record.Error = factory.Unknown(err.Error())
	}

	rm.logger.Error("Operation failed",
		"operation_id", id,
		"operation_name", record.Name,
		"state", record.State.String(),
		"duration", duration.String(),
		"error", err.Error())

	return nil
}

// RollbackOperation rolls back a failed operation
func (rm *RollbackManager) RollbackOperation(
	ctx context.Context,
	id OperationID,
) result.Result[shared.CleanResult] {
	rm.mutex.Lock()
	record := rm.operations[id]
	if record == nil {
		rm.mutex.Unlock()
		return result.Err[shared.CleanResult](
			errors.NewErrorFactory("rollback_manager").
				NotFound("operation not found").
				WithDetail("operation_id", id).
				ToResult())
	}

	if record.State != OperationStateFailed {
		rm.mutex.Unlock()
		return result.Err[shared.CleanResult](
			errors.NewErrorFactory("rollback_manager").
				InvalidState(fmt.Sprintf("operation is %s, cannot rollback", record.State.String())).
				WithDetail("operation_id", id).
				WithDetail("current_state", record.State.String()).
				ToResult())
	}

	if record.RollbackFunc == nil {
		rm.mutex.Unlock()
		return result.Err[shared.CleanResult](
			errors.NewErrorFactory("rollback_manager").
				Operation("no rollback function available").
				WithDetail("operation_id", id).
				ToResult())
	}

	record.State = OperationStateRollback
	rm.mutex.Unlock()

	rm.logger.Info("Starting operation rollback",
		"operation_id", id,
		"operation_name", record.Name,
		"state", record.State.String())

	// Execute rollback function
	rollbackResult := record.RollbackFunc()

	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	endTime := time.Now()

	if rollbackResult.IsOk() {
		record.State = OperationStateRolledBack
		rm.logger.Info("Operation rollback completed successfully",
			"operation_id", id,
			"operation_name", record.Name,
			"state", record.State.String())
	} else {
		record.State = OperationStateFailed
		rm.logger.Error("Operation rollback failed",
			"operation_id", id,
			"operation_name", record.Name,
			"state", record.State.String(),
			"rollback_error", rollbackResult.Error().Error())
	}

	record.EndTime = &endTime
	record.Duration = &endTime.Sub(record.StartTime)

	return rollbackResult
}

// RetryOperation retries a failed operation
func (rm *RollbackManager) RetryOperation(
	ctx context.Context,
	id OperationID,
	operationFunc func() result.Result[shared.CleanResult],
) result.Result[shared.CleanResult] {
	rm.mutex.Lock()
	record := rm.operations[id]
	if record == nil {
		rm.mutex.Unlock()
		return result.Err[shared.CleanResult](
			errors.NewErrorFactory("rollback_manager").
				NotFound("operation not found").
				WithDetail("operation_id", id).
				ToResult())
	}

	if !record.CanRetry() {
		rm.mutex.Unlock()
		return result.Err[shared.CleanResult](
			errors.NewErrorFactory("rollback_manager").
				Operation("operation cannot be retried").
				WithDetail("operation_id", id).
				WithDetail("current_state", record.State.String()).
				WithDetail("retryable", record.Retryable).
				ToResult())
	}

	// Reset operation state
	record.State = OperationStatePending
	record.StartTime = time.Now()
	record.EndTime = nil
	record.Duration = nil
	record.Result = nil
	record.Error = nil

	rm.mutex.Unlock()

	rm.logger.Info("Retrying operation",
		"operation_id", id,
		"operation_name", record.Name,
		"state", record.State.String())

	// Start operation
	if err := rm.StartOperation(id); err != nil {
		return result.Err[shared.CleanResult](err)
	}

	// Execute operation
	opResult := operationFunc()

	if opResult.IsOk() {
		// Complete operation
		if err := rm.CompleteOperation(id, opResult.Value()); err != nil {
			return result.Err[shared.CleanResult](err)
		}
		return opResult
	} else {
		// Fail operation
		if err := rm.FailOperation(id, opResult.Error()); err != nil {
			return result.Err[shared.CleanResult](err)
		}
		return opResult
	}
}

// GetOperation returns operation record
func (rm *RollbackManager) GetOperation(id OperationID) (*OperationRecord, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	record := rm.operations[id]
	if record == nil {
		return nil, errors.NewErrorFactory("rollback_manager").
			NotFound("operation not found").
			WithDetail("operation_id", id).
			ToResult()
	}

	return record, nil
}

// GetOperations returns all operations
func (rm *RollbackManager) GetOperations() map[OperationID]*OperationRecord {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	operations := make(map[OperationID]*OperationRecord)
	for id, record := range rm.operations {
		operations[id] = record
	}

	return operations
}

// GetPendingOperations returns all pending operations
func (rm *RollbackManager) GetPendingOperations() map[OperationID]*OperationRecord {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	pending := make(map[OperationID]*OperationRecord)
	for id, record := range rm.operations {
		if record.State == OperationStatePending {
			pending[id] = record
		}
	}

	return pending
}

// GetFailedOperations returns all failed operations
func (rm *RollbackManager) GetFailedOperations() map[OperationID]*OperationRecord {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	failed := make(map[OperationID]*OperationRecord)
	for id, record := range rm.operations {
		if record.State == OperationStateFailed || record.State == OperationStateRolledBack {
			failed[id] = record
		}
	}

	return failed
}

// RollbackAllOperations rolls back all failed operations
func (rm *RollbackManager) RollbackAllOperations(
	ctx context.Context,
) map[OperationID]result.Result[shared.CleanResult] {
	failed := rm.GetFailedOperations()
	results := make(map[OperationID]result.Result[shared.CleanResult])

	for id, record := range failed {
		if record.State == OperationStateFailed {
			results[id] = rm.RollbackOperation(ctx, id)
		} else {
			results[id] = result.Err[shared.CleanResult](
				errors.NewErrorFactory("rollback_manager").
					InvalidState("operation already rolled back").
					WithDetail("operation_id", id).
					ToResult())
		}
	}

	return results
}

// Cleanup removes old operations
func (rm *RollbackManager) Cleanup(olderThan time.Duration) int {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	cutoff := time.Now().Add(-olderThan)
	removed := 0

	for id, record := range rm.operations {
		if record.IsCompleted() && record.CreatedAt.Before(cutoff) {
			delete(rm.operations, id)
			removed++
		}
	}

	rm.logger.Info("Operation history cleaned up",
		"removed_count", removed,
		"cutoff_time", cutoff.String())

	return removed
}

// GetStatistics returns rollback manager statistics
func (rm *RollbackManager) GetStatistics() RollbackStatistics {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	stats := RollbackStatistics{
		TotalOperations:     len(rm.operations),
		PendingOperations:   0,
		RunningOperations:   0,
		CompletedOperations: 0,
		FailedOperations:    0,
		RollbackOperations:  0,
	}

	for _, record := range rm.operations {
		switch record.State {
		case OperationStatePending:
			stats.PendingOperations++
		case OperationStateRunning:
			stats.RunningOperations++
		case OperationStateCompleted:
			stats.CompletedOperations++
		case OperationStateFailed:
			stats.FailedOperations++
		case OperationStateRollback, OperationStateRolledBack:
			stats.RollbackOperations++
		}
	}

	return stats
}

// cleanupHistory removes old operations to maintain max history
func (rm *RollbackManager) cleanupHistory() {
	if len(rm.operations) <= rm.maxHistory {
		return
	}

	// Sort operations by creation time
	type operationTime struct {
		id       OperationID
		createAt time.Time
	}

	times := make([]operationTime, 0, len(rm.operations))
	for id, record := range rm.operations {
		times = append(times, operationTime{id, record.CreatedAt})
	}

	// Sort by creation time (oldest first)
	for i := 0; i < len(times)-1; i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i].createAt.After(times[j].createAt) {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	// Remove oldest operations
	removeCount := len(rm.operations) - rm.maxHistory
	for i := 0; i < removeCount; i++ {
		delete(rm.operations, times[i].id)
	}

	rm.logger.Info("Operation history truncated",
		"removed_count", removeCount,
		"current_count", len(rm.operations),
		"max_history", rm.maxHistory)
}

// RollbackStatistics represents rollback manager statistics
type RollbackStatistics struct {
	TotalOperations     int `json:"total_operations"`
	PendingOperations   int `json:"pending_operations"`
	RunningOperations   int `json:"running_operations"`
	CompletedOperations int `json:"completed_operations"`
	FailedOperations    int `json:"failed_operations"`
	RollbackOperations  int `json:"rollback_operations"`
}

// GenerateOperationID generates unique operation ID
func GenerateOperationID() OperationID {
	return OperationID(fmt.Sprintf("op_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()))
}
