/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cacheduserinfoendpointprovider

import (
	roidc "github.com/dsx-ai-factory/run-ai/runai-cluster/common/runai-auth/providers/oidc"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/runai-auth/providers/oidc/configuration"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/runai-auth/providers/oidc/userinfo"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/safemapcache"
	"time"
)

type mapCachedUserInfoEndpointProvider struct {
	asyncMapCache      *safemapcache.SafeMapCache[string, *roidc.UserInfo]
	oidcConfigProvider configuration.Provider
}

func New(oidcConfigProvider configuration.Provider, expirationPeriod time.Duration) userinfo.Provider {
	return &mapCachedUserInfoEndpointProvider{
		asyncMapCache:      safemapcache.NewWithPeriods[string, *roidc.UserInfo](expirationPeriod, expirationPeriod),
		oidcConfigProvider: oidcConfigProvider,
	}
}

func (m *mapCachedUserInfoEndpointProvider) Load(userAccessToken string) (*roidc.UserInfo, error) {
	value, found, valid := m.asyncMapCache.Load(userAccessToken)
	if !found || !valid {
		userInfo, err := m.oidcConfigProvider.UserInfo(userAccessToken)
		if err != nil {
			return nil, err
		}

		m.asyncMapCache.Store(userAccessToken, userInfo)
		return userInfo, err
	}

	return value, nil
}
