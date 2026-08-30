package services

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/control-plane/temp-backend-settings/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/dao"
	"github.com/run-ai/runai/control-plane/temp-backend-settings/pkg/types"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type tenantSettingValue struct {
	TenantID int         `json:"tenantId"`
	Value    interface{} `json:"value"`
}

type tenantSettingsByKey struct {
	Key    string               `json:"key"`
	Values []tenantSettingValue `json:"values"`
}

type InternalAPIService struct {
	settingsDao *dao.TenantSettingsDAO
}

func NewInternalAPIService(settingsDao *dao.TenantSettingsDAO) *InternalAPIService {
	return &InternalAPIService{
		settingsDao: settingsDao,
	}
}

func (s *InternalAPIService) GetSettingByKeyForAllTenants(ctx context.Context, key string) (api.ImplResponse, error) {
	dbKey, originalKey, ok := types.LookupTenantSettingKey(key)
	if !ok {
		return api.Response(http.StatusBadRequest, api.Error{
			Code:    http.StatusBadRequest,
			Message: "unknown setting key: " + key,
		}), nil
	}

	result, err := s.settingsDao.GetByKeyForAllTenants(ctx, dbKey)
	if err != nil {
		logging.Errorf("failed to get setting by key for all tenants: %v", err)
		return api.Response(http.StatusInternalServerError, api.Error{
			Code:    http.StatusInternalServerError,
			Message: "failed to get setting for all tenants",
		}), nil
	}

	values := make([]tenantSettingValue, 0, len(result))
	for tenantID, value := range result {
		values = append(values, tenantSettingValue{TenantID: tenantID, Value: value})
	}

	return api.Response(http.StatusOK, tenantSettingsByKey{Key: originalKey, Values: values}), nil
}
