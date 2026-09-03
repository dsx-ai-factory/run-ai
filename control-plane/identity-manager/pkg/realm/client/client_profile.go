/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"time"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

const (
	ClientTypeWorkloadsAuth = "WorkloadsAuth"
	ClientTypeRunaiAdminUI  = "RunaiAdminUI"
	ClientTypeCLI           = "CLI"
	ClientTypeRunaiCLI      = "RunaiCLI"
	ClientTypeRunai         = "Runai"
	ClientTypeTokenExchange = "TokenExchange"
	ClientTypeAgent         = "agent"
	ClientTypeInstaller     = "runai-installer"
	ClientTypeUserApp       = "user-application"
	ClientTypeTenantApp     = "tenant-application"
	ClientTypeMCP           = "MCP"
)

const DefaultServiceAccountTokenLifespan = 30 * time.Minute

type redirectStrategy int

const (
	redirectNone redirectStrategy = iota
	redirectDefault
	redirectStatic
)

type clientProfile struct {
	publicClient            bool
	directAccessGrants      bool
	standardFlow            bool
	serviceAccount          bool
	emailScopeDefault       bool
	isExternal              *bool
	pkceCodeChallengeMethod string
	tokenLifespan           *time.Duration

	clientID           string
	redirect           redirectStrategy
	staticRedirectURIs []string
	webOrigins         bool
	extraAttributes    map[string]string
}

var mcpRedirectURIs = []string{
	"http://localhost/callback",
	"cursor://anysphere.cursor-mcp/oauth/callback",
	"http://localhost/oauth/callback/debug",
	"http://localhost/oauth/callback",
}

var workloadsAuthAttributes = map[string]string{
	"backchannel.logout.revoke.offline.tokens":   "false",
	"backchannel.logout.session.required":        "true",
	"client_credentials.use_refresh_token":       "false",
	"refresh_token_max_age":                      "600",
	"display.on.consent.screen":                  "false",
	"require.pushed.authorization.requests":      "false",
	"acr.loa.map":                                "{}",
	"tls-client-certificate-bound-access-tokens": "false",
	"token.response.type.bearer.lower-case":      "false",
	"use.jwks.url":                               "false",
	"use.refresh.tokens":                         "true",
}

var clientProfiles = map[string]clientProfile{
	ClientTypeRunaiAdminUI: {
		clientID:                utils.RUNAI_ADMIN_UI_CLIENT_ID,
		standardFlow:            true,
		emailScopeDefault:       true,
		pkceCodeChallengeMethod: "S256",
		redirect:                redirectDefault,
		webOrigins:              true,
		staticRedirectURIs:      []string{"http://localhost/login/callback"},
	},
	ClientTypeRunaiCLI: {
		clientID:           utils.RUNAI_CLI_CLIENT_ID,
		publicClient:       true,
		directAccessGrants: true,
		standardFlow:       true,
		emailScopeDefault:  true,
		redirect:           redirectDefault,
		webOrigins:         true,
	},
	ClientTypeCLI: {
		clientID:           utils.CLI_CLIENT_ID,
		directAccessGrants: true,
		standardFlow:       true,
		emailScopeDefault:  true,
		redirect:           redirectDefault,
		webOrigins:         true,
	},
	ClientTypeRunai: {
		clientID:           utils.RUNAI_CLIENT_ID,
		publicClient:       true,
		directAccessGrants: true,
		emailScopeDefault:  true,
		redirect:           redirectDefault,
		webOrigins:         true,
	},
	ClientTypeTokenExchange: {
		clientID:      utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID,
		standardFlow:  true,
		tokenLifespan: ptr.To(10 * time.Minute),
		redirect:      redirectDefault,
		webOrigins:    true,
	},
	ClientTypeMCP: {
		clientID:                utils.RUNAI_MCP_CLIENT_ID,
		standardFlow:            true,
		emailScopeDefault:       true,
		pkceCodeChallengeMethod: "S256",
		redirect:                redirectStatic,
		staticRedirectURIs:      mcpRedirectURIs,
	},
	ClientTypeWorkloadsAuth: {
		standardFlow:       true,
		emailScopeDefault:  true,
		redirect:           redirectDefault,
		staticRedirectURIs: []string{"*"},
		extraAttributes:    workloadsAuthAttributes,
	},
	ClientTypeAgent: {
		standardFlow:   true,
		serviceAccount: true,
		isExternal:     ptr.To(false),
	},
	ClientTypeInstaller: {
		standardFlow:   true,
		serviceAccount: true,
		isExternal:     ptr.To(false),
		tokenLifespan:  ptr.To(2 * time.Hour),
	},
	ClientTypeUserApp: {
		standardFlow:   true,
		serviceAccount: true,
		tokenLifespan:  ptr.To(DefaultServiceAccountTokenLifespan),
	},
	ClientTypeTenantApp: {
		standardFlow:   true,
		serviceAccount: true,
		isExternal:     ptr.To(true),
		tokenLifespan:  ptr.To(DefaultServiceAccountTokenLifespan),
	},
}

func ProfileClientID(clientType string) string {
	return clientProfiles[clientType].clientID
}

func ProfileEmailScopeDefault(clientType string) bool {
	return clientProfiles[clientType].emailScopeDefault
}
