/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck_test

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/inputcheck"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
)

// Not one of the two accepted values; every operator that reaches OperatorHub
// has to reject it for the same reason, so the table shares it.
const invalidInstallPlanApproval = "sometimes"

func installPlanApprovalFindings(t *testing.T, s *spec.Spec) []string {
	t.Helper()

	var fields []string

	for _, finding := range inputcheck.Validate(s).Findings {
		if finding.Check == "install-plan-approval" {
			fields = append(fields, finding.Field)
		}
	}

	return fields
}

// Every operator that reaches OperatorHub on OpenShift consumes
// openshift.installPlanApproval, so an unknown value has to be rejected for all
// of them: the plan silently downgrades it to manual.
func TestInstallPlanApprovalCheckCoversEveryOLMOperator(t *testing.T) {
	cases := []struct {
		name  string
		spec  *spec.Spec
		field string
	}{
		{
			name:  "gpu operator",
			spec:  &spec.Spec{Prerequisites: &spec.Prerequisites{GPUOperator: &spec.Dependency{OpenShiftInstallPlanApproval: invalidInstallPlanApproval}}},
			field: "prerequisites.gpuOperator.openshift.installPlanApproval",
		},
		{
			name:  "network operator",
			spec:  &spec.Spec{Prerequisites: &spec.Prerequisites{NetworkOperator: &spec.Dependency{OpenShiftInstallPlanApproval: invalidInstallPlanApproval}}},
			field: "prerequisites.networkOperator.openshift.installPlanApproval",
		},
		{
			name:  "knative",
			spec:  &spec.Spec{Features: &spec.Features{Knative: &spec.Dependency{OpenShiftInstallPlanApproval: invalidInstallPlanApproval}}},
			field: "features.knative.openshift.installPlanApproval",
		},
		{
			name:  "lws",
			spec:  &spec.Spec{Features: &spec.Features{LWS: &spec.Dependency{OpenShiftInstallPlanApproval: invalidInstallPlanApproval}}},
			field: "features.lws.openshift.installPlanApproval",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, []string{tc.field}, installPlanApprovalFindings(t, tc.spec))
		})
	}
}

func TestInstallPlanApprovalCheckAcceptsTheSupportedValues(t *testing.T) {
	for _, approval := range []string{"", spec.OLMApprovalManual, spec.OLMApprovalAutomatic} {
		t.Run("approval="+approval, func(t *testing.T) {
			s := &spec.Spec{Features: &spec.Features{LWS: &spec.Dependency{OpenShiftInstallPlanApproval: approval}}}
			assert.Empty(t, installPlanApprovalFindings(t, s))
		})
	}
}
