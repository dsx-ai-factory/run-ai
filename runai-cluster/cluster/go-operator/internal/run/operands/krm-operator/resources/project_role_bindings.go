/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

// runaiProjectRoleBindings are the RoleBindings the project-controller replicates
// into every project namespace. KRM builds them into the ConfigMap the controller
// reads, so they move here from the ConfigMap the operand used to write.
//
// KRM binds every subject to a service account in the install namespace. The
// OpenShift bindings runai also ships cannot be expressed that way -- three bind
// the project namespace's own default service account, and one binds
// prometheus-k8s in the OpenShift monitoring namespace -- so they are not carried
// over. Those remain outstanding for OpenShift installs.
func runaiProjectRoleBindings(runaiConfig *runaiv1.RunaiConfig) []kaiv1alpha1.ProjectRoleBinding {
	bindings := []kaiv1alpha1.ProjectRoleBinding{
		{Name: "runai-accessrule-controller-project", ServiceAccountName: operands.RunaiAccessRuleControllerName},
		{Name: "researcher-service-rw", ServiceAccountName: "researcher-service"},
		{Name: "cluster-api-project", ServiceAccountName: "cluster-api"},
		{Name: "runai-scheduler-rw", ServiceAccountName: "runai-scheduler"},
		{Name: "runai-project-controller-project", ServiceAccountName: operands.RunaiProjectControllerName},
		{Name: "runai-admission-controller-project", ServiceAccountName: "runai-admission-controller"},
		{Name: "runai-binder-project", ServiceAccountName: "binder"},
		{Name: "status-updater", ServiceAccountName: "status-updater"},
		{Name: "pod-group-controller", ServiceAccountName: "podgroup-controller"},
		{Name: "numa-placement-exporter", ServiceAccountName: "numa-placement-exporter"},
		// Below, the binding and the ClusterRole are named differently.
		{
			Name:               "external-workload-integrator",
			ClusterRoleName:    "runai-external-workload-integrator-project",
			ServiceAccountName: "external-workload-integrator",
		},
		{Name: "cluster-api", ClusterRoleName: "runai-cluster-api-project", ServiceAccountName: "cluster-api"},
		{
			Name:               "shared-objects-controller",
			ClusterRoleName:    "runai-shared-objects-controller-project",
			ServiceAccountName: "shared-objects-controller",
		},
		{
			Name:               "advice-restart-controller",
			ClusterRoleName:    "runai-advice-restart-controller-project",
			ServiceAccountName: "advice-restart-controller",
		},
		{Name: "cluster-sync", ClusterRoleName: "cluster-sync-project", ServiceAccountName: "cluster-sync"},
		{Name: "runai-agent", ClusterRoleName: "runai-agent-project", ServiceAccountName: "runai-agent"},
		{
			Name:               "runai-agent-additional-workload-gvks",
			ClusterRoleName:    runaiConfig.Spec.ExternalWorkloadIntegrator.RIDRbacConfig.ClusterRoleCrudName,
			ServiceAccountName: "runai-agent",
		},
		{
			Name:               "anyworkload-controller",
			ClusterRoleName:    "runai-anyworkload-controller-project",
			ServiceAccountName: "anyworkload-controller",
		},
	}

	// The per-project LimitRange binding is not listed: KRM ships its own builtin
	// under the same feature flag, bound to the ClusterRole its chart renders.
	return bindings
}
