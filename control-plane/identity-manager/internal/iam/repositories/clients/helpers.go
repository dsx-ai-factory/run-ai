/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"fmt"
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/repositories/protocol_mapper"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
	ptr "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils/pointer"
)

var reservedClientAttrKeys = map[string]struct{}{
	utils.ClientTypeAttr:              {},
	utils.PkceCodeChallengeMethodAttr: {},
	utils.LogoutRedirectUris:          {},
	utils.Oauth2AuthGrantEnabled:      {},
	utils.OidcGrantEnabled:            {},
	utils.IsExternal:                  {},
	utils.UpdateTime:                  {},
	utils.CreationTime:                {},
	utils.CreatedBy:                   {},
	utils.LastLoginTime:               {},
	utils.TokenLifespanAttribute:      {},
	utils.ClusterIdAttr:               {},
}

func fromClientModel(client *models.Client) *gocloak.Client {
	kcClient := &gocloak.Client{}
	kcClient.Attributes = &map[string]string{}
	kcClient.ProtocolMappers = &[]gocloak.ProtocolMapperRepresentation{}
	if client.Name == nil {
		kcClient.Name = client.ClientID
	} else {
		kcClient.Name = client.Name
	}
	if client.ID != nil {
		kcClient.ID = client.ID
	}
	kcClient.ClientID = client.ClientID
	if client.Enabled == nil {
		kcClient.Enabled = gocloak.BoolP(true)
	} else {
		kcClient.Enabled = client.Enabled
	}
	kcClient.RedirectURIs = client.RedirectURIs
	kcClient.WebOrigins = client.WebOrigins
	if client.ClientConfig == nil {
		client.ClientConfig = &models.ClientConfig{}
	}
	if client.Attributes == nil {
		client.Attributes = &models.ClientAttributes{}
	}
	kcClient.PublicClient = client.ClientConfig.PublicClient
	kcClient.ServiceAccountsEnabled = client.ClientConfig.ServiceAccountsEnabled
	kcClient.StandardFlowEnabled = client.ClientConfig.StandardFlowEnabled
	kcClient.DirectAccessGrantsEnabled = client.ClientConfig.DirectAccessGrantsEnabled
	if client.GetClientType() == "" {
		return nil
	}
	setKcDefaults(kcClient, client)
	setProtocolMappers(kcClient, client)
	setAttributes(kcClient, client)
	return kcClient
}

func setKcDefaults(kcClient *gocloak.Client, client *models.Client) {
	kcClient.Protocol = gocloak.StringP(utils.AuthProtocolOidc)
	if client.ClientConfig.PublicClient != nil && !*client.ClientConfig.PublicClient {
		kcClient.ClientAuthenticatorType = gocloak.StringP(utils.AuthTypeClientSecret)
	}
	kcClient.FullScopeAllowed = gocloak.BoolP(false)
	kcClient.FrontChannelLogout = gocloak.BoolP(false)
	kcClient.AuthorizationServicesEnabled = gocloak.BoolP(false)
	kcClient.ImplicitFlowEnabled = gocloak.BoolP(false)
	setIamClientAttribute(kcClient, utils.LogoutRedirectUris, utils.GetLogoutRedirectUrisOrDefault(kcClient.Attributes))
	setIamClientAttribute(kcClient, utils.Oauth2AuthGrantEnabled, "false")
	setIamClientAttribute(kcClient, utils.OidcGrantEnabled, "false")
}

func setProtocolMappers(kcClient *gocloak.Client, client *models.Client) {
	cType := client.GetClientType()
	kcClient.ProtocolMappers = &[]gocloak.ProtocolMapperRepresentation{
		protocol_mapper.NewRunaiAudienceProtocolMapper(),
		protocol_mapper.NewRunaiDummyAudienceProtocolMapper(),
		protocol_mapper.NewTenantIdProtocolMapper(*client.TenantId),
		protocol_mapper.NewSubjectUidProtocolMapper(),
		protocol_mapper.NewClientTypeProtocolMapper(cType),
	}
	if client.IsServiceAccountsEnabled() {
		if cType == models.ClientTypeUserApp {
			*kcClient.ProtocolMappers = append(*kcClient.ProtocolMappers,
				protocol_mapper.NewCreatedByToSubProtocolMapper(*client.Attributes.CreatedBy),
				protocol_mapper.NewCreatedByToEmailProtocolMapper(*client.Attributes.CreatedBy),
				protocol_mapper.NewAccessKeyGroupsProtocolMapper(*client.TenantId),
				protocol_mapper.NewAccessKeyUIDProtocolMapper(*client.TenantId),
				protocol_mapper.NewAccessKeyGIDProtocolMapper(*client.TenantId),
				protocol_mapper.NewAccessKeySupplementaryGroupsProtocolMapper(*client.TenantId),
			)
		} else {
			*kcClient.ProtocolMappers = append(*kcClient.ProtocolMappers,
				protocol_mapper.NewEmailToClientIdProtocolMapper(),
				protocol_mapper.NewSubToClientIdProtocolMapper(),
			)
		}
	} else {
		*kcClient.ProtocolMappers = append(*kcClient.ProtocolMappers, protocol_mapper.NewSubToEmailProtocolMapper(),
			protocol_mapper.NewRunaiGroupsProtocolMapper(false, *client.TenantId),
			protocol_mapper.NewRunaiGroupsCountProtocolMapper(false),
			protocol_mapper.NewGIDProtocolMapper(),
			protocol_mapper.NewSupplementaryGroupsProtocolMapper(),
			protocol_mapper.NewUIDProtocolMapper(),
			protocol_mapper.NewIdentityProviderMapper(),
		)
	}
}

func clientTypeToAttrValue(clientType string) string {
	switch clientType {
	case models.ClientTypeWorkloadsAuth:
		return "workloads_auth"
	default:
		return clientType
	}
}

func setAttributes(kcClient *gocloak.Client, client *models.Client) {
	clientAttributes := client.Attributes
	currentTime := utils.UnixTimestampToString()
	if clientAttributes.ClientType != nil {
		setIamClientAttribute(kcClient, utils.ClientTypeAttr, clientTypeToAttrValue(client.GetClientType()))
	}
	if clientAttributes.IsExternal != nil {
		setIamClientAttribute(kcClient, utils.IsExternal, strconv.FormatBool(*clientAttributes.IsExternal))
	}
	if clientAttributes.UpdatedAt != nil {
		setIamClientAttribute(kcClient, utils.UpdateTime, *clientAttributes.UpdatedAt)
	} else {
		setIamClientAttribute(kcClient, utils.UpdateTime, currentTime)
	}
	if clientAttributes.CreatedAt != nil {
		setIamClientAttribute(kcClient, utils.CreationTime, *clientAttributes.CreatedAt)
	} else {
		setIamClientAttribute(kcClient, utils.CreationTime, currentTime)
	}
	if clientAttributes.CreatedBy != nil {
		setIamClientAttribute(kcClient, utils.CreatedBy, *clientAttributes.CreatedBy)
	}
	if clientAttributes.LastLogin != nil {
		setIamClientAttribute(kcClient, utils.LastLoginTime, *clientAttributes.LastLogin)
	}
	if clientAttributes.TokenLifespan != nil {
		setIamClientAttribute(kcClient, utils.TokenLifespanAttribute, fmt.Sprintf("%d", int64(clientAttributes.TokenLifespan.Seconds())))
	}
	if clientAttributes.ClusterId != nil {
		setIamClientAttribute(kcClient, utils.ClusterIdAttr, *clientAttributes.ClusterId)
	}
	if clientAttributes.PkceCodeChallengeMethod != nil {
		setIamClientAttribute(kcClient, utils.PkceCodeChallengeMethodAttr, *clientAttributes.PkceCodeChallengeMethod)
	}
	for k, v := range clientAttributes.Extra {
		if _, reserved := reservedClientAttrKeys[k]; reserved {
			continue
		}
		setIamClientAttribute(kcClient, k, v)
	}
}

func setIamClientAttribute(kc *gocloak.Client, key, value string) {
	if kc.Attributes == nil {
		kc.Attributes = &map[string]string{}
	}
	(*kc.Attributes)[key] = value
}

func toClientModel(kcClient *gocloak.Client) *models.Client {
	client := &models.Client{}
	client.ID = kcClient.ID
	client.Name = kcClient.Name
	client.ClientID = kcClient.ClientID
	client.Enabled = kcClient.Enabled
	client.RedirectURIs = kcClient.RedirectURIs
	client.WebOrigins = kcClient.WebOrigins
	client.Secret = kcClient.Secret
	client.ClientConfig = &models.ClientConfig{
		PublicClient:              kcClient.PublicClient,
		ServiceAccountsEnabled:    kcClient.ServiceAccountsEnabled,
		StandardFlowEnabled:       kcClient.StandardFlowEnabled,
		DirectAccessGrantsEnabled: kcClient.DirectAccessGrantsEnabled,
	}
	client.Attributes = &models.ClientAttributes{}
	client.Attributes.ClientType = ptr.ToString((*kcClient.Attributes)[utils.ClientTypeAttr])
	if isExternal, err := strconv.ParseBool((*kcClient.Attributes)[utils.IsExternal]); err == nil {
		client.Attributes.IsExternal = ptr.ToBool(isExternal)
	}
	if createdBy, ok := (*kcClient.Attributes)[utils.CreatedBy]; ok {
		client.Attributes.CreatedBy = ptr.ToString(createdBy)
	}
	if updatedAt, ok := (*kcClient.Attributes)[utils.UpdateTime]; ok {
		client.Attributes.UpdatedAt = ptr.ToString(updatedAt)
	}
	if createdAt, ok := (*kcClient.Attributes)[utils.CreationTime]; ok {
		client.Attributes.CreatedAt = ptr.ToString(createdAt)
	}
	if lastLogin, ok := (*kcClient.Attributes)[utils.LastLoginTime]; ok {
		client.Attributes.LastLogin = ptr.ToString(lastLogin)
	}
	if tokenLifespan, ok := (*kcClient.Attributes)[utils.TokenLifespanAttribute]; ok {
		client.Attributes.TokenLifespan = utils.StringToTimeDuration(&tokenLifespan)
	}
	if clusterId, ok := (*kcClient.Attributes)[utils.ClusterIdAttr]; ok {
		client.Attributes.ClusterId = ptr.ToString(clusterId)
	}
	if oauth2AuthGrant, ok := (*kcClient.Attributes)[utils.Oauth2AuthGrantEnabled]; ok {
		client.Attributes.Oauth2AuthGrantEnabled = ptr.ToString(oauth2AuthGrant)
	}
	if oidcGrant, ok := (*kcClient.Attributes)[utils.OidcGrantEnabled]; ok {
		client.Attributes.OidcGrantEnabled = ptr.ToString(oidcGrant)
	}
	if logoutRedirectUris, ok := (*kcClient.Attributes)[utils.LogoutRedirectUris]; ok {
		client.Attributes.LogoutRedirectUris = ptr.ToString(logoutRedirectUris)
	}
	if pkceCodeChallengeMethod, ok := (*kcClient.Attributes)[utils.PkceCodeChallengeMethodAttr]; ok {
		client.Attributes.PkceCodeChallengeMethod = ptr.ToString(pkceCodeChallengeMethod)
	}
	return client
}

func toKcFilters(filters *models.ClientFilter) *gocloak.GetClientsParams {
	filtersChain := ""
	kcFilter := gocloak.GetClientsParams{
		Search: gocloak.BoolP(true),
	}
	if filters == nil {
		kcFilter.Search = gocloak.BoolP(false)
		return &kcFilter
	}

	if filters.ClientID != nil {
		kcFilter.ClientID = filters.ClientID
	}
	if filters.MaxResults != nil {
		kcFilter.Max = filters.MaxResults
	}
	if filters.IsExternal != nil {
		addKcFilter(&filtersChain, utils.IsExternal, strconv.FormatBool(*filters.IsExternal))
	}
	if filters.ClientType != nil {
		addKcFilter(&filtersChain, utils.ClientTypeAttr, *filters.ClientType)
	}
	if filters.ClusterId != nil {
		addKcFilter(&filtersChain, utils.ClusterIdAttr, *filters.ClusterId)
	}
	if filters.ClientType != nil {
		switch *filters.ClientType {
		case string(models.ClientTypeUserApp):
			if filters.IsTenantAdmin {
				if filters.CreatedBy != nil && *filters.CreatedBy != "" {
					addKcFilter(&filtersChain, utils.CreatedBy, *filters.CreatedBy)
					addKcFilter(&filtersChain, utils.ClientTypeAttr, string(models.ClientTypeUserApp))
				} else {
					addKcFilter(&filtersChain, utils.ClientTypeAttr, string(models.ClientTypeUserApp))
				}
			} else {
				if filters.CreatedBy != nil && *filters.CreatedBy != "" {
					addKcFilter(&filtersChain, utils.CreatedBy, *filters.CreatedBy)
					addKcFilter(&filtersChain, utils.ClientTypeAttr, string(models.ClientTypeUserApp))
				}
			}
		case string(models.ClientTypeTenantApp):
			addKcFilter(&filtersChain, utils.IsExternal, "true")
		default:
		}
	}
	kcFilter.SearchableAttributes = gocloak.StringP(filtersChain)
	return &kcFilter
}

func addKcFilter(kcFilterChain *string, filterKey, filterValue string) {
	*kcFilterChain = fmt.Sprintf("%s %s:%s", *kcFilterChain, filterKey, filterValue)
}
