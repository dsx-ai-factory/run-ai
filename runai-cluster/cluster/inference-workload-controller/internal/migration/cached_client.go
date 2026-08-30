package migration

import (
	"context"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CachedClient interface {
	Start(ctx context.Context) error
	WaitForCacheSync(ctx context.Context) bool
	IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
	Get(ctx context.Context, key client.ObjectKey, obj client.Object) error
}

type MigrationCachedClient struct {
	k8sCachedClient cache.Cache
}

func NewMigrationCachedClient(cfg *rest.Config, options cache.Options) (CachedClient, error) {
	k8sCachedClient, err := cache.New(cfg, options)
	if err != nil {
		return nil, err
	}

	return &MigrationCachedClient{
		k8sCachedClient: k8sCachedClient,
	}, nil
}

func (m *MigrationCachedClient) Start(ctx context.Context) error {
	return m.k8sCachedClient.Start(ctx)
}

func (m *MigrationCachedClient) WaitForCacheSync(ctx context.Context) bool {
	return m.k8sCachedClient.WaitForCacheSync(ctx)
}

func (m *MigrationCachedClient) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return m.k8sCachedClient.IndexField(ctx, obj, field, extractValue)
}

func (m *MigrationCachedClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return m.k8sCachedClient.List(ctx, list, opts...)
}

func (m *MigrationCachedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return m.k8sCachedClient.Get(ctx, key, obj)
}
