/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"maps"

	configinputs "github.com/run-ai/runai/runai-installer/installer/internal/inputs"
	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq"
	"github.com/run-ai/runai/runai-installer/installer/internal/prereq/components"
	"github.com/run-ai/runai/runai-installer/installer/internal/session"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

func AlignInstallData(runaiData *model.RunAIInstallData, k8sClients k8s.K8sClients) {
	validateAndSetConfiguredPreReq(runaiData, prereq.GetPrerequisites(k8sClients))
	mergeInputData(runaiData, k8sClients)
}

func mergeInputData(runaiData *model.RunAIInstallData, clients k8s.K8sClients) {
	//Merges saved inputs with new inputs, new inputs from flags take precedence over saved inputs from secret
	tempInputs := make(map[configinputs.InstallInputType]string)
	savedInputs, err := session.LoadInputs(clients)
	if err == nil && len(savedInputs) > 0 {
		maps.Copy(tempInputs, savedInputs)
	}
	if len(runaiData.Inputs) > 0 {
		maps.Copy(tempInputs, runaiData.Inputs)
	}
	runaiData.Inputs = tempInputs
}

func validateAndSetConfiguredPreReq(runaiData *model.RunAIInstallData, preReqMap map[string]components.Prerequisite) {
	existingPreReqs := make(map[string]struct{})
	if len(runaiData.Config.PreReq) > 0 {
		if runaiData.Config.PreReq[0] == "none" {
			return
		}
		if runaiData.Config.PreReq[0] == "all" {
			fillAllPreReq(runaiData, preReqMap)
			return
		}
		for _, preReq := range runaiData.Config.PreReq {
			if _, exists := preReqMap[preReq]; !exists {
				logging.Infof("[PreReq validation] unknown prerequisite: %s", preReq)
			}
			runaiData.Prerequisites = append(runaiData.Prerequisites, preReq)
			existingPreReqs[preReq] = struct{}{}
		}
	}
	for _, preReq := range preReqMap {
		_, exists := existingPreReqs[preReq.GetName()]
		if preReq.IsMandatory() && !exists {
			runaiData.Prerequisites = append(runaiData.Prerequisites, preReq.GetName())
		}
	}
}

func fillAllPreReq(runaiData *model.RunAIInstallData, preReqMap map[string]components.Prerequisite) {
	for _, preReq := range preReqMap {
		if preReq.IsMandatory() {
			runaiData.Prerequisites = append(runaiData.Prerequisites, preReq.GetName())
		}
	}
}
