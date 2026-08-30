package state

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/state/mocks"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/slack"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const tenantId = "1"

func TestGet_EmailEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		EmailEnabled: "true",
	}}, nil)

	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(true)},
		&SlackState{},
	}
	state, err := ncs.Get(ctx, tenantId)

	require.NoError(t, err)
	require.NotNil(t, state)
	require.EqualValues(t, expectedState, state)
}

func TestGet_SlackEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		SlackEnabled:       "true",
		SlackWorkspaceName: "workspace",
	}}, nil)

	expectedState := []notification_state.Stater{
		&EmailState{},
		&SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace")},
	}
	state, err := ncs.Get(ctx, tenantId)

	require.Nil(t, err)
	require.EqualValues(t, expectedState, state)
}

func TestGet_AllEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		EmailEnabled:       "true",
		SlackEnabled:       "true",
		SlackWorkspaceName: "workspace",
	}}, nil)
	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(true)},
		&SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace")},
	}
	state, err := ncs.Get(ctx, tenantId)

	require.Nil(t, err)
	require.EqualValues(t, expectedState, state)
}

func TestGet_AllDisabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		EmailEnabled: "false",
		SlackEnabled: "false",
	}}, nil)
	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
		&SlackState{Enabled: ptr.To(false)},
	}
	state, err := ncs.Get(ctx, tenantId)

	require.Nil(t, err)
	require.EqualValues(t, expectedState, state)
}

func TestGet_RouteDoesntExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}
	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)
	state, err := ncs.Get(ctx, tenantId)

	require.Nil(t, err)
	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
		&SlackState{Enabled: ptr.To(false)},
	}
	require.Equal(t, expectedState, state)
}

func TestGet_FailedToRetrieveCurrentRoute(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("connection error"))
	state, err := ncs.Get(ctx, tenantId)

	require.ErrorContains(t, err, "connection error")
	require.ErrorContains(t, err, notification_state.ErrClientFailure.Error())
	require.Nil(t, state)
}

func TestGet_RouteWithoutMetadata_ReturnsNilPointers(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{}, nil) // metadata is nil

	expectedState := []notification_state.Stater{
		&EmailState{},
		&SlackState{},
	}

	state, err := ncs.Get(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, expectedState, state)
}

func TestGetDetails_EmailEnabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	slackCredsGetterMock := mocks.NewSlackCredentialsGetter(t)
	ncs := Client{tenantRoutesClient: trc, slackCredentialsGetter: slackCredsGetterMock}

	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		EmailEnabled: "true",
	}}, nil)
	slackCredsGetterMock.On("Get", ctx).Return(nil, slack.ErrMissingAppCredentials)

	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(true)},
		&SlackState{AppCredentialsConfigured: ptr.To(false)},
	}
	state, err := ncs.GetDetails(ctx, tenantId)

	require.NoError(t, err)
	require.NotNil(t, state)
	require.EqualValues(t, expectedState, state)
}

func TestGetDetails_RouteDoesntExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	slackCredsGetterMock := mocks.NewSlackCredentialsGetter(t)
	ncs := Client{tenantRoutesClient: trc, slackCredentialsGetter: slackCredsGetterMock}
	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)
	slackCredsGetterMock.On("Get", ctx).Return(nil, slack.ErrMissingAppCredentials)
	state, err := ncs.GetDetails(ctx, tenantId)

	require.Nil(t, err)
	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
		&SlackState{Enabled: ptr.To(false), AppCredentialsConfigured: ptr.To(false)},
	}
	require.Equal(t, expectedState, state)
}

func TestUpdate_FailedToGetCurrentRoute(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, mock.Anything).Return(nil, fmt.Errorf("error"))
	_, err := ncs.UpdateState(ctx, tenantId, []notification_state.Stater{})

	require.ErrorContains(t, err, "failed to retrieve route of tenant 1: error")
}

func TestUpdate_EnableOneChannel(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, mock.Anything).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		"email_enabled": "true",
	}}, nil)
	trc.On("CreateOrUpdateRoute", ctx, &types.BasicRouteConfig{
		Metadata: map[string]string{
			"email_enabled": "true",
			"slack_enabled": "true",
		},
	}).Return(nil)

	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(true)},
		&SlackState{Enabled: ptr.To(true)},
	}
	state, err := ncs.UpdateState(ctx, tenantId, []notification_state.Stater{
		&SlackState{Enabled: ptr.To(true)},
	})

	require.NoError(t, err)
	require.EqualValues(t, expectedState, state)
}

func TestUpdate_DisableOneChannel(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, mock.Anything).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		"email_enabled":        "true",
		"slack_enabled":        "true",
		"slack_workspace_name": "workspace",
	}}, nil)
	trc.On("CreateOrUpdateRoute", ctx, &types.BasicRouteConfig{
		Metadata: map[string]string{
			"email_enabled":        "false",
			"slack_enabled":        "true",
			"slack_workspace_name": "workspace",
		},
	}).Return(nil)

	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
		&SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace")},
	}
	state, err := ncs.UpdateState(ctx, tenantId, []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
	})

	require.NoError(t, err)
	require.EqualValues(t, expectedState, state)
}

func TestUpdate_CurrentRouteDoesntExist(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, mock.Anything).Return(nil, nil)
	expectedRouteToUpdate := getDefaultTenantRoute(tenantId)
	expectedRouteToUpdate.Metadata = map[string]string{
		"email_enabled": "false",
	}
	trc.On("CreateOrUpdateRoute", ctx, expectedRouteToUpdate).Return(nil)

	expectedState := []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
		&SlackState{},
	}

	state, err := ncs.UpdateState(ctx, tenantId, []notification_state.Stater{
		&EmailState{Enabled: ptr.To(false)},
	})

	require.NoError(t, err)
	require.EqualValues(t, expectedState, state)
}

func TestUpdate_FailedToUpdateRoute(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, mock.Anything).Return(&types.BasicRouteConfig{Metadata: map[string]string{}}, nil)
	trc.On("CreateOrUpdateRoute", ctx, mock.Anything).Return(fmt.Errorf("connection error"))
	state, err := ncs.UpdateState(ctx, tenantId, []notification_state.Stater{
		&EmailState{Enabled: ptr.To(true)},
	})

	require.ErrorContains(t, err, "connection error")
	require.Nil(t, state)
}

func TestGetEmail_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		EmailEnabled: "true",
	}}, nil)

	s, err := ncs.GetEmail(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, &EmailState{Enabled: ptr.To(true)}, s)
}

func TestGetEmail_RouteNil_ReturnsDisabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)

	s, err := ncs.GetEmail(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, &EmailState{Enabled: ptr.To(false)}, s)
}

func TestGetEmail_FetchRouteFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("connection error"))

	s, err := ncs.GetEmail(ctx, tenantId)
	require.ErrorContains(t, err, "connection error")
	require.True(t, errors.Is(err, notification_state.ErrClientFailure))
	require.Nil(t, s)
}

func TestUpdateEmailState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	emailState := &EmailState{Enabled: ptr.To(true)}
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{}}, nil)
	trc.On("CreateOrUpdateRoute", ctx, &types.BasicRouteConfig{
		Metadata: map[string]string{EmailEnabled: "true"},
	}).Return(nil)

	s, err := ncs.UpdateEmail(ctx, tenantId, emailState)
	require.NoError(t, err)
	require.Equal(t, emailState, s)
}

func TestUpdateEmailState_GetRouteFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("get route error"))

	s, err := ncs.UpdateEmail(ctx, tenantId, &EmailState{Enabled: ptr.To(true)})
	require.ErrorContains(t, err, "get route error")
	require.True(t, errors.Is(err, notification_state.ErrClientFailure))
	require.Nil(t, s)
}

func TestUpdateEmailState_UpdateRouteFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{}}, nil)
	trc.On("CreateOrUpdateRoute", ctx, mock.Anything).Return(fmt.Errorf("update error"))

	s, err := ncs.UpdateEmail(ctx, tenantId, &EmailState{Enabled: ptr.To(false)})
	require.ErrorContains(t, err, "update error")
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, s)
}

func TestGetSlack_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		SlackEnabled:       "true",
		SlackWorkspaceName: "workspace",
	}}, nil)

	s, err := ncs.GetSlack(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, &SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("workspace")}, s)
}

func TestGetSlack_RouteNil_ReturnsDisabled(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)

	s, err := ncs.GetSlack(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, &SlackState{Enabled: ptr.To(false)}, s)
}

func TestGetSlack_FetchRouteFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("connection error"))

	s, err := ncs.GetSlack(ctx, tenantId)
	require.ErrorContains(t, err, "connection error")
	require.True(t, errors.Is(err, notification_state.ErrClientFailure))
	require.Nil(t, s)
}

func TestGetSlackDetails_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	slackCredsGetterMock := mocks.NewSlackCredentialsGetter(t)
	ncs := Client{tenantRoutesClient: trc, slackCredentialsGetter: slackCredsGetterMock}

	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{
		SlackEnabled: "true",
	}}, nil)
	slackCredsGetterMock.On("Get", ctx).Return(&slack.SlackCredentials{}, nil)

	s, err := ncs.GetSlackDetails(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, &SlackState{Enabled: ptr.To(true), AppCredentialsConfigured: ptr.To(true)}, s)
}

func TestGetSlackDetails_RouteNil_ReturnsDefaultWithSlackCredsState(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	slackCredsGetterMock := mocks.NewSlackCredentialsGetter(t)
	ncs := Client{tenantRoutesClient: trc, slackCredentialsGetter: slackCredsGetterMock}

	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)
	slackCredsGetterMock.On("Get", ctx).Return(nil, slack.ErrMissingAppCredentials)

	s, err := ncs.GetSlackDetails(ctx, tenantId)
	require.NoError(t, err)
	require.Equal(t, &SlackState{Enabled: ptr.To(false), AppCredentialsConfigured: ptr.To(false)}, s)
}

func TestGetSlackDetails_FetchRouteFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	slackCredsGetterMock := mocks.NewSlackCredentialsGetter(t)
	ncs := Client{tenantRoutesClient: trc, slackCredentialsGetter: slackCredsGetterMock}

	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("connection error"))

	s, err := ncs.GetSlackDetails(ctx, tenantId)
	require.ErrorContains(t, err, "connection error")
	require.True(t, errors.Is(err, notification_state.ErrClientFailure))
	require.Nil(t, s)
}

func TestUpdateSlackState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	slackState := &SlackState{Enabled: ptr.To(true), WorkspaceName: ptr.To("ws")}
	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{}}, nil)
	trc.On("CreateOrUpdateRoute", ctx, &types.BasicRouteConfig{
		Metadata: map[string]string{SlackEnabled: "true", SlackWorkspaceName: "ws"},
	}).Return(nil)

	s, err := ncs.UpdateSlack(ctx, tenantId, slackState)
	require.NoError(t, err)
	require.Equal(t, slackState, s)
}

func TestUpdateSlackState_GetRouteFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("get route error"))

	s, err := ncs.UpdateSlack(ctx, tenantId, &SlackState{Enabled: ptr.To(true)})
	require.ErrorContains(t, err, "get route error")
	require.True(t, errors.Is(err, notification_state.ErrClientFailure))
	require.Nil(t, s)
}

func TestUpdateSlackState_UpdateRouteFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	trc := mocks.NewTenantRoutesClienter(t)
	ncs := Client{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(&types.BasicRouteConfig{Metadata: map[string]string{}}, nil)
	trc.On("CreateOrUpdateRoute", ctx, mock.Anything).Return(fmt.Errorf("update error"))

	s, err := ncs.UpdateSlack(ctx, tenantId, &SlackState{Enabled: ptr.To(false)})
	require.ErrorContains(t, err, "update error")
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, s)
}
