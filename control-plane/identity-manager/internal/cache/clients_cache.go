package cache

import (
	"context"

	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
)

//go:generate mockery --name ClientCache --output ./mocks --filename clients_cache.go
type ClientCache interface {
	SetClientForRealm(ctx context.Context, realm, clientId string, client iamModels.Client) error
	GetClientForRealm(ctx context.Context, realm, clientId string, fetchFunc func() (*iamModels.Client, error)) (*iamModels.Client, error)
	DeleteClientFromRealm(ctx context.Context, realm, clientId string) error
}

type ClientCacheManager struct {
	KeyBuilder *KeyBuilder
	Manager    CacheManagerInterface[iamModels.Client]
}

func NewClientCacheManager() *ClientCacheManager {
	return &ClientCacheManager{
		KeyBuilder: NewKeyBuilder(),
		Manager:    NewCacheManager[iamModels.Client](),
	}
}

func (cm *ClientCacheManager) SetClientForRealm(ctx context.Context, realm, clientId string, client iamModels.Client) error {
	cacheKey := cm.KeyBuilder.BuildClientKey(realm, clientId)
	return cm.Manager.Set(ctx, cacheKey, client)
}

func (cm *ClientCacheManager) GetClientForRealm(ctx context.Context, realm, clientId string, fetchFunc func() (*iamModels.Client, error)) (*iamModels.Client, error) {
	cacheKey := cm.KeyBuilder.BuildClientKey(realm, clientId)
	return cm.Manager.GetOrFetch(ctx, cacheKey, fetchFunc)
}

func (cm *ClientCacheManager) DeleteClientFromRealm(ctx context.Context, realm, clientId string) error {
	cacheKey := cm.KeyBuilder.BuildClientKey(realm, clientId)
	return cm.Manager.Delete(ctx, cacheKey)
}
