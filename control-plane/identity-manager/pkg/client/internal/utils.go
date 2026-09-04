/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

func isUriContained(existingUriSlice []string, uri string) bool {
	for _, v := range existingUriSlice {
		if v == uri {
			return true
		}
	}
	return false
}
