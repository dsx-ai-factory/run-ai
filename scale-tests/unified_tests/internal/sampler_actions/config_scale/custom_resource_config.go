/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_scale

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/scale-tests/unified_tests/internal/scale"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1global "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runaiv1wc "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	runaiv1we "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-exporter"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CustomResourceConfig struct {
	kubeClient    client.Client
	clusterConfig *scale.ClusterConfig
}

func NewCustomResourceConfig(kubeClient client.Client, clusterConfig *scale.ClusterConfig) ResourceScaleConfig {
	return &CustomResourceConfig{
		kubeClient:    kubeClient,
		clusterConfig: clusterConfig,
	}
}

func (c *CustomResourceConfig) identifier() string {
	return "Custom"
}

func (c *CustomResourceConfig) adjustSchedulingServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust scheduler services - Custom resources")

	if runaiConfig.Spec.Global.SchedulingServices == nil {
		runaiConfig.Spec.Global.SchedulingServices = &runaiv1global.SchedulingServices{}
	}

	// Use provided values or fall back to medium defaults
	cpuRequest := c.getOrDefault(c.clusterConfig.SchedulingServicesCPURequest, "1")
	cpuLimit := c.getOrDefault(c.clusterConfig.SchedulingServicesCPULimit, "2")
	memRequest := c.getOrDefault(c.clusterConfig.SchedulingServicesMemoryRequest, "1Gi")
	memLimit := c.getOrDefault(c.clusterConfig.SchedulingServicesMemoryLimit, "1Gi")

	runaiConfig.Spec.Global.SchedulingServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuRequest),
			v1.ResourceMemory: resource.MustParse(memRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuLimit),
			v1.ResourceMemory: resource.MustParse(memLimit),
		},
	}
}

func (c *CustomResourceConfig) adjustWorkloadServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust workload services - Custom resources")

	if runaiConfig.Spec.Global.WorkloadServices == nil {
		runaiConfig.Spec.Global.WorkloadServices = &runaiv1global.WorkloadServices{}
	}

	cpuRequest := c.getOrDefault(c.clusterConfig.WorkloadServicesCPURequest, "1")
	cpuLimit := c.getOrDefault(c.clusterConfig.WorkloadServicesCPULimit, "1")
	memRequest := c.getOrDefault(c.clusterConfig.WorkloadServicesMemoryRequest, "3Gi")
	memLimit := c.getOrDefault(c.clusterConfig.WorkloadServicesMemoryLimit, "6Gi")

	runaiConfig.Spec.Global.WorkloadServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuRequest),
			v1.ResourceMemory: resource.MustParse(memRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuLimit),
			v1.ResourceMemory: resource.MustParse(memLimit),
		},
	}

	if runaiConfig.Spec.WorkloadController == nil {
		runaiConfig.Spec.WorkloadController = &runaiv1wc.WorkloadController{}
	}
	if runaiConfig.Spec.WorkloadExporter == nil {
		runaiConfig.Spec.WorkloadExporter = &runaiv1we.WorkloadExporter{}
	}
	runaiConfig.Spec.WorkloadController.TrainingMaxConcurrentReconciles = ptr.To(15)

	// Configure workload exporter resources
	runaiConfig.Spec.WorkloadExporter.Resources = &runaiv1we.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("300m"),
			v1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("600m"),
			v1.ResourceMemory: resource.MustParse("4Gi"),
		},
	}
}

func (c *CustomResourceConfig) adjustSyncServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust sync services - Custom resources")

	if runaiConfig.Spec.Global.SyncServices == nil {
		runaiConfig.Spec.Global.SyncServices = &runaiv1global.SyncServices{}
	}

	cpuRequest := c.getOrDefault(c.clusterConfig.SyncServicesCPURequest, "1")
	cpuLimit := c.getOrDefault(c.clusterConfig.SyncServicesCPULimit, "1")
	memRequest := c.getOrDefault(c.clusterConfig.SyncServicesMemoryRequest, "3Gi")
	memLimit := c.getOrDefault(c.clusterConfig.SyncServicesMemoryLimit, "7Gi")

	runaiConfig.Spec.Global.SyncServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuRequest),
			v1.ResourceMemory: resource.MustParse(memRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuLimit),
			v1.ResourceMemory: resource.MustParse(memLimit),
		},
	}
}

func (c *CustomResourceConfig) adjustTrainingOperator(ctx context.Context) error {
	fmt.Println("Adjust training operator - Custom resources")

	deployment := &appsv1.Deployment{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "training-operator",
		Namespace: "kubeflow",
	}, deployment)
	if err != nil {
		return err
	}

	deployment.Spec.Template.Spec.Containers[0].Image = "kubeflow/training-operator:v1-3f15cb8"

	// Add controller threads and API tuning
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--controller-threads=40")
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--kube-api-qps=100")
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--kube-api-burst=300")

	cpuRequest := c.getOrDefault(c.clusterConfig.TrainingOperatorCPURequest, "1")
	cpuLimit := c.getOrDefault(c.clusterConfig.TrainingOperatorCPULimit, "2")
	memRequest := c.getOrDefault(c.clusterConfig.TrainingOperatorMemoryRequest, "3Gi")
	memLimit := c.getOrDefault(c.clusterConfig.TrainingOperatorMemoryLimit, "6Gi")

	deployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuRequest),
			v1.ResourceMemory: resource.MustParse(memRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuLimit),
			v1.ResourceMemory: resource.MustParse(memLimit),
		},
	}

	return c.kubeClient.Update(ctx, deployment)
}

func (c *CustomResourceConfig) adjustFakeGPUOperator(ctx context.Context) error {
	fmt.Println("Adjust fake GPU operator - Custom resources")

	deployment := &appsv1.Deployment{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "status-updater",
		Namespace: "gpu-operator",
	}, deployment)
	if err != nil {
		return err
	}

	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--controller-threads=10")

	cpuRequest := c.getOrDefault(c.clusterConfig.FakeGPUOperatorCPURequest, "200m")
	cpuLimit := c.getOrDefault(c.clusterConfig.FakeGPUOperatorCPULimit, "400m")
	memRequest := c.getOrDefault(c.clusterConfig.FakeGPUOperatorMemoryRequest, "1Gi")
	memLimit := c.getOrDefault(c.clusterConfig.FakeGPUOperatorMemoryLimit, "3Gi")

	deployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuRequest),
			v1.ResourceMemory: resource.MustParse(memRequest),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(cpuLimit),
			v1.ResourceMemory: resource.MustParse(memLimit),
		},
	}

	return c.kubeClient.Update(ctx, deployment)
}

func (c *CustomResourceConfig) adjustThanos(ctx context.Context) error {
	// Use medium defaults for Thanos (not critical for scale testing)
	return nil
}

func (c *CustomResourceConfig) adjustPostgres(ctx context.Context) error {
	cpuRequest := c.clusterConfig.PostgresCPURequest
	cpuLimit := c.clusterConfig.PostgresCPULimit
	memRequest := c.clusterConfig.PostgresMemoryRequest
	memLimit := c.clusterConfig.PostgresMemoryLimit

	if cpuRequest == "" && cpuLimit == "" && memRequest == "" && memLimit == "" {
		return nil
	}

	fmt.Println("Adjust postgres resources - Custom resources")
	postgresStatefulSet := &appsv1.StatefulSet{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "runai-backend-postgresql",
		Namespace: "runai-backend",
	}, postgresStatefulSet)

	if err != nil {
		return fmt.Errorf("get postgres statefulset: %w", err)
	}

	resources := v1.ResourceRequirements{}
	if cpuRequest != "" || memRequest != "" {
		resources.Requests = v1.ResourceList{}
		if cpuRequest != "" {
			resources.Requests[v1.ResourceCPU] = resource.MustParse(cpuRequest)
		}
		if memRequest != "" {
			resources.Requests[v1.ResourceMemory] = resource.MustParse(memRequest)
		}
	}
	if cpuLimit != "" || memLimit != "" {
		resources.Limits = v1.ResourceList{}
		if cpuLimit != "" {
			resources.Limits[v1.ResourceCPU] = resource.MustParse(cpuLimit)
		}
		if memLimit != "" {
			resources.Limits[v1.ResourceMemory] = resource.MustParse(memLimit)
		}
	}

	postgresStatefulSet.Spec.Template.Spec.Containers[0].Resources = resources
	err = c.kubeClient.Update(ctx, postgresStatefulSet)
	if err != nil {
		return fmt.Errorf("failed to update postgres statefulset: %v", err)
	}
	return nil
}

// Helper function to use provided value or fall back to default
func (c *CustomResourceConfig) getOrDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}
