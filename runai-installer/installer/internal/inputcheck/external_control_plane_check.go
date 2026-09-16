/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"

func init() { checks = append(checks, externalControlPlaneCheck) }

// externalControlPlaneCheck surfaces every problem with a cluster-only install
// that points at an external control plane: the URL and cluster domain, and
// exactly one way of identifying the cluster there, either a pre-registered
// uid with its bootstrap token / client secret, or control-plane credentials
// the installer registers the cluster with. The rules live in the spec so the
// hard gate (spec.Validate) and this aggregated report cannot drift apart.
func externalControlPlaneCheck(s *spec.Spec) []Finding {
	var findings []Finding
	for _, p := range s.ExternalControlPlaneProblems() {
		findings = append(findings, Finding{Check: "external-control-plane", Field: p.Field, Message: p.Message})
	}
	return findings
}
