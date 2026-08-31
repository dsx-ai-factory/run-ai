/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package components

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/model"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	k8s "github.com/run-ai/runai/runai-installer/installer/internal/utils/k8sclient"
)

type Prerequisite interface {
	step.InstallationComponent
	GetVersion() string
	IsMandatory() bool
	IsDefaultSelected() bool
	InstalledOnCluster(runaiData *model.RunAIInstallData) model.InstallPhase
}

type PreReqImpl struct {
	name              string
	version           string
	mandatory         bool
	isDefaultSelected bool
	k8sClient         k8s.K8sClients
}

func NewPreReqComponentImpl(name string, version string, mandatory bool, isDefaultSelected bool, k8sClient k8s.K8sClients) *PreReqImpl {
	return &PreReqImpl{
		name:              name,
		version:           version,
		mandatory:         mandatory,
		isDefaultSelected: isDefaultSelected,
		k8sClient:         k8sClient,
	}
}

func (p *PreReqImpl) GetName() string {
	return string(p.name)
}

func (p *PreReqImpl) ComponentType() step.ComponentType {
	return step.ComponentTypePrerequisite
}

func (p *PreReqImpl) GetVersion() string {
	return p.version
}

func (p *PreReqImpl) IsMandatory() bool {
	return p.mandatory
}

func (p *PreReqImpl) IsDefaultSelected() bool {
	return p.isDefaultSelected
}
