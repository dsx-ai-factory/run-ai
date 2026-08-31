/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils/pointer"
)

const (
	ClientTypeUserApp       = "user-application"
	ClientTypeTenantApp     = "tenant-application"
	ClientTypeWorkloadsAuth = "WorkloadsAuth"
	ClientTypeMCP           = "MCP"
)

type ClientConfig struct {
	PublicClient              *bool `json:"publicClient,omitempty"`
	DirectAccessGrantsEnabled *bool `json:"directAccessGrantsEnabled,omitempty"`
	StandardFlowEnabled       *bool `json:"standardFlowEnabled,omitempty"`
	ServiceAccountsEnabled    *bool `json:"serviceAccountsEnabled,omitempty"`
}

type ClientAttributes struct {
	IsExternal              *bool   `json:"runai.is_external,omitempty"`
	ClientType              *string `json:"runai.client_type,omitempty"`
	UpdatedAt               *string `json:"runai.update.time,omitempty"`
	CreatedAt               *string `json:"runai.creation.time,omitempty"`
	CreatedBy               *string `json:"runai.created.by,omitempty"`
	LastLogin               *string `json:"runai.login.time,omitempty"`
	TokenLifespan           *time.Duration
	ClusterId               *string           `json:"runai.cluster_id,omitempty"`
	Oauth2AuthGrantEnabled  *string           `json:"-"`
	OidcGrantEnabled        *string           `json:"-"`
	LogoutRedirectUris      *string           `json:"-"`
	PkceCodeChallengeMethod *string           `json:"-"`
	Extra                   map[string]string `json:"-"`
}

func (c ClientAttributes) MarshalJSON() ([]byte, error) {
	type Alias struct {
		IsExternal    *string `json:"runai.is_external,omitempty"`
		ClientType    *string `json:"runai.client_type,omitempty"`
		UpdatedAt     *string `json:"runai.update.time,omitempty"`
		CreatedAt     *string `json:"runai.creation.time,omitempty"`
		CreatedBy     *string `json:"runai.created.by,omitempty"`
		LastLogin     *string `json:"runai.login.time,omitempty"`
		TokenLifespan *string `json:"access.token.lifespan,omitempty"`
		ClusterId     *string `json:"runai.cluster_id,omitempty"`
	}
	a := Alias{
		ClientType: c.ClientType,
		UpdatedAt:  c.UpdatedAt,
		CreatedAt:  c.CreatedAt,
		CreatedBy:  c.CreatedBy,
		LastLogin:  c.LastLogin,
		ClusterId:  c.ClusterId,
	}
	if c.IsExternal != nil {
		s := strconv.FormatBool(*c.IsExternal)
		a.IsExternal = &s
	}
	if c.TokenLifespan != nil {
		s := strconv.FormatInt(int64(c.TokenLifespan.Seconds()), 10)
		a.TokenLifespan = &s
	}
	return json.Marshal(a)
}

type Client struct {
	ID           *string           `json:"id,omitempty"`
	Name         *string           `json:"name,omitempty"`
	ClientID     *string           `json:"clientId,omitempty"`
	Enabled      *bool             `json:"enabled,omitempty"`
	Attributes   *ClientAttributes `json:"attributes,omitempty"`
	RedirectURIs *[]string         `json:"redirectUris,omitempty"`
	WebOrigins   *[]string         `json:"webOrigins,omitempty"`
	Secret       *string           `json:"secret,omitempty"`
	TenantId     *string           `json:"tenantId,omitempty"`
	ClientConfig *ClientConfig     `json:"clientConfig,omitempty"`
}

type ClientFilter struct {
	ClientID      *string
	MaxResults    *int
	IsExternal    *bool
	ClientType    *string
	ClusterId     *string
	CreatedBy     *string
	IsTenantAdmin bool
}

type NewClientParams struct {
	Name         *string
	TenantID     *string
	ClientID     *string
	RedirectURIs *[]string
	WebOrigins   *[]string
	Attributes   *ClientAttributes
}

func NewClient(params NewClientParams) *Client {
	if params.Name == nil {
		params.Name = params.ClientID
	}
	if params.ClientID == nil {
		params.ClientID = params.Name
	}
	if params.Attributes == nil {
		params.Attributes = &ClientAttributes{}
	}
	return &Client{
		Name:         params.Name,
		ClientID:     params.ClientID,
		TenantId:     params.TenantID,
		Enabled:      pointer.ToBool(true),
		RedirectURIs: params.RedirectURIs,
		WebOrigins:   params.WebOrigins,
		ClientConfig: &ClientConfig{},
		Attributes:   params.Attributes,
	}
}

func (c *Client) IsServiceAccountsEnabled() bool {
	if c.ClientConfig == nil || c.ClientConfig.ServiceAccountsEnabled == nil {
		return false
	}
	return *c.ClientConfig.ServiceAccountsEnabled
}

func (c *Client) GetClientType() string {
	if c.Attributes.ClientType == nil {
		return ""
	}
	return *c.Attributes.ClientType
}

func (c *Client) IsClientAuthenticationEnabled() bool {
	return c.ClientConfig == nil || c.ClientConfig.PublicClient == nil || !*c.ClientConfig.PublicClient
}

func (c *Client) SetCreatedBy(createdBy *string) {
	c.Attributes.CreatedBy = createdBy
}

func (c *Client) SetCreatedAtNow() {
	c.Attributes.CreatedAt = pointer.ToString(strconv.FormatInt(time.Now().Unix(), 10))
}

func (c *Client) SetUpdatedAtNow() {
	c.Attributes.UpdatedAt = pointer.ToString(strconv.FormatInt(time.Now().Unix(), 10))
}
