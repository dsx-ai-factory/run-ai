/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"strings"

	"k8s.io/utils/ptr"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type ClientParams struct {
	CurrentState *iamModels.Client
	ClientID     string
	TenantId     string
	Realm        string
	AppHost      string
	HostTemplate string
	ClusterId    *string
}

func ApplyClientProfile(client *iamModels.Client, clientType string) {
	if client.Attributes == nil {
		client.Attributes = &iamModels.ClientAttributes{}
	}
	client.Attributes.ClientType = ptr.To(clientType)

	profile, ok := clientProfiles[clientType]
	if !ok {
		logging.WithCtx(runaictx.LoggerHook(context.Background())).Warnf("ApplyClientProfile: unknown client type %q — ClientConfig not set", clientType)
		return
	}
	if profile.isExternal != nil {
		client.Attributes.IsExternal = ptr.To(*profile.isExternal)
	}
	if profile.pkceCodeChallengeMethod != "" {
		client.Attributes.PkceCodeChallengeMethod = ptr.To(profile.pkceCodeChallengeMethod)
	}
	if profile.tokenLifespan != nil {
		client.Attributes.TokenLifespan = ptr.To(*profile.tokenLifespan)
	}
	if profile.extraAttributes != nil {
		client.Attributes.Extra = profile.extraAttributes
	}
	client.ClientConfig = &iamModels.ClientConfig{
		PublicClient:              ptr.To(profile.publicClient),
		DirectAccessGrantsEnabled: ptr.To(profile.directAccessGrants),
		StandardFlowEnabled:       ptr.To(profile.standardFlow),
		ServiceAccountsEnabled:    ptr.To(profile.serviceAccount),
	}
}

func BuildClientData(ctx context.Context, clientType string, cp ClientParams) *iamModels.Client {
	profile := clientProfiles[clientType]

	client := &iamModels.Client{}
	if cp.CurrentState != nil {
		client = cp.CurrentState
	}
	if client.Attributes == nil {
		client.Attributes = &iamModels.ClientAttributes{}
	}

	clientID := profile.clientID
	if clientID == "" {
		clientID = cp.ClientID
	}
	client.ClientID = &clientID
	client.TenantId = &cp.TenantId
	client.Enabled = ptr.To(true)

	if profile.serviceAccount {
		if cp.CurrentState != nil && cp.CurrentState.Name != nil {
			client.Name = cp.CurrentState.Name
		} else {
			client.Name = client.ClientID
		}
	}

	if err := applyClientURLs(ctx, profile, clientID, cp, client); err != nil {
		return nil
	}

	ApplyClientProfile(client, clientType)

	if cp.ClusterId != nil {
		client.Attributes.ClusterId = cp.ClusterId
	}

	return client
}

func applyClientURLs(ctx context.Context, profile clientProfile, clientID string, in ClientParams, client *iamModels.Client) error {
	var tenantURL string
	if profile.redirect == redirectDefault || profile.webOrigins {
		var err error
		tenantURL, err = utils.TemplateBaseUrl(in.HostTemplate, in.Realm)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to template base url with host template %s and realm %s in client %s: %v", in.HostTemplate, in.Realm, clientID, err)
			return err
		}
	}

	switch profile.redirect {
	case redirectDefault:
		client.RedirectURIs = getDefaultRedirectUris(tenantURL, in.AppHost, client.RedirectURIs)
		if len(profile.staticRedirectURIs) > 0 {
			uris := append(*client.RedirectURIs, profile.staticRedirectURIs...)
			client.RedirectURIs = &uris
		}
	case redirectStatic:
		uris := append([]string(nil), profile.staticRedirectURIs...)
		client.RedirectURIs = &uris
	}

	if profile.webOrigins {
		client.WebOrigins = &[]string{tenantURL, in.AppHost}
	}

	return nil
}

func getDefaultRedirectUris(tenantUrl string, appHost string, oldRedirectUris *[]string) *[]string {
	var redirectUris []string

	// Register the exact invite callback URI (with query params) as a specific allowed redirect URI.
	// Using exact URIs instead of wildcards follows the OAuth 2.0 security best practices:
	// https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics#name-countermeasures
	inviteCallback := tenantUrl + "/login/callback?" + utils.RedirectFlowQueryKey + "=" + string(utils.RedirectFlowInvite)
	redirectUris = append(redirectUris,
		tenantUrl+"/login/callback",
		inviteCallback,
		tenantUrl+"/oauth-code",
		tenantUrl,
		appHost+"/login/callback",
		appHost+"/oauth-code",
		"http://localhost",
	)

	if oldRedirectUris != nil {
		filteredOld := filterOutBroadRedirectUris(*oldRedirectUris, tenantUrl, appHost)
		redirectUris = utils.MergeListsWithUniqueness(redirectUris, filteredOld)
	}

	return &redirectUris
}

func filterOutBroadRedirectUris(uris []string, tenantUrl, appHost string) []string {
	broad := map[string]bool{
		appHost:                         true,
		appHost + "/*":                  true,
		appHost + "/":                   true,
		appHost + "/login/callback/*":   true,
		tenantUrl + "/*":                true,
		tenantUrl + "/":                 true,
		tenantUrl + "/login/callback/*": true,
		"http://localhost:*":            true,
	}
	filtered := make([]string, 0, len(uris))
	for _, uri := range uris {
		isGrafana := strings.Contains(uri, "grafana") && strings.Contains(uri, "/login/generic_oauth")
		if !isGrafana && !broad[uri] {
			filtered = append(filtered, uri)
		}
	}
	return filtered
}
