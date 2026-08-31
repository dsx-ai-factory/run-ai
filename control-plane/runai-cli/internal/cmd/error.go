/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
)

type UnsupportedError struct {
	workloadType api.WorkloadType
	framework    api.Framework
}

func (e UnsupportedError) Error() string {
	return fmt.Sprintf("unsupported framework %s and workload type %s", e.framework, e.workloadType)
}

func NewUnsupportedError(framework api.Framework, workloadType api.WorkloadType) error {
	return &UnsupportedError{
		framework:    framework,
		workloadType: workloadType,
	}
}

type UnsupportedAliasError struct {
	framework    api.Framework
	workloadType api.WorkloadType
}

func (e UnsupportedAliasError) Error() string {
	return fmt.Sprintf("unsupported alias for framework %s and workload type %s", e.framework, e.workloadType)
}

func NewUnsupportedAliasError(framework api.Framework, workloadType api.WorkloadType) error {
	return &UnsupportedAliasError{
		framework:    framework,
		workloadType: workloadType,
	}
}
