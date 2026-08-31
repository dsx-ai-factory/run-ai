/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricstores

import (
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
)

func APIToDBMetricStoreModel(ms *api.MetricStore) *models.MetricStore {
	if ms == nil {
		return nil
	}
	dbModel := &models.MetricStore{
		TenantId:     ms.TenantId,
		DatasourceId: ms.DatasourceId,
	}
	if details := APIToDBMetricStoreDetailsModel(ms.MetricStoreDetails); details != nil {
		dbModel.MetricStoreDetails = *details
	}
	return dbModel
}

func DBToAPIMetricStore(ms *models.MetricStore) *api.MetricStore {
	if ms == nil {
		return nil
	}

	return &api.MetricStore{
		TenantId:           ms.TenantId,
		DatasourceId:       ms.DatasourceId,
		MetricStoreDetails: DBToAPIMetricStoreDetails(&ms.MetricStoreDetails),
	}
}

func APIToDBMetricStoreDetailsModel(ms *api.MetricStoreDetails) *models.MetricStoreDetails {
	if ms == nil {
		return nil
	}

	return &models.MetricStoreDetails{
		Type:  models.MetricStoreType(ms.Type),
		Read:  *APIToDBMetricStoreEndpointModel(&ms.Read),
		Write: *APIToDBMetricStoreEndpointModel(&ms.Write),
	}
}

func DBToAPIMetricStoreDetails(ms *models.MetricStoreDetails) *api.MetricStoreDetails {
	if ms == nil {
		return nil
	}

	return &api.MetricStoreDetails{
		Type:  api.MetricStoreType(ms.Type),
		Read:  *DBToAPIMetricStoreEndpoint(&ms.Read),
		Write: *DBToAPIMetricStoreEndpoint(&ms.Write),
	}
}

func APIToDBMetricStoreEndpointModel(mse *api.MetricStoreEndpoint) *models.MetricStoreEndpoint {
	headers := map[string]string{}
	if mse.Headers != nil {
		headers = mse.Headers
	}
	return &models.MetricStoreEndpoint{
		Url:     mse.Url,
		Auth:    APIToDBMetricStoreEndpointAuthModel(mse.Auth),
		Headers: headers,
	}
}

func DBToAPIMetricStoreEndpoint(mse *models.MetricStoreEndpoint) *api.MetricStoreEndpoint {
	headers := map[string]string{}
	if mse.Headers != nil {
		headers = mse.Headers
	}

	return &api.MetricStoreEndpoint{
		Url:     mse.Url,
		Auth:    DBToAPIMetricStoreEndpointAuth(mse.Auth),
		Headers: headers,
	}
}

func APIToDBMetricStoreEndpointAuthModel(msea *api.MetricStoreEndpointAuth) *models.MetricStoreEndpointAuth {
	if msea == nil {
		return nil
	}

	return &models.MetricStoreEndpointAuth{
		Type:  models.AuthType(msea.Type),
		Basic: APIToDBMetricStoreEndpointBasicAuthModel(msea.Basic),
	}
}

func DBToAPIMetricStoreEndpointAuth(msea *models.MetricStoreEndpointAuth) *api.MetricStoreEndpointAuth {
	if msea == nil {
		return nil
	}

	return &api.MetricStoreEndpointAuth{
		Type:  api.MetricStoreEndpointAuthType(msea.Type),
		Basic: DBToAPIMetricStoreEndpointBasicAuth(msea.Basic),
	}
}

func APIToDBMetricStoreEndpointBasicAuthModel(ab *api.MetricStoreEndpointAuthBasic) *models.AuthBasic {
	if ab == nil {
		return nil
	}

	return &models.AuthBasic{
		Username: ab.Username, Password: ab.Password,
	}
}

func DBToAPIMetricStoreEndpointBasicAuth(ab *models.AuthBasic) *api.MetricStoreEndpointAuthBasic {
	if ab == nil {
		return nil
	}

	return &api.MetricStoreEndpointAuthBasic{
		Username: ab.Username, Password: ab.Password,
	}
}

func DatasourceToDBMetricStore(tenantId int32, datasourceId string, ds *Datasource) *models.MetricStore {
	if ds == nil {
		return nil
	}
	return &models.MetricStore{
		TenantId:     tenantId,
		DatasourceId: datasourceId,
		MetricStoreDetails: models.MetricStoreDetails{
			Type:  models.MetricStoreType(ds.Type),
			Read:  datasourceEndpointToDB(&ds.Read),
			Write: datasourceEndpointToDB(&ds.Write),
		},
	}
}

func datasourceEndpointToDB(ep *DatasourceEndpoint) models.MetricStoreEndpoint {
	return models.MetricStoreEndpoint{
		Url:  ep.Url,
		Auth: datasourceAuthToDB(ep.Auth),
	}
}

func datasourceAuthToDB(a *DatasourceAuth) *models.MetricStoreEndpointAuth {
	if a == nil {
		return nil
	}
	return &models.MetricStoreEndpointAuth{
		Type:  models.AuthType(a.Type),
		Basic: datasourceAuthBasicToDB(a.Basic),
	}
}

func datasourceAuthBasicToDB(b *AuthBasic) *models.AuthBasic {
	if b == nil {
		return nil
	}
	return &models.AuthBasic{
		Username: b.User,
		Password: b.Password,
	}
}
