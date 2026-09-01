/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/gateway-hub/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	os.Setenv("SERVICE_NAME", "gateway-hub-test")
	os.Setenv("SECRETS_MOUNT_PATH", ".")
	os.Exit(m.Run())
}

type testKeyPair struct {
	priv    *rsa.PrivateKey
	certB64 string
}

func generateKeyPair(t *testing.T) *testKeyPair {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "generate RSA key")

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err, "create certificate")
	return &testKeyPair{
		priv:    priv,
		certB64: base64.StdEncoding.EncodeToString(certDER),
	}
}

func (kp *testKeyPair) signRS256(claims map[string]any) (string, error) {
	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload

	h := sha256.New()
	h.Write([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, kp.priv, crypto.SHA256, h.Sum(nil))
	if err != nil {
		return "", err
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func keycloakServer(t *testing.T, realm string, kp *testKeyPair) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/realms/" + realm
		if r.URL.Path != expected {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp, _ := json.Marshal(map[string]string{
			"realm":      realm,
			"public_key": kp.certB64,
		})
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
}

func makeClaims(iss, sub string, tenantID int, ttl time.Duration, extra map[string]any) map[string]any {
	now := time.Now().UTC()
	c := map[string]any{
		"iss":       iss,
		"sub":       sub,
		"tenant_id": tenantID,
		"exp":       now.Add(ttl).Unix(),
		"iat":       now.Unix(),
	}
	for k, v := range extra {
		c[k] = v
	}
	return c
}

type authFlowCase struct {
	name string

	claims map[string]any

	sub      string
	tenantID int
	ttl      time.Duration
	extra    map[string]any

	managementRealm string
	bearerPrefix    string
	wrongSigningKey bool

	wantStatus        int
	wantHeaders       map[string]string
	wantMultiHeaders  map[string][]string
	wantAbsentHeaders []string
	wantErrSubstr     string
}

func TestAuthHandler_Flow(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()
	issuer := srv.URL + "/realms/runai"

	cases := []authFlowCase{
		{
			name: "valid jwt", sub: "alice@run.ai", tenantID: 42, ttl: time.Hour,
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{auth.HeaderTenantId: "42", auth.HeaderSubjectType: "User"},
		},
		{
			name: "service account", sub: "app-abc123", tenantID: 1, ttl: time.Hour,
			extra:       map[string]any{"client_id": "app-abc123"},
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{auth.HeaderSubjectType: "ServiceAccount"},
		},
		{
			name: "personal app", sub: "app-abc123", tenantID: 1, ttl: time.Hour,
			extra:       map[string]any{"client_id": "app-abc123", "client_type": "personal-application"},
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{auth.HeaderClientId: "app-abc123"},
		},
		{
			name: "client id empty for non personal app", sub: "app-abc123", tenantID: 1, ttl: time.Hour,
			extra:       map[string]any{"client_id": "app-abc123"},
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{auth.HeaderClientId: ""},
		},
		{
			name: "supplemental groups", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			extra:       map[string]any{"supplementarygroups": []int{100, 200}},
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{auth.HeaderSupplementalGroups: "100,200"},
		},
		{
			name: "groups present", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			extra:            map[string]any{"groups": []string{"eng", "ops"}},
			wantStatus:       http.StatusOK,
			wantMultiHeaders: map[string][]string{auth.HeaderGroups: {"eng", "ops"}},
		},
		{
			name: "groups absent - header not set", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			wantStatus:        http.StatusOK,
			wantAbsentHeaders: []string{auth.HeaderGroups},
		},
		{
			name: "group name containing comma is escaped, not split", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			extra:            map[string]any{"groups": []string{"x,runai-admins"}},
			wantStatus:       http.StatusOK,
			wantMultiHeaders: map[string][]string{auth.HeaderGroups: {"x%2Crunai-admins"}},
		},
		{
			name: "gid uid session headers", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			extra:       map[string]any{"gid": "1000", "uid": "1000", "sid": "sess-123"},
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{auth.HeaderGID: "1000", auth.HeaderUID: "1000", auth.HeaderSessionId: "sess-123"},
		},
		{
			name: "all headers set", sub: "app-abc123", tenantID: 42, ttl: time.Hour,
			extra: map[string]any{
				"client_id": "app-abc123", "client_type": "personal-application",
				"groups": []string{"eng", "ops"}, "supplementarygroups": []int{100, 200},
				"gid": "1000", "uid": "2000", "sid": "session-1",
			},
			managementRealm: "runai",
			wantStatus:      http.StatusOK,
			wantHeaders: map[string]string{
				auth.HeaderTenantId: "42", auth.HeaderSubjectId: "app-abc123", auth.HeaderSubjectType: "ServiceAccount",
				auth.HeaderSupplementalGroups: "100,200",
				auth.HeaderGID: "1000", auth.HeaderUID: "2000", auth.HeaderIsManagementTenant: "true",
				auth.HeaderClientId: "app-abc123", auth.HeaderSessionId: "session-1",
			},
			wantMultiHeaders: map[string][]string{auth.HeaderGroups: {"eng", "ops"}},
		},
		{
			name: "tenant id as string",
			claims: map[string]any{
				"iss": issuer, "sub": "alice@run.ai", "tenant_id": "99",
				"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
			},
			wantStatus:  http.StatusOK,
			wantHeaders: map[string]string{auth.HeaderTenantId: "99"},
		},
		{
			name: "lowercase bearer prefix", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			bearerPrefix: "bearer ",
			wantStatus:   http.StatusOK,
		},
		{
			name: "is management realm", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			managementRealm: "runai",
			wantStatus:      http.StatusOK,
			wantHeaders:     map[string]string{auth.HeaderIsManagementTenant: "true"},
		},
		{
			name: "not management realm", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			managementRealm: "other-realm",
			wantStatus:      http.StatusOK,
			wantHeaders:     map[string]string{auth.HeaderIsManagementTenant: "false"},
		},
		{
			name: "expired token", sub: "alice@run.ai", tenantID: 1, ttl: -10 * time.Second,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "not yet valid token", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			extra:         map[string]any{"nbf": time.Now().Add(time.Hour).Unix()},
			wantStatus:    http.StatusUnauthorized,
			wantErrSubstr: "token is not valid yet",
		},
		{
			name: "token used before issued", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			extra:         map[string]any{"iat": time.Now().Add(time.Hour).Unix()},
			wantStatus:    http.StatusUnauthorized,
			wantErrSubstr: "token used before issued",
		},
		{
			name: "missing tenant id",
			claims: map[string]any{
				"iss": issuer, "sub": "alice@run.ai",
				"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
			},
			wantStatus:    http.StatusUnauthorized,
			wantErrSubstr: "missing claim",
		},
		{
			name: "missing subject",
			claims: map[string]any{
				"iss": issuer, "tenant_id": 1,
				"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(),
			},
			wantStatus:    http.StatusUnauthorized,
			wantErrSubstr: "missing claim",
		},
		{
			name: "invalid signature", sub: "alice@run.ai", tenantID: 1, ttl: time.Hour,
			wrongSigningKey: true,
			wantStatus:      http.StatusUnauthorized,
			wantErrSubstr:   "signature is invalid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			claims := tc.claims
			if claims == nil {
				claims = makeClaims(issuer, tc.sub, tc.tenantID, tc.ttl, tc.extra)
			}

			signer := kp
			if tc.wrongSigningKey {
				signer = generateKeyPair(t)
			}
			token, err := signer.signRS256(claims)
			require.NoError(t, err, "sign")

			bearerPrefix := tc.bearerPrefix
			if bearerPrefix == "" {
				bearerPrefix = "Bearer "
			}

			h := auth.NewHandler(srv.URL, tc.managementRealm)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", bearerPrefix+token)
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			require.Equal(t, tc.wantStatus, rw.Code, "status mismatch: %s", rw.Body.String())
			for header, want := range tc.wantHeaders {
				assertHeader(t, rw, header, want)
			}
			for header, want := range tc.wantMultiHeaders {
				assert.Equal(t, want, rw.Header().Values(header), "header %s", header)
			}
			for _, header := range tc.wantAbsentHeaders {
				assert.Empty(t, rw.Header().Get(header), "header %s should be absent", header)
			}
			if tc.wantErrSubstr != "" {
				assertUnauthorized(t, rw, tc.wantErrSubstr)
			}
		})
	}
}

func TestAuthHandler_AuthorizationHeaderFormats(t *testing.T) {
	cases := []struct {
		name       string
		authHeader string
	}{
		{name: "missing auth header", authHeader: ""},
		{name: "malformed authorization header", authHeader: "Token abc123"},
		{name: "empty bearer token", authHeader: "Bearer "},
		{name: "oversized bearer token", authHeader: "Bearer " + strings.Repeat("a", 16*1024+1)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := auth.NewHandler("http://keycloak", "")
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)
			assertUnauthorized(t, rw, "authorization header is missing or invalid")
		})
	}
}

func TestAuthHandler_RealKeycloakPublicKeyFormat(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "generate RSA key")
	pkixDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err, "marshal PKIX public key")
	kp := &testKeyPair{priv: priv, certB64: base64.StdEncoding.EncodeToString(pkixDER)}

	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()

	issuer := srv.URL + "/realms/runai"
	claims := makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil)
	token, err := kp.signRS256(claims)
	require.NoError(t, err, "sign")

	h := auth.NewHandler(srv.URL, "")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	require.Equal(t, http.StatusOK, rw.Code, "expected 200: %s", rw.Body.String())
}

func TestAuthHandler_InvalidIssuer(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()

	for _, badIssuer := range []string{"", "notaurl", "onlyone/", "https://kc/auth/invalid"} {
		claims := makeClaims(badIssuer, "alice@run.ai", 1, time.Hour, nil)
		token, _ := kp.signRS256(claims)
		h := auth.NewHandler(srv.URL, "")
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		assertUnauthorized(t, rw, "token has invalid issuer")
	}
}

func TestAuthHandler_KeycloakUnreachable(t *testing.T) {
	kp := generateKeyPair(t)
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	dead.Close()

	issuer := dead.URL + "/realms/runai"
	claims := makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil)
	token, _ := kp.signRS256(claims)
	h := auth.NewHandler(dead.URL, "")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "failed to get public key")
}

func TestAuthHandler_PublicKeyCached(t *testing.T) {
	kp := generateKeyPair(t)
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp, _ := json.Marshal(map[string]string{
			"realm":      "runai",
			"public_key": kp.certB64,
		})
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer srv.Close()

	issuer := srv.URL + "/realms/runai"
	h := auth.NewHandler(srv.URL, "")

	for i := 0; i < 3; i++ {
		claims := makeClaims(issuer, fmt.Sprintf("user%d@run.ai", i), 1, time.Hour, nil)
		token, _ := kp.signRS256(claims)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		require.Equal(t, http.StatusOK, rw.Code, "request %d: expected 200", i)
	}

	assert.Equal(t, 1, callCount, "Keycloak should be called once (cache miss)")
}

func TestAuthHandler_MalformedJWTStructure(t *testing.T) {
	h := auth.NewHandler("http://keycloak", "")

	t.Run("wrong number of segments", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer abc.def")
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		assertUnauthorized(t, rw, "token could not be parsed")
	})

	t.Run("invalid base64 payload", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer aGVhZGVy.!!!not-base64!!!.sig")
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		assertUnauthorized(t, rw, "token could not be parsed")
	})
}

func TestAuthHandler_UnsupportedAlgorithm(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()

	issuer := srv.URL + "/realms/runai"
	h := auth.NewHandler(srv.URL, "")

	headerJSON, _ := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	payloadJSON, _ := json.Marshal(makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil))
	fakeToken := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON) + ".invalidsig"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "signature is invalid")
}

func TestAuthHandler_UnknownRealm(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()

	issuer := srv.URL + "/realms/does-not-exist"
	claims := makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil)
	token, _ := kp.signRS256(claims)
	h := auth.NewHandler(srv.URL, "")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "failed to get public key")
}

func TestAuthHandler_KeycloakMalformedResponse(t *testing.T) {
	kp := generateKeyPair(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	issuer := srv.URL + "/realms/runai"
	claims := makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil)
	token, _ := kp.signRS256(claims)
	h := auth.NewHandler(srv.URL, "")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "failed to get public key")
}

func TestAuthHandler_KeycloakBadPublicKey(t *testing.T) {
	kp := generateKeyPair(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, _ := json.Marshal(map[string]string{
			"realm":      "runai",
			"public_key": "not-a-valid-cert",
		})
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer srv.Close()

	issuer := srv.URL + "/realms/runai"
	claims := makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil)
	token, _ := kp.signRS256(claims)
	h := auth.NewHandler(srv.URL, "")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "failed to get public key")
}

func TestAuthHandler_KeycloakRetriesOnEachRequestWhileUnavailable(t *testing.T) {
	kp := generateKeyPair(t)
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	issuer := srv.URL + "/realms/runai"
	h := auth.NewHandler(srv.URL, "")

	for i := 0; i < 3; i++ {
		claims := makeClaims(issuer, fmt.Sprintf("user%d@run.ai", i), 1, time.Hour, nil)
		token, _ := kp.signRS256(claims)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rw := httptest.NewRecorder()
		h.ServeHTTP(rw, req)
		require.Equal(t, http.StatusUnauthorized, rw.Code, "request %d: expected 401", i)
	}

	assert.Equal(t, 3, callCount, "Keycloak should be called on every request while the public key is not cached")
}

func TestAuthHandler_KeycloakRecoversAfterTransientFailure(t *testing.T) {
	kp := generateKeyPair(t)
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp, _ := json.Marshal(map[string]string{
			"realm":      "runai",
			"public_key": kp.certB64,
		})
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer srv.Close()

	issuer := srv.URL + "/realms/runai"
	h := auth.NewHandler(srv.URL, "")

	claims := makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil)
	token, err := kp.signRS256(claims)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	require.Equal(t, http.StatusUnauthorized, rw.Code, "first request should fail while Keycloak is unhealthy")

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	require.Equal(t, http.StatusOK, rw.Code, "second request should recover immediately: %s", rw.Body.String())
	assert.Equal(t, 2, callCount, "Keycloak should be retried on the next request")
}

func assertHeader(t *testing.T, rw *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	assert.Equal(t, want, rw.Header().Get(key), "header %s", key)
}

func assertUnauthorized(t *testing.T, rw *httptest.ResponseRecorder, wantMsg string) {
	t.Helper()
	require.Equal(t, http.StatusUnauthorized, rw.Code)
	if wantMsg == "" {
		return
	}
	var got struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &got), "unmarshal error body (raw: %s)", rw.Body.String())
	assert.Equal(t, http.StatusUnauthorized, got.Code, "error body code")
	assert.Contains(t, got.Message, wantMsg, "error body message")
}

func TestAuthHandler_BadPayloadBase64(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()
	h := auth.NewHandler(srv.URL, "")

	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	fakeToken := base64.RawURLEncoding.EncodeToString(headerJSON) + ".!!!.validsig"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "token could not be parsed")
}

func TestAuthHandler_BadHeaderBase64(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()
	h := auth.NewHandler(srv.URL, "")
	issuer := srv.URL + "/realms/runai"

	payloadJSON, _ := json.Marshal(makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil))
	fakeToken := "!!!." + base64.RawURLEncoding.EncodeToString(payloadJSON) + ".validsig"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "signature is invalid")
}

func TestAuthHandler_NonJSONHeader(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()
	h := auth.NewHandler(srv.URL, "")
	issuer := srv.URL + "/realms/runai"

	payloadJSON, _ := json.Marshal(makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil))
	fakeToken := base64.RawURLEncoding.EncodeToString([]byte("not-json")) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON) + ".validsig"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "signature is invalid")
}

func TestAuthHandler_BadSignatureBase64(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()
	h := auth.NewHandler(srv.URL, "")
	issuer := srv.URL + "/realms/runai"

	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	payloadJSON, _ := json.Marshal(makeClaims(issuer, "alice@run.ai", 1, time.Hour, nil))
	fakeToken := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString(payloadJSON) + ".!!!"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "signature is invalid")
}

func TestAuthHandler_TenantIdNonNumericString(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()
	h := auth.NewHandler(srv.URL, "")
	issuer := srv.URL + "/realms/runai"

	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	rawPayload := fmt.Sprintf(`{"iss":%q,"sub":"alice@run.ai","tenant_id":"abc"}`, issuer)
	fakeToken := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(rawPayload)) + ".validsig"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "token could not be parsed")
}

func TestAuthHandler_TenantIdObjectJSON(t *testing.T) {
	kp := generateKeyPair(t)
	srv := keycloakServer(t, "runai", kp)
	defer srv.Close()
	h := auth.NewHandler(srv.URL, "")
	issuer := srv.URL + "/realms/runai"

	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	rawPayload := fmt.Sprintf(`{"iss":%q,"sub":"alice@run.ai","tenant_id":{}}`, issuer)
	fakeToken := base64.RawURLEncoding.EncodeToString(headerJSON) + "." +
		base64.RawURLEncoding.EncodeToString([]byte(rawPayload)) + ".validsig"

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+fakeToken)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assertUnauthorized(t, rw, "token could not be parsed")
}
