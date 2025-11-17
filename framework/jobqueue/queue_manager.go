package jobqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/framework/cache"
	"github.com/mdaxf/iac/framework/logs"
	"github.com/mdaxf/iac/models"

	"github.com/google/uuid"
)

const (
	// Redis keys
	JobQueueKey      = "job:queue"
	JobLockKeyPrefix = "job:lock:"
	JobStatusPrefix  = "job:status:"

	// Lock settings
	DefaultLockTimeout = 5 * time.Minute
	LockRetryDelay     = 100 * time.Millisecond
	MaxLockRetries     = 3
)

// DistributedQueueManager manages job queues using Redis for distributed coordination
type DistributedQueueManager struct {
	cache      cache.Cache
	instanceID string
	logger     logs.Logger
}

// NewDistributedQueueManager creates a new distributed queue manager
func NewDistributedQueueManager(cache cache.Cache) *DistributedQueueManager {
	return &DistributedQueueManager{
		cache:      cache,
		instanceID: uuid.New().String(),
		logger:     logs.Logger{ModuleName: "DistributedQueueManager"},
	}
}

// EnqueueJob adds a job to the distributed queue
func (dqm *DistributedQueueManager) EnqueueJob(ctx context.Context, jobID string, priority int) error {
	startTime := time.Now()
	defer func() {
		dqm.logger.Debug(fmt.Sprintf("EnqueueJob completed in %v", time.Since(startTime)))
	}()

	// Create job queue item with priority and timestamp
	queueItem := map[string]interface{}{
		"job_id":     jobID,
		"priority":   priority,
		"enqueued_at": time.Now().Unix(),
		"instance_id": dqm.instanceID,
	}

	queueItemJSON, err := json.Marshal(queueItem)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	// Use sorted set with priority as score for priority queue
	score := float64(priority*1000000000 - time.Now().Unix()) // Higher priority first, then FIFO

	// In this implementation, we'll use a list for simplicity
	// For production, consider using Redis sorted sets or lists with proper priority handling
	err = dqm.cache.Put(ctx, fmt.Sprintf("%s:%s", JobQueueKey, jobID), string(queueItemJSON), 24*time.Hour)
	if err != nil {
		dqm.logger.Error(fmt.Sprintf("Failed to enqueue job %s: %v", jobID, err))
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	// Add to pending list
	listKey := fmt.Sprintf("%s:pending", JobQueueKey)
	err = dqm.cache.Put(ctx, listKey, jobID, 24*time.Hour)

	dqm.logger.Info(fmt.Sprintf("Enqueued job %s with priority %d", jobID, priority))
	return nil
}

// DequeueJob removes and returns the next job from the queue
func (dqm *DistributedQueueManager) DequeueJob(ctx context.Context) (string, error) {
	listKey := fmt.Sprintf("%s:pending", JobQueueKey)

	// Try to get the first pending job
	jobID, err := dqm.cache.Get(ctx, listKey)
	if err != nil || jobID == nil {
		return "", nil // No jobs available
	}

	jobIDStr, ok := jobID.(string)
	if !ok {
		return "", fmt.Errorf("invalid job ID format")
	}

	// Remove from pending list
	err = dqm.cache.Delete(ctx, listKey)
	if err != nil {
		dqm.logger.Warning(fmt.Sprintf("Failed to remove job %s from pending list: %v", jobIDStr, err))
	}

	dqm.logger.Debug(fmt.Sprintf("Dequeued job %s", jobIDStr))
	return jobIDStr, nil
}

// AcquireLock attempts to acquire a distributed lock for a job
func (dqm *DistributedQueueManager) AcquireLock(ctx context.Context, jobID string, timeout time.Duration) (bool, error) {
	startTime := time.Now()
	defer func() {
		dqm.logger.Debug(fmt.Sprintf("AcquireLock for job %s completed in %v", jobID, time.Since(startTime)))
	}()

	if timeout == 0 {
		timeout = DefaultLockTimeout
	}

	lockKey := JobLockKeyPrefix + jobID

	// Create lock data
	lock := models.JobLock{
		JobID:      jobID,
		InstanceID: dqm.instanceID,
		LockedAt:   time.Now(),
		ExpiresAt:  time.Now().Add(timeout),
	}

	lockJSON, err := json.Marshal(lock)
	if err != nil {
		return false, fmt.Errorf("failed to marshal lock: %w", err)
	}

	// Try to acquire lock with retries
	for i := 0; i < MaxLockRetries; i++ {
		// Check if lock exists
		exists, err := dqm.cache.IsExist(ctx, lockKey)
		if err != nil {
			dqm.logger.Warning(fmt.Sprintf("Failed to check lock existence: %v", err))
			time.Sleep(LockRetryDelay)
			continue
		}

		if !exists {
			// Lock doesn't exist, try to create it
			err = dqm.cache.Put(ctx, lockKey, string(lockJSON), timeout)
			if err == nil {
				dqm.logger.Info(fmt.Sprintf("Acquired lock for job %s (instance: %s)", jobID, dqm.instanceID))
				return true, nil
			}
			dqm.logger.Warning(fmt.Sprintf("Failed to create lock: %v", err))
		} else {
			// Lock exists, check if it's expired
			lockData, err := dqm.cache.Get(ctx, lockKey)
			if err == nil && lockData != nil {
				var existingLock models.JobLock
				if lockDataStr, ok := lockData.(string); ok {
					json.Unmarshal([]byte(lockDataStr), &existingLock)
					if time.Now().After(existingLock.ExpiresAt) {
						// Lock expired, delete and retry
						dqm.cache.Delete(ctx, lockKey)
						dqm.logger.Info(fmt.Sprintf("Removed expired lock for job %s", jobID))
						continue
					}
				}
			}
			dqm.logger.Debug(fmt.Sprintf("Lock already held for job %s", jobID))
		}

		if i < MaxLockRetries-1 {
			time.Sleep(LockRetryDelay * time.Duration(i+1))
		}
	}

	return false, nil
}

// ReleaseLock releases the distributed lock for a job
func (dqm *DistributedQueueManager) ReleaseLock(ctx context.Context, jobID string) error {
	startTime := time.Now()
	defer func() {
		dqm.logger.Debug(fmt.Sprintf("ReleaseLock for job %s completed in %v", jobID, time.Since(startTime)))
	}()

	lockKey := JobLockKeyPrefix + jobID

	// Check if we own the lock
	lockData, err := dqm.cache.Get(ctx, lockKey)
	if err != nil {
		return fmt.Errorf("failed to get lock data: %w", err)
	}

	if lockData != nil {
		var lock models.JobLock
		if lockDataStr, ok := lockData.(string); ok {
			json.Unmarshal([]byte(lockDataStr), &lock)
			if lock.InstanceID != dqm.instanceID {
				dqm.logger.Warning(fmt.Sprintf("Attempted to release lock for job %s owned by different instance", jobID))
				return fmt.Errorf("lock owned by different instance")
			}
		}
	}

	// Delete the lock
	err = dqm.cache.Delete(ctx, lockKey)
	if err != nil {
		dqm.logger.Error(fmt.Sprintf("Failed to release lock for job %s: %v", jobID, err))
		return fmt.Errorf("failed to release lock: %w", err)
	}

	dqm.logger.Info(fmt.Sprintf("Released lock for job %s", jobID))
	return nil
}

// ExtendLock extends the expiration time of a lock
func (dqm *DistributedQueueManager) ExtendLock(ctx context.Context, jobID string, additionalTime time.Duration) error {
	lockKey := JobLockKeyPrefix + jobID

	lockData, err := dqm.cache.Get(ctx, lockKey)
	if err != nil {
		return fmt.Errorf("failed to get lock data: %w", err)
	}

	if lockData == nil {
		return fmt.Errorf("lock does not exist")
	}

	var lock models.JobLock
	if lockDataStr, ok := lockData.(string); ok {
		json.Unmarshal([]byte(lockDataStr), &lock)
		if lock.InstanceID != dqm.instanceID {
			return fmt.Errorf("lock owned by different instance")
		}
	}

	// Extend expiration
	lock.ExpiresAt = lock.ExpiresAt.Add(additionalTime)

	lockJSON, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal lock: %w", err)
	}

	err = dqm.cache.Put(ctx, lockKey, string(lockJSON), time.Until(lock.ExpiresAt))
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	dqm.logger.Debug(fmt.Sprintf("Extended lock for job %s by %v", jobID, additionalTime))
	return nil
}

// SetJobStatus sets the status of a job in cache
func (dqm *DistributedQueueManager) SetJobStatus(ctx context.Context, jobID string, status int) error {
	statusKey := JobStatusPrefix + jobID

	statusData := map[string]interface{}{
		"status":      status,
		"instance_id": dqm.instanceID,
		"updated_at":  time.Now().Unix(),
	}

	statusJSON, err := json.Marshal(statusData)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	err = dqm.cache.Put(ctx, statusKey, string(statusJSON), 1*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to set job status: %w", err)
	}

	return nil
}

// GetJobStatus retrieves the status of a job from cache
func (dqm *DistributedQueueManager) GetJobStatus(ctx context.Context, jobID string) (int, error) {
	statusKey := JobStatusPrefix + jobID

	statusData, err := dqm.cache.Get(ctx, statusKey)
	if err != nil {
		return 0, fmt.Errorf("failed to get job status: %w", err)
	}

	if statusData == nil {
		return 0, fmt.Errorf("status not found")
	}

	var status map[string]interface{}
	if statusDataStr, ok := statusData.(string); ok {
		json.Unmarshal([]byte(statusDataStr), &status)
		if statusInt, ok := status["status"].(float64); ok {
			return int(statusInt), nil
		}
	}

	return 0, fmt.Errorf("invalid status format")
}

// ClearJobData clears all cached data for a job
func (dqm *DistributedQueueManager) ClearJobData(ctx context.Context, jobID string) error {
	// Delete lock
	dqm.cache.Delete(ctx, JobLockKeyPrefix+jobID)

	// Delete status
	dqm.cache.Delete(ctx, JobStatusPrefix+jobID)

	// Delete queue item
	dqm.cache.Delete(ctx, fmt.Sprintf("%s:%s", JobQueueKey, jobID))

	dqm.logger.Debug(fmt.Sprintf("Cleared cache data for job %s", jobID))
	return nil
}

// GetInstanceID returns the instance ID of this queue manager
func (dqm *DistributedQueueManager) GetInstanceID() string {
	return dqm.instanceID
}

// HealthCheck performs a health check on the queue manager
func (dqm *DistributedQueueManager) HealthCheck(ctx context.Context) error {
	testKey := "job:queue:health"
	testValue := time.Now().String()

	// Test write
	err := dqm.cache.Put(ctx, testKey, testValue, 10*time.Second)
	if err != nil {
		return fmt.Errorf("health check write failed: %w", err)
	}

	// Test read
	val, err := dqm.cache.Get(ctx, testKey)
	if err != nil {
		return fmt.Errorf("health check read failed: %w", err)
	}

	if val == nil {
		return fmt.Errorf("health check read returned nil")
	}

	// Cleanup
	dqm.cache.Delete(ctx, testKey)

	return nil
}
