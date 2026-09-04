/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs/validation"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, fqdnCheck) }

// fqdnCheck validates the format of every domain the spec carries. It only
// checks values that are present: required-ness is the structural validator's
// job, and an empty cluster domain legitimately falls back to the control-plane
// domain.
func fqdnCheck(s *spec.Spec) []Finding {
	var findings []Finding

	if cp := s.ControlPlane; cp != nil && cp.Domain != "" {
		findings = append(findings, fromResult(
			validation.ValidateFQDNFormat(cp.Domain),
			"fqdn-format", "controlPlane.domain",
		)...)
	}

	if c := s.Cluster; c != nil {
		if c.Domain != "" {
			findings = append(findings, fromResult(
				validation.ValidateFQDNFormat(c.Domain),
				"fqdn-format", "cluster.domain",
			)...)
		}
		if c.Inference != nil && c.Inference.WildcardDomain != "" {
			findings = append(findings, fromResult(
				validation.ValidateWildcardFQDNFormat(c.Inference.WildcardDomain),
				"fqdn-format", "cluster.inference.wildcardDomain",
			)...)
		}
	}

	return findings
}
