package services

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type ConfigApiService struct {
	config *config.TenantRealm
}

func NewConfigApiService(config *config.TenantRealm) api.ConfigAPIServicer {
	return &ConfigApiService{
		config: config,
	}
}

func (s *ConfigApiService) GetTenantConfiguration(ctx context.Context) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if realm == "" {
		api.Response(http.StatusBadRequest, errors.New("failed to extract tenant name from request"))
	}
	redirectUrl, err := utils.TemplateBaseUrl(s.config.TenantDomainTemplate, realm)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to resolve redirect URL for realm %s: %v", realm, err)
	}
	return api.Response(http.StatusOK, api.AuthConfiguration{
		TenantName:    realm,
		AuthClientID:  s.config.DefaultClientId,
		AuthHost:      s.config.AppHost,
		IsMultiTenant: s.config.IsMultiTenant,
		RedirectUrl:   redirectUrl,
	}), nil
}
