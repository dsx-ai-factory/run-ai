package sampler_actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnvironmentInfo contains cluster environment details
type EnvironmentInfo struct {
	KubernetesVersion      string
	KubernetesDistribution string
	TotalNodes             int
	RealNodeCount          int
	FakeNodeCount          int
	NodepoolCount          int
	NodesPerPool           int
	KwokEnabled            bool
	GpuProducts            []string
	ResourceTier           string
	RunaiResources         *RunaiResourcesInfo
}

// RunaiResourcesInfo contains Run:AI component resource configurations
type RunaiResourcesInfo struct {
	SchedulingServices *ComponentResourceInfo
	SyncServices       *ComponentResourceInfo
	WorkloadServices   *ComponentResourceInfo
	WorkloadController *ControllerConfigInfo
}

// ComponentResourceInfo contains resource requests and limits for a component
type ComponentResourceInfo struct {
	CPURequest    string
	CPULimit      string
	MemoryRequest string
	MemoryLimit   string
}

// ControllerConfigInfo contains workload controller configuration
type ControllerConfigInfo struct {
	MaxConcurrentReconciles int
	Resources               *ComponentResourceInfo
}

// GatherEnvironmentInfo collects cluster environment information
func GatherEnvironmentInfo(ctx context.Context, kubeClient client.Client, restConfig *rest.Config, clusterConfig *scale.ClusterConfig) (map[string]string, error) {
	envInfo := make(map[string]string)

	// Get Kubernetes version
	k8sVersion, err := getKubernetesVersion(restConfig)
	if err != nil {
		fmt.Printf("Warning: failed to get Kubernetes version: %v\n", err)
		k8sVersion = "unknown"
	}
	envInfo["kubernetes_version"] = k8sVersion

	// Detect Kubernetes distribution
	k8sDistribution, err := getKubernetesDistribution(ctx, kubeClient)
	if err != nil {
		fmt.Printf("Warning: failed to detect Kubernetes distribution: %v\n", err)
		k8sDistribution = "unknown"
	}
	envInfo["kubernetes_distribution"] = k8sDistribution

	// Count real and fake nodes from the cluster
	realNodes, fakeNodes, err := countNodesByType(ctx, kubeClient)
	if err != nil {
		fmt.Printf("Warning: failed to count nodes by type: %v\n", err)
	}
	totalNodes := realNodes + fakeNodes
	envInfo["total_nodes"] = fmt.Sprintf("%d", totalNodes)
	envInfo["real_node_count"] = fmt.Sprintf("%d", realNodes)
	envInfo["fake_node_count"] = fmt.Sprintf("%d", fakeNodes)
	envInfo["nodepool_count"] = fmt.Sprintf("%d", clusterConfig.NodepoolCount)
	envInfo["nodes_per_pool"] = fmt.Sprintf("%d", clusterConfig.NodesPerPool)
	envInfo["kwok_enabled"] = fmt.Sprintf("%t", clusterConfig.KwokEnabled)

	// GPU products
	if len(clusterConfig.GpuProduct) > 0 {
		envInfo["gpu_products"] = strings.Join(clusterConfig.GpuProduct, ",")
	} else {
		envInfo["gpu_products"] = ""
	}

	// Resource tier
	envInfo["resource_tier"] = string(clusterConfig.Resources)

	// Get RunAI resource configurations
	runaiResources, err := getRunaiConfigResources(ctx, kubeClient)
	if err != nil {
		fmt.Printf("Warning: failed to get RunAI config resources: %v\n", err)
	} else {
		// Add RunAI resources to envInfo
		if runaiResources.SchedulingServices != nil {
			envInfo["scheduling_services_cpu_request"] = runaiResources.SchedulingServices.CPURequest
			envInfo["scheduling_services_cpu_limit"] = runaiResources.SchedulingServices.CPULimit
			envInfo["scheduling_services_memory_request"] = runaiResources.SchedulingServices.MemoryRequest
			envInfo["scheduling_services_memory_limit"] = runaiResources.SchedulingServices.MemoryLimit
		}
		if runaiResources.SyncServices != nil {
			envInfo["sync_services_cpu_request"] = runaiResources.SyncServices.CPURequest
			envInfo["sync_services_cpu_limit"] = runaiResources.SyncServices.CPULimit
			envInfo["sync_services_memory_request"] = runaiResources.SyncServices.MemoryRequest
			envInfo["sync_services_memory_limit"] = runaiResources.SyncServices.MemoryLimit
		}
		if runaiResources.WorkloadServices != nil {
			envInfo["workload_services_cpu_request"] = runaiResources.WorkloadServices.CPURequest
			envInfo["workload_services_cpu_limit"] = runaiResources.WorkloadServices.CPULimit
			envInfo["workload_services_memory_request"] = runaiResources.WorkloadServices.MemoryRequest
			envInfo["workload_services_memory_limit"] = runaiResources.WorkloadServices.MemoryLimit
		}
		if runaiResources.WorkloadController != nil {
			envInfo["workload_controller_max_concurrent_reconciles"] = fmt.Sprintf("%d", runaiResources.WorkloadController.MaxConcurrentReconciles)
			if runaiResources.WorkloadController.Resources != nil {
				envInfo["workload_exporter_cpu_request"] = runaiResources.WorkloadController.Resources.CPURequest
				envInfo["workload_exporter_cpu_limit"] = runaiResources.WorkloadController.Resources.CPULimit
				envInfo["workload_exporter_memory_request"] = runaiResources.WorkloadController.Resources.MemoryRequest
				envInfo["workload_exporter_memory_limit"] = runaiResources.WorkloadController.Resources.MemoryLimit
			}
		}
	}

	return envInfo, nil
}

// getKubernetesVersion queries the Kubernetes server version
func getKubernetesVersion(restConfig *rest.Config) (string, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return "", fmt.Errorf("failed to create discovery client: %w", err)
	}

	version, err := discoveryClient.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get server version: %w", err)
	}

	return version.GitVersion, nil
}

// getKubernetesDistribution detects the Kubernetes distribution
func getKubernetesDistribution(ctx context.Context, kubeClient client.Client) (string, error) {
	// Get a sample node to check labels
	nodeList := &corev1.NodeList{}
	err := kubeClient.List(ctx, nodeList, &client.ListOptions{Limit: 1})
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	if len(nodeList.Items) == 0 {
		return "unknown", nil
	}

	node := nodeList.Items[0]

	// Check node labels for distribution indicators
	for labelKey := range node.Labels {
		if strings.Contains(labelKey, "eks.amazonaws.com") {
			return "EKS", nil
		}
		if strings.Contains(labelKey, "cloud.google.com") || strings.Contains(labelKey, "gke.io") {
			return "GKE", nil
		}
		if strings.Contains(labelKey, "node.openshift.io") {
			return "OpenShift", nil
		}
		if strings.Contains(labelKey, "aks.azure.com") {
			return "AKS", nil
		}
	}

	// Check provider ID
	if node.Spec.ProviderID != "" {
		providerID := strings.ToLower(node.Spec.ProviderID)
		if strings.Contains(providerID, "aws") {
			return "EKS", nil
		}
		if strings.Contains(providerID, "gce") || strings.Contains(providerID, "gke") {
			return "GKE", nil
		}
		if strings.Contains(providerID, "azure") {
			return "AKS", nil
		}
	}

	return "Kubernetes", nil
}

// countNodesByType counts real nodes (without fake-node label) and fake nodes (with fake-node=true label)
func countNodesByType(ctx context.Context, kubeClient client.Client) (realNodes int, fakeNodes int, err error) {
	nodeList := &corev1.NodeList{}
	if err := kubeClient.List(ctx, nodeList); err != nil {
		return 0, 0, fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodeList.Items {
		if val, exists := node.Labels["fake-node"]; exists && val == "true" {
			fakeNodes++
		} else {
			realNodes++
		}
	}

	return realNodes, fakeNodes, nil
}

// getRunaiConfigResources reads the RunaiConfig and extracts resource configurations
func getRunaiConfigResources(ctx context.Context, kubeClient client.Client) (*RunaiResourcesInfo, error) {
	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      runaiv1.DefaultName,
		},
	}

	err := kubeClient.Get(ctx, client.ObjectKeyFromObject(runaiConfig), runaiConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get runai config: %w", err)
	}

	resourceInfo := &RunaiResourcesInfo{}

	// Extract scheduling services resources
	if runaiConfig.Spec.Global != nil && runaiConfig.Spec.Global.SchedulingServices != nil && runaiConfig.Spec.Global.SchedulingServices.Resources != nil {
		resourceInfo.SchedulingServices = &ComponentResourceInfo{
			CPURequest:    getResourceString(runaiConfig.Spec.Global.SchedulingServices.Resources.Requests, corev1.ResourceCPU),
			CPULimit:      getResourceString(runaiConfig.Spec.Global.SchedulingServices.Resources.Limits, corev1.ResourceCPU),
			MemoryRequest: getResourceString(runaiConfig.Spec.Global.SchedulingServices.Resources.Requests, corev1.ResourceMemory),
			MemoryLimit:   getResourceString(runaiConfig.Spec.Global.SchedulingServices.Resources.Limits, corev1.ResourceMemory),
		}
	}

	// Extract sync services resources
	if runaiConfig.Spec.Global != nil && runaiConfig.Spec.Global.SyncServices != nil && runaiConfig.Spec.Global.SyncServices.Resources != nil {
		resourceInfo.SyncServices = &ComponentResourceInfo{
			CPURequest:    getResourceString(runaiConfig.Spec.Global.SyncServices.Resources.Requests, corev1.ResourceCPU),
			CPULimit:      getResourceString(runaiConfig.Spec.Global.SyncServices.Resources.Limits, corev1.ResourceCPU),
			MemoryRequest: getResourceString(runaiConfig.Spec.Global.SyncServices.Resources.Requests, corev1.ResourceMemory),
			MemoryLimit:   getResourceString(runaiConfig.Spec.Global.SyncServices.Resources.Limits, corev1.ResourceMemory),
		}
	}

	// Extract workload services resources
	if runaiConfig.Spec.Global != nil && runaiConfig.Spec.Global.WorkloadServices != nil && runaiConfig.Spec.Global.WorkloadServices.Resources != nil {
		resourceInfo.WorkloadServices = &ComponentResourceInfo{
			CPURequest:    getResourceString(runaiConfig.Spec.Global.WorkloadServices.Resources.Requests, corev1.ResourceCPU),
			CPULimit:      getResourceString(runaiConfig.Spec.Global.WorkloadServices.Resources.Limits, corev1.ResourceCPU),
			MemoryRequest: getResourceString(runaiConfig.Spec.Global.WorkloadServices.Resources.Requests, corev1.ResourceMemory),
			MemoryLimit:   getResourceString(runaiConfig.Spec.Global.WorkloadServices.Resources.Limits, corev1.ResourceMemory),
		}
	}

	// Extract workload controller config
	if runaiConfig.Spec.WorkloadController != nil || runaiConfig.Spec.WorkloadExporter != nil {
		controllerInfo := &ControllerConfigInfo{}

		if runaiConfig.Spec.WorkloadController != nil && runaiConfig.Spec.WorkloadController.TrainingMaxConcurrentReconciles != nil {
			controllerInfo.MaxConcurrentReconciles = *runaiConfig.Spec.WorkloadController.TrainingMaxConcurrentReconciles
		}

		if runaiConfig.Spec.WorkloadExporter != nil && runaiConfig.Spec.WorkloadExporter.Resources != nil {
			controllerInfo.Resources = &ComponentResourceInfo{
				CPURequest:    getResourceString(runaiConfig.Spec.WorkloadExporter.Resources.Requests, corev1.ResourceCPU),
				CPULimit:      getResourceString(runaiConfig.Spec.WorkloadExporter.Resources.Limits, corev1.ResourceCPU),
				MemoryRequest: getResourceString(runaiConfig.Spec.WorkloadExporter.Resources.Requests, corev1.ResourceMemory),
				MemoryLimit:   getResourceString(runaiConfig.Spec.WorkloadExporter.Resources.Limits, corev1.ResourceMemory),
			}
		}

		resourceInfo.WorkloadController = controllerInfo
	}

	return resourceInfo, nil
}

// getResourceString extracts resource quantity as string
func getResourceString(resourceList corev1.ResourceList, resourceName corev1.ResourceName) string {
	if resourceList == nil {
		return ""
	}

	quantity, ok := resourceList[resourceName]
	if !ok {
		return ""
	}

	return quantity.String()
}

