/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type IdpCache interface {
	SetIdpForRealm(ctx context.Context, realm, idpAlias string, idp *iamModels.Idp) error
	GetIdpForRealm(ctx context.Context, realm, idpAlias string, fetchFunc func() (*iamModels.Idp, error)) (*iamModels.Idp, error)
	DeleteIdpFromRealm(ctx context.Context, realm, idpAlias string) error
}

type IdpMappersCache interface {
	SetIdpMappersForRealm(ctx context.Context, realm, idpAlias string, mappers IdpMappers) error
	GetIdpMappersForRealm(ctx context.Context, realm, idpAlias string, fetchFunc func() (*IdpMappers, error)) (*IdpMappers, error)
	DeleteIdpMappersFromRealm(ctx context.Context, realm, idpAlias string) error
}

type IdpsListCache interface {
	SetIdpsListForRealm(ctx context.Context, realm string, idps IdpsList) error
	GetIdpsListForRealm(ctx context.Context, realm string, fetchFunc func() (*IdpsList, error)) (*IdpsList, error)
	DeleteIdpsListFromRealm(ctx context.Context, realm string) error
}

type IdpCacheManager struct {
	KeyBuilder *KeyBuilder
	Manager    CacheManagerInterface[iamModels.Idp]
}

func NewIdpCacheManager() *IdpCacheManager {
	return &IdpCacheManager{
		KeyBuilder: NewKeyBuilder(),
		Manager:    NewCacheManager[iamModels.Idp](),
	}
}

func (im *IdpCacheManager) SetIdpForRealm(ctx context.Context, realm, idpAlias string, idp *iamModels.Idp) error {
	key := im.KeyBuilder.BuildIdpKey(realm, idpAlias)
	return im.Manager.Set(ctx, key, *idp)
}

func (im *IdpCacheManager) GetIdpForRealm(ctx context.Context, realm, idpAlias string, fetchFunc func() (*iamModels.Idp, error)) (*iamModels.Idp, error) {
	key := im.KeyBuilder.BuildIdpKey(realm, idpAlias)
	return im.Manager.GetOrFetch(ctx, key, fetchFunc)
}

func (im *IdpCacheManager) DeleteIdpFromRealm(ctx context.Context, realm, idpAlias string) error {
	key := im.KeyBuilder.BuildIdpKey(realm, idpAlias)
	return im.Manager.Delete(ctx, key)
}

type IdpMappers struct {
	Mappers []*iamModels.IdpMapper
}
type IdpMappersCacheManager struct {
	KeyBuilder *KeyBuilder
	Manager    CacheManagerInterface[IdpMappers]
}

func NewIdpMappersCacheManager() *IdpMappersCacheManager {
	return &IdpMappersCacheManager{
		KeyBuilder: NewKeyBuilder(),
		Manager:    NewCacheManager[IdpMappers](),
	}
}

func (imm *IdpMappersCacheManager) SetIdpMappersForRealm(ctx context.Context, realm, idpAlias string, mappers IdpMappers) error {
	key := imm.KeyBuilder.BuildIdpMappersKey(realm, idpAlias)
	return imm.Manager.Set(ctx, key, mappers)
}

func (imm *IdpMappersCacheManager) GetIdpMappersForRealm(ctx context.Context, realm, idpAlias string, fetchFunc func() (*IdpMappers, error)) (*IdpMappers, error) {
	key := imm.KeyBuilder.BuildIdpMappersKey(realm, idpAlias)
	return imm.Manager.GetOrFetch(ctx, key, fetchFunc)
}

func (imm *IdpMappersCacheManager) DeleteIdpMappersFromRealm(ctx context.Context, realm, idpAlias string) error {
	key := imm.KeyBuilder.BuildIdpMappersKey(realm, idpAlias)
	return imm.Manager.Delete(ctx, key)
}

type IdpsList struct {
	Idps []*iamModels.Idp
}
type IdpsListCacheManager struct {
	KeyBuilder *KeyBuilder
	Manager    CacheManagerInterface[IdpsList]
}

func NewIdpsListCacheManager() *IdpsListCacheManager {
	return &IdpsListCacheManager{
		KeyBuilder: NewKeyBuilder(),
		Manager:    NewCacheManager[IdpsList](),
	}
}

func (ilm *IdpsListCacheManager) SetIdpsListForRealm(ctx context.Context, realm string, idps IdpsList) error {
	key := ilm.KeyBuilder.BuildIdpsListKey(realm)
	return ilm.Manager.Set(ctx, key, idps)
}

func (ilm *IdpsListCacheManager) GetIdpsListForRealm(ctx context.Context, realm string, fetchFunc func() (*IdpsList, error)) (*IdpsList, error) {
	key := ilm.KeyBuilder.BuildIdpsListKey(realm)
	return ilm.Manager.GetOrFetch(ctx, key, fetchFunc)
}

func (ilm *IdpsListCacheManager) DeleteIdpsListFromRealm(ctx context.Context, realm string) error {
	key := ilm.KeyBuilder.BuildIdpsListKey(realm)
	return ilm.Manager.Delete(ctx, key)
}

type IdpCacheWrapper struct {
	IdpCache        IdpCache
	IdpMappersCache IdpMappersCache
	IdpsListCache   IdpsListCache
}

func NewIdpCacheWrapper() *IdpCacheWrapper {
	return &IdpCacheWrapper{
		IdpCache:        NewIdpCacheManager(),
		IdpMappersCache: NewIdpMappersCacheManager(),
		IdpsListCache:   NewIdpsListCacheManager(),
	}
}

func (icw *IdpCacheWrapper) InvalidateCache(ctx context.Context, realm, idpAlias string) {
	idpErr := icw.IdpCache.DeleteIdpFromRealm(ctx, realm, idpAlias)
	idpMappersErr := icw.IdpMappersCache.DeleteIdpMappersFromRealm(ctx, realm, idpAlias)
	idpsListErr := icw.IdpsListCache.DeleteIdpsListFromRealm(ctx, realm)
	if idpErr != nil || idpMappersErr != nil || idpsListErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"failed to delete idp from cache for realm %s and idp alias %s, "+
				"idp error: %v, idp mappers error: %v, idps list error: %v",
			realm, idpAlias, idpErr, idpMappersErr, idpsListErr,
		)
	}
}
