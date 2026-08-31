/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

package controller

import (
	"context"
	"errors"
	"fmt"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions"
	"github.com/run-ai/runai/runai-cluster/cluster/common/extensions/distributed"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	distributedpolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	core1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerUtil "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var DistributedJobCRDsSupportedMap = map[string]client.Object{
	constants.KindMpiJob: &v2beta1.MPIJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindMpiJob,
			APIVersion: "kubeflow.org/v2beta1",
		},
	},
	constants.KindPyTorchJob: &kubeflowv1.PyTorchJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindPyTorchJob,
			APIVersion: "kubeflow.org/v1",
		},
	},
	constants.KindTFJob: &kubeflowv1.TFJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindTFJob,
			APIVersion: "kubeflow.org/v1",
		},
	},
	constants.KindXGBoostJob: &kubeflowv1.XGBoostJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindXGBoostJob,
			APIVersion: "kubeflow.org/v1",
		},
	},
	constants.KindJAXJob: &kubeflowv1.JAXJob{
		TypeMeta: v1.TypeMeta{
			Kind:       constants.KindJAXJob,
			APIVersion: "kubeflow.org/v1",
		},
	},
}

type DistributedJobFrameworksStatus struct {
	JobObject      client.Object
	SuspendEnabled bool
}

// DistributedWorkloadReconciler reconciles a DistributedWorkload object
type DistributedWorkloadReconciler struct {
	client.Client
	apiReadClient       client.Reader
	enabledFrameworks   map[string]client.Object
	supportedFrameworks map[string]client.Object
	EventRecorder       record.EventRecorder
	podLabels           map[string]string
	podAnnotations      map[string]string
	controllerOptions   Options
}

func NewSimpleDistributedWorkloadReconciler(k8sClient client.Client, apiReadClient client.Reader,
	enabledFrameworks map[string]DistributedJobFrameworksStatus, supportedFrameworks map[string]client.Object,
	eventRecorder record.EventRecorder, podLabels map[string]string,
	podAnnotations map[string]string, controllerOptions Options) *DistributedWorkloadReconciler {

	// read only copy of the maps
	enabledFrameworkCopy := make(map[string]client.Object, len(enabledFrameworks))
	supportedFrameworkCopy := make(map[string]client.Object, len(supportedFrameworks))
	for k, v := range enabledFrameworks {
		enabledFrameworkCopy[k] = v.JobObject
	}
	for k, v := range supportedFrameworks {
		supportedFrameworkCopy[k] = v
	}

	return &DistributedWorkloadReconciler{
		Client:              k8sClient,
		apiReadClient:       apiReadClient,
		enabledFrameworks:   enabledFrameworkCopy,
		supportedFrameworks: supportedFrameworkCopy,
		EventRecorder:       eventRecorder,
		podLabels:           podLabels,
		podAnnotations:      podAnnotations,
		controllerOptions:   controllerOptions,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DistributedWorkloadReconciler) SetupWithManager(mgr ctrl.Manager, maxConcurrentReconcile int, mnnvlSupported bool) error {
	log := logger.Log.WithName("controller.distributed")
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&v2alpha1.DistributedWorkload{}).
		Owns(&core1.Service{}).
		Owns(&core1.PersistentVolumeClaim{}).
		Owns(&core1.Secret{}).
		Watches(
			&core1.Pod{},
			handler.EnqueueRequestsFromMapFunc(DistributedWorkloadPodsFilter()),
			builder.WithPredicates(DistributedWorkloadPodPredicate()),
		).
		Named("distributed-controller").
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
			manager.Owns(&netv1.Ingress{})
		}
	}
	if workloadconfig.GatewayEnabled() {
		manager.Owns(&gatewayv1.HTTPRoute{})
	}

	if mnnvlSupported {
		manager.Owns(&nvapi.ComputeDomain{})
	}

	for framework, object := range r.enabledFrameworks {
		if object != nil {
			log.Info(fmt.Sprintf("framework %s is enabled", framework))
			manager.Owns(object)
		}
	}

	return manager.Complete(r)
}

//+kubebuilder:rbac:groups=run.ai,resources=distributedworkload,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=distributedworkload/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=run.ai,resources=distributedworkload/finalizers,verbs=update
//+kubebuilder:rbac:groups=resource.nvidia.com,resources=computedomains,verbs=get;list;watch;create;update;delete

func (r *DistributedWorkloadReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var log = logger.Log.WithName("reconcile.distributed")
	logger.IntoContext(ctx, log)

	requestID := fmt.Sprintf("%s:%s", req.Namespace, req.Name)

	log.Info("reconcile started", "id", requestID)

	//
	//    obtain the resource being reconciled
	//
	workload := &v2alpha1.DistributedWorkload{}

	err := r.Client.Get(ctx, req.NamespacedName, workload)
	if err != nil {
		//
		//   there was an error to get the resource from k8s
		//
		var statusError *k8sErrors.StatusError
		errors.As(err, &statusError)
		if statusError == nil || statusError.Status().Code != 404 {
			log.Error(err, "failed to load the workload", "id", requestID)
		} else {
			log.Info("workload has being deleted", "id", requestID)
		}

		//
		//    assume that the resource has been deleted with so just return normally
		//
		return ctrl.Result{}, nil
	}

	// Skip workloads owned by another controller (e.g. the WorkloadRunner restart flow).
	if workload.Spec.ManagedBy != "" {
		log.V(1).Info("workload is managed by another controller, skipping", "id", requestID, "managedBy", workload.Spec.ManagedBy)
		return ctrl.Result{}, nil
	}

	err = r.validateJobFramework(workload, log, requestID)
	if err != nil {
		return ctrl.Result{}, err
	}
	materializeImplicitMasterSpec(workload)
	var jobWorkload jobs.JobWorkload
	workload.CopyTo(&jobWorkload)
	workloadControllerFactory := NewDistributedWorkloadControllerFactory(r.Client, r.apiReadClient,
		workloadconfig.IsOpenshift(), r.EventRecorder, r.podLabels, r.podAnnotations, r.controllerOptions)
	workloadController, err := workloadControllerFactory.CreateController()
	if err != nil {
		log.Error(err, "failed to create workload controller", "id", requestID)
		return ctrl.Result{}, err
	}
	result, err := workloadController.Reconcile(ctx, req, &jobWorkload, workload)
	if err != nil {
		log.Error(err, "failed to reconcile the workload", "id", requestID)
		return ctrl.Result{}, err
	}

	return result, nil
}

func materializeImplicitMasterSpec(workload *v2alpha1.DistributedWorkload) {
	if workload.MasterSpec != nil || workload.Spec.NoMaster.Deref() {
		return
	}

	// A missing masterSpec means that the master inherits the worker pod fields. This used to be
	// materialized by the retired mutation webhook, so do it in-memory before building child jobs.
	masterSpec := &distributedpolicy.DistributedMasterSpec{}
	workload.Spec.WorkloadFields.DeepCopyInto(&masterSpec.WorkloadFields)
	workload.Spec.ResourceAllocationFields.DeepCopyInto(&masterSpec.ResourceAllocationFields)
	workload.Spec.RunaiSchedulerFields.DeepCopyInto(&masterSpec.RunaiSchedulerFields)
	workload.Spec.AccessControlFields.DeepCopyInto(&masterSpec.AccessControlFields)
	workload.MasterSpec = masterSpec
}

func (r *DistributedWorkloadReconciler) validateJobFramework(workload *v2alpha1.DistributedWorkload, log logr.Logger, resrcId string) error {
	if supportedFramework := r.enabledFrameworks[workload.Spec.JobType]; supportedFramework == nil {
		frameworkObj := r.supportedFrameworks[workload.Spec.JobType]
		if frameworkObj == nil {
			err := fmt.Errorf("job type %s is not recognized", workload.Spec.JobType)
			log.Error(err, "failed to validate the job framework", "id", resrcId)
			return err
		}
		err := fmt.Errorf("job type %s is not enabled", workload.Spec.JobType)
		log.Error(err, fmt.Sprintf("%v CRD should be installed before submiting distributed workload",
			frameworkObj.GetObjectKind().GroupVersionKind()), "id", resrcId)
		return err
	}
	return nil
}

func GetDistributedJobFrameworksStatus(extensionsGroupStatus extensions.Group) (map[string]DistributedJobFrameworksStatus, error) {
	enabledJobFrameworks := make(map[string]DistributedJobFrameworksStatus, len(extensionsGroupStatus))

	for _, extension := range extensionsGroupStatus {
		if !extension.IsEnabled() {
			continue
		}
		jobKind, ok := distributed.ExtensionToJobKindMap[extension.GetType()]
		if !ok {
			return nil, fmt.Errorf("the job kind for the extension %s were not found", extension.GetType().String())
		}
		jobObject, ok := DistributedJobCRDsSupportedMap[jobKind]
		if !ok {
			return nil, fmt.Errorf("the job object for the extension %s were not found", extension.GetType().String())
		}
		enabledJobFrameworks[jobKind] = DistributedJobFrameworksStatus{
			jobObject,
			extension.IsCapable(distributed.Suspension),
		}
	}

	return enabledJobFrameworks, nil
}
