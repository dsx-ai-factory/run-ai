/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validator

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/validator/gvk"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Validator interface {
	Name() string
	Validate(ctx context.Context, existingRunner, desiredRunner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, field.ErrorList)
}

type ControllerValidator struct {
	validators []Validator
}

func NewControllerValidator() *ControllerValidator {
	return &ControllerValidator{validators: []Validator{gvk.NewGVKValidator()}}
}

func (v *ControllerValidator) Run(ctx context.Context, existingRunner, desiredRunner *dromosv1alpha1.WorkloadRunner) (admission.Warnings, field.ErrorList) {
	logger := log.FromContext(ctx)

	logger.Info("Running ControllerValidator", "validators", len(v.validators))

	var allWarnings admission.Warnings
	var allErrors field.ErrorList

	for _, validator := range v.validators {
		warnings, errors := validator.Validate(ctx, existingRunner, desiredRunner)
		if len(warnings) > 0 {
			logger.Info("Validation warnings", "validator", validator.Name(), "warnings", warnings)
			allWarnings = append(allWarnings, warnings...)
		}
		if len(errors) > 0 {
			logger.Error(errors.ToAggregate(), "Validation errors", "validator", validator.Name(), "errors", errors)
			allErrors = append(allErrors, errors...)
		}
	}

	return allWarnings, allErrors
}
