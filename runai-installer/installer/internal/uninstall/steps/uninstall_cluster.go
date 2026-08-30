package steps

import (
	"context"
	"errors"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
)

type UninstallClusterChart struct {
	*step.RunAIInstallStepImpl
	helmWrapper uninstall.HelmWrapper
}

func NewUninstallClusterChart(cfg *config.UninstallConfig, helmWrapper uninstall.HelmWrapper) *UninstallClusterChart {
	return &UninstallClusterChart{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Uninstall cluster chart", cfg.StepTimeout),
		helmWrapper:          helmWrapper,
	}
}

func (s *UninstallClusterChart) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	// TODO: InstallPhase values (Installing/Installed) are install-centric but used as generic
	// in-progress/completed signals. Refactor to use neutral names (e.g. InProgress/Done).
	s.SetStatus(model.Installing, nil)

	waitTimeout := time.Duration(0)
	if deadline, ok := ctx.Deadline(); ok {
		waitTimeout = time.Until(deadline)
	}

	// A previously interrupted install may leave the release in pending state.
	// Fix it before uninstalling so Helm's pending-state protection does not block us.
	// We warn and continue rather than failing: if the fix fails the worst case is that
	// UninstallRelease below returns an error, which is handled there.
	if err := s.helmWrapper.FixPendingRelease(step.ClusterReleaseName, step.ClusterNamespace); err != nil {
		logging.Warnf("[Uninstall] could not fix pending cluster release: %v", err)
	}

	logging.Infof("[Uninstall] Uninstalling cluster chart")
	if err := s.helmWrapper.UninstallRelease(step.ClusterReleaseName, step.ClusterNamespace, waitTimeout); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			s.SetStatus(model.Timeout, err)
		} else {
			s.SetStatus(model.Failed, err)
		}
		return err
	}

	logging.Infof("[Uninstall] Cluster chart uninstalled successfully")
	s.SetStatus(model.Installed, nil)
	return nil
}
