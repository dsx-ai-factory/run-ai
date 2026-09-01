/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dynamic_schema_upgrades

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"

	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

type runaiConfigDynamicSchemaUpgrader interface {
	updateRunaiConfig(resource *unstructured.Unstructured) (*unstructured.Unstructured, error)
	shouldPerformUpgrade(resource *unstructured.Unstructured) bool
	upgradeHandlerIdentifier() string
}

type runaiConfigUpgrade struct {
	specUpgraders []runaiConfigDynamicSchemaUpgrader
}

func NewRunaiConfigUpgrade() *runaiConfigUpgrade {
	var specUpgraders []runaiConfigDynamicSchemaUpgrader
	specUpgraders = append(specUpgraders, newMpsServerUpgrader())
	specUpgraders = append(specUpgraders, newLegacyWorkloadConfigurationUpgrader())
	return &runaiConfigUpgrade{specUpgraders}
}

func (r *runaiConfigUpgrade) Upgrade(ctx context.Context, clients *clients.Clients) error {
	runaiConfigResource := getRunaiConfigDynamicResourceInterface(clients.LocalClusterDynamicClient)
	resource, err := runaiConfigResource.Get(ctx, "runai", metav1.GetOptions{})
	if err != nil {
		// In the case that runaiconfig does not exist, there's nothing to upgrade
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get runaiconfig, err:[%v]", err)
	}
	for _, upgrader := range r.specUpgraders {
		resource, err = r.upgradeSpec(ctx, resource, upgrader)
		if err != nil {
			return err
		}
	}
	logger := log.FromContext(ctx)
	var lastErr error
	for attempt := 1; ; attempt++ {
		logger.Info("attempting to save upgraded runaiconfig", "attempt", attempt)
		lastErr = r.saveUpgradedRunaiConfig(runaiConfigResource, resource)
		if lastErr == nil {
			break
		}
		logger.Error(lastErr, "failed to save upgraded runaiconfig", "attempt", attempt, "conflict", apierrors.IsConflict(lastErr), "reason", apierrors.ReasonForError(lastErr), "error", lastErr.Error())
		time.Sleep(1 * time.Second)
		if apierrors.IsConflict(lastErr) {
			resource, err = runaiConfigResource.Get(ctx, "runai", metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return nil
				}
				return fmt.Errorf("failed to re-read runaiconfig after conflict, err:[%v]", err)
			}
			for _, upgrader := range r.specUpgraders {
				resource, err = r.upgradeSpec(ctx, resource, upgrader)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (r *runaiConfigUpgrade) upgradeSpec(ctx context.Context, resource *unstructured.Unstructured, upgrader runaiConfigDynamicSchemaUpgrader) (*unstructured.Unstructured, error) {
	logger := log.FromContext(ctx)
	shouldUpgrade := upgrader.shouldPerformUpgrade(resource)
	if !shouldUpgrade {
		logger.Info(fmt.Sprintf("No runaiconfig upgrade needed for handler:[%s]", upgrader.upgradeHandlerIdentifier()))
		return resource, nil
	}
	resource, err := upgrader.updateRunaiConfig(resource)
	if err != nil {
		return nil, err
	}
	logger.Info(fmt.Sprintf("Completed runaiconfig upgrade, for handler:[%s]", upgrader.upgradeHandlerIdentifier()))
	return resource, nil

}

func (r *runaiConfigUpgrade) saveUpgradedRunaiConfig(runaiConfigResource dynamic.ResourceInterface, resource *unstructured.Unstructured) error {
	data, err := yaml.Marshal(resource)
	if err != nil {
		return fmt.Errorf("failed to marshal unstructured runaiconfig: %w", err)
	}
	newRunaiConf := &unstructured.Unstructured{Object: map[string]interface{}{}}
	err = yaml.Unmarshal(data, &newRunaiConf)
	if err != nil {
		return fmt.Errorf("failed to update runaiconfig. Failed to marshal config data: %w", err)
	}
	_, err = runaiConfigResource.Update(context.Background(), newRunaiConf, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update runaiconfig: %w", err)
	}
	return nil
}

func getRunaiConfigDynamicResourceInterface(dynamicClient dynamic.Interface) dynamic.ResourceInterface {
	return dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
}
