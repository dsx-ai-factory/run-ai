//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/integration_tests/mailhog/mailhogclient"
	smtp_test "github.com/run-ai/runai/control-plane/notifications-service/integration_tests/tests/smtp"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/smtp"
	notifyclient "github.com/run-ai/runai/control-plane/notifications-service/pkg/client"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/stretchr/testify/require"
)

var configClient *config_client.APIClient

func TestMain(m *testing.M) {
	notificationHost = GetEnv("NOTIFICATION_HOST", NotificationHost)
	smtpHost = GetEnv("SMTP_HOST", SmtpHost)
	InitClients()
	os.Exit(m.Run())
}

func InitClients() {
	initNotifyClient(1)
	cfgServer := config_client.NewConfiguration()
	cfgServer.Host = ConfigUrl()
	cfgServer.Scheme = "http"

	configClient = config_client.NewAPIClient(cfgServer)
	initHistoryClient()
}

func initNotifyClient(batchSize int) (notifyclient.NotifyClienter, error) {
	return notifyclient.Init(&notifyclient.NotifyClientConfig{
		NotifyUrl:        "http://" + GetNotifyURL(),
		ConfigUrl:        "http://" + ConfigUrl(),
		BatchSize:        batchSize,
		BatchIntervalSec: 5,
	}, context.Background())
}

func GetNotifyURL() string {
	return fmt.Sprintf("%s:%s",
		GetEnv("NOTIFICATION_HOST", NotificationHost),
		GetEnv("NOTIFICATION_PORT", NotificationPort))
}

func ConfigUrl() string {
	return fmt.Sprintf("%s:%s",
		GetEnv("NOTIFICATION_HOST", NotificationHost),
		GetEnv("CONFIG_PORT", ConfigPort))
}

func TestSendEmailTriggeredByNotification(t *testing.T) {
	mailTo := getUniqueEmailAddr()
	sinkId := "SendEmailTriggeredByNotification"

	setupEmailSink(t, sinkId, mailTo, "user", "password")
	setupNotificationRoute(t, sinkId)
	sendAndVerifyEmail(t, mailTo)
	_, _, err := configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
}

func TestSendEmailTriggeredByNotification_WithoutCredentials(t *testing.T) {
	mailTo := getUniqueEmailAddr()
	sinkId := "SendEmailTriggeredByNotification_WithoutCredentials"

	setupEmailSink(t, sinkId, mailTo, "", "")
	setupNotificationRoute(t, sinkId)
	sendAndVerifyEmail(t, mailTo)
	_, _, err := configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
}

func TestSendEmailsBatchTriggeredByNotification(t *testing.T) {
	batchSize := 3
	initNotifyClient(batchSize)

	mailTo := getUniqueEmailAddr()
	sinkId := "SendEmailTriggeredByNotification_Batch"

	setupEmailSink(t, sinkId, mailTo, "user", "password")
	setupNotificationRoute(t, sinkId)
	sendAndVerifyBatchEmails(t, mailTo, batchSize)

	initNotifyClient(1)
	_, _, err := configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
}

func Test_Health(t *testing.T) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/%s", notificationHost, ConfigPort, "api/v1/config/health"), strings.NewReader(""))
	client := http.Client{}
	response, err := client.Do(req)
	require.NoError(t, err)

	defer response.Body.Close()
	_, err = io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func TestRegisterNotificationKinds_CreatesAnewOne(t *testing.T) {
	notification := &notifyclient.NotificationKind{
		Category:    "Workload timeout TestRegisterNotificationKinds_CreatesAnewOne",
		Type:        "Interactive workload stop due to timout",
		Description: "Interactive workload is about to be stopped due to timeout",
		Template:    "template",
	}
	err := notifyclient.RegisterNotification("integration-tests", []*notifyclient.NotificationKind{notification}, nil)
	require.NoError(t, err)

	cfgServer := config_client.NewConfiguration()
	cfgServer.Host = fmt.Sprintf("%s:%s", notificationHost, ConfigPort)
	cfgServer.Scheme = "http"
	client := config_client.NewAPIClient(cfgServer)
	notifications, _, err := client.RegisterNotificationAPI.GetNotificationKinds(context.Background()).Execute()
	require.NoError(t, err)
	require.Equal(t, "Workload timeout TestRegisterNotificationKinds_CreatesAnewOne", notifications[0].Category)
	require.Equal(t, "Interactive workload stop due to timout", notifications[0].Type)
	require.Equal(t, "Interactive workload is about to be stopped due to timeout", notifications[0].Description)
	require.Equal(t, "template", notifications[0].Template)
	remove := false
	_ = notifyclient.RegisterNotification("integration-tests", []*notifyclient.NotificationKind{}, &remove)
	_, _, err = configClient.RegisterNotificationAPI.DeleteNotificationKind(context.Background()).Category("Workload timeout TestRegisterNotificationKinds_CreatesAnewOne").Type_("Interactive workload stop due to timout").Execute()
	require.NoError(t, err)
}

func TestCreateNotification_AndGetByCategoryAndType(t *testing.T) {
	nKind := &notifyclient.NotificationKind{
		Category:    "Workload timeout TestCreateNotification_AndGetByCategoryAndType",
		Type:        "Interactive workload stop due to timeout",
		Description: "Interactive workload is about to be stopped due to timeout",
		Template:    "template",
	}
	err := notifyclient.RegisterNotification("integration-tests", []*notifyclient.NotificationKind{nKind}, nil)
	require.NoError(t, err)

	n, _, err := configClient.RegisterNotificationAPI.GetNotificationKinds(context.Background()).
		Category("Workload timeout TestCreateNotification_AndGetByCategoryAndType").Execute()
	require.NoError(t, err)
	require.Equal(t, 1, len(n))

	remove := true
	nKind2 := &notifyclient.NotificationKind{
		Category:    "Workload status change",
		Type:        "deleted",
		Description: "Workload resources were removed from the cluster",
		Template:    "template",
	}
	err = notifyclient.RegisterNotification("integration-tests", []*notifyclient.NotificationKind{nKind2}, &remove)
	n, _, err = configClient.RegisterNotificationAPI.GetNotificationKinds(context.Background()).Category("Workload status change").
		Type_("deleted").Execute()
	require.NoError(t, err)

	notification := n[0]
	require.Equal(t, 1, len(n))
	require.NoError(t, err)
	require.Equal(t, "Workload status change", notification.Category)
	require.Equal(t, "deleted", notification.Type)
	require.Equal(t, "Workload resources were removed from the cluster", notification.Description)
	require.Equal(t, "template", notification.Template)

	err = notifyclient.RegisterNotification("integration-tests", []*notifyclient.NotificationKind{}, &remove)
	require.NoError(t, err)
}

func TestAddAndGetTenantRoute(t *testing.T) {
	routeConfig := &config_client.RouteData{
		RouteId:  "TestAddAndGetTenantRoute_email_k8s_notifications",
		TenantId: TenantId,
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: ListenerId,
			NotifyIds:  []string{"email1"},
		}),
	}
	_, _, err := configClient.TenantRouteAPI.CreateRoute(context.Background(), routeConfig.TenantId).
		RouteData(*routeConfig).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeConfig.RouteId).Execute()

	retrievedRoute, _, err := configClient.TenantRouteAPI.GetRoute(context.Background(), TenantId, routeConfig.RouteId).Execute()
	require.NoError(t, err)

	routeConfigBytes, err := json.Marshal(routeConfig.RouteConfig)
	require.NoError(t, err)

	retrievedRouteBytes, err := json.Marshal(retrievedRoute.RouteConfig)
	require.NoError(t, err)
	require.JSONEq(t, string(routeConfigBytes), string(retrievedRouteBytes),
		"sent and retrieved should be same")
}

func FindByMailTo(messages *mailhogclient.Messages, mailTo string) []*mailhogclient.Message {
	msg := FindAllInArray(messages.Items,
		func(message mailhogclient.Message) bool {
			e := fmt.Sprintf("%s@%s", message.To[0].GetMailbox(), message.To[0].GetDomain())
			log.Info().Msgf("Found an email inbox with target email %s", e)
			return e == mailTo
		})
	return msg
}

func waitForTheEmail(t *testing.T, mailTo string, numEmails int) {
	smtpClient := smtp_test.NewSmtpClient(fmt.Sprintf("%s:%d", smtpHost, SmtpApiPort))
	_, _, err := smtpClient.GetMessagesWithRetry(5, func(messages *mailhogclient.Messages, resp *http.Response, err error) bool {
		return err == nil &&
			messages.HasCount() &&
			len(FindByMailTo(messages, mailTo)) == numEmails
	})
	require.NoErrorf(t, err, "Failed to find a message with destination mail %s", mailTo)
}

func getUniqueEmailAddr() string {
	return uuid.New().String() + "@run.ai"
}

func setupEmailSink(t *testing.T, sinkId, mailTo, user, password string) {
	emailConfig := StructToMap(
		smtp.SmtpConfig{
			SmtpHost:       SmtpHost,
			SmtpPort:       SmtpPort,
			SmtpTlsEnabled: false,
			User:           user,
			Password:       password,
		},
	)
	emailConfig["recipients"] = []string{mailTo}
	sinkConfig := config_client.SinkData{
		Id:     sinkId,
		Type:   "email",
		Config: emailConfig,
	}
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(sinkConfig).Execute()
	require.NoError(t, err)
}

func setupNotificationRoute(t *testing.T, routeId string) {
	routeData := config_client.RouteData{
		TenantId: TenantId,
		RouteId:  routeId,
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: ListenerId,
			NotifyIds:  []string{routeId},
		}),
	}
	_, _, err := configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).RouteData(routeData).Execute()
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _, err := configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()
		require.NoError(t, err)
	})
}

func sendAndVerifyEmail(t *testing.T, mailTo string) {
	notifyclient.BuildNotification("name", "type").
		WithTenant(TenantId).WithUser(mailTo).Notify()
	waitForTheEmail(t, mailTo, 1)
}

func sendAndVerifyBatchEmails(t *testing.T, mailTo string, batchSize int) {
	for i := 0; i < batchSize; i++ {
		notifyclient.BuildNotification("name", "type").
			WithTenant(TenantId).WithUser(mailTo).Notify()
	}
	waitForTheEmail(t, mailTo, batchSize)
}
