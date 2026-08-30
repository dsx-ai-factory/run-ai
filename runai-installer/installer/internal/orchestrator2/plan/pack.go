package plan

import (
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/steps"
)

// imageDownloadTimeout is added to the per-step timeout for the image pull: it
// is dominated by network throughput and disk speed, so it far exceeds the
// per-step norm.
const imageDownloadTimeout = 120 * time.Minute

// buildAirgapPack assembles the preflight and bundle-building plans. Preflight
// is read-only, so it runs for real even under --dry-run.
func buildAirgapPack(s *spec.Spec) (preflight, pack *orchestrator2.Plan) {
	p := s.AirgapPack

	preflight = &orchestrator2.Plan{Pipelines: []*orchestrator2.Pipeline{{
		Name:  "preflight",
		Title: "Preflight checks",
		Flows: []*orchestrator2.Flow{{
			Name:  "preflight-checks",
			Title: "Running checks",
			Steps: []orchestrator2.Step{
				steps.NewActionStepTitled("preflight-pack-sources", "Artifact sources", actions.CheckPackSources{
					Lock:           p.Lock,
					Credentials:    actions.ToCredentials(p.Credentials),
					ImageOverrides: p.ImageOverrides,
				}),
			},
		}},
	}}}

	flow := &orchestrator2.Flow{
		Name:  "airgap-pack",
		Title: "Air-gapped bundle creation",
		Steps: []orchestrator2.Step{
			steps.NewActionStepTitled("resolve-artifacts", "Resolving artifact inventory", actions.ResolveArtifacts{
				Lock:           p.Lock,
				WorkDir:        p.WorkDir,
				ImageOverrides: p.ImageOverrides,
			}),
			steps.WithExtraTimeout(
				steps.NewActionStepTitled("download-images", "Downloading container images (this can take a while)", actions.PullImages{
					WorkDir: p.WorkDir,
				}),
				imageDownloadTimeout,
			),
			steps.NewActionStepTitled("download-charts", "Downloading Helm charts", actions.PullCharts{
				WorkDir: p.WorkDir,
			}),
			steps.NewActionStepTitled("download-clis", "Downloading CLI artifacts", actions.PullCLIs{
				WorkDir: p.WorkDir,
			}),
			steps.NewActionStepTitled("assemble-bundle", "Assembling the bundle", actions.AssembleBundle{
				WorkDir:    p.WorkDir,
				OutputPath: p.OutputPath,
			}),
		},
	}

	pack = &orchestrator2.Plan{Pipelines: []*orchestrator2.Pipeline{{
		Name:   "pack",
		Title:  "Air-gapped bundle",
		DryRun: s.DryRun,
		Flows:  []*orchestrator2.Flow{flow},
	}}}
	return preflight, pack
}
