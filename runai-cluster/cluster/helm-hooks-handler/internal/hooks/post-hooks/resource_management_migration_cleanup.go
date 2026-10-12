/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	dynamicupgrades "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/dynamic-schema-upgrades"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// newDepartmentCRDName is the kai department CRD (kai.resources/v1alpha1) the
	// legacy departments were migrated to. The cleanup refuses to delete anything
	// unless this one is present, so departments are never left without a CRD.
	newDepartmentCRDName = "departments.kai.resources"

	// newProjectCRDName is the kai project CRD (kai.resources/v1alpha1) the legacy
	// projects were migrated to. The cleanup refuses to delete the legacy CRD unless
	// this one is present, so projects are never left without a CRD.
	newProjectCRDName = "projects.kai.resources"

	// newNodePoolCRDName is the kai nodepool CRD (kai.resources/v1alpha1) the legacy
	// nodepools were migrated to. The cleanup refuses to delete the legacy CRD unless
	// this one is present, so nodepools are never left without a CRD.
	newNodePoolCRDName = "nodepools.kai.resources"

	// newManagedNodesConfigCRDName is the kai managed-nodes-config CRD
	// (kai.resources/v1alpha1) the legacy one was migrated to. The cleanup refuses to
	// delete the legacy CRD unless this one is present.
	newManagedNodesConfigCRDName = "managednodesconfigs.kai.resources"
)

// newDepartmentGVK / newProjectGVK are the migrated kai types (kai.resources/
// v1alpha1). The post-upgrade client's scheme does not register the kai.resources
// types, so twins are looked up as unstructured objects via the REST mapper.
var (
	newDepartmentGVK         = schema.GroupVersionKind{Group: "kai.resources", Version: "v1alpha1", Kind: "Department"}
	newProjectGVK            = schema.GroupVersionKind{Group: "kai.resources", Version: "v1alpha1", Kind: "Project"}
	newNodePoolGVK           = schema.GroupVersionKind{Group: "kai.resources", Version: "v1alpha1", Kind: "NodePool"}
	newManagedNodesConfigGVK = schema.GroupVersionKind{Group: "kai.resources", Version: "v1alpha1", Kind: "ManagedNodesConfig"}
)

// cleanupLegacyDepartments removes the legacy departments and their CRD once the
// resource-management migration has produced a kai twin for every legacy
// department. It is gated by the resource-management-migration-cleanup flag and
// self-gating on the legacy CRD still being installed, so it runs once and no-ops
// thereafter.
//
// Guard rails, in order:
//   - legacy department CRD not installed -> nothing to clean up (fresh installs
//     and clusters already cleaned up).
//   - new (kai) department CRD not installed -> error; the migration target is
//     missing, deleting the legacy departments would lose them.
//   - a legacy department without a kai twin (same name) -> error; that
//     department was not migrated, deleting it would lose it.
//   - otherwise -> delete every legacy department, then the legacy CRD.
func (p *PostUpgradeHook) cleanupLegacyDepartments(ctx context.Context, k8sClient client.Client, logger logr.Logger) error {
	oldCRDInstalled, err := hooks.OldDepartmentCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !oldCRDInstalled {
		logger.Info("Legacy department CRD not installed, nothing to clean up")
		return nil
	}

	newCRDInstalled, err := crdInstalled(ctx, k8sClient, newDepartmentCRDName)
	if err != nil {
		return err
	}
	if !newCRDInstalled {
		return fmt.Errorf("refusing to delete legacy departments: new department CRD %q is not installed", newDepartmentCRDName)
	}

	oldDepartments, err := hooks.ListLegacyDepartments(ctx, k8sClient)
	if err != nil {
		return err
	}

	// Verify every legacy department has a migrated kai twin (same name) before
	// deleting anything, so a department that was not migrated is never lost.
	for i := range oldDepartments {
		name := oldDepartments[i].Name
		migrated, err := newDepartmentExists(ctx, k8sClient, name)
		if err != nil {
			return err
		}
		if !migrated {
			return fmt.Errorf("refusing to delete legacy departments: legacy department %q has no migrated kai department, migration is not complete", name)
		}
	}

	// Every legacy department has a kai twin: clear the finalizers, delete the
	// legacy departments, then the legacy CRD.
	for i := range oldDepartments {
		old := &oldDepartments[i]
		if err := clearLegacyFinalizers(ctx, k8sClient, old, "department", logger); err != nil {
			return err
		}
		if err := deleteLegacyObject(ctx, k8sClient, old, "department", logger); err != nil {
			return err
		}
	}

	return deleteCRD(ctx, k8sClient, hooks.OldDepartmentCRDName)
}

// cleanupLegacyProjects removes the legacy projects and their CRD once the
// resource-management migration has produced a kai twin for every legacy project.
// It mirrors cleanupLegacyDepartments.
//
// Guard rails, in order:
//   - legacy project CRD not installed -> nothing to clean up.
//   - new (kai) project CRD not installed -> error; the migration target is
//     missing, deleting the legacy projects would lose them.
//   - a legacy project without a kai twin (same name) -> error; that project was
//     not migrated, deleting it would lose it.
//   - otherwise -> delete every legacy project, then the legacy CRD.
func (p *PostUpgradeHook) cleanupLegacyProjects(ctx context.Context, k8sClient client.Client) error {
	logger := log.FromContext(ctx)
	oldCRDInstalled, err := hooks.OldProjectCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !oldCRDInstalled {
		logger.Info("Legacy project CRD not installed, nothing to clean up")
		return nil
	}

	newCRDInstalled, err := crdInstalled(ctx, k8sClient, newProjectCRDName)
	if err != nil {
		return err
	}
	if !newCRDInstalled {
		return fmt.Errorf("refusing to delete legacy projects: new project CRD %q is not installed", newProjectCRDName)
	}

	oldProjects, err := hooks.ListLegacyProjects(ctx, k8sClient)
	if err != nil {
		return err
	}

	// Verify every legacy project has a migrated kai twin (same name) before
	// deleting anything, so a project that was not migrated is never lost.
	for i := range oldProjects {
		name := oldProjects[i].Name
		migrated, err := newProjectExists(ctx, k8sClient, name)
		if err != nil {
			return err
		}
		if !migrated {
			return fmt.Errorf("refusing to delete legacy projects: legacy project %q has no migrated kai project, migration is not complete", name)
		}
	}

	// Every legacy project has a kai twin: clear the finalizers, delete the legacy
	// projects, then the legacy CRD.
	for i := range oldProjects {
		old := &oldProjects[i]
		if err := clearLegacyFinalizers(ctx, k8sClient, old, "project", logger); err != nil {
			return err
		}
		if err := deleteLegacyObject(ctx, k8sClient, old, "project", logger); err != nil {
			return err
		}
	}

	return deleteCRD(ctx, k8sClient, hooks.OldProjectCRDName)
}

// cleanupLegacyNodePools removes the legacy nodepools and their CRD once every one
// has a kai twin, mirroring cleanupLegacyProjects. It no-ops if the legacy CRD is
// gone, and errors (deleting nothing) if the kai CRD is missing or any nodepool
// lacks a twin, so a nodepool is never lost.
func (p *PostUpgradeHook) cleanupLegacyNodePools(ctx context.Context, k8sClient client.Client) error {
	logger := log.FromContext(ctx)
	oldCRDInstalled, err := hooks.OldNodePoolCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !oldCRDInstalled {
		logger.Info("Legacy nodepool CRD not installed, nothing to clean up")
		return nil
	}

	newCRDInstalled, err := crdInstalled(ctx, k8sClient, newNodePoolCRDName)
	if err != nil {
		return err
	}
	if !newCRDInstalled {
		return fmt.Errorf("refusing to delete legacy nodepools: new nodepool CRD %q is not installed", newNodePoolCRDName)
	}

	oldNodePools, err := hooks.ListLegacyNodePools(ctx, k8sClient)
	if err != nil {
		return err
	}

	// Verify every legacy nodepool has a kai twin before deleting anything.
	for i := range oldNodePools {
		name := oldNodePools[i].Name
		migrated, err := newNodePoolExists(ctx, k8sClient, name)
		if err != nil {
			return err
		}
		if !migrated {
			return fmt.Errorf("refusing to delete legacy nodepools: legacy nodepool %q has no migrated kai nodepool, migration is not complete", name)
		}
	}

	// All twins exist: delete the legacy nodepools, then the legacy CRD.
	for i := range oldNodePools {
		old := &oldNodePools[i]
		logger.Info("Deleting legacy nodepool", "name", old.Name)
		if err = k8sClient.Delete(ctx, old); err != nil && !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
			return fmt.Errorf("failed to delete legacy nodepool %q: %w", old.Name, err)
		}
	}

	return deleteCRD(ctx, k8sClient, hooks.OldNodePoolCRDName)
}

// cleanupLegacyManagedNodesConfigs removes the legacy managed-nodes-config and its CRD
// once the kai twin exists, mirroring cleanupLegacyNodePools. Because it is a singleton
// the twin check is a single lookup rather than a loop, and a legacy CRD with no object
// still gets its CRD deleted.
func (p *PostUpgradeHook) cleanupLegacyManagedNodesConfigs(ctx context.Context, k8sClient client.Client) error {
	logger := log.FromContext(ctx)
	oldCRDInstalled, err := hooks.OldManagedNodesConfigCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !oldCRDInstalled {
		logger.Info("Legacy managed-nodes-config CRD not installed, nothing to clean up")
		return nil
	}

	newCRDInstalled, err := crdInstalled(ctx, k8sClient, newManagedNodesConfigCRDName)
	if err != nil {
		return err
	}
	if !newCRDInstalled {
		return fmt.Errorf("refusing to delete legacy managed-nodes-config: new managed-nodes-config CRD %q is not installed", newManagedNodesConfigCRDName)
	}

	old, found, err := hooks.GetLegacyManagedNodesConfig(ctx, k8sClient)
	if err != nil {
		return err
	}

	if found {
		var migrated bool
		migrated, err = newManagedNodesConfigExists(ctx, k8sClient, old.Name)
		if err != nil {
			return err
		}
		if !migrated {
			return fmt.Errorf("refusing to delete legacy managed-nodes-config: %q has no migrated kai managed-nodes-config, migration is not complete", old.Name)
		}

		logger.Info("Deleting legacy managed-nodes-config", "name", old.Name)
		if err = k8sClient.Delete(ctx, old); err != nil && !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
			return fmt.Errorf("failed to delete legacy managed-nodes-config %q: %w", old.Name, err)
		}
	}

	return deleteCRD(ctx, k8sClient, hooks.OldManagedNodesConfigCRDName)
}

// finalizeResourceManagementMigration completes the migration in post-upgrade: it
// re-enables the operands the pre-upgrade hook disabled (so they come back on the
// new version), then, once they are ready, deletes the migrated legacy resources
// and their CRDs. The org-unit and nodepool cleanups are gated independently.
// Operands are re-enabled first so they recover even if the cleanup fails.
func (p *PostUpgradeHook) finalizeResourceManagementMigration(ctx context.Context, resources *clients.Clients) error {
	logger := log.FromContext(ctx)

	// The org-unit (department/project) and nodepool cleanups are gated
	// independently, so each can be rolled out and tested on its own.
	orgUnitsMigrationEnabled := p.params.ResourceManagementMigrationEnabled
	orgUnitsCleanup := p.params.ResourceManagementMigrationCleanup
	nodePoolsMigrationEnabled := p.params.ResourceManagementMigrationNodePoolsEnabled
	nodePoolsCleanup := p.params.ResourceManagementMigrationNodePoolsCleanup

	if !orgUnitsMigrationEnabled && !nodePoolsMigrationEnabled {
		logger.Info("Skipping post-resource-management-upgrade hook: resource-management migration disabled")
		return nil
	}

	// Nodepool cleanup runs before the operands are restored: the pre-upgrade hook
	// stripped the legacy nodepools' and managed-nodes-config's finalizers, so a
	// restarted controller would only re-add them.
	if nodePoolsCleanup {
		if err := p.cleanupLegacyNodePools(ctx, resources.LocalClusterClient); err != nil {
			return fmt.Errorf("failed to clean up legacy nodepools: %w", err)
		}
		if err := p.cleanupLegacyManagedNodesConfigs(ctx, resources.LocalClusterClient); err != nil {
			return fmt.Errorf("failed to clean up legacy managed-nodes-config: %w", err)
		}
	}

	// Re-enable the operands each active migration flow disabled, so they come back
	// up on the new version - regardless of whether cleanup runs.
	if err := p.restoreResourceManagementOperands(ctx, resources, logger, orgUnitsMigrationEnabled, nodePoolsMigrationEnabled); err != nil {
		return fmt.Errorf("failed to restore resource-management operands: %w", err)
	}

	if orgUnitsCleanup {
		// The agent and project-controller must be back up before the cleanup deletes
		// the legacy org units: the agent serves the validate-project webhook that
		// guards the delete, and the project-controller clears the departments' /
		// projects' finalizers so the deletes actually complete.
		for _, name := range []string{hooks.AgentOperandName, hooks.ProjectControllerDeploymentName} {
			if err := waitForDeploymentReady(ctx, resources.LocalClusterClient, name, p.params.OperandDownIntervalMillis, logger); err != nil {
				return fmt.Errorf("failed waiting for %s before cleanup: %w", name, err)
			}
		}
		if err := p.cleanupLegacyDepartments(ctx, resources.LocalClusterClient, logger); err != nil {
			return fmt.Errorf("failed to clean up legacy departments: %w", err)
		}
		if err := p.cleanupLegacyProjects(ctx, resources.LocalClusterClient); err != nil {
			return fmt.Errorf("failed to clean up legacy projects: %w", err)
		}
	}

	return nil
}

// restoreResourceManagementOperands re-enables the operands the pre-upgrade hook
// disabled - project-controller for the org-unit flow, nodepool-controller for the
// nodepool flow, and the agent whenever either ran. It runs in post-upgrade so they
// return on the new version, attempts all even if one fails, and re-enabling an
// already-enabled operand is a no-op.
func (p *PostUpgradeHook) restoreResourceManagementOperands(ctx context.Context, resources *clients.Clients, logger logr.Logger,
	orgUnits, nodePools bool) error {
	var errs []error

	if orgUnits {
		logger.Info("Re-enabling project-controller after resource-management migration")
		if err := dynamicupgrades.NewRunaiConfigEnableProjectController().Upgrade(ctx, resources); err != nil {
			errs = append(errs, fmt.Errorf("failed to re-enable project-controller: %w", err))
		}
	}
	if orgUnits || nodePools {
		logger.Info("Re-enabling agent after resource-management migration")
		if err := dynamicupgrades.NewRunaiConfigEnableAgent().Upgrade(ctx, resources); err != nil {
			errs = append(errs, fmt.Errorf("failed to re-enable agent: %w", err))
		}
	}
	if nodePools {
		logger.Info("Re-enabling nodepool-controller after resource-management migration")
		if err := dynamicupgrades.NewRunaiConfigEnableNodepoolController().Upgrade(ctx, resources); err != nil {
			errs = append(errs, fmt.Errorf("failed to re-enable nodepool-controller: %w", err))
		}
	}
	return errors.Join(errs...)
}

// waitForDeploymentReady blocks until the named Deployment (in the runai namespace)
// reports at least one available replica. go-operator scales the operands back up
// asynchronously after restoreResourceManagementOperands re-enables them, so this
// closes the race between the re-enable and the operand actually serving.
func waitForDeploymentReady(ctx context.Context, k8sClient client.Client, name string, intervalMillis int, logger logr.Logger) error {
	logger.Info("Waiting for deployment to be ready before cleanup", "deployment", name)
	ticker := time.NewTicker(time.Duration(intervalMillis) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ready, err := deploymentReady(ctx, k8sClient, name)
			if err != nil {
				return err
			}
			if ready {
				logger.Info("Deployment is ready", "deployment", name)
				return nil
			}
			logger.Info("Deployment not ready yet", "deployment", name)
		case <-ctx.Done():
			return fmt.Errorf("timeout reached while waiting for deployment %q to be ready", name)
		}
	}
}

// deploymentReady reports whether the named Deployment has at least one available
// replica. A missing Deployment counts as not-ready (go-operator has not recreated
// it yet), not an error.
func deploymentReady(ctx context.Context, k8sClient client.Client, name string) (bool, error) {
	dep := &appsv1.Deployment{}
	err := k8sClient.Get(ctx, types.NamespacedName{Namespace: runaiv1.DefaultNamespace, Name: name}, dep)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get deployment %q: %w", name, err)
	}
	return dep.Status.AvailableReplicas >= 1, nil
}

// objectExists reports whether the named object exists (false on NotFound). obj is
// an empty object of the wanted type - typed (e.g. a CRD) or an unstructured with
// its GVK set.
func objectExists(ctx context.Context, k8sClient client.Client, name, kindStr string, obj client.Object) (bool, error) {
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name}, obj)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get %s %q: %w", kindStr, name, err)
	}
	return true, nil
}

func crdInstalled(ctx context.Context, k8sClient client.Client, name string) (bool, error) {
	return objectExists(ctx, k8sClient, name, "CRD", &apiextensionsv1.CustomResourceDefinition{})
}

// clearLegacyFinalizers strips a legacy org unit's finalizers so its deletion can
// complete. The project-controller only reconciles the kai.resources types, so
// nothing removes the legacy run.ai finalizers once it is on the new version:
// without this the deletes below are accepted but never finish, the objects sit
// Terminating forever, and the legacy CRD stays stuck on its own
// customresourcecleanup finalizer waiting for instances that can never go away.
//
// It is a no-op for an object with no finalizers, and tolerates an object (or CRD)
// that is already gone, so re-running the hook on a cluster wedged by an earlier
// upgrade clears it.
func clearLegacyFinalizers(ctx context.Context, k8sClient client.Client, obj client.Object,
	kindStr string, logger logr.Logger) error {
	if len(obj.GetFinalizers()) == 0 {
		return nil
	}

	logger.Info("Clearing finalizers on legacy "+kindStr, "name", obj.GetName(), "finalizers", obj.GetFinalizers())
	patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
	obj.SetFinalizers(nil)
	return retryLegacyOperation(ctx, obj, kindStr, "clear finalizers on", logger, func(ctx context.Context) error {
		return k8sClient.Patch(ctx, obj, patch)
	})
}

// deleteCRD deletes the named CRD, tolerating a CRD that is already gone or whose
// type is not registered on the cluster.
func deleteCRD(ctx context.Context, k8sClient client.Client, name string) error {
	log.FromContext(ctx).Info("Deleting legacy CRD", "name", name)
	crd := &apiextensionsv1.CustomResourceDefinition{}
	crd.SetName(name)
	if err := k8sClient.Delete(ctx, crd); err != nil && !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
		return fmt.Errorf("failed to delete legacy CRD %q: %w", name, err)
	}
	return nil
}

// newKaiResourceExists reports whether a migrated kai.resources object of the given
// GVK and name exists. It uses an unstructured lookup because the post-upgrade
// client's scheme does not register the kai.resources types.
func newKaiResourceExists(ctx context.Context, k8sClient client.Client, name, kindStr string, gvk schema.GroupVersionKind) (bool, error) {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	return objectExists(ctx, k8sClient, name, kindStr, obj)
}

func newDepartmentExists(ctx context.Context, k8sClient client.Client, name string) (bool, error) {
	return newKaiResourceExists(ctx, k8sClient, name, "kai department", newDepartmentGVK)
}

func newProjectExists(ctx context.Context, k8sClient client.Client, name string) (bool, error) {
	return newKaiResourceExists(ctx, k8sClient, name, "kai project", newProjectGVK)
}

func newNodePoolExists(ctx context.Context, k8sClient client.Client, name string) (bool, error) {
	return newKaiResourceExists(ctx, k8sClient, name, "kai nodepool", newNodePoolGVK)
}

func newManagedNodesConfigExists(ctx context.Context, k8sClient client.Client, name string) (bool, error) {
	return newKaiResourceExists(ctx, k8sClient, name, "kai managed-nodes-config", newManagedNodesConfigGVK)
}
