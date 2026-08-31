/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	dynamicupgrades "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/dynamic-schema-upgrades"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	schedv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// departmentKind is the Kind used in the OwnerReference stamped on owned
	// objects re-parented to the new kai department.
	departmentKind = "Department"

	// departmentQueueNamePrefix is the prefix the agent puts before the department
	// id when naming a kai department's queues (q-<department-id>[-<nodepool>]).
	departmentQueueNamePrefix = "q-"

	// operandRestoreTimeout bounds the best-effort attempt to bring the
	// project-controller (and, the same way, the agent) back up. It uses a detached
	// context so the operand is restored even when the migration's own context has
	// already been cancelled.
	operandRestoreTimeout = 30 * time.Second
)

// migrationTrueRef is the addressable bool the OwnerReference Controller /
// BlockOwnerDeletion pointers reference, matching the project-controller's
// queue ownership stamp.
var migrationTrueRef = true

// reownedQueue records a queue whose owner was changed during migration, so the
// change can be undone if a later step in the same department's migration fails.
type reownedQueue struct {
	queue    schedv2.Queue
	oldOwner []metav1.OwnerReference
}

type PreResourceManagementUpgradeHook struct {
	params *hooks.ClusterHookParams
}

func NewPreResourceManagementUpgradeHook(params *hooks.ClusterHookParams) *PreResourceManagementUpgradeHook {
	return &PreResourceManagementUpgradeHook{params: params}
}

func (p *PreResourceManagementUpgradeHook) GetIdentifier() string {
	return hooks.PreResourceManagementUpgradeHookIdentifier
}

func (p *PreResourceManagementUpgradeHook) Call(ctx context.Context, resources *clients.Clients) error {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("Starting %s", p.GetIdentifier()))
	defer logger.Info(fmt.Sprintf("Done %s", p.GetIdentifier()))

	orgUnitsEnabled := p.params.ResourceManagementMigrationEnabled
	nodePoolsEnabled := p.params.ResourceManagementMigrationNodePoolsEnabled
	if !orgUnitsEnabled && !nodePoolsEnabled {
		logger.Info("Skipping pre-resource-management-upgrade hook: resource-management migration disabled")
		return nil
	}

	return p.migrateResourceManagementResources(ctx, resources, resources.LocalClusterClient, orgUnitsEnabled, nodePoolsEnabled)
}

// migrateResourceManagementResources runs each enabled migration flow. The org-unit
// (department + project) and nodepool flows are independent - each disables the
// operands it needs, migrates, and re-enables them on failure - so they live in
// their own functions and don't share state here.
func (p *PreResourceManagementUpgradeHook) migrateResourceManagementResources(ctx context.Context, resources *clients.Clients,
	k8sClient client.Client, orgUnitsEnabled, nodePoolsEnabled bool) error {
	if orgUnitsEnabled {
		if err := p.migrateOrgUnits(ctx, resources, k8sClient); err != nil {
			return err
		}
	}
	if nodePoolsEnabled {
		if err := p.migrateNodePoolResources(ctx, resources, k8sClient); err != nil {
			return err
		}
	}
	return nil
}

// migrateOrgUnits disables the project-controller and agent, migrates the legacy
// departments and projects to their kai twins, and re-enables the operands on
// failure (on success the post-upgrade hook does, so they return on the new
// version). It is a no-op when neither legacy CRD is installed. Operands are toggled
// via RunaiConfig, not by scaling deployments (which go-operator would revert).
func (p *PreResourceManagementUpgradeHook) migrateOrgUnits(ctx context.Context, resources *clients.Clients, k8sClient client.Client) (err error) {
	logger := log.FromContext(ctx)

	oldDepartmentInstalled, err := hooks.OldDepartmentCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	oldProjectInstalled, err := hooks.OldProjectCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !oldDepartmentInstalled && !oldProjectInstalled {
		logger.Info("Neither legacy department nor project CRD installed, nothing to migrate")
		return nil
	}

	// On success the project-controller and agent are restarted in post-hook.
	// On failure (or panic) they are re-enabled here. Registered before the disables so
	// a failure between them still restores the one already taken down; re-enabling an
	// operand that was never disabled is a no-op.
	defer func() {
		if err != nil {
			reEnableOperands(resources, logger,
				operandRestore{"project-controller", dynamicupgrades.NewRunaiConfigEnableProjectController().Upgrade},
				operandRestore{"agent", dynamicupgrades.NewRunaiConfigEnableAgent().Upgrade})
		}
	}()

	logger.Info("Running resource-management migration - disabling project-controller")
	if err = dynamicupgrades.NewRunaiConfigDisableProjectController().Upgrade(ctx, resources); err != nil {
		return fmt.Errorf("failed to disable project-controller: %w", err)
	}
	logger.Info("Running departments migration - disabling agent")
	if err = dynamicupgrades.NewRunaiConfigDisableAgent().Upgrade(ctx, resources); err != nil {
		return fmt.Errorf("failed to disable agent: %w", err)
	}

	if err = p.waitForOperandsDown(ctx, resources, hooks.ProjectControllerOperandName, hooks.AgentOperandName); err != nil {
		return err
	}

	// Departments first - a project's parent department must already have a kai twin.
	if oldDepartmentInstalled {
		logger.Info("Old department CRD installed, running department migration")
		if err = migrateDepartments(ctx, k8sClient); err != nil {
			return err
		}
	}
	if oldProjectInstalled {
		logger.Info("Old project CRD installed, running project migration")
		if err = migrateProjects(ctx, k8sClient); err != nil {
			return err
		}
	}
	return nil
}

// migrateNodePoolResources disables the nodepool-controller and agent, migrates the
// legacy nodepools and managed-nodes-config to their kai twins, and re-enables the
// operands on failure. Both are owned by the same two operands, so they share one
// disable/restore window. It is a no-op when neither legacy CRD is installed.
func (p *PreResourceManagementUpgradeHook) migrateNodePoolResources(ctx context.Context, resources *clients.Clients, k8sClient client.Client) (err error) {
	logger := log.FromContext(ctx)

	oldNodePoolInstalled, err := hooks.OldNodePoolCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	oldManagedNodesConfigInstalled, err := hooks.OldManagedNodesConfigCRDInstalled(ctx, k8sClient)
	if err != nil {
		return err
	}
	if !oldNodePoolInstalled && !oldManagedNodesConfigInstalled {
		logger.Info("Neither legacy nodepool nor managed-nodes-config CRD installed, nothing to migrate")
		return nil
	}

	// Registered before the disables so a failure between them still restores the one
	// already taken down; re-enabling an operand that was never disabled is a no-op.
	defer func() {
		if err != nil {
			reEnableOperands(resources, logger,
				operandRestore{"nodepool-controller", dynamicupgrades.NewRunaiConfigEnableNodepoolController().Upgrade},
				operandRestore{"agent", dynamicupgrades.NewRunaiConfigEnableAgent().Upgrade})
		}
	}()

	logger.Info("Running nodepool migration - disabling nodepool-controller and agent")
	if err = dynamicupgrades.NewRunaiConfigDisableNodepoolController().Upgrade(ctx, resources); err != nil {
		return fmt.Errorf("failed to disable nodepool-controller: %w", err)
	}
	if err = dynamicupgrades.NewRunaiConfigDisableAgent().Upgrade(ctx, resources); err != nil {
		return fmt.Errorf("failed to disable agent: %w", err)
	}

	if err = p.waitForOperandsDown(ctx, resources, hooks.NodePoolControllerOperandName, hooks.AgentOperandName); err != nil {
		return err
	}

	if oldNodePoolInstalled {
		logger.Info("Running nodepool migration")
		if err = migrateNodePools(ctx, k8sClient); err != nil {
			return err
		}
	}
	if oldManagedNodesConfigInstalled {
		logger.Info("Running managed-nodes-config migration")
		if err = migrateManagedNodesConfig(ctx, k8sClient); err != nil {
			return err
		}
	}
	return nil
}

// waitForOperandsDown waits for each named operand's pods to be gone. Disabling an
// operand in RunaiConfig is asynchronous - go-operator has to observe it and scale
// the deployment down - so this must complete before mutating the resources the
// operand manages.
func (p *PreResourceManagementUpgradeHook) waitForOperandsDown(ctx context.Context, resources *clients.Clients, operands ...string) error {
	for _, operand := range operands {
		if err := waitForOperandsToBeDown(ctx, resources.LocalClusterClient, operand, p.params.OperandDownIntervalMillis); err != nil {
			return err
		}
	}
	return nil
}

// operandRestore names an operand and the RunaiConfig upgrade that re-enables it.
type operandRestore struct {
	name  string
	apply func(context.Context, *clients.Clients) error
}

// reEnableOperands re-enables the given operands on a detached context (so they are
// restored even if the migration's own context was already cancelled), logging any
// failure rather than returning it so one stuck operand doesn't block the others.
func reEnableOperands(resources *clients.Clients, logger logr.Logger, restores ...operandRestore) {
	restoreCtx, cancel := context.WithTimeout(context.Background(), operandRestoreTimeout)
	defer cancel()
	for _, r := range restores {
		if err := r.apply(restoreCtx, resources); err != nil {
			logger.Error(err, fmt.Sprintf("failed to re-enable %s after migration failure", r.name))
		}
	}
}

// migrateDepartments lists every old department (scheduling.run.ai/v2alpha2) and
// migrates each to a new kai department (kai.resources/v1alpha1), re-owning the
// queues the old department owns.
//
// It is idempotent: a department that already has a kai twin (same name) is
// skipped. Each department is migrated as a single chain - if re-owning any of
// its queues fails, the queue owners already changed are rolled back and the
// just-created new department is deleted before the error is returned.
func migrateDepartments(ctx context.Context, k8sClient client.Client) error {
	logger := log.FromContext(ctx)
	created := 0
	skipped := 0

	oldDepartments, err := hooks.ListLegacyDepartments(ctx, k8sClient)
	if err != nil {
		return err
	}

	for i := range oldDepartments {
		old := &oldDepartments[i]

		var existing kaiv1alpha1.Department
		err := k8sClient.Get(ctx, types.NamespacedName{Name: old.Name}, &existing)
		if err == nil {
			skipped++
			logger.Info(fmt.Sprintf("kai department %s already exists, skipping", old.Name))
			continue
		}
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get kai department %s: %w", old.Name, err)
		}

		if err := migrateDepartment(ctx, k8sClient, old); err != nil {
			return fmt.Errorf("migrate department %s: %w", old.Name, err)
		}
		created++
		logger.Info(fmt.Sprintf("migrated department %s", old.Name))
	}

	logger.Info(fmt.Sprintf("department migration complete: old=%d created=%d skipped=%d", len(oldDepartments), created, skipped))
	return nil
}

// migrateDepartment migrates a single old department: it creates the new kai
// department, then re-owns the queues the old department owns. On any failure
// after the new department is created, it rolls back the whole chain.
func migrateDepartment(ctx context.Context, k8sClient client.Client, old *schedv2alpha2.Department) error {
	newDept, err := buildNewDepartment(old)
	if err != nil {
		return err
	}

	if err = k8sClient.Create(ctx, newDept); err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create kai department: %w", err)
	}
	err = reownQueues(ctx, k8sClient, old, newDept)
	if err != nil {
		rollbackDepartmentUpgrade(ctx, k8sClient, newDept, nil)
		return err
	}

	return nil
}

// buildNewDepartment maps an old department onto a new kai department per the
// design: same name; copy annotations / finalizers / labels; move the status
// sync timestamp to a label; and convert the queues spec via a JSON round-trip
// (the resource shapes are structurally identical).
func buildNewDepartment(old *schedv2alpha2.Department) (*kaiv1alpha1.Department, error) {
	newDept := &kaiv1alpha1.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name:        old.Name,
			Annotations: copyStringMap(old.Annotations),
			Labels:      copyStringMap(old.Labels),
			Finalizers:  copyStringSlice(old.Finalizers),
		},
	}

	if ts := old.Status.SyncStatus.LastSyncTimestamp; !ts.IsZero() {
		if newDept.Annotations == nil {
			newDept.Annotations = map[string]string{}
		}
		newDept.Annotations[hooks.LastSyncTimestampAnnotation] = ts.UTC().Format("2006-01-02T15-04-05Z")
	}

	queues, err := convertQueues(old.Spec.Queues, departmentQueueBaseName(old))
	if err != nil {
		return nil, fmt.Errorf("convert queues for department %s: %w", old.Name, err)
	}
	newDept.Spec.Queues = queues

	return newDept, nil
}

// departmentQueueBaseName returns the base the agent builds a kai department's
// queue names from: "q-<department-id>", the id taken from the label the agent
// stamps on the department. It returns "" when the department carries no id (a
// department not created by the agent), in which case the queue names are kept
// as they are.
func departmentQueueBaseName(old *schedv2alpha2.Department) string {
	id := old.Labels[schedv2alpha2.DepartmentIdLabel]
	if id == "" {
		return ""
	}
	return departmentQueueNamePrefix + id
}

// convertQueues converts the old department queue specs to the new QueueConfig
// shape via a JSON round-trip. Only the fields shared by both shapes
// (name/nodepool/resources/priority) survive; the rest are dropped per design.
//
// The old department's queues are named after the department name, while the
// agent names a kai department's queues after the department id, so the migrated
// queues are renamed to the id-based scheme the agent would write - otherwise the
// first agent sync after the upgrade would rename them anyway.
func convertQueues(old []schedv2alpha2.QueueDepartmentSpec, queueBaseName string) ([]kaiv1alpha1.QueueConfig, error) {
	if old == nil {
		return nil, nil
	}
	data, err := json.Marshal(old)
	if err != nil {
		return nil, fmt.Errorf("marshal old queues: %w", err)
	}
	var queues []kaiv1alpha1.QueueConfig
	if err := json.Unmarshal(data, &queues); err != nil {
		return nil, fmt.Errorf("unmarshal queues into kai shape: %w", err)
	}
	if queueBaseName == "" {
		return queues, nil
	}
	for i := range queues {
		queues[i].Name = departmentQueueName(queueBaseName, queues[i].Nodepool)
	}
	return queues, nil
}

// departmentQueueName mirrors the agent's queue naming (handlers.GetQueueName):
// the queue of the default nodepool is named after the base alone, every other
// queue is suffixed with its nodepool.
func departmentQueueName(baseName, nodepool string) string {
	if nodepool == runaiv1alpha1.DefaultNodePoolName {
		return baseName
	}
	return fmt.Sprintf("%s-%s", baseName, nodepool)
}

// reownQueues re-points every queue owned by the old department at the new kai
// department. Ownership is determined by the queue's OwnerReference (a Department
// owner named after the old department), not by any label. It tracks each
// successful change and, on the first failure, rolls back all changes already
// made and deletes the new department before returning.
func reownQueues(ctx context.Context, k8sClient client.Client, old *schedv2alpha2.Department, newDept *kaiv1alpha1.Department) error {
	newOwnerRef := metav1.OwnerReference{
		APIVersion:         kaiv1alpha1.GroupVersion.String(),
		Kind:               departmentKind,
		Name:               newDept.Name,
		UID:                newDept.UID,
		Controller:         &migrationTrueRef,
		BlockOwnerDeletion: &migrationTrueRef,
	}

	var reowned []reownedQueue
	listOpts := []client.ListOption{
		client.MatchingLabels{schedv2alpha2.DepartmentNameLabel: old.Name},
		client.Limit(hooks.ListPageSize),
	}
	for {
		var queues schedv2.QueueList
		if err := k8sClient.List(ctx, &queues, listOpts...); err != nil {
			rollbackDepartmentUpgrade(ctx, k8sClient, newDept, reowned)
			return fmt.Errorf("list queues for department %s: %w", old.Name, err)
		}

		for i := range queues.Items {
			queue := &queues.Items[i]
			oldOwner := queue.OwnerReferences

			updated, replaced := replaceDepartmentOwner(oldOwner, old, newOwnerRef)
			if !replaced {
				continue
			}

			patch := client.MergeFrom(queue.DeepCopy())
			queue.OwnerReferences = updated
			if err := k8sClient.Patch(ctx, queue, patch); err != nil {
				rollbackDepartmentUpgrade(ctx, k8sClient, newDept, reowned)
				return fmt.Errorf("patch owner of queue %s: %w", queue.Name, err)
			}
			reowned = append(reowned, reownedQueue{queue: *queue, oldOwner: oldOwner})
		}

		if queues.Continue == "" {
			break
		}
		listOpts = []client.ListOption{
			client.MatchingLabels{schedv2alpha2.DepartmentNameLabel: old.Name},
			client.Limit(hooks.ListPageSize),
			client.Continue(queues.Continue),
		}
	}

	return nil
}

// replaceDepartmentOwner returns a copy of owners with any OwnerReference that
// targets the old department (a Department owner with the old department's name)
// replaced by newOwnerRef, and reports whether such an owner was found. Other
// owners are kept untouched. A queue with no matching owner is left unchanged.
func replaceDepartmentOwner(owners []metav1.OwnerReference, old *schedv2alpha2.Department,
	newOwnerRef metav1.OwnerReference) ([]metav1.OwnerReference, bool) {
	result := make([]metav1.OwnerReference, 0, len(owners))
	replaced := false
	for _, owner := range owners {
		if owner.Kind == departmentKind && owner.Name == old.Name {
			result = append(result, newOwnerRef)
			replaced = true
			continue
		}
		result = append(result, owner)
	}
	return result, replaced
}

// rollbackDepartmentUpgrade undoes a failed department migration: it restores the owner of every
// queue already re-owned and deletes the new department.
func rollbackDepartmentUpgrade(ctx context.Context, k8sClient client.Client, newDept *kaiv1alpha1.Department,
	reowned []reownedQueue) {
	logger := log.FromContext(ctx)
	for i := range reowned {
		queue := reowned[i].queue
		patch := client.MergeFrom(queue.DeepCopy())
		queue.OwnerReferences = reowned[i].oldOwner
		if err := k8sClient.Patch(ctx, &queue, patch); err != nil {
			logger.Error(err, fmt.Sprintf("rollback: failed to restore queue %s owner", queue.Name))
		}
	}

	// Strip finalizers first: the new department's controller is not running
	// during the migration window, so a copied finalizer would otherwise leave
	// the department stuck terminating instead of actually deleted.
	if len(newDept.Finalizers) > 0 {
		newDept.Finalizers = nil
		if err := k8sClient.Update(ctx, newDept); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf("rollback: failed to strip new department %s finalizers", newDept.Name))
		}
	}

	if err := k8sClient.Delete(ctx, newDept); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("rollback: failed to delete new department %s", newDept.Name))
	}
}

func copyStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyStringSlice(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
