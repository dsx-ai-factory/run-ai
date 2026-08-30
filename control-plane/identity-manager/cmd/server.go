package main

import (
	"os"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaihttp"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
	"github.com/run-ai/runai/runai-common-packages/setting"

	internalCache "github.com/run-ai/runai/control-plane/identity-manager/internal/cache"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	auth_manager "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/authorization"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/pubsub"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/identity_provider"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/router"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/services"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/settings"
)

func initServer() {
	runaipubsub.SetLogger(logging.Actual())
	if _, err := runaipubsub.InitAsync(nil); err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	}

	setting.Init(nil, nil)
	setting.Register[*settings.LogoutRedirectUri]()
	setting.Register[*settings.AutoRedirectSSO]()
	setting.Register[*settings.BrowserSessionTimeout]()

	tenantRealm := runai_config.Get[*config.TenantRealm]()
	internalCache.InitCacheStores(tenantRealm.ClientCachingExpiration)
}

func runServer(
	iamServices *iam.IamServiceContainer,
	masterReconciler reconciler.MasterReconcilerI,
	realmDal business.RealmDal,
	idpDal business.IdpDal,
	tenantRealmArgs *config.TenantRealm,
	iamArgs *config.Iam,
	notifyArgs *config.Notify,
) {
	initServer()

	authManager := auth_manager.NewAuthManager(realmDal)
	serviceAccountsApiService := services.NewServiceAccountsApiService(iamServices, authManager, realmDal)
	idpRepresentationFactory := identity_provider.NewIdpRepresentationFactory()

	pubsub.Register(serviceAccountsApiService)
	serveHTTP(*iamServices, tenantRealmArgs, iamArgs, notifyArgs, masterReconciler, idpRepresentationFactory, realmDal, idpDal, tenantRealmArgs.AppHost, authManager)
}

func serveHTTP(iamServices iam.IamServiceContainer,
	tenantRealmArgs *config.TenantRealm,
	iamArgs *config.Iam,
	notifyArgs *config.Notify,
	masterReconciler reconciler.MasterReconcilerI,
	idpFactory identity_provider.IdpRepresentationFactory,
	realmDal business.RealmDal,
	idpDal business.IdpDal,
	appHost string,
	authManager auth_manager.AuthManagerInterface,
) {
	runaihttp.ServeHTTP(
		router.CreateRouter(iamServices, tenantRealmArgs, iamArgs, notifyArgs, masterReconciler, idpFactory, realmDal, idpDal, appHost, authManager),
		nil,
	)
	os.Exit(0)
}
