package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/auth"
	iamError "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/errors"
	gocloakInterface "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/gocloak"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/runaicloak"
)

type KeycloakAuthRepository struct {
	auth             auth.Auth
	kcClient         gocloakInterface.GoCloak
	runaicloakClient runaicloak.RunAICloakClient
}

func NewKeycloakAuthRepository(client *auth.KeycloakAuthenticator, runaicloakClient *runaicloak.Client) *KeycloakAuthRepository {
	return &KeycloakAuthRepository{
		auth:             client,
		kcClient:         client.KcClient,
		runaicloakClient: runaicloakClient,
	}
}

func (k *KeycloakAuthRepository) ExchangeExternalIdpToken(ctx context.Context, realm, token, clientId, clientSecret, idpAlias string) (*models.Tokens, error) {
	adminToken, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	tokens, err := k.runaicloakClient.ExchangeExternalIdpToken(ctx, *adminToken, realm, token, clientId, clientSecret, idpAlias)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	return &models.Tokens{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (k *KeycloakAuthRepository) GetUserInfo(ctx context.Context, realm string, accessToken string) (*models.UserInfo, error) {
	user, err := k.kcClient.GetRawUserInfo(ctx, accessToken, realm)
	if err != nil {
		return nil, err
	}
	userInfo := models.NewUserInfo(user)

	if userInfo == nil {
		return nil, fmt.Errorf("failed to parse userinfo :%v, in realm %s", user, realm)
	}
	return userInfo, nil
}

func (k *KeycloakAuthRepository) RefreshToken(ctx context.Context, realm string, refreshToken string, clientId string, clientSecret string) (*models.Tokens, error) {
	options := gocloak.TokenOptions{
		ClientID:     &clientId,
		ClientSecret: &clientSecret,
		RefreshToken: &refreshToken,
		GrantType:    gocloak.StringP("refresh_token"),
	}

	tokens, err := k.kcClient.GetToken(ctx, realm, options)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	tokensRep := models.Tokens{
		RefreshToken: tokens.RefreshToken,
		IdToken:      tokens.IDToken,
		AccessToken:  tokens.AccessToken,
	}

	return &tokensRep, nil
}

func (k *KeycloakAuthRepository) AuthenticateClient(ctx context.Context, realm string, clientId string, clientSecret string) (*models.Tokens, error) {
	options := gocloak.TokenOptions{
		ClientID:     &clientId,
		ClientSecret: &clientSecret,
		Scope:        gocloak.StringP("openid"),
		ResponseType: gocloak.StringP("id_token"),
		GrantType:    gocloak.StringP("client_credentials"),
	}
	tokens, err := k.kcClient.GetToken(ctx, realm, options)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	tokensRep := models.Tokens{
		AccessToken: tokens.AccessToken,
		IdToken:     tokens.IDToken,
		ExpiresIn:   tokens.ExpiresIn,
	}
	return &tokensRep, nil
}

func (k *KeycloakAuthRepository) ExchangeCodeForToken(ctx context.Context, realm string, redirectUri string, code string, clientId string, clientSecret string, codeVerifier string) (*models.Tokens, error) {
	if codeVerifier != "" {
		return k.exchangeCodeForTokenWithPKCE(ctx, realm, redirectUri, code, clientId, clientSecret, codeVerifier)
	}
	return k.exchangeCodeForToken(ctx, realm, redirectUri, code, clientId, clientSecret)
}

func (k *KeycloakAuthRepository) exchangeCodeForToken(ctx context.Context, realm, redirectUri, code, clientId, clientSecret string) (*models.Tokens, error) {
	options := gocloak.TokenOptions{
		Code:         &code,
		ClientID:     &clientId,
		RedirectURI:  &redirectUri,
		ClientSecret: &clientSecret,
		GrantType:    gocloak.StringP("authorization_code"),
	}

	tokens, err := k.kcClient.GetToken(ctx, realm, options)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	return &models.Tokens{
		RefreshToken: tokens.RefreshToken,
		IdToken:      tokens.IDToken,
		AccessToken:  tokens.AccessToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (k *KeycloakAuthRepository) exchangeCodeForTokenWithPKCE(ctx context.Context, realm, redirectUri, code, clientId, clientSecret, codeVerifier string) (*models.Tokens, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.auth.GetHost(), realm)

	resp, err := resty.New().R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     clientId,
			"client_secret": clientSecret,
			"code":          code,
			"redirect_uri":  redirectUri,
			"code_verifier": codeVerifier,
		}).
		Post(tokenURL)
	if err != nil {
		return nil, fmt.Errorf("PKCE token exchange failed: %w", iamError.ErrInternal)
	}
	if resp.IsError() {
		return nil, iamError.HandleKeyCloakError(&gocloak.APIError{
			Code:    resp.StatusCode(),
			Message: fmt.Sprintf("PKCE token exchange failed: %s", string(resp.Body())),
		})
	}

	var jwt gocloak.JWT
	if err := json.Unmarshal(resp.Body(), &jwt); err != nil {
		return nil, fmt.Errorf("failed to parse PKCE token response: %w", iamError.ErrInternal)
	}

	return &models.Tokens{
		RefreshToken: jwt.RefreshToken,
		IdToken:      jwt.IDToken,
		AccessToken:  jwt.AccessToken,
		ExpiresIn:    jwt.ExpiresIn,
	}, nil
}

func (k *KeycloakAuthRepository) ExchangeUsernameAndPasswordForToken(ctx context.Context, realm string, username string, password string, clientId string, clientSecret string) (*models.Tokens, error) {
	options := gocloak.TokenOptions{
		ClientID:     &clientId,
		ClientSecret: &clientSecret,
		Username:     gocloak.StringP(username),
		Password:     gocloak.StringP(password),
		GrantType:    gocloak.StringP("password"),
		Scope:        gocloak.StringP("openid email"),
	}

	tokens, err := k.kcClient.GetToken(ctx, realm, options)
	if err != nil {
		return nil, iamError.HandleKeyCloakError(err)
	}
	tokensRep := models.Tokens{
		RefreshToken: tokens.RefreshToken,
		IdToken:      tokens.IDToken,
		AccessToken:  tokens.AccessToken,
		ExpiresIn:    tokens.ExpiresIn,
	}

	return &tokensRep, nil
}

// GetClaimsFromToken extracts the claims from a JWT WITHOUT verifying its signature.
//
// Security contract: tokenString MUST be a trusted, Keycloak-issued token that this
// service obtained itself (e.g. the access token returned by a token-grant exchange) —
// never an untrusted, caller-supplied token. The returned claims MUST NOT be used for
// any authorization or access-control decision; they are only safe for non-security
// purposes such as populating the audit-log subject. Signature verification is enforced
// upstream by the API gateway. See Sonar S5659.
func (k *KeycloakAuthRepository) GetClaimsFromToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*jwt.MapClaims); ok {
		return *claims, nil
	}
	return nil, errors.New("failed to parse claims")
}

func (k *KeycloakAuthRepository) ValidateServiceAccess(ctx context.Context) error {
	_, err := k.auth.GetAdminToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get keycloak admin token. err: %s", err.Error())
	}
	return nil
}
