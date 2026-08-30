package inputcheck

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, kubernetesDistributionCheck) }

func kubernetesDistributionCheck(s *spec.Spec) []Finding {
	return fromResult(
		validation.ValidateKubernetesDistribution(s.KubernetesDistribution),
		"kubernetes-distribution", "install.kubernetesDistribution",
	)
}
