/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"
	"fmt"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	inferencecontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	jobcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/job-controller"
	workloadcontroller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func commitControllerDisable(ctx context.Context, k8sClient client.Client) error {
	key := client.ObjectKey{Namespace: runaiv1.DefaultNamespace, Name: runaiv1.DefaultName}
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		config := &runaiv1.RunaiConfig{}
		if err := k8sClient.Get(ctx, key, config); err != nil {
			return err
		}
		state, err := classifyControllerState(config)
		if err != nil {
			return err
		}
		if state == controllersDisabled {
			return nil
		}
		setControllerFlags(config, false)
		return k8sClient.Update(ctx, config)
	})
	if err != nil {
		return fmt.Errorf("atomically disable legacy controllers: %w", err)
	}
	return nil
}

func setControllerFlags(config *runaiv1.RunaiConfig, enabled bool) {
	if config.Spec.WorkloadController == nil {
		config.Spec.WorkloadController = &workloadcontroller.WorkloadController{}
	}
	if config.Spec.InferenceWorkloadController == nil {
		config.Spec.InferenceWorkloadController = &inferencecontroller.InferenceWorkloadController{}
	}
	if config.Spec.JobController == nil {
		config.Spec.JobController = &jobcontroller.JobController{}
	}
	config.Spec.WorkloadController.Enabled = ptr.To(enabled)
	config.Spec.InferenceWorkloadController.Enabled = ptr.To(enabled)
	config.Spec.JobController.Enabled = ptr.To(enabled)
}
