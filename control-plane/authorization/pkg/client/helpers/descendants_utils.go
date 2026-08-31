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

func GetDescendantsCache(ctx context.Context, resourceType string, resourceId string) (*Descendants, error) {
	authDescendants, err := cache.Get[Descendants](ctx, GetDescendantsKey(resourceId, resourceType))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got an error trying to fetch the descendant list of Resource <%s> <%s> from cache. err: %v", resourceType, resourceId, err)
		return nil, fmt.Errorf("could not fetch cache data")
	}

	if authDescendants == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache miss. Resource <%s> <%s> descendants are not in cache.", resourceType, resourceId)
		return nil, fmt.Errorf("could not fetch cache data")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cache hit. Resource <%s> <%s> descendants is in cache.", resourceType, resourceId)
	return authDescendants, nil
}

func SetDescendantsCache(ctx context.Context, resourceType string, resourceId string, descendants Descendants) {
	err := cache.Set[Descendants](ctx, GetDescendantsKey(resourceId, resourceType), descendants)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set descendants data to cache, err: %v", err)
	}
}
