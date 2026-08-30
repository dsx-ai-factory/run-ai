package integrationtests

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/schedule"
	"github.com/stretchr/testify/require"
)

func TestSchedule_PGCron_ExecutesQuery(t *testing.T) {
	db := runaidb.DB()

	_, err := db.Exec("DELETE FROM test_schema.search_docs")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job := schedule.NewJob(
		db,
		"integ-goroutine-job",
		500*time.Millisecond,
		[]string{"INSERT INTO test_schema.search_docs (name, description) VALUES ('test', 'test')"},
	)
	job.Schedule(ctx)

	require.Eventually(t, func() bool {
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM test_schema.search_docs").Row().Scan(&count); err != nil {
			return false
		}
		return count >= 2
	}, 5*time.Second, 100*time.Millisecond)

	cancel()
}

func TestSchedule_PGCron_RegistersJob(t *testing.T) {
	db := runaidb.DB()

	var count int64
	err := db.Table("pg_extension").Where("extname = ?", "pg_cron").Count(&count)
	require.NoError(t, err)
	require.Equal(t, int64(1), count, "pg_cron extension should be installed")

	const jobName = "integ-pgcron-job"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job := schedule.NewJob(
		db,
		jobName,
		1*time.Minute,
		[]string{"INSERT INTO test_schema.search_docs (name, description) VALUES ('test', 'test')"},
	)
	job.Schedule(ctx)

	var registered int64
	err = db.Raw("SELECT COUNT(*) FROM cron.job WHERE jobname = ?", jobName).Row().Scan(&registered)
	require.NoError(t, err)
	require.Equal(t, int64(1), registered)

	_, err = db.Exec("SELECT cron.unschedule(?)", jobName)
	require.NoError(t, err)
}
