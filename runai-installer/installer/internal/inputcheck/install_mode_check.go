/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, installModeCheck) }

// installModeCheck enforces that the declared install.mode matches the
// configured components, so what gets installed follows the user's stated
// intent rather than whichever config blocks happen to be present:
//
//   - control_plane_only: a control plane must be configured and a cluster must
//     not.
//   - cluster_only: a cluster must be configured and a control plane must not.
//   - full_installation (and the empty default): no component exclusivity — the
//     configured components install as-is.
//
// Prerequisites and features are unaffected by the mode; they are governed by
// their own checks.
func installModeCheck(s *spec.Spec) []Finding {
	mode := s.InstallMode
	if !mode.IsKnown() {
		return []Finding{{
			Check: "install-mode",
			Field: "install.mode",
			Message: fmt.Sprintf("unknown install mode %q: must be one of %q, %q, %q",
				string(mode), spec.InstallModeFull, spec.InstallModeControlPlaneOnly, spec.InstallModeClusterOnly),
		}}
	}

	switch mode {
	case spec.InstallModeControlPlaneOnly:
		return controlPlaneOnlyFindings(s)
	case spec.InstallModeClusterOnly:
		return clusterOnlyFindings(s)
	}
	return nil
}

func controlPlaneOnlyFindings(s *spec.Spec) []Finding {
	var findings []Finding

	if s.ControlPlane == nil {
		findings = append(findings, Finding{
			Check:   "install-mode",
			Field:   "controlPlane.enabled",
			Message: "control-plane-only install requires the control plane: set controlPlane.enabled to true",
		})
	}
	if s.Cluster != nil {
		findings = append(findings, Finding{
			Check:   "install-mode",
			Field:   "cluster",
			Message: "control-plane-only install must not configure a cluster: remove the cluster block (or set cluster.enabled to false), or change install.mode to full_installation",
		})
	}

	return findings
}

func clusterOnlyFindings(s *spec.Spec) []Finding {
	var findings []Finding

	if s.Cluster == nil {
		findings = append(findings, Finding{
			Check:   "install-mode",
			Field:   "cluster.enabled",
			Message: "cluster-only install requires the cluster: set cluster.enabled to true",
		})
	}
	if s.ControlPlane != nil {
		findings = append(findings, Finding{
			Check:   "install-mode",
			Field:   "controlPlane",
			Message: "cluster-only install must not configure a control plane: remove the controlPlane block (or set controlPlane.enabled to false), or change install.mode to full_installation",
		})
	}

	return findings
}
