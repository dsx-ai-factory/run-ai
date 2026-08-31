/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import "errors"

type WorkloadPolicyValidationError struct {
	Reason string
}

func (self *WorkloadPolicyValidationError) Error() string {
	return self.Reason
}

func NewWorkloadPolicyValidationError(reason string) error {
	return &WorkloadPolicyValidationError{
		Reason: reason,
	}

}

func IsWorkloadPolicyValidationError(err error) bool {
	var workloadPolicyValidationError *WorkloadPolicyValidationError
	ok := errors.As(err, &workloadPolicyValidationError)
	return ok
}
