/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

// StepRunner is a utility class for running sequentially a set of functions that are
// all return workloadsApi.Error. it stops running the set on the first function that
// returns error.
// the functions needs to be wrapped up like this/:
//
//	func f(....) func *workloadsApi.Error {
//	    return func *workloadsApi.Error {
//	       // the function itself
//	    }
//	}
type StepRunner struct {
	err *workloadsApi.Error
}

type StepFunc func() *workloadsApi.Error

// Run executes one given function only if no previous error occured in the runner.
func (s *StepRunner) Run(f func() *workloadsApi.Error) {
	if s.err != nil {
		return
	}
	s.err = f()
}

// Batch runs a batch of functions
func (s *StepRunner) Batch(steps ...StepFunc) {
	for _, step := range steps {
		if s.Error() != nil {
			return // Stop early if previous batch failed
		}
		s.Run(step)
	}
}

// Returns the error that the failing function returned
func (s *StepRunner) Error() *workloadsApi.Error {
	return s.err
}
