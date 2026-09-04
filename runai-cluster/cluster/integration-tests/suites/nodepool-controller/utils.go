/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"
	"fmt"

	. "github.com/onsi/gomega"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/integration-tests/common"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1global "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodePool struct {
	LabelKey   string
	LabelValue string
	IsDefault  bool
}

func buildNodePoolObject(nodePoolName, labelKey, labelValue, cpuStrategy, gpuStrategy string) *v1alpha1.NodePool {
	return &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodePoolName,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   labelKey,
			LabelValue: labelValue,
			SchedulingShardConfig: &v1alpha1.SchedulingShardConfig{
				PlacementStrategy: &kaiv1.PlacementStrategy{
					CPU: &cpuStrategy,
					GPU: &gpuStrategy,
				},
			},
		},
	}
}

func buildAffinityFromNodePools(nodePools []NodePool) *v1.Affinity {
	var nodeSelectorTerms []v1.NodeSelectorTerm

	for _, nodePool := range nodePools {
		if nodePool.IsDefault {
			nodeSelectorTerms = append(nodeSelectorTerms, v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      "runai/node-pool",
						Operator: v1.NodeSelectorOpDoesNotExist,
					},
				},
			})
		} else {
			nodeSelectorTerms = append(nodeSelectorTerms, v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					{
						Key:      nodePool.LabelKey,
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{nodePool.LabelValue},
					},
				},
			})
		}
	}

	return &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: nodeSelectorTerms,
			},
		},
	}
}

func getWorkerNodeName(ctx context.Context, kubeClient client.Client) (string, error) {
	var nodeList v1.NodeList
	err := kubeClient.List(ctx, &nodeList)
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodeList.Items {
		// Check if the node does NOT have the control-plane label
		if _, exists := node.Labels["node-role.kubernetes.io/control-plane"]; !exists {
			return node.Name, nil
		}
	}

	return "", fmt.Errorf("no worker nodes found")
}

func labelNode(ctx context.Context, kubeClient client.Client, nodeName, key, value string) error {
	var node v1.Node
	err := kubeClient.Get(ctx, client.ObjectKey{Name: nodeName}, &node)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels[key] = value

	err = kubeClient.Update(ctx, &node)
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

func removeNodeLabel(ctx context.Context, kubeClient client.Client, nodeName, key string) error {
	var node v1.Node
	err := kubeClient.Get(ctx, client.ObjectKey{Name: nodeName}, &node)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	if node.Labels != nil {
		delete(node.Labels, key)
	}

	err = kubeClient.Update(ctx, &node)
	if err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	return nil
}

func validateNodeAndNodePoolCoupling(ctx context.Context, kubeClient client.Client, nodePoolName, nodeName string) error {
	var nodePool v1alpha1.NodePool
	err := kubeClient.Get(ctx, client.ObjectKey{Name: nodePoolName}, &nodePool)
	if err != nil {
		return fmt.Errorf("failed to get node pool: %w", err)
	}

	if len(nodePool.Status.Nodes) != 1 {
		return fmt.Errorf("node pool %s does not have the correct number of nodes", nodePoolName)
	}

	nodeInNodePool := nodePool.Status.Nodes[0]
	if nodeInNodePool.Name != nodeName {
		return fmt.Errorf("node %s is not part of node pool %s", nodeName, nodePoolName)
	}

	var node v1.Node
	err = kubeClient.Get(ctx, client.ObjectKey{Name: nodeName}, &node)
	if err != nil {
		return fmt.Errorf("failed to get node: %w", err)
	}

	if _, exists := node.Labels[nodePool.Spec.LabelKey]; !exists {
		return fmt.Errorf("node %s does not have the correct label", nodeName)
	}

	if node.Labels[nodePool.Spec.LabelKey] != nodePool.Spec.LabelValue {
		return fmt.Errorf("node %s does not have the correct label", nodeName)
	}

	if node.Labels["runai/node-pool"] != nodePoolName {
		return fmt.Errorf("node %s does not have the correct node pool annotation", nodeName)
	}

	return nil
}

func submitRunaiJobWithPriorityAndNodeAffinity(ctx context.Context, kubeClient client.Client, nodePoolList []NodePool, runaiJobName, projectName, priorityClassName string) error {

	namespace := "runai-" + projectName

	runaiJob := &runaiv1.RunaiJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      runaiJobName,
			Namespace: namespace,
			Labels: map[string]string{
				"priorityClassName": priorityClassName,
				"workloadKind":      "TrainingWorkload",
				"workloadName":      runaiJobName,
			},
		},
		Spec: runaiv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					SchedulerName: "runai-scheduler",
					Containers: []v1.Container{
						{
							Name:    runaiJobName,
							Image:   "ubuntu",
							Command: []string{"sleep", "infinity"},
							Env: []v1.EnvVar{
								{
									Name:  "RUNAI_JOB_NAME",
									Value: runaiJobName,
								},
								{
									Name:  "RUNAI_PROJECT",
									Value: projectName,
								},
							},
						},
					},
					Resources: &v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"nvidia.com/gpu": resource.MustParse("1"),
						},
						Requests: v1.ResourceList{
							"cpu": resource.MustParse("300m"),
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					Affinity:      buildAffinityFromNodePools(nodePoolList),
				},
			},
		},
	}

	err := kubeClient.Create(ctx, runaiJob)
	if err != nil {
		return fmt.Errorf("failed to create runaiJob: %w", err)
	}

	return nil

}

func editRunaiConfigAndExcludeNodesByLabelFromRunai(ctx context.Context, kubeClient client.Client, labelKey string) error {
	managedNodes := &runaiv1global.ManagedNodes{
		InclusionCriteria: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      labelKey,
							Operator: v1.NodeSelectorOpDoesNotExist,
						},
					},
				},
			},
		},
	}

	var runaiConfig runaiv1.RunaiConfig

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      "runai",
			Namespace: "runai"},
			&runaiConfig)
		if err != nil {
			return fmt.Errorf("failed to get runaiConfig: %w", err)
		}

		runaiConfig.Spec.Global.ManagedNodes = managedNodes

		err = kubeClient.Update(ctx, &runaiConfig)
		if err != nil {
			return fmt.Errorf("failed to update runaiConfig: %w", err)
		}

		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("failed to exclude nodes from Run:AI: %w", retryErr)
	}

	return nil
}

func editRunaiConfigAndRemoveNodeExclusion(ctx context.Context, kubeClient client.Client) error {
	var runaiConfig runaiv1.RunaiConfig
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		err := kubeClient.Get(ctx, types.NamespacedName{
			Name:      "runai",
			Namespace: "runai"},
			&runaiConfig)
		if err != nil {
			return fmt.Errorf("failed to get runaiConfig: %w", err)
		}

		runaiConfig.Spec.Global.ManagedNodes = nil

		err = kubeClient.Update(ctx, &runaiConfig)
		if err != nil {
			return fmt.Errorf("failed to update runaiConfig: %w", err)
		}
		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("failed to remove node exclusion from Run:AI: %w", retryErr)
	}

	return nil
}

func submitPod(ctx context.Context, kubeClient client.Client, projectName string) (*v1.Pod, error) {
	namespace := "runai-" + projectName
	podName := common.GetRandomName("test-pod")
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
		Spec: common.CreatePodSpec(podName, "", "", "", "", "ubuntu"),
	}

	err := kubeClient.Create(ctx, pod)
	if err != nil {
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	return pod, nil
}

func eventuallyValidateNodeLabel(ctx context.Context, kubeClient client.Client, nodeName, labelKey, expectedLabelValue string, shouldExist bool) {
	node := &v1.Node{}
	EventuallyWithOffset(1, func() bool {
		err := kubeClient.Get(ctx, client.ObjectKey{Name: nodeName}, node)
		if err != nil {
			return false
		}
		labelValue, exists := node.Labels[labelKey]
		if shouldExist {
			return exists && labelValue == expectedLabelValue
		}
		return !exists
	}, "30s", "2s").Should(BeTrue())
}

// NodePoolCreationConfig holds configuration for creating a nodepool with full integration
type NodePoolCreationConfig struct {
	NodePoolName    string
	LabelKey        string
	LabelValue      string
	CpuStrategy     *string
	GpuStrategy     *string
	DepartmentNames []string
	ProjectNames    []string
	GpuDeserved     *int64
	CpuDeserved     *int64
	MemoryDeserved  *int64
}

// CreateNodePoolWithFullIntegration creates a nodepool and sets up all necessary department/project resources
func CreateNodePoolWithFullIntegration(ctx context.Context, kubeClient client.Client, config NodePoolCreationConfig) error {
	// Set default values if not provided
	gpuDeserved := int64(1)
	cpuDeserved := int64(1000)
	memoryDeserved := int64(10000)
	cpuStrategy := "binpack"
	gpuStrategy := "binpack"

	if config.GpuDeserved != nil {
		gpuDeserved = *config.GpuDeserved
	}
	if config.CpuDeserved != nil {
		cpuDeserved = *config.CpuDeserved
	}
	if config.MemoryDeserved != nil {
		memoryDeserved = *config.MemoryDeserved
	}
	if config.CpuStrategy != nil {
		cpuStrategy = *config.CpuStrategy
	}
	if config.GpuStrategy != nil {
		gpuStrategy = *config.GpuStrategy
	}

	// Step 1: Create the NodePool CRD
	nodePool := buildNodePoolObject(config.NodePoolName, config.LabelKey, config.LabelValue, cpuStrategy, gpuStrategy)
	err := kubeClient.Create(ctx, nodePool)
	if err != nil {
		return fmt.Errorf("failed to create nodepool %s: %w", config.NodePoolName, err)
	}

	// Step 2: Get all departments to update (or use specified ones)
	departmentNames := config.DepartmentNames
	if len(departmentNames) == 0 {
		departmentNames, err = listDepartmentNames(ctx, kubeClient)
		if err != nil {
			return err
		}
	}

	// Step 3: Add nodepool resources to all departments
	for _, deptName := range departmentNames {
		err = AddNodePoolToDepartmentResources(ctx, kubeClient, deptName, config.NodePoolName,
			gpuDeserved, cpuDeserved, memoryDeserved)
		if err != nil {
			return fmt.Errorf("failed to add nodepool %s to department %s: %w", config.NodePoolName, deptName, err)
		}
	}

	// Step 4: Get all projects to update (or use specified ones)
	projectNames := config.ProjectNames
	if len(projectNames) == 0 {
		projectNames, err = listProjectNames(ctx, kubeClient)
		if err != nil {
			return err
		}
	}

	// Step 5: Add nodepool resources to all projects
	for _, projName := range projectNames {
		err = AddNodePoolToProjectResources(ctx, kubeClient, projName, config.NodePoolName,
			gpuDeserved, cpuDeserved, memoryDeserved)
		if err != nil {
			return fmt.Errorf("failed to add nodepool %s to project %s: %w", config.NodePoolName, projName, err)
		}
	}

	return nil
}

func getNodePoolCondition(nodePool *v1alpha1.NodePool, conditionType v1alpha1.NodePoolConditionType) *v1alpha1.NodePoolCondition {
	for i := range nodePool.Status.Conditions {
		if nodePool.Status.Conditions[i].Type == conditionType {
			return &nodePool.Status.Conditions[i]
		}
	}
	return nil
}

// listProjectNames returns the names of every project in the cluster.
func listProjectNames(ctx context.Context, kubeClient client.Client) ([]string, error) {
	projectList := &v1alpha1.ProjectList{}
	if err := kubeClient.List(ctx, projectList); err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projectNames := make([]string, 0, len(projectList.Items))
	for _, proj := range projectList.Items {
		projectNames = append(projectNames, proj.Name)
	}
	return projectNames, nil
}

// listDepartmentNames returns the names of every department in the cluster.
func listDepartmentNames(ctx context.Context, kubeClient client.Client) ([]string, error) {
	departmentList := &v1alpha1.DepartmentList{}
	if err := kubeClient.List(ctx, departmentList); err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	departmentNames := make([]string, 0, len(departmentList.Items))
	for _, dept := range departmentList.Items {
		departmentNames = append(departmentNames, dept.Name)
	}
	return departmentNames, nil
}

// RemoveNodePoolFromProjectResources removes the nodepool-specific queue from a project
func RemoveNodePoolFromProjectResources(ctx context.Context, kubeClient client.Client, projectName, nodePoolName string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		project := &v1alpha1.Project{}
		err := kubeClient.Get(ctx, client.ObjectKey{Name: projectName}, project)
		if err != nil {
			return fmt.Errorf("failed to get project %s: %w", projectName, err)
		}

		remainingQueues := make([]v1alpha1.QueueConfig, 0, len(project.Spec.Queues))
		for _, queue := range project.Spec.Queues {
			if queue.Nodepool != nodePoolName {
				remainingQueues = append(remainingQueues, queue)
			}
		}
		project.Spec.Queues = remainingQueues

		err = kubeClient.Update(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to update project %s: %w", projectName, err)
		}
		return nil
	})
}

// AddNodePoolToDepartmentResources adds nodepool-specific resources to a department
func AddNodePoolToDepartmentResources(ctx context.Context, kubeClient client.Client, departmentName, nodePoolName string, gpuDeserved, cpuDeserved, memoryDeserved int64) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		department := &v1alpha1.Department{}
		err := kubeClient.Get(ctx, client.ObjectKey{Name: departmentName}, department)
		if err != nil {
			return fmt.Errorf("failed to get department %s: %w", departmentName, err)
		}

		// Check if queue already exists
		for _, queue := range department.Spec.Queues {
			if queue.Nodepool == nodePoolName {
				// Queue exists, no need for updates
				return nil
			}
		}

		// Add nodepool queue to department
		department.Spec.Queues = append(department.Spec.Queues, v1alpha1.QueueConfig{
			Name:      fmt.Sprintf("q-%s-%s", departmentName, nodePoolName),
			Nodepool:  nodePoolName,
			Resources: common.BuildQueueResources(gpuDeserved, cpuDeserved, memoryDeserved),
		})

		err = kubeClient.Update(ctx, department)
		if err != nil {
			return fmt.Errorf("failed to update department %s: %w", departmentName, err)
		}
		return nil
	})
}

// AddNodePoolToProjectResources adds nodepool-specific resources to a project
func AddNodePoolToProjectResources(ctx context.Context, kubeClient client.Client, projectName, nodePoolName string, gpuDeserved, cpuDeserved, memoryDeserved int64) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		project := &v1alpha1.Project{}
		err := kubeClient.Get(ctx, client.ObjectKey{Name: projectName}, project)
		if err != nil {
			return fmt.Errorf("failed to get project %s: %w", projectName, err)
		}

		// Check if queue already exists
		for _, queue := range project.Spec.Queues {
			if queue.Nodepool == nodePoolName {
				return nil
			}
		}

		// Add nodepool queue to project
		project.Spec.Queues = append(project.Spec.Queues, v1alpha1.QueueConfig{
			Name:      fmt.Sprintf("%s-%s", projectName, nodePoolName),
			Nodepool:  nodePoolName,
			Resources: common.BuildQueueResources(gpuDeserved, cpuDeserved, memoryDeserved),
		})

		err = kubeClient.Update(ctx, project)
		if err != nil {
			return fmt.Errorf("failed to update project %s: %w", projectName, err)
		}
		return nil
	})
}
