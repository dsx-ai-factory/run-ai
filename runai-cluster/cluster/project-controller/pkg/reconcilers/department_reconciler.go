/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcilers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	rateLimiterBaseDelay = time.Millisecond * 500
	rateLimiterMaxDelay  = time.Second * 10

	ProjectDepartmentIndexField = "spec.parent"
)

// IndexProjectByDepartment extracts the department a project references, for use
// as a field indexer value. Returns nil for projects with no department so they
// are not indexed.
func IndexProjectByDepartment(obj client.Object) []string {
	project, ok := obj.(*kaiv1alpha1.Project)
	if !ok || project.Spec.Parent == "" {
		return nil
	}
	return []string{project.Spec.Parent}
}

// DepartmentReconciler reconciles Department objects and creates associated queue resources
type DepartmentReconciler struct {
	client            client.Client
	Log               logr.Logger
	departmentHandler handlers.DepartmentHandler
}

func NewDepartmentReconciler(client client.Client) *DepartmentReconciler {
	logger := ctrl.Log.WithName(common.LogDepartmentTag)

	return &DepartmentReconciler{
		client:            client,
		Log:               logger,
		departmentHandler: handlers.NewDepartmentHandler(client),
	}
}

func (r *DepartmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	department := &kaiv1alpha1.Department{}
	if shouldContinue, err := r.getDepartmentForRequest(ctx, req, department); err != nil {
		return ctrl.Result{Requeue: true}, err
	} else if !shouldContinue {
		return ctrl.Result{Requeue: false}, nil
	}

	if !department.DeletionTimestamp.IsZero() {
		// The department is being deleted - finalize it. The finalizer blocks
		// deletion while any project still references this department in its spec.
		return r.finalize(ctx, department)
	}

	// The department is not being deleted - make sure this controller is registered
	// as a finalizer so we get a chance to validate it before it is actually deleted.
	if err := r.addFinalizerIfNeeded(ctx, department); err != nil {
		return ctrl.Result{Requeue: true}, err
	}

	err := r.departmentHandler.Handle(ctx, department)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	return ctrl.Result{}, nil
}

// finalize is invoked when a Department has a deletion timestamp. It only allows
// the deletion to proceed (by removing our finalizer) once no Project references
// the department in its spec.
func (r *DepartmentReconciler) finalize(ctx context.Context,
	department *kaiv1alpha1.Department) (ctrl.Result, error) {
	if !common.ContainsString(config.DepartmentFinalizerName(), department.Finalizers) {
		r.Log.Info("Got a finalization event for Department, but this controller is "+
			"not in its finalizers list. Skipping finalization.",
			common.LogDepartmentTag, department.Name)
		return ctrl.Result{}, nil
	}

	projects, err := r.getProjectsOfDepartment(ctx, department.Name)
	if err != nil {
		r.Log.Error(err, "Failed listing projects while finalizing department",
			common.LogDepartmentTag, department.Name)
		return ctrl.Result{}, err
	}

	if len(projects) > 0 {
		r.Log.Info("Department cannot be deleted - it still has projects assigned to it. "+
			"Blocking deletion until all assigned projects are removed.",
			common.LogDepartmentTag, department.Name, "projectsCount", len(projects),
			common.LogProjectTag, projects[0].Name)

		condition := kaiv1alpha1.DepartmentCondition{
			Type:   kaiv1alpha1.DepartmentDeletionBlocked,
			Status: corev1.ConditionTrue,
			Reason: "ProjectsStillAssigned",
			Message: fmt.Sprintf("%d project(s) still reference this department, e.g. %q",
				len(projects), projects[0].Name),
		}
		if err := r.updateDepartmentStatusCondition(ctx, department, condition); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, fmt.Errorf(
			"department %q cannot be deleted: %d project(s) still reference it (e.g. %q)",
			department.Name, len(projects), projects[0].Name)
	}

	if err := r.removeFinalizer(ctx, department); err != nil {
		return ctrl.Result{}, err
	}

	r.Log.Info("No projects reference the department - finalization complete",
		common.LogDepartmentTag, department.Name)
	return ctrl.Result{}, nil
}

// updateDepartmentStatusCondition upserts the given condition into the department's status
// (replacing any existing condition of the same type) and patches the status subresource.
// It is a no-op when the condition is already present with the same value, to avoid
// generating redundant status updates on every requeue.
func (r *DepartmentReconciler) updateDepartmentStatusCondition(ctx context.Context,
	department *kaiv1alpha1.Department, condition kaiv1alpha1.DepartmentCondition) error {
	if !handlers.UpdateDepartmentCondition(&department.Status, &condition) {
		return nil
	}

	patchBytes, err := json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		"conditions": department.Status.Conditions}})
	if err != nil {
		r.Log.Error(err, "Failed to json.Marshal patch - update status conditions of department",
			common.LogDepartmentTag, department.Name)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	if err := r.client.Status().Patch(ctx, department, patch); err != nil {
		r.Log.Error(err, "Failed to patch status conditions of department",
			common.LogDepartmentTag, department.Name)
		return err
	}
	return nil
}

// getProjectsOfDepartment returns all projects that reference the given department
// by name in their spec.
func (r *DepartmentReconciler) getProjectsOfDepartment(ctx context.Context,
	departmentName string) ([]kaiv1alpha1.Project, error) {
	projectList := &kaiv1alpha1.ProjectList{}
	if err := r.client.List(ctx, projectList,
		client.MatchingFields{ProjectDepartmentIndexField: departmentName}); err != nil {
		return nil, err
	}
	return projectList.Items, nil
}

func (r *DepartmentReconciler) addFinalizerIfNeeded(ctx context.Context,
	department *kaiv1alpha1.Department) error {
	if common.ContainsString(config.DepartmentFinalizerName(), department.Finalizers) {
		// Controller already present in finalizers list
		return nil
	}
	r.Log.Info("Adding controller to finalizer list in Department",
		common.LogDepartmentTag, department.Name)
	department.Finalizers = append(department.Finalizers, config.DepartmentFinalizerName())
	if err := r.updateFinalizers(ctx, department); err != nil {
		r.Log.Error(err, "Error adding controller to finalizer list in Department",
			common.LogDepartmentTag, department.Name)
		return err
	}
	return nil
}

func (r *DepartmentReconciler) removeFinalizer(ctx context.Context,
	department *kaiv1alpha1.Department) error {
	r.Log.Info("Removing finalizer from department's finalizers list",
		common.LogDepartmentTag, department.Name)
	department.Finalizers = common.DeleteTerm(config.DepartmentFinalizerName(), department.Finalizers)
	if err := r.updateFinalizers(ctx, department); err != nil {
		r.Log.Error(err, "Error removing controller from finalizer list in Department",
			common.LogDepartmentTag, department.Name)
		return err
	}
	return nil
}

func (r *DepartmentReconciler) updateFinalizers(ctx context.Context,
	department *kaiv1alpha1.Department) error {
	patchBytes, err := json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"finalizers": department.Finalizers}})
	if err != nil {
		r.Log.Error(err, "Failed to json.Marshal patch - update finalizers of department",
			common.LogDepartmentTag, department.Name)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	return r.client.Patch(ctx, department, patch)
}

func (r *DepartmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Register a field indexer so departments can look up the projects that reference
	// them by spec.parent (used by getProjectsOfDepartment during finalization).
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kaiv1alpha1.Project{},
		ProjectDepartmentIndexField, IndexProjectByDepartment); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kaiv1alpha1.Department{}).
		Owns(&kaiv2.Queue{}).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(r)
}

func (r *DepartmentReconciler) getDepartmentForRequest(ctx context.Context,
	req ctrl.Request, department *kaiv1alpha1.Department) (shouldContinue bool, err error) {
	err = r.client.Get(ctx, req.NamespacedName, department)
	if err == nil {
		return true, nil
	}

	if errors.IsNotFound(err) {
		r.Log.Info("Can't locate department in cluster, this might also mean it was purposefully deleted",
			common.LogDepartmentTag, req.Name)
		// Nothing really more to do.
		return false, nil
	}

	r.Log.Error(err, "Could not extract department from incoming request", common.LogDepartmentTag, req.Name)
	return false, err
}
