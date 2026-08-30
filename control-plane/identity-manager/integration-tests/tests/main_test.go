package tests

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"

	"strconv"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	iamAuth "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/auth"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/database/dao"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func newTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, TestUserName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, BaseTenantId)
	return ctx
}

var (
	host   = getenv("IDM_HOST", "localhost")
	dbHost = getenv("DB_HOST", "localhost")

	port               = getenv("IDM_PORT", "8080")
	dbPort             = getenv("DB_PORT", "5432")
	dbUser             = getenv("DB_USER", "runai")
	dbPassword         = getenv("DB_PASSWORD", "password")
	dbName             = getenv("DB_NAME", "runai")
	dbScheme           = getenv("DB_SCHEMA", "identity_manager")
	dbSslMode          = getenv("DB_SSL_MODE", "disable")
	dbMigrationCommand = getenv("DB_MIGRATION_COMMAND", "None")
	iamHost            = getenv("IAM_URL", "http://localhost:8090")
	iamUsername        = getenv("IAM_USERNAME", "admin")
	iamPassword        = getenv("IAM_PASSWORD", "admin")
	hostUrl            = fmt.Sprintf("http://%s:%s", host, port)
	FakeHostHeader     = getenv("HOST_HEADER", "FakeHostHeader")
	// Shared DAOs
	RealmDal        business.RealmDal
	IamServices     *iam.IamServiceContainer
	OpenapiClient   *openapi.APIClient
	TestStepsClient *TestSteps

	DefaultRealmPrefix = "realm-test-"
	BaseRealmName      = "runai"
	BaseTenantId       string
	MasterRealmName    = "master"
)

func initClients() {
	OpenapiClient = openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{{
			URL: hostUrl,
		}},
		HTTPClient: &http.Client{
			Transport: &FakeHostHeaderTransport{HeaderName: FakeHostHeader, Base: http.DefaultTransport},
		},
	})
	// Initialize database connection
	portInt, _ := strconv.Atoi(dbPort)
	autoMigrate := dbMigrationCommand != "None"
	dbConf := &dbconfig.Config{
		Connection: dbconfig.ConnectionConfig{
			Host:     dbHost,
			Port:     portInt,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			Schema:   dbScheme,
			SslMode:  dbconfig.SslMode(dbSslMode),
		},
		Migration: dbconfig.MigrationConfig{
			Auto:      &autoMigrate,
			Direction: dbconfig.MigrationDirectionUp,
		},
		ServiceName: "identity-manager",
	}
	if err := runaidb.Init(dbConf); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	RealmDal = dao.NewRealmDao()
	// Initialize iam services
	IamServices = iam.NewIamServices(iamHost, iamUsername, iamPassword)
	IamServices.SetLoginType(context.Background(), iamAuth.LoginTypeServiceAccount)
	TestStepsClient = NewTestSteps(IamServices, OpenapiClient, FakeHostHeader)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}
	logging.Init(nil)
	initClients()

	var err error
	BaseRealmName, BaseTenantId, err = TestStepsClient.CreateRealm(context.Background(), &BaseRealmName, StrPtr("1"))
	if err != nil {
		logging.Fatalf("Failed to create base realm: %v", err)
		os.Exit(1)
	}
	logging.Infof("Base realm created: %s, %s", BaseRealmName, BaseTenantId)

	code := m.Run()

	os.Exit(code)
}

func TestHealth(t *testing.T) {
	t.Parallel()
	health, err := OpenapiClient.HealthAPI.Health(context.Background()).Execute()
	require.NoError(t, err)
	require.Equal(t, health.StatusCode, http.StatusOK)
}

func TestGetConfig(t *testing.T) {
	t.Parallel()
	ctx := TestStepsClient.SetFakeHostHeader(newTestContext(), OpenapiClient, BaseRealmName)
	config, resp, err := OpenapiClient.ConfigAPI.GetTenantConfiguration(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK)
	require.Equal(t, config.AuthClientID, "runai-admin-ui")
	require.True(t, config.IsMultiTenant)
	require.Equal(t, config.TenantName, BaseRealmName)
}
