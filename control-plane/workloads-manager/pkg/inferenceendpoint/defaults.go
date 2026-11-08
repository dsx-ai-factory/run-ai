/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inferenceendpoint

import (
	"fmt"

	clustersapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

// PathPrefix is the request path whose requests go through the endpoint picker. The
// control plane sets it for the supported workload types; users do not choose it.
const PathPrefix = "/v1"

// Defaults are cluster administrator choices shared by supported workload
// adapters. They do not establish runtime, authentication or URL compatibility.
type Defaults struct {
	GatewayName      string
	GatewayNamespace string
	EPPProfile       string
}

// ResolveDefaults returns nil for clusters that do not declare this integration.
// A declared but unavailable or misconfigured integration returns an error;
// callers must not silently fall back after selecting inference routing.
func ResolveDefaults(status *clustersapi.ClusterReportedStatus) (*Defaults, error) {
	return resolveDefaults(status, true)
}

// ResolveDefaultsForUpdate is ResolveDefaults for a workload that already
// opted in. It does not require the operands to be ready: an operand that
// restarts or rolls out must not block ordinary updates of a running workload,
// and the update is queued for the cluster either way. The integration version and
// the Gateway settings are still checked, because the update rebuilds the block.
func ResolveDefaultsForUpdate(status *clustersapi.ClusterReportedStatus) (*Defaults, error) {
	return resolveDefaults(status, false)
}

func resolveDefaults(status *clustersapi.ClusterReportedStatus, requireReadyOperands bool) (*Defaults, error) {
	if status == nil || status.Config.Get() == nil {
		return nil, nil
	}
	config := status.Config.Get()
	version := config.InferenceEndpointSchemaVersion.Get()
	if version == nil {
		return nil, nil
	}
	if *version != runv1.InferenceEndpointIntegrationV1 {
		return nil, fmt.Errorf("unsupported inference endpoint integration %q", *version)
	}
	// Operand status is keyed by the operator's operand name, not the Deployment
	// name: the agent reports as "agent" while its Deployment is runai-agent.
	if requireReadyOperands {
		operands := status.GetOperands()
		for _, name := range []string{"agent", "anyworkload-controller", "inference-endpoint-controller"} {
			if operand, found := operands[name]; !found || !operand.Ready {
				return nil, fmt.Errorf("inference endpoint integration requires ready operand %s", name)
			}
		}
	}
	if enabled := config.GatewayAPIEnabled.Get(); enabled == nil || !*enabled {
		return nil, fmt.Errorf("inference endpoint integration requires an administrator-enabled Gateway")
	}
	name, namespace := config.GetGatewayName(), config.GetGatewayNamespace()
	if name == "" || len(validation.IsDNS1123Subdomain(name)) != 0 || namespace == "" || len(validation.IsDNS1123Label(namespace)) != 0 {
		return nil, fmt.Errorf("inference endpoint integration requires a valid administrator Gateway name and namespace")
	}
	return &Defaults{GatewayName: name, GatewayNamespace: namespace, EPPProfile: "gie-v1.5"}, nil
}
