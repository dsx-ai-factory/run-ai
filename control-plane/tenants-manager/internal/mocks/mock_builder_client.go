package mocks

import (
	"context"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
)

type MockBuilder struct{}

func NewMockBuilder() *MockBuilder { return &MockBuilder{} }

func (b *MockBuilder) Name() string { return "mock-builder" }

func (b *MockBuilder) Construct(_ context.Context, _ *builders.TenantInfo, _ map[string]interface{}) error {
	return nil
}

func (b *MockBuilder) Destruct(_ context.Context, _ *builders.TenantInfo) error {
	return nil
}
