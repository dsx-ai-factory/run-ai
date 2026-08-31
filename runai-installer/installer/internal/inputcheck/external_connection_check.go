/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, externalConnectionCheck) }

// openShiftDistribution is the kubernetesDistribution value that selects
// OpenShift, where external routing uses Routes rather than Ingress/Gateway API.
const openShiftDistribution = "openshift"

func isOpenShift(s *spec.Spec) bool { return s.KubernetesDistribution == openShiftDistribution }

// isGatewayAPI reports whether external routing is configured to use Gateway API
// (the platform provides the Gateway; Run:ai/Knative attach HTTPRoutes to it).
func isGatewayAPI(s *spec.Spec) bool {
	return s.ExternalConnection != nil && s.ExternalConnection.Type == spec.ExternalConnectionGatewayAPI
}

// externalConnectionCheck validates the external routing configuration. On
// non-OpenShift installs it is required (ingress or gatewayApi). On OpenShift
// external routing uses Routes by default; gatewayApi may additionally be set,
// in which case both the OpenShift Route and Gateway API parameters are sent to
// the charts. The settings of the selected type must be complete, and the other
// type's settings must not be present.
func externalConnectionCheck(s *spec.Spec) []Finding {
	ec := s.ExternalConnection

	if isOpenShift(s) {
		if ec == nil {
			return nil
		}
		if ec.Type != spec.ExternalConnectionGatewayAPI {
			return []Finding{ecFinding("externalConnection.type",
				fmt.Sprintf("on OpenShift only %q may be set for externalConnection; other external routing uses OpenShift Routes", spec.ExternalConnectionGatewayAPI))}
		}
		return gatewayAPIFindings(ec)
	}

	// Non-OpenShift: only meaningful when something is actually being installed.
	if s.ControlPlane == nil && s.Cluster == nil {
		return nil
	}
	if ec == nil {
		return []Finding{ecFinding("externalConnection",
			"externalConnection is required: set type to ingress or gatewayApi")}
	}

	switch ec.Type {
	case spec.ExternalConnectionIngress:
		var findings []Finding
		if ec.GatewayAPI != nil {
			findings = append(findings, ecFinding("externalConnection.gatewayApi", "gatewayApi must not be set when type is ingress"))
		}
		if ec.Ingress == nil || ec.Ingress.ClassName == "" {
			findings = append(findings, ecFinding("externalConnection.ingress.className", "ingress.className is required when type is ingress"))
		}
		return findings
	case spec.ExternalConnectionGatewayAPI:
		return gatewayAPIFindings(ec)
	case spec.ExternalConnectionRoute:
		return []Finding{ecFinding("externalConnection.type",
			"route is selected automatically on OpenShift; on other distributions set type to ingress or gatewayApi")}
	default:
		return []Finding{ecFinding("externalConnection.type",
			fmt.Sprintf("invalid externalConnection type %q: must be %q or %q", ec.Type, spec.ExternalConnectionIngress, spec.ExternalConnectionGatewayAPI))}
	}
}

func gatewayAPIFindings(ec *spec.ExternalConnection) []Finding {
	var findings []Finding
	if ec.Ingress != nil {
		findings = append(findings, ecFinding("externalConnection.ingress", "ingress must not be set when type is gatewayApi"))
	}
	if ec.GatewayAPI == nil || ec.GatewayAPI.Name == "" || ec.GatewayAPI.Namespace == "" {
		findings = append(findings, ecFinding("externalConnection.gatewayApi", "gatewayApi.name and gatewayApi.namespace are required when type is gatewayApi"))
	}
	return findings
}

func ecFinding(field, msg string) Finding {
	return Finding{Check: "external-connection", Field: field, Message: msg}
}
