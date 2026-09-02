/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	KeycloakDomain     = "discovery.url.com"
	dockerKeycloakHost = "keycloak"
	dockerKeycloakPort = "8080"
	localhostIP        = "127.0.0.1"
	defaultTimeout     = 30 * time.Second
	defaultKeepAlive   = 30 * time.Second
	dockerEnvPath      = "/.dockerenv"
	defaultRealmPrefix = "test-realm1"
)

var (
	usedTenantIDs = make(map[string]struct{})
	tenantIDMutex sync.Mutex
)

func GenerateRandomString(prefix string) string {
	return prefix + uuid.New().String()
}

func GenerateUsername() string {
	u := uuid.New().String()
	return fmt.Sprintf("test-%s@run.ai", u)
}

func RandomRealmName(prefix string) string {
	if prefix == "" {
		prefix = defaultRealmPrefix
	}
	n, err := crand.Int(crand.Reader, big.NewInt(1<<62))
	if err != nil {
		logging.Errorf("Error generating random int for realm name: %v", err)
		n = big.NewInt(time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), n.Int64())
}

func GenerateRandomTenantID() string {
	for {
		n, err := crand.Int(crand.Reader, big.NewInt(100000))
		if err != nil {
			logging.Errorf("Error generating random tenant ID: %v", err)
			n = big.NewInt(time.Now().UnixNano() % 100000)
		}
		id := fmt.Sprintf("%d", n.Int64()+2)
		tenantIDMutex.Lock()
		if _, exists := usedTenantIDs[id]; !exists {
			usedTenantIDs[id] = struct{}{}
			tenantIDMutex.Unlock()
			return id
		}
		tenantIDMutex.Unlock()
		// else, try again
	}
}

// isDocker returns true if running inside a Docker container
func isDocker() bool {
	_, err := os.Stat(dockerEnvPath)
	return err == nil
}

// NewLocalhostOverrideClient returns an http.Client that resolves discovery.url.com to the correct address
// - If running in Docker, resolves to keycloak:8080
// - If not running in Docker, resolves to 127.0.0.1 (with the original port)
func NewLocalhostOverrideClient() *http.Client {
	dialer := &net.Dialer{
		Timeout:   defaultTimeout,
		KeepAlive: defaultKeepAlive,
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(addr)
			if host == KeycloakDomain {
				if isDocker() {
					return dialer.DialContext(ctx, network, net.JoinHostPort(dockerKeycloakHost, dockerKeycloakPort))
				}
				// If running locally, resolve to localhost IP
				return dialer.DialContext(ctx, network, net.JoinHostPort(localhostIP, port))
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	return &http.Client{Transport: transport}
}

// getKeycloakToken retrieves an access token from Keycloak using the specified grant type.
// For password grant, provide username and password. For client_credentials, leave them empty.
// Example usage:
//
//	token, err := getKeycloakToken(tokenURL, clientID, clientSecret, username, password, "password")
//	token, err := getKeycloakToken(tokenURL, clientID, clientSecret, "", "", "client_credentials")
//
//nolint:unused
func GetKeycloakToken(tokenURL, clientID, clientSecret, username, password, grantType string) (string, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("grant_type", grantType)
	data.Set("scope", "openid profile email")
	if grantType == "password" {
		data.Set("username", username)
		data.Set("password", password)
	}

	client := NewLocalhostOverrideClient()
	resp, err := client.PostForm(tokenURL, data) //nolint:noctx // integration test helper, no context available
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("no access_token in response: %s", string(body))
	}
	return token, nil
}

func StrPtr(s string) *string { return &s }

func GetDefaultIDPMappers() *openapi.Mappers {
	return &openapi.Mappers{
		Gid:                 StrPtr("GID"),
		Uid:                 StrPtr("UID"),
		Groups:              StrPtr("GROUPS"),
		SupplementaryGroups: StrPtr("SUPPLEMENTARYGROUPS"),
		Email:               StrPtr("email"),
		FirstName:           StrPtr("firstName"),
		LastName:            StrPtr("lastName"),
	}
}
