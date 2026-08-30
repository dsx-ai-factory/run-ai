package main

import (
	"context"
	"os"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/database/dao"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/auth"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	iamAuth "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/auth"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/notification"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler/factory"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/tracing"

	_ "github.com/run-ai/runai/runai-common-packages/fips"
)

func init() {
	logging.Init(nil)
	metrics.Init(nil)
	tracing.Init(nil)
	audit.Init(nil)
	if err := cache.Init(nil); err != nil {
		logging.Errorf("Failed to initialize cache: %v", err)
	}

	config.InitConfig()
	authorization_client.Init(runai_config.Get[*authorization_client.Config]())
	cfg := runai_config.Get[*config.Configuration]()
	if cfg.UseMocks {
		authorization_client.SetMockAuthorizationClient(auth.MockService{})
	}
	if cfg.AllowHTTPIdpUrls {
		utils.SetAllowHTTPScheme(true)
	}
	runaidb.SetLogger(logging.Actual())
}

func initDB() error {
	if !shouldReconcile() {
		return runaidb.Init(nil)
	}
	if err := dbconfig.InitConfig(); err != nil {
		return err
	}
	cfg := runai_config.Get[*dbconfig.Config]()
	noMigration := false
	cfg.Migration.Auto = &noMigration
	return runaidb.Init(cfg)
}

func main() {
	ctx := context.Background()
	if err := initDB(); err != nil {
		logging.Fatalf("Failed to initialize database: %v", err)
	}
	tenantRealmArgs := *runai_config.Get[*config.TenantRealm]()
	iamArgs := runai_config.Get[*config.Iam]()
	notifyArgs := runai_config.Get[*config.Notify]()
	iamServices := iam.NewIamServices(iamArgs.Url, iamArgs.AdminUsername, iamArgs.AdminPassword)
	err := iamServices.SetLoginType(ctx, iamAuth.LoginTypeServiceAccount)
	if err != nil {
		logging.Errorf("Failed to set iam token: %v", err)
	}

	realmDal := dao.NewRealmDao()
	reconcilerVersionsDal := dao.NewReconcilerVersionsDao()
	idpDal := dao.NewIdpDao()

	masterReconciler := factory.BuildMasterReconciler(ctx, *iamServices, realmDal, reconcilerVersionsDal, iamArgs)

	validateAccessToIam(iamServices, ctx)

	notification.Init(ctx)

	if shouldReconcile() {
		runReconcile(ctx, iamServices, masterReconciler, realmDal, iamArgs)
		return
	}
	runServer(iamServices, masterReconciler, realmDal, idpDal, &tenantRealmArgs, iamArgs, notifyArgs)
}

func shouldReconcile() bool {
	return len(os.Args) > 1 && os.Args[1] == "reconcile"
}

func validateAccessToIam(iamServices *iam.IamServiceContainer, ctx context.Context) {
	for {
		err := iamServices.AuthService.ValidateServiceAccess(ctx)
		if err == nil {
			return
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("No Access to Keycloak; waiting 5 seconds before retrying to get the access token.. err: %s", err)
			time.Sleep(5 * time.Second)
		}
	}
}
