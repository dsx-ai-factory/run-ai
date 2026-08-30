package metric_store

import (
	"context"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/errors"
	metric_store "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/metric-store"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	builderIdentifier = "metric_store"
)

type Builder struct {
	metricsInfraService metric_store.MetricStoreService
}

func NewBuilder(
	metricsInfraService metric_store.MetricStoreService,
) builders.Interface {
	return &Builder{
		metricsInfraService: metricsInfraService,
	}
}

func (b *Builder) Name() string {
	return builderIdentifier
}

func (b *Builder) Construct(ctx context.Context, tenant *builders.TenantInfo, _ map[string]interface{}) (err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Constructing Metric store for tenant: [%s] id:[%d]", tenant.Name, tenant.Id)
	request := &metric_store.MetricStoreInfraCreateRequest{
		TenantId:            tenant.Id,
		TenantName:          tenant.Name,
		MimirId:             tenant.MimirId,
		ExistingMetricStore: tenant.MetricStore,
	}
	_, businessErr := b.metricsInfraService.CreateMetricsStoreInfra(ctx, request)
	if businessErr != nil && businessErr.Reason != errors.Conflict {
		return businessErr.Err
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Succesfully construct metric store for tenant: [%s] id:[%d]", tenant.Name, tenant.Id)
	return nil
}

func (b *Builder) Destruct(ctx context.Context, tenant *builders.TenantInfo) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Destructing Metric store for tenant: [%s] id:[%d]", tenant.Name, tenant.Id)
	businessErr := b.metricsInfraService.DeleteMetricsStoreInfra(ctx, tenant.Id)
	if businessErr != nil {
		if businessErr.Reason == errors.NotFound || businessErr.Reason == errors.TenantNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Failed to find tenant or metrics store in db. tenant: %s. skipping this step... %v", tenant.Name, businessErr.Error())
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete metrics store for tenant: %s: %v", tenant.Name, businessErr.Error())
			return businessErr.Err
		}
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Succesfully deleted metrics store for tenant: %s", tenant.Name)
	return nil
}
