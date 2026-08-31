/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

const controlPlaneTLSPort = 443

// CheckControlPlaneCertificate is a read-only preflight check that verifies the
// certificate the Gateway front door serves is trusted for the control-plane
// domain (the provided CA, or system roots). Reported as a warning, not a hard
// failure.
type CheckControlPlaneCertificate struct {
	Host   string
	CAPath string
}

func (a CheckControlPlaneCertificate) Kind() string { return "preflight.control_plane_certificate" }

func (a CheckControlPlaneCertificate) Args() map[string]any {
	return map[string]any{"host": a.Host}
}

func (a CheckControlPlaneCertificate) Describe() string {
	return fmt.Sprintf("verify the control-plane endpoint serves a trusted certificate for %s", a.Host)
}

func (a CheckControlPlaneCertificate) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.TLSProber == nil {
		return orchestrator2.ActionResult{}, nil
	}
	res, err := deps.TLSProber.ProbeCertificate(ctx, a.Host, controlPlaneTLSPort, a.CAPath)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not verify the control-plane certificate for %s: %v", a.Host, err),
			Remediation: "check that the provided CA file is readable and valid",
		}, nil
	}
	if !res.Reachable {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("the control-plane endpoint https://%s is not reachable yet (%s); its certificate will be verified again during install", a.Host, res.Detail),
			Remediation: fmt.Sprintf("ensure DNS for %s resolves to the ingress/gateway front door and it is reachable from where the installer runs", a.Host),
		}, nil
	}
	if !res.Verified {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("the control-plane endpoint https://%s serves a certificate that is not trusted for it (%s)", a.Host, res.Detail),
			Remediation: "the Gateway listener — not the installer — terminates external TLS, so it must present a certificate signed by the provided CA (controlPlane.tls.caChainPath) and covering the domain; otherwise the control-plane health check may fail during install",
		}, nil
	}
	return orchestrator2.ActionResult{}, nil
}
