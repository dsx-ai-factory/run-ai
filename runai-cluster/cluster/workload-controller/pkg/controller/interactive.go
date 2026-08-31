/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package controller

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/record"

	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaijob "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	iwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerUtil "sigs.k8s.io/controller-runtime/pkg/controller"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var logiw = logger.Log.WithName("reconcile.iw")

// InteractiveReconciler reconciles a InteractiveWorkload object
type InteractiveReconciler struct {
	client.Client
	APIReadClient     client.Reader
	Scheme            *runtime.Scheme
	Config            *rest.Config
	EventRecorder     record.EventRecorder
	controllerOptions Options
}

func NewSimpleInteractiveReconciler(client client.Client, apiReadClient client.Reader, config *rest.Config, eventRecorder record.EventRecorder, controllerOptions Options) *InteractiveReconciler {
	return &InteractiveReconciler{
		Client:            client,
		APIReadClient:     apiReadClient,
		Scheme:            client.Scheme(),
		Config:            config,
		EventRecorder:     eventRecorder,
		controllerOptions: controllerOptions,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *InteractiveReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconcile int) error {
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&iwv1.InteractiveWorkload{}).
		Owns(&runaijob.RunaiJob{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		WithOptions(controllerUtil.Options{MaxConcurrentReconciles: maxConcurrentReconcile,
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				constants.RateLimiterBaseDelay, constants.RateLimiterMaxDelay),
		}).
		WithEventFilter(ignoreDeletionPredicate())

	// Watch owned routing resources based on the active routing mode(s).
	// Both can be enabled simultaneously during migration from Ingress to Gateway API.
	if workloadconfig.IngressEnabled() {
		if workloadconfig.IsOpenshift() {
			manager.Owns(&openshiftV1.Route{})
		} else {
			manager.Owns(&v1.Ingress{})
		}
	}
	if workloadconfig.GatewayEnabled() {
		manager.Owns(&gatewayv1.HTTPRoute{})
	}

	return manager.Complete(r)

}

// +kubebuilder:rbac:groups=run.ai,resources=interactiveworkload,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.ai,resources=interactiveworkload/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=run.ai,resources=interactiveworkload/finalizers,verbs=update
// +kubebuilder:rbac:groups=run.ai,resources=runaijob,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.ai,resources=runaijob/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=run.ai,resources=trainingworkload/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=ingress,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=ingress/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=PersistentVolumeClaim,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=resource.nvidia.com,resources=computedomains,verbs=get;list;watch;create;update;delete

func (r *InteractiveReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger.IntoContext(ctx, logiw)

	resrcId := fmt.Sprintf("%s:%s", req.Namespace, req.Name)

	logiw.Info("reconcile started", "id", resrcId)

	//
	//    obtain the resource being reconciled
	//
	interactiveWorkload := &iwv1.InteractiveWorkload{}

	err := r.Client.Get(ctx, req.NamespacedName, interactiveWorkload)
	if err != nil {
		//
		//   there was an error to get the resource from k8s
		//
		statusError := err.(*k8sErrors.StatusError)
		if statusError == nil || statusError.Status().Code != 404 {
			logiw.Error(err, "failed to load the workload", "id", resrcId)
		} else {
			logiw.Info("workload has being deleted", "id", resrcId)
		}

		//
		//    assume that the resource has been deleted with so just return normally
		//
		return ctrl.Result{}, nil
	}

	var jobWorkload jobs.JobWorkload
	interactiveWorkload.CopyTo(&jobWorkload)
	workloadControllerFactory := NewInteractiveWorkloadControllerFactory(r.Client, r.APIReadClient,
		workloadconfig.IsOpenshift(), r.EventRecorder, r.controllerOptions)
	workloadController, err := workloadControllerFactory.CreateController()
	if err != nil {
		logiw.Error(err, "failed to create workload controller", "id", resrcId)
		return ctrl.Result{}, err
	}
	result, err := workloadController.Reconcile(ctx, req, &jobWorkload, interactiveWorkload)
	if err != nil {
		logiw.Error(err, "failed to reconcile the workload", "id", resrcId)
		return ctrl.Result{}, err
	}

	return result, nil
}
