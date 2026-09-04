//go:build integration
// +build integration

/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"testing"

	"github.com/stretchr/testify/require"
)

var DefaultSmtpListenAddr = "notifications-service:2525"

func TestSendEmail_PlainTextBlock(t *testing.T) {
	mailTo := getUniqueEmailAddr()
	sender := "test@example.com"

	sinkConfig := createDirectEmailSink()
	sinkConfig.SetId(emailSinkID)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	routeData := buildMatcherRoute(TenantId, emailRouteID, []string{emailSinkID})
	_, _, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()

	listenerAddr := GetEnv("SMTP_LISTENER_ADDR", DefaultSmtpListenAddr)

	jsonBody := fmt.Sprintf(`-----BEGIN NOTIFICATION-----{"tenant_id": "%s", "user_id": "%s"}-----END NOTIFICATION-----`, TenantId, mailTo)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Test JSON\r\nContent-Type: text/html\r\n\r\n%s", sender, mailTo, jsonBody)

	err = smtp.SendMail(listenerAddr, nil, sender, []string{mailTo}, []byte(msg))
	require.NoError(t, err)

	waitForTheEmail(t, mailTo, 1)
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), emailSinkID).Execute()
	require.NoError(t, err)
}

func TestSendEmail_MultipartNotificationBlock(t *testing.T) {
	mailTo := getUniqueEmailAddr()
	sender := "test@example.com"

	sinkConfig := createDirectEmailSink()
	sinkConfig.SetId(emailSinkID)
	_, _, err := configClient.SinksAPI.CreateSink(context.Background()).SinkData(*sinkConfig).Execute()
	require.NoError(t, err)

	routeData := buildMatcherRoute(TenantId, emailRouteID, []string{emailSinkID})
	_, _, err = configClient.TenantRouteAPI.CreateRoute(context.Background(), TenantId).
		RouteData(routeData).Execute()
	require.NoError(t, err)
	defer configClient.TenantRouteAPI.DeleteRoute(context.Background(), TenantId, routeData.RouteId).Execute()

	listenerAddr := GetEnv("SMTP_LISTENER_ADDR", DefaultSmtpListenAddr)

	jsonBlock := fmt.Sprintf(`-----BEGIN NOTIFICATION-----{"tenant_id": "%s", "user_id": "%s"}-----END NOTIFICATION-----`, TenantId, mailTo)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreatePart(nil)
	require.NoError(t, err)
	_, err = part.Write([]byte(jsonBlock))
	require.NoError(t, err)
	w.Close()

	email := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: Multipart Test\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n%s", sender, mailTo, w.Boundary(), buf.String())

	err = smtp.SendMail(listenerAddr, nil, sender, []string{mailTo}, []byte(email))
	require.NoError(t, err)

	waitForTheEmail(t, mailTo, 1)
	_, _, err = configClient.SinksAPI.DeleteSink(context.Background(), emailSinkID).Execute()
	require.NoError(t, err)
}
