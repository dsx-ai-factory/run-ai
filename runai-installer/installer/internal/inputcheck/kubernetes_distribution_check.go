/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputs/validation"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

func init() { checks = append(checks, kubernetesDistributionCheck) }

func kubernetesDistributionCheck(s *spec.Spec) []Finding {
	return fromResult(
		validation.ValidateKubernetesDistribution(s.KubernetesDistribution),
		"kubernetes-distribution", "install.kubernetesDistribution",
	)
}
