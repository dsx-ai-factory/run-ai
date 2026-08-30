package api_exerciser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/run-ai/runai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
)

const (
	summaryInterval = 30 * time.Second
	requestTimeout  = 30 * time.Second
)

type errorStat struct {
	total    int64
	lastSeen time.Time
}

type Exerciser struct {
	clients       *sampler_actions.Clients
	config        *Config
	clusterConfig *scale.ClusterConfig

	mu         sync.Mutex
	errorStats map[int]*errorStat // HTTP status code → stats; status 0 means timeout/no response
	prevCounts map[int]int64      // counts at last summary, for computing deltas
}

func NewExerciser(clients *sampler_actions.Clients, cfg *Config, clusterConfig *scale.ClusterConfig) *Exerciser {
	return &Exerciser{
		clients:       clients,
		config:        cfg,
		clusterConfig: clusterConfig,
		errorStats:    make(map[int]*errorStat),
		prevCounts:    make(map[int]int64),
	}
}

func (e *Exerciser) Run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if e.config.DurationSecs > 0 {
		var durationCancel context.CancelFunc
		ctx, durationCancel = context.WithTimeout(ctx, time.Duration(e.config.DurationSecs)*time.Second)
		defer durationCancel()
	}

	logf("starting %d users, think time %.1fs, page size %d",
		e.config.NUsers, e.config.ThinkTimeSecs, e.config.PageSize)

	var wg sync.WaitGroup

	for i := range e.config.NUsers {
		wg.Add(1)
		// i is passed as an argument rather than captured by the closure to
		// avoid all goroutines sharing the same loop variable.
		go func(userID int) {
			defer wg.Done()
			e.userLoop(ctx, userID)
		}(i)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		e.summaryLoop(ctx)
	}()

	wg.Wait()
	e.printSummary()
}

func (e *Exerciser) userLoop(ctx context.Context, userID int) {
	think := time.Duration(float64(time.Second) * e.config.ThinkTimeSecs)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(think):
		}

		// Per-request timeout so a hung server doesn't stall this goroutine
		// indefinitely. reqCancel is called immediately (not deferred) because
		// defer runs at function return — deferring in a loop would accumulate
		// one live timer per iteration until the goroutine exits.
		reqCtx, reqCancel := context.WithTimeout(ctx, requestTimeout)
		_, resp, err := e.clients.WorkloadServiceClient.WorkloadsAPI.
			GetWorkloads(reqCtx).
			Limit(e.config.PageSize).
			Offset(0).
			Execute()
		reqCancel()

		if err != nil {
			code := statusCode(resp)
			e.recordError(code)
			body := readBody(resp)
			logf("user %d: request error (status %d): %v%s", userID, code, err, body)
			continue
		}
		if resp != nil && resp.StatusCode >= http.StatusBadRequest {
			e.recordError(resp.StatusCode)
			logf("user %d: unexpected status %d%s", userID, resp.StatusCode, readBody(resp))
		}
	}
}

func (e *Exerciser) recordError(code int) {
	e.mu.Lock()
	if e.errorStats[code] == nil {
		e.errorStats[code] = &errorStat{}
	}
	e.errorStats[code].total++
	e.errorStats[code].lastSeen = time.Now()
	e.mu.Unlock()
}

func (e *Exerciser) summaryLoop(ctx context.Context) {
	ticker := time.NewTicker(summaryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.printSummary()
		}
	}
}

func logf(format string, args ...any) {
	_, file, line, _ := runtime.Caller(1)
	// trim to just the filename
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}
	prefix := fmt.Sprintf("[api-exerciser] %s %s:%d ", time.Now().Format("15:04:05"), file, line)
	fmt.Printf(prefix+format+"\n", args...)
}

func (e *Exerciser) printSummary() {
	// Copy under lock so user goroutines are not blocked during the print calls.
	e.mu.Lock()
	type snapshot struct {
		total    int64
		lastSeen time.Time
	}
	snap := make(map[int]snapshot, len(e.errorStats))
	for code, s := range e.errorStats {
		snap[code] = snapshot{total: s.total, lastSeen: s.lastSeen}
	}
	e.mu.Unlock()

	if len(snap) == 0 {
		logf("error summary: no errors")
		return
	}

	codes := make([]int, 0, len(snap))
	for c := range snap {
		codes = append(codes, c)
	}
	sort.Ints(codes)

	summary := "error summary:"
	for _, c := range codes {
		s := snap[c]
		delta := s.total - e.prevCounts[c]
		ago := time.Since(s.lastSeen).Round(time.Second)
		summary += fmt.Sprintf(" [status=%d total=%d +%d last=%s ago]", c, s.total, delta, ago)
		e.prevCounts[c] = s.total
	}
	logf("%s", summary)
}

func statusCode(resp *http.Response) int {
	if resp != nil {
		return resp.StatusCode
	}
	return 0
}

// readBody returns the response body as a log-friendly string, or empty string if nil.
func readBody(resp *http.Response) string {
	if resp == nil || resp.Body == nil {
		return ""
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 512))
	if err != nil || len(b) == 0 {
		return ""
	}
	return fmt.Sprintf(" body=%q", b)
}
