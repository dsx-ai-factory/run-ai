package client_reconciler

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ClientsReconcilerConfig struct {
	hostTemplate string
	appHost      string
}

type ClientsReconciler struct {
	iamServices            iam.IamServiceContainer
	config                 ClientsReconcilerConfig
	tenantResolver         reconciler.TenantResolver
	clientReconcilerHelper ClientReconcilerHelperInterface
}

func NewClientsReconciler(ctx context.Context, iamServices *iam.IamServiceContainer, tenantRealm *config.TenantRealm, realmDal business.RealmDal) reconciler.Reconciler {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Init Clients Reconciler")
	tenantResolver := reconciler.NewRealmTenantResolver(ctx, realmDal)
	clientReconcilerHelper := NewClientReconcilerHelper(ctx, *iamServices, tenantResolver)

	cr := ClientsReconciler{
		iamServices:            *iamServices,
		tenantResolver:         tenantResolver,
		clientReconcilerHelper: clientReconcilerHelper,
		config: ClientsReconcilerConfig{
			hostTemplate: tenantRealm.TenantDomainTemplate,
			appHost:      tenantRealm.AppHost,
		},
	}
	return &cr
}

func (cr *ClientsReconciler) Reconcile(ctx context.Context, realm string, force bool) error {
	start := time.Now()
	defer reconciler.LogReconcileDuration(ctx, start, realm, "ClientsReconciler")
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling clients for realm %s", realm)

	tenantId := cr.tenantResolver.GetRealmTenantId(ctx, realm)
	if tenantId == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenant id for realm %s", realm)
		return fmt.Errorf("failed to get tenant id for realm %s", realm)
	}
	err := cr.reconcileDefaultClients(ctx, realm, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile default clients for realm %s: %v", realm, err)
		return err
	}

	return nil
}

func (cr *ClientsReconciler) reconcileDefaultClients(ctx context.Context, realm string, tenantId string) error {
	clients := []realm_client.Client{
		realm_client.NewAdminUiClient(realm, tenantId, cr.config.appHost, cr.config.hostTemplate),
		realm_client.NewRunaiCliClient(realm, tenantId, cr.config.appHost, cr.config.hostTemplate),
		realm_client.NewRunaiClient(realm, tenantId, cr.config.appHost, cr.config.hostTemplate),
		realm_client.NewRunaiServiceAccountClient(utils.RunaiInstallerClientId, tenantId, nil, realm_client.ClientTypeInstaller),
		realm_client.NewCLIClient(realm, tenantId, cr.config.appHost, cr.config.hostTemplate),
		realm_client.NewTokenExchangeClient(realm, tenantId, cr.config.appHost, cr.config.hostTemplate),
		realm_client.NewMcpClient(realm, tenantId),
	}

	for _, client := range clients {
		if err := cr.clientReconcilerHelper.ReconcileClient(ctx, realm, client); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to reconcile default client %s in realm %s: %v", client.GetClientID(), realm, err)
			return err
		}
	}
	return nil
}
