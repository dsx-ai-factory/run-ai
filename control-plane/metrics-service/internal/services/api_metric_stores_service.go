package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	metricerrors "github.com/run-ai/runai/control-plane/metrics-service/internal/errors"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
)

type MetricStoresAPIService struct {
	metricStoreManager MetricStoreManager
}

func NewMetricStoresAPIService(metricStoreManager MetricStoreManager) *MetricStoresAPIService {
	return &MetricStoresAPIService{
		metricStoreManager: metricStoreManager,
	}
}

func (mss *MetricStoresAPIService) GetMetricStores(ctx context.Context, overrideTenantId int32) (api.ImplResponse, error) {
	tenantId, err := GetTenantId(ctx, overrideTenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing tenant id, err: %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id")
	}

	store, err := mss.metricStoreManager.GetByTenantId(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to fetch metric store, tenant [%d], err: %v", tenantId, err)
		apiErr := api.NewError(http.StatusInternalServerError, "Failed to fetch metric store", err)
		return NewErrorResponse(apiErr), fmt.Errorf("%+v", *apiErr)
	}

	if store == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("missing metric store for tenant [%d]", tenantId)
		return api.Response(http.StatusOK, []api.MetricStore{}), nil
	}

	return api.Response(http.StatusOK, []api.MetricStore{*store}), nil
}

func (mss *MetricStoresAPIService) CreateMetricStore(ctx context.Context, metricStore api.MetricStore) (api.ImplResponse, error) {
	ms, err := mss.metricStoreManager.Create(ctx, metricStore)
	if err == nil {
		return api.Response(http.StatusOK, ms), nil
	}

	if errors.Is(err, metricerrors.ErrMetricStoreAlreadyExists) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("metric store for tenant [%d] already exists, skipping create", metricStore.TenantId)
		return NewErrorResponse(api.NewError(http.StatusConflict, "Metric store already exists", err)), errors.New("metric store already exists")
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create metric store, tenant [%d], err: %v", metricStore.TenantId, err)
	apiErr := api.NewError(http.StatusInternalServerError, "Failed to create metric store", err)
	return NewErrorResponse(apiErr), fmt.Errorf("%+v", *apiErr)
}

func (mss *MetricStoresAPIService) DeleteMetricStores(ctx context.Context, overrideTenantId int32) (api.ImplResponse, error) {
	tenantId, err := GetTenantId(ctx, overrideTenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing tenant id, err: %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("missing tenant id")
	}

	err = mss.metricStoreManager.DeleteByTenantId(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete metric stores of tenant %d: %v", tenantId, err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to delete metric stores of tenant %d", tenantId)
	}

	return api.Response(http.StatusNoContent, nil), nil
}
