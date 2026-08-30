package workload_submitter

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
)

const (
	submissionTimeout = 5 * time.Second
	statusLogInterval = 30 * time.Second
)

// Params holds derived values computed from Config before the Submitter is created.
type Params struct {
	// Shared
	WCap  int
	DSecs int // workload lifetime in seconds; drives KWOK stage duration and sleep command

	// Rate mode
	WTotal         int
	RampRate       float64 // submissions/min during ramp phase
	SteadyDuration time.Duration

	// Burst mode (zero in rate mode)
	BurstTarget       int
	ObservationWindow time.Duration
}

// submitFunc is the function signature used to submit one workload.
// Overridable in tests to inject fakes without a live cluster.
type submitFunc func(ctx context.Context, clients *Clients, p submitParams) (string, *http.Response, error)

type Submitter struct {
	cfg           *Config
	clients       *Clients
	params        Params
	projectIDs    []string
	nodepoolNames []string // may be nil if NodepoolCount==0; workloads go to any pool

	sem       *semaphore.Weighted
	submitted atomic.Int64 // incremented before each ramp goroutine fires; gates ramp exit
	counter   atomic.Int64 // global sequence number; used for workload names + round-robin

	// in-flight tracking for queue depth warnings
	pending    atomic.Int64
	warnActive atomic.Bool // true while pending > MaxConcurrent

	failMu       sync.Mutex
	failures     map[int]int // HTTP status code → count; 0 = connection error
	successCount atomic.Int64

	// observability
	startTime atomic.Value // stores time.Time; set at Run()/RunBurst() start
	phase     atomic.Value // stores string: "ramp" | "steady" | "burst" | "observe"

	submitFn             submitFunc    // defaults to SubmitOne; overridable in tests
	latencyWarnThreshold time.Duration // defaults to 2s; overridable in tests
}

func NewSubmitter(cfg *Config, clients *Clients, params Params, projectIDs, nodepoolNames []string) *Submitter {
	return &Submitter{
		cfg:                  cfg,
		clients:              clients,
		params:               params,
		projectIDs:           projectIDs,
		nodepoolNames:        nodepoolNames,
		sem:                  semaphore.NewWeighted(int64(cfg.MaxConcurrent)),
		failures:             make(map[int]int),
		submitFn:             SubmitOne,
		latencyWarnThreshold: 2 * time.Second,
	}
}

// Run executes the two-phase submission loop:
//  1. Ramp: submit at RampRate/min until WTotal workloads have been dispatched.
//  2. Steady state: submit at SubmissionRatePerMin/min for SteadyDuration.
//
// In both phases each tick launches a goroutine that blocks on the semaphore before
// calling the API, so the ticker advances independently of API latency. Each submission
// goroutine has a hard 5s timeout so a slow/unavailable API never blocks a slot
// indefinitely. A status line is logged every 30s throughout the run.
func (s *Submitter) Run(ctx context.Context) {
	s.startTime.Store(time.Now())
	s.phase.Store("ramp")

	statusDone := make(chan struct{})
	go s.statusLogger(ctx, statusDone)
	defer close(statusDone)

	s.runRamp(ctx)
	if ctx.Err() != nil {
		return
	}
	s.phase.Store("steady")
	s.runSteadyState(ctx)
}

// RunBurst submits BurstTarget workloads as fast as the API accepts, using a fixed pool
// of BurstWorkers goroutines, then waits ObservationWindow for the queue to drain naturally
// via KWOK stage completions.
func (s *Submitter) RunBurst(ctx context.Context) {
	s.startTime.Store(time.Now())
	s.phase.Store("burst")

	statusDone := make(chan struct{})
	go s.statusLogger(ctx, statusDone)
	defer close(statusDone)

	// work is a buffered channel sized to BurstWorkers so the feed loop stays up to
	// BurstWorkers items ahead of the workers and no worker starves between submissions.
	work := make(chan int64, s.cfg.BurstWorkers)

	var wg sync.WaitGroup
	for range s.cfg.BurstWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range work {
				if ctx.Err() != nil {
					return
				}
				s.doSubmit(ctx, idx)
			}
		}()
	}

	for range s.params.BurstTarget {
		if ctx.Err() != nil {
			break
		}
		work <- s.counter.Add(1) - 1
	}
	close(work)
	wg.Wait()

	if ctx.Err() != nil {
		return
	}
	log.Printf("[burst] submitted %d workloads (workers=%d D=%ds), observing for %v",
		s.params.BurstTarget, s.cfg.BurstWorkers, s.params.DSecs, s.params.ObservationWindow)
	s.phase.Store("observe")
	select {
	case <-ctx.Done():
	case <-time.After(s.params.ObservationWindow):
		log.Printf("[observe] window complete")
	}
}

func (s *Submitter) runRamp(ctx context.Context) {
	rampInterval := time.Duration(float64(time.Minute) / s.params.RampRate)
	log.Printf("[ramp] starting: target=%d workloads rate=%.2f/min interval=%v",
		s.params.WTotal, s.params.RampRate, rampInterval)

	ticker := time.NewTicker(rampInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n := s.submitted.Add(1)
			if n > int64(s.params.WTotal) {
				log.Printf("[ramp] complete: dispatched %d workloads", s.params.WTotal)
				return
			}
			idx := s.counter.Add(1) - 1
			go s.launchSubmit(ctx, idx)
		}
	}
}

func (s *Submitter) runSteadyState(ctx context.Context) {
	steadyInterval := time.Duration(float64(time.Minute) / s.cfg.SubmissionRatePerMin)
	log.Printf("[steady] starting: rate=%.2f/min interval=%v duration=%v",
		s.cfg.SubmissionRatePerMin, steadyInterval, s.params.SteadyDuration)

	ticker := time.NewTicker(steadyInterval)
	defer ticker.Stop()
	deadline := time.NewTimer(s.params.SteadyDuration)
	defer deadline.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-deadline.C:
			log.Printf("[steady] complete: duration %v elapsed", s.params.SteadyDuration)
			return
		case <-ticker.C:
			idx := s.counter.Add(1) - 1
			go s.launchSubmit(ctx, idx)
		}
	}
}

// launchSubmit is the goroutine body for each tick. It tracks pending depth for queue
// warnings, blocks on the semaphore, then calls doSubmit.
func (s *Submitter) launchSubmit(ctx context.Context, idx int64) {
	depth := s.pending.Add(1)
	defer func() {
		// recovery fires on exit: when pending drops back to ≤ MaxConcurrent after a warning
		if s.pending.Add(-1) <= int64(s.cfg.MaxConcurrent) {
			if s.warnActive.CompareAndSwap(true, false) {
				log.Printf("[INFO] submission queue cleared — back to normal rate")
			}
		}
	}()

	// warn once when the queue starts backing up
	if depth > int64(s.cfg.MaxConcurrent) {
		if s.warnActive.CompareAndSwap(false, true) {
			log.Printf("[WARNING] submission queue depth %d exceeds max concurrent %d — API latency is exceeding tick interval; submissions will be delayed",
				depth, s.cfg.MaxConcurrent)
		}
	}

	if err := s.sem.Acquire(ctx, 1); err != nil {
		return // context cancelled
	}
	defer s.sem.Release(1)

	s.doSubmit(ctx, idx)
}

func (s *Submitter) doSubmit(ctx context.Context, idx int64) {
	name := fmt.Sprintf("workload-%d", idx)

	projectID := s.projectIDs[idx%int64(len(s.projectIDs))]

	nodepoolName := ""
	if len(s.nodepoolNames) > 0 {
		nodepoolName = s.nodepoolNames[idx%int64(len(s.nodepoolNames))]
	}

	p := submitParams{
		Name:            name,
		ProjectID:       projectID,
		ClusterID:       s.cfg.ClusterID,
		NodepoolName:    nodepoolName,
		DSecs:           s.params.DSecs,
		PodsPerWorkload: s.cfg.PodsPerWorkload,
		AutoDeleteSecs:  s.cfg.AutoDeletionSeconds,
	}

	submitCtx, cancel := context.WithTimeout(ctx, submissionTimeout)
	defer cancel()

	start := time.Now()
	_, resp, err := s.submitFn(submitCtx, s.clients, p)
	latency := time.Since(start)

	if latency > s.latencyWarnThreshold {
		log.Printf("[WARNING] slow submission: %s took %v", name, latency)
	}

	if err != nil || (resp != nil && resp.StatusCode != http.StatusAccepted) {
		s.recordFailure(resp, err)
		return
	}
	s.successCount.Add(1)
}

func (s *Submitter) recordFailure(resp *http.Response, err error) {
	code := 0
	if resp != nil {
		code = resp.StatusCode
	}
	s.failMu.Lock()
	s.failures[code]++
	s.failMu.Unlock()
	if err != nil {
		log.Printf("[ERROR] submission failed (status=%d): %v", code, err)
	} else {
		log.Printf("[ERROR] submission rejected (status=%d)", code)
	}
}

// statusLogger emits a status line every statusLogInterval until ctx is cancelled or done is closed.
func (s *Submitter) statusLogger(ctx context.Context, done <-chan struct{}) {
	ticker := time.NewTicker(statusLogInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-done:
			return
		case <-ticker.C:
			s.logStatus()
		}
	}
}

// failureCount returns actual recorded failures, matching the [ERROR] lines and the
// summary breakdown.
func (s *Submitter) failureCount() int64 {
	s.failMu.Lock()
	defer s.failMu.Unlock()
	var n int64
	for _, c := range s.failures {
		n += int64(c)
	}
	return n
}

func (s *Submitter) logStatus() {
	phase, _ := s.phase.Load().(string)
	dispatched := s.counter.Load()
	succeeded := s.successCount.Load()
	failed := s.failureCount()
	queue := s.pending.Load()
	elapsed := time.Since(s.startTime.Load().(time.Time)).Round(time.Second)
	log.Printf("[status] phase=%-6s dispatched=%-6d succeeded=%-6d failed=%-4d queue=%-3d elapsed=%v",
		phase, dispatched, succeeded, failed, queue, elapsed)
}

// PrintSummary logs a final breakdown of submission outcomes.
// RetentionInfo is written next to the run output so the TTL and the workload count it
// implies are stamped onto every result row.
type RetentionInfo struct {
	AutoDeletionSeconds  int     `json:"auto_deletion_seconds"`
	WorkloadDurationSec  int     `json:"workload_duration_seconds"`
	SubmissionRatePerMin float64 `json:"submission_rate_per_min"`
	// ExpectedResidentWorkloads is how many workloads the cluster is expected to hold. With a
	// TTL it settles at R/60 x (D + TTL); without one nothing is ever removed, so it is
	// everything the run submits — unbounded in principle, but a known figure for this run.
	ExpectedResidentWorkloads int `json:"expected_resident_workloads"`
}

func (s *Submitter) Retention() RetentionInfo {
	info := RetentionInfo{
		AutoDeletionSeconds:  s.cfg.AutoDeletionSeconds,
		WorkloadDurationSec:  s.params.DSecs,
		SubmissionRatePerMin: s.cfg.SubmissionRatePerMin,
	}
	if info.AutoDeletionSeconds > 0 {
		info.ExpectedResidentWorkloads = int(s.cfg.SubmissionRatePerMin / 60 * float64(s.params.DSecs+info.AutoDeletionSeconds))
	} else {
		info.ExpectedResidentWorkloads = s.totalSubmissions()
	}
	return info
}

// totalSubmissions is how many workloads the run submits end to end.
func (s *Submitter) totalSubmissions() int {
	if s.cfg.SubmissionMode == SubmissionModeBurst {
		return s.params.BurstTarget
	}
	// The ramp fills W_total, then steady state adds R per minute.
	return s.params.WTotal + int(s.cfg.SubmissionRatePerMin*s.params.SteadyDuration.Minutes())
}

// WriteRetention emits retention.json for the orchestrator to merge into run metadata.
func (s *Submitter) WriteRetention(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", outputDir, err)
	}
	data, err := json.MarshalIndent(s.Retention(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outputDir, "retention.json"), data, 0o644)
}

func (s *Submitter) autoDeletionSummary() string {
	r := s.Retention()
	if r.AutoDeletionSeconds <= 0 {
		return "disabled — completed workloads are never deleted, resident count grows without bound"
	}
	return fmt.Sprintf("%ds after completion; expected resident workloads = %d (R=%.0f/min x (D=%ds + TTL=%ds))",
		r.AutoDeletionSeconds, r.ExpectedResidentWorkloads, r.SubmissionRatePerMin, r.WorkloadDurationSec, r.AutoDeletionSeconds)
}

func (s *Submitter) PrintSummary() {
	total := s.counter.Load()
	successes := s.successCount.Load()
	failures := s.failureCount()

	log.Printf("=== submission summary ===")
	log.Printf("  total dispatched : %d", total)
	log.Printf("  succeeded        : %d", successes)
	log.Printf("  failed           : %d", failures)
	log.Printf("  auto-deletion    : %s", s.autoDeletionSummary())

	if len(s.failures) > 0 {
		codes := make([]int, 0, len(s.failures))
		for c := range s.failures {
			codes = append(codes, c)
		}
		sort.Ints(codes)
		log.Printf("  failure breakdown:")
		for _, c := range codes {
			label := fmt.Sprintf("HTTP %d", c)
			if c == 0 {
				label = "connection error"
			}
			log.Printf("    %-20s : %d", label, s.failures[c])
		}
	}
}
