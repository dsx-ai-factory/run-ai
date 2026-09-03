/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaictx

import (
	"context"
)

// the returned value can be passed to logger.WithCtx
// example:
//
//	logger.WithCtx(runaictx.LoggerHook(ctx)).Info()....
func LoggerHook(ctx context.Context) func() (map[interface{}]string, context.Context) {
	mapping := getRunaiCtxMapping()
	return func() (map[interface{}]string, context.Context) { return mapping, ctx }
}
