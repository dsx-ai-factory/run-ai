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
)

func init() { builders = append(builders, openShiftInferenceCertBuilder) }

// openShiftInferenceCertBuilder warns when an inference wildcard certificate is
// provided on OpenShift with Knative enabled. On OpenShift, Knative serving is
// exposed through OpenShift Routes (the router's certificate), so the OpenShift
// Serverless install path never consumes cluster.inference.wildcardTls. Surfacing
// this turns the silent no-op of a cert the user expected to take effect into an
// explicit warning. It returns nil when it does not apply.
func openShiftInferenceCertBuilder(s *spec.Spec, _ orchestrator2.Deps) orchestrator2.Step {
	if !s.IsOpenShift() {
		return nil
	}
	f := s.Features
	if f == nil || f.Skip || !f.Knative.IsInstall() {
		return nil
	}
	c := s.Cluster
	if c == nil || c.Inference == nil || c.Inference.WildcardTLS == nil || c.Inference.WildcardTLS.CertPath == "" {
		return nil
	}
	return steps.NewActionStepTitled("preflight-openshift-inference-certificate", "Inference certificate (OpenShift)",
		actions.CheckOpenShiftInferenceCertIgnored{})
}
