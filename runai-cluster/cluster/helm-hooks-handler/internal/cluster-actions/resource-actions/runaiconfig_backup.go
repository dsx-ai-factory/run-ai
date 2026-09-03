/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_actions

import (
	"context"
	"fmt"
	"log"
	"time"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	configMapName      = runaiv1.DefaultName
	configmapNamespace = runaiv1.DefaultNamespace
	runaiConfigKey     = configMapName
)

// BackupRunaiConfig saves the existing RunaiConfig named 'runai' in the 'runai' namespace to a configmap
// named 'runai' in the 'runai' namespace
func BackupRunaiConfig(k8s client.Client) error {
	runaiConfig, err := GetExistingRunaiConfig(k8s)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return fmt.Errorf("failed to get runaiconfig: %w", err)
	}

	// If runaiconfig does not exist in the cluster, preserve the existing backup
	if k8sErrors.IsNotFound(err) {
		return nil
	}

	// Delete previous backup if runaiconfig exists (which is the latest configuration)
	err = k8s.Delete(context.Background(), runaiConfigBackupConfigMap())
	if err != nil && !k8sErrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete previous runaiconfig backup: %w", err)
	}

	runaiConfigYAML, err := yaml.Marshal(runaiConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal runaiconfig: %w", err)
	}

	cm := runaiConfigBackupConfigMap()
	cm.Data = map[string]string{
		runaiConfigKey: string(runaiConfigYAML),
	}

	err = k8s.Create(context.TODO(), cm)
	if err != nil {
		return fmt.Errorf("failed to create runaiconfig backup configmap: %w", err)
	}

	return nil
}

// WaitForRunaiConfigToBeDeletedIfNeeded waits for the RunaiConfig resource to be deleted from the cluster.
// If forceWaitForDeletion is true, it will wait for deletion even if the resource is not marked for deletion.
// If forceWaitForDeletion is false, it will only wait if the resource has a deletion timestamp.
// Returns an error if the operation times out.
func WaitForRunaiConfigToBeDeletedIfNeeded(ctx context.Context, k8s client.Client, forceWaitForDeletion bool) error {
	runaiConfig := &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      runaiv1.DefaultName,
		},
	}

	interval := 10 * time.Second
	attempts := 30
	for ; attempts > 0; attempts-- {
		// Check cancellation before each attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		err := k8s.Get(ctx, client.ObjectKeyFromObject(runaiConfig), runaiConfig)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				// runaiConfig was deleted, we can continue
				break
			}

			return err
		}

		if !forceWaitForDeletion && runaiConfig.DeletionTimestamp.IsZero() {
			// Runaiconfig should not be deleted, we can continue
			break
		}

		log.Printf("waiting for runaiconfig to be deleted, will attempt %d more times", attempts)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}

	// The operation timed out, exit with an error
	if attempts == 0 {
		return fmt.Errorf("timed out waiting for RunaiConfig to be deleted")
	}

	return nil
}

func runaiConfigBackupConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: configMapName,
			Name:      configmapNamespace,
		},
	}
}

func getRunaiConfigBackupConfigMap(k8s client.Client) (*corev1.ConfigMap, error) {
	runaiConfigBackupCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: configmapNamespace,
			Name:      configMapName,
		},
	}

	err := k8s.Get(context.Background(), client.ObjectKeyFromObject(runaiConfigBackupCM), runaiConfigBackupCM)
	if err != nil {
		return nil, err
	}

	return runaiConfigBackupCM, nil
}

func GetRunaiConfigFromBackupConfigMapOrDefault(k8s client.Client) (*runaiv1.RunaiConfig, error) {
	runaiConfigBackupCM, err := getRunaiConfigBackupConfigMap(k8s)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return nil, err
	}

	if k8sErrors.IsNotFound(err) {
		return DefaultRunaiConfig(), nil
	}

	runaiConfigBackup := &runaiv1.RunaiConfig{}
	err = yaml.Unmarshal([]byte(runaiConfigBackupCM.Data[runaiConfigKey]), runaiConfigBackup)
	if err != nil {
		return nil, err
	}

	return runaiConfigBackup, nil
}
