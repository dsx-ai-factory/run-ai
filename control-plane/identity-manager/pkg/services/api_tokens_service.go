/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	internalCache "github.com/run-ai/runai/control-plane/identity-manager/internal/cache"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	iamError "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/errors"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	realms "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/repositories/realms/helpers"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	legacyClientIdPrefix = "runai_agent_"
	internalClientPrefix = "runai-internal."
)

const (
	InvalidInputError           = "Invalid input"
	InvalidGrantTypeError       = "invalid grant type"
	InvalidRefreshTokenError    = "Invalid refresh token"
	AuthenticationFailedError   = "Authentication failed"
	ClientDoesntExistError      = "Client doesn't exists"
	MaxAllowedRefreshTokenError = "Maximum allowed refresh token reuse exceeded"
	TokenIsNotActiveError       = "Token is not active"
)

const (
	OAuth2ErrorInvalidRequest       = "invalid_request"
	OAuth2ErrorInvalidClient        = "invalid_client"
	OAuth2ErrorInvalidGrant         = "invalid_grant"
	OAuth2ErrorUnauthorizedClient   = "unauthorized_client"
	OAuth2ErrorUnsupportedGrantType = "unsupported_grant_type"
	OAuth2ErrorInvalidScope         = "invalid_scope"
)

type TokensApiService struct {
	iamServices iam.IamServiceContainer
	config      *config.TenantRealm
	clientCache internalCache.ClientCache
}

func NewTokensApiService(config *config.TenantRealm, iamServices *iam.IamServiceContainer) api.TokensAPIServicer {
	return &TokensApiService{
		config:      config,
		iamServices: *iamServices,
		clientCache: internalCache.NewClientCacheManager(),
	}
}

func (s *TokensApiService) GrantToken(ctx context.Context, tokenRequest api.TokenRequest) (api.ImplResponse, error) {
	if tokenRequest.GrantType == models.GrantTypeRefreshToken && tokenRequest.RefreshToken == "" {
		if cookieToken, ok := ctx.Value(utils.RefreshTokenFromCookieKey).(string); ok {
			tokenRequest.RefreshToken = cookieToken
		}
	}

	if err := s.validateTokenRequest(tokenRequest); err != nil {
		return api.Response(http.StatusBadRequest, nil), err
	}

	grantType := tokenRequest.GrantType
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("GrantToken was called with tokenRequest: %v", tokenRequest)

	audit.AuditFromCtx(ctx).WithAction(utils.AuditActionLogin)
	audit.AuditFromCtx(ctx).WithMetadata("grant_type", string(grantType))

	realmValue := ctx.Value(utils.AuthRealmKey)
	if realmValue == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "could not identify user realm"}
	}
	realm := realmValue.(string)
	if models.GrantTypeAppTokenDeprecated == grantType {
		return s.grantAppTokenDeprecated(ctx, realm, tokenRequest)
	}
	if models.GrantTypeAppToken == grantType {
		return s.grantAppToken(ctx, realm, tokenRequest)
	}

	if models.GrantTypeExternalTokenExchange == grantType {
		return s.grantExternalToken(ctx, realm, tokenRequest.ExternalToken)
	}
	if !s.shouldAudit(grantType) {
		audit.AuditFromCtx(ctx).Drop()
	}
	return s.grantUserToken(ctx, realm, tokenRequest)
}

func (s *TokensApiService) GrantTokenV2(
	ctx context.Context,
	grantType string,
	clientId string,
	clientSecret string,
	username string,
	password string,
	externalToken string,
	refreshToken string,
	code string,
	redirectUri string,
	codeVerifier string,
) (api.ImplResponse, error) {
	if grantType == "" {
		return api.Response(http.StatusBadRequest, api.OAuth2Error{
			Error:            OAuth2ErrorInvalidRequest,
			ErrorDescription: "The request is missing the grant_type parameter or the content type is not application/x-www-form-urlencoded",
		}), fmt.Errorf("missing grant_type parameter")
	}

	validGrantTypes := make(map[string]bool)
	for _, gt := range models.OAuth2StandardGrantTypes {
		validGrantTypes[gt] = true
	}
	validGrantTypes[models.GrantTypeExternalTokenExchange] = true

	if !validGrantTypes[grantType] {
		return api.Response(http.StatusBadRequest, api.OAuth2Error{
			Error:            OAuth2ErrorUnsupportedGrantType,
			ErrorDescription: fmt.Sprintf("The grant type '%s' is not supported. Supported grant types: authorization_code, client_credentials, password, refresh_token", grantType),
		}), fmt.Errorf("unsupported grant type: %s", grantType)
	}

	tokenRequest := api.TokenRequest{
		GrantType:     api.TokenRequestGrantTypeEnum(grantType),
		ClientID:      clientId,
		ClientSecret:  clientSecret,
		Username:      username,
		Password:      password,
		ExternalToken: externalToken,
		RefreshToken:  refreshToken,
		Code:          code,
		RedirectUri:   redirectUri,
		CodeVerifier:  codeVerifier,
	}

	result, err := s.GrantToken(ctx, tokenRequest)
	if err != nil {
		return s.convertToOAuth2Error(result, err)
	}

	if result.Code == http.StatusOK {
		return s.convertToOAuth2Response(result), nil
	}

	return s.convertToOAuth2Error(result, err)
}

// Deprecated
func (s *TokensApiService) AppToken(ctx context.Context, appTokenRequest api.AppTokenRequest) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("AppToken was called with appTokenRequest: %v", appTokenRequest)
	audit.AuditFromCtx(ctx).Drop()
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if realm == "app" || realm == "test" || realm == "staging" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("AppToken: realm is %s, getting realm name by client id", realm)
		if strings.HasPrefix(appTokenRequest.Name, legacyClientIdPrefix) {
			realm = strings.TrimPrefix(appTokenRequest.Name, legacyClientIdPrefix)
		} else {
			return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: "realm is not allowed to get app token"}
		}
	}
	tokens, err := s.iamServices.AuthService.AuthenticateClient(ctx, realm, appTokenRequest.Name, appTokenRequest.Secret)
	if err != nil {
		msg := fmt.Sprintf("failed to authenticate app %s, in realm %s. err: %s", appTokenRequest.Name, realm, err.Error())
		utils.LogHandler(ctx, msg, err)
		if errors.Is(err, iamError.ErrInvalidInput) {
			return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: "failed to authenticate app"}
		}
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	resp := api.AppTokenResponse{
		AccessToken: tokens.AccessToken,
		IdToken:     tokens.IdToken,
	}
	return api.Response(http.StatusOK, resp), nil
}

func (s *TokensApiService) grantAppTokenDeprecated(ctx context.Context,
	realm string, tokenRequest api.TokenRequest,
) (api.ImplResponse, error) {
	appID := tokenRequest.AppID
	if strings.HasPrefix(appID, internalClientPrefix) || strings.HasPrefix(appID, legacyClientIdPrefix) {
		audit.AuditFromCtx(ctx).Drop()
	} else {
		audit.AuditFromCtx(ctx).WithSubject(appID).WithSubjectType(utils.AuditServiceAccountSubjectType)
		audit.AuditFromCtx(ctx).WithMetadata("client_id", appID)
	}
	tokens, err := s.iamServices.AuthService.AuthenticateClient(ctx, realm, appID, tokenRequest.AppSecret)
	return s.handleTokenResponse(ctx, tokens, appID, err)
}

func (s *TokensApiService) grantAppToken(ctx context.Context,
	realm string, tokenRequest api.TokenRequest,
) (api.ImplResponse, error) {
	clientID := tokenRequest.ClientID
	if strings.HasPrefix(clientID, internalClientPrefix) {
		audit.AuditFromCtx(ctx).Drop()
	} else {
		audit.AuditFromCtx(ctx).WithSubject(clientID).WithSubjectType(utils.AuditServiceAccountSubjectType)
		audit.AuditFromCtx(ctx).WithMetadata("client_id", clientID)
	}
	tokens, err := s.iamServices.AuthService.AuthenticateClient(ctx, realm, clientID, tokenRequest.ClientSecret)
	return s.handleTokenResponse(ctx, tokens, clientID, err)
}

func (s *TokensApiService) grantUserToken(ctx context.Context, realm string, tokenRequest api.TokenRequest) (api.ImplResponse, error) {
	audit.AuditFromCtx(ctx).WithSubjectType(utils.AuditUserSubjetType)
	clientID := s.getClientIDForRequest(tokenRequest)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Using clientID: %s", clientID)
	audit.AuditFromCtx(ctx).WithMetadata("client_id", clientID)

	client, err := s.getCachedClientForRealm(ctx, realm, clientID)

	if client == nil || err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get client by client id %s. err: %s", clientID, err.Error())
		if err.Error() == ClientDoesntExistError {
			return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: fmt.Sprintf("Authentication failed. %s", ClientDoesntExistError)}
		}
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "Authentication failed. client ID error"}
	}
	if client.Secret == nil {
		s.deleteClientFromCache(ctx, realm, clientID)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("grantToken: failed to get client secret for client id %s in realm %s", clientID, realm)
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "Authentication failed. client authentication details not found."}
	}
	tokens, err := s.getTokens(ctx, string(tokenRequest.GrantType), realm, client, tokenRequest)
	if err != nil {
		s.deleteClientFromCache(ctx, realm, clientID)
		return s.handleTokenResponse(ctx, tokens, clientID, err)
	}
	sub, err := s.getSubFromToken(ctx, tokens.AccessToken)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get sub from token for auditing: %s", err.Error())
	} else {
		audit.AuditFromCtx(ctx).WithSubject(sub)
	}
	return s.handleTokenResponse(ctx, tokens, clientID, err)
}

func (s *TokensApiService) grantExternalToken(ctx context.Context, realm string, accessToken string) (api.ImplResponse, error) {
	idpAlias, err := s.iamServices.IdpService.GetIdpAliasForExchangeToken(ctx, realm)
	if err != nil && !errors.Is(err, iamError.ErrNotFound) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error searching for idp alias supported for exchange in realm %s. err: %s", realm, err.Error())
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: ""}
	}

	if errors.Is(err, iamError.ErrNotFound) || idpAlias == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("unable to find idp alias supported for exchange in realm %s.", realm)
		return api.Response(http.StatusMethodNotAllowed, nil), utils.Error{Code: http.StatusMethodNotAllowed, Message: "missing support in oidc configuration"}
	}

	client, err := s.getCachedClientForRealm(ctx, realm, utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID)
	if err != nil || client == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Errorf("failed to get client by client id %s. err: %v", utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, err)
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: ""}
	}
	if client.Secret == nil {
		if client.ID == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get client secret for %s in realm %s: client is missing required fields", utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, realm)
			return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: ""}
		}
		secret, secretErr := s.iamServices.ClientService.GetClientSecret(ctx, realm, *client.ID)
		if secretErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get client secret for %s in realm %s. err: %v", utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID, realm, secretErr)
			return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: ""}
		}
		client.Secret = &secret
	}
	audit.AuditFromCtx(ctx).WithMetadata("client_id", *client.ClientID)
	tokens, err := s.iamServices.AuthService.ExchangeExternalIdpToken(ctx, realm, accessToken, *client.ClientID, *client.Secret, *idpAlias)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).
			Errorf("failed to exchange external token from idp alias %v in realm %s. err: %v", idpAlias, realm, err.Error())
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to exchange external token"}
	}

	return s.handleTokenResponse(ctx, tokens, *client.ClientID, nil)
}

func (s *TokensApiService) getCachedClientForRealm(ctx context.Context, realm string, clientID string) (*iamModels.Client, error) {
	return s.clientCache.GetClientForRealm(ctx, realm, clientID, func() (*iamModels.Client, error) {
		return s.getClientForRealmNoCache(ctx, realm, clientID)
	})
}

func (s *TokensApiService) getClientForRealmNoCache(ctx context.Context, realm string, clientID string) (*iamModels.Client, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get no cached client was called with clientID: %s, in realm %s", clientID, realm)
	client, err := s.iamServices.ClientService.GetClientByClientId(ctx, realm, clientID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get client by client id %s in realm %s. err: %s", clientID, realm, err.Error())
		return nil, err
	}
	if client == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get client by client id %s in realm %s. client doesn't exist", clientID, realm)
		return nil, fmt.Errorf("%s", ClientDoesntExistError)
	}
	if client.Secret == nil {
		if client.ID == nil {
			err := fmt.Errorf("client %s in realm %s is missing required fields", clientID, realm)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%v", err)
			return nil, err
		}
		secret, secretErr := s.iamServices.ClientService.GetClientSecret(ctx, realm, *client.ID)
		if secretErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("getClientForRealmNoCache: failed to get client secret for %s in realm %s: %v", clientID, realm, secretErr)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("getClientForRealmNoCache: fetched secret for %s in realm %s, empty=%v", clientID, realm, secret == "")
			client.Secret = &secret
		}
	}
	return client, nil
}

func (s *TokensApiService) deleteClientFromCache(ctx context.Context, realm string, clientID string) {
	if err := s.clientCache.DeleteClientFromRealm(ctx, realm, clientID); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to delete client from cache, clientId: %s in realm %s. err: %s", clientID, realm, err.Error())
	}
}

func (s *TokensApiService) getTokens(ctx context.Context, grantType string,
	realm string, client *iamModels.Client, tokenRequest api.TokenRequest,
) (*iamModels.Tokens, error) {
	switch grantType {
	case models.GrantTypeAuthorizationCode:
		return s.iamServices.AuthService.ExchangeCodeForToken(ctx,
			realm,
			tokenRequest.RedirectUri,
			tokenRequest.Code,
			*client.ClientID,
			*client.Secret,
			tokenRequest.CodeVerifier)
	case models.GrantTypeExchangeToken:
		return s.iamServices.AuthService.ExchangeCodeForToken(ctx,
			realm,
			tokenRequest.RedirectUri,
			tokenRequest.Code,
			*client.ClientID,
			*client.Secret,
			tokenRequest.CodeVerifier)
	case models.GrantTypeRefreshToken:
		return s.iamServices.AuthService.RefreshToken(ctx,
			realm,
			tokenRequest.RefreshToken,
			*client.ClientID,
			*client.Secret)
	case models.GrantTypeUserPassword:
		audit.AuditFromCtx(ctx).WithSubject(tokenRequest.Username)
		return s.iamServices.AuthService.ExchangeUsernameAndPasswordForToken(ctx,
			realm,
			tokenRequest.Username,
			tokenRequest.Password,
			*client.ClientID,
			*client.Secret)
	}
	return nil, errors.New(InvalidGrantTypeError)
}

func (s *TokensApiService) getClientIDForRequest(tokenRequest api.TokenRequest) string {
	if tokenRequest.ClientID != "" {
		return tokenRequest.ClientID
	}
	if (tokenRequest.GrantType == models.GrantTypeUserPassword) || (tokenRequest.GrantType == models.GrantTypeExchangeToken) {
		return utils.CLI_CLIENT_ID
	}
	return s.config.DefaultClientId
}

func (s *TokensApiService) handleTokenResponse(ctx context.Context, tokens *iamModels.Tokens, clientID string, err error) (api.ImplResponse, error) {
	if err != nil {
		msg := fmt.Sprintf("authentication failed with error: %s", err.Error())
		utils.LogHandler(ctx, msg, err)
		errorString := err.Error()
		statusCode := http.StatusUnauthorized
		if strings.Contains(errorString, TokenIsNotActiveError) {
			msg = TokenIsNotActiveError
		} else if strings.Contains(errorString, InvalidRefreshTokenError) {
			msg = InvalidRefreshTokenError
		} else if strings.Contains(errorString, MaxAllowedRefreshTokenError) {
			msg = MaxAllowedRefreshTokenError
		} else if strings.Contains(errorString, AuthenticationFailedError) {
			msg = AuthenticationFailedError
		} else if strings.Contains(errorString, InvalidGrantTypeError) {
			statusCode = http.StatusBadRequest
			msg = InvalidGrantTypeError
		} else if errors.Is(err, iamError.ErrInvalidInput) {
			statusCode = http.StatusBadRequest
			msg = InvalidInputError
		} else if errors.Is(err, iamError.ErrUnauthorized) {
			statusCode = http.StatusUnauthorized
			msg = AuthenticationFailedError
		} else {
			statusCode = http.StatusInternalServerError
			msg = "Internal server error"
		}
		return api.Response(statusCode, nil), utils.Error{Code: statusCode, Message: msg}

	}
	resp := api.TokenResponse{
		AccessToken: tokens.AccessToken,
		IdToken:     tokens.IdToken,
		ExpiresIn:   int32(tokens.ExpiresIn),
	}
	if s.shouldReturnRefreshToken(clientID) {
		resp.RefreshToken = tokens.RefreshToken
	}

	cookie := utils.FormatSecureCookie(
		utils.RefreshTokenCookieName,
		tokens.RefreshToken,
		utils.RefreshTokenCookiePath,
		realms.SsoSessionMaxLifespanSeconds,
	)
	headers := utils.CreateSetCookieHeaders(cookie)

	return api.ResponseWithHeaders(http.StatusOK, headers, resp), nil
}

func (s *TokensApiService) getSubFromToken(ctx context.Context, token string) (string, error) {
	claims, err := s.iamServices.AuthService.GetClaimsFromToken(ctx, token)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("extracting sub from token claims, claims: %v", claims)
	if err != nil {
		return "", err
	}
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}
	return "", errors.New("failed to get sub from token")
}

func (s *TokensApiService) missingFieldErr(field string) error {
	return fmt.Errorf("missing required field: %s", field)
}

func (s *TokensApiService) shouldAudit(grantType api.TokenRequestGrantTypeEnum) bool {
	return grantType != models.GrantTypeRefreshToken
}

func (s *TokensApiService) shouldReturnRefreshToken(clientID string) bool {
	return clientID != utils.RUNAI_ADMIN_UI_CLIENT_ID
}

func (s *TokensApiService) convertToOAuth2Response(result api.ImplResponse) api.ImplResponse {
	internalResponse, ok := result.Body.(api.TokenResponse)
	if !ok {
		return result
	}

	oauth2Response := api.OAuth2TokenResponse{
		AccessToken:  internalResponse.AccessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int32(internalResponse.ExpiresIn),
		RefreshToken: internalResponse.RefreshToken,
		IdToken:      internalResponse.IdToken,
	}

	return api.ResponseWithHeaders(http.StatusOK, result.Headers, oauth2Response)
}

func (s *TokensApiService) convertToOAuth2Error(result api.ImplResponse, err error) (api.ImplResponse, error) {
	errorCode := OAuth2ErrorInvalidRequest
	description := "Invalid input"

	if err != nil {
		description = err.Error()
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "invalid") && strings.Contains(errMsg, "grant") {
			errorCode = OAuth2ErrorInvalidGrant
		} else if strings.Contains(errMsg, "invalid") && strings.Contains(errMsg, "client") {
			errorCode = OAuth2ErrorInvalidClient
		} else if strings.Contains(errMsg, "unsupported") && strings.Contains(errMsg, "grant") {
			errorCode = OAuth2ErrorUnsupportedGrantType
		} else if strings.Contains(errMsg, "unauthorized") {
			errorCode = OAuth2ErrorUnauthorizedClient
		} else if strings.Contains(errMsg, "invalid") && strings.Contains(errMsg, "scope") {
			errorCode = OAuth2ErrorInvalidScope
		}
	}

	statusCode := result.Code
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	return api.Response(statusCode, api.OAuth2Error{
		Error:            errorCode,
		ErrorDescription: description,
	}), err
}

func (s *TokensApiService) validateTokenRequest(tokenRequest api.TokenRequest) error {
	switch tokenRequest.GrantType {
	case models.GrantTypeAppToken:
		if tokenRequest.ClientID == "" {
			return s.missingFieldErr(models.TokenRequestFieldClientID)
		}
		if tokenRequest.ClientSecret == "" {
			return s.missingFieldErr(models.TokenRequestFieldClientSecret)
		}
	case models.GrantTypeUserPassword:
		if tokenRequest.Username == "" {
			return s.missingFieldErr(models.TokenRequestFieldUsername)
		}
		if tokenRequest.Password == "" {
			return s.missingFieldErr(models.TokenRequestFieldPassword)
		}
	case models.GrantTypeRefreshToken:
		if tokenRequest.RefreshToken == "" {
			return s.missingFieldErr(models.TokenRequestFieldRefreshToken)
		}
	case models.GrantTypeAuthorizationCode:
		if tokenRequest.Code == "" {
			return s.missingFieldErr(models.TokenRequestFieldCode)
		}
		if tokenRequest.RedirectUri == "" {
			return s.missingFieldErr(models.TokenRequestFieldRedirectUri)
		}
	case models.GrantTypeExchangeToken:
		if tokenRequest.Code == "" {
			return s.missingFieldErr(models.TokenRequestFieldCode)
		}
		if tokenRequest.RedirectUri == "" {
			return s.missingFieldErr(models.TokenRequestFieldRedirectUri)
		}
	case models.GrantTypeExternalTokenExchange:
		if tokenRequest.ExternalToken == "" {
			return s.missingFieldErr(models.TokenRequestFieldExternalToken)
		}

	case models.GrantTypeAppTokenDeprecated:
		if tokenRequest.AppID == "" {
			return s.missingFieldErr(models.TokenRequestFieldAppID)
		}
		if tokenRequest.AppSecret == "" {
			return s.missingFieldErr(models.TokenRequestFieldAppSecret)
		}
	}

	if err := validateCodeVerifier(tokenRequest.CodeVerifier); err != nil {
		return err
	}

	return nil
}

var codeVerifierPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)

func validateCodeVerifier(codeVerifier string) error {
	if codeVerifier == "" {
		return nil
	}
	if len(codeVerifier) < 43 || len(codeVerifier) > 128 {
		return fmt.Errorf("codeVerifier must be between 43 and 128 characters")
	}
	if !codeVerifierPattern.MatchString(codeVerifier) {
		return fmt.Errorf("codeVerifier contains invalid characters")
	}
	return nil
}
