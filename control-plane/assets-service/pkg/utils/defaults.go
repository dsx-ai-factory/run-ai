/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import "github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

func ApplyBoolDefault(optionalValue *bool, defaultValue bool) *bool {
	if optionalValue == nil {
		return convutils.Ref(defaultValue)
	}
	return optionalValue
}
