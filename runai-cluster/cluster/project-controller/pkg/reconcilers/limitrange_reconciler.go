/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcilers

import (
	"context"

	"github.com/go-logr/logr"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type LimitRangeReconciler struct {
	IntraEventSender
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func NewLimitRangeReconciler(client client.Client, scheme *runtime.Scheme, projectEvents chan event.GenericEvent) *LimitRangeReconciler {
	return &LimitRangeReconciler{
		Client:           client,
		Log:              ctrl.Log.WithName("reconcilers").WithName(common.LogLimitRangeTag),
		Scheme:           scheme,
		IntraEventSender: IntraEventSender{projectEvents: projectEvents},
	}
}

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile reconciles by means of triggering events on all Projects when a change is detected on the default limit range configmap.
func (reconciler LimitRangeReconciler) Reconcile(_ context.Context, req ctrl.Request) (ctrl.Result, error) {
	// The filter already makes sure the incoming event is for a configmap under the scheduler namespace, named default-limit-range
	// So all that's left is just to trigger events for all Projects.
	reconciler.Log.Info("Handling event for Configmap: ", "Configmap", req.String())
	projectsList := kaiv1alpha1.ProjectList{}
	if err := reconciler.List(context.Background(), &projectsList); err != nil {
		reconciler.Log.Error(err, "Error while listing Projects during reconciliation of Configmap: ", "Configmap", req.String())
		return ctrl.Result{}, err
	}
	return reconciler.triggerEventsForAllProjects(projectsList)
}

func (reconciler *LimitRangeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).Named("LimitRangeReconciler").
		For(&corev1.ConfigMap{}).
		WithEventFilter(predicate.NewPredicateFuncs(FilterConfigmapEvent)).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(reconciler)
}

// FilterConfigmapEvent filters configmap events for the LimitRange reconciler.
func FilterConfigmapEvent(obj client.Object) (shouldAllowEvent bool) {
	switch obj.(type) {
	case *corev1.ConfigMap:
		isInstallNamespace := obj.GetNamespace() == config.Get().InstallNamespace
		isLimitRangeConfigMap := obj.GetName() == common.DefaultLimitRangeConfigMapName
		shouldAllowEvent = isInstallNamespace && isLimitRangeConfigMap
		log.V(8).Info("Event for Configmap being allowed through the filter", "Resource Name", obj.GetName(), "Resource Namespace", obj.GetNamespace())
	}
	return shouldAllowEvent
}
