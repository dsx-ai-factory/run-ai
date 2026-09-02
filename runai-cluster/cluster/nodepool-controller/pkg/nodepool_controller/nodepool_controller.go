/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"

	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	requeueTimeout = time.Second * 10

	rateLimiterBaseDelay = 500 * time.Millisecond
	rateLimiterMaxDelay  = 30 * time.Second

	serviceMonitorCRDName = "servicemonitors.monitoring.coreos.com"
)

type NodePoolController struct {
	Client client.Client
	Scheme *runtime.Scheme
	params *common.NodePoolControllerParams

	// nrtEnabled is true when the NodeResourceTopology CRD is installed at startup;
	// it gates NRT-health evaluation.
	nrtEnabled bool

	// serviceMonitorEnabled is true when the ServiceMonitor CRD is installed at startup;
	// it gates both the ServiceMonitor watch and the per-nodepool ServiceMonitor creation,
	// so the controller runs on clusters without Prometheus.
	serviceMonitorEnabled bool
}

func NewNodePoolController(client client.Client, scheme *runtime.Scheme,
	nodePoolControllerParams *common.NodePoolControllerParams) *NodePoolController {
	return &NodePoolController{
		Client: client,
		Scheme: scheme,
		params: nodePoolControllerParams,
	}
}

//+kubebuilder:rbac:groups=kai.resources,resources=nodepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kai.resources,resources=nodepools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kai.resources,resources=nodepools/finalizers,verbs=update
//+kubebuilder:rbac:groups=run.ai,resources=projects,verbs=get;list;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;update;watch;patch;list
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;watch;list
//+kubebuilder:rbac:groups=kai.scheduler,resources=schedulingshards,verbs=get;list;watch;create;update;patch;delete

func (npc *NodePoolController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	nodePool := &v1alpha1.NodePool{}
	if shouldContinue, err := npc.getNodePoolForRequest(ctx, req, nodePool); err != nil {
		return ctrl.Result{}, err
	} else if !shouldContinue {
		return ctrl.Result{}, nil
	}

	var err error

	if nodePool.DeletionTimestamp.IsZero() {
		if err = npc.addControllerAsFinalizerIfNeeded(ctx, nodePool); err != nil {
			return ctrl.Result{}, err
		}

		err = npc.reconcileNodePool(ctx, nodePool)
	} else {
		err = npc.finalize(ctx, nodePool)
	}

	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (npc *NodePoolController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	err := npc.indexFields(ctx, mgr)
	if err != nil {
		return err
	}

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.NodePool{}).
		Owns(&kaiv1.SchedulingShard{}).
		Watches(&corev1.Node{}, handler.EnqueueRequestsFromMapFunc(npc.MapNodeToNodePoolEvent)).
		Watches(projectWatchObject(), handler.EnqueueRequestsFromMapFunc(npc.MapProjectToNodePoolEvents)).
		WithEventFilter(ControllerPredicateFuncs()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](rateLimiterBaseDelay, rateLimiterMaxDelay)})

	npc.serviceMonitorEnabled = utils.IsCRDInstalled(ctx, mgr.GetAPIReader(), serviceMonitorCRDName)
	if npc.serviceMonitorEnabled {
		builder = builder.Owns(&monitorv1.ServiceMonitor{})
	} else {
		log.Info().Msg("ServiceMonitor CRD not installed (no Prometheus); skipping ServiceMonitor watch and creation")
	}

	npc.nrtEnabled = isNrtCRDInstalled(ctx, mgr.GetAPIReader())
	if npc.nrtEnabled {
		log.Info().Msg("NodeResourceTopology CRD detected; watching NRT for node NRT-health conditions")
		return builder.Watches(&nrtv1alpha2.NodeResourceTopology{},
			handler.EnqueueRequestsFromMapFunc(npc.MapNrtToNodePoolEvent)).Complete(npc)
	}

	log.Info().Msg("NodeResourceTopology CRD not installed; will restart to watch NRT once its CRD is installed")
	if err = builder.Complete(npc); err != nil {
		return err
	}
	return npc.watchForNrtCRD(ctx, mgr)
}

func (npc *NodePoolController) indexFields(ctx context.Context, mgr ctrl.Manager) (err error) {
	err = mgr.GetFieldIndexer().IndexField(
		ctx, &v1alpha1.NodePool{},
		common.IsDeletingPhaseField, NodePoolIsDeletingPhaseIndexer)
	if err != nil {
		log.Error().Msgf("Failed indexing nodepool field: %v, err: %v", common.IsDeletingPhaseField, err.Error())
		return err
	}

	err = mgr.GetFieldIndexer().IndexField(
		ctx, &corev1.Pod{},
		common.PodRunningWithRunaiSchedulerNodeNameField, PodRunningWithRunaiSchedulerNodeNameIndexer)
	if err != nil {
		log.Error().Msgf("Failed indexing pod field: %v, err: %v", common.PodRunningWithRunaiSchedulerNodeNameField, err.Error())
		return err
	}

	return nil
}

func (npc *NodePoolController) getNodePoolForRequest(ctx context.Context, req ctrl.Request, nodePool *v1alpha1.NodePool) (shouldContinue bool, err error) {
	err = npc.Client.Get(ctx, req.NamespacedName, nodePool)
	if err == nil {
		return true, err
	}

	if errors.IsNotFound(err) {
		log.Error().Msgf("Can't locate node pool <%v> in cluster, this might also mean it was purposefully deleted", req.Name)
		// Nothing really more to do.
		return false, nil
	}
	log.Error().Msgf("Could not extract node pool <%v> from incoming request: %v, err: %v", req.Name, req, err.Error())
	return false, err
}

func NodePoolIsDeletingPhaseIndexer(object client.Object) (indexedKeys []string) {
	nodePool, ok := object.(*v1alpha1.NodePool)
	if !ok {
		log.Error().Msgf("NodePoolIsDeletingPhaseIndexer: Cannot convert object to *v1alpha1.NodePool: %v", object)
		return indexedKeys
	}

	isDeleting := !nodePool.DeletionTimestamp.IsZero() || nodePool.Status.Phase == v1alpha1.NodePoolDeleting
	return []string{strconv.FormatBool(isDeleting)}
}

func PodRunningWithRunaiSchedulerNodeNameIndexer(object client.Object) (indexedKeys []string) {
	pod, ok := object.(*corev1.Pod)
	if !ok {
		log.Error().Msgf("PodRunningWithRunaiSchedulerNodeNameIndexer: Cannot convert object to *corev1.Pod: %v", object)
		return indexedKeys
	}

	// this will filter out pods that are not running
	// and don't have runai-scheduler as scheduler name
	indexedKey := ""
	if pod.Spec.SchedulerName == config.Get().SchedulerName && isPodBoundToNode(pod) {
		indexedKey = pod.Spec.NodeName
	}
	return []string{indexedKey}
}

func isPodBoundToNode(pod *corev1.Pod) bool {
	return pod.Status.Phase == corev1.PodRunning ||
		(pod.Status.Phase == corev1.PodPending && pod.Spec.NodeName != "")
}

func (npc *NodePoolController) SetServiceMonitorEnabled(enabled bool) {
	npc.serviceMonitorEnabled = enabled
}

func (npc *NodePoolController) SetNrtEnabled(enabled bool) {
	npc.nrtEnabled = enabled
}
