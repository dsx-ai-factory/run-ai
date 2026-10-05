/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	"fmt"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func applyInferenceEndpoint(runner metav1.Object, config *workloadsapi.AgentInferenceEndpoint) error {
	if config == nil {
		return nil
	}
	// Require a wire version explicitly; the shared setter's authoring default
	// must not reinterpret missing or future control-plane versions as v1.
	if config.SchemaVersion != endpoints.SchemaVersionV1 {
		return fmt.Errorf("inference endpoint schemaVersion must be %q", endpoints.SchemaVersionV1)
	}
	spec, err := json.Marshal(config.Spec)
	if err != nil {
		return fmt.Errorf("encode inference endpoint spec: %w", err)
	}
	return endpoints.SetInferenceEndpoint(runner, &endpoints.InferenceEndpointConfig{
		SchemaVersion: config.SchemaVersion,
		Spec:          spec,
	})
}
