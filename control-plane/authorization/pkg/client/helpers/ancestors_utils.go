/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helpers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func GetAncestorsCache(ctx context.Context, resourceType string, resourceId string) (*Ancestors, error) {
	authAncestors, err := cache.Get[Ancestors](ctx, GetKey(resourceId, resourceType))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got an error trying to fetch the ancestor list of Resource <%s> <%s> from cache. err: %v", resourceType, resourceId, err)
		return nil, fmt.Errorf("could not fetch cache data")
	}

	if authAncestors == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss. Resource <%s> <%s> ancestors are not in cache.", resourceType, resourceId)
		return nil, fmt.Errorf("could not fetch cache data")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache hit. Resource <%s> <%s> ancestors is in cache.", resourceType, resourceId)
	return authAncestors, nil
}

func SetAncestorsCache(ctx context.Context, resourceType string, resourceId string, ancestors Ancestors) {
	err := cache.Set[Ancestors](ctx, GetKey(resourceId, resourceType), ancestors)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set ancestors data to cache, err: %v", err)
	}
}
