package models

const (
	GrantTypeRefreshToken          = "refresh_token"
	GrantTypeAppTokenDeprecated    = "app_token"
	GrantTypeAppToken              = "client_credentials"
	GrantTypeExchangeToken         = "exchange_token"
	GrantTypeAuthorizationCode     = "authorization_code"
	GrantTypeUserPassword          = "password"
	GrantTypeExternalTokenExchange = "external_token_exchange"
)

var OAuth2StandardGrantTypes = []string{
	GrantTypeAuthorizationCode,
	GrantTypeAppToken,
	GrantTypeUserPassword,
	GrantTypeRefreshToken,
}

const (
	TokenRequestFieldClientID      = "clientId"
	TokenRequestFieldClientSecret  = "clientSecret"
	TokenRequestFieldUsername      = "username"
	TokenRequestFieldPassword      = "password"
	TokenRequestFieldRefreshToken  = "refreshToken"
	TokenRequestFieldCode          = "code"
	TokenRequestFieldRedirectUri   = "redirectUri"
	TokenRequestFieldAppID         = "appId"
	TokenRequestFieldAppSecret     = "appSecret"
	TokenRequestFieldExternalToken = "externalToken"
	TokenRequestFieldCodeVerifier  = "codeVerifier"
)

type Tokens struct {
	IdToken       string `json:"idToken,omitempty"`
	RefreshToken  string `json:"refreshToken,omitempty"`
	ExternalToken string `json:"externalToken,omitempty"`
	AccessToken   string `json:"accessToken,omitempty"`
}

type ExternalTokenResponse struct {
	AccessToken string `json:"access_token"`
}

type IdpMetadata struct {
	DeleteBaseModel
	TenantId     string `gorm:"not null;unique_index"`
	Alias        string `gorm:"not null;unique_index"`
	ResourceName string `gorm:"not null"`
	Resource     string `gorm:"column:resource_content;not null"`
	ResourceType string `gorm:"not null"`
}
