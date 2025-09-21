package cronjobs

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// CronJob represents a scheduled job
type CronJob struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Schedule    string                 `json:"schedule" db:"schedule"` // Cron expression
	Command     string                 `json:"command" db:"command"`
	Args        []string               `json:"args" db:"args"`
	Environment map[string]string      `json:"environment" db:"environment"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	Timeout     time.Duration          `json:"timeout" db:"timeout"`
	Retries     int                    `json:"retries" db:"retries"`
	TenantID    string                 `json:"tenant_id" db:"tenant_id"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	LastRun     *time.Time             `json:"last_run" db:"last_run"`
	NextRun     *time.Time             `json:"next_run" db:"next_run"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// Execution represents a job execution
type Execution struct {
	ID          string                 `json:"id" db:"id"`
	JobID       string                 `json:"job_id" db:"job_id"`
	Status      string                 `json:"status" db:"status"` // scheduled, running, completed, failed, timeout
	StartedAt   time.Time              `json:"started_at" db:"started_at"`
	FinishedAt  *time.Time             `json:"finished_at" db:"finished_at"`
	Duration    time.Duration          `json:"duration" db:"duration"`
	ExitCode    *int                   `json:"exit_code" db:"exit_code"`
	Output      string                 `json:"output" db:"output"`
	Error       string                 `json:"error" db:"error"`
	Attempt     int                    `json:"attempt" db:"attempt"`
	TriggerType string                 `json:"trigger_type" db:"trigger_type"` // scheduled, manual
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// Storage interface for CronJob persistence
type Storage interface {
	// Job management
	CreateJob(ctx context.Context, job *CronJob) error
	GetJob(ctx context.Context, id string) (*CronJob, error)
	UpdateJob(ctx context.Context, job *CronJob) error
	DeleteJob(ctx context.Context, id string) error
	ListJobs(ctx context.Context, tenantID string) ([]*CronJob, error)
	GetEnabledJobs(ctx context.Context) ([]*CronJob, error)

	// Execution management
	CreateExecution(ctx context.Context, execution *Execution) error
	GetExecution(ctx context.Context, id string) (*Execution, error)
	UpdateExecution(ctx context.Context, execution *Execution) error
	ListJobExecutions(ctx context.Context, jobID string, limit int) ([]*Execution, error)
	CleanupOldExecutions(ctx context.Context, retentionDays int) error
}

// Scheduler manages cron job scheduling
type Scheduler struct {
	storage    Storage
	logger     interface{}
	jobs       map[string]*CronJob
	timers     map[string]*time.Timer
	executions map[string]*Execution
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wsHub      interface{} // WebSocket hub for real-time updates
}

// NewScheduler creates a new cron job scheduler
func NewScheduler(storage Storage, logger interface{}, wsHub interface{}) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		storage:    storage,
		logger:     logger,
		jobs:       make(map[string]*CronJob),
		timers:     make(map[string]*time.Timer),
		executions: make(map[string]*Execution),
		ctx:        ctx,
		cancel:     cancel,
		wsHub:      wsHub,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	// Load enabled jobs from storage
	jobs, err := s.storage.GetEnabledJobs(s.ctx)
	if err != nil {
		return fmt.Errorf("failed to load jobs: %w", err)
	}

	// Schedule all enabled jobs
	for _, job := range jobs {
		s.scheduleJob(job)
	}

	// Start cleanup routine
	go s.cleanupRoutine()

	// Start monitoring routine
	go s.monitoringRoutine()

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel all timers
	for _, timer := range s.timers {
		timer.Stop()
	}

	s.timers = make(map[string]*time.Timer)
}

// AddJob adds a new cron job
func (s *Scheduler) AddJob(job *CronJob) error {
	job.ID = generateJobID()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()

	// Calculate next run time
	nextRun, err := s.calculateNextRun(job.Schedule)
	if err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}
	job.NextRun = &nextRun

	// Save to storage
	if err := s.storage.CreateJob(s.ctx, job); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	// Schedule if enabled
	if job.Enabled {
		s.scheduleJob(job)
	}

	return nil
}

// UpdateJob updates a cron job
func (s *Scheduler) UpdateJob(jobID string, updates *CronJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current job
	currentJob, err := s.storage.GetJob(s.ctx, jobID)
	if err != nil {
		return fmt.Errorf("job not found: %w", err)
	}

	// Update fields
	updates.ID = jobID
	updates.CreatedAt = currentJob.CreatedAt
	updates.UpdatedAt = time.Now()

	// Recalculate next run if schedule changed
	if updates.Schedule != currentJob.Schedule {
		nextRun, err := s.calculateNextRun(updates.Schedule)
		if err != nil {
			return fmt.Errorf("invalid schedule: %w", err)
		}
		updates.NextRun = &nextRun
	}

	// Save to storage
	if err := s.storage.UpdateJob(s.ctx, updates); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Reschedule
	s.unscheduleJob(jobID)
	if updates.Enabled {
		s.scheduleJob(updates)
	}

	return nil
}

// DeleteJob deletes a cron job
func (s *Scheduler) DeleteJob(jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from scheduler
	s.unscheduleJob(jobID)

	// Delete from storage
	if err := s.storage.DeleteJob(s.ctx, jobID); err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	return nil
}

// TriggerJob manually triggers a job execution
func (s *Scheduler) TriggerJob(jobID string) (*Execution, error) {
	job, err := s.storage.GetJob(s.ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	execution := &Execution{
		ID:          generateExecutionID(),
		JobID:       jobID,
		Status:      "scheduled",
		StartedAt:   time.Now(),
		Attempt:     1,
		TriggerType: "manual",
		Metadata:    make(map[string]interface{}),
	}

	if err := s.storage.CreateExecution(s.ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Execute job in background
	go s.executeJob(job, execution)

	return execution, nil
}

// scheduleJob schedules a job for execution
func (s *Scheduler) scheduleJob(job *CronJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Calculate next run time
	nextRun, err := s.calculateNextRun(job.Schedule)
	if err != nil {
		return
	}

	// Update job's next run time
	job.NextRun = &nextRun
	s.jobs[job.ID] = job

	// Schedule timer
	duration := time.Until(nextRun)
	timer := time.AfterFunc(duration, func() {
		s.executeScheduledJob(job.ID)
	})
	s.timers[job.ID] = timer
}

// unscheduleJob removes a job from scheduling
func (s *Scheduler) unscheduleJob(jobID string) {
	if timer, exists := s.timers[jobID]; exists {
		timer.Stop()
		delete(s.timers, jobID)
	}
	delete(s.jobs, jobID)
}

// executeScheduledJob executes a scheduled job
func (s *Scheduler) executeScheduledJob(jobID string) {
	s.mu.RLock()
	job, exists := s.jobs[jobID]
	s.mu.RUnlock()

	if !exists {
		return
	}

	// Create execution record
	execution := &Execution{
		ID:          generateExecutionID(),
		JobID:       jobID,
		Status:      "scheduled",
		StartedAt:   time.Now(),
		Attempt:     1,
		TriggerType: "scheduled",
		Metadata:    make(map[string]interface{}),
	}

	if err := s.storage.CreateExecution(s.ctx, execution); err != nil {
		return
	}

	// Execute job
	go s.executeJob(job, execution)

	// Reschedule for next run
	s.scheduleJob(job)
}

// executeJob executes a job with retry logic
func (s *Scheduler) executeJob(job *CronJob, execution *Execution) {
	// Update execution status
	execution.Status = "running"
	s.storage.UpdateExecution(s.ctx, execution)

	// Broadcast status update via WebSocket
	s.broadcastJobUpdate(job.ID, "running", execution)

	// Execute with timeout
	ctx, cancel := context.WithTimeout(s.ctx, job.Timeout)
	defer cancel()

	// Prepare command
	cmd := exec.CommandContext(ctx, job.Command, job.Args...)

	// Set environment variables
	for key, value := range job.Environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	finishedAt := time.Now()
	execution.FinishedAt = &finishedAt
	execution.Duration = finishedAt.Sub(execution.StartedAt)
	execution.Output = string(output)

	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()

		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			execution.ExitCode = &exitCode
		}

		// Retry if attempts remaining
		if execution.Attempt < job.Retries {
			execution.Attempt++
			execution.Status = "retrying"
			s.storage.UpdateExecution(s.ctx, execution)

			// Wait before retry (exponential backoff)
			retryDelay := time.Duration(execution.Attempt) * time.Second
			time.Sleep(retryDelay)

			// Retry execution
			s.executeJob(job, execution)
			return
		}
	} else {
		execution.Status = "completed"
		exitCode := 0
		execution.ExitCode = &exitCode
	}

	// Update execution record
	s.storage.UpdateExecution(s.ctx, execution)

	// Update job's last run time
	job.LastRun = &execution.StartedAt
	s.storage.UpdateJob(s.ctx, job)

	// Broadcast completion status
	s.broadcastJobUpdate(job.ID, execution.Status, execution)
}

// calculateNextRun calculates the next run time based on cron expression
func (s *Scheduler) calculateNextRun(schedule string) (time.Time, error) {
	// Simple implementation - in production, use a proper cron library
	// For now, support basic intervals

	now := time.Now()

	switch schedule {
	case "*/5 * * * *": // Every 5 minutes
		return now.Add(5 * time.Minute), nil
	case "0 * * * *": // Every hour
		return now.Truncate(time.Hour).Add(time.Hour), nil
	case "0 0 * * *": // Daily at midnight
		return now.Truncate(24 * time.Hour).Add(24 * time.Hour), nil
	case "0 2 * * *": // Daily at 2 AM
		next := now.Truncate(24 * time.Hour).Add(2 * time.Hour)
		if next.Before(now) {
			next = next.Add(24 * time.Hour)
		}
		return next, nil
	default:
		// Default to 1 hour from now
		return now.Add(time.Hour), nil
	}
}

// cleanupRoutine periodically cleans up old execution records
func (s *Scheduler) cleanupRoutine() {
	ticker := time.NewTicker(24 * time.Hour) // Daily cleanup
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.storage.CleanupOldExecutions(s.ctx, 30) // Keep 30 days
		}
	}
}

// monitoringRoutine monitors for stuck executions
func (s *Scheduler) monitoringRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkStuckExecutions()
		}
	}
}

// checkStuckExecutions checks for and handles stuck executions
func (s *Scheduler) checkStuckExecutions() {
	s.mu.RLock()
	executions := make([]*Execution, 0, len(s.executions))
	for _, exec := range s.executions {
		executions = append(executions, exec)
	}
	s.mu.RUnlock()

	for _, execution := range executions {
		if execution.Status == "running" {
			duration := time.Since(execution.StartedAt)
			if duration > 30*time.Minute { // Consider stuck after 30 minutes
				execution.Status = "timeout"
				execution.Error = "Execution timeout - killed by monitoring"
				finishedAt := time.Now()
				execution.FinishedAt = &finishedAt
				execution.Duration = duration

				s.storage.UpdateExecution(s.ctx, execution)
				s.broadcastJobUpdate(execution.JobID, "timeout", execution)
			}
		}
	}
}

// broadcastJobUpdate broadcasts job status updates via WebSocket
func (s *Scheduler) broadcastJobUpdate(jobID, status string, execution *Execution) {
	// In real implementation, this would use the WebSocket hub
	// For now, just log the update
	fmt.Printf("Job %s status: %s\n", jobID, status)
}

// Helper functions

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().UnixNano())
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}
