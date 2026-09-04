/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return e.Message
}
