/*
 * SPDX-FileCopyrightText: Copyright (c) 2023 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"fmt"
	"strings"

	clusterrelations "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/cluster_relations"
	"github.com/go-logr/logr"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const runaiSchedulerName = "runai-scheduler"

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	DynamicClient   dynamic.Interface
	DiscoveryClient discovery.DiscoveryInterface
	EventRecorder   record.EventRecorder
	Scheme          *runtime.Scheme
}

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.V(1).Info("Reconciling pod")
	pod, err := r.getPodObject(ctx, req)
	if err != nil {
		if errors.IsNotFound(err) {
			return handleReconcileDeletedPod(logger, req)
		} else {
			logger.V(0).Error(err, "Failed to get pod", req.Namespace, req.Name)
			return ctrl.Result{}, err
		}
	}

	err = updatePod(ctx, pod, r.Client)
	if err != nil {
		logger.V(0).Error(err,
			fmt.Sprintf("Failed to update pod status: <%v/%v>", pod.Namespace, pod.Name))
		return ctrl.Result{}, err
	}

	podGroup, err := clusterrelations.GetPodGroupFromPod(ctx, pod, r.Client)
	if err != nil {
		return handleGetPodGroupError(err, logger, pod)
	}

	err = handleWorkload(ctx, r, pod, podGroup)
	if err != nil {
		logger.V(0).Error(err, "Reconcile pod handlePodGroupAndWorkload error")
	}

	return ctrl.Result{}, err
}

// handleDeletePodEvent complements the Reconcile function of PodReconciler for "delete pod" events.
// The regular reconciler can't work on a delete-pod event - it drops the "previous state" object and gives you only
// the name and the namespace of the reconciled pod. This isn't good enough for the status-updater,
// because we still need to update the relevant pod group and top owner.
func (r *PodReconciler) handleDeletePodEvent(_ context.Context, deleteEvent event.DeleteEvent,
	_ workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	ctx := context.Background()
	logger := log.FromContext(ctx)

	pod := deleteEvent.Object
	logger.V(1).Info(
		fmt.Sprintf("Delete event for pod <%s/%s>", pod.GetNamespace(), pod.GetName()))

	podGroup, err := clusterrelations.GetPodGroupFromDeletedPodObject(ctx, pod, r.Client)
	if err != nil {
		if strings.HasSuffix(err.Error(), "not found") {
			logger.V(1).Info(fmt.Sprintf("failed to find pod group on pod <%s/%s> delete event",
				pod.GetNamespace(), pod.GetName()))
		} else {
			logger.V(0).Error(err,
				fmt.Sprintf("failed to extract podGroup related to pod <%s/%s> with error.",
					pod.GetNamespace(), pod.GetName()))
		}
		return
	}

	err = handleWorkload(ctx, r, pod, podGroup)
	if err != nil {
		logger.V(0).Error(err, "Deleted pod handleWorkload error")
	}
}

func handleWorkload(ctx context.Context, r *PodReconciler, pod client.Object, podGroup *kaiv2alpha2.PodGroup) error {
	topOwnerRef, err := clusterrelations.GetPodGroupTopOwnerRef(podGroup)
	if err != nil {
		return fmt.Errorf("failed to get top owner reference from the pod group <%s/%s>. Error %w",
			podGroup.Namespace, podGroup.Name, err)
	}
	err = updateWorkloadData(ctx, topOwnerRef, podGroup.Namespace, r.Client, r.DynamicClient, r.DiscoveryClient,
		r.EventRecorder)
	if err != nil {
		return fmt.Errorf("failed to update workload data on top owner and pod groups. Error: %w", err)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &v1.Pod{}, clusterrelations.PodGroupAnnotationForPodPath,
		podGroupNameIndexerFunc)
	if err != nil {
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(
		context.Background(), &kaiv2alpha2.PodGroup{}, clusterrelations.TopOwnerUidForPodGropPath,
		topOwnerUidIndexerFunc)
	if err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Pod{}).
		Watches(&v1.Pod{},
			handler.Funcs{
				DeleteFunc: r.handleDeletePodEvent,
			},
			builder.WithPredicates(isRunAiPodPredicate()),
		).
		WithEventFilter(isRunAiPodPredicate()).
		Complete(r)
}

func (r *PodReconciler) getPodObject(ctx context.Context, req ctrl.Request) (
	*v1.Pod, error) {
	pod := v1.Pod{}
	err := r.Client.Get(ctx, req.NamespacedName, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

func handleReconcileDeletedPod(logger logr.Logger, req ctrl.Request) (ctrl.Result, error) {
	logger.V(1).Info(
		fmt.Sprintf("Pod %v not found, was probably deleted. The event will be handled by handleDeletePodEvent",
			req))
	return ctrl.Result{}, nil
}

func handleGetPodGroupError(err error, logger logr.Logger, pod *v1.Pod) (ctrl.Result, error) {
	if strings.HasSuffix(err.Error(), "not found") {
		logger.V(1).Info(
			fmt.Sprintf("No podGroup for pod <%s/%s> phase %s", pod.Namespace, pod.Name, pod.Status.Phase))
		return ctrl.Result{}, nil
	} else {
		return ctrl.Result{}, fmt.Errorf("failed to extract podGroup related to pod <%s/%s> with error %w",
			pod.Namespace, pod.Name, err)
	}
}
