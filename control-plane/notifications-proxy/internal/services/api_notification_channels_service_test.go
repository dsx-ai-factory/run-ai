package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/channel"
	channelClient "github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/channel"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/services/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	tenantId = "1"
	ctx      = context.WithValue(context.WithValue(context.Background(), runaictx.CfTenantId, tenantId), runaictx.CfSubjectID, user)
	cfgEmail = api.NotificationChannel{
		Config: api.NotificationChannelConfig{},
		Type:   "email",
	}
	cfgSlack = api.NotificationChannel{
		Config: api.NotificationChannelConfig{
			ApiToken: "xoxp-api-token",
		},
		Type: "slack",
	}
)

func initNotificationsChannelServiceForTesting(t *testing.T) (*NotificationChannelsAPIService, *mocks.ChannelsManager, *mocks.Authorizer) {
	authMock := mocks.NewAuthorizer(t)
	channelsManager := mocks.NewChannelsManager(t)
	return &NotificationChannelsAPIService{channelsManager: channelsManager, authorizer: authMock}, channelsManager, authMock
}

func TestCreateNotificationChannel_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)
	channelsManager.On("CreateChannel", ctx, tenantId, mock.Anything).Return(nil)
	res, err := ncs.CreateNotificationChannel(ctx, cfgEmail)

	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestCreateNotificationChannel_AlreadyExist(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)
	channelsManager.On("CreateChannel", ctx, tenantId, mock.Anything).Return(channel.ErrAlreadyExists)
	res, err := ncs.CreateNotificationChannel(ctx, cfgEmail)

	require.NotNil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusConflict, res.Code)
}

func TestCreateNotificationChannel_FailedToCreate(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)
	channelsManager.On("CreateChannel", ctx, tenantId, mock.Anything).Return(channel.ErrClientFailure)
	res, err := ncs.CreateNotificationChannel(ctx, cfgEmail)

	require.NotNil(t, err)
	require.NotNil(t, res)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestCreateNotificationChannelSlack_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsTenant", ctx).Return(true)

	channelsManager.On("CreateChannel", ctx, tenantId, mock.Anything).Return(nil)
	res, err := ncs.CreateNotificationChannel(ctx, cfgSlack)

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestCreateNotificationChannel_NoPermissions(t *testing.T) {
	ncs, _, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsTenant", ctx).Return(false)
	res, _ := ncs.CreateNotificationChannel(ctx, cfgSlack)

	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestDeleteNotificationChannel_Failed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)

	authMock.On("IsSystem", ctx).Return(true)
	channelsManager.On("DeleteChannel", ctx, tenantId, channelClient.EmailChannelName()).Return(errors.Join(fmt.Errorf("failed to delete sink"), channel.ErrClientFailure))

	res, err := ncs.DeleteNotificationChannel(ctx, channelClient.EmailChannelName())

	require.NotNil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestDeleteNotificationChannel_FailedNoPermissions(t *testing.T) {
	ncs, _, _ := initNotificationsChannelServiceForTesting(t)

	res, _ := ncs.DeleteNotificationChannel(ctx, "invalid-channelClient-id")
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestDeleteNotificationChannel_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)

	channelsManager.On("DeleteChannel", ctx, tenantId, channelClient.EmailChannelName()).Return(nil)
	authMock.On("IsSystem", ctx).Return(true)

	res, err := ncs.DeleteNotificationChannel(ctx, channelClient.EmailChannelName())

	require.Nil(t, err)
	require.Equal(t, http.StatusNoContent, res.Code)
}

func TestUpdateNotificationChannel_FailedToUpdate(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)

	channelsManager.On("UpdateChannel", ctx, tenantId, channelClient.EmailChannelName(), mock.Anything).Return(fmt.Errorf("failed to update"))
	res, err := ncs.UpdateNotificationChannel(ctx, channelClient.EmailChannelName(), cfgEmail)

	require.NotNil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateNotificationChannel_FailedNoPermissions(t *testing.T) {
	ncs, _, _ := initNotificationsChannelServiceForTesting(t)

	res, _ := ncs.UpdateNotificationChannel(ctx, "invalid-channelClient-id", cfgEmail)
	require.Equal(t, http.StatusBadRequest, res.Code)
}

func TestUpdateNotificationChannelEmail_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)

	channelsManager.On("UpdateChannel", ctx, tenantId, channelClient.EmailChannelName(), mock.Anything).Return(nil)
	res, err := ncs.UpdateNotificationChannel(ctx, channelClient.EmailChannelName(), cfgEmail)

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestUpdateNotificationChannelSlack_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsTenant", ctx).Return(true)

	channelsManager.On("UpdateChannel", ctx, tenantId, channelClient.SlackChannelName(tenantId), mock.Anything).Return(nil)
	res, err := ncs.UpdateNotificationChannel(ctx, channelClient.SlackChannelName(tenantId), cfgSlack)

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestUpdateNotificationChannel_ChannelDoesNotExist(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsTenant", ctx).Return(true)

	channelsManager.On("UpdateChannel", ctx, tenantId, channelClient.SlackChannelName(tenantId), mock.Anything).Return(channel.ErrNotFound)
	res, err := ncs.UpdateNotificationChannel(ctx, channelClient.SlackChannelName(tenantId), cfgSlack)

	require.EqualError(t, err, "channel not found")
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestPatchNotificationChannel_NotExists(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)

	channelsManager.On("PatchChannel", ctx, tenantId, channelClient.EmailChannelName(), mock.Anything).Return(channel.ErrNotFound)
	res, err := ncs.PatchNotificationChannel(ctx, channelClient.EmailChannelName(), api.NotificationChannelForPatch{Config: cfgEmail.Config, Type: cfgEmail.Type})

	require.NotNil(t, err)
	require.Equal(t, http.StatusNotFound, res.Code)
}

func TestPatchNotificationChannel_FailedToUpdate(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)

	channelsManager.On("PatchChannel", ctx, tenantId, channelClient.EmailChannelName(), mock.Anything).Return(fmt.Errorf("failed to update"))
	res, err := ncs.PatchNotificationChannel(ctx, channelClient.EmailChannelName(), api.NotificationChannelForPatch{Config: cfgEmail.Config, Type: cfgEmail.Type})

	require.NotNil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestPatchNotificationChannel_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)

	channelsManager.On("PatchChannel", ctx, tenantId, channelClient.EmailChannelName(), mock.Anything).Return(nil)
	res, err := ncs.PatchNotificationChannel(ctx, channelClient.EmailChannelName(), api.NotificationChannelForPatch{Config: cfgEmail.Config, Type: cfgEmail.Type})

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
}

func TestValidateNotificationChannelConfig_FailedToCheck(t *testing.T) {
	ncs, channelsManager, _ := initNotificationsChannelServiceForTesting(t)

	channelsManager.On("ValidateChannel", ctx, tenantId, mock.Anything).Return(channel.ValidationResponse{Status: false}, fmt.Errorf("failed to connect to service"))
	res, err := ncs.ValidateNotificationChannel(ctx, api.NotificationChannelForValidate{
		Config: cfgEmail.Config,
		Type:   cfgEmail.Type,
	})

	require.NotNil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestValidateNotificationChannelConfig_Valid(t *testing.T) {
	ncs, channelsManager, _ := initNotificationsChannelServiceForTesting(t)

	channelsManager.On("ValidateChannel", ctx, tenantId, mock.Anything).Return(channel.ValidationResponse{Status: true}, nil)
	res, err := ncs.ValidateNotificationChannel(ctx, api.NotificationChannelForValidate{
		Config: cfgEmail.Config,
		Type:   cfgEmail.Type,
	})

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.True(t, res.Body.(bool))
}

func TestValidateNotificationChannelConfig_Invalid(t *testing.T) {
	ncs, channelsManager, _ := initNotificationsChannelServiceForTesting(t)

	channelsManager.On("ValidateChannel", ctx, tenantId, mock.Anything).Return(channel.ValidationResponse{Status: false}, nil)
	res, err := ncs.ValidateNotificationChannel(ctx, api.NotificationChannelForValidate{
		Config: cfgEmail.Config,
		Type:   cfgEmail.Type,
	})

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.False(t, res.Body.(bool))
}

func TestValidateNotificationChannelConfig_ValidationFailed(t *testing.T) {
	ncs, channelsManager, _ := initNotificationsChannelServiceForTesting(t)

	channelsManager.On("ValidateChannel", ctx, tenantId, mock.Anything).Return(channel.ValidationResponse{Status: false}, channel.ErrValidationFailed)
	res, err := ncs.ValidateNotificationChannel(ctx, api.NotificationChannelForValidate{
		Config: cfgEmail.Config,
		Type:   cfgEmail.Type,
	})

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.False(t, res.Body.(bool))
}

func TestGetNotificationChannel_Failed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)

	channelsManager.On("GetChannel", ctx, tenantId, channelClient.EmailChannelName()).Return(nil, fmt.Errorf("failed to connect"))
	res, err := ncs.GetNotificationChannel(ctx, channelClient.EmailChannelName())

	require.NotNil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetNotificationChannel_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsSystem", ctx).Return(true)

	channelsManager.On("GetChannel", ctx, tenantId, channelClient.EmailChannelName()).Return(&channel.NotificationChannel{Type: "email"}, nil)
	res, err := ncs.GetNotificationChannel(ctx, channelClient.EmailChannelName())

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, res.Body.(*api.NotificationChannelResponse).Type, api.NOTIFICATIONCHANNELTYPE_EMAIL)
}

func TestGetNotificationChannel_Slack_Succeed(t *testing.T) {
	ncs, channelsManager, authMock := initNotificationsChannelServiceForTesting(t)
	authMock.On("IsTenant", ctx).Return(true)

	channelsManager.On("GetChannel", ctx, tenantId, channelClient.SlackChannelName(tenantId)).Return(&channel.NotificationChannel{Type: "slack"}, nil)
	res, err := ncs.GetNotificationChannel(ctx, channelClient.SlackChannelName(tenantId))

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, res.Body.(*api.NotificationChannelResponse).Type, api.NOTIFICATIONCHANNELTYPE_SLACK)
}

func TestGetNotificationChannels_Failed(t *testing.T) {
	ncs, channelsManager, _ := initNotificationsChannelServiceForTesting(t)

	channelsManager.On("GetNotificationsChannels", ctx, tenantId).Return(nil, errors.Join(fmt.Errorf("failed to connect"), channel.ErrClientFailure))
	res, err := ncs.GetNotificationChannels(ctx)

	require.NotNil(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestGetNotificationChannels_Succeed(t *testing.T) {
	ncs, channelsManager, _ := initNotificationsChannelServiceForTesting(t)
	sinkConfig := channel.NotificationChannelMetadata{
		Id:   channelClient.EmailChannelName(),
		Type: "email",
	}

	channelsManager.On("GetNotificationsChannels", ctx, tenantId).Return([]channel.NotificationChannelMetadata{sinkConfig}, nil)
	res, err := ncs.GetNotificationChannels(ctx)
	expected := []*api.NotificationChannelMeta{{
		Type: string(sinkConfig.Type),
		Name: sinkConfig.Id,
	}}

	require.Nil(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, expected, res.Body)
}

func TestAllSinksToChannelMeta(t *testing.T) {
	sinkConfigs := []channel.NotificationChannelMetadata{
		{
			Id:   "id",
			Type: "email",
		},
		{
			Id:   "id1",
			Type: "email",
		},
	}
	res := channelsMetadatToApi(sinkConfigs)
	expected := []*api.NotificationChannelMeta{
		{
			Type: string(sinkConfigs[0].Type),
			Name: sinkConfigs[0].Id,
		},
		{
			Type: string(sinkConfigs[1].Type),
			Name: sinkConfigs[1].Id,
		},
	}
	require.Equal(t, expected, res)
}
