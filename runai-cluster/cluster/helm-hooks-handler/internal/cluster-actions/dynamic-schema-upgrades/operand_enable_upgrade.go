/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dynamic_schema_upgrades

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
)

const (
	defaultOperandEnabled = true
)

type runaiOperandEnableUpgrade struct {
	desiredOperandEnabled bool
	operandName           string
}

func newOperandEnableUpgrader(desiredOperandEnabled bool, operandName string) *runaiOperandEnableUpgrade {
	return &runaiOperandEnableUpgrade{
		desiredOperandEnabled: desiredOperandEnabled,
		operandName:           operandName,
	}
}

func (u *runaiOperandEnableUpgrade) shouldPerformUpgrade(resource *unstructured.Unstructured) bool {
	operandEnabledInRunaiConfig := u.getOperandEnabledInRunaiConfig(resource)

	if operandEnabledInRunaiConfig != u.desiredOperandEnabled {
		// if desired state is different from the current state,
		// we need to continue and preform the updateRunaiConfig
		return true
	}
	return false
}

func (u *runaiOperandEnableUpgrade) getOperandEnabledInRunaiConfig(resource *unstructured.Unstructured) bool {
	spec, ok := resource.Object["spec"]
	if !ok {
		return false
	}
	specObj, ok := spec.(map[string]interface{})
	if !ok {
		return false
	}
	runaiOperandConf, ok := specObj[u.operandName]
	if !ok {
		return defaultOperandEnabled
	}
	runaiOperandConfObj, ok := runaiOperandConf.(map[string]interface{})
	if !ok {
		// this means an error in the runai config, so we return false
		return false
	}
	enabled, ok := runaiOperandConfObj["enabled"]
	if !ok {
		return defaultOperandEnabled
	}

	enabledBool, ok := enabled.(bool)
	if !ok {
		// this means an error in the runai config, so we return false
		return false
	}
	return enabledBool
}

func (u *runaiOperandEnableUpgrade) updateRunaiConfig(resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	spec, ok := resource.Object["spec"].(map[string]interface{})
	if !ok {
		return resource, nil // No spec found, nothing to update
	}

	runaiOperandConf, ok := spec[u.operandName].(map[string]interface{})
	if !ok {
		runaiOperandConf = map[string]interface{}{"enabled": ptr.To(u.desiredOperandEnabled)}
		spec[u.operandName] = runaiOperandConf
		return resource, nil
	}
	runaiOperandConf["enabled"] = ptr.To(u.desiredOperandEnabled)
	return resource, nil
}

func (u *runaiOperandEnableUpgrade) upgradeHandlerIdentifier() string {
	if u.desiredOperandEnabled {
		return fmt.Sprintf("%s-enable-upgrade", u.operandName)
	}
	return fmt.Sprintf("%s-disable-upgrade", u.operandName)
}
