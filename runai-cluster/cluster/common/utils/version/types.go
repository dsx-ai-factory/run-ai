/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version

type ComparisonResult int

const (
	EQUAL     ComparisonResult = 0
	LESS      ComparisonResult = -1
	GREATER   ComparisonResult = 1
	UNDEFINED ComparisonResult = -2
)
