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

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaijob "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	twv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	openshiftV1 "github.com/openshift/api/route/v1"
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

var logtw = logger.Log.WithName("reconcile.tw")

// TrainingReconciler reconciles a TrainingWorkload object
type TrainingReconciler struct {
	client.Client
	APIReadClient     client.Reader
	Scheme            *runtime.Scheme
	Config            *rest.Config
	EventRecorder     record.EventRecorder
	podLabels         map[string]string
	podAnnotations    map[string]string
	controllerOptions Options
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrainingReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconcile int) error {
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&twv1.TrainingWorkload{}).
		Owns(&runaijob.RunaiJob{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Secret{}).
		WithOptions(controllerUtil.Options{MaxConcurrentReconciles: maxConcurrentReconcile,
			RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
				constants.RateLimiterBaseDelay, constants.RateLimiterMaxDelay)}).
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

func NewSimpleTrainingReconciler(client client.Client, apiReadClient client.Reader, config *rest.Config,
	eventRecorder record.EventRecorder, podLabels map[string]string, podAnnotations map[string]string, controllerOptions Options) *TrainingReconciler {

	return &TrainingReconciler{
		Client:            client,
		APIReadClient:     apiReadClient,
		Scheme:            client.Scheme(),
		Config:            config,
		EventRecorder:     eventRecorder,
		podLabels:         podLabels,
		podAnnotations:    podAnnotations,
		controllerOptions: controllerOptions,
	}
}

//+kubebuilder:rbac:groups=run.ai,resources=trainingworkload,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=trainingworkload/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=run.ai,resources=runaijob,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=runaijob/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=run.ai,resources=trainingworkload/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=ingress,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=core,resources=ingress/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=PersistentVolumeClaim,verbs=get;list;watch;create;update;delete
//+kubebuilder:rbac:groups=resource.nvidia.com,resources=computedomains,verbs=get;list;watch;create;update;delete

func (r *TrainingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger.IntoContext(ctx, logtw)

	resrcId := fmt.Sprintf("%s:%s", req.Namespace, req.Name)

	logtw.Info("reconcile started", "id", resrcId)

	//
	//    obtain the resource being reconciled
	//
	trainingWorkload := &twv1.TrainingWorkload{}

	err := r.Client.Get(ctx, req.NamespacedName, trainingWorkload)
	if err != nil {
		//
		//   there was an error to get the resource from k8s
		//
		statusError := err.(*k8sErrors.StatusError)
		if statusError == nil || statusError.Status().Code != 404 {
			logtw.Error(err, "failed to load the workload", "id", resrcId)
		} else {
			logtw.Info("workload has being deleted", "id", resrcId)
		}

		//
		//    assume that the resource has been deleted with so just return normally
		//
		return ctrl.Result{}, nil
	}

	var jobWorkload jobs.JobWorkload
	trainingWorkload.CopyTo(&jobWorkload)
	trainingWorkloadFactory := NewTrainingWorkloadControllerFactory(r.Client, r.APIReadClient, workloadconfig.IsOpenshift(),
		r.EventRecorder, r.podLabels, r.podAnnotations, r.controllerOptions)
	trainingWorkloadController, err := trainingWorkloadFactory.CreateController()
	if err != nil {
		logtw.Error(err, "failed to create training workload controller", "id", resrcId)
		return ctrl.Result{}, err
	}
	result, err := trainingWorkloadController.Reconcile(ctx, req, &jobWorkload, trainingWorkload)
	if err != nil {
		logtw.Error(err, "failed to reconcile the workload", "id", resrcId)
		return ctrl.Result{}, err
	}

	return result, nil
}
