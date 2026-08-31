/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helper_job

import "os"

const JobModeArg = "job-mode"

func ShouldRunAsJob() bool {
	for _, arg := range os.Args {
		if arg == JobModeArg {
			return true
		}
	}
	return false
}
