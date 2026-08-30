package tests

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	openapi "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	FieldEnabled = "enabled"

	// 1×1 transparent PNG as a data URL, used as a safe valid logo in tests.
	TRANSPARENT_PNG_DATA_URL = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+V4nEAAAAASUVORK5CYII="

	// Minimal JPEG header as a data URL, used to trigger 415 in tests.
	// /9j/ is exactly the base64 encoding of the 3-byte JPEG magic FF D8 FF.
	JPEG_DATA_URL = "data:image/jpeg;base64,/9j/"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

var (
	host          = getenv("BFF_SERVICE_HOST", "localhost")
	port          = getenv("BFF_SERVICE_PORT", "8080")
	dbHost        = getenv("DB_HOST", "localhost")
	dbPort        = getenv("DB_PORT", "5432")
	dbUser        = getenv("DB_USER", "runai")
	dbPassword    = getenv("DB_PASSWORD", "password")
	dbName        = getenv("DB_NAME", "runai")
	dbScheme      = getenv("DB_SCHEMA", "bff_service")
	hostUrl       = fmt.Sprintf("http://%s:%s", host, port)
	OpenapiClient *openapi.APIClient
	BaseTenantId  = "base-tenant-id"
)

func initClients() {
	OpenapiClient = openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{{
			URL: hostUrl,
		}},
	})
	logging.Init(nil)
}

func newTestContextWithTenant(tenantID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "test@run.ai")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)
	return ctx
}

func newTestContextWithUniqueTenant() context.Context {
	return newTestContextWithTenant(uuid.NewString())
}

func newTestContextWithManagementTenant() context.Context {
	ctx := newTestContextWithTenant(BaseTenantId)
	return context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}
	initClients()

	code := m.Run()

	os.Exit(code)
}

func TestHealth(t *testing.T) {
	t.Parallel()
	health, err := OpenapiClient.HealthAPI.Health(context.Background()).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, health.StatusCode)
}

func TestVersion(t *testing.T) {
	t.Parallel()
	version, resp, err := OpenapiClient.VersionAPI.GetVersion(context.Background()).Execute()
	require.NoError(t, err)
	require.Equal(t, "1.0.0", version.Version)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestBootConfigIsAvailableWithoutAuthentication(t *testing.T) {
	t.Parallel()
	bootConfig, resp, err := OpenapiClient.BootConfigAPI.GetBootConfig(context.Background()).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "1.0.0", bootConfig.Version.Display)
	require.Equal(t, "https://run-ai-docs.nvidia.com/self-hosted/1.0", bootConfig.Docs.BaseUrl)
	require.Equal(t, "What's New in 1.0", bootConfig.Docs.WhatsNew.Label)
	require.True(t, bootConfig.Capabilities.ShowsClusterLocationStep)
	require.True(t, bootConfig.UnleashConfig.IsSet())
	require.Nil(t, bootConfig.UnleashConfig.Get())
}
