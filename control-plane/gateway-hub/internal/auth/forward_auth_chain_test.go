/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/gateway-hub/internal/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	runaictxhttp "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var allRunaiHeaders = []string{
	auth.HeaderTenantId,
	auth.HeaderSubjectId,
	auth.HeaderSubjectType,
	auth.HeaderGroups,
	auth.HeaderSupplementalGroups,
	auth.HeaderGID,
	auth.HeaderUID,
	auth.HeaderIsManagementTenant,
	auth.HeaderClientId,
	auth.HeaderSessionId,
}

// applyForwardAuth simulates Traefik's forwardAuth: calls the handler in-process,
// strips all X-Runai-* headers from clientReq, copies the handler's response headers.
func applyForwardAuth(h *auth.Handler, clientReq *http.Request) (allowed bool, status int, upstream *http.Request) {
	authReq := httptest.NewRequest(http.MethodGet, "/", nil)
	authReq.Header.Set("Authorization", clientReq.Header.Get("Authorization"))
	authRec := httptest.NewRecorder()
	h.ServeHTTP(authRec, authReq)

	if authRec.Code < 200 || authRec.Code >= 300 {
		return false, authRec.Code, nil
	}

	upstream = clientReq.Clone(clientReq.Context())
	for _, header := range allRunaiHeaders {
		upstream.Header.Del(header)
		for _, v := range authRec.Header().Values(header) {
			upstream.Header.Add(header, v)
		}
	}
	return true, http.StatusOK, upstream
}

func TestForwardAuth_MandatoryIdentityHeaders(t *testing.T) {
	kp := generateKeyPair(t)
	keycloak := keycloakServer(t, "runai", kp)
	defer keycloak.Close()

	h := auth.NewHandler(keycloak.URL, "")
	issuer := keycloak.URL + "/realms/runai"
	token, _ := kp.signRS256(makeClaims(issuer, "alice@run.ai", 42, time.Hour, nil))

	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	allowed, _, upstream := applyForwardAuth(h, req)
	require.True(t, allowed, "expected request to be allowed")

	for _, tc := range []struct{ header, want string }{
		{auth.HeaderSubjectId, "alice@run.ai"},
		{auth.HeaderTenantId, "42"},
		{auth.HeaderSubjectType, "User"},
	} {
		assert.Equal(t, tc.want, upstream.Header.Get(tc.header), "upstream %s", tc.header)
	}
}

func TestForwardAuth_SEC1_ClientHeadersStripped(t *testing.T) {
	kp := generateKeyPair(t)
	keycloak := keycloakServer(t, "runai", kp)
	defer keycloak.Close()

	h := auth.NewHandler(keycloak.URL, "")
	issuer := keycloak.URL + "/realms/runai"
	token, _ := kp.signRS256(makeClaims(issuer, "alice@run.ai", 42, time.Hour, nil))

	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set(auth.HeaderSubjectId, "attacker-subject")
	req.Header.Set(auth.HeaderTenantId, "999")
	req.Header.Set(auth.HeaderSubjectType, "ServiceAccount")
	req.Header.Set(auth.HeaderIsManagementTenant, "true")

	allowed, _, upstream := applyForwardAuth(h, req)
	require.True(t, allowed, "expected request to be allowed")

	for _, tc := range []struct{ header, want string }{
		{auth.HeaderSubjectId, "alice@run.ai"},
		{auth.HeaderTenantId, "42"},
		{auth.HeaderSubjectType, "User"},
		{auth.HeaderIsManagementTenant, "false"},
	} {
		assert.Equal(t, tc.want, upstream.Header.Get(tc.header), "SEC-1: %s attacker value leaked", tc.header)
	}
}

func TestForwardAuth_SEC1_ClientGroupsHeaderStripped(t *testing.T) {
	kp := generateKeyPair(t)
	keycloak := keycloakServer(t, "runai", kp)
	defer keycloak.Close()

	h := auth.NewHandler(keycloak.URL, "")
	issuer := keycloak.URL + "/realms/runai"
	token, _ := kp.signRS256(makeClaims(issuer, "alice@run.ai", 42, time.Hour, map[string]any{"groups": []string{"eng"}}))

	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Add(auth.HeaderGroups, "attacker-group")
	req.Header.Add(auth.HeaderGroups, "runai-admins")

	allowed, _, upstream := applyForwardAuth(h, req)
	require.True(t, allowed, "expected request to be allowed")

	assert.Equal(t, []string{"eng"}, upstream.Header.Values(auth.HeaderGroups), "SEC-1: attacker-supplied groups leaked")
}

func TestForwardAuth_GroupNameContainingComma_RoundTripsThroughRealMiddleware(t *testing.T) {
	kp := generateKeyPair(t)
	keycloak := keycloakServer(t, "runai", kp)
	defer keycloak.Close()

	h := auth.NewHandler(keycloak.URL, "")
	authServer := httptest.NewServer(h)
	defer authServer.Close()

	issuer := keycloak.URL + "/realms/runai"
	token, _ := kp.signRS256(makeClaims(issuer, "alice@run.ai", 42, time.Hour, map[string]any{"groups": []string{"x,runai-admins"}}))

	authReq, err := http.NewRequest(http.MethodGet, authServer.URL, nil)
	require.NoError(t, err)
	authReq.Header.Set("Authorization", "Bearer "+token)

	authResp, err := http.DefaultClient.Do(authReq)
	require.NoError(t, err)
	defer authResp.Body.Close()
	require.Equal(t, http.StatusOK, authResp.StatusCode)

	upstream := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	for _, header := range allRunaiHeaders {
		upstream.Header.Del(header)
		for _, v := range authResp.Header.Values(header) {
			upstream.Header.Add(header, v)
		}
	}

	var gotGroups []string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotGroups, _ = r.Context().Value(runaictx.CfGroups).([]string)
	})
	runaictxhttp.ContextMiddleware(next).ServeHTTP(httptest.NewRecorder(), upstream)

	assert.Equal(t, []string{"x%2Crunai-admins"}, gotGroups, "comma in group name must not be split into extra groups")
}

func TestForwardAuth_Unauthenticated_Returns401(t *testing.T) {
	h := auth.NewHandler("http://keycloak-unreachable", "")

	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)

	allowed, status, upstream := applyForwardAuth(h, req)
	assert.False(t, allowed, "expected request to be rejected")
	assert.Nil(t, upstream, "upstream request must not be created for unauthenticated input")
	assert.Equal(t, http.StatusUnauthorized, status)
}
