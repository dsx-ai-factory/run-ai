package manager_reconciler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"

	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/reconciler"
	tmc "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	tenantStatusReady     = "Ready"
	jsonField             = "jsonField"
	emailToUsernameMapper = "EmailToUsername Mapper"
)

type IamReconcileManager struct {
	realmDal             business.RealmDal
	masterReconciler     reconciler.MasterReconcilerI
	tenantsManagerClient tmc.ITenantsManager
	iam                  iam.IamServiceContainer
}

func NewReconcileManager(ctx context.Context,
	iamServices *iam.IamServiceContainer,
	masterReconciler reconciler.MasterReconcilerI,
	tenantsManagerClient tmc.ITenantsManager,
	realmDal business.RealmDal,
) reconciler.ReconcileManager {
	return IamReconcileManager{
		realmDal:             realmDal,
		masterReconciler:     masterReconciler,
		tenantsManagerClient: tenantsManagerClient,
		iam:                  *iamServices,
	}
}

func (rm IamReconcileManager) ShouldReconcile() bool {
	return rm.isEvenEqual(1, "reconcile")
}

func (rm IamReconcileManager) ShouldUpdateTenants(ctx context.Context) bool {
	updateTenantsConfig := rm.isEvenEqual(1, "updateTenantRealms") || rm.isEvenEqual(2, "updateTenantRealms")
	if !updateTenantsConfig {
		return false
	}
	realms, err := rm.realmDal.GetRealms(ctx)
	if err == nil && realms != nil && len(*realms) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("already updated tenants realms, skipping update")
		return false
	}
	return true
}

func (rm IamReconcileManager) isEvenEqual(i int, command string) bool {
	if len(os.Args) > i {
		arg := os.Args[i]
		if arg == command {
			return true
		}
	}
	return false
}

func (rm IamReconcileManager) AddAllTenantsToLocalSchema(ctx context.Context) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("adding tenant realms to local schema")
	tenants, err := rm.listAllTenants(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list all Tenants. Error %s", err.Error())
		return
	}
	size, err := rm.realmDal.CreateRealms(ctx, &tenants)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to add Tenants to local schema. Error %s", err.Error())
		return
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("added %d new Tenants (out of %d that were found) to local schema", size, len(tenants))
}

func (rm IamReconcileManager) listAllTenants(ctx context.Context) ([]models.Realm, error) {
	tenants, resp, err := rm.tenantsManagerClient.ListAllTenantsInternal(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list tenants. resp: %s, status code: %d", resp.Body, resp.StatusCode)
	}

	res := []models.Realm{}
	for _, tenant := range tenants {
		if tenant.Status == tenantStatusReady {
			res = append(res, models.Realm{
				TenantId: strconv.Itoa(int(tenant.Id)),
				Realm:    tenant.Name,
			})
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("skipping adding of tenant %s to local schema. tenant status not ready: %s", tenant.Name, tenant.Status)
		}
	}
	return res, nil
}

func (rm IamReconcileManager) RunReconcile(ctx context.Context) error {
	iamConfigIgnoredRealms := utils.ConvertArrayToMap(runai_config.Get[*config.Iam]().IgnoredRealms)
	realms, err := rm.realmDal.GetRealms(ctx)
	if err != nil {
		errorMsg := "failed to get realms for tenant reconcile, err: %v"
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(errorMsg, err)
		return fmt.Errorf(errorMsg, err)
	}
	return rm.reconcileTenantRealms(ctx, realms, iamConfigIgnoredRealms, false)
}

func (rm IamReconcileManager) reconcileTenantRealms(ctx context.Context, realms *[]models.Realm, ignoredRealms map[string]bool, force bool) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling tenant realms command was called")

	err := rm.masterReconciler.ReconcileRealms(ctx, realms, ignoredRealms, force)
	if err != nil {
		errorMsg := "failed to reconcile all tenant realms, err: %v"
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(errorMsg, err)
		return fmt.Errorf(errorMsg, err)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Reconciling tenant realms was finished successfully")
	return nil
}
