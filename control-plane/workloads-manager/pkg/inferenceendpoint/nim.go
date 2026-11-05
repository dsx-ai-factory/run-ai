/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inferenceendpoint

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
)

const nimServiceKind = "NIM service"

// BuildNIMServiceIntent returns the externalAccess.inference block for the serving
// endpoint of a NIMService. The NIM operator creates the serving Service, named
// after the workload, so the endpoint references it instead of creating one; the
// networking plugin reads the workload selector and target port from that live
// Service, which also covers multi-node NIM, where the operator selects the leaders.
// This function checks that the endpoint references that Service and is exposed
// only through the administrator Gateway at a root hostname.
// Only public serving is representable in this version: protected NIM puts the
// auth proxy in front of the Service, and the endpoint picker would bypass it.
// It does not establish trusted submission origin or the serving protocol; the
// caller must establish both, because gRPC serving is not visible in the endpoint.
func BuildNIMServiceIntent(endpoint api.NetworkingEndpoint, workloadName string, rewrites []api.ModelRewrite, defaults *Defaults) (*api.ExternalAccessInference, error) {
	if !validDefaults(defaults) {
		return nil, fmt.Errorf("NIM service requires valid administrator Gateway defaults and the gie-v1.5 EPP profile")
	}
	if endpoint.Services == nil || len(*endpoint.Services) != 1 {
		return nil, fmt.Errorf("NIM service requires exactly one referenced serving Service")
	}
	service := (*endpoint.Services)[0]
	if service.Spec != nil || service.Ref == nil || workloadName == "" || service.Ref.Name != workloadName ||
		service.Ref.Port < 1 || service.Ref.Port > 65535 {
		return nil, fmt.Errorf("NIM service requires a reference to the serving port of the operator Service %q", workloadName)
	}
	if service.Authorization != nil {
		return nil, fmt.Errorf("NIM service with a protected serving port is not supported yet")
	}
	if _, err := servingHostname(nimServiceKind, endpoint.ExternalAccess, service.Ref.Name, service.Ref.Port, defaults); err != nil {
		return nil, err
	}
	return newInferenceIntent(defaults, rewrites), nil
}
