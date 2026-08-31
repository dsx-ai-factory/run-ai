/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"sync"

	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

var (
	globalInstance *Installer
	once           sync.Once
	initErr        error
)

// Init initializes the global installer instance with the provided RunAIInstallData
// This should be called once at application startup
// Returns an error if initialization fails
func Init(runaiData *model.RunAIInstallData) error {
	once.Do(func() {
		if runaiData != nil && runaiData.Config != nil {
			step.SetDefaultStepTimeout(runaiData.Config.StepTimeout)
		}
		globalInstance, initErr = NewInstaller(runaiData)
	})
	return initErr
}

// InitWithK8sClient initializes the global installer instance with the provided RunAIInstallData and K8s client
// This is useful for testing scenarios where you want to provide a mock K8s client
// Returns an error if initialization fails
func InitWithK8sClient(runaiData *model.RunAIInstallData, k8sClients k8s.K8sClients) error {
	once.Do(func() {
		if runaiData != nil && runaiData.Config != nil {
			step.SetDefaultStepTimeout(runaiData.Config.StepTimeout)
		}
		globalInstance, initErr = NewInstallerWithK8sClient(runaiData, k8sClients)
	})
	return initErr
}

// Get returns the global installer instance
// Panics if Init() or InitWithK8sClient() hasn't been called first
func Get() *Installer {
	if globalInstance == nil {
		panic("installer not initialized, call Init() or InitWithK8sClient() first")
	}
	return globalInstance
}

// Reset clears the global instance (useful for testing)
// This should NOT be used in production code
func Reset() {
	globalInstance = nil
	initErr = nil
	once = sync.Once{}
}

// IsInitialized returns true if the global installer has been initialized
func IsInitialized() bool {
	return globalInstance != nil
}
