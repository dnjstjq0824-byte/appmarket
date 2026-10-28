package db

import (
	"context"
	"fmt"
	"time"
)

// PartitionTarget represents a table configured for monthly range partitioning.
type PartitionTarget struct {
	ParentTable string
	Prefix      string
}

var DefaultPartitionTargets = []PartitionTarget{
	{ParentTable: "install_events", Prefix: "install_events"},
	{ParentTable: "search_query_logs", Prefix: "search_query_logs"},
}

// PartitionDefinition contains calculated metadata for a monthly partition.
type PartitionDefinition struct {
	ParentTable   string
	PartitionName string
	RangeStart    time.Time
	RangeEnd      time.Time
	DDL           string
}

// GenerateMonthlyPartition computes the partition table name and DDL for a specific year and month.
func GenerateMonthlyPartition(target PartitionTarget, year int, month time.Month) PartitionDefinition {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	partitionName := fmt.Sprintf("%s_y%04dm%02d", target.Prefix, year, int(month))

	ddl := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s');",
		partitionName,
		target.ParentTable,
		start.Format("2006-01-02 15:04:05-07"),
		end.Format("2006-01-02 15:04:05-07"),
	)

	return PartitionDefinition{
		ParentTable:   target.ParentTable,
		PartitionName: partitionName,
		RangeStart:    start,
		RangeEnd:      end,
		DDL:           ddl,
	}
}

// DBExecutor represents an interface for executing SQL queries (like pgxpool or sql.DB).
type DBExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (any, error)
}

// EnsureFuturePartitions creates partitions for the current month and the specified number of future months.
func EnsureFuturePartitions(ctx context.Context, exec func(ctx context.Context, sql string) error, targets []PartitionTarget, baseDate time.Time, monthsAhead int) ([]PartitionDefinition, error) {
	if len(targets) == 0 {
		targets = DefaultPartitionTargets
	}
	var created []PartitionDefinition

	for i := 0; i <= monthsAhead; i++ {
		t := baseDate.AddDate(0, i, 0)
		year, month := t.Year(), t.Month()

		for _, target := range targets {
			pDef := GenerateMonthlyPartition(target, year, month)
			if exec != nil {
				if err := exec(ctx, pDef.DDL); err != nil {
					return created, fmt.Errorf("failed to create partition %s: %w", pDef.PartitionName, err)
				}
			}
			created = append(created, pDef)
		}
	}
	return created, nil
}

