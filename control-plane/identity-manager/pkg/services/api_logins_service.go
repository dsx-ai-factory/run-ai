/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

const (
	RedirectURIFormat = "auth/realms/%s/protocol/openid-connect/auth?%s"
)

type LoginsApiService struct {
	appURL string
}

func NewLoginsApiService(appURL string) api.LoginsAPIServicer {
	return &LoginsApiService{
		appURL: appURL,
	}
}

func (s *LoginsApiService) GetLoginScreenURL(
	ctx context.Context,
	userAgent string,
	authenticationFlow string,
	redirectURL string,
	codeChallenge string,
	codeChallengeMethod api.CodeChallengeMethod,
) (api.ImplResponse, error) {
	host := utils.GetStringFromCtx(ctx, utils.RequestHostKey)
	if host == "" {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "could not identify user host"}
	}

	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if realm == "" {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "could not identify user realm"}
	}

	redirectURI := s.buildRedirectUriByUserAgent(userAgent, authenticationFlow, host, redirectURL, realm, codeChallenge, codeChallengeMethod)

	headers := map[string][]string{
		utils.AbsoluteLocationHeaderKey: {fmt.Sprintf("%s/%s", s.appURL, redirectURI)},
		utils.LocationHeaderKey:         {fmt.Sprintf("%s/%s", s.appURL, redirectURI)},
	}

	return api.ResponseWithHeaders(http.StatusTemporaryRedirect, headers, nil), nil
}

func (s *LoginsApiService) buildRedirectUriByUserAgent(userAgent, authenticationFlow, host, redirectURL, realm, codeChallenge string, codeChallengeMethod api.CodeChallengeMethod) string {
	var redirectURI string
	if isCLIUserAgent(userAgent) {
		redirectUriParam := buildCliRedirectUriParam(authenticationFlow, host, redirectURL)
		redirectURI = buildOidcRedirectURI(realm, utils.CLI_CLIENT_ID, redirectUriParam, utils.CliScope, codeChallenge, codeChallengeMethod)

	} else {
		redirectUriParam := fmt.Sprintf("https://%s/%s", host, "login/callback")
		redirectURI = buildOidcRedirectURI(realm, utils.RUNAI_ADMIN_UI_CLIENT_ID, redirectUriParam, utils.Scope, codeChallenge, codeChallengeMethod)
	}
	return redirectURI
}

func buildCliRedirectUriParam(authenticationFlow, host, redirectURL string) string {
	if authenticationFlow == utils.AuthenticationFlowExternalBrowser {
		return fmt.Sprintf("https://%s/%s", host, "oauth-code")
	} else {
		if redirectURL != "" {
			return redirectURL
		}
	}
	return "http://localhost:8080"
}

func buildOidcRedirectURI(realm, clientId, redirectUriParam, scopes, codeChallenge string, codeChallengeMethod api.CodeChallengeMethod) string {
	queryParams := buildCommonQueryParams(scopes, codeChallenge, codeChallengeMethod)

	queryParams.Add(utils.ClientIDParam, clientId)
	queryParams.Add(utils.RedirectURIParam, redirectUriParam)

	return fmt.Sprintf(RedirectURIFormat, realm, queryParams.Encode())
}

func buildCommonQueryParams(scopes, codeChallenge string, codeChallengeMethod api.CodeChallengeMethod) url.Values {
	queryParams := url.Values{}

	queryParams.Add(utils.ResponseTypeParam, "code")
	queryParams.Add(utils.ConnectionParam, utils.RUNAI_CLIENT_ID)
	queryParams.Add(utils.ScopeParam, scopes)

	if codeChallenge != "" {
		queryParams.Add(utils.CodeChallengeParam, codeChallenge)
		if codeChallengeMethod == "" {
			codeChallengeMethod = api.CODECHALLENGEMETHOD_PLAIN
		}
		queryParams.Add(utils.CodeChallengeMethodParam, string(codeChallengeMethod))
	}

	return queryParams
}

func isCLIUserAgent(userAgent string) bool {
	return strings.HasPrefix(userAgent, "runai-cli/")
}
