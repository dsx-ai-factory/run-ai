package step

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

var defaultStepTimeout = 5 * time.Minute

func SetDefaultStepTimeout(d time.Duration) {
	if d > 0 {
		defaultStepTimeout = d
	}
}

type RunAIInstallStepImpl struct {
	name           string
	status         model.InstallPhase
	stepError      error
	mu             sync.RWMutex
	executionStart time.Time
	timeout        time.Duration
	probeInFlight  atomic.Bool
	executeInFlight atomic.Bool
	statusEvaluator func(runaiData *model.RunAIInstallData)
}

func NewRunAIInstallStepImpl(name string, timeout ...time.Duration) *RunAIInstallStepImpl {
	var t time.Duration
	if len(timeout) > 0 {
		t = timeout[0]
	}
	return &RunAIInstallStepImpl{
		name:    name,
		status:  model.NotInstalled,
		timeout: t,
	}
}

// SetStatusEvaluator registers the step-specific status evaluation function.
// This must be called after construction to enable GetStatusAsync/GetStatusSync.
func (r *RunAIInstallStepImpl) SetStatusEvaluator(statusEvaluator func(runaiData *model.RunAIInstallData)) {
	r.statusEvaluator = statusEvaluator
}

func (r *RunAIInstallStepImpl) HasStatusEvaluator() bool {
	return r.statusEvaluator != nil
}

func (r *RunAIInstallStepImpl) GetName() string {
	return r.name
}

func (r *RunAIInstallStepImpl) GetError() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stepError
}

// GetCachedStatus returns the last known status without any I/O (implements InstallStep interface)
func (r *RunAIInstallStepImpl) GetCachedStatus() model.InstallPhase {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status
}

func (r *RunAIInstallStepImpl) SetStatus(status model.InstallPhase, err error) model.InstallPhase {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status = status
	if err != nil {
		r.stepError = err
	} else {
		r.stepError = nil
	}
	return r.status
}

// checkContextCancellation checks if the context is cancelled and handles it appropriately
func (r *RunAIInstallStepImpl) CheckContextCancellation(ctx context.Context) error {
	select {
	case <-ctx.Done():
		err := fmt.Errorf("Step '%s' timed out", r.GetName())
		r.SetStatus(model.Failed, err)
		return err
	default:
		return nil
	}
}

func (r *RunAIInstallStepImpl) IsTerminalStatus() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.status == model.Failed || r.status == model.Timeout || r.status == model.Installed
}

// IsExecuteInFlight returns true if Execute() is currently running.
// This uses the executeInFlight atomic guard, not the status field.
func (r *RunAIInstallStepImpl) IsExecuteInFlight() bool {
	return r.executeInFlight.Load()
}

// ShouldSkipStatusEvaluation returns true if status evaluation should be skipped.
// This is the case when we're in a terminal state OR when Execute() is in progress.
// Steps that override GetStatusAsync should call this at the start and return
// GetCachedStatus() if it returns true, to prevent race conditions where status
// is set to Installed before Execute() completes and sets required data fields.
func (r *RunAIInstallStepImpl) ShouldSkipStatusEvaluation() bool {
	return r.IsTerminalStatus() || r.IsExecuteInFlight()
}

func (r *RunAIInstallStepImpl) PostInstall(runaiData *model.RunAIInstallData) {}

// StartExecutionTimer starts the timer when Execute() begins.
// This is called automatically by TryStartExecute() - steps should NOT call this manually.
func (r *RunAIInstallStepImpl) StartExecutionTimer() {
	if r.executionStart.IsZero() {
		r.executionStart = time.Now()
	}
}

func (r *RunAIInstallStepImpl) CheckTimeout() error {
	timeout := r.GetTimeout()
	if !r.executionStart.IsZero() && time.Since(r.executionStart) > timeout {
		return fmt.Errorf("'%s' timed out after %v", r.GetName(), timeout)
	}
	return nil
}

// GetTimeoutHint returns an optional troubleshooting hint shown when the step times out.
// Override in step implementations to provide step-specific guidance.
// Default implementation returns empty string (no hint).
func (r *RunAIInstallStepImpl) GetTimeoutHint() string {
	return ""
}

// GetElapsedTime returns the elapsed time since execution started
func (r *RunAIInstallStepImpl) GetElapsedTime() time.Duration {
	if r.executionStart.IsZero() {
		return 0
	}
	return time.Since(r.executionStart).Round(time.Second)
}

func (r *RunAIInstallStepImpl) GetTimeout() time.Duration {
	if r.timeout > 0 {
		return r.timeout
	}
	return defaultStepTimeout
}

// TryStartProbe attempts to acquire the probe lock. Returns true if successful
// (caller should start the probe goroutine), false if a probe is already running.
// Use this in GetStatus() to prevent goroutine pileup from repeated UI ticks.
func (r *RunAIInstallStepImpl) TryStartProbe() bool {
	return r.probeInFlight.CompareAndSwap(false, true)
}

// EndProbe releases the probe lock. Must be called (typically via defer) when
// the background probe goroutine completes.
func (r *RunAIInstallStepImpl) EndProbe() {
	r.probeInFlight.Store(false)
}

// TryStartExecute attempts to acquire the execute lock. Returns true if successful
// (caller should proceed with Execute()), false if Execute() is already running.
// This prevents multiple concurrent Execute() calls from UI ticks racing.
func (r *RunAIInstallStepImpl) TryStartExecute() bool {
	if r.executeInFlight.CompareAndSwap(false, true) {
		r.StartExecutionTimer()
		logging.Infof("[RunAI] Starting '%s' (timeout: %v)", r.GetName(), r.GetTimeout())
		return true
	}
	return false
}

// EndExecute releases the execute lock. Must be called (typically via defer) when
// Execute() completes (whether successful or not).
func (r *RunAIInstallStepImpl) EndExecute() {
	r.executeInFlight.Store(false)
}

// GetStatusAsync checks status asynchronously (non-blocking, for UI polling).
// It starts a background probe if none is running and returns cached status immediately.
// The actual evaluation logic is provided via SetStatusEvaluator().
// Note: Does not run probes while Execute() is in progress to prevent race conditions.
func (r *RunAIInstallStepImpl) GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase {
	// Skip evaluation if in terminal state or Execute() is running
	if r.ShouldSkipStatusEvaluation() {
		return r.GetCachedStatus()
	}
	if r.statusEvaluator != nil && r.TryStartProbe() {
		go func() {
			defer r.EndProbe()
			r.statusEvaluator(runaiData)
		}()
	}
	return r.GetCachedStatus()
}

// GetStatusSync checks status synchronously (blocking, for initial checks).
// Use this to determine initial state (e.g., checking if already installed).
// The actual evaluation logic is provided via SetStatusEvaluator().
func (r *RunAIInstallStepImpl) GetStatusSync(runaiData *model.RunAIInstallData) model.InstallPhase {
	if r.IsTerminalStatus() {
		return r.GetCachedStatus()
	}
	if r.statusEvaluator != nil {
		r.statusEvaluator(runaiData)
	}
	return r.GetCachedStatus()
}
