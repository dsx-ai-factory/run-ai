package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplication_GetInstallerApp_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.ServiceAccountsAPI.GetInstallerServiceAccount(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)

	require.Equal(t, "runai-internal.installer", resp.Name)
	require.NotNil(t, resp.Id)
	require.NotNil(t, resp.ClientSecret)
	require.NotNil(t, resp.ClientId)
}
