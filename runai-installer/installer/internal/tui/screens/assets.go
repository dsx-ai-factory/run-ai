/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package screens

import (
	"sync"

	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq"
	prereqids "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/screens/common"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

var (
	screens     map[common.ScreenType]common.Screen
	screensOnce sync.Once
)

func GetScreens() map[common.ScreenType]common.Screen {
	screensOnce.Do(func() {
		screens = initScreens()
	})
	return screens
}

func GetScreensWithK8s(k8sClient k8s.K8sClients) map[common.ScreenType]common.Screen {
	prereqMap := prereq.GetPrerequisites(k8sClient)

	// Create screens with proper next screen references
	installationScreen := NewRunaiInstallationScreen(k8sClient, nil)
	prerequisiteInstallScreen := NewPrerequisiteInstallScreen(k8sClient, prereqMap, installationScreen)
	inputValidationScreen := NewInputValidationScreen(prerequisiteInstallScreen, prereqMap)
	configurationScreen := NewConfigurationScreen(inputValidationScreen, k8sClient, prereqMap)
	prerequisitesScreen := NewPrerequisitesScreen(prereqMap, configurationScreen)

	return map[common.ScreenType]common.Screen{
		common.ScreenPrerequisites:       prerequisitesScreen,
		common.ScreenConfiguration:       configurationScreen,
		common.ScreenPrerequisiteInstall: prerequisiteInstallScreen,
		common.ScreenInputValidation:     inputValidationScreen,
		common.ScreenInstallation:        installationScreen,
	}
}

func initScreens() map[common.ScreenType]common.Screen {
	k8sClient, err := k8s.GetK8sClients()
	if err != nil {
		panic("k8s client should be initialized before calling initScreens: " + err.Error())
	}
	return GetScreensWithK8s(k8sClient)
}

var (
	prerequisiteViewOrder     map[string]int
	prerequisiteViewOrderOnce sync.Once
)

func GetPrerequisiteViewOrder() map[string]int {
	prerequisiteViewOrderOnce.Do(func() {
		prerequisiteViewOrder = map[string]int{
			prereqids.PrerequisiteIdentifierIngress:          0,
			prereqids.PrerequisiteIdentifierPrometheus:       1,
			prereqids.PrerequisiteIdentifierGpuOperator:      2,
			prereqids.PrerequisiteIdentifierNetworkOperator:  3,
			prereqids.PrerequisiteIdentifierKnative:          4,
			prereqids.PrerequisiteIdentifierTrainingOperator: 5,
			prereqids.PrerequisiteIdentifierLWS:              6,
		}
	})
	return prerequisiteViewOrder
}

var (
	inputsViewOrder     map[configinputs.InstallInputType]int
	inputsViewOrderOnce sync.Once
)

func GetInputsViewOrder() map[configinputs.InstallInputType]int {
	inputsViewOrderOnce.Do(func() {
		inputsViewOrder = map[configinputs.InstallInputType]int{
			configinputs.InstallInputClusterName:                0,
			configinputs.InstallInputRegistryPullSecret:         1,
			configinputs.InstallInputFQDN:                       2,
			configinputs.InstallInputIngressControllerHTTPSPort: 3,
			configinputs.InstallInputSSLPath:                    4,
			configinputs.InstallInputSSLKeyPath:                 5,
			configinputs.InstallInputCAChainPath:                6,
			configinputs.InstallInputWildcardFQDN:               7,
			configinputs.InstallInputWildcardCertPath:           8,
			configinputs.InstallInputWildcardCertKeyPath:        9,
			configinputs.InstallInputWorkloadRouting:            10,
			configinputs.InstallInputHostRoutingCertPath:        11,
			configinputs.InstallInputHostRoutingKeyPath:         12,
		}
	})
	return inputsViewOrder
}
