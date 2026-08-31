/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils/pointer"
)

var (
	now              = time.Now()
	clientID         = pointer.ToString("testClientID")
	clientName       = pointer.ToString("testClientName")
	tenantID         = pointer.ToString("testTenantID")
	redirectUris     = &[]string{"testRedirectURI"}
	webOrigins       = &[]string{"testWebOrigin"}
	clientAttributes = &ClientAttributes{
		ClientType: pointer.ToString("testClientType"),
		IsExternal: pointer.ToBool(true),
		UpdatedAt:  pointer.ToString(now.Format(time.RFC3339)),
		CreatedAt:  pointer.ToString(now.Format(time.RFC3339)),
		CreatedBy:  pointer.ToString("testCreatedBy"),
		LastLogin:  pointer.ToString(now.Format(time.RFC3339)),
	}
)

func TestNewClient(t *testing.T) {
	type args struct {
		name             *string
		tenantID         string
		clientID         *string
		redirectUris     *[]string
		webOrigins       *[]string
		clientAttributes *ClientAttributes
	}
	tests := []struct {
		name string
		args args
		want *Client
	}{
		{
			name: "Test NewClients",
			args: args{
				name:             clientName,
				tenantID:         *tenantID,
				clientID:         clientID,
				redirectUris:     redirectUris,
				webOrigins:       webOrigins,
				clientAttributes: clientAttributes,
			},
			want: &Client{
				Name:         clientName,
				TenantId:     tenantID,
				ClientID:     clientID,
				RedirectURIs: redirectUris,
				WebOrigins:   webOrigins,
				Enabled:      pointer.ToBool(true),
				ClientConfig: &ClientConfig{},
				Attributes:   clientAttributes,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NewClient(NewClientParams{
				Name:         tt.args.name,
				TenantID:     &tt.args.tenantID,
				ClientID:     tt.args.clientID,
				RedirectURIs: tt.args.redirectUris,
				WebOrigins:   tt.args.webOrigins,
				Attributes:   tt.args.clientAttributes,
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_IsServiceAccountsEnabled(t *testing.T) {
	tests := []struct {
		name   string
		client *Client
		want   bool
	}{
		{
			name: "Service accounts enabled",
			client: &Client{
				ClientConfig: &ClientConfig{
					ServiceAccountsEnabled: pointer.ToBool(true),
				},
			},
			want: true,
		},
		{
			name: "Service accounts disabled",
			client: &Client{
				ClientConfig: &ClientConfig{
					ServiceAccountsEnabled: pointer.ToBool(false),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.client.IsServiceAccountsEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_GetClientType(t *testing.T) {
	tests := []struct {
		name   string
		client *Client
		want   string
	}{
		{
			name: "Client type set",
			client: &Client{
				Attributes: &ClientAttributes{
					ClientType: pointer.ToString("test-type"),
				},
			},
			want: "test-type",
		},
		{
			name: "Client type nil",
			client: &Client{
				Attributes: &ClientAttributes{
					ClientType: nil,
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.client.GetClientType()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClient_SetCreatedBy(t *testing.T) {
	t.Parallel()

	client := &Client{
		Attributes: &ClientAttributes{},
	}

	client.SetCreatedBy(pointer.ToString("test-created-by"))
	assert.Equal(t, "test-created-by", *client.Attributes.CreatedBy)
}

func TestClient_SetCreatedAtNow(t *testing.T) {
	t.Parallel()

	client := &Client{
		Attributes: &ClientAttributes{},
	}

	client.SetCreatedAtNow()
	assert.NotNil(t, client.Attributes.CreatedAt)
	_, err := strconv.ParseInt(*client.Attributes.CreatedAt, 10, 64)
	assert.NoError(t, err)
}

func TestClient_SetUpdatedAtNow(t *testing.T) {
	t.Parallel()

	client := &Client{
		Attributes: &ClientAttributes{},
	}

	client.SetUpdatedAtNow()
	assert.NotNil(t, client.Attributes.UpdatedAt)
	_, err := strconv.ParseInt(*client.Attributes.UpdatedAt, 10, 64)
	assert.NoError(t, err)
}

func TestClientAttributes_MarshalJSON_IncludesTokenLifespan(t *testing.T) {
	t.Parallel()

	tokenLifespan := 10 * time.Minute
	attrs := ClientAttributes{
		ClientType:    pointer.ToString("RunaiAdminUI"),
		TokenLifespan: &tokenLifespan,
	}

	data, err := json.Marshal(attrs)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"access.token.lifespan":"600"`)
}
