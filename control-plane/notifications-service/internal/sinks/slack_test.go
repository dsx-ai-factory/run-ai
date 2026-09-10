/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks/mocks"
	slack_go "github.com/slack-go/slack"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	mockApiToken          = "xoxb-123456789012-1234567890123-12345678901234567890123456789012"
	slackChannelRecipient = "C086FBSPFXT"
	slackUserRecipient    = "U080U9XK7MY"
	recipientIds          = []string{slackChannelRecipient, slackUserRecipient}

	slackConfig = &SlackSinkConfig{
		ApiToken:        mockApiToken,
		Recipients:      recipientIds,
		TemplatesFolder: "../../sinks_templates/slack",
		TemplateFile:    "generic_notification.tpl",
	}
)

func TestNewSlackSinkConfigDefaults(t *testing.T) {
	cfg, err := NewSlackSinkConfig(viper.New(), "id")

	assert.NoError(t, err)
	assert.Equal(t, cfg.ApiToken, "")
	assert.Equal(t, cfg.Recipients, []string{})
	assert.Equal(t, cfg.DirectNotifications, false)
}

func TestSinkFunc_HappyFlow(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)

	conf := &SlackSinkConfig{
		ApiToken:        mockApiToken,
		Recipients:      recipientIds,
		TemplatesFolder: "../../sinks_templates/slack",
		TemplateFile:    "generic_notification.tpl",
	}
	sinker := SlackSink{
		config:      conf,
		slackClient: slackClientMock,
	}

	slackClientMock.On("PostMessage", slackChannelRecipient, mock.Anything).Return(nil)
	slackClientMock.On("PostDMMessage", []string{slackUserRecipient}, mock.Anything).Return(nil)

	err := sinker.Sink(&eventExample)
	assert.NoError(t, err)

	slackClientMock.AssertNotCalled(t, "GetUserByEmail", validUser)
	slackClientMock.AssertCalled(t, "PostMessage", slackChannelRecipient, mock.Anything)
	slackClientMock.AssertCalled(t, "PostDMMessage", []string{slackUserRecipient}, mock.Anything)
}

func TestSinkFunc__DirectNotificationsSetTrue(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)

	conf := &SlackSinkConfig{
		ApiToken:            mockApiToken,
		Recipients:          recipientIds,
		TemplatesFolder:     "../../sinks_templates/slack",
		TemplateFile:        "generic_notification.tpl",
		DirectNotifications: true,
	}
	sinker := SlackSink{
		config:      conf,
		slackClient: slackClientMock,
	}

	slackClientMock.On("GetUserByEmail", validUser).Return(&slack_go.User{ID: "U180U9XK7MY"}, nil)
	slackClientMock.On("PostMessage", slackChannelRecipient, mock.Anything).Return(nil)
	slackClientMock.On("PostDMMessage", []string{slackUserRecipient}, mock.Anything).Return(nil)
	slackClientMock.On("PostDMMessage", []string{"U180U9XK7MY"}, mock.Anything).Return(nil)

	err := sinker.Sink(&eventExample)
	assert.NoError(t, err)

	slackClientMock.AssertCalled(t, "GetUserByEmail", validUser)
	slackClientMock.AssertCalled(t, "PostMessage", slackChannelRecipient, mock.Anything)
	slackClientMock.AssertCalled(t, "PostDMMessage", []string{slackUserRecipient}, mock.Anything)
	slackClientMock.AssertCalled(t, "PostDMMessage", []string{"U180U9XK7MY"}, mock.Anything)
}

func TestSink_EmailRecipients(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)

	conf := &SlackSinkConfig{
		ApiToken:        mockApiToken,
		Recipients:      []string{"user1@example.com", "user2@example.com"},
		TemplatesFolder: "../../sinks_templates/slack",
		TemplateFile:    "generic_notification.tpl",
	}
	sinker := SlackSink{
		config:      conf,
		slackClient: slackClientMock,
	}

	slackClientMock.On("GetUserByEmail", "user1@example.com").Return(&slack_go.User{ID: "U111111111"}, nil)
	slackClientMock.On("GetUserByEmail", "user2@example.com").Return(&slack_go.User{ID: "U222222222"}, nil)
	slackClientMock.On("PostDMMessage", []string{"U111111111"}, mock.Anything).Return(nil)
	slackClientMock.On("PostDMMessage", []string{"U222222222"}, mock.Anything).Return(nil)

	err := sinker.Sink(&eventExample)
	assert.NoError(t, err)
}

func TestSink_MixedRecipients(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)

	conf := &SlackSinkConfig{
		ApiToken:        mockApiToken,
		Recipients:      []string{slackChannelRecipient, "user@example.com"},
		TemplatesFolder: "../../sinks_templates/slack",
		TemplateFile:    "generic_notification.tpl",
	}
	sinker := SlackSink{
		config:      conf,
		slackClient: slackClientMock,
	}

	slackClientMock.On("GetUserByEmail", "user@example.com").Return(&slack_go.User{ID: "U333333333"}, nil)
	slackClientMock.On("PostMessage", slackChannelRecipient, mock.Anything).Return(nil)
	slackClientMock.On("PostDMMessage", []string{"U333333333"}, mock.Anything).Return(nil)

	err := sinker.Sink(&eventExample)
	assert.NoError(t, err)
}

func TestVerifyFuncInvoked_ValidToken_Success(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)
	sinker := SlackSink{
		config:      slackConfig,
		slackClient: slackClientMock,
	}

	slackClientMock.On("AuthTest").Return(&slack_go.AuthTestResponse{}, nil)
	assert.NoError(t, sinker.Verify())
}

func TestVerifyFuncInvoked_InvalidToken_Fail(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)
	sinker := SlackSink{
		config:      &SlackSinkConfig{ApiToken: "invalid token"},
		slackClient: slackClientMock,
	}

	expectedError := errors.New("invalid token")
	slackClientMock.On("AuthTest").Return(nil, expectedError)
	assert.Error(t, sinker.Verify())
}

func TestSinkFunc_NoRecipients(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)

	conf := &SlackSinkConfig{
		TemplatesFolder: "../../sinks_templates/slack",
		TemplateFile:    "generic_notification.tpl",
	}
	sinker := SlackSink{
		config:      conf,
		slackClient: slackClientMock,
	}

	event := events.Event{
		UID: "test",
		Properties: map[string]interface{}{
			"reason": "Scheduled",
		},
	}

	err := sinker.Sink(&event)
	assert.NoError(t, err)

	slackClientMock.AssertNotCalled(t, "PostDMMessage")
	slackClientMock.AssertNotCalled(t, "PostMessage")
}

func TestSinkFunc_InvalidEmailFromEvent(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)

	conf := &SlackSinkConfig{
		TemplatesFolder:     "../../sinks_templates/slack",
		TemplateFile:        "generic_notification.tpl",
		DirectNotifications: true,
	}
	sinker := SlackSink{
		config:      conf,
		slackClient: slackClientMock,
	}

	slackClientMock.On("GetUserByEmail", "invalid@example.com").Return(nil, errors.New("user not found"))
	event := events.Event{
		UID: "test",
		Properties: map[string]interface{}{
			"user_id": "invalid@example.com",
			"reason":  "Scheduled",
		},
	}
	err := sinker.Sink(&event)
	assert.NoError(t, err)

	slackClientMock.AssertCalled(t, "GetUserByEmail", "invalid@example.com")
	slackClientMock.AssertNotCalled(t, "PostDMMessage", mock.Anything, mock.Anything)
}

func TestSendMessageToEmail_CachesResolvedUserId(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)
	sinker := SlackSink{slackClient: slackClientMock}

	email := "cached-user@example.com"
	blocks := []slack_go.Block{}

	t.Cleanup(func() {
		slackUserByEmailCache.InvalidateAll()
	})

	slackClientMock.On("GetUserByEmail", email).Return(&slack_go.User{ID: "U999999999"}, nil).Once()
	slackClientMock.On("PostDMMessage", []string{"U999999999"}, blocks).Return(nil).Twice()

	sinker.sendMessageToEmail(email, blocks)
	sinker.sendMessageToEmail(email, blocks)
}

func TestSendMessageToEmail_FailedLookupNotCached(t *testing.T) {
	slackClientMock := mocks.NewSlackClienter(t)
	sinker := SlackSink{slackClient: slackClientMock}

	email := "failing-user@example.com"
	blocks := []slack_go.Block{}

	t.Cleanup(func() {
		slackUserByEmailCache.InvalidateAll()
	})

	slackClientMock.On("GetUserByEmail", email).Return(nil, errors.New("user not found")).Once()

	sinker.sendMessageToEmail(email, blocks)

	userId, ok := slackUserByEmailCache.GetIfPresent(email)
	assert.Empty(t, userId)
	assert.False(t, ok)
}

func TestIsSlackChannelId_WorksAsExpected(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"C086FBSPFGT", true},
		{"CABCDEFGHA", true},
		{"D082M36118R", false},
		{"", false},
		{"C", false},
		{"C1234@", false},
		{"CaBCDEFGH", false},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, isSlackChannelId(tc.input))
	}
}

func TestIsSlackUserId_WorksAsExpected(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"U080U9XK7MY", true},
		{"W080U9XK7MY", true},
		{"D082M36118R", false},
		{"", false},
		{"U", false},
		{"U1234@", false},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, isSlackUserId(tc.input))
	}
}

func TestIsEmail(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"user@example.com", true},
		{"test@example.com", true},
		{"C086FBSPFGT", false},
		{"U080U9XK7MY", false},
		{"", false},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expected, isEmail(tc.input))
	}
}

func TestConvertDescriptionHTMLToMarkdown(t *testing.T) {
	html := "<h><b>Hello world</b></p><br><p>This is a new paragraph</p>"
	expected := "**Hello world**\n\nThis is a new paragraph"
	result := convertDescriptionHTMLToMarkdown(html)
	require.Equal(t, expected, result)
}
