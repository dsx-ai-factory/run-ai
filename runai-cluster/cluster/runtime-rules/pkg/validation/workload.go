/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation

import "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"

type WorkloadValidator struct {
	validators []Validator
}

func NewWorkloadValidator(validators ...Validator) *WorkloadValidator {
	return &WorkloadValidator{validators: validators}
}

func (wv *WorkloadValidator) Validate(workload *jobs.JobWorkload) error {
	for _, validator := range wv.validators {
		err := validator.Validate(workload)
		if err != nil {
			return err
		}
	}

	return nil
}
