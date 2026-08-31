/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/inputs/validation"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	pkgvalidation "github.com/run-ai/runai/runai-installer/installer/pkg/validation"
)

func init() { checks = append(checks, workloadRoutingCheck) }

// workloadRoutingCheck validates the cluster's workload-routing selection and
// guards against it being contradicted by a subdomainSupport custom helm value.
// The first-class cluster.workloadRouting setting and the chart value must not
// disagree: the installer emits subdomainSupport from the resolved routing mode,
// so a conflicting override would make the effective routing ambiguous.
func workloadRoutingCheck(s *spec.Spec) []Finding {
	c := s.Cluster
	if c == nil {
		return nil
	}

	var findings []Finding

	if res := validation.ValidateWorkloadRouting(c.WorkloadRouting); !res.Status {
		findings = append(findings, Finding{
			Check:   "workload-routing",
			Field:   "cluster.workloadRouting",
			Message: res.Message,
		})
	}

	overrideHost, present, valid := c.SubdomainOverride()
	if !present || !valid {
		return findings
	}

	var explicitHost bool
	switch c.WorkloadRouting {
	case pkgvalidation.WorkloadRoutingHost:
		explicitHost = true
	case pkgvalidation.WorkloadRoutingPath:
		explicitHost = false
	default:
		return findings // not explicitly set: the override drives routing, no conflict
	}

	if explicitHost != overrideHost {
		findings = append(findings, Finding{
			Check: "workload-routing",
			Field: "cluster.customHelmValues." + spec.SubdomainSupportHelmKey,
			Message: fmt.Sprintf(
				"cluster.workloadRouting is %q but customHelmValues sets %s=%t; they must agree",
				c.WorkloadRouting, spec.SubdomainSupportHelmKey, overrideHost,
			),
		})
	}

	return findings
}
