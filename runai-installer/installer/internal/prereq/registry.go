/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prereq

import (
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/components"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/prereq/ids"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

var (
	prerequisites     map[string]components.Prerequisite
	prerequisitesOnce sync.Once
)

func GetPrerequisites(k8sClient k8sclient.K8sClients) map[string]components.Prerequisite {
	prerequisitesOnce.Do(func() {
		prerequisites = map[string]components.Prerequisite{
			ids.PrerequisiteIdentifierIngress:          components.NewNgnixIngress(k8sClient),
			ids.PrerequisiteIdentifierKnative:          components.NewKnative(k8sClient),
			ids.PrerequisiteIdentifierGpuOperator:      components.NewGpuOperator(k8sClient),
			ids.PrerequisiteIdentifierPrometheus:       components.NewPrometheus(k8sClient),
			ids.PrerequisiteIdentifierTrainingOperator: components.NewTrainingOperator(k8sClient),
			ids.PrerequisiteIdentifierNetworkOperator:  components.NewNetworkOperator(k8sClient),
			ids.PrerequisiteIdentifierLWS:              components.NewLWS(k8sClient),
		}
	})
	return prerequisites
}
