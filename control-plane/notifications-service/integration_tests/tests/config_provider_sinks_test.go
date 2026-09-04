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
	"net/http"
	"testing"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks/smtp"
	"github.com/stretchr/testify/require"
)

const SlackUrl = "http://slack-mock:3000/"

func TestPostgresSink_Create(t *testing.T) {
	sinkId := fmt.Sprintf("TestPostgresSink_CreateGetDelete_%s", uuid.New().String())
	sinkConfig := createPostgresSink()
	sinkConfig.SetId(sinkId)
	res, httpRes, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)
	require.NotNil(t, httpRes)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)
	require.Contains(t, res, "sink created")
	defer configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()

	sink, _, err := configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)

	retrievedSink, err := json.Marshal(sink.GetConfig())
	require.NoError(t, err)
	require.Equal(t, sinks.PostgresSinkType, sink.GetType())

	delete(sinkConfig.Config, "password")
	expectedSinkConfigNoPass, err := json.Marshal(sinkConfig.Config)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedSinkConfigNoPass), string(retrievedSink),
		"sent and retrieved should match except removed password")
}

func TestPostgresSink_CreateDuplicate(t *testing.T) {
	sinkId := fmt.Sprintf("TestPostgresSink_CreateDuplicate_%s", uuid.New().String())
	sinkConfig := createPostgresSink()
	sinkConfig.SetId(sinkId)

	_, httpRes, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)
	defer configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()

	duplicateConfig := createPostgresSink()
	duplicateConfig.SetId(sinkId)

	_, httpRes, err = configClient.SinksAPI.CreateSink(context.Background()).SinkData(*duplicateConfig).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, httpRes.StatusCode)
	require.EqualError(t, err, "409 Conflict")
}

func TestPostgresSink_UpdateAndGetSinkNoPass(t *testing.T) {
	sinkId := fmt.Sprintf("TestPostgresSink_UpdateGet_%s", uuid.New().String())
	sinkConfig := createPostgresSink()
	sinkConfig.SetId(sinkId)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)
	defer configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()

	updated := sinks.PostgresSinkConfig{
		Host:     host,
		Port:     "5432",
		Database: "postgres",
		Schema:   "public",
		Table:    "notification_events_v2",
		User:     user,
		SslMode:  "disable",
	}
	updateMap := StructToMap(updated)
	update := config_client.NewSinkDataForUpdate(sinks.PostgresSinkType)
	update.SetConfig(updateMap)
	_, httpRes, err := configClient.SinksAPI.UpdateSink(context.Background(), sinkId).SinkDataForUpdate(*update).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	sink, _, err := configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
	require.Equal(t, sinks.PostgresSinkType, sink.GetType())

	retrievedSink, err := json.Marshal(sink.GetConfig())
	require.NoError(t, err)

	expectedMap := StructToMap(updated)
	delete(expectedMap, "password")
	expectedSinkConfigNoPass, err := json.Marshal(expectedMap)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedSinkConfigNoPass), string(retrievedSink))
}

func TestPostgresSink_Delete(t *testing.T) {
	sinkId := fmt.Sprintf("TestPostgresSink_Delete_%s", uuid.New().String())
	sinkConfig := createPostgresSink()
	sinkConfig.SetId(sinkId)
	_, httpRes, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)
	require.NotNil(t, httpRes)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	sink, httpRes, err := configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
	require.NotNil(t, sink)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	_, httpRes, err = configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
	require.NotNil(t, httpRes)
	require.Equal(t, http.StatusOK, httpRes.StatusCode)

	sink, httpRes, err = configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.Error(t, err)
	require.Nil(t, sink)
	require.Equal(t, http.StatusNotFound, httpRes.StatusCode)
}

func TestUpdateAndGetSinkNoPass(t *testing.T) {
	sinkId := "TestUpdateAndGetSinkNoPass"
	sinkConfig := createEmailSinkWithRecipient(getUniqueEmailAddr())
	sinkConfig.SetId(sinkId)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	sink, _, err := configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)

	retrievedSink, err := json.Marshal(sink.GetConfig())
	delete(sinkConfig.Config, "password")

	require.Equal(t, "email", sink.Type)
	expectedSinkConfigNoPass, err := json.Marshal(sinkConfig.Config)
	require.JSONEq(t, string(expectedSinkConfigNoPass), string(retrievedSink),
		"sent and retrieved should be same, except removed password")
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
}

func TestUpdateAndGetSlackSinkNoPass(t *testing.T) {
	sinkId := "slack_notifications"
	sinkConfig := createSlackSinkWithRecipient("dummy_token", []string{getUniqueEmailAddr()})
	sinkConfig.SetId(sinkId)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	sink, _, err := configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)

	retrievedSink, err := json.Marshal(sink.GetConfig())
	require.Equal(t, "slack", sink.Type)

	delete(sinkConfig.Config, "api_token")
	expectedSinkConfigWithoutApiToken, err := json.Marshal(sinkConfig.Config)
	require.JSONEq(t, string(expectedSinkConfigWithoutApiToken), string(retrievedSink),
		"sent and retrieved should be same, except removed api token")
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
}

func TestDeleteSink(t *testing.T) {
	sinkId := fmt.Sprintf("TestDeleteSink_%s", uuid.New().String())
	sinkConfig := createEmailSinkWithRecipient(getUniqueEmailAddr())
	sinkConfig.SetId(sinkId)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	sink, httpResponse, err := configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)
	require.NotNil(t, sink)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), sinkId).Execute()
	require.NoError(t, err)

	sink, httpResponse, err = configClient.SinksAPI.GetSink(context.Background(), sinkId).Execute()
	require.Error(t, err)
	require.Nil(t, sink)
	require.Equal(t, http.StatusNotFound, httpResponse.StatusCode)
}

func TestDeleteSink_Error_NotFound(t *testing.T) {
	res, httpResponse, err := configClient.SinksAPI.DeleteSink(context.Background(), "non_existing_sink_id").Execute()
	require.Equal(t, "", res)
	require.EqualError(t, err, "404 Not Found")
	require.Equal(t, http.StatusNotFound, httpResponse.StatusCode)
}

func TestEmailSettingsVerification_Valid(t *testing.T) {
	responseBody, err := sendEmailSettingsVerificationRequest(smtp.AuthLogin, smtpHost, SmtpPort, false, http.StatusOK)
	require.NoError(t, err)
	require.NotNil(t, responseBody)
	require.Equal(t, http_listener.ErrSuccess, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should have succeeded")
}

func TestSlackSettingsVerification_Valid(t *testing.T) {
	responseBody, err := sendSlackSettingsVerificationRequest("dummy_token_pass_validation")
	require.NoError(t, err)
	require.NotNil(t, responseBody)
	require.Equal(t, http_listener.ErrSuccess, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should have succeeded")
}

func TestSlackSettingsVerification_NotValid(t *testing.T) {
	responseBody, err := sendSlackSettingsVerificationRequest("not_valid_token")
	require.NoError(t, err)
	require.NotNil(t, responseBody)
	require.Equal(t, http_listener.ErrValidationFailed, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should fail")
}

func TestEmailSettingsVerification_InvalidPort(t *testing.T) {
	responseBody, err := sendEmailSettingsVerificationRequest(smtp.AuthLogin, smtpHost, 666, false, http.StatusOK)
	require.NoError(t, err)
	require.NotNil(t, responseBody, "expected validation response body")
	require.Equal(t, http_listener.ErrValidationFailed, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should have failed due to wrong port")
}

func TestEmailSettingsVerification_InvalidHost(t *testing.T) {
	responseBody, err := sendEmailSettingsVerificationRequest(smtp.AuthLogin, "wrong_host", SmtpPort, false, http.StatusOK)
	require.NoError(t, err)
	require.NotNil(t, responseBody, "expected validation response body")
	require.Equal(t, http_listener.ErrValidationFailed, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should have failed due to wrong host")
}

func TestEmailSettingsVerification_TlsNotSupported(t *testing.T) {
	responseBody, err := sendEmailSettingsVerificationRequest(smtp.AuthLogin, smtpHost, SmtpPort, true, http.StatusOK)
	require.NoError(t, err)
	require.NotNil(t, responseBody, "expected validation response body")
	require.Equal(t, http_listener.ErrValidationFailed, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should have failed due to TLS not supported")
}

func TestEmailSettingsVerification_TlsForPlainAuth(t *testing.T) {
	for _, tlsEnabled := range []bool{false, true} {
		responseBody, err := sendEmailSettingsVerificationRequest(smtp.AuthPlain, smtpHost, SmtpPort, tlsEnabled, http.StatusOK)
		require.NoError(t, err)
		require.NotNil(t, responseBody)
		if tlsEnabled {
			// MailHog does not support SMTP authentication or TLS encryption. It’s intentionally insecure for local use only
			require.Equal(t, http_listener.ErrValidationFailed, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should have failed because TLS is not supported")
		} else {
			require.Equal(t, http_listener.ErrSuccess, http_listener.ErrorCode(responseBody.GetErrorCode()), "Should have succeeded with TLS disabled")
		}
	}
}

func TestEmailSettingsVerification_NoneAuthServerRequiresAuth(t *testing.T) {
	responseBody, err := sendEmailSettingsVerificationRequest(smtp.AuthNone, "sandbox.smtp.mailtrap.io", 587, false, http.StatusOK)
	require.NoError(t, err)
	require.NotNil(t, responseBody)
	log.Info().Msgf("TestEmailSettingsVerification_NoneAuthServerRequiresAuth: responseBody: %+v", responseBody)
	require.Equal(t,
		http_listener.ErrValidationFailed,
		http_listener.ErrorCode(responseBody.GetErrorCode()),
		"verification must fail when smtp server rejects unauthenticated delivery and auth_type is auth_none",
	)
}

// TestEmailSettingsVerification_NoneAuthServerAllowsUnauthenticated guards against a
// regression: a server that advertises the AUTH extension but permits unauthenticated
// delivery (e.g. an IP-allowlisted relay, as emulated by the local mailhog) must pass
// verification with auth_type auth_none.
func TestEmailSettingsVerification_NoneAuthServerAllowsUnauthenticated(t *testing.T) {
	responseBody, err := sendEmailSettingsVerificationRequest(smtp.AuthNone, smtpHost, SmtpPort, false, http.StatusOK)
	require.NoError(t, err)
	require.NotNil(t, responseBody)
	log.Info().Msgf("TestEmailSettingsVerification_NoneAuthServerAllowsUnauthenticated: responseBody: %+v", responseBody)
	require.Equal(t,
		http_listener.ErrSuccess,
		http_listener.ErrorCode(responseBody.GetErrorCode()),
		"verification must succeed when smtp server permits unauthenticated delivery and auth_type is auth_none",
	)
}

func sendEmailSettingsVerificationRequest(authType smtp.AuthType, smtpHost string, smtpPort int, smtpTlsEnabled bool, expectedStatusCode int) (*config_client.ValidateSink200Response, error) {
	cfg := smtp.SmtpConfig{
		SmtpTlsEnabled: smtpTlsEnabled,
		SmtpHost:       smtpHost,
		SmtpPort:       smtpPort,
		AuthType:       authType,
	}
	if authType != smtp.AuthNone {
		cfg.User = "user"
		cfg.Password = "password"
	}
	config := StructToMap(cfg)
	config["from"] = "runai-notifications@example.com"
	requestData := config_client.SinkData{
		Id:   "validation_sink",
		Type: "email", Config: config,
	}

	res, r, err := configClient.SinksAPI.ValidateSink(context.Background()).SinkData(requestData).Execute()
	if err != nil {
		return nil, err
	}

	if r.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("got an unexpected error code. Recieved - %d, expected - %d", r.StatusCode, expectedStatusCode)
	}

	return res, nil
}

func sendSlackSettingsVerificationRequest(token string) (*config_client.ValidateSink200Response, error) {
	requestData := createSlackSinkWithRecipient(token, []string{})
	requestData.SetId("validation_sink")
	res, r, err := configClient.SinksAPI.ValidateSink(context.Background()).SinkData(*requestData).Execute()
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("r: %+v, res: %+v", r, res)
	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got an unexpected error code. Recieved - %d, expected - %d", r.StatusCode, http.StatusOK)
	}

	return res, nil
}

func createEmailSinkWithRecipient(mailTo string) *config_client.SinkData {
	emailConfig := StructToMap(sinks.EmailSinkConfig{
		Smtp: smtp.SmtpConfig{
			User:           "user",
			Password:       "password",
			SmtpTlsEnabled: false,
			SmtpHost:       smtpHost,
			SmtpPort:       SmtpPort,
			AuthType:       smtp.AuthLogin,
		},
		From:       "notification-service-integration-tests@run.ai",
		Recipients: []string{mailTo},
	})

	return &config_client.SinkData{
		Type: "email", Config: emailConfig,
	}
}

func createSlackSinkWithRecipient(apiToken string, recipients []string) *config_client.SinkData {
	return &config_client.SinkData{
		Type: "slack", Config: StructToMap(sinks.SlackSinkConfig{
			Url:                 SlackUrl,
			ApiToken:            apiToken,
			DirectNotifications: true,
			Recipients:          recipients,
		}),
	}
}

func createPostgresSink() *config_client.SinkData {
	return &config_client.SinkData{
		Type: sinks.PostgresSinkType,
		Config: StructToMap(sinks.PostgresSinkConfig{
			Host:     host,
			Port:     "5432",
			Database: "postgres",
			Schema:   "notifications_service",
			Table:    "notification_events",
			User:     user,
			Password: pwd,
			SslMode:  "disable",
		}),
	}
}

func createDirectEmailSink() *config_client.SinkData {
	emailConfig := StructToMap(sinks.EmailSinkConfig{
		Smtp: smtp.SmtpConfig{
			User:           "user",
			Password:       "password",
			SmtpTlsEnabled: false,
			SmtpHost:       smtpHost,
			SmtpPort:       SmtpPort,
			AuthType:       smtp.AuthLogin,
		},
		From:                "notification-service-integration-tests@run.ai",
		DirectNotifications: true,
	})

	return &config_client.SinkData{
		Type: "email", Config: emailConfig,
	}
}
