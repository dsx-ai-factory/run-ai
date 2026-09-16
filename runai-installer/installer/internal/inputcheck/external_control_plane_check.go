/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"

func init() { checks = append(checks, externalControlPlaneCheck) }

func externalControlPlaneCheck(s *spec.Spec) []Finding {
	var findings []Finding
	for _, p := range s.ExternalControlPlaneProblems() {
		findings = append(findings, Finding{Check: "external-control-plane", Field: p.Field, Message: p.Message})
	}
	return findings
}
