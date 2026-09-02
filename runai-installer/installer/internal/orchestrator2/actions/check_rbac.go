/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

// requiredPermissions is the fixed, minimal set of cluster-scoped permissions
// every install needs: it creates namespaces, secrets, CRDs, and cluster roles
// (the latter via the Helm charts). This is intentionally a static set rather
// than one derived from the resolved plan; plan-derived permissions are a later
// refinement.
var requiredPermissions = []orchestrator2.ResourceAttributes{
	{Verb: "create", Resource: "namespaces"},
	{Verb: "create", Resource: "secrets"},
	{Verb: "create", Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"},
	{Verb: "delete", Group: "apiextensions.k8s.io", Resource: "customresourcedefinitions"},
	{Verb: "create", Group: "rbac.authorization.k8s.io", Resource: "clusterroles"},
	{Verb: "list", Group: "storage.k8s.io", Resource: "storageclasses"},
}

// CheckRBAC is a read-only preflight check that confirms the caller holds the
// permissions the installer needs, via SelfSubjectAccessReview (no writes).
// Missing permissions are reported as a warning, not a failure: the
// SelfSubjectAccessReview is an approximation (it does not account for every
// resource the charts touch), so it must not hard-block an install that would
// otherwise succeed.
type CheckRBAC struct{}

func (a CheckRBAC) Kind() string { return "preflight.rbac" }

func (a CheckRBAC) Args() map[string]any { return map[string]any{} }

func (a CheckRBAC) Describe() string {
	return "verify the current user has the permissions the installer needs"
}

func (a CheckRBAC) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	var denied []string
	for _, ra := range requiredPermissions {
		allowed, _, err := deps.ClusterProbe.CanI(ctx, ra)
		if err != nil {
			return orchestrator2.ActionResult{
				Status:      orchestrator2.StatusWarn,
				Reason:      fmt.Sprintf("could not evaluate permissions: %v", err),
				Remediation: "check the kubeconfig/context and that the API server is reachable",
			}, nil
		}
		if !allowed {
			denied = append(denied, describePermission(ra))
		}
	}
	if len(denied) > 0 {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      "missing required permissions: " + strings.Join(denied, ", "),
			Remediation: "grant the current user the listed permissions (e.g. bind a role with cluster-admin)",
		}, nil
	}
	return orchestrator2.ActionResult{}, nil
}

func describePermission(ra orchestrator2.ResourceAttributes) string {
	resource := ra.Resource
	if ra.Group != "" {
		resource = ra.Resource + "." + ra.Group
	}
	return ra.Verb + " " + resource
}
