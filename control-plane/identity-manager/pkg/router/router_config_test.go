package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	auth_manager "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/authorization"
	businessMocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/identity_provider"
	"github.com/stretchr/testify/require"
)

func TestConfigRouteDoesNotRevealWhetherTenantExists(t *testing.T) {
	tenantRealmArgs := &config.TenantRealm{}
	tenantRealmArgs.SetDefaults()
	tenantRealmArgs.IsMultiTenant = true

	realmDal := businessMocks.NewRealmDal(t)
	router := CreateRouter(
		iam.IamServiceContainer{},
		tenantRealmArgs,
		&config.Iam{},
		&config.Notify{},
		nil,
		identity_provider.IdpRepresentationFactory{},
		realmDal,
		nil,
		tenantRealmArgs.AppHost,
		auth_manager.AuthManagerInterface(nil),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/config", nil)
	req.Host = "unknown-tenant.run.ai"
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var response api.AuthConfiguration
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, "unknown-tenant", response.TenantName)
	require.Equal(t, "https://unknown-tenant.run.ai", response.RedirectUrl)
}
