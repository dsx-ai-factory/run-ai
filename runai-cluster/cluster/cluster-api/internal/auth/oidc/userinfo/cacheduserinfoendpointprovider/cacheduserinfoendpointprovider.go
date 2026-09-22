/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cacheduserinfoendpointprovider

import (
	"context"
	"time"

	roidc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth/oidc"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth/oidc/configuration"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth/oidc/userinfo"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/rs/zerolog/log"
)

type mapCachedUserInfoEndpointProvider struct {
	oidcConfigProvider configuration.Provider
}

func New(oidcConfigProvider configuration.Provider, expirationPeriod time.Duration) userinfo.Provider {
	err := cache.RegisterStore[roidc.UserInfo](options.Ttl(expirationPeriod))
	if err != nil {
		log.Warn().Err(err).Msg("failed to register user info store to cache")
	}

	return &mapCachedUserInfoEndpointProvider{
		oidcConfigProvider: oidcConfigProvider,
	}
}

func (m *mapCachedUserInfoEndpointProvider) Load(userAccessToken string) (*roidc.UserInfo, error) {
	userInfo, err := cache.Get[roidc.UserInfo](context.Background(), userAccessToken)
	if err != nil {
		log.Warn().Err(err).Msg("failed to get user info from cache")
	}
	if userInfo != nil {
		return userInfo, nil
	}

	userInfo, err = m.oidcConfigProvider.UserInfo(userAccessToken)
	if err != nil {
		return nil, err
	}

	err = cache.Set[roidc.UserInfo](context.Background(), userAccessToken, *userInfo)
	if err != nil {
		log.Warn().Err(err).Msg("failed to set user info in cache")
	}
	return userInfo, nil
}
