/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	configinputs "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/model"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/steps"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type knative struct {
	*PreReqImpl
}

func NewKnative(k8sClient k8s.K8sClients) *knative {
	return &knative{
		PreReqImpl: NewPreReqComponentImpl(ids.PrerequisiteIdentifierKnative, "1.18.2", false, true, k8sClient),
	}
}

func (k *knative) GetMandatoryInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{
		configinputs.InstallInputWildcardFQDN,
		configinputs.InstallInputWildcardCertPath,
		configinputs.InstallInputWildcardCertKeyPath,
	}
}

func (k *knative) GetOptionalInputs(runaiData *model.RunAIInstallData) []configinputs.InstallInputType {
	return []configinputs.InstallInputType{}
}

func (k *knative) GetInstallSteps(k8sClient k8s.K8sClients, helmInstaller step.HelmInstaller, _ step.RunAIAPIClientProvider, runaiData *model.RunAIInstallData) []step.InstallStep {
	installSteps := []step.InstallStep{}
	installSteps = append(installSteps, steps.NewKnativeOperatorInstallStep(k8sClient, k.version, helmInstaller))
	createNameSpaceStep := step.NewCreateNamespace(steps.KnativeServingNamespace, k8sClient)
	installSteps = append(installSteps, createNameSpaceStep)
	inferenceTlsSecretStep := step.NewInstallTlsSecret(steps.RunaiInferenceSecretName,
		steps.KnativeServingNamespace,
		configinputs.InstallInputWildcardCertPath,
		configinputs.InstallInputWildcardCertKeyPath,
		configinputs.InstallInputWildcardFQDN,
		k8sClient)
	installSteps = append(installSteps, inferenceTlsSecretStep)
	installSteps = append(installSteps, steps.NewKnativeServingInstallStep(k8sClient, k.version))
	installSteps = append(installSteps, steps.NewKnativeServingHPAInstallStep(k8sClient, k.version))
	return installSteps
}

func (k *knative) InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase {
	phase1 := steps.NewKnativeOperatorInstallStep(k.k8sClient, k.GetVersion(), step.NewHelmInstaller()).GetStatusSync(runaiData)
	phase2 := steps.NewKnativeServingInstallStep(k.k8sClient, k.GetVersion()).GetStatusSync(runaiData)
	return CombinePhases(phase1, phase2)
}

func CombinePhases(phases ...model.InstallPhase) model.InstallPhase {
	if len(phases) == 0 {
		return model.NotInstalled
	}

	statusSet := make(map[model.InstallPhase]struct{}, len(phases))
	allNotInstalled := true
	for _, p := range phases {
		statusSet[p] = struct{}{}
		if p != model.NotInstalled {
			allNotInstalled = false
		}
	}

	_, hasFailed := statusSet[model.Failed]
	_, hasTimeout := statusSet[model.Timeout]
	_, isInstalling := statusSet[model.Installing]

	switch {
	case hasFailed:
		return model.Failed
	case hasTimeout:
		return model.Failed
	case allNotInstalled:
		return model.NotInstalled
	case isInstalling:
		return model.Installing
	default:
		return model.Installed
	}
}
