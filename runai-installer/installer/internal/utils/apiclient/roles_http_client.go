/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-installer/clients/pkg/tls"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/logging"
	"github.com/golang-jwt/jwt/v5"
	"k8s.io/utils/ptr"
)

type Role struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	System      bool   `json:"system,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// The roles API is paginated; request the maximum page size to keep the number of pages small.
const rolesPageLimit = 500

type rolesResponse struct {
	Roles []Role `json:"roles"`
	Next  *int32 `json:"next,omitempty"`
}

type RolesHTTPClient struct {
	httpClient  *http.Client
	tokenApi    idmclient.TokensAPI
	serverURL   string
	username    string
	password    string
	accessToken string
	tokenExpiry time.Time
	mu          sync.Mutex
}

func NewRolesHTTPClient(fqdn, username, password, caCertPath string) (*RolesHTTPClient, error) {
	serverURL := normalizeServerURL(fqdn)

	// Create HTTP client with optional custom CA
	httpClient, err := tls.NewHTTPClientWithCA(caCertPath, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create IDM client configuration with custom HTTP client
	idmCfg := idmclient.NewConfiguration()
	idmCfg.Servers = idmclient.ServerConfigurations{{URL: serverURL}}
	idmCfg.HTTPClient = httpClient
	idmClient := idmclient.NewAPIClient(idmCfg)

	client := &RolesHTTPClient{
		httpClient: httpClient,
		tokenApi:   idmClient.TokensAPI,
		serverURL:  serverURL,
		username:   username,
		password:   password,
	}

	if err := client.refreshToken(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to acquire initial token: %w", err)
	}

	return client, nil
}

func normalizeServerURL(fqdn string) string {
	if len(fqdn) > 0 && fqdn[0] != 'h' {
		return fmt.Sprintf("https://%s", fqdn)
	}
	return fqdn
}

func (c *RolesHTTPClient) refreshToken(ctx context.Context) error {
	logging.Infof("[RolesHTTPClient] Refreshing OAuth token for %s (user: %s)", c.serverURL, c.username)

	request := idmclient.TokenRequest{
		Username:  ptr.To(c.username),
		Password:  ptr.To(c.password),
		GrantType: idmclient.TokenRequestGrantTypeEnum("password").Ptr(),
	}
	ctx = context.WithValue(ctx, "user-agent", "runai-client/1.0.0")
	token, _, err := c.tokenApi.GrantToken(ctx).TokenRequest(request).Execute()
	if err != nil {
		logging.Errorf("[RolesHTTPClient] Failed to refresh token: %v", err)
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	accessToken := token.GetAccessToken()
	expiry, err := extractTokenExpiry(accessToken)
	if err != nil {

		expiry = time.Now().Add(5 * time.Minute)
		logging.Infof("[RolesHTTPClient] Could not parse JWT expiry, using fallback: %v", expiry)
	}

	c.accessToken = accessToken
	c.tokenExpiry = expiry

	logging.Infof("[RolesHTTPClient] Token refreshed successfully, expires at %s", expiry.Format(time.RFC3339))
	return nil
}

func extractTokenExpiry(tokenString string) (time.Time, error) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, claims)
	if err != nil {
		logging.Errorf("[RolesHTTPClient] Failed to parse token: %v", err)
		return time.Time{}, err
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		logging.Errorf("[RolesHTTPClient] Invalid or missing 'exp' claim in token")
		return time.Time{}, fmt.Errorf("exp claim not found or invalid type in JWT token")
	}

	timeObj := time.Unix(int64(exp), 0)
	return timeObj, nil
}

func (c *RolesHTTPClient) getValidToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken == "" || time.Now().Add(30*time.Second).After(c.tokenExpiry) {
		if err := c.refreshToken(ctx); err != nil {
			return "", err
		}
	}

	return c.accessToken, nil
}

func (c *RolesHTTPClient) GetRoles(ctx context.Context) ([]Role, int, error) {
	logging.Infof("[RolesHTTPClient] Fetching roles from %s", c.serverURL)
	roles := []Role{}
	offset := int32(0)
	for {
		page, statusCode, err := c.getRolesPage(ctx, offset)
		if err != nil {
			return nil, statusCode, err
		}
		roles = append(roles, page.Roles...)
		// Stop on a next that does not move forward as well, so a bad response cannot loop forever.
		if page.Next == nil || *page.Next <= offset {
			logging.Infof("[RolesHTTPClient] Successfully fetched %d roles", len(roles))
			return roles, statusCode, nil
		}
		offset = *page.Next
	}
}

func (c *RolesHTTPClient) getRolesPage(ctx context.Context, offset int32) (*rolesResponse, int, error) {
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get valid token: %w", err)
	}

	rolesURL := fmt.Sprintf("%s/api/v2/authorization/roles?limit=%d&offset=%d", c.serverURL, rolesPageLimit, offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rolesURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "runai-client/1.0.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logging.Errorf("[RolesHTTPClient] Failed to execute request: %v", err)
		return nil, 0, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logging.Errorf("[RolesHTTPClient] Error closing response body: %v", err)
		}
	}()
	statusCode := resp.StatusCode
	if statusCode == http.StatusUnauthorized {
		logging.Infof("[RolesHTTPClient] Received 401, attempting token refresh and retry")
		c.mu.Lock()

		c.tokenExpiry = time.Time{}
		c.mu.Unlock()

		token, err = c.getValidToken(ctx)
		if err != nil {
			return nil, statusCode, fmt.Errorf("failed to refresh token after 401: %w", err)
		}
		req, _ = http.NewRequestWithContext(ctx, http.MethodGet, rolesURL, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "runai-client/1.0.0")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to execute retry request: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logging.Errorf("[RolesHTTPClient] Error closing response body: %v", err)
			}
		}()
		statusCode = resp.StatusCode
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, statusCode, fmt.Errorf("failed to read response body: %w", err)
	}
	if statusCode != http.StatusOK {
		logging.Errorf("[RolesHTTPClient] GetRoles failed with status %d: %s", statusCode, string(body))
		return nil, statusCode, fmt.Errorf("GetRoles failed with status %d: %s", statusCode, string(body))
	}
	var response rolesResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, statusCode, fmt.Errorf("failed to parse roles response: %w", err)
	}
	return &response, statusCode, nil
}

func (c *RolesHTTPClient) GetRoleIDByName(ctx context.Context, roleName string) (int32, error) {
	logging.Infof("[RolesHTTPClient] Getting role ID for role: %s", roleName)
	roles, _, err := c.GetRoles(ctx)
	if err != nil {
		return 0, err
	}
	for _, role := range roles {
		if role.Name == roleName {
			logging.Infof("[RolesHTTPClient] Found role '%s' with ID: %d", roleName, role.ID)
			return role.ID, nil
		}
	}
	return 0, fmt.Errorf("role '%s' not found", roleName)
}
