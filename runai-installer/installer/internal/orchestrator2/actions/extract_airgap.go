package actions

import (
	"context"
	"fmt"
	"os"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/airgap"
)

type ExtractAirgap struct {
	BundlePath string
	WorkingDir string
}

func (a ExtractAirgap) Kind() string { return "airgap.extract" }

func (a ExtractAirgap) Args() map[string]any {
	return map[string]any{
		"bundlePath": a.BundlePath,
		"workingDir": a.WorkingDir,
	}
}

func (a ExtractAirgap) Describe() string {
	return fmt.Sprintf("extract airgap bundle %s to %s", a.BundlePath, a.WorkingDir)
}

func (a ExtractAirgap) Execute(ctx context.Context, _ orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if err := os.MkdirAll(a.WorkingDir, 0755); err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{}, airgap.ExtractTar(ctx, a.BundlePath, a.WorkingDir)
}
