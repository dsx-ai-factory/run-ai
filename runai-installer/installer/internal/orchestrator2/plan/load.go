/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plan

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/steps"
)

// imagePushTimeout is added to the per-step timeout for the image push: like
// the pack's download it is dominated by network throughput, so it far exceeds
// the per-step norm.
const imagePushTimeout = 120 * time.Minute

// buildAirgapLoad assembles the preflight and bundle-loading plans. Preflight
// is read-only, so it runs for real even under --dry-run.
func buildAirgapLoad(s *spec.Spec) (preflight, load *orchestrator2.Plan) {
	l := s.AirgapLoad

	preflight = &orchestrator2.Plan{Pipelines: []*orchestrator2.Pipeline{{
		Name:  "preflight",
		Title: "Preflight checks",
		Flows: []*orchestrator2.Flow{{
			Name:  "preflight-checks",
			Title: "Running checks",
			Steps: []orchestrator2.Step{
				airgapLoadPreflightStep(l),
			},
		}},
	}}}

	load = &orchestrator2.Plan{Pipelines: []*orchestrator2.Pipeline{{
		Name:   "load",
		Title:  "Air-gapped load",
		DryRun: s.DryRun,
		Flows:  []*orchestrator2.Flow{airgapLoadFlow(l)},
	}}}
	return preflight, load
}

func airgapLoadPreflightStep(l *spec.AirgapLoad) orchestrator2.Step {
	return steps.NewActionStepTitled("preflight-load-registry", "Destination registry", actions.CheckLoadRegistry{
		Destination: l.Destination,
	})
}

func airgapLoadFlow(l *spec.AirgapLoad) *orchestrator2.Flow {
	return &orchestrator2.Flow{
		Name:  "airgap-load",
		Title: "Air-gapped bundle load",
		Steps: []orchestrator2.Step{
			steps.NewActionStepTitled("extract-bundle", "Extracting the bundle", actions.ExtractBundle{
				BundlePath: l.BundlePath,
				WorkingDir: l.WorkingDir,
			}),
			steps.WithExtraTimeout(
				steps.NewActionStepTitled("push-images", "Pushing container images (this can take a while)", actions.PushImages{
					WorkingDir:  l.WorkingDir,
					Destination: l.Destination,
				}),
				imagePushTimeout,
			),
			steps.NewActionStepTitled("generate-values", "Generating Helm values", actions.GenerateAirgapValues{
				WorkingDir:  l.WorkingDir,
				ValuesDir:   l.ValuesDir,
				Destination: l.Destination,
			}),
		},
	}
}
