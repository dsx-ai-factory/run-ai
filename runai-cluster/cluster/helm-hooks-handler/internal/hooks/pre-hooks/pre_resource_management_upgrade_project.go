/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"strconv"

	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// projectKind is the Kind used in the OwnerReference stamped on owned objects
	// re-parented to the new kai project (unchanged between the old and new CRD).
	projectKind = "Project"

	// projectNameLabel is the label queues owned by a project carry, valued with the
	// project name (project-controller's config.ProjectLabelKey, default "project").
	// The migration lists a project's queues by it.
	projectNameLabel = "project"
)

// reownedObject records an object whose ownership was re-pointed during a project
// migration, capturing enough of its pre-change state (owner references and labels)
// to undo the change if a later step in the same project's migration fails.
type reownedObject struct {
	obj       client.Object
	oldOwners []metav1.OwnerReference
	oldLabels map[string]string
}

// migrateProjects lists every legacy project (run.ai/v2) and migrates each to a new
// kai project (kai.resources/v1alpha1), re-parenting every object the old project
// owns. It is idempotent: a project that already has a kai twin (same name) is
// skipped. Each project is migrated as a single chain - if re-parenting any owned
// object fails, every change already made for that project is rolled back and the
// just-created kai project is deleted before the error is returned.
func migrateProjects(ctx context.Context, k8sClient client.Client) error {
	logger := log.FromContext(ctx)
	created := 0
	skipped := 0

	oldProjects, err := hooks.ListLegacyProjects(ctx, k8sClient)
	if err != nil {
		return err
	}

	for i := range oldProjects {
		old := &oldProjects[i]

		var existing kaiv1alpha1.Project
		err := k8sClient.Get(ctx, types.NamespacedName{Name: old.Name}, &existing)
		if err == nil {
			skipped++
			logger.Info(fmt.Sprintf("kai project %s already exists, skipping", old.Name))
			continue
		}
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("get kai project %s: %w", old.Name, err)
		}

		if err := migrateProject(ctx, k8sClient, old); err != nil {
			return fmt.Errorf("migrate project %s: %w", old.Name, err)
		}
		created++
		logger.Info(fmt.Sprintf("migrated project %s", old.Name))
	}

	logger.Info(fmt.Sprintf("project migration complete: old=%d created=%d skipped=%d", len(oldProjects), created, skipped))
	return nil
}

// migrateProject migrates a single legacy project: it creates the new kai project,
// then re-parents every object the old project owns. On any failure after the new
// project is created, it rolls back the whole chain.
func migrateProject(ctx context.Context, k8sClient client.Client, old *runaiv2.Project) error {
	newProject, err := buildNewProject(old)
	if err != nil {
		return err
	}

	if err = k8sClient.Create(ctx, newProject); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("create kai project: %w", err)
		}
		// The kai project already exists (a concurrent migration or retry created it
		// after migrateProjects' existence check). Create leaves newProject without a
		// UID, so fetch the persisted object before reparenting - the ownership
		// reassignment below stamps newProject.UID onto every owned object and would
		// otherwise write an empty UID.
		if err = k8sClient.Get(ctx, types.NamespacedName{Name: newProject.Name}, newProject); err != nil {
			return fmt.Errorf("get existing kai project %s: %w", newProject.Name, err)
		}
	}

	var reowned []reownedObject
	if err = reownProjectOwnedObjects(ctx, k8sClient, old, newProject, &reowned); err != nil {
		rollbackProject(ctx, k8sClient, newProject, reowned)
		return err
	}

	return nil
}

// buildNewProject maps a legacy project onto a new kai project per the design: same
// name; copy annotations / finalizers / labels; rename department->parent and
// enforceRunaiScheduler->enforceKaiScheduler; convert queues via a JSON round-trip;
// map the deletion request; and move the dropped legacy spec fields (and the status
// sync timestamp) onto labels/annotations since the new CRD has no home for them.
func buildNewProject(old *runaiv2.Project) (*kaiv1alpha1.Project, error) {
	newProject := &kaiv1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name:        old.Name,
			Annotations: copyStringMap(old.Annotations),
			Labels:      copyStringMap(old.Labels),
			Finalizers:  copyStringSlice(old.Finalizers),
		},
		Spec: kaiv1alpha1.ProjectSpec{
			Parent:              old.Spec.Department,
			DefaultNodePools:    copyStringSlice(old.Spec.DefaultNodePools),
			Namespace:           old.Spec.Namespace,
			EnforceKaiScheduler: old.Spec.EnforceRunaiScheduler,
		},
	}

	queues, err := convertProjectQueues(old.Spec.Queues)
	if err != nil {
		return nil, fmt.Errorf("convert queues for project %s: %w", old.Name, err)
	}
	newProject.Spec.Queues = queues

	// A cascade deletion request becomes the new blocking deletion type; any other
	// value (orphan / unset) leaves DeletionType nil.
	if old.Spec.DeletionRequest != nil && old.Spec.DeletionRequest.Type == runaiv2.Cascade {
		blocking := kaiv1alpha1.Blocking
		newProject.Spec.DeletionType = &blocking
	}

	moveDroppedProjectFields(newProject, old)

	return newProject, nil
}

// moveDroppedProjectFields stashes the legacy project spec fields the new CRD drops
// onto labels (scalars) and annotations (lists), plus the status sync timestamp.
func moveDroppedProjectFields(newProject *kaiv1alpha1.Project, old *runaiv2.Project) {
	setAnnotation := func(key string, value string) {
		if value == "" {
			return
		}
		if newProject.Annotations == nil {
			newProject.Annotations = map[string]string{}
		}
		newProject.Annotations[key] = value
	}
	setAnnotationInt := func(key string, value int64) {
		if value != 0 {
			setAnnotation(key, strconv.FormatInt(value, 10))
		}
	}
	setAnnotationList := func(key string, values []string) {
		if len(values) == 0 {
			return
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			return
		}
		setAnnotation(key, string(encoded))
	}

	setAnnotationInt(constants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation, old.Spec.InteractivePreemptibleJobMaxIdleDurationSecs)
	setAnnotationInt(constants.InteractiveJobTimeLimitSecondsAnnotation, old.Spec.InteractiveJobTimeLimitSecs)
	setAnnotationInt(constants.InteractiveJobMaxIdleDurationSecondsAnnotation, old.Spec.InteractiveJobMaxIdleDurationSecs)
	setAnnotationInt(constants.TrainingJobTimeLimitSecondsAnnotation, old.Spec.TrainingJobTimeLimitSecs)
	setAnnotationInt(constants.TrainingJobMaxIdleDurationSecondsAnnotation, old.Spec.TrainingJobMaxIdleDurationSecs)

	if old.Spec.DeletionRequest != nil && old.Spec.DeletionRequest.Force {
		setAnnotation(constants.ForceDeleteAnnotation, strconv.FormatBool(true))
	}

	setAnnotationList(constants.NodeAffinityInteractiveAnnotation, old.Spec.NodeAffinityInteractive)
	setAnnotationList(constants.NodeAffinityTrainAnnotation, old.Spec.NodeAffinityTrain)

	if ts := old.Status.SyncStatus.LastSyncTimestamp; !ts.IsZero() {
		setAnnotation(hooks.LastSyncTimestampAnnotation, ts.UTC().Format("2006-01-02T15-04-05Z"))
	}
}

// convertProjectQueues converts the legacy project queue specs to the new
// QueueConfig shape via a JSON round-trip. Only the fields shared by both shapes
// (name/nodepool/resources/priority) survive; the rest are dropped per design.
func convertProjectQueues(old []runaiv2.QueueProjectSpec) ([]kaiv1alpha1.QueueConfig, error) {
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
	return queues, nil
}

// reownProjectOwnedObjects re-points every object owned by the old project at the
// new kai project. Queues are cluster-scoped and located by the project-name label;
// LimitRanges and RoleBindings live in the project namespace and are located by the
// project OwnerReference; the replicated Secrets/PVCs/ConfigMaps are located by the
// replication labels and re-pointed via those labels (not an owner ref). Every
// change is appended to reowned so migrateProject can roll the chain back on failure.
func reownProjectOwnedObjects(ctx context.Context, k8sClient client.Client, old *runaiv2.Project,
	newProject *kaiv1alpha1.Project, reowned *[]reownedObject) error {
	// Queues labelled with the project name.
	if err := reownOwnerRefObjects(ctx, k8sClient, &schedv2.QueueList{}, old, newProject, reowned,
		client.MatchingLabels{projectNameLabel: old.Name}); err != nil {
		return err
	}

	// Namespaced owned objects live in the project namespace.
	ns := extractProjectNamespace(old)

	// LimitRange and RoleBinding (project-controller + accessrule-controller) carry a
	// project OwnerReference but no identifying label, so they are matched by owner ref.
	if err := reownOwnerRefObjects(ctx, k8sClient, &corev1.LimitRangeList{}, old, newProject, reowned,
		client.InNamespace(ns)); err != nil {
		return err
	}
	if err := reownOwnerRefObjects(ctx, k8sClient, &rbacv1.RoleBindingList{}, old, newProject, reowned,
		client.InNamespace(ns)); err != nil {
		return err
	}

	// Cluster-wide replicated copies are tracked by label, not owner ref.
	for _, list := range []client.ObjectList{&corev1.SecretList{}, &corev1.PersistentVolumeClaimList{}, &corev1.ConfigMapList{}} {
		if err := relabelReplicatedObjects(ctx, k8sClient, list, ns, old, newProject, reowned); err != nil {
			return err
		}
	}

	return nil
}

// reownOwnerRefObjects lists objects matching baseOpts (paginated) and, for every
// object carrying an OwnerReference to the old project, rewrites that reference to
// point at the new kai project. Each rewritten object is recorded in reowned.
func reownOwnerRefObjects(ctx context.Context, k8sClient client.Client, list client.ObjectList, old *runaiv2.Project,
	newProject *kaiv1alpha1.Project, reowned *[]reownedObject, baseOpts ...client.ListOption) error {
	continueToken := ""
	for {
		listOpts := append([]client.ListOption{}, baseOpts...)
		listOpts = append(listOpts, client.Limit(hooks.ListPageSize))
		if continueToken != "" {
			listOpts = append(listOpts, client.Continue(continueToken))
		}
		if err := k8sClient.List(ctx, list, listOpts...); err != nil {
			return fmt.Errorf("list %T for project %s: %w", list, old.Name, err)
		}
		items, err := apimeta.ExtractList(list)
		if err != nil {
			return fmt.Errorf("extract %T items: %w", list, err)
		}

		for _, item := range items {
			obj := item.(client.Object)
			oldOwners := obj.GetOwnerReferences()
			updated, replaced := replaceProjectOwner(oldOwners, old, newProject)
			if !replaced {
				continue
			}
			patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
			obj.SetOwnerReferences(updated)
			if err := k8sClient.Patch(ctx, obj, patch); err != nil {
				return fmt.Errorf("patch owner of %T %s: %w", obj, obj.GetName(), err)
			}
			*reowned = append(*reowned, reownedObject{
				obj:       obj.DeepCopyObject().(client.Object),
				oldOwners: oldOwners,
				oldLabels: copyStringMap(obj.GetLabels()),
			})
		}

		if continueToken = list.GetContinue(); continueToken == "" {
			return nil
		}
	}
}

// relabelReplicatedObjects re-points the per-project replicated copies of a
// cluster-wide resource kind at the new kai project by rewriting the replication
// labels (project name + uid) and stripping any stale legacy project OwnerReference.
func relabelReplicatedObjects(ctx context.Context, k8sClient client.Client, list client.ObjectList, ns string,
	old *runaiv2.Project, newProject *kaiv1alpha1.Project, reowned *[]reownedObject) error {
	continueToken := ""
	for {
		listOpts := []client.ListOption{
			client.InNamespace(ns),
			client.MatchingLabels{constants.ReplicatedProjectUIDLabel: string(old.UID)},
			client.Limit(hooks.ListPageSize),
		}
		if continueToken != "" {
			listOpts = append(listOpts, client.Continue(continueToken))
		}
		if err := k8sClient.List(ctx, list, listOpts...); err != nil {
			return fmt.Errorf("list replicated %T for project %s: %w", list, old.Name, err)
		}
		items, err := apimeta.ExtractList(list)
		if err != nil {
			return fmt.Errorf("extract %T items: %w", list, err)
		}

		for _, item := range items {
			obj := item.(client.Object)
			oldOwners := obj.GetOwnerReferences()
			oldLabels := copyStringMap(obj.GetLabels())

			patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
			labels := copyStringMap(obj.GetLabels())
			if labels == nil {
				labels = map[string]string{}
			}
			labels[constants.ReplicatedProjectNameLabel] = newProject.Name
			labels[constants.ReplicatedProjectUIDLabel] = string(newProject.UID)
			obj.SetLabels(labels)

			// The replicated copies are tracked by label, but older ones may still
			// carry a legacy project OwnerReference (a pre-deprecation fallback) - strip
			// it so the deleted legacy project does not garbage-collect them.
			if stripped, changed := stripProjectOwner(oldOwners, old); changed {
				obj.SetOwnerReferences(stripped)
			}

			if err := k8sClient.Patch(ctx, obj, patch); err != nil {
				return fmt.Errorf("patch replicated %T %s: %w", obj, obj.GetName(), err)
			}
			*reowned = append(*reowned, reownedObject{
				obj:       obj.DeepCopyObject().(client.Object),
				oldOwners: oldOwners,
				oldLabels: oldLabels,
			})
		}

		if continueToken = list.GetContinue(); continueToken == "" {
			return nil
		}
	}
}

// replaceProjectOwner returns a copy of owners with any OwnerReference targeting the
// old project (a Project owner with the old project's UID) replaced by a reference to
// the new kai project, and reports whether such an owner was found. The replaced
// reference keeps its Controller / BlockOwnerDeletion flags, preserving the per-source
// difference (project-controller sets BlockOwnerDeletion; accessrule-controller does not).
func replaceProjectOwner(owners []metav1.OwnerReference, old *runaiv2.Project,
	newProject *kaiv1alpha1.Project) ([]metav1.OwnerReference, bool) {
	result := make([]metav1.OwnerReference, 0, len(owners))
	replaced := false
	for _, owner := range owners {
		if owner.Kind == projectKind && owner.UID == old.UID {
			newOwner := owner
			newOwner.APIVersion = kaiv1alpha1.GroupVersion.String()
			newOwner.Name = newProject.Name
			newOwner.UID = newProject.UID
			result = append(result, newOwner)
			replaced = true
			continue
		}
		result = append(result, owner)
	}
	return result, replaced
}

// stripProjectOwner returns a copy of owners with any OwnerReference targeting the
// old project removed, and whether one was removed.
func stripProjectOwner(owners []metav1.OwnerReference, old *runaiv2.Project) ([]metav1.OwnerReference, bool) {
	result := make([]metav1.OwnerReference, 0, len(owners))
	stripped := false
	for _, owner := range owners {
		if owner.Kind == projectKind && owner.UID == old.UID {
			stripped = true
			continue
		}
		result = append(result, owner)
	}
	return result, stripped
}

// rollbackProject undoes a failed project migration: it restores the owner
// references and labels of every object already re-pointed and deletes the new project.
func rollbackProject(ctx context.Context, k8sClient client.Client, newProject *kaiv1alpha1.Project, reowned []reownedObject) {
	logger := log.FromContext(ctx)
	for i := range reowned {
		obj := reowned[i].obj
		patch := client.MergeFrom(obj.DeepCopyObject().(client.Object))
		obj.SetOwnerReferences(reowned[i].oldOwners)
		obj.SetLabels(reowned[i].oldLabels)
		if err := k8sClient.Patch(ctx, obj, patch); err != nil {
			logger.Error(err, fmt.Sprintf("rollback: failed to restore %T %s", obj, obj.GetName()))
		}
	}

	// Strip finalizers first: the new project's controller is not running during the
	// migration window, so a copied finalizer would otherwise leave the project stuck
	// terminating instead of actually deleted.
	if len(newProject.Finalizers) > 0 {
		newProject.Finalizers = nil
		if err := k8sClient.Update(ctx, newProject); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf("rollback: failed to strip new project %s finalizers", newProject.Name))
		}
	}

	if err := k8sClient.Delete(ctx, newProject); err != nil && !apierrors.IsNotFound(err) {
		logger.Error(err, fmt.Sprintf("rollback: failed to delete new project %s", newProject.Name))
	}
}

// extractProjectNamespace returns the namespace a project's namespaced owned objects live
// in. Status.Namespace is authoritative (set by project-controller once the namespace
// exists); Spec.Namespace is the requested value; otherwise fall back to the default
// runai naming scheme.
func extractProjectNamespace(old *runaiv2.Project) string {
	if old.Status.Namespace != "" {
		return old.Status.Namespace
	}
	if old.Spec.Namespace != "" {
		return old.Spec.Namespace
	}
	return "runai-" + old.Name
}
