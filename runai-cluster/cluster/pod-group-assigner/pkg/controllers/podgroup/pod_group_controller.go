/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package podgroup

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup/assigner"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	commonconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	requeueTimeout = time.Second * 2

	rateLimiterBaseDelay = time.Second
	rateLimiterMaxDelay  = requeueTimeout
)

// PodGroupReconciler reconciles a PodGroup object
type PodGroupReconciler struct {
	cachedClient     client.Client
	podGroupAssigner *assigner.PodGroupAssigner
}

func NewPodGroupReconciler(mgrClient client.Client) *PodGroupReconciler {
	return &PodGroupReconciler{
		cachedClient:     mgrClient,
		podGroupAssigner: assigner.NewPodGroupAssigner(mgrClient),
	}
}

//+kubebuilder:rbac:groups=scheduling.run.ai,resources=podgroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scheduling.run.ai,resources=podgroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scheduling.run.ai,resources=podgroups/finalizers,verbs=update
//+kubebuilder:rbac:groups=kai.scheduler,resources=topologies,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PodGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)

	podgroup := &kaiv2alpha2.PodGroup{}
	if shouldContinue, err := r.getPodGroupForRequest(ctx, req, podgroup); err != nil {
		return ctrl.Result{Requeue: true}, err
	} else if !shouldContinue {
		return ctrl.Result{Requeue: false}, nil
	}

	log.Ctx(ctx).Info().Msgf("Reconciling podgroup <%v>", req.NamespacedName)

	if !podgroup.DeletionTimestamp.IsZero() {
		log.Ctx(ctx).Info().Msgf("Podgroup <%v> is being deleted - not assigning to Node Pools",
			req.NamespacedName)

		return ctrl.Result{}, nil
	}

	err := r.podGroupAssigner.Run(ctx, podgroup)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: requeueTimeout}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodGroupReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := r.indexFields(ctx, mgr)
	if err != nil {
		return err
	}

	r.podGroupAssigner = assigner.NewPodGroupAssigner(mgr.GetClient())

	// watches Node Pools as well - converting each to Reconcile on all Pod Groups of that Node Pool;
	// but with event filter - that filters in only Node Pools that are Deleting or Unschedulable.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kaiv2alpha2.PodGroup{}).
		Watches(nodePoolWatchObject(), handler.EnqueueRequestsFromMapFunc(r.MapNodePoolToPodGroupEvent)).
		WithEventFilter(predicate.NewPredicateFuncs(r.FilterPodGroupControllerEvents)).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](rateLimiterBaseDelay, rateLimiterMaxDelay)}).
		Complete(r)
}

func nodePoolWatchObject() client.Object {
	return &v1alpha1.NodePool{}
}

func (r *PodGroupReconciler) getPodGroupForRequest(ctx context.Context, req ctrl.Request, podGroup *kaiv2alpha2.PodGroup) (shouldContinue bool, bool error) {
	err := r.cachedClient.Get(ctx, req.NamespacedName, podGroup)
	if err == nil {
		return true, nil
	}

	if errors.IsNotFound(err) {
		log.Ctx(ctx).Warn().Msgf("Can't locate podgroup <%v> in cluster, this might also mean it was purposefully deleted",
			req.NamespacedName)
		// Nothing really more to do.
		return false, nil
	}

	log.Ctx(ctx).Error().Msgf("Could not extract pod group <%v> from incoming request, err: <%s>",
		req.NamespacedName, err.Error())

	return false, err
}

func (r *PodGroupReconciler) indexFields(ctx context.Context, mgr ctrl.Manager) error {
	err := mgr.GetFieldIndexer().IndexField(
		ctx, &corev1.Pod{},
		common.PodByPodGroupIndexerName, PodByPodGroupIndexer)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed indexing pod field: %s, err: <%s>",
			common.PodByPodGroupIndexerName, err.Error())

		return err
	}

	return nil
}

func PodByPodGroupIndexer(object client.Object) (indexedKeys []string) {
	pod, ok := object.(*corev1.Pod)
	if !ok {
		log.Error().Msgf("PodByPodGroupIndexer: Cannot convert object to *corev1.Pod: %v", object)
		return indexedKeys
	}

	podGroupName := pod.Annotations[commonconstants.PodGroupAnnotationForPod]

	return []string{podGroupName}
}
