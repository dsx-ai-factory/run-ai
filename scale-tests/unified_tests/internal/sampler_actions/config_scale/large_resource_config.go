/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_scale

import (
	"context"
	"fmt"
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

type LargeResourceConfig struct {
	kubeClient client.Client
}

func NewLargeResourceConfig(kubeClient client.Client) ResourceScaleConfig {
	return &LargeResourceConfig{
		kubeClient: kubeClient,
	}
}

func (c *LargeResourceConfig) identifier() string {
	return "Large"
}

func (c *LargeResourceConfig) adjustFakeGPUOperator(ctx context.Context) error {
	fmt.Println("Adjust fake gpu operator resources")
	deployment := &appsv1.Deployment{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "status-updater",
		Namespace: "gpu-operator",
	}, deployment)

	if err != nil {
		return err
	}

	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--controller-threads=10")
	deployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("200m"),
			v1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("400m"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
	}
	err = c.kubeClient.Update(ctx, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (c *LargeResourceConfig) adjustTrainingOperator(ctx context.Context) error {
	fmt.Println("Adjust training operator resources")
	deployment := &appsv1.Deployment{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "training-operator",
		Namespace: "kubeflow",
	}, deployment)

	if err != nil {
		return err
	}
	deployment.Spec.Template.Spec.Containers[0].Image = "kubeflow/training-operator:v1-3f15cb8"

	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--controller-threads=40")
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--kube-api-qps=100")
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--kube-api-burst=300")
	deployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("2"),
			v1.ResourceMemory: resource.MustParse("6Gi"),
		},
	}

	err = c.kubeClient.Update(ctx, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (c *LargeResourceConfig) adjustWorkloadServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust workload services")
	if runaiConfig.Spec.Global.WorkloadServices == nil {
		runaiConfig.Spec.Global.WorkloadServices = &runaiv1global.WorkloadServices{}
	}
	runaiConfig.Spec.Global.WorkloadServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("6Gi"),
		},
	}
	if runaiConfig.Spec.WorkloadController == nil {
		runaiConfig.Spec.WorkloadController = &runaiv1wc.WorkloadController{}
	}
	if runaiConfig.Spec.WorkloadExporter == nil {
		runaiConfig.Spec.WorkloadExporter = &runaiv1we.WorkloadExporter{}
	}
	runaiConfig.Spec.WorkloadController.TrainingMaxConcurrentReconciles = ptr.To(15)
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

func (c *LargeResourceConfig) adjustSyncServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust sync services")
	if runaiConfig.Spec.Global.SyncServices == nil {
		runaiConfig.Spec.Global.SyncServices = &runaiv1global.SyncServices{}
	}
	runaiConfig.Spec.Global.SyncServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("7Gi"),
		},
	}
}

func (c *LargeResourceConfig) adjustSchedulingServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust scheduler services- Large scale")

	if runaiConfig.Spec.Global.SchedulingServices == nil {
		runaiConfig.Spec.Global.SchedulingServices = &runaiv1global.SchedulingServices{}
	}

	runaiConfig.Spec.Global.SchedulingServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("3"),
			v1.ResourceMemory: resource.MustParse("7Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("5"),
			v1.ResourceMemory: resource.MustParse("8Gi"),
		},
	}
}

func (c *LargeResourceConfig) adjustThanos(ctx context.Context) error {
	fmt.Println("Adjust thanos resources - Large scale")
	thanosReceiveStatefulSet := &appsv1.StatefulSet{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "runai-backend-thanos-receive",
		Namespace: "runai-backend",
	}, thanosReceiveStatefulSet)

	if err != nil {
		return fmt.Errorf("get thanos receive statefulset: %w", err)
	}

	thanosReceiveStatefulSet.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("6Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("6Gi"),
		},
	}
	err = c.kubeClient.Update(ctx, thanosReceiveStatefulSet)
	if err != nil {
		return fmt.Errorf("failed to update thanos receive statefulset: %v", err)
	}

	thanosQueryDeployment := &appsv1.Deployment{}
	err = c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "runai-backend-thanos-query",
		Namespace: "runai-backend",
	}, thanosQueryDeployment)

	if err != nil {
		return fmt.Errorf("failed to update thanos query deployment: %v", err)
	}

	thanosQueryDeployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
	}
	err = c.kubeClient.Update(ctx, thanosQueryDeployment)
	if err != nil {
		return fmt.Errorf("failed to update thanos query deployment: %v", err)
	}
	return nil
}

func (c *LargeResourceConfig) adjustPostgres(ctx context.Context) error {
	fmt.Println("Adjust postgres resources - Large scale")
	postgresStatefulSet := &appsv1.StatefulSet{}
	err := c.kubeClient.Get(ctx, types.NamespacedName{
		Name:      "runai-backend-postgresql",
		Namespace: "runai-backend",
	}, postgresStatefulSet)

	if err != nil {
		return fmt.Errorf("get postgres statefulset: %w", err)
	}

	postgresStatefulSet.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
	}
	err = c.kubeClient.Update(ctx, postgresStatefulSet)
	if err != nil {
		return fmt.Errorf("failed to update postgres statefulset: %v", err)
	}
	return nil
}
