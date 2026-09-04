/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node_pools

import (
	"context"
	"fmt"
	"strings"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	kwokopv1beta1 "github.com/run-ai/kwok-operator/api/v1beta1"
	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type NodepoolConfig struct {
	NodepoolCount         int
	NodesPerPool          int
	KwokEnabled           bool
	ClusterId             string
	MonitorWaitTime       time.Duration
	NodepoolCreateTimeOut string
	GpuProduct            []string
	GPUsPerNode           int
}
type ClusterNodePools struct {
	config  *NodepoolConfig
	clients *sampler_actions.Clients
}

func NewClusterNodePools(nodepoolConfig *NodepoolConfig, clients *sampler_actions.Clients) *ClusterNodePools {
	return &ClusterNodePools{
		config:  nodepoolConfig,
		clients: clients,
	}
}

func (n *ClusterNodePools) Create(ctx context.Context) error {
	//Allow retry on getting existing nodepools, since cluster might be adjusting after resource configuration
	existingNodePoolsInCluster, err := n.getExistingNodePoolsWithRetry(3)

	numNodePools := n.config.NodepoolCount
	// Only use GpuProduct length when KWOK is disabled AND GPU products are specified
	// Otherwise, use NodepoolCount to create nodepools with generic labels
	if !n.config.KwokEnabled && len(n.config.GpuProduct) > 0 {
		numNodePools = len(n.config.GpuProduct)
	}
	if err != nil {
		return fmt.Errorf("failed to get existing nodepools: %w", err)
	}

	kwokProfileName := "default"
	if n.config.KwokEnabled {
		kwokProfileName, err = n.ensureTopologyProfile(ctx)
		if err != nil {
			return fmt.Errorf("failed to ensure topology profile: %w", err)
		}
	}

	for i := 1; i <= numNodePools; i++ {
		err = n.createRunAINodePool(ctx, i, n.config.GpuProduct, existingNodePoolsInCluster, kwokProfileName)
		if err != nil {
			return fmt.Errorf("failed to create RunAI nodepool: %w", err)
		}
	}
	return nil
}

func (n *ClusterNodePools) Delete(ctx context.Context) error {
	if n.config.KwokEnabled {
		existingNodePoolsInCluster, err := n.getExistingNodePools()
		if err != nil {
			return fmt.Errorf("failed to get existing nodepools: %w", err)
		}
		for _, nodepool := range existingNodePoolsInCluster {
			if nodepool.Name == "default" {
				continue
			}
			err = n.deleteRunaiNodePool(ctx, nodepool)
			if err != nil {
				return fmt.Errorf("failed to delete RunAI nodepool: %v", err)
			}
		}
	}
	return nil
}

func (n *ClusterNodePools) WaitForNodepoolsReadiness() error {
	ticker := time.NewTicker(n.config.MonitorWaitTime)
	startTime := time.Now()
	defer ticker.Stop()
	durationForProjectWait, _ := time.ParseDuration(n.config.NodepoolCreateTimeOut)
	timeout := time.After(durationForProjectWait)

	diagTicker := time.NewTicker(2 * time.Minute)
	defer diagTicker.Stop()

	for {
		select {
		case <-ticker.C:
			ready, err := n.allNodePoolsReady()
			if err != nil {
				return fmt.Errorf("failed to check nodepools readiness: %w", err)
			}
			if ready {
				duration := time.Since(startTime)
				fmt.Printf("All nodepools are ready, after: %v\n", duration)
				return nil
			}
		case <-diagTicker.C:
			fmt.Printf("\n--- Diagnostic info (waiting %v) ---\n", time.Since(startTime))
			n.logNodepoolStatus(context.Background())
			n.logPendingPodsInRunaiNamespace(context.Background())
			fmt.Print("-----------------------------------\n\n")
		case <-timeout:
			fmt.Println("\n=== TIMEOUT DIAGNOSTICS ===")
			n.logNodepoolStatus(context.Background())
			n.logPendingPodsInRunaiNamespace(context.Background())
			fmt.Print("===========================\n\n")
			return fmt.Errorf("nodepool readiness timeout has reached not all nodepools are ready")
		}
	}
}

func ToNodePoolConfig(clusterConfig *scale.ClusterConfig) *NodepoolConfig {
	return &NodepoolConfig{
		NodepoolCount:         clusterConfig.NodepoolCount,
		NodesPerPool:          clusterConfig.NodesPerPool,
		KwokEnabled:           clusterConfig.KwokEnabled,
		ClusterId:             clusterConfig.ClusterID,
		MonitorWaitTime:       time.Duration(2) * time.Second,
		NodepoolCreateTimeOut: "20m",
		GpuProduct:            clusterConfig.GpuProduct,
		GPUsPerNode:           clusterConfig.GPUsPerNode,
	}
}

func (n *ClusterNodePools) allNodePoolsReady() (bool, error) {
	if !n.config.KwokEnabled {
		return true, nil
	}

	nodepools, _, err := n.clients.ClusterServiceClient.NodePoolsAPI.GetNodepools(context.Background()).Execute()
	if err != nil {
		return false, err
	}
	if nodepools == nil {
		return false, fmt.Errorf("failed to get nodepools")
	}
	if len(nodepools.Nodepools) == 0 {
		return false, nil
	}
	for _, nodepool := range nodepools.Nodepools {
		if nodepool.Phase == nil || *nodepool.Phase != clusterclient.NODEPOOLPHASE_READY {
			return false, nil
		}
	}
	return true, nil
}

func (n *ClusterNodePools) logNodepoolStatus(ctx context.Context) {
	nodepools, _, err := n.clients.ClusterServiceClient.NodePoolsAPI.GetNodepools(ctx).Execute()
	if err != nil {
		fmt.Printf("Warning: failed to get nodepools from Control Plane: %v\n", err)
		return
	}

	fmt.Println("Nodepool status (Control Plane):")
	for _, np := range nodepools.Nodepools {
		phase := "Unknown"
		if np.Phase != nil {
			phase = string(*np.Phase)
		}
		fmt.Printf("  - %s: %s\n", np.Name, phase)
	}
}

func (n *ClusterNodePools) logPendingPodsInRunaiNamespace(ctx context.Context) {
	podList := &corev1.PodList{}
	if err := n.clients.KubeClient.List(ctx, podList, client.InNamespace("runai")); err != nil {
		fmt.Printf("Warning: failed to list pods in runai namespace: %v\n", err)
		return
	}

	var pendingPods []string
	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodPending {
			reason := "Unknown"
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
					reason = condition.Message
					if len(reason) > 100 {
						reason = reason[:100] + "..."
					}
					break
				}
			}
			pendingPods = append(pendingPods, fmt.Sprintf("  - %s: %s", pod.Name, reason))
		}
	}

	if len(pendingPods) > 0 {
		fmt.Printf("⚠️  %d pending pods in runai namespace:\n%s\n", len(pendingPods), strings.Join(pendingPods, "\n"))
	} else {
		fmt.Println("✓ No pending pods in runai namespace")
	}
}

func (n *ClusterNodePools) createKwokNodePool(ctx context.Context, nodepool string, nodesPerPool int, profileName string) error {
	kwokNp := kwokopv1beta1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodepool,
		},
		Spec: kwokopv1beta1.NodePoolSpec{
			NodeCount: int32(nodesPerPool),
			NodeTemplate: corev1.Node{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"run.ai/simulated-gpu-node-pool": profileName,
						"type":                           "kwok",
						"fake-node":                      "true",
						"nodepool":                       nodepool,
					},
				},
			},
		},
	}

	err := n.clients.KubeClient.Create(ctx, &kwokNp)
	if k8serrors.IsAlreadyExists(err) {
		existing := &kwokopv1beta1.NodePool{}
		if err := n.clients.KubeClient.Get(ctx, client.ObjectKey{Name: nodepool}, existing); err != nil {
			return fmt.Errorf("failed to get existing kwok nodepool %s: %w", nodepool, err)
		}
		existing.Spec = kwokNp.Spec
		if err := n.clients.KubeClient.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update kwok nodepool %s: %w", nodepool, err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to apply kwok %s: %w", nodepool, err)
	}
	return nil
}

func (n *ClusterNodePools) deleteKwokNodePool(ctx context.Context, nodepool string) error {
	kwokNp := kwokopv1beta1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodepool,
		},
	}
	err := n.clients.KubeClient.Delete(ctx, &kwokNp)
	if err != nil {
		return fmt.Errorf("failed to delete kwok np %s: %w", nodepool, err)
	}
	return nil
}

func (n *ClusterNodePools) createRunAINodePool(ctx context.Context, i int, GpuProduct []string, existingNodePools map[string]*clusterclient.Nodepool, kwokProfileName string) error {
	ps, err := clusterclient.NewPlacementStrategyFromValue("binpack")
	if err != nil {
		return fmt.Errorf("failed to create RunAI placement strategy: %w", err)
	}
	name := fmt.Sprintf("nodepool-%d", i)
	labelKey, LabelValue := "nodepool", name
	if !n.config.KwokEnabled && len(GpuProduct) > 0 {
		labelKey = "nvidia.com/gpu.product"
		LabelValue = GpuProduct[(i-1)%len(GpuProduct)]
	}
	np, resp, _ := n.clients.ClusterServiceClient.NodePoolsAPI.CreateNodepool(ctx).NodepoolCreateFields(clusterclient.NodepoolCreateFields{
		Name:       name,
		LabelKey:   labelKey,
		LabelValue: LabelValue,
		ClusterId:  n.config.ClusterId,
		SchedulingConfiguration: *clusterclient.NewNullableSchedulingConfiguration(&clusterclient.SchedulingConfiguration{
			GpuPodsPlacementStrategy: *clusterclient.NewNullableGpuPodsPlacementStrategy(&clusterclient.GpuPodsPlacementStrategy{
				Node: *clusterclient.NewNullablePlacementStrategy(ps),
			}),
			CpuPodsPlacementStrategy: *clusterclient.NewNullableCpuPodsPlacementStrategy(&clusterclient.CpuPodsPlacementStrategy{
				Node: *clusterclient.NewNullablePlacementStrategy(ps),
			}),
		}),
	}).Execute()

	// Not checking error here since we can get an error even if the Status Code is 201
	// due to response body parsing error which is not relevant for us
	if resp.StatusCode != 201 && resp.StatusCode != 409 {
		return fmt.Errorf("failed to create RunAI nodepool-%d: got status code %d", i, resp.StatusCode)
	}
	if np == nil {
		np = existingNodePools[name]
	}
	if n.config.KwokEnabled {
		if err := n.reconcileNodePool(np, kwokProfileName); err != nil {
			return fmt.Errorf("failed to reconcile kwok nodepool: %w", err)
		}
	}
	return nil
}

func (n *ClusterNodePools) reconcileNodePool(np *clusterclient.Nodepool, profileName string) error {
	err := n.createKwokNodePool(context.Background(), np.Name, n.config.NodesPerPool, profileName)
	if err != nil {
		return fmt.Errorf("failed to reconcile kwok nodepool: %w", err)
	}
	return nil
}

// ensureTopologyProfile adds a GPU profile to the fake-GPU operator topology
// ConfigMap so KWOK nodes with that profile label get the correct GPU count.
// Returns the profile name to use on KWOK node templates.
func (n *ClusterNodePools) ensureTopologyProfile(ctx context.Context) (string, error) {
	gpuCount := n.config.GPUsPerNode
	if gpuCount <= 0 {
		gpuCount = 8
	}
	profileName := fmt.Sprintf("runai-fake-%dgpu", gpuCount)

	cm := &corev1.ConfigMap{}
	if err := n.clients.KubeClient.Get(ctx, client.ObjectKey{Namespace: "gpu-operator", Name: "topology"}, cm); err != nil {
		return "", fmt.Errorf("failed to get topology configmap: %w", err)
	}

	var topology map[string]interface{}
	if err := yaml.Unmarshal([]byte(cm.Data["topology.yml"]), &topology); err != nil {
		return "", fmt.Errorf("failed to parse topology configmap: %w", err)
	}

	nodePools, _ := topology["nodePools"].(map[string]interface{})
	if nodePools == nil {
		nodePools = make(map[string]interface{})
		topology["nodePools"] = nodePools
	}

	if _, exists := nodePools[profileName]; !exists {
		nodePools[profileName] = map[string]interface{}{
			"gpuCount":   gpuCount,
			"gpuMemory":  80000,
			"gpuProduct": fmt.Sprintf("runai-fake-%dgpu", gpuCount),
		}
		updated, err := yaml.Marshal(topology)
		if err != nil {
			return "", fmt.Errorf("failed to marshal updated topology: %w", err)
		}
		cm.Data["topology.yml"] = string(updated)
		if err := n.clients.KubeClient.Update(ctx, cm); err != nil {
			return "", fmt.Errorf("failed to update topology configmap: %w", err)
		}
		fmt.Printf("Added topology profile %q with gpuCount=%d gpuMemory=80000\n", profileName, gpuCount)

		// status-updater caches the topology ConfigMap at startup; restart it so it
		// picks up the new profile before KWOK nodes are created.
		if err := n.restartDeployment(ctx, "gpu-operator", "status-updater"); err != nil {
			return "", fmt.Errorf("failed to restart status-updater: %w", err)
		}
		fmt.Println("Restarted status-updater to reload topology profile")
		time.Sleep(5 * time.Second)
	}

	return profileName, nil
}

// restartDeployment triggers a rollout restart by patching the pod template annotation.
func (n *ClusterNodePools) restartDeployment(ctx context.Context, namespace, name string) error {
	patch := []byte(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"` + time.Now().UTC().Format(time.RFC3339) + `"}}}}}`)
	dep := &appsv1.Deployment{}
	if err := n.clients.KubeClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, dep); err != nil {
		return err
	}
	return n.clients.KubeClient.Patch(ctx, dep, client.RawPatch(types.MergePatchType, patch))
}

func (n *ClusterNodePools) getExistingNodePoolsWithRetry(maxAttempts int) (map[string]*clusterclient.Nodepool, error) {
	var (
		nodePools map[string]*clusterclient.Nodepool
		err       error
	)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		nodePools, err = n.getExistingNodePools()
		if err == nil {
			return nodePools, nil
		}

		fmt.Printf("Attempt %d: failed to get node pools: %v\n", attempt, err)

		if attempt < maxAttempts {
			time.Sleep(10 * time.Second)
		}
	}
	return nil, fmt.Errorf("all %d attempts to get node pools failed: %w", maxAttempts, err)
}

func (n *ClusterNodePools) getExistingNodePools() (map[string]*clusterclient.Nodepool, error) {
	nodepoolMap := map[string]*clusterclient.Nodepool{}
	request := n.clients.ClusterServiceClient.NodePoolsAPI.GetNodepools(context.Background())
	request = request.FilterBy([]string{"clusterId==" + n.config.ClusterId})
	nodepools, _, err := request.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodepools: %v", err)
	}
	for _, nodepool := range nodepools.Nodepools {
		nodepoolMap[nodepool.Name] = &nodepool
	}
	return nodepoolMap, nil
}

func (n *ClusterNodePools) deleteRunaiNodePool(ctx context.Context, nodepool *clusterclient.Nodepool) error {
	_, _, err := n.clients.ClusterServiceClient.NodePoolsAPI.DeleteNodepool(ctx, nodepool.Id).Execute()
	if err != nil {
		return fmt.Errorf("failed to delete RunAI nodepool %s: %w", nodepool.Name, err)
	}
	if n.config.KwokEnabled {
		err = n.deleteKwokNodePool(ctx, nodepool.Name)
		if err != nil {
			return fmt.Errorf("failed to delete kwok nodepool %s: %w", nodepool.Name, err)
		}
	}
	return nil
}

func getNodePoolNumber(str string) int {
	var number int
	_, err := fmt.Sscanf(str, "nodepool-%d", &number)
	if err != nil {
		return -1
	}
	return number
}
