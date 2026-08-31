/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	runaictxhttp "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	tracinghttp "github.com/run-ai/runai/runai-common-packages/tracing/http"
)

const keyCacheTTL = 24 * time.Hour

var cacheInitOnce sync.Once

func ensureCacheReady() {
	cacheInitOnce.Do(func() {
		if err := cache.Init(nil); err != nil {
			logging.Errorf("failed to initialize cache: %v", err)
			return
		}
		if err := cache.RegisterStore[rsa.PublicKey](options.Ttl(keyCacheTTL)); err != nil {
			logging.Errorf("failed to register keycloak public key cache store: %v", err)
		}
	})
}

type KeycloakClient struct {
	baseURL string
	http    *http.Client
}

func NewKeycloakClient(baseURL string) *KeycloakClient {
	return &KeycloakClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (kc *KeycloakClient) GetPublicKey(ctx context.Context, realm string) (*rsa.PublicKey, error) {
	ensureCacheReady()
	cacheKey := kc.baseURL + "|" + realm

	if key, err := cache.Get[rsa.PublicKey](ctx, cacheKey); err == nil && key != nil {
		return key, nil
	}

	key, err := kc.fetchPublicKey(ctx, realm)
	if err != nil {
		return nil, err
	}

	if err := cache.Set(ctx, cacheKey, *key); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to cache keycloak public key for realm %s: %v", realm, err)
	}

	return key, nil
}

func (kc *KeycloakClient) fetchPublicKey(ctx context.Context, realm string) (*rsa.PublicKey, error) {
	realmURL := fmt.Sprintf("%s/realms/%s", kc.baseURL, url.PathEscape(realm))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, realmURL, nil)
	if err != nil {
		return nil, fmt.Errorf("keycloak request build: %w", err)
	}
	runaictxhttp.SetHeaders(ctx, req)
	tracinghttp.SetHeaders(ctx, req)

	resp, err := kc.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("keycloak unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("keycloak returned %d for realm %s", resp.StatusCode, realm)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return parseRSACert(result.PublicKey)
}

func parseRSACert(b64cert string) (*rsa.PublicKey, error) {
	pemStr := "-----BEGIN CERTIFICATE-----\n" + b64cert + "\n-----END CERTIFICATE-----"
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from Keycloak public key")
	}

	var pub any
	if parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		pub = parsedKey
	} else if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		pub = cert.PublicKey
	} else {
		return nil, fmt.Errorf("failed to parse keycloak public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("keycloak public key is not RSA")
	}
	return rsaKey, nil
}

func RealmFromIssuer(issuer string) (string, error) {
	parts := strings.Split(strings.TrimRight(issuer, "/"), "/")
	if len(parts) < 2 || parts[len(parts)-2] != "realms" {
		return "", fmt.Errorf("invalid issuer URL: %s", issuer)
	}
	realm := parts[len(parts)-1]
	if realm == "" {
		return "", fmt.Errorf("empty realm in issuer URL: %s", issuer)
	}
	return realm, nil
}
