/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metric_store

import (
	"context"
	"fmt"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	metrics_api "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	berrors "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/business/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

const (
	XScopeOrgId string = "X-Scope-OrgID"
)

type MetricStoreService interface {
	CreateMetricsStoreInfra(ctx context.Context, request *MetricStoreInfraCreateRequest) (*models.MetricStoreDetails, *berrors.BusinessError)
	DeleteMetricsStoreInfra(ctx context.Context, tenantId uint) *berrors.BusinessError
	GetTenantMetricStore(ctx context.Context, tenantId uint) (*models.MetricStore, *berrors.BusinessError)
	GetTenantExtendedMetricStore(ctx context.Context, tenantId uint) (*api.ExtendedMetricStore, *berrors.BusinessError)
}

type MetricStoreInfraCreateRequest struct {
	TenantId            uint
	TenantName          string
	MimirId             string
	ExistingMetricStore *models.MetricStoreDetails
}

type MetricStoreServiceImpl struct {
	metricStoreDao       dao.MetricStore
	metricStoreClient    metrics_client.MetricStoresClienter
	tenantDao            dao.Tenant
	businessErrorHandler berrors.BusinessErrorHandler
}

func NewMetricStoreService(
	metricStoreDao dao.MetricStore,
	tenantDao dao.Tenant,
	metricStoreClient metrics_client.MetricStoresClienter,
) MetricStoreService {
	return &MetricStoreServiceImpl{
		metricStoreDao:       metricStoreDao,
		metricStoreClient:    metricStoreClient,
		tenantDao:            tenantDao,
		businessErrorHandler: berrors.NewBusinessErrorHandler(),
	}
}

func (m *MetricStoreServiceImpl) CreateMetricsStoreInfra(ctx context.Context, request *MetricStoreInfraCreateRequest) (*models.MetricStoreDetails, *berrors.BusinessError) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Creating metric store for tenant [%d:%s]", request.TenantId, request.TenantName)
	metricStoreReq := m.buildCreateMetricStoreRequest(ctx, request)
	createdMetricStore, err := m.metricStoreClient.CreateMetricStore(ctx, metricStoreReq)
	if err != nil {
		apiErr, ok := err.(*metrics_api.Error)
		if !ok {
			return nil, &berrors.BusinessError{Reason: berrors.GeneralError, Err: err}
		}
		return nil, m.businessErrorHandler.MetricsAPIErrorToBusinessError(apiErr, berrors.MetricStore)
	}
	details := getMetricStoreDetails(createdMetricStore.MetricStoreDetails.Get())
	return &details, nil
}

func (m *MetricStoreServiceImpl) DeleteMetricsStoreInfra(ctx context.Context, tenantId uint) *berrors.BusinessError {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("deleting metric store for tenant: %d.", tenantId)
	err := m.metricStoreClient.DeleteMetricStores(ctx, int32(tenantId))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("error deleting metric store with id: %d. [%v]", tenantId, err)
		apiErr, ok := err.(*metrics_api.Error)
		if !ok {
			return &berrors.BusinessError{Reason: berrors.GeneralError, Err: err}
		}
		return m.businessErrorHandler.MetricsAPIErrorToBusinessError(apiErr, berrors.MetricStore)
	}
	return nil
}

func (m *MetricStoreServiceImpl) GetTenantMetricStore(ctx context.Context, tenantId uint) (metricStore *models.MetricStore, businessError *berrors.BusinessError) {
	tenant, daoErr := m.tenantDao.GetById(ctx, tenantId)
	if daoErr != nil {
		businessError = m.businessErrorHandler.DaoErrorToBusinessError(daoErr, berrors.Tenant)
		errMsg := fmt.Sprintf("failed to get tenant: %d. error: %v", tenantId, businessError.Error())
		if businessError.Reason == berrors.GeneralError {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMsg)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", errMsg)
		}
		return nil, businessError
	}
	return m.getTenantMetricStoreFromMetricsAPI(ctx, tenant)
}

func (m *MetricStoreServiceImpl) GetTenantExtendedMetricStore(ctx context.Context, tenantId uint) (extendedMetricStore *api.ExtendedMetricStore, businessError *berrors.BusinessError) {
	// This method uses the DAO for reading from the internal database. unlike GetTenantMetricStore that uses the metrics API.
	// It's necessary for the migration to work.
	tenant, daoErr := m.tenantDao.GetById(ctx, tenantId)
	if daoErr != nil {
		return nil, m.businessErrorHandler.DaoErrorToBusinessError(daoErr, berrors.Tenant)
	}
	metricStore, businessError := m.getTenantMetricStoreFromDAO(ctx, tenant)
	if businessError != nil {
		return nil, businessError
	}
	return m.extendMetricStore(ctx, metricStore, tenantId), nil
}

func (m *MetricStoreServiceImpl) getTenantMetricStoreFromMetricsAPI(ctx context.Context, tenant *models.Tenant) (*models.MetricStore, *berrors.BusinessError) {
	metricStore, err := m.metricStoreClient.GetMetricStore(ctx, int32(tenant.ID))
	if err != nil {
		businessErr := m.businessErrorHandler.MetricsAPIErrorToBusinessError(err.(*metrics_api.Error), berrors.MetricStore)
		if businessErr.Reason == berrors.NotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Metricstore not found in API for tenant with id: %d", tenant.ID)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error while getting metricstore from API for tenant: %d, error: %v", tenant.ID, businessErr.Err)
		}
		return nil, businessErr
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Metricstore found for tenant with id: %d, returning metrics store from db", tenant.ID)
	return fromAPIToModelMetricStore(metricStore), nil
}

func (m *MetricStoreServiceImpl) getTenantMetricStoreFromDAO(ctx context.Context, tenant *models.Tenant) (*models.MetricStore, *berrors.BusinessError) {
	metricStore, daoErr := m.metricStoreDao.GetByTenantId(ctx, tenant.ID)
	if daoErr != nil {
		businessErr := m.businessErrorHandler.DaoErrorToBusinessError(daoErr, berrors.MetricStore)
		if businessErr.Reason == berrors.NotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Metricstore not found in DAO for tenant with id: %d", tenant.ID)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error while getting metricstore from DAO for tenant: %d, error: %v", tenant.ID, businessErr.Err)
		}
		return nil, businessErr
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Metricstore found for tenant with id: %d, returning metrics store from db", tenant.ID)
	return metricStore, nil
}

func (m *MetricStoreServiceImpl) buildCreateMetricStoreRequest(ctx context.Context, request *MetricStoreInfraCreateRequest) *metrics_api.MetricStore {
	if request.ExistingMetricStore == nil {
		return &metrics_api.MetricStore{TenantId: int32(request.TenantId)}
	}
	metricStore := &models.MetricStore{
		TenantId:    request.TenantId,
		MetricStore: *request.ExistingMetricStore,
		Info:        models.MetricStoreInfo{Type: models.ExternalMetricStoreType},
	}
	extended := m.extendMetricStore(ctx, metricStore, request.TenantId)
	return fromExtendedToAPIMetricStore(extended)
}

func (m *MetricStoreServiceImpl) extendMetricStore(ctx context.Context, metricStore *models.MetricStore, tenantId uint) *api.ExtendedMetricStore {
	ems := &api.ExtendedMetricStore{
		TenantId: int32(tenantId),
		Read:     metricStore.MetricStore.Read,
		Write:    metricStore.MetricStore.Write,
	}

	if metricStore.Info.Type == models.GrafanaLabMetricStoreType && metricStore.Info.GrafanaLabIdentifier != "" {
		ems.Type = string(metrics_api.METRICSTORETYPE_GRAFANALABS)
		ems.DatasourceId = metricStore.Info.GrafanaLabIdentifier
		return ems
	}

	if orgId, ok := getXScopeOrgId(metricStore); ok {
		ems.Type = string(metrics_api.METRICSTORETYPE_MIMIR)
		ems.DatasourceId = orgId
		return ems
	}

	ems.Type = string(metrics_api.METRICSTORETYPE_THANOS)
	ems.DatasourceId = "default-metric-store"

	tenant, err := m.tenantDao.GetById(ctx, uint(tenantId))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant for extended metric store: %v", err)
	} else if tenant != nil {
		ems.DatasourceId = tenant.MimirId
	}

	return ems
}

func getXScopeOrgId(metricStore *models.MetricStore) (string, bool) {
	if metricStore.MetricStore.Write.Headers != nil {
		if orgId, ok := (*metricStore.MetricStore.Write.Headers)[XScopeOrgId]; ok {
			return orgId, ok
		}
	}
	if metricStore.MetricStore.Read.Headers != nil {
		orgId, ok := (*metricStore.MetricStore.Read.Headers)[XScopeOrgId]
		return orgId, ok
	}

	return "", false
}

func fromExtendedToAPIMetricStore(extendedMetricStore *api.ExtendedMetricStore) *metrics_api.MetricStore {
	if extendedMetricStore == nil {
		return nil
	}

	return &metrics_api.MetricStore{
		TenantId:     int32(extendedMetricStore.TenantId),
		DatasourceId: ptr.To(extendedMetricStore.DatasourceId),
		MetricStoreDetails: *metrics_api.NewNullableMetricStoreDetails(&metrics_api.MetricStoreDetails{
			Type:  metrics_api.MetricStoreType(extendedMetricStore.Type),
			Read:  fromAPIEndpointToMetricsAPIEndpoint(extendedMetricStore.Read),
			Write: fromAPIEndpointToMetricsAPIEndpoint(extendedMetricStore.Write),
		}),
	}
}

func fromAPIEndpointToMetricsAPIEndpoint(endpoint api.MetricStoreEndpoint) metrics_api.MetricStoreEndpoint {
	result := metrics_api.MetricStoreEndpoint{
		Url:     endpoint.Url,
		Headers: endpoint.Headers,
	}

	if endpoint.Auth != nil && endpoint.Auth.Basic != nil {
		result.Auth = *metrics_api.NewNullableMetricStoreEndpointAuth(&metrics_api.MetricStoreEndpointAuth{
			Type: ptr.To(metrics_api.METRICSTOREENDPOINTAUTHTYPE_BASIC),
			Basic: *metrics_api.NewNullableMetricStoreEndpointAuthBasic(
				&metrics_api.MetricStoreEndpointAuthBasic{
					Username: endpoint.Auth.Basic.Username,
					Password: endpoint.Auth.Basic.Password,
				},
			),
		})
	}

	return result
}

func fromAPIToModelMetricStore(metricStore *metrics_api.MetricStore) *models.MetricStore {
	if metricStore == nil {
		return nil
	}

	return &models.MetricStore{
		TenantId:    uint(metricStore.TenantId),
		Active:      true,
		MetricStore: getMetricStoreDetails(metricStore.MetricStoreDetails.Get()),
		Info:        getMetricStoreInfo(metricStore),
	}
}

func getMetricStoreDetails(details *metrics_api.MetricStoreDetails) models.MetricStoreDetails {
	if details == nil {
		return models.MetricStoreDetails{}
	}
	return models.MetricStoreDetails{
		Read: api.MetricStoreEndpoint{
			Url:     details.Read.Url,
			Auth:    getAuth(details.Read.Auth),
			Headers: details.Read.Headers,
		},
		Write: api.MetricStoreEndpoint{
			Url:     details.Write.Url,
			Auth:    getAuth(details.Write.Auth),
			Headers: details.Write.Headers,
		},
	}
}

func getAuth(auth metrics_api.NullableMetricStoreEndpointAuth) *api.MetricStoreEndpointAuth {
	if !auth.IsSet() || auth.Get() == nil {
		return nil
	}

	authValue := auth.Get()
	if !authValue.Basic.IsSet() || authValue.Basic.Get() == nil {
		return nil
	}

	basicValue := authValue.Basic.Get()
	return &api.MetricStoreEndpointAuth{
		Basic: &api.MetricStoreEndpointAuthBasic{
			Username: basicValue.Username,
			Password: basicValue.Password,
		},
	}
}

func getMetricStoreInfo(metricStore *metrics_api.MetricStore) models.MetricStoreInfo {
	details := metricStore.MetricStoreDetails.Get()
	if details == nil {
		return models.MetricStoreInfo{Type: models.ExternalMetricStoreType}
	}
	switch details.Type {
	case metrics_api.METRICSTORETYPE_GRAFANALABS:
		return models.MetricStoreInfo{
			Type:                 models.GrafanaLabMetricStoreType,
			GrafanaLabIdentifier: ptr.Deref(metricStore.DatasourceId, ""),
		}
	default:
		return models.MetricStoreInfo{
			Type: models.ExternalMetricStoreType,
		}
	}
}
