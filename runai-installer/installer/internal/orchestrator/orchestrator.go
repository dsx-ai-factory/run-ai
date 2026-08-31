/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/config"
	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/install"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type Installer struct {
	k8sClient k8s.K8sClients
	runaiData *model.RunAIInstallData
	deps      InstallDeps
}

func NewInstaller(runaiData *model.RunAIInstallData) (*Installer, error) {
	k8sClients, err := k8s.GetK8sClients()
	if err != nil {
		return nil, fmt.Errorf("cannot get Kubernetes clients: %w", err)
	}
	return NewInstallerWithDeps(runaiData, ProductionInstallDeps(k8sClients, controlPlaneVersion(runaiData)))
}

func NewInstallerWithK8sClient(runaiData *model.RunAIInstallData, k8sClients k8s.K8sClients) (*Installer, error) {
	return NewInstallerWithDeps(runaiData, ProductionInstallDeps(k8sClients, controlPlaneVersion(runaiData)))
}

func controlPlaneVersion(runaiData *model.RunAIInstallData) string {
	if runaiData == nil || runaiData.Config == nil {
		return ""
	}
	return runaiData.Config.CpVersion
}

func NewInstallerWithDeps(runaiData *model.RunAIInstallData, deps InstallDeps) (*Installer, error) {
	if runaiData == nil {
		return nil, fmt.Errorf("runaiData cannot be nil")
	}
	if deps.K8sClient == nil {
		return nil, fmt.Errorf("k8sClients cannot be nil")
	}

	if runaiData.Config == nil {
		runaiData.Config = &config.InstallConfig{}
	}
	err := ValidateInputs(runaiData, deps.K8sClient, configinputs.GetConfigInputs())
	if err != nil {
		return nil, fmt.Errorf("failed to validate installer inputs: %v", err)
	}
	return &Installer{
		k8sClient: deps.K8sClient,
		runaiData: runaiData,
		deps:      deps,
	}, nil
}

func ValidateInputs(runaiData *model.RunAIInstallData, k8sClients k8s.K8sClients, configItemsMap map[configinputs.InstallInputType]*configinputs.ConfigurationItem) error {
	if runaiData.Config.Mode != config.ModeCLI {
		return nil
	}
	installerInputs := install.GetRelevantInputTypes(runaiData, prereq.GetPrerequisites(k8sClients))
	for inputType, mandatory := range installerInputs {
		configItem := configItemsMap[inputType]
		if configItem == nil {
			return fmt.Errorf("configuration item not found for input type: %v", inputType)
		}
		if value, exists := runaiData.Inputs[inputType]; !exists && mandatory {
			return fmt.Errorf("missing mandatory input: %s", configItem.Label)
		} else if exists {
			validationResult := configItem.ValidationFunc(value)
			if !validationResult.Status {
				return fmt.Errorf("invalid value for input %s: %s", configItem.Label, validationResult.Message)
			}
		}
	}
	return nil
}

func (i *Installer) RunInstallation(flowGroups []*FlowGroup) error {
	logging.Infof("================================================================================")
	logging.Infof("=                Starting Run:AI Installation (Library Mode)                  =")
	logging.Infof("================================================================================")

	for idx, flowGroup := range flowGroups {
		if flowGroup == nil || len(flowGroup.Flows) == 0 {
			logging.Infof("Skipping flow group %d: no flows to execute", idx+1)
			continue
		}

		logging.Infof("Executing flow group %d with %d flows", idx+1, len(flowGroup.Flows))
		err := i.executeFlowGroup(flowGroup.Flows)
		if err != nil {
			return fmt.Errorf("flow group %d execution failed: %w", idx+1, err)
		}
		logging.Infof("Flow group %d completed successfully", idx+1)
	}

	return nil
}

func (i *Installer) executeFlowGroup(flows []*flow.InstallFlow) error {
	if len(flows) == 0 {
		return nil
	}
	manager := flow.NewFlowManager(flows)
	events := make(chan interface{}, 100)
	manager.ExecuteFlow(context.Background(), i.runaiData, events)

	manager.Wait()

	close(events)
	return nil
}

func (i *Installer) RunInstallationWithEventListner() (<-chan *flow.FlowStatusReportingMsg, error) {
	executor := NewInstallationExecutor(i, i.runaiData)
	eventsChan := executor.Start()

	go func() {
		_ = executor.Wait()
	}()

	return eventsChan, nil
}
