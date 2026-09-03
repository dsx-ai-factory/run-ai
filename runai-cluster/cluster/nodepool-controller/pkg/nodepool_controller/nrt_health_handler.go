/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	toolscache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// osExit is a seam so the CRD-installed restart in watchForNrtCRD can be asserted in tests.
var osExit = os.Exit

const (
	// NrtHealthyConditionType is set on a node in a NUMA-aware nodepool when its NRT data is
	// missing or unusable. It is present ONLY when unhealthy.
	NrtHealthyConditionType corev1.NodeConditionType = "MissingNrtHealthyPrerequisite"

	NrtHealthReasonMissing            = "Missing"            // no NRT object for the node
	NrtHealthReasonPolicyNotEnforcing = "PolicyNotEnforcing" // Topology Manager not enforcing alignment
	NrtHealthReasonNoNumaZones        = "NoNumaZones"        // NRT exposes no NUMA-node zones

	NrtAttrTopologyManagerPolicy = "topologyManagerPolicy"
	NrtPolicyNone                = "none"
	NrtZoneTypeNode              = "Node" // NRT Zone.Type of a NUMA node

	nrtCRDName = "noderesourcetopologies.topology.node.k8s.io"

	// NumaPluginName is the scheduler plugin key that enables NUMA-aware scheduling for a
	// nodepool, set under NodePool.Spec.SchedulingShardConfig.Plugins.
	NumaPluginName = "numa"

	NrtPrereqMessageTemplate = "The following prerequisites are not met on the node: %s. " +
		"As a result, NUMA-aware scheduling may be impacted."
	NrtNodePoolPrereqMessageTemplate = "The following prerequisites are not met on one or more nodes " +
		"in the node pool: %s. As a result, NUMA-aware scheduling may be impacted."

	// Per-reason unmet-prerequisite phrases (the %s in NrtPrereqMessageTemplate). The braces are
	// intentional and kept per product.
	NrtMissingItemNoObject    = "NodeResourceTopology custom resource missing"
	NrtMissingItemNoNumaZones = "NodeResourceTopology custom resource invalid"
	NrtMissingItemPolicy      = "Kubelet Topology Manager Policy misconfigured"
)

// nrtPrereqMessage builds the customer-facing NRT-health message for a missing item.
func nrtPrereqMessage(missingItem string) string {
	return fmt.Sprintf(NrtPrereqMessageTemplate, missingItem)
}

func nrtNodePoolPrereqMessage(missingItems []string) string {
	if len(missingItems) == 0 {
		return ""
	}
	return fmt.Sprintf(NrtNodePoolPrereqMessageTemplate, strings.Join(missingItems, ", "))
}

func nrtNodePoolMissingItemForReason(reason string) string {
	switch reason {
	case NrtHealthReasonMissing:
		return NrtMissingItemNoObject
	case NrtHealthReasonNoNumaZones:
		return NrtMissingItemNoNumaZones
	case NrtHealthReasonPolicyNotEnforcing:
		return NrtMissingItemPolicy
	default:
		return ""
	}
}

func setNodePoolNrtHealthCondition(nodePool *v1alpha1.NodePool, missingItems []string) {
	condition := v1alpha1.NodePoolCondition{
		Type:    v1alpha1.NodePoolMissingNrtHealthyPrerequisite,
		Reason:  v1alpha1.NodePoolMissingNrtHealthyPrerequisiteReason,
		Message: nrtNodePoolPrereqMessage(missingItems),
	}
	condition.SetConditionStatusValue(len(missingItems) > 0)
	nodePool.SetNodePoolCondition(condition)
}

// isNrtCRDInstalled reports whether the NodeResourceTopology CRD is installed and established.
func isNrtCRDInstalled(ctx context.Context, reader client.Reader) bool {
	return utils.IsCRDInstalled(ctx, reader, nrtCRDName)
}

// watchForNrtCRD sets up an informer on CustomResourceDefinition and, once the NodeResourceTopology
// CRD is installed, restarts the process so it comes up watching NRT.
// Used only when the CRD is absent at startup.
func (npc *NodePoolController) watchForNrtCRD(ctx context.Context, mgr ctrl.Manager) error {
	informer, err := mgr.GetCache().GetInformer(ctx, &apiextensionsv1.CustomResourceDefinition{})
	if err != nil {
		return err
	}
	_, err = informer.AddEventHandler(toolscache.FilteringResourceEventHandler{
		FilterFunc: func(obj any) bool {
			crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
			if !ok || crd.Name != nrtCRDName {
				return false
			}
			for _, condition := range crd.Status.Conditions {
				if condition.Type == apiextensionsv1.Established &&
					condition.Status == apiextensionsv1.ConditionTrue {
					return true
				}
			}
			return false
		},
		Handler: toolscache.ResourceEventHandlerFuncs{
			AddFunc: func(any) {
				log.Info().Msg("NodeResourceTopology CRD established; restarting to start watching it...")
				osExit(0)
			},
		},
	})
	return err
}

// isNumaEnabled reports whether NUMA-aware scheduling is enabled for the nodepool.
func isNumaEnabled(nodePool *v1alpha1.NodePool) bool {
	cfg := nodePool.Spec.SchedulingShardConfig
	if cfg == nil {
		return false
	}
	plugin, ok := cfg.Plugins[NumaPluginName]
	if !ok {
		return false
	}
	return plugin.Enabled != nil && *plugin.Enabled
}

// reconcileNrtHealthCondition sets or clears the NRT-health condition on a node according to
// its nodepool's NUMA setting and the node's NRT data.
func (npc *NodePoolController) reconcileNrtHealthCondition(ctx context.Context, node *corev1.Node, nodePool *v1alpha1.NodePool) error {
	if !isNumaEnabled(nodePool) {
		return npc.removeNrtCondition(ctx, node)
	}

	if !npc.nrtEnabled {
		// NRT CRD not installed: we cannot (and must not) read NRT objects, and a NUMA-aware
		// node then has no NRT data at all, so the prerequisite is Missing.
		return npc.ensureNrtCondition(ctx, node, NrtHealthReasonMissing, nrtPrereqMessage(NrtMissingItemNoObject))
	}

	nrt, found, err := npc.getNodeResourceTopology(ctx, node.Name)
	if err != nil {
		return err
	}
	if !found {
		nrt = nil // evaluateNrtHealth treats nil as "missing"
	}

	if healthy, reason, message := evaluateNrtHealth(nrt); !healthy {
		return npc.ensureNrtCondition(ctx, node, reason, message)
	}
	return npc.removeNrtCondition(ctx, node)
}

// evaluateNrtHealth returns the NRT health of a node that is already known to be in a
// NUMA-aware nodepool. nrt is nil when no NRT object exists. The reason distinguishes the
// failure mode; the message shares one template with a per-reason missing item.
func evaluateNrtHealth(nrt *nrtv1alpha2.NodeResourceTopology) (healthy bool, reason, message string) {
	if nrt == nil {
		return false, NrtHealthReasonMissing, nrtPrereqMessage(NrtMissingItemNoObject)
	}

	if !hasNumaNodeZone(nrt) {
		return false, NrtHealthReasonNoNumaZones, nrtPrereqMessage(NrtMissingItemNoNumaZones)
	}

	if !isPolicyEnforcing(nrt) {
		return false, NrtHealthReasonPolicyNotEnforcing, nrtPrereqMessage(NrtMissingItemPolicy)
	}

	return true, "", ""
}

func hasNumaNodeZone(nrt *nrtv1alpha2.NodeResourceTopology) bool {
	for _, zone := range nrt.Zones {
		if zone.Type == NrtZoneTypeNode {
			return true
		}
	}
	return false
}

// isPolicyEnforcing reports whether the node's Topology Manager policy enforces NUMA
// alignment. A "none" policy, or no policy attribute at all, counts as not enforcing.
func isPolicyEnforcing(nrt *nrtv1alpha2.NodeResourceTopology) bool {
	if policy, found := topologyManagerPolicy(nrt); found {
		return policy != NrtPolicyNone
	}
	return false
}

func topologyManagerPolicy(nrt *nrtv1alpha2.NodeResourceTopology) (policy string, found bool) {
	for _, attr := range nrt.Attributes {
		if attr.Name == NrtAttrTopologyManagerPolicy {
			return strings.ToLower(attr.Value), true
		}
	}
	return "", false
}

func (npc *NodePoolController) getTopologyManagerPolicyFromNodeNRT(ctx context.Context,
	node *corev1.Node, nodePool *v1alpha1.NodePool) (string, error) {
	if !isNumaEnabled(nodePool) || !npc.nrtEnabled {
		return "", nil
	}

	nrt, found, err := npc.getNodeResourceTopology(ctx, node.Name)
	if err != nil || !found {
		return "", err
	}

	policy, _ := topologyManagerPolicy(nrt)
	return policy, nil
}

func (npc *NodePoolController) getNodeResourceTopology(ctx context.Context, nodeName string) (*nrtv1alpha2.NodeResourceTopology, bool, error) {
	nrt := &nrtv1alpha2.NodeResourceTopology{}
	if err := npc.Client.Get(ctx, client.ObjectKey{Name: nodeName}, nrt); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, false, nil
		}
		log.Error().Msgf("Failed to get NodeResourceTopology for node <%s>, err: %v", nodeName, err.Error())
		return nil, false, err
	}
	return nrt, true, nil
}

// ensureNrtCondition sets the NRT-health condition (status True). It is idempotent and
// patches only our condition type, leaving other node conditions untouched.
func (npc *NodePoolController) ensureNrtCondition(ctx context.Context, node *corev1.Node, reason, message string) error {
	existing := findNodeCondition(node, NrtHealthyConditionType)
	if existing != nil && existing.Status == corev1.ConditionTrue &&
		existing.Reason == reason && existing.Message == message {
		return nil
	}

	now := metav1.Now()
	// the condition only exists with status True; keep the original transition time if so
	lastTransition := now
	if existing != nil {
		lastTransition = existing.LastTransitionTime
	}

	condition := corev1.NodeCondition{
		Type:               NrtHealthyConditionType,
		Status:             corev1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		LastHeartbeatTime:  now,
		LastTransitionTime: lastTransition,
	}
	patchBytes, err := nrtConditionUpsertPatch(condition)
	if err != nil {
		return err
	}
	if err = npc.patchNodeStatus(ctx, node, patchBytes); err != nil {
		log.Error().Msgf("Failed to set NRT-health condition <%s> on node <%s>, err: %v", reason, node.Name, err.Error())
		return err
	}
	upsertNodeCondition(node, condition)

	log.Info().Msgf("Set NRT-health condition on node <%s> (reason: %s)", node.Name, reason)
	return nil
}

func (npc *NodePoolController) removeNrtCondition(ctx context.Context, node *corev1.Node) error {
	if findNodeCondition(node, NrtHealthyConditionType) == nil {
		return nil
	}
	patchBytes, err := nrtConditionDeletePatch()
	if err != nil {
		return err
	}
	if err = npc.patchNodeStatus(ctx, node, patchBytes); err != nil {
		log.Error().Msgf("Failed to clear NRT-health condition on node <%s>, err: %v", node.Name, err.Error())
		return err
	}
	deleteNodeCondition(node, NrtHealthyConditionType)

	log.Info().Msgf("Cleared NRT-health condition on node <%s>", node.Name)
	return nil
}

// patchNodeStatus patches the node status subresource, where conditions live (patchObject
// patches metadata/spec via the non-status client and cannot write conditions).
func (npc *NodePoolController) patchNodeStatus(ctx context.Context, node *corev1.Node, patchBytes []byte) error {
	return npc.Client.Status().Patch(ctx, node, client.RawPatch(types.StrategicMergePatchType, patchBytes))
}

// nrtConditionUpsertPatch builds a strategic merge patch that adds/updates our condition
// (merge-keyed by "type").
func nrtConditionUpsertPatch(cond corev1.NodeCondition) ([]byte, error) {
	return json.Marshal(map[string]any{
		"status": map[string]any{
			"conditions": []corev1.NodeCondition{cond},
		},
	})
}

// nrtConditionDeletePatch builds a strategic merge patch that deletes our condition by "type".
func nrtConditionDeletePatch() ([]byte, error) {
	return json.Marshal(map[string]any{
		"status": map[string]any{
			"conditions": []map[string]any{
				{"type": NrtHealthyConditionType, "$patch": "delete"},
			},
		},
	})
}

func findNodeCondition(node *corev1.Node, condType corev1.NodeConditionType) *corev1.NodeCondition {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == condType {
			return &node.Status.Conditions[i]
		}
	}
	return nil
}

func upsertNodeCondition(node *corev1.Node, condition corev1.NodeCondition) {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == condition.Type {
			node.Status.Conditions[i] = condition
			return
		}
	}
	node.Status.Conditions = append(node.Status.Conditions, condition)
}

func deleteNodeCondition(node *corev1.Node, conditionType corev1.NodeConditionType) {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == conditionType {
			node.Status.Conditions = append(node.Status.Conditions[:i], node.Status.Conditions[i+1:]...)
			return
		}
	}
}
