/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// out goal in the tests is to simply use the same client as both cached and uncached client.
// we use this wrapper interface as in the real code the cached client is different than the uncached one

type MigrationFakeCachedClient struct {
	k8sCachedClient client.Client
}

func NewMigrationFakeCachedClient(k8sClient client.Client) CachedClient {
	return &MigrationFakeCachedClient{
		k8sCachedClient: k8sClient,
	}
}

func (m *MigrationFakeCachedClient) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (m *MigrationFakeCachedClient) WaitForCacheSync(ctx context.Context) bool {
	return true
}

func (m *MigrationFakeCachedClient) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

func (m *MigrationFakeCachedClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return m.k8sCachedClient.List(ctx, list, opts...)
}

func (m *MigrationFakeCachedClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return m.k8sCachedClient.Get(ctx, key, obj)
}
