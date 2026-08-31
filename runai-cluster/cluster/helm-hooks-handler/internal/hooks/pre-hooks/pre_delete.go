/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	resource_actions "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions/resource-actions"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PreDeleteHook struct{}

func NewPreDeleteHook() *PreDeleteHook {
	return &PreDeleteHook{}
}

func (p *PreDeleteHook) GetIdentifier() string {
	return hooks.PreDeleteHookIdentifier
}

func (p *PreDeleteHook) Call(ctx context.Context, resources *clients.Clients) error {
	logger := log.FromContext(ctx)
	logger.Info(fmt.Sprintf("Starting %s", p.GetIdentifier()))
	defer logger.Info(fmt.Sprintf("Done %s", p.GetIdentifier()))

	for {
		deleteOptions := client.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationForeground),
		}
		err := resources.LocalClusterClient.Delete(context.Background(),
			&runaiv1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: runaiv1.ClusterName,
				},
			}, &deleteOptions)
		if k8sErrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			break
		} else if err != nil {
			return err
		}

		logger.Info("Waiting for cluster and all its owned resources to be deleted...")
	}

	logger.Info("Runai cluster was successfully deleted")

	// Wait for RunaiConfig to be deleted and finalized
	logger.Info("Waiting for RunaiConfig to be deleted and finalized...")
	err := resource_actions.WaitForRunaiConfigToBeDeletedIfNeeded(ctx, resources.LocalClusterClient, true)
	if err != nil {
		return err
	}
	logger.Info("RunaiConfig was successfully deleted and finalized")

	return nil
}
