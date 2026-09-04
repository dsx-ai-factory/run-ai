/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import "strings"

const (
	AuthProtocolOidc     = "openid-connect"
	AuthTypeClientSecret = "client-secret"
)

// default application attributes
const (
	Oauth2AuthGrantEnabled      = "oauth2.device.authorization.grant.enabled"
	OidcGrantEnabled            = "oidc.ciba.grant.enabled"
	LogoutRedirectUris          = "post.logout.redirect.uris"
	LogoutRedirectUrisDefault   = "+"
	LogoutRedirectUrisDelimiter = "##"
	TokenLifespanAttribute      = "access.token.lifespan"
	LightWeightToken            = "client.use.lightweight.access.token.enabled"
	PkceCodeChallengeMethodAttr = "pkce.code.challenge.method"
)

func BuildLogoutRedirectUrisValue(customUri string) string {
	if customUri == "" {
		return LogoutRedirectUrisDefault
	}
	return LogoutRedirectUrisDefault + LogoutRedirectUrisDelimiter + customUri
}

func IsValidLogoutRedirectUris(value string) bool {
	// "*" is preserved for existing customers who had a wildcard redirect URI before the security fix
	if value == LogoutRedirectUrisDefault || value == "*" {
		return true
	}
	uri, found := strings.CutPrefix(value, LogoutRedirectUrisDefault+LogoutRedirectUrisDelimiter)
	return found && uri != ""
}

func GetLogoutRedirectUrisOrDefault(attributes *map[string]string) string {
	if attributes != nil {
		if val, ok := (*attributes)[LogoutRedirectUris]; ok && IsValidLogoutRedirectUris(val) {
			return val
		}
	}
	return LogoutRedirectUrisDefault
}

// keycloak scope names
const (
	EmailScope = "email"
)

// runai custom attributes
const (
	IsLocalUser    = "runai.is_local"
	CreatedBy      = "runai.created.by"
	LastLoginTime  = "runai.login.time"
	UpdateTime     = "runai.update.time"
	CreationTime   = "runai.creation.time"
	ClusterIdAttr  = "runai.cluster_id"
	ClientTypeAttr = "runai.client_type"
	IsExternal     = "runai.is_external"
)

// custom user attributes
const (
	UID                 = "UID"
	GID                 = "GID"
	Upn                 = "upn"
	Groups              = "GROUPS"
	SupplementaryGroups = "SUPPLEMENTARYGROUPS"
)

// flow alias
const (
	BrowserFlowAlias = "browser"
)

// auth flows providers ids
const (
	IdpProviderID = "identity-provider-redirector"
)
