/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_scale

import (
	"context"
	"fmt"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1global "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MediumResourceConfig struct {
	kubeClient client.Client
}

func NewMediumResourceConfig(kubeClient client.Client) ResourceScaleConfig {
	return &MediumResourceConfig{
		kubeClient: kubeClient,
	}
}

func (c *MediumResourceConfig) identifier() string {
	return "Medium"

}

func (c *MediumResourceConfig) adjustFakeGPUOperator(ctx context.Context) error {
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

func (c *MediumResourceConfig) adjustTrainingOperator(ctx context.Context) error {
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

	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--controller-threads=20")
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--kube-api-qps=100")
	deployment.Spec.Template.Spec.Containers[0].Args = append(deployment.Spec.Template.Spec.Containers[0].Args, "--kube-api-burst=300")
	deployment.Spec.Template.Spec.Containers[0].Resources = v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("3Gi"),
		},
	}

	err = c.kubeClient.Update(ctx, deployment)
	if err != nil {
		return err
	}
	return nil
}

func (c *MediumResourceConfig) adjustWorkloadServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
}

func (c *MediumResourceConfig) adjustSyncServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
}

func (c *MediumResourceConfig) adjustSchedulingServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig) {
	fmt.Println("Adjust scheduler services - Medium scale")
	if runaiConfig.Spec.Global.SchedulingServices == nil {
		runaiConfig.Spec.Global.SchedulingServices = &runaiv1global.SchedulingServices{}
	}
	runaiConfig.Spec.Global.SchedulingServices.Resources = &runaiv1global.Resources{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: resource.MustParse("1Gi"),
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("2"),
			v1.ResourceMemory: resource.MustParse("1Gi"),
		},
	}
}

func (c *MediumResourceConfig) adjustThanos(ctx context.Context) error {
	return nil
}

func (c *MediumResourceConfig) adjustPostgres(ctx context.Context) error {
	return nil
}
