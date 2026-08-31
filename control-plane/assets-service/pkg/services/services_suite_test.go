/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	datavolumes_service "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/datavolumes-service"
	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/backend_db"
	kubeClient "github.com/run-ai/runai/control-plane/assets-service/pkg/kube_client"
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	dvapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	dvclientmock "github.com/run-ai/runai/runai-clients/datavolumes-service-client/mocks"
	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupDatavolumeClientWithGetDataVolumesResponse() *dvclientmock.IDatavolumesClient {
	datavolumeClientMock := &dvclientmock.IDatavolumesClient{}
	datavolumes_service.SetDatavolumesMockClient(datavolumeClientMock)

	datavolumeClientMock.
		On("GetDataVolumesNotPaginated", mock.Anything, dvapi.DATAVOLUMEREQUESTTYPE_ORIGINATED,
			"", mock.Anything, "", "").
		Return([]dvapi.DatavolumeNoSharedScopes{}, nil)

	return datavolumeClientMock
}

func TestServices(t *testing.T) {
	config.MockConfig()
	config.Init()
	RegisterFailHandler(Fail)

	//
	//   we need fake client for k8s mock
	//
	fakeClient := NewFakeClient(t)

	kubeClient.SetKubeClientForTest(fakeClient)

	//
	//   we run the service test with "real" database, so need to start it prior
	//   to running any services related test
	//
	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "assets", MigrationDirPath: "file://../database/migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	// tests run against a single-user DB; use the owner role (not runaidb's limited
	// service session) so runtime AutoMigrate can create tables.
	conf.UseServiceSession = new(bool)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	cache_testing.EnableCache()

	backend_db.CreateTables(runaidb.DB().Actual())

	authClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient())

	setupDatavolumeClientWithGetDataVolumesResponse()

	t.Cleanup(func() {
		cache_testing.DisableCache()
	})

	RunSpecs(t, "Services Suite")
}

// create a fake k8s client which holds a set of policies
func NewFakeClient(_ *testing.T, initObjs ...runtime.Object) client.WithWatch {
	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).Build()
}
