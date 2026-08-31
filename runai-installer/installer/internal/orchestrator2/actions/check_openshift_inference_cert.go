/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// CheckOpenShiftInferenceCertIgnored is a read-only preflight check that warns
// when a spec-provided inference wildcard certificate is set on OpenShift. There
// Knative serving is exposed through OpenShift Routes, which terminate TLS with
// the OpenShift router's certificate; the OpenShift Serverless install path never
// consumes cluster.inference.wildcardTls, so the certificate silently has no
// effect. It is only assembled (by the preflight builder) when that certificate
// is set on OpenShift with Knative enabled, so its Execute always warns.
type CheckOpenShiftInferenceCertIgnored struct{}

func (a CheckOpenShiftInferenceCertIgnored) Kind() string {
	return "preflight.openshift_inference_cert_ignored"
}

func (a CheckOpenShiftInferenceCertIgnored) Args() map[string]any { return map[string]any{} }

func (a CheckOpenShiftInferenceCertIgnored) Describe() string {
	return "verify the inference wildcard certificate is used (warning only on OpenShift, where Routes serve the router certificate)"
}

func (a CheckOpenShiftInferenceCertIgnored) Execute(_ context.Context, _ orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	return orchestrator2.ActionResult{
		Status:      orchestrator2.StatusWarn,
		Reason:      "cluster.inference.wildcardTls is set but has no effect on OpenShift: Knative serving is exposed through OpenShift Routes, which terminate TLS with the OpenShift router's certificate",
		Remediation: "configure the inference wildcard certificate on the OpenShift router/IngressController, or via features.knative.openshift.customResource; the installer does not apply cluster.inference.wildcardTls on OpenShift",
	}, nil
}
