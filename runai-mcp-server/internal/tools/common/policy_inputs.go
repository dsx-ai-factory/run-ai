/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"strings"

	policy "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

const (
	PolicyVersionLegacy = "legacy"
	PolicyVersionGVK    = "gvk"
)

// GVK identifies a Kubernetes workload type in policy API requests.
type GVK struct {
	Group   string `json:"group,omitempty" jsonschema:"Kubernetes API group, for example apps. Omit for core resources such as v1/Pod."`
	Version string `json:"version" jsonschema:"Required Kubernetes API version, for example v1."`
	Kind    string `json:"kind" jsonschema:"Required Kubernetes resource kind, for example Deployment or Pod."`
}

func (g GVK) Validate() error {
	if strings.TrimSpace(g.Version) == "" {
		return fmt.Errorf("workloadType.version is required")
	}
	if strings.TrimSpace(g.Kind) == "" {
		return fmt.Errorf("workloadType.kind is required")
	}
	return nil
}

func ValidatePolicyScope(scope, cluster, department, project string) error {
	switch policy.ScopeType(scope) {
	case policy.SCOPETYPE_TENANT:
		return nil
	case policy.SCOPETYPE_CLUSTER:
		return ValidateUUID("clusterId", cluster)
	case policy.SCOPETYPE_DEPARTMENT:
		if strings.TrimSpace(department) == "" {
			return fmt.Errorf("departmentId is required")
		}
	case policy.SCOPETYPE_PROJECT:
		return ValidateNumericString("projectId", project)
	default:
		return fmt.Errorf("scope must be tenant, cluster, department, or project")
	}
	return nil
}
