/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/steps"
)

const nimAdmissionControllerValue = "operator.admissionController.enabled"

func init() { builders = append(builders, nimAdmissionBuilder) }

// nimAdmissionBuilder warns when the NIM Operator is installed with its
// admission controller left enabled (the default), since that webhook needs
// cert-manager. It returns nil when NIM is not installed or when the admission
// controller is explicitly disabled via custom helm values.
func nimAdmissionBuilder(s *spec.Spec, _ orchestrator2.Deps) orchestrator2.Step {
	f := s.Features
	if f == nil || f.Skip || !f.NIMOperator.IsInstall() {
		return nil
	}
	if f.NIMOperator.CustomHelmValues[nimAdmissionControllerValue] == "false" {
		return nil
	}
	return steps.NewActionStepTitled("preflight-nim-admission-webhook", "NIM admission webhook", actions.CheckNIMAdmissionWebhook{})
}
