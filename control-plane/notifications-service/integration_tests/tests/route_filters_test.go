//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/integration_tests/mailhog/mailhogclient"
	smtp_test "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/integration_tests/tests/smtp"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/filter/filters"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	notifyclient "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/config_client"
	_ "github.com/lib/pq"
	"github.com/sethvargo/go-retry"
	slack_go "github.com/slack-go/slack"
	"github.com/stretchr/testify/require"
)

const (
	emailSinkID    = "email_prom"
	postgresSinkID = "postgres"
	slackSinkID    = "slack_notifier"

	emailRouteID    = "email_matcher_notification"
	postgresRouteID = "postgres_matcher_notification"
	slackRouteID    = "slack_matcher_notification"
)

var (
	host = GetEnv("POSTGRES_HOST", "postgres")
	pwd  = GetEnv("POSTGRES_PASSWORD", "admin")
	user = GetEnv("POSTGRES_USER", "runai")
)

func buildMatcherRoute(tenantId, routeId string, notifyIds []string) config_client.RouteData {
	return config_client.RouteData{
		RouteId:  routeId,
		TenantId: tenantId,
		RouteConfig: StructToMap(route.BasicRouteConfiguration{
			ListenerId: ListenerId,
			Filtering: filter.FilteringConfig{Filters: []*filter.FilterConfig{{
				Type: "matcher",
				Matcher: &filters.MatcherFilterConfig{&filters.PropertyMatcher{
					MatchType: "==",
					Key:       "tenant_id",
					Value:     tenantId,
				}},
			}}},
			NotifyIds: notifyIds,
		}),
	}
}

func TestSendPostgresWithFilter(t *testing.T) {
	mailTo := getUniqueEmailAddr()
	sinkConfig := &config_client.SinkData{
		Type: sinks.PostgresSinkType,
		Config: StructToMap(sinks.PostgresSinkConfig{
			Host:     host,
			Port:     "5432",
			Database: "runai",
			Schema:   "public",
			Table:    "UserEvents",
			User:     user,
			Password: pwd,
			SslMode:  "disable",
		}),
	}
	sinkConfig.SetId(postgresSinkID)
	_, httpRes, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)
	defer func() {
		_, httpRes, err := configClient.SinksAPI.DeleteSink(context.Background(), postgresSinkID).Execute()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, httpRes.StatusCode)
	}()

	routeData := buildMatcherRoute(TenantId, postgresRouteID, []string{postgresSinkID})
	_, httpRes, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)
	defer func() {
		_, httpRes, err := configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, httpRes.StatusCode)
	}()

	const description = "postgres sink integration filter"
	notifyclient.BuildNotification("name", "type").
		WithTenant(TenantId).WithUser(mailTo).WithDescription(description).Notify()

	waitForPostgresEventsRow(t, TenantId, mailTo, description, 1)
}

func TestSendEmailWithFilter(t *testing.T) {
	mailTo := getUniqueEmailAddr()

	sinkConfig := createEmailSinkWithRecipient(mailTo)
	sinkConfig.SetId(emailSinkID)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	routeData := buildMatcherRoute(TenantId, emailRouteID, []string{emailSinkID})
	_, _, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()

	notifyclient.BuildNotification("name", "type").
		WithTenant(TenantId).WithUser(mailTo).Notify()

	waitForTheEmail(t, mailTo, 1)
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), emailSinkID).Execute()
	require.NoError(t, err)
}

func TestSendSlackMessageWithFilter(t *testing.T) {
	tenantId := "12345"
	category, typ, channel := "some category", "some type", "CSOMECHANNEL"

	sinkConfig := createSlackSinkWithRecipient("dummy_token_pass_validation", []string{channel})
	sinkConfig.SetId(slackSinkID)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	routeData := buildMatcherRoute(tenantId, slackRouteID, []string{slackSinkID})
	_, _, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), tenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), tenantId, routeData.RouteId).Execute()

	notifyclient.BuildNotification(category, typ).WithTenant(tenantId).Notify()

	validateMessageSent(t, SlackUrl, channel, tenantId, category, typ)
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), slackSinkID).Execute()
	require.NoError(t, err)
}

func TestSendEmailWithFilter_ShouldNotSendEmail(t *testing.T) {
	mailTo := getUniqueEmailAddr()

	sinkConfig := createEmailSinkWithRecipient(mailTo)
	sinkConfig.SetId(emailSinkID)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	routeData := buildMatcherRoute(TenantId, emailRouteID, []string{emailSinkID})
	_, _, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()

	notifyclient.BuildNotification("name", "type").WithTenant("some other tenant").WithUser(mailTo).Notify()

	smtpClient := smtp_test.NewSmtpClient(fmt.Sprintf("%s:%d", smtpHost, SmtpApiPort))
	_, _, err = smtpClient.GetMessagesWithRetry(5, func(messages *mailhogclient.Messages, resp *http.Response, err error) bool {
		return err == nil &&
			messages.HasCount() &&
			len(FindByMailTo(messages, mailTo)) > 0
	})
	require.Error(t, err)
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), emailSinkID).Execute()
	require.NoError(t, err)
}

func TestSendEmailWithFilterAndTimeout(t *testing.T) {
	initNotifyClient(10)
	defer initNotifyClient(1)
	mailTo := getUniqueEmailAddr()

	sinkConfig := createEmailSinkWithRecipient(mailTo)
	sinkConfig.SetId(emailSinkID)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	routeData := buildMatcherRoute(TenantId, emailRouteID, []string{emailSinkID})
	_, _, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()

	notifyclient.BuildNotification("name", "type").
		WithTenant(TenantId).WithUser(mailTo).Notify()

	time.Sleep(time.Second * 10)
	waitForTheEmail(t, mailTo, 1)
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), emailSinkID).Execute()
	require.NoError(t, err)
}

func TestSendMailNoNotifiers_ShouldNotSendEmail(t *testing.T) {
	const routeID = "no_notifiers_route"
	const testSinkID = "no_notifiers_sink"

	mailTo := getUniqueEmailAddr()

	sinkConfig := createEmailSinkWithRecipient(mailTo)
	sinkConfig.SetId(testSinkID)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)
	defer configClient.SinksAPI.DeleteSink(context.Background(), testSinkID).Execute()

	routeData := buildMatcherRoute(TenantId, routeID, []string{})
	_, _, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()

	_, _, err = configClient.UserRouteAPI.CreateUserRoute(context.Background(), TenantId, mailTo).
		UserRouteData(*config_client.NewUserRouteData(TenantId, mailTo, routeData.RouteId, routeData.RouteConfig)).Execute()
	require.NoError(t, err)
	defer configClient.UserRouteAPI.DeleteUserRoute(context.Background(), TenantId, mailTo, routeData.RouteId).Execute()

	notifyclient.BuildNotification("name", "type").WithTenant(TenantId).WithUser(mailTo).Notify()

	smtpClient := smtp_test.NewSmtpClient(fmt.Sprintf("%s:%d", smtpHost, SmtpApiPort))
	_, _, err = smtpClient.GetMessagesWithRetry(5, func(messages *mailhogclient.Messages, resp *http.Response, err error) bool {
		return err == nil &&
			messages.HasCount() &&
			len(FindByMailTo(messages, mailTo)) > 0
	})
	require.Error(t, err, "no notify_ids on tenant and user routes: email must not be delivered")
}

func validateMessageSent(t *testing.T, url, channel, tenantId, category, typ string) {
	jsonResponse := map[string]interface{}{}
	retryableGetMessagesFunc := func(ctx context.Context) error {
		jsonResponse, err := getSlackMessages(t, url, jsonResponse)
		if err == nil && jsonResponse["token"] == "dummy_token_pass_validation" {
			return nil
		}
		return retry.RetryableError(fmt.Errorf("failed retry condition. Retrying... (Error:  %w)", err))
	}

	require.NoError(t, retry.Do(
		context.Background(),
		retry.WithMaxRetries(60, retry.NewConstant(time.Second)),
		retryableGetMessagesFunc))

	require.Equal(t, "dummy_token_pass_validation", jsonResponse["token"])
	require.Equal(t, channel, jsonResponse["channel"])

	validateSentMessageData(t, jsonResponse, tenantId, category, typ)
}

func validateSentMessageData(t *testing.T, jsonResponse map[string]interface{}, tenantId, category, typ string) {
	var blocks []slack_go.SectionBlock
	err := json.Unmarshal([]byte(jsonResponse["blocks"].(string)), &blocks)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(blocks), 1)
	cnt := blocks[len(blocks)-1].Text.Text
	require.True(t, strings.Contains(cnt, category))
	require.True(t, strings.Contains(cnt, typ))
	require.True(t, strings.Contains(cnt, tenantId))
}

func getSlackMessages(t *testing.T, url string, jsonResponse map[string]interface{}) (map[string]interface{}, error) {
	resp, err := http.Get(url + "/chat.getMessage")
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &jsonResponse)
	return jsonResponse, err
}

func waitForPostgresEventsRow(t *testing.T, tenantID, userID, wantDescriptionSubstring string, minCount int) {
	dsn := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=runai sslmode=disable", host, user, pwd)
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	err = retry.Do(context.Background(), retry.WithMaxRetries(30, retry.NewConstant(time.Second)), func(ctx context.Context) error {
		var n int
		q := `SELECT COUNT(*) FROM public."UserEvents" WHERE tenant_id = $1 AND user_id = $2 AND description LIKE $3`
		err := db.QueryRowContext(ctx, q, tenantID, userID, "%"+wantDescriptionSubstring+"%").Scan(&n)
		if err != nil {
			return retry.RetryableError(err)
		}
		if n < minCount {
			return retry.RetryableError(fmt.Errorf("expected at least %d matching rows in UserEvents, got %d", minCount, n))
		}
		return nil
	})
	require.NoError(t, err)
}
