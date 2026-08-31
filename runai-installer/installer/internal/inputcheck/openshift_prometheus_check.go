/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() {
	checks = append(checks, openShiftForbidsPrometheus)
}

// openShiftForbidsPrometheus rejects requesting the Prometheus prerequisite on
// OpenShift. OpenShift ships its own monitoring stack (openshift-monitoring),
// which Run:ai integrates with; installing a second Prometheus is unsupported.
// The openshift-monitoring stack is verified by a preflight check instead.
func openShiftForbidsPrometheus(s *spec.Spec) []Finding {
	if !s.IsOpenShift() {
		return nil
	}
	p := s.Prerequisites
	if p == nil || p.Skip || !p.Prometheus.IsInstall() {
		return nil
	}
	return []Finding{{
		Check:   "openshift-no-prometheus",
		Field:   "prerequisites.prometheus",
		Message: "installing Prometheus is not supported on OpenShift; Run:ai requires the built-in OpenShift monitoring stack (openshift-monitoring) instead — set prerequisites.prometheus.mode to skip",
	}}
}
