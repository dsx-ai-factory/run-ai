package notification_state_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state/mocks"
	statemocks "github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const tenantId = "1"

// TestState is a lightweight implementation of Stater used for unit tests.
type TestState struct {
	Enabled *bool
	Type    string
	Chan    string
	Meta    any
}

func (ts *TestState) IsEnabled() bool {
	return *ts.Enabled
}

func (ts *TestState) GetType() string     { return ts.Type }
func (ts *TestState) GetChanName() string { return ts.Chan }
func (ts *TestState) GetMetadata() any    { return ts.Meta }

func TestUpdateNotificationState_FailedToUpdateSubscription(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, mock.Anything).Return(true, nil)
	stateClientMock.On("UpdateState", ctx, tenantId, mock.Anything).Return(nil, fmt.Errorf("connection error"))

	state, err := nsm.UpdateNotificationState(ctx, tenantId, []notification_state.Stater{&TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}})

	require.ErrorContains(t, err, "connection error")
	require.Nil(t, state)
}

func TestUpdateNotificationState_ChannelMisconfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, mock.Anything).Return(false, fmt.Errorf("incompatible sink type"))

	res, err := nsm.UpdateNotificationState(ctx, tenantId, []notification_state.Stater{&TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}})

	require.Error(t, err)
	require.Nil(t, res)
}

func TestUpdateNotificationState_CallsDisable(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, mock.Anything).Return(true, nil)
	subUserClientMock.On("DisableChannel", ctx, tenantId, mock.Anything).Return(nil)
	stateClientMock.On("UpdateState", ctx, tenantId, mock.Anything).Return([]notification_state.Stater{&TestState{Enabled: ptr.To(false), Type: "email", Chan: "runai-email"}, &TestState{Type: "slack", Chan: "slack-1"}}, nil)

	state, err := nsm.UpdateNotificationState(ctx, tenantId, []notification_state.Stater{&TestState{Enabled: ptr.To(false), Type: "email", Chan: "runai-email"}})

	require.NoError(t, err)
	require.EqualValues(t, []notification_state.Stater{&TestState{Enabled: ptr.To(false), Type: "email", Chan: "runai-email"}, &TestState{Type: "slack", Chan: "slack-1"}}, state)
}

func TestUpdateNotificationState_MigrateFails_ReturnsWrappedError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(fmt.Errorf("migration failed"))

	state, err := nsm.UpdateNotificationState(ctx, tenantId, []notification_state.Stater{&TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}})

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, state)
}

func TestGetNotificationState_MigrateFails_ReturnsWrappedError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(fmt.Errorf("migration failed"))

	state, err := nsm.GetNotificationState(ctx, tenantId)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, state)
}

func TestGetDetailedNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	detailedStates := []notification_state.Stater{
		&TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"},
		&TestState{Enabled: ptr.To(false), Type: "slack", Chan: "slack-1"},
	}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	stateClientMock.On("GetDetails", ctx, tenantId).Return(detailedStates, nil)

	state, err := nsm.GetDetailedNotificationState(ctx, tenantId)

	require.NoError(t, err)
	require.EqualValues(t, detailedStates, state)
}

func TestGetDetailedNotificationState_GetDetailsFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	stateClientMock.On("GetDetails", ctx, tenantId).Return(nil, fmt.Errorf("backend error"))

	state, err := nsm.GetDetailedNotificationState(ctx, tenantId)

	require.ErrorContains(t, err, "backend error")
	require.Nil(t, state)
}

func TestGetDetailedNotificationState_MigrateFails_ReturnsWrappedError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(fmt.Errorf("migration failed"))

	state, err := nsm.GetDetailedNotificationState(ctx, tenantId)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, state)
}

func TestGetEmailNotificationState_MigrateFails_ReturnsWrappedError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(fmt.Errorf("migration failed"))

	s, err := nsm.GetEmailNotificationState(ctx, tenantId)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, s)
}

func TestGetEmailNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	emailState := &TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	stateClientMock.On("GetEmail", ctx, tenantId).Return(emailState, nil)

	s, err := nsm.GetEmailNotificationState(ctx, tenantId)

	require.NoError(t, err)
	require.Equal(t, emailState, s)
}

func TestGetEmailNotificationState_GetEmailFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	stateClientMock.On("GetEmail", ctx, tenantId).Return(nil, fmt.Errorf("backend error"))

	s, err := nsm.GetEmailNotificationState(ctx, tenantId)

	require.ErrorContains(t, err, "backend error")
	require.Nil(t, s)
}

func TestGetSlackNotificationState_MigrateFails_ReturnsWrappedError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(fmt.Errorf("migration failed"))

	s, err := nsm.GetSlackNotificationState(ctx, tenantId)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, s)
}

func TestGetSlackNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	slackState := &TestState{Enabled: ptr.To(false), Type: "slack", Chan: "slack-1"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	stateClientMock.On("GetSlack", ctx, tenantId).Return(slackState, nil)

	s, err := nsm.GetSlackNotificationState(ctx, tenantId)

	require.NoError(t, err)
	require.Equal(t, slackState, s)
}

func TestGetSlackNotificationState_GetSlackFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	stateClientMock.On("GetSlack", ctx, tenantId).Return(nil, fmt.Errorf("backend error"))

	s, err := nsm.GetSlackNotificationState(ctx, tenantId)

	require.ErrorContains(t, err, "backend error")
	require.Nil(t, s)
}

func TestUpdateEmailNotificationState_MigrateFails_ReturnsWrappedError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(fmt.Errorf("migration failed"))

	emailState := &TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}
	s, err := nsm.UpdateEmailNotificationState(ctx, tenantId, emailState)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, s)
}

func TestUpdateEmailNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	emailState := &TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}
	updatedState := &TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, emailState).Return(true, nil)
	stateClientMock.On("UpdateEmail", ctx, tenantId, emailState).Return(updatedState, nil)

	s, err := nsm.UpdateEmailNotificationState(ctx, tenantId, emailState)

	require.NoError(t, err)
	require.Equal(t, updatedState, s)
}

func TestUpdateEmailNotificationState_ChannelMisconfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	emailState := &TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, emailState).Return(false, nil)

	s, err := nsm.UpdateEmailNotificationState(ctx, tenantId, emailState)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrNoSinkConfigured))
	require.Nil(t, s)
}

func TestUpdateEmailNotificationState_CallsDisable(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	emailState := &TestState{Enabled: ptr.To(false), Type: "email", Chan: "runai-email"}
	updatedState := &TestState{Enabled: ptr.To(false), Type: "email", Chan: "runai-email"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, emailState).Return(true, nil)
	subUserClientMock.On("DisableChannel", ctx, tenantId, "runai-email").Return(nil)
	stateClientMock.On("UpdateEmail", ctx, tenantId, emailState).Return(updatedState, nil)

	s, err := nsm.UpdateEmailNotificationState(ctx, tenantId, emailState)

	require.NoError(t, err)
	require.Equal(t, updatedState, s)
}

func TestUpdateEmailNotificationState_UpdateEmailStateFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	emailState := &TestState{Enabled: ptr.To(true), Type: "email", Chan: "runai-email"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, emailState).Return(true, nil)
	stateClientMock.On("UpdateEmail", ctx, tenantId, emailState).Return(nil, fmt.Errorf("backend error"))

	s, err := nsm.UpdateEmailNotificationState(ctx, tenantId, emailState)

	require.ErrorContains(t, err, "backend error")
	require.Nil(t, s)
}

func TestUpdateSlackNotificationState_MigrateFails_ReturnsWrappedError(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	stateMigrator.On("Migrate", ctx, tenantId).Return(fmt.Errorf("migration failed"))

	slackState := &TestState{Enabled: ptr.To(true), Type: "slack", Chan: "slack-1"}
	s, err := nsm.UpdateSlackNotificationState(ctx, tenantId, slackState)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrFailedToUpdateNotificationState))
	require.Nil(t, s)
}

func TestUpdateSlackNotificationState_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	slackState := &TestState{Enabled: ptr.To(true), Type: "slack", Chan: "slack-1"}
	updatedState := &TestState{Enabled: ptr.To(true), Type: "slack", Chan: "slack-1"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, slackState).Return(true, nil)
	stateClientMock.On("UpdateSlack", ctx, tenantId, slackState).Return(updatedState, nil)

	s, err := nsm.UpdateSlackNotificationState(ctx, tenantId, slackState)

	require.NoError(t, err)
	require.Equal(t, updatedState, s)
}

func TestUpdateSlackNotificationState_ChannelMisconfigured(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	slackState := &TestState{Enabled: ptr.To(true), Type: "slack", Chan: "slack-1"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, slackState).Return(false, nil)

	s, err := nsm.UpdateSlackNotificationState(ctx, tenantId, slackState)

	require.Error(t, err)
	require.True(t, errors.Is(err, notification_state.ErrNoSinkConfigured))
	require.Nil(t, s)
}

func TestUpdateSlackNotificationState_CallsDisable(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	slackState := &TestState{Enabled: ptr.To(false), Type: "slack", Chan: "slack-1"}
	updatedState := &TestState{Enabled: ptr.To(false), Type: "slack", Chan: "slack-1"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, slackState).Return(true, nil)
	subUserClientMock.On("DisableChannel", ctx, tenantId, "slack-1").Return(nil)
	stateClientMock.On("UpdateSlack", ctx, tenantId, slackState).Return(updatedState, nil)

	s, err := nsm.UpdateSlackNotificationState(ctx, tenantId, slackState)

	require.NoError(t, err)
	require.Equal(t, updatedState, s)
}

func TestUpdateSlackNotificationState_UpdateSlackStateFails(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	stateClientMock := statemocks.NewStateClienter(t)
	channelClientMock := mocks.NewChannelClienter(t)
	subUserClientMock := mocks.NewSubUserClienter(t)
	stateMigrator := mocks.NewStateMigrationClient(t)
	nsm := notification_state.NewNotificationStateManager(stateClientMock, channelClientMock, subUserClientMock, stateMigrator)

	slackState := &TestState{Enabled: ptr.To(true), Type: "slack", Chan: "slack-1"}
	stateMigrator.On("Migrate", ctx, tenantId).Return(nil)
	channelClientMock.On("IsExists", ctx, tenantId, slackState).Return(true, nil)
	stateClientMock.On("UpdateSlack", ctx, tenantId, slackState).Return(nil, fmt.Errorf("backend error"))

	s, err := nsm.UpdateSlackNotificationState(ctx, tenantId, slackState)

	require.ErrorContains(t, err, "backend error")
	require.Nil(t, s)
}
