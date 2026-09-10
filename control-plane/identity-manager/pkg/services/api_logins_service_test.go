/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

const appURL = "https://mock-q7z3k.example.com"

func initLoginsServiceTest() (*LoginsApiService, context.Context) {
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, "runai")
	ctx = context.WithValue(ctx, utils.RequestHostKey, "example.com")
	return &LoginsApiService{
		appURL: appURL,
	}, ctx
}

func TestCLIRedirection_BrowserFlow(t *testing.T) {
	las, ctx := initLoginsServiceTest()

	userAgent := "runai-cli/"

	res, err := las.GetLoginScreenURL(ctx, userAgent, "", "http://localhoster:8080", "", "")
	assert.Nil(t, err)

	assert.Equal(t, 307, res.Code)
	assert.Nil(t, res.Body)
	assert.NotNil(t, res.Headers)

	assert.NotEmpty(t, res.Headers[utils.AbsoluteLocationHeaderKey])
	assert.Contains(t, res.Headers[utils.AbsoluteLocationHeaderKey][0], appURL)

	query := strings.Split(res.Headers[utils.AbsoluteLocationHeaderKey][0], "?")[1]

	params, err := url.ParseQuery(query)
	assert.Nil(t, err)

	assert.Equal(t, []string{utils.CLI_CLIENT_ID}, params[utils.ClientIDParam])
	assert.Contains(t, params[utils.RedirectURIParam][0], "localhoster")
}

func TestCLIRedirection_ExternalBrowserFlow(t *testing.T) {
	las, ctx := initLoginsServiceTest()

	userAgent := "runai-cli/"

	res, err := las.GetLoginScreenURL(ctx, userAgent, utils.AuthenticationFlowExternalBrowser, "", "", "")
	assert.Nil(t, err)

	assert.Equal(t, 307, res.Code)
	assert.Nil(t, nil, res.Body)
	assert.NotNil(t, res.Headers)

	assert.NotEmpty(t, res.Headers[utils.AbsoluteLocationHeaderKey])
	assert.Contains(t, res.Headers[utils.AbsoluteLocationHeaderKey][0], appURL)

	query := strings.Split(res.Headers[utils.AbsoluteLocationHeaderKey][0], "?")[1]

	params, err := url.ParseQuery(query)
	assert.Nil(t, err)

	assert.Equal(t, []string{utils.CLI_CLIENT_ID}, params[utils.ClientIDParam])
	assert.Contains(t, params[utils.RedirectURIParam][0], "oauth-code")
	assert.Contains(t, params[utils.RedirectURIParam][0], "https")
}

func TestUIRedirection(t *testing.T) {
	las, ctx := initLoginsServiceTest()

	res, err := las.GetLoginScreenURL(ctx, "", "", "", "", "")
	assert.Nil(t, err)

	assert.Equal(t, 307, res.Code)
	assert.Nil(t, nil, res.Body)
	assert.NotNil(t, res.Headers)

	assert.NotEmpty(t, res.Headers[utils.AbsoluteLocationHeaderKey])
	assert.Contains(t, res.Headers[utils.AbsoluteLocationHeaderKey][0], appURL)

	query := strings.Split(res.Headers[utils.AbsoluteLocationHeaderKey][0], "?")[1]

	params, err := url.ParseQuery(query)
	assert.Nil(t, err)

	assert.Equal(t, []string{utils.RUNAI_ADMIN_UI_CLIENT_ID}, params[utils.ClientIDParam])
	assert.Contains(t, params[utils.RedirectURIParam][0], "callback")
	assert.Contains(t, params[utils.RedirectURIParam][0], "https")
}
