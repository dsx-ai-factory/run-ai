/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"testing"
)

func TestNewTokenSource(t *testing.T) {
	for _, tc := range []struct {
		name      string
		creds     Credentials
		wantErr   bool
		wantToken string
	}{
		{name: "bearer token", creds: Credentials{Token: "bearer"}, wantToken: "bearer"},
		{name: "token wins over password", creds: Credentials{Username: "u", Password: "p", Token: "bearer"}, wantToken: "bearer"},
		{name: "password grant", creds: Credentials{Username: "u", Password: "p"}},
		{name: "missing password", creds: Credentials{Username: "u"}, wantErr: true},
		{name: "empty", creds: Credentials{}, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			src, err := newTokenSource("https://cp.example.com", "", tc.creds)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantToken == "" {
				if _, ok := src.(*TokenSource); !ok {
					t.Fatalf("expected the password-grant token source, got %T", src)
				}
				return
			}
			tok, err := src.Token()
			if err != nil {
				t.Fatalf("static token source failed: %v", err)
			}
			if tok.AccessToken != tc.wantToken {
				t.Fatalf("access token = %q, want %q", tok.AccessToken, tc.wantToken)
			}
		})
	}
}
