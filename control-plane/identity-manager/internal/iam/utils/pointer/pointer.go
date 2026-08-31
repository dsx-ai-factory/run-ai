/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pointer

import "time"

func ToString(s string) *string {
	return &s
}

func ToBool(b bool) *bool {
	return &b
}

func ToTime(t time.Time) *time.Time {
	return &t
}
