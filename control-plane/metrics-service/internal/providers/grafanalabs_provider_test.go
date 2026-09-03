/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/grafana"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metricstores"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/providers"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/providers/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newTestInstance() *grafana.Instance {
	return &grafana.Instance{
		Id:     123,
		UserId: 1234,
		Slug:   "runai-test",
		Region: "us-central1",
		Url:    "https://prometheus-us-central1.grafana.net",
	}
}

func TestCreate(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	instance := newTestInstance()
	readPolicy := &grafana.Policy{Id: "read-policy-id", Name: "read-policy"}
	readToken := &grafana.Token{Id: "read-token-id", Token: "read-secret"}
	writePolicy := &grafana.Policy{Id: "write-policy-id", Name: "write-policy"}
	writeToken := &grafana.Token{Id: "write-token-id", Token: "write-secret"}

	clientMock.On("CreateInstance", mock.Anything, "ds-1").Return(instance, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:read", instance).Return(readPolicy, nil)
	clientMock.On("CreateToken", mock.Anything, "us-central1", readPolicy).Return(readToken, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:write", instance).Return(writePolicy, nil)
	clientMock.On("CreateToken", mock.Anything, "us-central1", writePolicy).Return(writeToken, nil)

	ds, err := provider.Create(context.Background(), "ds-1")

	require.NoError(t, err)
	require.Equal(t, &metricstores.Datasource{
		Type: metricstores.DatasourceTypeGrafanaLabs,
		Read: metricstores.DatasourceEndpoint{
			Url: "https://prometheus-us-central1.grafana.net/api/prom",
			Auth: &metricstores.DatasourceAuth{
				Type: metricstores.AuthTypeBasic,
				Basic: &metricstores.AuthBasic{
					User:     "1234",
					Password: "read-secret",
				},
			},
		},
		Write: metricstores.DatasourceEndpoint{
			Url: "https://prometheus-us-central1.grafana.net/api/prom/push",
			Auth: &metricstores.DatasourceAuth{
				Type: metricstores.AuthTypeBasic,
				Basic: &metricstores.AuthBasic{
					User:     "1234",
					Password: "write-secret",
				},
			},
		},
	}, ds)
}

func TestCreate_FailedToCreateInstance(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	clientMock.On("CreateInstance", mock.Anything, "ds-1").Return(nil, fmt.Errorf("failed"))

	ds, err := provider.Create(context.Background(), "ds-1")
	require.EqualError(t, err, "failed to create instance: failed")
	require.Nil(t, ds)
}

func TestCreate_FailedToCreateReadAccessPolicy(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	instance := newTestInstance()
	clientMock.On("CreateInstance", mock.Anything, "ds-1").Return(instance, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:read", instance).Return(nil, fmt.Errorf("failed"))

	ds, err := provider.Create(context.Background(), "ds-1")
	require.EqualError(t, err, "failed to create read access policy: failed")
	require.Nil(t, ds)
}

func TestCreate_FailedToCreateReadToken(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	instance := newTestInstance()
	readPolicy := &grafana.Policy{Id: "read-policy-id", Name: "read-policy"}

	clientMock.On("CreateInstance", mock.Anything, "ds-1").Return(instance, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:read", instance).Return(readPolicy, nil)
	clientMock.On("CreateToken", mock.Anything, "us-central1", readPolicy).Return(nil, fmt.Errorf("failed"))

	ds, err := provider.Create(context.Background(), "ds-1")
	require.EqualError(t, err, "failed to create read token: failed")
	require.Nil(t, ds)
}

func TestCreate_FailedToCreateWriteAccessPolicy(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	instance := newTestInstance()
	readPolicy := &grafana.Policy{Id: "read-policy-id", Name: "read-policy"}
	readToken := &grafana.Token{Id: "read-token-id", Token: "read-secret"}

	clientMock.On("CreateInstance", mock.Anything, "ds-1").Return(instance, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:read", instance).Return(readPolicy, nil)
	clientMock.On("CreateToken", mock.Anything, "us-central1", readPolicy).Return(readToken, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:write", instance).Return(nil, fmt.Errorf("failed"))

	ds, err := provider.Create(context.Background(), "ds-1")
	require.EqualError(t, err, "failed to create write access policy: failed")
	require.Nil(t, ds)
}

func TestCreate_FailedToCreateWriteToken(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	instance := newTestInstance()
	readPolicy := &grafana.Policy{Id: "read-policy-id", Name: "read-policy"}
	readToken := &grafana.Token{Id: "read-token-id", Token: "read-secret"}
	writePolicy := &grafana.Policy{Id: "write-policy-id", Name: "write-policy"}

	clientMock.On("CreateInstance", mock.Anything, "ds-1").Return(instance, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:read", instance).Return(readPolicy, nil)
	clientMock.On("CreateToken", mock.Anything, "us-central1", readPolicy).Return(readToken, nil)
	clientMock.On("CreateAccessPolicy", mock.Anything, "metrics:write", instance).Return(writePolicy, nil)
	clientMock.On("CreateToken", mock.Anything, "us-central1", writePolicy).Return(nil, fmt.Errorf("failed"))

	ds, err := provider.Create(context.Background(), "ds-1")
	require.EqualError(t, err, "failed to create write token: failed")
	require.Nil(t, ds)
}

func TestDelete(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	clientMock.On("DeleteInstance", mock.Anything, "ds-1").Return(nil).Once()
	clientMock.On("DeleteInstance", mock.Anything, "ds-1").Return(grafana.ErrInstanceNotFound).Once()

	err := provider.Delete(context.Background(), "ds-1")
	require.NoError(t, err)
}

func TestDelete_AlreadyGone(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	clientMock.On("DeleteInstance", mock.Anything, "ds-1").Return(grafana.ErrInstanceNotFound).Once()

	err := provider.Delete(context.Background(), "ds-1")
	require.NoError(t, err)
}

func TestDelete_FirstDeleteFails(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	clientMock.On("DeleteInstance", mock.Anything, "ds-1").Return(fmt.Errorf("transient")).Once()

	err := provider.Delete(context.Background(), "ds-1")
	require.EqualError(t, err, "failed to delete instance: transient")
}

func TestDelete_SilentFailureStillReturnsNil(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	clientMock.On("DeleteInstance", mock.Anything, "ds-1").Return(nil).Times(2)

	err := provider.Delete(context.Background(), "ds-1")
	require.NoError(t, err)
}

func TestDelete_VerificationErrorStillReturnsNil(t *testing.T) {
	clientMock := mocks.NewGrafanaLabsClienter(t)
	provider := providers.GrafanaLabsProvider{Client: clientMock}

	clientMock.On("DeleteInstance", mock.Anything, "ds-1").Return(nil).Once()
	clientMock.On("DeleteInstance", mock.Anything, "ds-1").Return(fmt.Errorf("transient")).Once()

	err := provider.Delete(context.Background(), "ds-1")
	require.NoError(t, err)
}
