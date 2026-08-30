package mcphttp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/runai-mcp-server/internal/auth"
)

type testOIDC struct {
	verifier *auth.JWKSVerifier
	key      *rsa.PrivateKey
	issuer   string
}

func newTestOIDC(t *testing.T) *testOIDC {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	oidcMux := http.NewServeMux()
	oidcServer := httptest.NewServer(oidcMux)
	t.Cleanup(oidcServer.Close)

	issuer := oidcServer.URL

	oidcMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":   issuer,
			"jwks_uri": issuer + "/jwks",
		})
	})
	oidcMux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(rsaPublicKeyToJWKS(&key.PublicKey))
	})

	verifier, err := auth.NewJWKSVerifier(context.Background(), auth.JWKSVerifierConfig{
		IssuerURL:  issuer,
		HTTPClient: oidcServer.Client(),
		Logger:     zerolog.Nop(),
	})
	require.NoError(t, err)

	// Pre-warm the JWKS key cache. keyfunc fetches keys lazily on first use;
	// without this, the first JWT validation may return 400 instead of 401.
	o := &testOIDC{verifier: verifier, key: key, issuer: issuer}
	_, _ = verifier.Verify(context.Background(), o.signToken(t, time.Now().Add(time.Hour)), nil)
	return o
}

func (o *testOIDC) validToken(t *testing.T) string {
	t.Helper()
	return o.signToken(t, time.Now().Add(time.Hour))
}

func (o *testOIDC) expiredToken(t *testing.T) string {
	t.Helper()
	return o.signToken(t, time.Now().Add(-time.Hour))
}

func (o *testOIDC) signToken(t *testing.T, expiry time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"iss": o.issuer,
		"sub": "test-user",
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = "test-key"
	s, err := tok.SignedString(o.key)
	require.NoError(t, err)
	return s
}

func newTestServer(t *testing.T, verifier *auth.JWKSVerifier, getServer func(context.Context) (*mcp.Server, error)) *httptest.Server {
	t.Helper()
	cfg := Config{ListenPort: 9999, Verifier: verifier}
	r := newRouter(getServer, cfg, zerolog.Nop())
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

type testEnv struct {
	*testOIDC
	server *httptest.Server
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	oidc := newTestOIDC(t)
	getServer := func(_ context.Context) (*mcp.Server, error) {
		return mcp.NewServer(&mcp.Implementation{Name: "test"}, nil), nil
	}
	return &testEnv{testOIDC: oidc, server: newTestServer(t, oidc.verifier, getServer)}
}

func rsaPublicKeyToJWKS(pub *rsa.PublicKey) []byte {
	eBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(eBytes, uint32(pub.E))
	for len(eBytes) > 1 && eBytes[0] == 0 {
		eBytes = eBytes[1:]
	}
	doc := map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"alg": "RS256",
			"use": "sig",
			"kid": "test-key",
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(eBytes),
		}},
	}
	b, _ := json.Marshal(doc)
	return b
}

func TestHealthz(t *testing.T) {
	env := newTestEnv(t)
	resp, err := http.Get(env.server.URL + "/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestProtectedResourceMetadata(t *testing.T) {
	env := newTestEnv(t)
	resp, err := http.Get(env.server.URL + wellKnownPath)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
}

func TestMCPNoToken(t *testing.T) {
	env := newTestEnv(t)
	resp, err := http.Post(env.server.URL+"/mcp", "application/json", strings.NewReader("{}"))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("WWW-Authenticate"), "Bearer resource_metadata=")
}

func TestMCPExpiredToken(t *testing.T) {
	env := newTestEnv(t)
	req, err := http.NewRequest(http.MethodPost, env.server.URL+"/mcp", strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+env.expiredToken(t))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestMCPTamperedToken(t *testing.T) {
	env := newTestEnv(t)
	token := env.validToken(t)
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	sig := []rune(parts[2])
	sig[len(sig)-1] = 'X'
	parts[2] = string(sig)
	tampered := strings.Join(parts, ".")

	req, err := http.NewRequest(http.MethodPost, env.server.URL+"/mcp", strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tampered)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	// 401 = invalid signature (key cached); 400 = key-cache miss (ErrOAuth from keyfunc).
	// Both mean the token was rejected.
	assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest)
}

func TestMCPValidTokenPassesAuth(t *testing.T) {
	env := newTestEnv(t)
	req, err := http.NewRequest(http.MethodPost, env.server.URL+"/mcp", strings.NewReader("{}"))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+env.validToken(t))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode)
	assert.NotEqual(t, http.StatusForbidden, resp.StatusCode)
}

func TestMCPRecoveryAfterInitFailure(t *testing.T) {
	oidc := newTestOIDC(t)

	var calls atomic.Int32
	getServer := func(_ context.Context) (*mcp.Server, error) {
		if calls.Add(1) == 1 {
			return nil, fmt.Errorf("transient init failure")
		}
		return mcp.NewServer(&mcp.Implementation{Name: "test"}, nil), nil
	}
	srv := newTestServer(t, oidc.verifier, getServer)

	do := func() *http.Response {
		req, err := http.NewRequest(http.MethodPost, srv.URL+"/mcp", strings.NewReader("{}"))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+oidc.validToken(t))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	resp1 := do()
	defer resp1.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp1.StatusCode)

	resp2 := do()
	defer resp2.Body.Close()
	assert.NotEqual(t, http.StatusServiceUnavailable, resp2.StatusCode)
	assert.NotEqual(t, http.StatusUnauthorized, resp2.StatusCode)
}
