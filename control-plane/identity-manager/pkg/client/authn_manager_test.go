/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity_manager_client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	identity_manager_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var _ http.RoundTripper = (*identity_manager_client.AuthnManager)(nil)

func authnManagerConfig(mock *mocks.IIdmExternalClient, clientId, clientSecret string, refreshBuffer time.Duration) identity_manager_client.AuthnManagerConfig {
	return identity_manager_client.AuthnManagerConfig{
		ClientId:      clientId,
		ClientSecret:  clientSecret,
		IdmClient:     mock,
		RefreshBuffer: refreshBuffer,
	}
}

func tokenResponse(accessToken string, expiresInSeconds int32) *openapi.OAuth2TokenResponse {
	t := openapi.NewOAuth2TokenResponse(accessToken, "Bearer")
	t.SetExpiresIn(expiresInSeconds)
	return t
}

func TestNewAuthnManager(t *testing.T) {
	const clientId = "test-client"
	const clientSecret = "test-secret"

	tests := []struct {
		name    string
		setup   func(*mocks.IIdmExternalClient)
		wantErr bool
		check   func(*testing.T, *identity_manager_client.AuthnManager, *mocks.IIdmExternalClient)
	}{
		{
			name: "success",
			setup: func(m *mocks.IIdmExternalClient) {
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return(tokenResponse("token-1", 3600), nil).Once()
			},
			wantErr: false,
			check: func(t *testing.T, mgr *identity_manager_client.AuthnManager, _ *mocks.IIdmExternalClient) {
				token, err := mgr.Token(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, "token-1", token)
			},
		},
		{
			name: "ClientCredentialsLogin error",
			setup: func(m *mocks.IIdmExternalClient) {
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return((*openapi.OAuth2TokenResponse)(nil), assert.AnError).Once()
			},
			wantErr: true,
			check:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIdm := mocks.NewIIdmExternalClient(t)
			tt.setup(mockIdm)
			config := authnManagerConfig(mockIdm, clientId, clientSecret, 0)
			mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, mgr)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, mgr)
			if tt.check != nil {
				tt.check(t, mgr, mockIdm)
			}
		})
	}
}

func TestAuthnManager_Token(t *testing.T) {
	const clientId = "token-client"
	const clientSecret = "token-secret"

	tests := []struct {
		name  string
		setup func(*mocks.IIdmExternalClient) (refreshBuffer time.Duration, waitBeforeSecondCall time.Duration)
		check func(*testing.T, *identity_manager_client.AuthnManager, *mocks.IIdmExternalClient)
	}{
		{
			name: "returns cached token",
			setup: func(m *mocks.IIdmExternalClient) (time.Duration, time.Duration) {
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return(tokenResponse("cached-token", 3600), nil).Once()
				return 0, 0
			},
			check: func(t *testing.T, mgr *identity_manager_client.AuthnManager, _ *mocks.IIdmExternalClient) {
				token1, err := mgr.Token(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, "cached-token", token1)
				token2, err := mgr.Token(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, "cached-token", token2)
			},
		},
		{
			name: "refresh when expired",
			setup: func(m *mocks.IIdmExternalClient) (time.Duration, time.Duration) {
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return(tokenResponse("token-1", 1), nil).Once()
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return(tokenResponse("token-2", 3600), nil).Once()
				return 100 * time.Millisecond, 0
			},
			check: func(t *testing.T, mgr *identity_manager_client.AuthnManager, _ *mocks.IIdmExternalClient) {
				token1, err := mgr.Token(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, "token-1", token1)
				time.Sleep(1100 * time.Millisecond)
				token2, err := mgr.Token(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, "token-2", token2)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIdm := mocks.NewIIdmExternalClient(t)
			refreshBuffer, wait := tt.setup(mockIdm)
			config := authnManagerConfig(mockIdm, clientId, clientSecret, refreshBuffer)
			mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
			assert.NoError(t, err)
			assert.NotNil(t, mgr)
			if wait > 0 {
				time.Sleep(wait)
			}
			tt.check(t, mgr, mockIdm)
		})
	}
}

func TestAuthnManager_RoundTrip(t *testing.T) {
	const clientId = "roundtrip-client"
	const clientSecret = "roundtrip-secret"
	const expectedToken = "bearer-token-123"

	tests := []struct {
		name       string
		setup      func(*mocks.IIdmExternalClient)
		refreshBuf time.Duration
		waitBefore time.Duration
		wantErr    bool
		check      func(*testing.T, *http.Request)
	}{
		{
			name: "injects Authorization header",
			setup: func(m *mocks.IIdmExternalClient) {
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return(tokenResponse(expectedToken, 3600), nil).Once()
			},
			wantErr: false,
			check: func(t *testing.T, req *http.Request) {
				assert.Equal(t, "Bearer "+expectedToken, req.Header.Get("Authorization"))
			},
		},
		{
			name: "propagates Token error",
			setup: func(m *mocks.IIdmExternalClient) {
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return(tokenResponse("first", 0), nil).Once()
				m.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
					Return((*openapi.OAuth2TokenResponse)(nil), assert.AnError).Once()
			},
			refreshBuf: 10 * time.Millisecond,
			waitBefore: 20 * time.Millisecond,
			wantErr:    true,
			check:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIdm := mocks.NewIIdmExternalClient(t)
			tt.setup(mockIdm)
			config := authnManagerConfig(mockIdm, clientId, clientSecret, tt.refreshBuf)
			mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
			assert.NoError(t, err)
			assert.NotNil(t, mgr)

			if tt.waitBefore > 0 {
				time.Sleep(tt.waitBefore)
			}

			reqCh := make(chan *http.Request, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reqCh <- r
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
			resp, err := mgr.RoundTrip(req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			if tt.check != nil {
				tt.check(t, <-reqCh)
			}
		})
	}

	t.Run("retries once on 401 after token refresh", func(t *testing.T) {
		mockIdm := mocks.NewIIdmExternalClient(t)
		mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
			Return(tokenResponse("old-token", 3600), nil).Once()
		mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
			Return(tokenResponse("new-token", 3600), nil).Once()
		config := authnManagerConfig(mockIdm, clientId, clientSecret, 0)
		mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		var attempts atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch n := attempts.Add(1); n {
			case 1:
				assert.Equal(t, "Bearer old-token", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusUnauthorized)
			case 2:
				assert.Equal(t, "Bearer new-token", r.Header.Get("Authorization"))
				w.WriteHeader(http.StatusOK)
			default:
				t.Errorf("unexpected extra request %d", n)
			}
		}))
		defer server.Close()

		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		assert.NoError(t, err)
		resp, err := mgr.RoundTrip(req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
		assert.Equal(t, int32(2), attempts.Load())
		mockIdm.AssertExpectations(t)
	})
}

func TestAuthnManager_HttpClient(t *testing.T) {
	const clientId = "httpclient-client"
	const clientSecret = "httpclient-secret"

	t.Run("returns client with manager as transport", func(t *testing.T) {
		mockIdm := mocks.NewIIdmExternalClient(t)
		mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
			Return(tokenResponse("token", 3600), nil).Once()
		config := authnManagerConfig(mockIdm, clientId, clientSecret, 0)
		mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
		assert.NoError(t, err)
		assert.NotNil(t, mgr)

		client := mgr.HttpClient()
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client.Transport)
		transport, ok := client.Transport.(*identity_manager_client.AuthnManager)
		assert.True(t, ok, "Transport should be *AuthnManager")
		assert.Same(t, mgr, transport)
	})
}

func TestAuthnManager_Token_no_deadlock_on_refresh(t *testing.T) {
	const clientId = "deadlock-client"
	const clientSecret = "deadlock-secret"

	mockIdm := mocks.NewIIdmExternalClient(t)
	mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
		Return(tokenResponse("initial", 1), nil).Once()
	mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
		Return(tokenResponse("refreshed", 3600), nil).Once()

	config := authnManagerConfig(mockIdm, clientId, clientSecret, 100*time.Millisecond)
	mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
	assert.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)

	type result struct {
		token string
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		token, err := mgr.Token(context.Background())
		ch <- result{token, err}
	}()

	select {
	case res := <-ch:
		assert.NoError(t, res.err)
		assert.Equal(t, "refreshed", res.token)
	case <-time.After(3 * time.Second):
		t.Fatal("Token() deadlocked: did not return within timeout")
	}
	mockIdm.AssertExpectations(t)
}

func TestAuthnManager_Token_thundering_herd(t *testing.T) {
	const clientId = "herd-client"
	const clientSecret = "herd-secret"
	const numGoroutines = 50

	mockIdm := mocks.NewIIdmExternalClient(t)
	mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
		Return(tokenResponse("initial", 1), nil).Once()
	mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
		Return(tokenResponse("refreshed", 3600), nil).Once()

	config := authnManagerConfig(mockIdm, clientId, clientSecret, 100*time.Millisecond)
	mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
	assert.NoError(t, err)

	time.Sleep(1100 * time.Millisecond)

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			token, err := mgr.Token(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, "refreshed", token)
		}()
	}
	close(start)
	wg.Wait()

	mockIdm.AssertExpectations(t)
}

func TestAuthnManager_Token_concurrent(t *testing.T) {
	const clientId = "concurrent-client"
	const clientSecret = "concurrent-secret"
	const tokenValue = "concurrent-token"
	numGoroutines := 30
	callsPerGoroutine := 10

	mockIdm := mocks.NewIIdmExternalClient(t)
	mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
		Return(tokenResponse(tokenValue, 3600), nil).Once()
	config := authnManagerConfig(mockIdm, clientId, clientSecret, 0)
	mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				token, err := mgr.Token(ctx)
				assert.NoError(t, err)
				assert.Equal(t, tokenValue, token)
			}
		}()
	}
	wg.Wait()
	mockIdm.AssertExpectations(t)
}

func TestAuthnManager_RoundTrip_concurrent(t *testing.T) {
	const clientId = "roundtrip-concurrent-client"
	const clientSecret = "roundtrip-concurrent-secret"
	const tokenValue = "roundtrip-concurrent-token"
	numGoroutines := 20
	requestsPerGoroutine := 5

	mockIdm := mocks.NewIIdmExternalClient(t)
	mockIdm.On("ClientCredentialsLoginV2", mock.Anything, clientId, clientSecret).
		Return(tokenResponse(tokenValue, 3600), nil).Once()
	config := authnManagerConfig(mockIdm, clientId, clientSecret, 0)
	mgr, err := identity_manager_client.NewAuthnManager(context.Background(), config)
	assert.NoError(t, err)
	assert.NotNil(t, mgr)

	client := mgr.HttpClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if want := "Bearer " + tokenValue; r.Header.Get("Authorization") != want {
			http.Error(w, "wrong auth", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerGoroutine; j++ {
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
				resp, err := client.Transport.RoundTrip(req)
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		}()
	}
	wg.Wait()
	mockIdm.AssertExpectations(t)
}
