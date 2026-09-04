package mocks

import (
	"context"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type MockBffClient struct {
	mu    sync.RWMutex
	logos map[string]string
}

func NewMockBffClient() *MockBffClient { return &MockBffClient{logos: make(map[string]string)} }

func (m *MockBffClient) GetLogo(ctx context.Context, tenantID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.logos[resolveTenantID(ctx, tenantID)], nil
}

func (m *MockBffClient) UpdateLogo(ctx context.Context, tenantID string, logoDataURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logos[resolveTenantID(ctx, tenantID)] = logoDataURL
	return nil
}

func resolveTenantID(ctx context.Context, tenantID string) string {
	if tenantID != "" {
		return tenantID
	}
	if ctxTenant := runaictx.GetString(ctx, runaictx.CfTenantId); ctxTenant != nil {
		return *ctxTenant
	}
	return tenantID
}
