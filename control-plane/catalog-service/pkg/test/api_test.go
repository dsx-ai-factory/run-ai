/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

const (
	TestTenant     = "1"
	TestTenantUint = 1
	TestTenant2    = "2"
	TestUser       = "user@run.ai"
	TestUser2      = "user2@run.ai"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)

	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "catalog", MigrationDirPath: "file://../database/migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	// Set up mock authorization client
	authorization.SetMock(&SimpleMockFactory{})

	RunSpecs(t, "Registries API tests")
	t.Cleanup(func() {
		CleanTables()
	})
}

func SetContext(tenantId string, subjectId string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)

	return ctx
}

func CleanTables() {
	db := runaidb.DB()
	db.Where("1 = 1").Delete(&models.Registry{})
	db.Where("1 = 1").Delete(&models.HelmRepo{})
}

// SimpleMockFactory implements authorization.FactoryInterface
type SimpleMockFactory struct{}

func (f *SimpleMockFactory) Client() authorization.ClientInterface {
	return &SimpleMockAuthClient{}
}

// SimpleMockAuthClient implements authorization.ClientInterface
type SimpleMockAuthClient struct{}

func (m *SimpleMockAuthClient) CheckPermittedScopes(_ context.Context, _ authClient.ResourceType, _ authClient.Action) (*authClient.PermittedScopes, *authClient.Error) {
	return &authClient.PermittedScopes{
		Tenants: []string{TestTenant, TestTenant2},
	}, nil
}

// NoPermissionsMockFactory returns a mock that denies all permissions
type NoPermissionsMockFactory struct{}

func (f *NoPermissionsMockFactory) Client() authorization.ClientInterface {
	return &NoPermissionsMockClient{}
}

type NoPermissionsMockClient struct{}

func (m *NoPermissionsMockClient) CheckPermittedScopes(_ context.Context, _ authClient.ResourceType, _ authClient.Action) (*authClient.PermittedScopes, *authClient.Error) {
	return &authClient.PermittedScopes{}, nil
}
