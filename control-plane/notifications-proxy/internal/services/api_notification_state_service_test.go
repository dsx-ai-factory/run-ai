/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/state"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/services/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestGetNotificationState_NoState(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}

	managerMock.On("GetNotificationState", ctx, tenantId).Return(nil, nil)
	res, err := ncs.GetNotificationState(ctx)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.NotificationState{}, res.Body)
}

func TestGetNotificationState_SuccessEmailEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	nstate := []notification_state.Stater{&state.EmailState{Enabled: ptr.To(true)}}
	managerMock.On("GetNotificationState", ctx, tenantId).Return(nstate, nil)
	res, err := ncs.GetNotificationState(ctx)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.NotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
	}, res.Body)
}

func TestGetNotificationState_SuccessSlackEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	nstate := []notification_state.Stater{&state.SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace")}}
	managerMock.On("GetNotificationState", ctx, tenantId).Return(nstate, nil)
	res, err := ncs.GetNotificationState(ctx)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.NotificationState{
		Email: api.NotificationStateStatus{},
		Slack: api.NotificationStateStatus{
			Enabled: ptr.To(true),
		},
	}, res.Body)
}

func TestGetNotificationStateWithDetails_SuccessSlackEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	nstate := []notification_state.Stater{&state.SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace"), AppCredentialsConfigured: ptr.To(true)}}
	managerMock.On("GetDetailedNotificationState", ctx, tenantId).Return(nstate, nil)
	res, err := ncs.GetDetailedNotificationState(ctx)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.DetailedNotificationState{
		Email: api.NotificationStateStatus{},
		Slack: api.DetailedNotificationStateSlack{
			Enabled: ptr.To(true),
			Settings: api.SlackDetailsSettings{
				WorkspaceName:            ptr.To("workspace"),
				AppCredentialsConfigured: ptr.To(true)},
		},
	}, res.Body)
}

func TestGetNotificationState_FailedToGetNotificationState(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	managerMock.On("GetNotificationState", ctx, tenantId).Return(nil, errors.Join(errors.New("error"), notification_state.ErrClientFailure))
	res, err := ncs.GetNotificationState(ctx)

	require.ErrorContains(t, err, "request to get notification state failed, internal error")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetNotificationState_Failed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	managerMock.On("GetNotificationState", ctx, tenantId).Return(nil, errors.New("error"))
	res, err := ncs.GetNotificationState(ctx)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetNotificationState_MissingTenantId(t *testing.T) {
	ncs := NotificationStateAPIService{}
	res, err := ncs.GetNotificationState(context.Background())

	require.EqualError(t, err, "missing tenant id in request context")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestGetDetailedNotificationState_SuccessSlackEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	nstate := []notification_state.Stater{&state.EmailState{Enabled: ptr.To(true)}, &state.SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace"), AppCredentialsConfigured: ptr.To(true)}}
	managerMock.On("GetDetailedNotificationState", ctx, tenantId).Return(nstate, nil)
	res, err := ncs.GetDetailedNotificationState(ctx)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.DetailedNotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
		Slack: api.DetailedNotificationStateSlack{Enabled: ptr.To(true), Settings: api.SlackDetailsSettings{WorkspaceName: ptr.To("workspace"), AppCredentialsConfigured: ptr.To(true)}},
	}, res.Body)
}

func TestUpdateNotificationState_MissingTenantId(t *testing.T) {
	ncs := NotificationStateAPIService{}
	res, err := ncs.UpdateNotificationState(context.Background(), api.NotificationState{})

	require.EqualError(t, err, "missing tenant id in request context")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestUpdateNotificationState_FailedToRemoveSinkFromUserRoutes(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	managerMock.On("UpdateNotificationState", ctx, tenantId, mock.Anything).Return(nil, errors.Join(errors.New("failed"), notification_state.ErrFailedToRemoveSinkFromUserRoutes))

	res, err := ncs.UpdateNotificationState(ctx, api.NotificationState{
		Slack: api.NotificationStateStatus{Enabled: ptr.To(true)},
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
	})

	require.ErrorContains(t, err, "failed to remove sink from user routes")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateNotificationState_FailedToUpdateNotificationState(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	managerMock.On("UpdateNotificationState", ctx, tenantId, mock.Anything).Return(nil, errors.Join(errors.New("failed"), notification_state.ErrFailedToUpdateNotificationState))
	res, err := ncs.UpdateNotificationState(ctx, api.NotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
	})

	require.ErrorContains(t, err, "failed to update notification state")
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateNotificationState_NoSinkConfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	managerMock.On("UpdateNotificationState", ctx, tenantId, mock.Anything).Return(nil, errors.Join(errors.New("failed"), notification_state.ErrNoSinkConfigured))
	res, err := ncs.UpdateNotificationState(ctx, api.NotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
	})

	require.ErrorContains(t, err, "no sinks server configured, bad request")
	require.NotNil(t, res)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestUpdateNotificationState_IncompatibleSinkType(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	managerMock.On("UpdateNotificationState", ctx, tenantId, mock.Anything).Return(nil, errors.Join(errors.New("failed"), notification_state.ErrIncompatibleSinkType))

	res, err := ncs.UpdateNotificationState(ctx, api.NotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
	})

	require.ErrorContains(t, err, "incompatible sink type")
	require.NotNil(t, res)
	require.Equal(t, http.StatusConflict, res.Code)
}

func TestUpdateNotificationState_Failed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	managerMock.On("UpdateNotificationState", ctx, tenantId, mock.Anything).Return(nil, errors.New("failed"))
	res, err := ncs.UpdateNotificationState(ctx, api.NotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	managerMock := mocks.NewNotificationStateManager(t)
	ncs := NotificationStateAPIService{manager: managerMock}
	nstate := []notification_state.Stater{&state.EmailState{Enabled: ptr.To(true)}}
	updatedState := []notification_state.Stater{&state.EmailState{Enabled: ptr.To(true)}, &state.SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace-test"), AppCredentialsConfigured: ptr.To(false)}}
	managerMock.On("UpdateNotificationState", ctx, tenantId, nstate).Return(updatedState, nil)
	res, err := ncs.UpdateNotificationState(ctx, api.NotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
	})

	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, api.NotificationState{
		Email: api.NotificationStateStatus{Enabled: ptr.To(true)},
		Slack: api.NotificationStateStatus{Enabled: ptr.To(true)},
	}, res.Body)
}
