package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/uninstall"
)

// filterRunAICRDs identifies Run:AI CRDs by matching the ".run.ai" API group
// suffix rather than by label (Run:AI Helm charts do not carry a consistent
// ownership label on CRDs).  All Run:AI CRDs follow the "<resource>.run.ai"
// naming convention, so a suffix match is both precise and safe — it avoids
// accidentally matching third-party CRDs whose names happen to contain "runai"
// as a substring.
func filterRunAICRDs(crds []string) []string {
	var result []string
	for _, crd := range crds {
		if strings.HasSuffix(strings.ToLower(crd), ".run.ai") {
			result = append(result, crd)
		}
	}
	return result
}

type DeleteRunAICRDs struct {
	*step.RunAIInstallStepImpl
	k8sWrapper uninstall.K8sWrapper
}

func NewDeleteRunAICRDs(cfg *config.UninstallConfig, k8sWrapper uninstall.K8sWrapper) *DeleteRunAICRDs {
	return &DeleteRunAICRDs{
		RunAIInstallStepImpl: step.NewRunAIInstallStepImpl("Delete RunAI CRDs", cfg.StepTimeout),
		k8sWrapper:           k8sWrapper,
	}
}

func (s *DeleteRunAICRDs) Execute(ctx context.Context, _ *model.RunAIInstallData) error {
	if err := s.CheckContextCancellation(ctx); err != nil {
		return err
	}
	s.SetStatus(model.Installing, nil)

	crds, err := s.k8sWrapper.ListCRDs()
	if err != nil {
		s.SetStatus(model.Failed, err)
		return fmt.Errorf("failed to list CRDs: %w", err)
	}

	runaiCRDs := filterRunAICRDs(crds)
	if len(runaiCRDs) == 0 {
		logging.Infof("[Uninstall] No RunAI CRDs found, skipping CRD deletion")
		s.SetStatus(model.Installed, nil)
		return nil
	}

	logging.Infof("[Uninstall] Deleting %d RunAI CRDs", len(runaiCRDs))
	for _, crdName := range runaiCRDs {
		if err := s.CheckContextCancellation(ctx); err != nil {
			return err
		}
		if err := s.k8sWrapper.DeleteCRD(ctx, crdName); err != nil {
			s.SetStatus(model.Failed, err)
			return fmt.Errorf("failed to delete CRD %s: %w", crdName, err)
		}
	}

	s.SetStatus(model.Installed, nil)
	return nil
}
