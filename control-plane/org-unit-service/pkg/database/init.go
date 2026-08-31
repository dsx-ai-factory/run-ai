/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func ClearCache() {
	err := cache.ClearCache()
	if err != nil {
		logging.Warnf("failed clearing cache. %+v", err)
	}
}
