/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

type RedirectFlow string

const (
	RedirectFlowQueryKey              = "runai_flow"
	RedirectFlowInvite   RedirectFlow = "invite"
)

// AuditLog
const (
	AuditActionLogin               = "Login"
	AuditActionLogout              = "Logout"
	AuditActionResetPassword       = "Reset password"
	AuditActionChangePassword      = "Update password"
	AuditEntityTypeIdp             = "identity provider"
	AuditEntityTypeUser            = "user"
	AuditEntityTypeServiceAccount  = "service account"
	AuditEntityTypeAccessKey       = "access key"
	AuditUserSubjetType            = "User"
	AuditServiceAccountSubjectType = "ServiceAccount"
)

// context keys
type contextKey string

const (
	AuthRealmKey   = contextKey("AuthRealm")
	RequestHostKey = "requestHost"
)

// entity types
const (
	ServiceAccount  = "service account"
	User            = "user"
	InternalClient  = "internal client"
	AccessKey       = "access key"
	SecuritySetting = "security setting"
)

// identifiers
const (
	Name      = "name"
	EMAIL     = "email"
	ID        = "Id"
	ClientId  = "client id"
	Realm     = "realm"
	SessionID = "session id"
	Tenant    = "tenant"
)

// realms
const (
	MASTER_REALM      = "master"
	SSO_DEFAULT_REALM = "sso"
)

// idps
const (
	// idp mapper names
	IdpExternalUserMapperName = "External User Mapper"
	IdpUpnMapperName          = "Upn Mapper"
	IdpEmailMapperName        = "Email Mapper"
	IdpFirstNameMapperName    = "First Name Mapper"
	IdpLastNameMapperName     = "Last Name Mapper"
	IdpGroupsMapperName       = "Groups Mapper"
	IdpSupGroupsMapperName    = "Supplementary Groups Mapper"
	IdpUidMapperName          = "Uid Mapper"
	IdpGidMapperName          = "Gid Mapper"
	OidcMandatoryClaim        = "oidc-mandatory-claim"
)

// clients
const (
	RUNAI_TOKEN_EXCHANGE_CLIENT_ID = "runai-token-exchange" // #nosec G101
	RUNAI_ADMIN_UI_CLIENT_ID       = "runai-admin-ui"
	RUNAI_CLIENT_ID                = "runai"
	RUNAI_CLI_CLIENT_ID            = "runai-cli"
	RunaiInstallerClientId         = "runai-internal.installer"
	CLI_CLIENT_ID                  = "cli"
	RUNAI_MCP_CLIENT_ID            = "runai-mcp"
	ClientTypeWorkloadsAuth        = "workloads_auth"
)

// realms authflows
const (
	LOGIN_FLOW_NAME   = "runai first login"
	BROWSER_FLOW_NAME = "browser"
)

// SAML metadataXML types
const (
	File = "file"
	Url  = "url"
)

// SAML config fields
const (
	MetadataXmlUrl         = "metadataXmlUrl"
	MetadataXmlFile        = "metadataXmlFile"
	EntityId               = "entityId"
	FileName               = "fileName"
	MetadataXmlType        = "metadataXmlType"
	SigningCertificate     = "signingCertificate"
	ValidateSignature      = "validateSignature"
	SingleSignOnServiceUrl = "singleSignOnServiceUrl"
)

// OIDC config fields
const (
	Oidc_ClientId       = "clientId"
	Oidc_ClientSecret   = "clientSecret"
	DiscoverDocumentUrl = "discoverDocumentUrl"
	Scope               = "openid email profile offline_access"
	ScopesSeperator     = " "
	CliScope            = "openid email"
)

// OCP config fields
const (
	IdpBaseUrl         = "baseUrl"
	IdpBaseUrlGetParam = "idpBaseUrl"
	Ocp_ClientId       = "clientId"
	Ocp_ClientSecret   = "clientSecret"
)

// Default scope
const (
	Scopes     string = "defaultScope"
	ScopeRegex string = `^([a-zA-Z0-9-._:]+)(\s[a-zA-Z0-9-._:]+)*$`
)

const (
	LocationHeaderKey         = "Location"
	AbsoluteLocationHeaderKey = "Absolute-Location"
)

const (
	RefreshTokenCookieName    = "refresh-token"
	RefreshTokenCookiePath    = "/api/v1/token"
	RefreshTokenFromCookieKey = "RefreshTokenFromCookie"
)

const (
	ResponseTypeParam        = "response_type"
	ConnectionParam          = "connection"
	ScopeParam               = "scope"
	RedirectURIParam         = "redirect_uri"
	ClientIDParam            = "client_id"
	CodeChallengeParam       = "code_challenge"
	CodeChallengeMethodParam = "code_challenge_method"
)

const (
	AuthenticationFlowExternalBrowser = "external-browser"
)

// settings categories
const (
	SecurityCategory = "security"
)

// keycloak user attributes
const (
	IS_LOCAL_USER = "runai.is_local"
	GROUPS        = "GROUPS"
)
