/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"fmt"
	"time"

	authmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/auth"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"

	"github.com/golang/mock/gomock"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"golang.org/x/oauth2"
)

const (
	appIdName     = "app-lies"
	expiryMinutes = 10
	// tokenWithClaims is a synthetic JWT: fake issuer/subject, no real signature ("mock").
	tokenWithClaims = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwiaXNzIjoiaHR0cHM6Ly9mYWtlLWlzc3Vlci5leGFtcGxlLmNvbS9hdXRoL3JlYWxtcy9ydW5haSIsImF1ZCI6WyJydW5haSIsImJyb2tlciIsImFjY291bnQiXSwic3ViIjoidGVzdEBydW4uYWkiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJjbGkiLCJhbGxvd2VkX29yaWdpbnMiOlsiaHR0cHM6Ly9mYWtlLWlzc3Vlci5leGFtcGxlLmNvbSJdLCJyZWFsbV9hY2Nlc3MiOnsicm9sZXMiOlsib2ZmbGluZV9hY2Nlc3MiLCJ1bWFfYXV0aG9yaXphdGlvbiIsImRlZmF1bHQtcm9sZXMtcnVuYWkiXX0sInJlc291cmNlX2FjY2VzcyI6eyJicm9rZXIiOnsicm9sZXMiOlsicmVhZC10b2tlbiJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwic2NvcGUiOiJvcGVuaWQgb2ZmbGluZV9hY2Nlc3MgcHJvZmlsZSBlbWFpbCIsInRlbmFudF9pZCI6MTAwMSwiZW1haWxfdmVyaWZpZWQiOnRydWUsInN1YmplY3RfdXVpZCI6IjU4YmJmNDZhLTEyNDUtNDg2Mi1hMmZkLTI3YTVlNTIwMjhiNyIsInByZWZlcnJlZF91c2VybmFtZSI6InRlc3RAcnVuLmFpIiwiZW1haWwiOiJ0ZXN0QHJ1bi5haSJ9.mock" //nolint:gosec,lll // no security hazard
)

func MockAuthentication(mockCtrl *gomock.Controller) {
	mockAuth := authmocks.NewMockAuthenticator(mockCtrl)
	mockAuth.EXPECT().Authenticate(gomock.Any()).Return(&oauth2.Token{
		AccessToken:  "lies",
		TokenType:    "",
		RefreshToken: "",
		Expiry:       time.Now().Add(expiryMinutes * time.Minute),
	}, nil).AnyTimes()
	auth.SetAuthenticator(mockAuth)
}

func MockAuthenticationClaims(mockCtrl *gomock.Controller) {
	mockAuth := authmocks.NewMockAuthenticator(mockCtrl)
	mockAuth.EXPECT().Authenticate(gomock.Any()).Return(&oauth2.Token{
		AccessToken:  tokenWithClaims,
		TokenType:    "",
		RefreshToken: "",
		Expiry:       time.Now().Add(expiryMinutes * time.Minute),
	}, nil).AnyTimes()
	auth.SetAuthenticator(mockAuth)
}

func AuthValidation(configPath, configFile string, envMap map[string]string) error {
	_, err := cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
	if err != nil {
		return err
	}

	_, err = cli.RunCommand(fmt.Sprintf("login app --name=%s --secret=shhhhh --interactive=disabled --save", appIdName), envMap)
	return err
}
