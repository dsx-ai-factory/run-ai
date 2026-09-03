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

func init() { builders = append(builders, controlPlaneCertificateBuilder) }

// controlPlaneCertificateBuilder assembles the control-plane certificate check.
// It applies only to an existing Gateway, whose listener certificate the
// installer does not control; ingress (installer-provided secret) and OpenShift
// (Routes) are skipped. Needs a TLSProber.
func controlPlaneCertificateBuilder(s *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step {
	if deps.TLSProber == nil || s.IsOpenShift() {
		return nil
	}
	ec := s.ExternalConnection
	if ec == nil || ec.Type != spec.ExternalConnectionGatewayAPI || ec.GatewayAPI == nil {
		return nil
	}
	host := controlPlaneHost(s)
	if host == "" {
		return nil
	}
	return steps.NewActionStepTitled("preflight-control-plane-certificate", "Control-plane certificate",
		actions.CheckControlPlaneCertificate{Host: host, CAPath: controlPlaneCAPath(s)})
}

// controlPlaneHost is the domain the control-plane endpoint is served on: the
// control-plane domain, falling back to the cluster domain.
func controlPlaneHost(s *spec.Spec) string {
	if s.ControlPlane != nil && s.ControlPlane.Domain != "" {
		return s.ControlPlane.Domain
	}
	if s.Cluster != nil {
		return s.Cluster.Domain
	}
	return ""
}

// controlPlaneCAPath is the CA the served certificate is verified against: the
// control-plane TLS CA chain, falling back to the cluster TLS CA chain. Empty
// means system roots (a publicly trusted certificate).
func controlPlaneCAPath(s *spec.Spec) string {
	if tls := s.ControlPlaneTLS(); tls != nil {
		return tls.CAChainPath
	}
	if tls := s.ClusterTLS(); tls != nil {
		return tls.CAChainPath
	}
	return ""
}
