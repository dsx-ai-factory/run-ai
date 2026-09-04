/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dynamic_schema_upgrades

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	workloadControllerOperandName          = "workload-controller"
	inferenceWorkloadControllerOperandName = "inference-workload-controller"
	anyworkloadControllerOperandName       = "anyworkload-controller"

	gpuNetworkAccelerationEnabledField = "GPUNetworkAccelerationEnabled"
)

// gpuNetworkAccelerationUpgrade propagates the GPUNetworkAccelerationEnabled (MNNVL)
// feature to the anyworkload-controller when it was enabled on either the legacy
// workload-controller or inference-workload-controller. This keeps the feature turned
// on when workloads are handled by the anyworkload-controller after an upgrade.
type gpuNetworkAccelerationUpgrade struct{}

func newGpuNetworkAccelerationUpgrader() *gpuNetworkAccelerationUpgrade {
	return &gpuNetworkAccelerationUpgrade{}
}

// NewRunaiConfigEnableAnyworkloadGPUNetworkAcceleration enables the
// GPUNetworkAccelerationEnabled feature on the anyworkload-controller when it was
// enabled on the workload-controller or inference-workload-controller.
func NewRunaiConfigEnableAnyworkloadGPUNetworkAcceleration() *runaiConfigUpgrade {
	return &runaiConfigUpgrade{[]runaiConfigDynamicSchemaUpgrader{
		newGpuNetworkAccelerationUpgrader(),
	}}
}

func (u *gpuNetworkAccelerationUpgrade) shouldPerformUpgrade(resource *unstructured.Unstructured) bool {
	legacyEnabled := u.operandGpuNetworkAccelerationEnabled(resource, workloadControllerOperandName) ||
		u.operandGpuNetworkAccelerationEnabled(resource, inferenceWorkloadControllerOperandName)
	if !legacyEnabled {
		return false
	}

	// Nothing to do if anyworkload-controller already has the feature enabled.
	return !u.operandGpuNetworkAccelerationEnabled(resource, anyworkloadControllerOperandName)
}

func (u *gpuNetworkAccelerationUpgrade) operandGpuNetworkAccelerationEnabled(resource *unstructured.Unstructured, operandName string) bool {
	spec, ok := resource.Object["spec"].(map[string]interface{})
	if !ok {
		return false
	}
	operandConf, ok := spec[operandName].(map[string]interface{})
	if !ok {
		return false
	}
	enabled, ok := operandConf[gpuNetworkAccelerationEnabledField].(bool)
	if !ok {
		return false
	}
	return enabled
}

func (u *gpuNetworkAccelerationUpgrade) updateRunaiConfig(resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	spec, ok := resource.Object["spec"].(map[string]interface{})
	if !ok {
		return resource, nil // No spec found, nothing to update
	}

	anyworkloadConf, ok := spec[anyworkloadControllerOperandName].(map[string]interface{})
	if !ok {
		anyworkloadConf = map[string]interface{}{}
		spec[anyworkloadControllerOperandName] = anyworkloadConf
	}
	anyworkloadConf[gpuNetworkAccelerationEnabledField] = true
	return resource, nil
}

func (u *gpuNetworkAccelerationUpgrade) upgradeHandlerIdentifier() string {
	return "anyworkload-controller-gpu-network-acceleration"
}
