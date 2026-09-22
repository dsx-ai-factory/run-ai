/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/steps"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

func init() { builders = append(builders, controlPlaneAccessBuilder) }

func controlPlaneAccessBuilder(s *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step {
	if deps.RunAIClientWithCredentials == nil || s.Cluster == nil || !s.Cluster.UsesControlPlaneCredentials() {
		return nil
	}
	c := s.Cluster
	return steps.NewActionStepTitled("preflight-control-plane-access", "Control-plane access",
		actions.CheckControlPlaneAccess{
			ClusterName: c.Name,
			ControlPlane: actions.ControlPlaneRef{
				URL:        c.ControlPlaneURL,
				Domain:     c.Domain,
				CACertPath: controlPlaneCAPath(s),
				Credentials: &step.ControlPlaneCredentials{
					Username: c.ControlPlaneCredentials.Username,
					Password: c.ControlPlaneCredentials.Password,
					Token:    c.ControlPlaneCredentials.Token,
				},
			},
		})
}
