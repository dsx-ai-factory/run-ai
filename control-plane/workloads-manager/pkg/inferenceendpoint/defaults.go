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
	operands := status.GetOperands()
	for _, name := range []string{"runai-agent", "anyworkload-controller", "inference-endpoint-controller"} {
		if operand, found := operands[name]; !found || !operand.Ready {
			return nil, fmt.Errorf("inference endpoint integration requires ready operand %s", name)
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
