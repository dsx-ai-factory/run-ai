/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

type General struct {
	PrefixName      string
	UseNameAsPrefix bool
	Labels          []string
	Annotations     []string
	DryRun          bool
	Template        string
}
