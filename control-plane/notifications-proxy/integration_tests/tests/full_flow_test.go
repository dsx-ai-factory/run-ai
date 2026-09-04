//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/integration_tests/mailhog/mailhogclient"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/integration_tests/tests/smtp"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const (
	SmtpHost    = "smtp"
	SmtpApiPort = 8025
	SmtpPort    = 1025
	ListenerId  = "http-runai"
)

func TestShouldSendEmail_succeed(t *testing.T) {
	channelId := "runai-email"
	mailTo := uuid.New().String() + "@run.ai"
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "1")
	smtpPort := int32(SmtpPort)
	channelConfig := &openapi.NotificationChannel{
		Type: "email",
		Config: openapi.NotificationChannelConfig{
			Username:       ptr.To("user"),
			Password:       ptr.To("password"),
			SmtpTlsEnabled: *openapi.NewNullableBool(ptr.To(false)),
			SmtpHost:       ptr.To(SmtpHost),
			SmtpPort:       &smtpPort,
			AuthType:       openapi.EMAILAUTHTYPE_AUTH_LOGIN.Ptr(),
		},
	}

	_, res, err := client.NotificationChannelsAPI.ValidateNotificationChannel(ctx).NotificationChannelForValidate(openapi.NotificationChannelForValidate{
		Type: "email", Config: channelConfig.Config,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, res.StatusCode, http.StatusOK)

	_, err = client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(*channelConfig).Execute()
	require.NoError(t, err)

	// patch should not remove configuration credentials
	client.NotificationChannelsAPI.PatchNotificationChannel(ctx, channelId).NotificationChannelForPatch(openapi.NotificationChannelForPatch{
		Type: openapi.NOTIFICATIONCHANNELTYPE_EMAIL.Ptr(),
		Config: &openapi.NotificationChannelConfig{
			Recipients: []string{mailTo},
		}}).Execute()

	routeId := "route_" + uuid.New().String()
	enabled := true
	_, _, err = client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Email: &openapi.NotificationStateStatus{Enabled: *openapi.NewNullableBool(&enabled)},
	}).Execute()
	require.NoError(t, err)

	_, _, err = client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events: []openapi.SubscriptionEvent{
			{
				Category: "Workloads",
				Types:    []string{"All"},
			},
		},
		Channels: []string{channelId},
		Id:       &routeId,
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, routeId).Execute()

	RunEventsEmitter(5000, "1", 10, map[string]interface{}{"category": "Workloads", "type": "All", "user_id": "1", "user": mailTo})
	waitForTheEmail(t, mailTo, 10)

	_, _ = client.NotificationChannelsAPI.DeleteNotificationChannel(ctx, channelId).Execute()
}

func TestShouldNotSendEmail(t *testing.T) {
	channelId := "runai-email"
	mailTo := "testShouldNotSendEmail_" + uuid.New().String() + "@run.ai"
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, mailTo)
	smtpPort := int32(SmtpPort)
	channelConfig := &openapi.NotificationChannel{
		Type: "email",
		Config: openapi.NotificationChannelConfig{
			Username:            ptr.To("user"),
			Password:            ptr.To("password"),
			SmtpTlsEnabled:      *openapi.NewNullableBool(ptr.To(false)),
			SmtpHost:            ptr.To(SmtpHost),
			SmtpPort:            &smtpPort,
			AuthType:            openapi.EMAILAUTHTYPE_AUTH_LOGIN.Ptr(),
			DirectNotifications: *openapi.NewNullableBool(ptr.To(true)),
		},
	}

	_, res, err := client.NotificationChannelsAPI.ValidateNotificationChannel(ctx).NotificationChannelForValidate(openapi.NotificationChannelForValidate{
		Type: "email", Config: channelConfig.Config,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, res.StatusCode, http.StatusOK)

	client.NotificationChannelsAPI.DeleteNotificationChannel(ctx, channelId).Execute()
	_, err = client.NotificationChannelsAPI.CreateNotificationChannel(ctx).NotificationChannel(*channelConfig).Execute()
	require.NoError(t, err)

	routeId := "route_" + uuid.New().String()
	enabled := true
	_, _, err = client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Email: &openapi.NotificationStateStatus{Enabled: *openapi.NewNullableBool(&enabled)},
	}).Execute()

	_, _, err = client.SubscriptionsAPI.CreateSubscription(ctx).Subscription(openapi.Subscription{
		Events: []openapi.SubscriptionEvent{
			{
				Category: "Workloads",
				Types:    []string{"All"},
			},
		},
		Channels: []string{channelId},
		Id:       &routeId,
	}).Execute()
	require.NoError(t, err)
	defer client.SubscriptionsAPI.DeleteSubscription(ctx, routeId).Execute()

	require.NoError(t, err)
	enabled = false
	_, _, err = client.NotificationStateAPI.UpdateNotificationState(ctx).NotificationState(openapi.NotificationState{
		Email: &openapi.NotificationStateStatus{Enabled: *openapi.NewNullableBool(&enabled)},
	}).Execute()
	require.NoError(t, err)

	RunEventsEmitter(5000, "1", 10, map[string]interface{}{"category": "Workloads", "type": "All", "user_id": mailTo, "user": mailTo})
	validateNoEmail(t, mailTo)

	_, _ = client.NotificationChannelsAPI.DeleteNotificationChannel(ctx, channelId).Execute()
}

func validateNoEmail(t *testing.T, mailTo string) {
	smtpClient := smtp.NewSmtpClient(fmt.Sprintf("%s:%d", SmtpHost, SmtpApiPort))
	for attempt := 0; attempt < 10; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Second)
		}
		messages, _, err := smtpClient.GetMessages()
		require.NoError(t, err)
		fmt.Println("validateNoEmail: messages: ", len(messages.Items))
		require.Equal(t, 0, len(FindByMailTo(messages, mailTo)))
	}
}

func waitForTheEmail(t *testing.T, mailTo string, numEmails int) {
	smtpClient := smtp.NewSmtpClient(fmt.Sprintf("%s:%d", SmtpHost, SmtpApiPort))
	_, _, err := smtpClient.GetMessagesWithRetry(5, func(messages *mailhogclient.Messages, resp *http.Response, err error) bool {
		return err == nil &&
			messages.HasCount() &&
			len(FindByMailTo(messages, mailTo)) >= numEmails
	})
	require.NoError(t, err)
}

func FindByMailTo(messages *mailhogclient.Messages, mailTo string) []*mailhogclient.Message {
	msg := FindAllInArray(messages.Items,
		func(message mailhogclient.Message) bool {
			e := fmt.Sprintf("%s@%s", message.To[0].GetMailbox(), message.To[0].GetDomain())
			return e == mailTo
		})
	return msg
}

func FindAllInArray[V interface{}](arr []V, predicate func(V) bool) []*V {
	elements := []*V{}
	for _, element := range arr {
		if predicate(element) {
			elements = append(elements, &element)
		}
	}
	return elements
}
