package schedule

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	loggermocks "github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger/mocks"
	dbtesting "github.com/run-ai/runai/runai-common-packages/runaidb/testing"
	"github.com/stretchr/testify/require"
)

func TestToCronExpression_Seconds(t *testing.T) {
	require.Equal(t, "30 seconds", toCronExpression(30*time.Second))
}

func TestToCronExpression_Minutes(t *testing.T) {
	require.Equal(t, "*/20 * * * *", toCronExpression(20*time.Minute))
}

func TestToCronExpression_Hours(t *testing.T) {
	require.Equal(t, "0 */2 * * *", toCronExpression(2*time.Hour))
}

func TestGetInitialDelay_ZeroTime(t *testing.T) {
	require.Equal(t, time.Duration(0), getInitialDelay(time.Time{}))
}

func TestGetInitialDelay_FutureTime(t *testing.T) {
	future := time.Now().Add(1 * time.Hour)
	delay := getInitialDelay(future)
	require.Greater(t, delay, time.Duration(0))
}

func TestIsPGCronInstalled_NotInstalled(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT count(*) FROM pg_extension WHERE extname = ?", "pg_cron").
		Returns([]struct{ Count int64 }{{Count: 0}})

	require.False(t, isPGCronInstalled(db))
}

func TestIsPGCronInstalled_Installed(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT count(*) FROM pg_extension WHERE extname = ?", "pg_cron").
		Returns([]struct{ Count int64 }{{Count: 1}})

	require.True(t, isPGCronInstalled(db))
}

func TestSchedule_InvalidInterval(t *testing.T) {
	ctx := context.Background()

	db, _ := dbtesting.NewMocker(t)
	loggerMock := loggermocks.NewLogger(t)
	loggerMock.On("Errorf", "schedule: job %q has non-positive interval %s, not scheduling", "test-job", time.Duration(0)).Return()
	logger.SetLogger(loggerMock)
	defer logger.SetLogger(logger.NewDummyLogger())

	job := NewJob(db, "test-job", 0, []string{"INSERT INTO schedule_log DEFAULT VALUES"})
	job.Schedule(ctx)

}

func TestSchedule_FallsBackToGoroutineWhenPGCronAbsent(t *testing.T) {
	db, mocker := dbtesting.NewMocker(t)

	mocker.ExpectQuery("SELECT count(*) FROM pg_extension WHERE extname = ?", "pg_cron").
		Returns([]struct{ Count int64 }{{Count: 0}})

	mocker.ExpectExec("INSERT INTO schedule_log DEFAULT VALUES").Returns(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	job := NewJob(db, "test-job", 10*time.Second, []string{"INSERT INTO schedule_log DEFAULT VALUES"})
	job.Schedule(ctx)

	time.Sleep(50 * time.Millisecond) // give goroutine time to fire the first execution
}
