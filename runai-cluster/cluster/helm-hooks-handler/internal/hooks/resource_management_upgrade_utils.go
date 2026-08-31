/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hooks

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	schedv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// OldDepartmentCRDName is the name of the legacy department CRD. The
	// department migration is a no-op when it is not installed (fresh installs).
	OldDepartmentCRDName = "departments.scheduling.run.ai"

	// OldProjectCRDName is the name of the legacy project CRD. The project
	// migration is a no-op when it is not installed (fresh installs).
	OldProjectCRDName = "projects.run.ai"

	// OldNodePoolCRDName is the name of the legacy nodepool CRD. The nodepool
	// migration is a no-op when it is not installed (fresh installs).
	OldNodePoolCRDName = "nodepools.run.ai"

	// OldManagedNodesConfigCRDName is the name of the legacy managed-nodes-config
	// CRD. The migration is a no-op when it is not installed (fresh installs).
	OldManagedNodesConfigCRDName = "managednodesconfigs.run.ai"

	// ListPageSize bounds each List response so a cluster with
	// many departments is paged through instead of fetched in a single call.
	ListPageSize = 500

	// lastSyncTimestampAnnotation holds the value moved off the old department's /
	// project's status.syncStatus.lastSyncTimestamp (the new CRD has no such status
	// field). It is stored as an annotation, not a label - it is a free-form
	// timestamp that nothing selects on.
	LastSyncTimestampAnnotation = "run.ai/last-sync-timestamp"

	// ProjectControllerOperandName is the app-label prefix of the project-controller
	// pods the migration waits to disappear after disabling it. go-operator owns and
	// scales it from RunaiConfig, so the migration takes it down by disabling it in
	// RunaiConfig (same mechanism pre-crd-upgrade uses for the nodepool-controller)
	// rather than scaling the Deployment directly (which go-operator would revert).
	ProjectControllerOperandName = "runai-project-controller"

	// AgentOperandName is the app-label prefix of the runai-agent pods the
	// migration waits to disappear after disabling it. Like the project-controller
	// it is owned and scaled by go-operator from RunaiConfig, so the migration
	// takes it down by disabling it in RunaiConfig rather than scaling the
	// Deployment directly (which go-operator would revert).
	AgentOperandName = "runai-agent"

	// NodePoolControllerOperandName is the app-label prefix of the legacy
	// nodepool-controller pods. It owns the SchedulingShards being re-parented, so
	// the migration takes it down (via RunaiConfig, like the others) to keep it from
	// re-stamping the old ownership.
	NodePoolControllerOperandName = "nodepool-controller"
)

// ListLegacyDepartments returns every legacy (scheduling.run.ai/v2alpha2)
// department on the cluster, paging through large lists.
func ListLegacyDepartments(ctx context.Context, k8sClient client.Client) ([]schedv2alpha2.Department, error) {
	var all []schedv2alpha2.Department
	listOpts := []client.ListOption{client.Limit(ListPageSize)}
	for {
		var departments schedv2alpha2.DepartmentList
		if err := k8sClient.List(ctx, &departments, listOpts...); err != nil {
			return nil, fmt.Errorf("list legacy departments: %w", err)
		}
		all = append(all, departments.Items...)
		if departments.Continue == "" {
			break
		}
		listOpts = []client.ListOption{client.Limit(ListPageSize), client.Continue(departments.Continue)}
	}
	return all, nil
}

// ListLegacyProjects returns every legacy (run.ai/v2) project on the cluster,
// paging through large lists.
func ListLegacyProjects(ctx context.Context, k8sClient client.Client) ([]runaiv2.Project, error) {
	var all []runaiv2.Project
	listOpts := []client.ListOption{client.Limit(ListPageSize)}
	for {
		var projects runaiv2.ProjectList
		if err := k8sClient.List(ctx, &projects, listOpts...); err != nil {
			return nil, fmt.Errorf("list legacy projects: %w", err)
		}
		all = append(all, projects.Items...)
		if projects.Continue == "" {
			break
		}
		listOpts = []client.ListOption{client.Limit(ListPageSize), client.Continue(projects.Continue)}
	}
	return all, nil
}

// ListLegacyNodePools returns every legacy (run.ai/v1alpha1) nodepool on the
// cluster, paging through large lists.
func ListLegacyNodePools(ctx context.Context, k8sClient client.Client) ([]runv1alpha1.NodePool, error) {
	var all []runv1alpha1.NodePool
	listOpts := []client.ListOption{client.Limit(ListPageSize)}
	for {
		var nodePools runv1alpha1.NodePoolList
		if err := k8sClient.List(ctx, &nodePools, listOpts...); err != nil {
			return nil, fmt.Errorf("list legacy nodepools: %w", err)
		}
		all = append(all, nodePools.Items...)
		if nodePools.Continue == "" {
			break
		}
		listOpts = []client.ListOption{client.Limit(ListPageSize), client.Continue(nodePools.Continue)}
	}
	return all, nil
}

// GetLegacyManagedNodesConfig returns the legacy (run.ai/v1alpha1) managed-nodes-config,
// reporting false when it does not exist. Unlike the other migrated resources it is a
// cluster-scoped singleton with a fixed name, so it is fetched by name rather than listed.
func GetLegacyManagedNodesConfig(ctx context.Context, k8sClient client.Client) (*runv1alpha1.ManagedNodesConfig, bool, error) {
	mnc := &runv1alpha1.ManagedNodesConfig{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: constants.ManagedNodesConfigName}, mnc)
	if apierrors.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get legacy managed nodes config %q: %w", constants.ManagedNodesConfigName, err)
	}
	return mnc, true, nil
}

// OldDepartmentCRDInstalled reports whether the legacy department CRD exists on
// the cluster. The department migration must be skipped when it does not (fresh
// installs have nothing to migrate).
func OldDepartmentCRDInstalled(ctx context.Context, k8sClient client.Client) (bool, error) {
	return crdInstalled(ctx, k8sClient, OldDepartmentCRDName)
}

// OldProjectCRDInstalled reports whether the legacy project CRD exists on the
// cluster. The project migration must be skipped when it does not (fresh installs
// have nothing to migrate).
func OldProjectCRDInstalled(ctx context.Context, k8sClient client.Client) (bool, error) {
	return crdInstalled(ctx, k8sClient, OldProjectCRDName)
}

// OldNodePoolCRDInstalled reports whether the legacy nodepool CRD exists on the
// cluster. The nodepool migration must be skipped when it does not (fresh
// installs have nothing to migrate).
func OldNodePoolCRDInstalled(ctx context.Context, k8sClient client.Client) (bool, error) {
	return crdInstalled(ctx, k8sClient, OldNodePoolCRDName)
}

// OldManagedNodesConfigCRDInstalled reports whether the legacy managed-nodes-config
// CRD exists on the cluster. The migration must be skipped when it does not (fresh
// installs have nothing to migrate).
func OldManagedNodesConfigCRDInstalled(ctx context.Context, k8sClient client.Client) (bool, error) {
	return crdInstalled(ctx, k8sClient, OldManagedNodesConfigCRDName)
}

// crdInstalled reports whether the named CRD exists on the cluster.
func crdInstalled(ctx context.Context, k8sClient client.Client, name string) (bool, error) {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, crd)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get CRD %q: %w", name, err)
	}
	return true, nil
}
