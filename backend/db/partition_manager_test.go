package db

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGenerateMonthlyPartition(t *testing.T) {
	target := PartitionTarget{ParentTable: "install_events", Prefix: "install_events"}
	pDef := GenerateMonthlyPartition(target, 2026, time.October)

	expectedName := "install_events_y2026m10"
	if pDef.PartitionName != expectedName {
		t.Fatalf("expected partition name %s, got %s", expectedName, pDef.PartitionName)
	}

	if !strings.Contains(pDef.DDL, "PARTITION OF install_events") {
		t.Fatalf("expected DDL to target install_events, got %s", pDef.DDL)
	}

	if !strings.Contains(pDef.DDL, "2026-10-01 00:00:00+00") || !strings.Contains(pDef.DDL, "2026-11-01 00:00:00+00") {
		t.Fatalf("unexpected date range in DDL: %s", pDef.DDL)
	}
}

func TestEnsureFuturePartitions(t *testing.T) {
	baseDate := time.Date(2026, time.October, 15, 0, 0, 0, 0, time.UTC)
	var executedSQL []string

	exec := func(ctx context.Context, sql string) error {
		executedSQL = append(executedSQL, sql)
		return nil
	}

	partitions, err := EnsureFuturePartitions(context.Background(), exec, nil, baseDate, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 default targets * (1 current + 2 future months = 3) = 6 partitions
	if len(partitions) != 6 {
		t.Fatalf("expected 6 partitions, got %d", len(partitions))
	}
	if len(executedSQL) != 6 {
		t.Fatalf("expected 6 executed SQL statements, got %d", len(executedSQL))
	}
}

