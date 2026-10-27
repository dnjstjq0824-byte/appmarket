package worker

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
)

type JobStatus string

const (
	JobStatusQueued          JobStatus = "QUEUED"
	JobStatusProcessing      JobStatus = "PROCESSING"
	JobStatusCompleted       JobStatus = "COMPLETED"
	JobStatusRetryScheduled  JobStatus = "RETRY_SCHEDULED"
	JobStatusDeadLetter      JobStatus = "DEAD_LETTER"
)

type BackgroundJob struct {
	ID             string     `json:"id"`
	QueueName      string     `json:"queue_name"`
	JobType        string     `json:"job_type"`
	PayloadJSON    string     `json:"payload_json"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
	Status         JobStatus  `json:"status"`
	Attempts       int        `json:"attempts"`
	MaxAttempts    int        `json:"max_attempts"`
	ScheduledAt    time.Time  `json:"scheduled_at"`
	LockedAt       *time.Time `json:"locked_at,omitempty"`
	LockedBy       string     `json:"locked_by,omitempty"`
	LastHeartbeat  time.Time  `json:"last_heartbeat"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type JobRepository interface {
	Enqueue(ctx context.Context, job *BackgroundJob) (*BackgroundJob, bool, error)
	AcquireNextJob(ctx context.Context, queueName, workerID string, now time.Time) (*BackgroundJob, error)
	UpdateHeartbeat(ctx context.Context, jobID string, now time.Time) error
	CompleteJob(ctx context.Context, jobID string, now time.Time) error
	FailJob(ctx context.Context, jobID, errStr string, now time.Time) (*BackgroundJob, error)
	GetJobByID(ctx context.Context, jobID string) (*BackgroundJob, error)
	ListDeadLetterJobs(ctx context.Context, limit int) ([]*BackgroundJob, error)
	ResetDeadLetterJob(ctx context.Context, jobID string) error
}

type InMemoryJobRepository struct {
	mu             sync.RWMutex
	jobs           map[string]*BackgroundJob
	idempotencyMap map[string]string // idempotency_key -> job_id
}

func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{
		jobs:           make(map[string]*BackgroundJob),
		idempotencyMap: make(map[string]string),
	}
}

// Enqueue persists a job with idempotency deduplication (TC-FUNC-006).
func (r *InMemoryJobRepository) Enqueue(ctx context.Context, job *BackgroundJob) (*BackgroundJob, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if job.IdempotencyKey != "" {
		if existingID, exists := r.idempotencyMap[job.IdempotencyKey]; exists {
			if existing, ok := r.jobs[existingID]; ok {
				return existing, false, nil // deduplicated!
			}
		}
	}

	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.MaxAttempts <= 0 {
		job.MaxAttempts = 5
	}
	if job.Status == "" {
		job.Status = JobStatusQueued
	}
	if job.ScheduledAt.IsZero() {
		job.ScheduledAt = time.Now().UTC()
	}
	job.CreatedAt = time.Now().UTC()
	job.LastHeartbeat = job.CreatedAt

	r.jobs[job.ID] = job
	if job.IdempotencyKey != "" {
		r.idempotencyMap[job.IdempotencyKey] = job.ID
	}

	return job, true, nil
}

// AcquireNextJob simulates FOR UPDATE SKIP LOCKED to grab the next ready job.
func (r *InMemoryJobRepository) AcquireNextJob(ctx context.Context, queueName, workerID string, now time.Time) (*BackgroundJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, job := range r.jobs {
		if (queueName == "" || job.QueueName == queueName) &&
			(job.Status == JobStatusQueued || job.Status == JobStatusRetryScheduled) &&
			!job.ScheduledAt.After(now) {

			job.Status = JobStatusProcessing
			job.LockedAt = &now
			job.LockedBy = workerID
			job.LastHeartbeat = now
			job.Attempts++

			c := *job
			return &c, nil
		}
	}

	return nil, nil // no job available
}

func (r *InMemoryJobRepository) UpdateHeartbeat(ctx context.Context, jobID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}
	job.LastHeartbeat = now
	return nil
}

func (r *InMemoryJobRepository) CompleteJob(ctx context.Context, jobID string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}
	job.Status = JobStatusCompleted
	job.CompletedAt = &now
	job.LockedBy = ""
	return nil
}

func (r *InMemoryJobRepository) FailJob(ctx context.Context, jobID, errStr string, now time.Time) (*BackgroundJob, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	job.ErrorMessage = errStr
	job.LockedBy = ""

	if job.Attempts >= job.MaxAttempts {
		job.Status = JobStatusDeadLetter
	} else {
		job.Status = JobStatusRetryScheduled
		// Exponential backoff: base 2 seconds * (2 ^ (attempts - 1))
		backoffSeconds := int(math.Pow(2, float64(job.Attempts-1)) * 2)
		job.ScheduledAt = now.Add(time.Duration(backoffSeconds) * time.Second)
	}

	c := *job
	return &c, nil
}

func (r *InMemoryJobRepository) GetJobByID(ctx context.Context, jobID string) (*BackgroundJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	c := *job
	return &c, nil
}

func (r *InMemoryJobRepository) ListDeadLetterJobs(ctx context.Context, limit int) ([]*BackgroundJob, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var dlq []*BackgroundJob
	for _, job := range r.jobs {
		if job.Status == JobStatusDeadLetter {
			c := *job
			dlq = append(dlq, &c)
			if len(dlq) >= limit {
				break
			}
		}
	}
	return dlq, nil
}

func (r *InMemoryJobRepository) ResetDeadLetterJob(ctx context.Context, jobID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	job, ok := r.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}
	if job.Status != JobStatusDeadLetter {
		return fmt.Errorf("job is not in DEAD_LETTER status")
	}

	job.Status = JobStatusQueued
	job.Attempts = 0
	job.ScheduledAt = time.Now().UTC()
	job.ErrorMessage = ""
	return nil
}

type JobHandler func(ctx context.Context, job *BackgroundJob) error

type Dispatcher struct {
	repo     JobRepository
	handlers map[string]JobHandler
	workerID string
	mu       sync.RWMutex
}

func NewDispatcher(repo JobRepository, workerID string) *Dispatcher {
	if workerID == "" {
		workerID = "worker-" + uuid.New().String()[:8]
	}
	return &Dispatcher{
		repo:     repo,
		handlers: make(map[string]JobHandler),
		workerID: workerID,
	}
}

func (d *Dispatcher) RegisterHandler(jobType string, handler JobHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[jobType] = handler
}

// ExecuteNext processes one ready job from the given queue.
func (d *Dispatcher) ExecuteNext(ctx context.Context, queueName string) (bool, error) {
	now := time.Now().UTC()
	job, err := d.repo.AcquireNextJob(ctx, queueName, d.workerID, now)
	if err != nil || job == nil {
		return false, err
	}

	d.mu.RLock()
	handler, ok := d.handlers[job.JobType]
	d.mu.RUnlock()

	if !ok {
		errStr := fmt.Sprintf("no registered handler for job type: %s", job.JobType)
		_, _ = d.repo.FailJob(ctx, job.ID, errStr, time.Now().UTC())
		return true, fmt.Errorf("%s", errStr)
	}

	execErr := handler(ctx, job)
	completedTime := time.Now().UTC()

	if execErr != nil {
		_, _ = d.repo.FailJob(ctx, job.ID, execErr.Error(), completedTime)
		return true, execErr
	}

	_ = d.repo.CompleteJob(ctx, job.ID, completedTime)
	return true, nil
}
