/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

func TestCreateApp_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := CreateAppResponse{
			Ok:    true,
			AppId: "A12345",
			Credentials: CreateAppCredentials{
				ClientId:     "test-client-id",
				ClientSecret: "test-client-secret",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	client := &Client{
		httpClient:        resty.New(),
		createManifestUrl: mockServer.URL,
	}
	response, err := client.CreateApp("xox")
	require.NoError(t, err)
	require.Equal(t, "A12345", response.AppId)
	require.Equal(t, "test-client-id", response.ClientId)
	require.Equal(t, "test-client-secret", response.ClientSecret)
}

func TestCreateApp_Failure_InvalidAuth(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockResponse := CreateAppResponse{
			Ok:    false,
			Error: "invalid_auth",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	client := &Client{
		httpClient:        resty.New(),
		createManifestUrl: mockServer.URL,
	}

	response, err := client.CreateApp("xox")
	require.Error(t, err)
	require.Nil(t, response)
}

func TestExchangeCode_Success(t *testing.T) {
	clientId, clientSecret, code, token := "clientId", "clientSecret", "code", "token"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)
		require.Equal(t, clientId, r.FormValue("client_id"))
		require.Equal(t, clientSecret, r.FormValue("client_secret"))
		require.Equal(t, code, r.FormValue("code"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(oauthAccessResponse{
			AccessToken: token,
			Team: oauthAccessResponseTeam{
				Name: "Test Team",
			},
			Ok: true,
		})
	}))
	defer ts.Close()

	slackClient := &Client{
		oauthAccessUrl: ts.URL,
		httpClient:     resty.New(),
	}
	exchangeRes, err := slackClient.ExchangeCode(code, clientId, clientSecret)
	require.NoError(t, err)
	require.Equal(t, token, exchangeRes.AccessToken)
	require.Equal(t, "Test Team", exchangeRes.TeamName)
}

func TestExchangeCode_Failed(t *testing.T) {
	clientId, clientSecret, code := "clientId", "clientSecret", "code"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		er := r.ParseForm()
		require.NoError(t, er)
		require.Equal(t, clientId, r.FormValue("client_id"))
		require.Equal(t, clientSecret, r.FormValue("client_secret"))
		require.Equal(t, code, r.FormValue("code"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(oauthAccessResponse{
			Ok: false,
		})
	}))
	defer ts.Close()

	slackClient := &Client{
		oauthAccessUrl: ts.URL,
		httpClient:     resty.New(),
	}
	exchangeRes, er := slackClient.ExchangeCode(code, clientId, clientSecret)
	require.Error(t, er)
	require.ErrorContains(t, er, "failed to exchange slack code")
	require.Empty(t, exchangeRes)
}

func TestGetAuthUrl(t *testing.T) {
	clientId := "clientId"

	slackClient := &Client{
		redirectUrl: "https://example.com/callback",
	}

	uri := slackClient.GetAuthUrl("state123", clientId)

	parsed, err := url.Parse(uri)
	require.NoError(t, err)

	query := parsed.Query()
	require.Equal(t, clientId, query.Get("client_id"))
	require.Equal(t, slackClient.redirectUrl, query.Get("redirect_uri"))
	require.Equal(t, scopesForQuery, query.Get("scope"))
	require.Equal(t, "state123", query.Get("state"))
}
