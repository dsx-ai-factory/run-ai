package seed_cluster

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NodeChaosTaintKey   = "chaos-test/node-failure"
	NodeChaosTaintValue = "true"
)

type NodeChaosMetrics struct {
	Enabled                bool          `json:"enabled"`
	Completed              bool          `json:"completed"`
	Error                  string        `json:"error,omitempty"`
	ChaosStartTime         time.Time     `json:"chaos_start_time,omitempty"`
	AllNodesDownTime       time.Time     `json:"all_nodes_down_time,omitempty"`
	NodesRestoredTime      time.Time     `json:"nodes_restored_time,omitempty"`
	FailedNodes            []string      `json:"failed_nodes,omitempty"`
	ConfiguredDownDuration time.Duration `json:"configured_down_duration,omitempty"`
	ActualDownDuration     time.Duration `json:"actual_down_duration,omitempty"`
}

func (w *WorkloadSampler) startNodeChaos(ctx context.Context) {
	if !w.config.NodeChaosEnabled {
		return
	}

	w.nodeChaosMetrics = &NodeChaosMetrics{
		Enabled: true,
	}

	go func() {
		fmt.Printf("[NodeChaos] Starting chaos routine - will wait %s before starting\n", w.config.NodeChaosStartDelay)
		fmt.Printf("[NodeChaos] Config: count=%d, downDuration=%s, delayBetween=%s\n",
			w.config.NodeChaosCount, w.config.NodeChaosDownDuration, w.config.NodeChaosDelayBetween)

		time.Sleep(w.config.NodeChaosStartDelay)

		nodes, err := w.selectNodesToFail(ctx)
		if err != nil {
			fmt.Printf("[NodeChaos] ERROR: Failed to select nodes: %v\n", err)
			w.nodeChaosMetrics.Error = err.Error()
			return
		}

		if len(nodes) == 0 {
			fmt.Println("[NodeChaos] No nodes selected for chaos, skipping")
			w.nodeChaosMetrics.Error = "no nodes selected"
			return
		}

		w.nodeChaosMetrics.ChaosStartTime = time.Now()
		w.nodeChaosMetrics.FailedNodes = nodes
		w.nodeChaosMetrics.ConfiguredDownDuration = w.config.NodeChaosDownDuration

		fmt.Printf("[NodeChaos] Bringing down %d nodes: %v\n", len(nodes), nodes)
		for i, nodeName := range nodes {
			if err := w.bringNodeDown(ctx, nodeName); err != nil {
				fmt.Printf("[NodeChaos] ERROR: Failed to bring down node %s: %v\n", nodeName, err)
			}

			if i < len(nodes)-1 {
				time.Sleep(w.config.NodeChaosDelayBetween)
			}
		}
		w.nodeChaosMetrics.AllNodesDownTime = time.Now()
		fmt.Printf("[NodeChaos] All %d nodes are now down, waiting %s\n", len(nodes), w.config.NodeChaosDownDuration)

		time.Sleep(w.config.NodeChaosDownDuration)

		fmt.Printf("[NodeChaos] Restoring %d nodes\n", len(nodes))
		for _, nodeName := range nodes {
			if err := w.bringNodeUp(ctx, nodeName); err != nil {
				fmt.Printf("[NodeChaos] ERROR: Failed to restore node %s: %v\n", nodeName, err)
			}
		}
		w.nodeChaosMetrics.NodesRestoredTime = time.Now()
		w.nodeChaosMetrics.ActualDownDuration = w.nodeChaosMetrics.NodesRestoredTime.Sub(w.nodeChaosMetrics.AllNodesDownTime)
		w.nodeChaosMetrics.Completed = true

		fmt.Printf("[NodeChaos] Chaos complete. Nodes were down for %s\n", w.nodeChaosMetrics.ActualDownDuration)
		fmt.Printf("[NodeChaos] Summary: started=%s, allDown=%s, restored=%s\n",
			w.nodeChaosMetrics.ChaosStartTime.Format(time.RFC3339),
			w.nodeChaosMetrics.AllNodesDownTime.Format(time.RFC3339),
			w.nodeChaosMetrics.NodesRestoredTime.Format(time.RFC3339))
	}()
}

func (w *WorkloadSampler) bringNodeDown(ctx context.Context, nodeName string) error {
	node := &corev1.Node{}
	if err := w.clients.KubeClient.Get(ctx, client.ObjectKey{Name: nodeName}, node); err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	node.Spec.Unschedulable = true

	taintExists := false
	for _, taint := range node.Spec.Taints {
		if taint.Key == NodeChaosTaintKey {
			taintExists = true
			break
		}
	}

	if !taintExists {
		node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
			Key:       NodeChaosTaintKey,
			Value:     NodeChaosTaintValue,
			Effect:    corev1.TaintEffectNoExecute,
			TimeAdded: ptr.To(metav1.Now()),
		})
	}

	if err := w.clients.KubeClient.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to update node %s: %w", nodeName, err)
	}

	fmt.Printf("Node %s marked as down (cordoned + tainted with NoExecute)\n", nodeName)
	return nil
}

func (w *WorkloadSampler) bringNodeUp(ctx context.Context, nodeName string) error {
	node := &corev1.Node{}
	if err := w.clients.KubeClient.Get(ctx, client.ObjectKey{Name: nodeName}, node); err != nil {
		return fmt.Errorf("failed to get node %s: %w", nodeName, err)
	}

	node.Spec.Unschedulable = false

	newTaints := []corev1.Taint{}
	for _, taint := range node.Spec.Taints {
		if taint.Key != NodeChaosTaintKey {
			newTaints = append(newTaints, taint)
		}
	}
	node.Spec.Taints = newTaints

	if err := w.clients.KubeClient.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to update node %s: %w", nodeName, err)
	}

	fmt.Printf("Node %s restored (uncordoned + taint removed)\n", nodeName)
	return nil
}

func (w *WorkloadSampler) selectNodesToFail(ctx context.Context) ([]string, error) {
	nodeList := &corev1.NodeList{}
	listOpts := []client.ListOption{}

	if w.config.NodeChaosNodeSelector != "" {
		selector, err := metav1.ParseToLabelSelector(w.config.NodeChaosNodeSelector)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node selector: %w", err)
		}
		labelSelector, err := metav1.LabelSelectorAsSelector(selector)
		if err != nil {
			return nil, fmt.Errorf("failed to convert label selector: %w", err)
		}
		listOpts = append(listOpts, client.MatchingLabelsSelector{Selector: labelSelector})
	}

	if err := w.clients.KubeClient.List(ctx, nodeList, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	eligibleNodes := []string{}
	for _, node := range nodeList.Items {
		if isControlPlaneNode(&node) {
			continue
		}
		eligibleNodes = append(eligibleNodes, node.Name)
	}

	if len(eligibleNodes) < w.config.NodeChaosCount {
		return nil, fmt.Errorf("not enough eligible nodes: need %d, found %d", w.config.NodeChaosCount, len(eligibleNodes))
	}

	selectedNodes := eligibleNodes[:w.config.NodeChaosCount]

	fmt.Printf("Selected %d nodes for chaos: %v\n", len(selectedNodes), selectedNodes)
	return selectedNodes, nil
}

func isControlPlaneNode(node *corev1.Node) bool {
	controlPlaneLabels := []string{
		"node-role.kubernetes.io/control-plane",
		"node-role.kubernetes.io/master",
	}

	for _, label := range controlPlaneLabels {
		if _, exists := node.Labels[label]; exists {
			return true
		}
	}
	return false
}

func (w *WorkloadSampler) GetNodeChaosMetrics() *NodeChaosMetrics {
	return w.nodeChaosMetrics
}

func (m *NodeChaosMetrics) ToMetadataMap() map[string]string {
	if m == nil || !m.Enabled {
		return map[string]string{
			"node_chaos_enabled": "false",
		}
	}

	metadata := map[string]string{
		"node_chaos_enabled":   "true",
		"node_chaos_completed": fmt.Sprintf("%t", m.Completed),
	}

	if m.Error != "" {
		metadata["node_chaos_error"] = m.Error
		return metadata
	}

	if m.Completed {
		metadata["node_chaos_start_time"] = m.ChaosStartTime.Format(time.RFC3339)
		metadata["node_chaos_all_nodes_down_time"] = m.AllNodesDownTime.Format(time.RFC3339)
		metadata["node_chaos_nodes_restored_time"] = m.NodesRestoredTime.Format(time.RFC3339)
	}

	return metadata
}
