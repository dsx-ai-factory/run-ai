/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

type fakeIDM struct {
	t      *testing.T
	logins int
	tokens []string
}

func (f *fakeIDM) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != loginPath {
		f.t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
	}
	var request tokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		f.t.Errorf("decoding login body: %v", err)
	}
	if request.GrantType != "password" || request.Username == "" || request.Password == "" {
		f.t.Errorf("unexpected login body: %+v", tokenRequest{Username: request.Username, GrantType: request.GrantType})
	}

	i := f.logins
	if i >= len(f.tokens) {
		i = len(f.tokens) - 1
	}
	f.logins++
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"accessToken":%q,"refreshToken":"refresh"}`, f.tokens[i])
}

func mintJWT(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func newTestSource(t *testing.T, fake *fakeIDM) oauth2.TokenSource {
	t.Helper()
	fake.t = t
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	return newTokenSource(context.Background(), server.Client(), server.URL+loginPath, "user@test", "secret")
}

func TestValidTokenIsCached(t *testing.T) {
	fake := &fakeIDM{tokens: []string{
		mintJWT(t, jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}),
	}}
	source := newTestSource(t, fake)

	for range 3 {
		token, err := source.Token()
		if err != nil {
			t.Fatalf("Token() failed: %v", err)
		}
		if token.AccessToken == "" {
			t.Fatal("empty access token")
		}
	}

	if fake.logins != 1 {
		t.Fatalf("expected 1 login, got %d", fake.logins)
	}
}

func TestExpiredTokenRefreshesWithoutError(t *testing.T) {
	fake := &fakeIDM{tokens: []string{
		mintJWT(t, jwt.MapClaims{"exp": time.Now().Add(-time.Minute).Unix()}),
		mintJWT(t, jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}),
	}}
	source := newTestSource(t, fake)

	if _, err := source.Token(); err != nil {
		t.Fatalf("first Token() failed: %v", err)
	}
	token, err := source.Token()
	if err != nil {
		t.Fatalf("Token() after expiry failed: %v", err)
	}
	if !token.Valid() {
		t.Fatal("expected a valid token after refresh")
	}
	if fake.logins != 2 {
		t.Fatalf("expected 2 logins, got %d", fake.logins)
	}
}

func TestTokenWithoutExpiryClaimFails(t *testing.T) {
	fake := &fakeIDM{tokens: []string{mintJWT(t, jwt.MapClaims{})}}
	source := newTestSource(t, fake)

	_, err := source.Token()
	if err == nil || !strings.Contains(err.Error(), "no expiry claim") {
		t.Fatalf("expected expiry-claim error, got: %v", err)
	}
}

func TestRejectedLoginNamesTheStep(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)
	source := newTokenSource(context.Background(), server.Client(), server.URL+loginPath, "user@test", "wrong")

	_, err := source.Token()
	if err == nil || !strings.Contains(err.Error(), "control-plane login") || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected login error with status, got: %v", err)
	}
	if strings.Contains(err.Error(), "wrong") {
		t.Fatal("error must not contain the password")
	}
}

func TestLoginDoesNotFollowRedirects(t *testing.T) {
	leaked := 0
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { leaked++ }))
	t.Cleanup(target.Close)
	redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+loginPath, http.StatusTemporaryRedirect)
	}))
	t.Cleanup(redirecting.Close)

	base, err := url.Parse(redirecting.URL)
	if err != nil {
		t.Fatalf("parsing test server URL: %v", err)
	}
	env := core.Env{ControlPlaneURL: redirecting.URL, Username: "user@test", Password: "secret"}
	source, err := credentials(context.Background(), env, base)
	if err != nil {
		t.Fatalf("credentials: %v", err)
	}

	_, err = source.Token()
	if err == nil || !strings.Contains(err.Error(), "307") {
		t.Fatalf("expected login to fail on redirect, got: %v", err)
	}
	if leaked != 0 {
		t.Fatalf("credentials were sent to the redirect target %d time(s)", leaked)
	}
}

func TestTokenIdentity(t *testing.T) {
	cases := []struct {
		name    string
		claims  jwt.MapClaims
		want    string
		wantErr bool
	}{
		{name: "subject claim", claims: jwt.MapClaims{"sub": "8f2c-user-uuid", "email": "ci@run.ai"}, want: "8f2c-user-uuid"},
		{name: "email alone is not an identity", claims: jwt.MapClaims{"email": "ci@run.ai"}, wantErr: true},
		{name: "no claims", claims: jwt.MapClaims{"exp": 1}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tokenIdentity(mintJWT(t, tc.claims))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q (%v), want %q", got, err, tc.want)
			}
		})
	}
}

func TestControlPlaneURL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{name: "empty", in: "", wantErr: "--cp-url"},
		{name: "no scheme", in: "cp.runailabs.com", wantErr: "must use https"},
		{name: "plain http rejected", in: "http://cp.runailabs.com", wantErr: "must use https"},
		{name: "no host", in: "https://", wantErr: "no host"},
		{name: "embedded credentials rejected", in: "https://user:pw@cp.runailabs.com", wantErr: "must not contain"},
		{name: "query rejected", in: "https://cp.runailabs.com?debug=1", wantErr: "must not contain"},
		{name: "fragment rejected", in: "https://cp.runailabs.com#top", wantErr: "must not contain"},
		{name: "https", in: "https://cp.runailabs.com", want: "https://cp.runailabs.com"},
		{name: "trailing slash trimmed", in: "https://cp.local/", want: "https://cp.local"},
		{name: "path prefix kept", in: "https://cp.local/runai/", want: "https://cp.local/runai"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := controlPlaneURL(tc.in)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got: %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("got %q, want %q", got.String(), tc.want)
			}
			if login := got.JoinPath(loginPath).String(); login != tc.want+loginPath {
				t.Fatalf("login endpoint %q, want %q", login, tc.want+loginPath)
			}
		})
	}
}
