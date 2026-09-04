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

func init() { builders = append(builders, externalConnectionBuilder) }

// externalConnectionBuilder assembles the external-connection readiness check:
// on an ingress install it verifies the configured IngressClass exists; on a
// Gateway API install it verifies the Gateway API is installed and the
// referenced Gateway is programmed. The installer brings neither the ingress
// controller nor the Gateway, so both are preconditions.
//
// It is skipped on OpenShift (routing there uses Routes) and returns nil when no
// ClusterProbe is wired or the external connection is not configured.
func externalConnectionBuilder(s *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step {
	if deps.ClusterProbe == nil || s.IsOpenShift() {
		return nil
	}
	ec := s.ExternalConnection
	if ec == nil {
		return nil
	}

	switch ec.Type {
	case spec.ExternalConnectionIngress:
		if ec.Ingress == nil || ec.Ingress.ClassName == "" {
			return nil
		}
		return steps.NewActionStepTitled("preflight-external-connection", "External connection (ingress)",
			actions.CheckExternalConnection{IngressClassName: ec.Ingress.ClassName})
	case spec.ExternalConnectionGatewayAPI:
		if ec.GatewayAPI == nil || ec.GatewayAPI.Name == "" || ec.GatewayAPI.Namespace == "" {
			return nil
		}
		return steps.NewActionStepTitled("preflight-external-connection", "External connection (Gateway API)",
			actions.CheckExternalConnection{GatewayName: ec.GatewayAPI.Name, GatewayNamespace: ec.GatewayAPI.Namespace})
	default:
		return nil
	}
}
