/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package httputils

import (
	"fmt"
	"os"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// PrintSubmissionWarnings leaves stdout available for the workload name or JSON.
func PrintSubmissionWarnings(warnings []openapi.SubmissionWarning) {
	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "Warning (%s): %s\n", warning.Field, warning.Message)
	}
}
