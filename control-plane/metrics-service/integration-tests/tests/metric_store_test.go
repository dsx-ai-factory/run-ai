//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/stretchr/testify/suite"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
)

var (
	random       = rand.New(rand.NewSource(time.Now().UnixNano()))
	datasourceId = "some_datasource_id"
	endPoint     = openapi.MetricStoreEndpoint{Url: "url"}
)

type MetricStoreTestSuite struct {
	suite.Suite
	startTime time.Time
}

func TestMetricStoreTestSuite(t *testing.T) {
	t.Setenv("PUBSUB_TYPE", getEnvOrDefault("PUBSUB_TYPE", "redis"))
	t.Setenv("PUBSUB_URL", getEnvOrDefault("PUBSUB_URL", "localhost:6379"))
	t.Setenv("SERVICE_NAME", getEnvOrDefault("SERVICE_NAME", "metrics-service"))
	t.Setenv("DB_HOST", getEnvOrDefault("DB_HOST", "localhost"))
	t.Setenv("DB_PORT", getEnvOrDefault("DB_PORT", "5432"))
	t.Setenv("DB_NAME", getEnvOrDefault("DB_NAME", "runai"))
	t.Setenv("SECRETS_MOUNT_PATH", getEnvOrDefault("SECRETS_MOUNT_PATH", "../secrets/"))
	t.Setenv("METRICS_SERVER_HOST", getEnvOrDefault("METRICS_SERVER_HOST", "localhost:8080"))
	t.Setenv("DB_SCHEMA", getEnvOrDefault("DB_SCHEMA", "metrics"))
	if err := runaidb.Init(nil); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}

	basic := openapi.NewMetricStoreEndpointAuth()
	basic.SetType(openapi.METRICSTOREENDPOINTAUTHTYPE_BASIC)
	basic.SetBasic(openapi.MetricStoreEndpointAuthBasic{
		Username: "username", Password: "password",
	})
	endPoint.SetAuth(*basic)
	suite.Run(t, new(MetricStoreTestSuite))
}

func (suite *MetricStoreTestSuite) SetupSuite() {
	fmt.Println("Setting up metric store integration tests suite")
	suite.startTime = time.Now()
}

func (suite *MetricStoreTestSuite) TearDownSuite() {
}

func (suite *MetricStoreTestSuite) TestCreateMetricStore() {
	tenantId := random.Int31()
	fmt.Printf("running TestCreateMetricStore, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@example.com")
	tenantMetricStore := getMetricStore(tenantId, datasourceId)

	res, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, tenantMetricStore)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), res)

	res, err = metrics_client.GetMetricStoresClient().GetMetricStore(ctx, tenantId)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), res)
	require.Equal(suite.T(), tenantMetricStore.TenantId, res.TenantId)
	require.Equal(suite.T(), *tenantMetricStore.DatasourceId, *res.DatasourceId)
	sentDetails := tenantMetricStore.MetricStoreDetails.Get()
	gotDetails := res.MetricStoreDetails.Get()
	require.Equal(suite.T(), sentDetails.Type, gotDetails.Type)
	require.Equal(suite.T(), sentDetails.Write.Url, gotDetails.Write.Url)
	require.Equal(suite.T(), *sentDetails.Write.Auth.Get(), *gotDetails.Write.Auth.Get())
	require.Equal(suite.T(), sentDetails.Read.Url, gotDetails.Read.Url)
	require.Equal(suite.T(), *sentDetails.Read.Auth.Get(), *gotDetails.Read.Auth.Get())
}

func (suite *MetricStoreTestSuite) TestCreateMetricStore_TenantMetricStoreAlreadyExist() {
	tenantId := random.Int31()
	fmt.Printf("running TestCreateMetricStore_TenantMetricStoreAlreadyExist, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@example.com")

	res, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, getMetricStore(tenantId, datasourceId))
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), res)

	res, err = metrics_client.GetMetricStoresClient().GetMetricStore(ctx, tenantId)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), datasourceId, *res.DatasourceId)

	res, err = metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, getMetricStore(tenantId, "something_else"))
	require.Error(suite.T(), err)
	require.Nil(suite.T(), res)

	var apiErr *openapi.Error
	require.True(suite.T(), errors.As(err, &apiErr))
	require.Equal(suite.T(), http.StatusConflict, int(apiErr.Code))
	require.True(suite.T(), strings.Contains(apiErr.Message, "409 Conflict"))
}

func (suite *MetricStoreTestSuite) TestCreateMetricStore_NoMetricStore() {
	tenantId := random.Int31()
	fmt.Printf("running TestCreateMetricStore_NoMetricStore, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ms := &openapi.MetricStore{
		TenantId:           int32(tenantId),
		DatasourceId:       ptr.To[string](datasourceId),
		MetricStoreDetails: *openapi.NewNullableMetricStoreDetails(&openapi.MetricStoreDetails{}),
	}

	res, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, ms)
	require.Error(suite.T(), err)
	require.Nil(suite.T(), res)

	var apiErr *openapi.Error
	require.True(suite.T(), errors.As(err, &apiErr))
	require.Equal(suite.T(), http.StatusBadRequest, int(apiErr.Code))
	require.Contains(suite.T(), apiErr.Message, "tenant metric store configured not appropriately: no metric store type")
}

func (suite *MetricStoreTestSuite) TestCreateMetricStore_NoMetricStoreType() {
	tenantId := random.Int31()
	fmt.Printf("running TestCreateMetricStore_NoMetricStoreType, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ms := &openapi.MetricStore{
		TenantId:     int32(tenantId),
		DatasourceId: ptr.To[string](datasourceId),
		MetricStoreDetails: *openapi.NewNullableMetricStoreDetails(&openapi.MetricStoreDetails{
			Read:  endPoint,
			Write: endPoint,
		}),
	}

	res, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, ms)
	require.Error(suite.T(), err)
	require.Nil(suite.T(), res)

	var apiErr *openapi.Error
	require.True(suite.T(), errors.As(err, &apiErr))
	require.Equal(suite.T(), http.StatusBadRequest, int(apiErr.Code))
	require.Contains(suite.T(), apiErr.Message, "tenant metric store configured not appropriately: no metric store type")
}

func (suite *MetricStoreTestSuite) TestGetNoMetricStores() {
	tenantId := random.Int31()
	fmt.Printf("running TestGetNoMetricStores, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	res, err := metrics_client.GetMetricStoresClient().GetMetricStore(ctx, tenantId)
	require.Error(suite.T(), err)
	require.Nil(suite.T(), res)

	var apiErr *openapi.Error
	require.True(suite.T(), errors.As(err, &apiErr))
	require.Equal(suite.T(), http.StatusNotFound, int(apiErr.Code))
	require.Contains(suite.T(), apiErr.Message, "metric store not found")
}

func (suite *MetricStoreTestSuite) TestCreateMetricStore_WithoutDatasourceId() {
	tenantId := random.Int31()
	fmt.Printf("running TestCreateMetricStore_WithoutDatasourceId, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@example.com")
	tenantMetricStore := getMetricStore(tenantId, "")

	created, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, tenantMetricStore)
	require.NoError(suite.T(), err)
	defer metrics_client.GetMetricStoresClient().DeleteMetricStores(ctx, tenantId)
	require.NotNil(suite.T(), created)
	require.Equal(suite.T(), "external", *created.DatasourceId)
}

func (suite *MetricStoreTestSuite) TestCreateMetricStore_WithoutDatasourceAndDetails() {
	tenantId := random.Int31()
	fmt.Printf("running TestCreateMetricStore_WithoutDatasourceAndDetails, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@example.com")

	ms := &openapi.MetricStore{TenantId: tenantId}

	created, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, ms)
	require.NoError(suite.T(), err)
	defer metrics_client.GetMetricStoresClient().DeleteMetricStores(ctx, tenantId)
	require.NotNil(suite.T(), created)
	require.NotNil(suite.T(), created.DatasourceId)
	require.NotEmpty(suite.T(), *created.DatasourceId)
	require.NotEqual(suite.T(), "external", *created.DatasourceId)
	require.NotNil(suite.T(), created.MetricStoreDetails.Get())
}

func (suite *MetricStoreTestSuite) TestDeleteMetricStore() {
	tenantId := random.Int31()
	fmt.Printf("running TestDeleteMetricStore, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@example.com")
	tenantMetricStore := getMetricStore(tenantId, datasourceId)

	metricStore, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, tenantMetricStore)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), metricStore)

	metricStore, err = metrics_client.GetMetricStoresClient().GetMetricStore(ctx, tenantId)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), metricStore)

	err = metrics_client.GetMetricStoresClient().DeleteMetricStores(ctx, tenantId)
	require.NoError(suite.T(), err)

	metricStore, err = metrics_client.GetMetricStoresClient().GetMetricStore(ctx, tenantId)
	require.Error(suite.T(), err)
	require.Nil(suite.T(), metricStore)

	var apiErr *openapi.Error
	require.True(suite.T(), errors.As(err, &apiErr))
	require.Equal(suite.T(), http.StatusNotFound, int(apiErr.Code))
}

func (suite *MetricStoreTestSuite) TestDeleteMetricStore_NotExisting() {
	tenantId := random.Int31()
	fmt.Printf("running TestDeleteMetricStore_NotExisting, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@example.com")

	require.NoError(suite.T(), metrics_client.GetMetricStoresClient().DeleteMetricStores(ctx, tenantId))
}

func (suite *MetricStoreTestSuite) TestTenantDeletePubsubEvent_MetricStoreDeleted() {
	tenantId := random.Int31()
	fmt.Printf("running TestTenantDeletePubsubEvent_MetricStoreDeleted, tenant [%d]\n", tenantId)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@example.com")
	tenantMetricStore := getMetricStore(tenantId, datasourceId)

	res, err := metrics_client.GetMetricStoresClient().CreateMetricStore(ctx, tenantMetricStore)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), res)

	res, err = metrics_client.GetMetricStoresClient().GetMetricStore(ctx, tenantId)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), res)

	err = runaipubsub.Init(nil)
	require.NoError(suite.T(), err)
	tenantDeleteEvent := &events.TenantDelete{
		TenantId: strconv.Itoa(int(tenantId)),
	}
	err = runaipubsub.Publish[events.TenantDelete](ctx, tenantDeleteEvent)
	require.NoError(suite.T(), err)

	require.Eventually(suite.T(), func() bool {
		res, err := metrics_client.GetMetricStoresClient().GetMetricStore(ctx, tenantId)
		if res != nil || err == nil {
			return false
		}
		var apiErr *openapi.Error
		if errors.As(err, &apiErr) {
			return int(apiErr.Code) == http.StatusNotFound
		}
		return false
	}, 15*time.Second, 3*time.Second, "expected error when getting non-existent metric store")
}

func getMetricStore(tenantId int32, datasourceId string) *openapi.MetricStore {
	tenantMetricStore := &openapi.MetricStore{
		TenantId:     int32(tenantId),
		DatasourceId: ptr.To[string](datasourceId),
		MetricStoreDetails: *openapi.NewNullableMetricStoreDetails(&openapi.MetricStoreDetails{
			Type:  openapi.METRICSTORETYPE_MIMIR,
			Read:  endPoint,
			Write: endPoint,
		}),
	}
	return tenantMetricStore
}

func getEnvOrDefault(key, defaultVal string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Trim(value, " ")
	}
	return defaultVal
}
