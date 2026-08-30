package managed_nodes_config

import (
	"context"
	"errors"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

//+kubebuilder:rbac:groups=kai.resources,resources=managednodesconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kai.resources,resources=managednodesconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kai.resources,resources=managednodesconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;update;watch;patch;list

func (mncc *ManagedNodesConfigController) Reconcile(ctx context.Context, req MNCReconcileRequest) (ctrl.Result, error) {
	mnc, err := mncc.getManagedNodesConfig(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}

	var toBeExcludedNodes []corev1.Node
	if req.isNode {
		err = mncc.reconcileNode(ctx, req, mnc)
	} else {
		toBeExcludedNodes, err = mncc.reconcileManagedNodesConfig(ctx, req, mnc)
	}

	innerErr := mncc.reconcileStatus(ctx, req, mnc, toBeExcludedNodes)
	err = errors.Join(err, innerErr)

	if err != nil {
		log.Error().Msgf("Got an error in reconcile loop of %s. error: %v", req.Name, err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (mncc *ManagedNodesConfigController) reconcileManagedNodesConfig(ctx context.Context, req MNCReconcileRequest, mnc *v1alpha1.ManagedNodesConfig) (nodes []corev1.Node, err error) {
	log.Info().Msgf("Reconciling Managed nodes config <%v>", req.Name)

	allNodes, err := mncc.NodePoolController.ListNodesWithRequirements(ctx)
	if err != nil {
		return nodes, err
	}

	includedNodes, excludedNodes := mncc.splitNodeToIncludedExcluded(allNodes.Items)

	toBeExcludedNodes, innerErr := mncc.reconcileIncludedNodesThatShouldBeExcluded(ctx, mnc, includedNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	innerErr = mncc.reconcileExcludedNodesThatShouldBeIncluded(ctx, mnc, excludedNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	innerErr = mncc.reconcileMarkedToBeExcludedAndRevertedToBeIncluded(ctx, mnc, includedNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	return toBeExcludedNodes, err
}

func (mncc *ManagedNodesConfigController) reconcileNode(ctx context.Context, req MNCReconcileRequest, mnc *v1alpha1.ManagedNodesConfig) (err error) {
	log.Info().Msgf("Reconciling node <%v>", req.Name)

	var node corev1.Node
	err = mncc.Client.Get(ctx, types.NamespacedName{
		Name: req.Name,
	}, &node)
	if err != nil {
		return err
	}

	includedNodes, excludedNodes := mncc.splitNodeToIncludedExcluded([]corev1.Node{node})

	_, innerErr := mncc.reconcileIncludedNodesThatShouldBeExcluded(ctx, mnc, includedNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	innerErr = mncc.reconcileExcludedNodesThatShouldBeIncluded(ctx, mnc, excludedNodes)
	err = utils.AppendErrIfNotNil(err, innerErr)

	return err
}

func (mncc *ManagedNodesConfigController) splitNodeToIncludedExcluded(nodes []corev1.Node) (includedNodes []corev1.Node, excludedNodes []corev1.Node) {
	for _, node := range nodes {
		if node.Labels[config.Get().NodePoolNameLabel] == config.Get().ExcludedNodepoolName {
			excludedNodes = append(excludedNodes, node)
		} else {
			includedNodes = append(includedNodes, node)
		}
	}
	return includedNodes, excludedNodes
}
