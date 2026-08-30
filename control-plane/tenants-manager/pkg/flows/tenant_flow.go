package flows

import (
	"context"
	"net/http"

	tenant_provider "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/tenant"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type tenantProvisionAction int

const (
	actionCreate         tenantProvisionAction = iota // tenant does not exist
	actionRequeueAndWait                              // tenant is Failed; requeue then wait
	actionWait                                        // tenant is Creating; already in flight
	actionSkip                                        // tenant is Ready/Deleting
)

func checkTenantProvisionState(
	ctx context.Context,
	provider tenant_provider.TenantProvider,
	name string,
) (tenantProvisionAction, *models.Tenant) {
	tenant, err := provider.GetTenantByName(ctx, name)
	if err != nil {
		if err.Code == http.StatusNotFound {
			return actionCreate, nil
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"failed to get tenant %s: code=%d — skipping", name, err.Code)
		return actionSkip, nil
	}
	switch tenant.Status {
	case models.TenantFailed:
		return actionRequeueAndWait, tenant
	case models.TenantCreating:
		return actionWait, tenant
	default:
		return actionSkip, tenant
	}
}
