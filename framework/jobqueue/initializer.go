package jobqueue

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/cache"
	"github.com/mdaxf/iac/framework/logs"
	"github.com/mdaxf/iac-signalr/signalr"
)

var (
	// Global instances
	GlobalQueueManager   *DistributedQueueManager
	GlobalJobWorker      *JobWorker
	GlobalJobScheduler   *JobScheduler
	GlobalJobCreator     *IntegrationJobCreator
	JobSystemInitialized bool
)

// InitializeJobSystem initializes the background job system
func InitializeJobSystem(
	ctx context.Context,
	db *sql.DB,
	docDB *documents.DocDB,
	cacheInstance cache.Cache,
	signalRClient signalr.Client,
) error {

	logger := logs.Logger{ModuleName: "JobSystemInitializer"}

	// Check if job system is enabled
	if !config.GlobalConfiguration.JobsConfig.Enabled {
		logger.Info("Job system is disabled in configuration")
		return nil
	}

	logger.Info("Initializing background job system...")

	// Validate configuration
	if err := validateJobConfiguration(); err != nil {
		return fmt.Errorf("invalid job configuration: %w", err)
	}

	// Initialize distributed queue manager with any available cache
	// This works with Redis, Memcache, or any cache.Cache implementation
	if cacheInstance != nil {
		GlobalQueueManager = NewDistributedQueueManager(cacheInstance)
		logger.Info(fmt.Sprintf("Initialized distributed queue manager with configured cache (instance: %s)", GlobalQueueManager.GetInstanceID()))

		// Health check
		if err := GlobalQueueManager.HealthCheck(ctx); err != nil {
			logger.Warning(fmt.Sprintf("Queue manager health check failed: %v - continuing without cache", err))
			logger.Warning("Job system will run in single-instance mode without distributed locking")
			GlobalQueueManager = nil
		} else {
			cacheType := "configured cache"
			if config.GlobalConfiguration.JobsConfig.UseRedis {
				cacheType = "Redis"
			}
			logger.Info(fmt.Sprintf("Queue manager using %s for distributed coordination", cacheType))
		}
	} else {
		logger.Warning("No cache configured - job system will run in single-instance mode")
		logger.Warning("For distributed processing across multiple instances, configure a cache (Redis, Memcache, etc.)")
	}

	// Initialize job worker
	workerID := fmt.Sprintf("worker-%s", config.GlobalConfiguration.InstanceName)
	GlobalJobWorker = NewJobWorker(workerID, db, docDB, signalRClient, GlobalQueueManager)

	// Start the worker
	if err := GlobalJobWorker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start job worker: %w", err)
	}
	logger.Info(fmt.Sprintf("Started job worker with %d workers", config.GlobalConfiguration.JobsConfig.Workers))

	// Initialize job scheduler
	GlobalJobScheduler = NewJobScheduler(db, GlobalQueueManager)

	// Start the scheduler
	if err := GlobalJobScheduler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start job scheduler: %w", err)
	}
	logger.Info("Started job scheduler")

	// Initialize integration job creator
	GlobalJobCreator = NewIntegrationJobCreator(db, GlobalQueueManager)
	logger.Info("Initialized integration job creator")

	JobSystemInitialized = true
	logger.Info("Background job system initialized successfully")

	// Print status
	printJobSystemStatus(logger)

	return nil
}

// ShutdownJobSystem gracefully shuts down the job system
func ShutdownJobSystem() error {
	logger := logs.Logger{ModuleName: "JobSystemInitializer"}

	if !JobSystemInitialized {
		logger.Info("Job system not initialized, skipping shutdown")
		return nil
	}

	logger.Info("Shutting down background job system...")

	// Stop worker
	if GlobalJobWorker != nil {
		if err := GlobalJobWorker.Stop(); err != nil {
			logger.Error(fmt.Sprintf("Error stopping job worker: %v", err))
		}
	}

	// Stop scheduler
	if GlobalJobScheduler != nil {
		if err := GlobalJobScheduler.Stop(); err != nil {
			logger.Error(fmt.Sprintf("Error stopping job scheduler: %v", err))
		}
	}

	JobSystemInitialized = false
	logger.Info("Background job system shut down successfully")

	return nil
}

// validateJobConfiguration validates the job system configuration
func validateJobConfiguration() error {
	if config.GlobalConfiguration.JobsConfig.Workers <= 0 {
		config.GlobalConfiguration.JobsConfig.Workers = 5
	}

	if config.GlobalConfiguration.JobsConfig.PollInterval <= 0 {
		config.GlobalConfiguration.JobsConfig.PollInterval = 5
	}

	if config.GlobalConfiguration.JobsConfig.MaxRetries < 0 {
		config.GlobalConfiguration.JobsConfig.MaxRetries = 3
	}

	if config.GlobalConfiguration.JobsConfig.SchedulerCheckInterval <= 0 {
		config.GlobalConfiguration.JobsConfig.SchedulerCheckInterval = 60
	}

	if config.GlobalConfiguration.JobsConfig.JobHistoryRetentionDays <= 0 {
		config.GlobalConfiguration.JobsConfig.JobHistoryRetentionDays = 90
	}

	return nil
}

// printJobSystemStatus prints the current status of the job system
func printJobSystemStatus(logger logs.Logger) {
	logger.Info("========================================")
	logger.Info("Background Job System Status")
	logger.Info("========================================")
	logger.Info(fmt.Sprintf("Enabled: %v", config.GlobalConfiguration.JobsConfig.Enabled))
	logger.Info(fmt.Sprintf("Workers: %d", config.GlobalConfiguration.JobsConfig.Workers))
	logger.Info(fmt.Sprintf("Poll Interval: %d seconds", config.GlobalConfiguration.JobsConfig.PollInterval))
	logger.Info(fmt.Sprintf("Max Retries: %d", config.GlobalConfiguration.JobsConfig.MaxRetries))
	logger.Info(fmt.Sprintf("Use Redis: %v", config.GlobalConfiguration.JobsConfig.UseRedis))

	if GlobalJobWorker != nil {
		status := GlobalJobWorker.GetStatus()
		logger.Info(fmt.Sprintf("Worker Status: Running=%v", status["running"]))
	}

	if GlobalJobScheduler != nil {
		status := GlobalJobScheduler.GetStatus()
		logger.Info(fmt.Sprintf("Scheduler Status: Running=%v, Jobs=%v", status["running"], status["scheduled_jobs"]))
	}

	if GlobalQueueManager != nil {
		logger.Info(fmt.Sprintf("Queue Manager Instance: %s", GlobalQueueManager.GetInstanceID()))
	}

	logger.Info("========================================")
}

// GetJobSystemStatus returns the current status of the job system
func GetJobSystemStatus() map[string]interface{} {
	status := make(map[string]interface{})

	status["initialized"] = JobSystemInitialized
	status["enabled"] = config.GlobalConfiguration.JobsConfig.Enabled

	if GlobalJobWorker != nil {
		status["worker"] = GlobalJobWorker.GetStatus()
	}

	if GlobalJobScheduler != nil {
		status["scheduler"] = GlobalJobScheduler.GetStatus()
	}

	if GlobalQueueManager != nil {
		status["queue_manager"] = map[string]interface{}{
			"instance_id": GlobalQueueManager.GetInstanceID(),
		}
	}

	return status
}

// IsJobSystemRunning returns whether the job system is running
func IsJobSystemRunning() bool {
	if !JobSystemInitialized {
		return false
	}

	if GlobalJobWorker != nil && !GlobalJobWorker.IsRunning() {
		return false
	}

	if GlobalJobScheduler != nil && !GlobalJobScheduler.IsRunning() {
		return false
	}

	return true
}
