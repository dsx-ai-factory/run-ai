package inputcheck

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() {
	checks = append(checks, airgapForbidsDependencies)
}

// airgapForbidsDependencies rejects enabling prerequisites or features on an
// air-gapped install: the offline bundle carries only the Run:ai charts, so
// third-party components cannot be pulled. One finding is emitted per group
// (prerequisites / features) listing the enabled members, since the group
// shares one remediation (prerequisites.skip / features.skip).
func airgapForbidsDependencies(s *spec.Spec) []Finding {
	if !s.IsAirGapped() {
		return nil
	}

	var findings []Finding

	if p := s.Prerequisites; p != nil && !p.Skip {
		enabled := enabledDependencies([]namedDependency{
			{"prometheus", p.Prometheus},
			{"gpuOperator", p.GPUOperator},
			{"networkOperator", p.NetworkOperator},
		})
		if len(enabled) > 0 {
			findings = append(findings, Finding{
				Check: "airgap-no-prerequisites",
				Field: "prerequisites",
				Message: fmt.Sprintf(
					"the installer does not support installing prerequisites in air-gapped mode (the offline bundle contains only the Run:ai charts); set prerequisites.skip or the mode of each enabled dependency to skip. Enabled: %s",
					strings.Join(enabled, ", "),
				),
			})
		}
	}

	if f := s.Features; f != nil && !f.Skip {
		enabled := enabledDependencies([]namedDependency{
			{"knative", f.Knative},
			{"kubeflow", f.Kubeflow},
			{"mpiOperator", f.MPIOperator},
			{"lws", f.LWS},
			{"nimOperator", f.NIMOperator},
			{"grove", f.Grove},
			{"dynamo", f.Dynamo},
		})
		if len(enabled) > 0 {
			findings = append(findings, Finding{
				Check: "airgap-no-features",
				Field: "features",
				Message: fmt.Sprintf(
					"the installer does not support installing features in air-gapped mode (the offline bundle contains only the Run:ai charts); set features.skip or the mode of each enabled feature to skip. Enabled: %s",
					strings.Join(enabled, ", "),
				),
			})
		}
	}

	return findings
}

type namedDependency struct {
	name string
	dep  *spec.Dependency
}

// enabledDependencies returns the names of the dependencies set to install,
// preserving the given order.
func enabledDependencies(deps []namedDependency) []string {
	var names []string
	for _, d := range deps {
		if d.dep.IsInstall() {
			names = append(names, d.name)
		}
	}
	return names
}
