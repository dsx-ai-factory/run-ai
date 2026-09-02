/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_scale

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv1global "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type ResourceScaleConfig interface {
	adjustFakeGPUOperator(ctx context.Context) error
	adjustTrainingOperator(ctx context.Context) error
	adjustWorkloadServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig)
	adjustSyncServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig)
	adjustSchedulingServices(ctx context.Context, runaiConfig *runaiv1.RunaiConfig)
	adjustThanos(ctx context.Context) error
	adjustPostgres(ctx context.Context) error
	identifier() string
}

func ConfigureResourceForScaleRun(kubeClient client.Client, clusterConfig *scale.ClusterConfig) ResourceScaleConfig {
	switch clusterConfig.Resources {
	case "large":
		return NewLargeResourceConfig(kubeClient)
	case "medium":
		return NewMediumResourceConfig(kubeClient)
	case "small":
		return NewSmallResourceConfig(kubeClient)
	case "custom":
		return NewCustomResourceConfig(kubeClient, clusterConfig)
	default:
		panic("Unknown resource type: " + string(clusterConfig.Resources))
	}

}

func Configure(ctx context.Context, kubeClient client.Client, clusterConfig *scale.ClusterConfig, resourceScaleConfig ResourceScaleConfig) error {
	fmt.Println(fmt.Sprintf("Adjust test resources, configured size: %s", resourceScaleConfig.identifier()))
	if err := adjustRunaiServicesResources(ctx, kubeClient, resourceScaleConfig); err != nil {
		return err
	}
	if err := resourceScaleConfig.adjustTrainingOperator(ctx); err != nil {
		return fmt.Errorf("error adjusting training operator: %v", err)
	}
	if clusterConfig.KwokEnabled {
		if err := resourceScaleConfig.adjustFakeGPUOperator(ctx); err != nil {
			return err
		}
	}
	if err := resourceScaleConfig.adjustThanos(ctx); err != nil {
		return fmt.Errorf("error adjusting Thanos: %v", err)
	}

	if err := resourceScaleConfig.adjustPostgres(ctx); err != nil {
		return fmt.Errorf("error adjusting postgres: %v", err)
	}
	time.Sleep(1 * time.Minute) // Wait for the changes to take effect
	fmt.Printf("Successfully adjusted test resources")
	return nil
}

func adjustRunaiServicesResources(ctx context.Context, kubeClient client.Client, resourceScaleConfig ResourceScaleConfig) error {
	fmt.Println("Adjust runai config services")
	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      runaiv1.DefaultName,
		},
	}
	if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(runaiConfig), runaiConfig); err != nil {
		return fmt.Errorf("failed to get runai config: %v", err)
	}

	if runaiConfig.Spec.Global == nil {
		runaiConfig.Spec.Global = &runaiv1global.Global{}
	}
	resourceScaleConfig.adjustSchedulingServices(ctx, runaiConfig)
	resourceScaleConfig.adjustSyncServices(ctx, runaiConfig)
	resourceScaleConfig.adjustWorkloadServices(ctx, runaiConfig)

	if err := kubeClient.Update(ctx, runaiConfig); err != nil {
		return fmt.Errorf("failed to update runai config: %v", err)
	}
	return nil
}
