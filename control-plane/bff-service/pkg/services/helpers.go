package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/bff-service/internal/generated/pkg/api"
	auth_manager "github.com/run-ai/runai/control-plane/bff-service/pkg/authorization"
	settings_manager "github.com/run-ai/runai/control-plane/bff-service/pkg/settings"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	errHandler "github.com/run-ai/runai/runai-common-packages/setting/error"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

// SettingResponseTransformer is an interface for settings that need custom response formatting
type SettingResponseTransformer interface {
	manager.Settinger
	TransformToResponse() interface{}
}

type SettingsHelper struct {
	authManager auth_manager.AuthManagerInterface
}

func NewSettingsHelper() *SettingsHelper {
	return &SettingsHelper{
		authManager: auth_manager.NewAuthManager(),
	}
}

func (s *SettingsHelper) HandleTenantAuthorization(ctx context.Context, tenantId string, operationDescription string) (*string, *api.ImplResponse, error) {
	tenantID, err := s.authManager.GetTenantID(ctx, tenantId)
	if err != nil {
		if isUnauthorizedError(err) {
			errMsg := fmt.Sprintf("unauthorized to %s. for tenant id: %s.", operationDescription, tenantId)
			response := api.Response(http.StatusForbidden, err.Error())
			return nil, &response, utils.Error{Code: http.StatusForbidden, Message: errMsg}
		}
		errMsg := fmt.Sprintf("failed to %s for tenant id: %s.", operationDescription, tenantId)
		response := api.Response(http.StatusInternalServerError, nil)
		return nil, &response, utils.Error{Code: http.StatusInternalServerError, Message: errMsg}
	}
	return tenantID, nil, nil
}

func (s *SettingsHelper) HandleSettingsError(err error, tenantId string, operationDescription string) (api.ImplResponse, error) {
	if strings.Contains(strings.ToLower(err.Error()), settings_manager.ErrUnknownSettingKey) {
		return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: err.Error()}
	}
	if _, ok := err.(errHandler.ErrorBadRequest); ok {
		return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: err.Error()}
	}
	errMsg := fmt.Sprintf("failed to %s for tenant id: %s.", operationDescription, tenantId)
	return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: errMsg}
}

func isUnauthorizedError(err error) bool {
	errMsg := strings.ToLower(err.Error())
	return strings.HasPrefix(errMsg, "unauthorized")
}

func (s *SettingsHelper) WrapSettingInResponse(setting manager.Settinger) interface{} {
	if transformer, ok := setting.(SettingResponseTransformer); ok {
		return transformer.TransformToResponse()
	}

	return map[string]interface{}{
		"value":       setting,
		"displayName": setting.DisplayName(),
		"description": setting.Description(),
	}
}
