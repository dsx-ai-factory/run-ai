package step

import (
	"context"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
)

type InstallStep interface {
	GetName() string
	Execute(ctx context.Context, runaiData *model.RunAIInstallData) error
	// GetStatusAsync checks status asynchronously (non-blocking, for UI polling during installation)
	GetStatusAsync(runaiData *model.RunAIInstallData) model.InstallPhase
	// GetStatusSync checks status synchronously (blocking, for initial "already installed" checks)
	GetStatusSync(runaiData *model.RunAIInstallData) model.InstallPhase
	// GetCachedStatus returns the last known status without any I/O (safe for frequent calls)
	GetCachedStatus() model.InstallPhase
	GetError() error
	SetStatus(status model.InstallPhase, err error) model.InstallPhase
	PostInstall(runaiData *model.RunAIInstallData)
	GetElapsedTime() time.Duration
	GetTimeout() time.Duration
	CheckTimeout() error
	// GetTimeoutHint returns an optional troubleshooting hint shown when the step times out.
	// Override in step implementations to provide step-specific guidance (e.g., "Check pods with: kubectl get pods -n X").
	// Returns empty string by default, meaning no additional hint is shown.
	GetTimeoutHint() string
	// TryStartExecute attempts to acquire the execute lock and starts the execution timer.
	// Returns true if successful, false if Execute() is already running.
	// Prevents concurrent Execute() calls and ensures timer is always started.
	TryStartExecute() bool
	// EndExecute releases the execute lock. Must be called when Execute() completes.
	EndExecute()
}
