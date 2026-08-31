/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package db

import (
	"gorm.io/gorm"
)

func DBMiddleware(errorCallback func(e error)) gorm.Plugin {
	if errorCallback == nil {
		errorCallback = func(e error) {}
	}
	return newGormPrometheusMetricsMiddleware(errorCallback)
}
