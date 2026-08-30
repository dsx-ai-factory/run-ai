package flows

import (
	"context"
	"net/http"
	"time"

	tenant_provider "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/tenant"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	pollInterval           = 3 * time.Second
	tenantReadyPollTimeout = 3 * time.Minute
)

func WaitForTenantReady(ctx context.Context, tenantProvider tenant_provider.TenantProvider, tenantName string) *models.Tenant {
	deadline := time.Now().Add(tenantReadyPollTimeout)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed waiting for tenant %s to be ready: %v", tenantName, ctx.Err())
			return nil
		}
		tenant, apiErr := tenantProvider.GetTenantByName(ctx, tenantName)
		if apiErr != nil || tenant == nil {
			if apiErr != nil && apiErr.Code != http.StatusNotFound {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("GetTenantByName(%s) failed (will retry): code=%d, message=%s", tenantName, apiErr.Code, apiErr.Message)
			}
			time.Sleep(pollInterval)
			continue
		}
		switch tenant.Status {
		case models.TenantReady:
			return tenant
		case models.TenantFailed:
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("tenant %s is Failed, waiting for recovery", tenantName)
		}
		time.Sleep(pollInterval)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed waiting for tenant %s to be ready: timed out", tenantName)
	return nil
}
