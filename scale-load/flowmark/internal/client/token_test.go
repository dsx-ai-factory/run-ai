/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type fakeIDM struct {
	logins int
	tokens []string
}

func (f *fakeIDM) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
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
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)

	config := idmclient.NewConfiguration()
	config.Servers = idmclient.ServerConfigurations{{URL: server.URL}}
	return newTokenSource(context.Background(), idmclient.NewAPIClient(config).TokensAPI, "user@test", "secret")
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

func TestControlPlaneURL(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		want    string
		wantErr string
	}{
		{name: "empty", in: "", wantErr: "--cp-url"},
		{name: "no scheme", in: "cp.runailabs.com", wantErr: "http"},
		{name: "https", in: "https://cp.runailabs.com", want: "https://cp.runailabs.com"},
		{name: "trailing slash trimmed", in: "http://cp.local/", want: "http://cp.local"},
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
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
