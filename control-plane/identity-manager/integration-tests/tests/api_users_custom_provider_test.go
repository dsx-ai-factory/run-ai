package tests

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// customUsersEndpoint is the Run:ai custom user-listing SPI, mounted under Keycloak's
// standard admin API tree at /admin/realms/{realm}/runaienteties/users.
const customUsersEndpoint = "http://" + KeycloakDomain + ":8090/auth/admin/realms/%s/runaienteties/users"

// TestCustomUserProviderEndpoint_AdminTokenSucceeds is a sanity check that the custom
// user-listing endpoint is reachable under the standard admin API path, accepts a valid
// admin token (HTTP 200), and returns the expected {count, users} payload.
func TestCustomUserProviderEndpoint_AdminTokenSucceeds(t *testing.T) {
	ctx := newTestContext()

	const numTestUsers = 2
	createdUserIDs := createTestUsers(t, ctx, numTestUsers)
	defer cleanupTestUsers(t, ctx, createdUserIDs)

	token, err := GetKeycloakToken(
		fmt.Sprintf(keycloakTokenEndpoint, MasterRealmName),
		"admin-cli", "", iamUsername, iamPassword, "password")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf(customUsersEndpoint, BaseRealmName), nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := NewLocalhostOverrideClient().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result struct {
		Count int                 `json:"count"`
		Users []map[string]string `json:"users"`
	}
	require.NoError(t, json.Unmarshal(body, &result))
	require.GreaterOrEqual(t, result.Count, numTestUsers)
	require.GreaterOrEqual(t, len(result.Users), numTestUsers)
}

// TestCustomUserProviderEndpoint_RejectsMissingToken is a sanity check that the endpoint
// is protected by the standard admin authentication: an unauthenticated request is
// rejected with 401 instead of leaking user data.
func TestCustomUserProviderEndpoint_RejectsMissingToken(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequest(http.MethodGet, //nolint:noctx // simple sanity check
		fmt.Sprintf(customUsersEndpoint, BaseRealmName), nil)
	require.NoError(t, err)

	resp, err := NewLocalhostOverrideClient().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
