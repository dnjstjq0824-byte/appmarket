package worker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bookingnetwork/appmarket/backend/internal/worker"
)

// TC-FUNC-006: Outbox Event Relay and Idempotent Job Execution (10 duplicates -> 1 execution)
func TestOutboxRelayAndIdempotentJobExecution(t *testing.T) {
	ctx := context.Background()

	// 1. Test Outbox Event Relay to 6 core queues
	outboxRepo := worker.NewInMemoryOutboxRepository()
	streamPub := worker.NewInMemoryStreamPublisher()
	relayWorker := worker.NewOutboxRelayWorker(outboxRepo, streamPub)

	ev1 := worker.CreateOutboxEvent("APK", "apk-123", "APK_UPLOADED", `{"file":"app.apk"}`)
	ev2 := worker.CreateOutboxEvent("CATALOG", "cat-456", "APP_CATALOG_UPDATED", `{"name":"Cool App"}`)
	ev3 := worker.CreateOutboxEvent("ORDER", "ord-789", "ORDER_COMPLETED", `{"amount":999}`)

	outboxRepo.AddEvent(ev1)
	outboxRepo.AddEvent(ev2)
	outboxRepo.AddEvent(ev3)

	count, err := relayWorker.ProcessBatch(ctx, 10)
	if err != nil {
		t.Fatalf("relay batch failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 events relayed, got %d", count)
	}

	if len(streamPub.Messages[worker.QueueAPKAnalysis]) != 1 {
		t.Fatalf("expected 1 event in queue:apk_analysis")
	}
	if len(streamPub.Messages[worker.QueueSearchIndexer]) != 1 {
		t.Fatalf("expected 1 event in queue:search_indexer")
	}
	if len(streamPub.Messages[worker.QueueAdBilling]) != 1 {
		t.Fatalf("expected 1 event in queue:ad_billing")
	}

	// 2. Test Idempotency Guard: Submit same idempotency key 10 times
	jobRepo := worker.NewInMemoryJobRepository()
	dispatcher := worker.NewDispatcher(jobRepo, "worker-alpha")

	var executionCount int32
	dispatcher.RegisterHandler("PROCESS_PAYMENT", func(ctx context.Context, job *worker.BackgroundJob) error {
		atomic.AddInt32(&executionCount, 1)
		return nil
	})

	idempotencyKey := "idempotent-order-999-settle"
	for i := 1; i <= 10; i++ {
		_, _, err := jobRepo.Enqueue(ctx, &worker.BackgroundJob{
			QueueName:      worker.QueueAdBilling,
			JobType:        "PROCESS_PAYMENT",
			PayloadJSON:    `{"order_id":999}`,
			IdempotencyKey: idempotencyKey,
			ScheduledAt:    time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("enqueue failed on iteration %d: %v", i, err)
		}
	}

	// Dispatch the job
	executed, err := dispatcher.ExecuteNext(ctx, worker.QueueAdBilling)
	if err != nil || !executed {
		t.Fatalf("expected job to execute: executed=%v, err=%v", executed, err)
	}

	// Attempting to execute again from queue: no more jobs should be ready
	executedAgain, err := dispatcher.ExecuteNext(ctx, worker.QueueAdBilling)
	if err != nil {
		t.Fatalf("unexpected error on second acquire: %v", err)
	}
	if executedAgain {
		t.Fatalf("expected no additional job to execute due to deduplication")
	}

	// VERIFY: Executed exactly once!
	if atomic.LoadInt32(&executionCount) != 1 {
		t.Fatalf("expected exactly 1 execution for 10 idempotency submissions, got %d", executionCount)
	}
}

// TC-OPS-004: Zombie Job Heartbeat Timeout Reclamation & DLQ Retry
func TestZombieJobReaperAndDLQRetry(t *testing.T) {
	ctx := context.Background()
	jobRepo := worker.NewInMemoryJobRepository()
	reaper := worker.NewZombieReaper(jobRepo, 60*time.Second)

	now := time.Now().UTC()

	// 1. Enqueue a job
	job, _, err := jobRepo.Enqueue(ctx, &worker.BackgroundJob{
		QueueName:   worker.QueueAPKAnalysis,
		JobType:     "VIRUS_SCAN",
		PayloadJSON: `{"apk_id":"apk-virus-scan-1"}`,
		ScheduledAt: now,
		MaxAttempts: 3,
	})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	// 2. Worker 1 acquires the job
	acquired, err := jobRepo.AcquireNextJob(ctx, worker.QueueAPKAnalysis, "worker-crashed", now)
	if err != nil || acquired == nil {
		t.Fatalf("failed to acquire job: %v", err)
	}
	if acquired.Status != worker.JobStatusProcessing {
		t.Fatalf("expected status PROCESSING, got %s", acquired.Status)
	}

	// 3. Worker crashes! Time advances 65 seconds (heartbeat expired)
	futureTime := now.Add(65 * time.Second)

	reapedCount, reapedIDs, err := reaper.ReapZombies(ctx, futureTime)
	if err != nil {
		t.Fatalf("reap failed: %v", err)
	}
	if reapedCount != 1 || len(reapedIDs) != 1 || reapedIDs[0] != job.ID {
		t.Fatalf("expected job %s to be reaped, got count=%d, ids=%v", job.ID, reapedCount, reapedIDs)
	}

	// 4. Job is now in RETRY_SCHEDULED; Worker 2 can acquire it!
	worker2Job, err := jobRepo.AcquireNextJob(ctx, worker.QueueAPKAnalysis, "worker-healthy", futureTime)
	if err != nil || worker2Job == nil {
		t.Fatalf("healthy worker failed to acquire reaped job: %v", err)
	}
	if worker2Job.ID != job.ID {
		t.Fatalf("expected to acquire reaped job %s, got %s", job.ID, worker2Job.ID)
	}
	if worker2Job.LockedBy != "worker-healthy" {
		t.Fatalf("expected locked by worker-healthy, got %s", worker2Job.LockedBy)
	}

	// 5. Simulate repeated failures until DEAD_LETTER
	_, _ = jobRepo.FailJob(ctx, job.ID, "failure 2", futureTime)
	_, _ = jobRepo.AcquireNextJob(ctx, worker.QueueAPKAnalysis, "worker-healthy", futureTime.Add(10*time.Minute))
	finalFailedJob, _ := jobRepo.FailJob(ctx, job.ID, "failure 3 (final)", futureTime.Add(10*time.Minute))

	if finalFailedJob.Status != worker.JobStatusDeadLetter {
		t.Fatalf("expected DEAD_LETTER after max attempts, got %s", finalFailedJob.Status)
	}

	// 6. Test DLQ manual retry
	dlqList, err := jobRepo.ListDeadLetterJobs(ctx, 10)
	if err != nil || len(dlqList) != 1 {
		t.Fatalf("expected 1 job in DLQ, got %d", len(dlqList))
	}

	if err := reaper.RetryDeadLetterJob(ctx, job.ID); err != nil {
		t.Fatalf("failed to retry dead letter job: %v", err)
	}

	retriedJob, _ := jobRepo.GetJobByID(ctx, job.ID)
	if retriedJob.Status != worker.JobStatusQueued || retriedJob.Attempts != 0 {
		t.Fatalf("expected queued job with 0 attempts after DLQ retry, got status=%s, attempts=%d",
			retriedJob.Status, retriedJob.Attempts)
	}
}
