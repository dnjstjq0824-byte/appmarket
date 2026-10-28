package worker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ZombieReaper struct {
	repo             JobRepository
	heartbeatTimeout time.Duration
	mu               sync.Mutex
}

func NewZombieReaper(repo JobRepository, heartbeatTimeout time.Duration) *ZombieReaper {
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = 60 * time.Second
	}
	return &ZombieReaper{
		repo:             repo,
		heartbeatTimeout: heartbeatTimeout,
	}
}

// ReapZombies scans for jobs whose workers stopped heartbeating for > 60 seconds (TC-OPS-004).
func (r *ZombieReaper) ReapZombies(ctx context.Context, now time.Time) (int, []string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	inMem, ok := r.repo.(*InMemoryJobRepository)
	if !ok {
		return 0, nil, fmt.Errorf("repository does not support direct zombie inspection")
	}

	inMem.mu.Lock()
	defer inMem.mu.Unlock()

	var reapedIDs []string
	cutoff := now.Add(-r.heartbeatTimeout)

	for _, job := range inMem.jobs {
		if job.Status == JobStatusProcessing && job.LastHeartbeat.Before(cutoff) {
			reapedIDs = append(reapedIDs, job.ID)
			job.LockedBy = ""
			job.LockedAt = nil
			job.ErrorMessage = fmt.Sprintf("zombie job reaped: worker heartbeat lost since %s", job.LastHeartbeat.Format(time.RFC3339))

			if job.Attempts >= job.MaxAttempts {
				job.Status = JobStatusDeadLetter
			} else {
				job.Status = JobStatusRetryScheduled
				job.ScheduledAt = now // immediately eligible for next worker
			}
		}
	}

	return len(reapedIDs), reapedIDs, nil
}

// RetryDeadLetterJob allows manual retry of a dead letter job by an administrator.
func (r *ZombieReaper) RetryDeadLetterJob(ctx context.Context, jobID string) error {
	return r.repo.ResetDeadLetterJob(ctx, jobID)
}

