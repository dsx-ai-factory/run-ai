/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type OpenIDConfigurationService struct {
	config      *config.TenantRealm
	iamServices *iam.IamServiceContainer
}

func NewOpenIDConfigurationService(config *config.TenantRealm, iamServices *iam.IamServiceContainer) api.OpenIDConnectAPIServicer {
	return &OpenIDConfigurationService{
		config:      config,
		iamServices: iamServices,
	}
}

func (s *OpenIDConfigurationService) GetOpenidConfiguration(ctx context.Context) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if realm == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to extract realm from context")
		return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: "failed to extract realm from context"}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Fetching OpenID configuration for realm: %s", realm)

	keycloakConfig, err := s.iamServices.FetchKeycloakConfiguration(ctx, realm)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to fetch Keycloak configuration: %v", err)
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to fetch OpenID configuration"}
	}

	return api.Response(http.StatusOK, s.modifyConfiguration(ctx, keycloakConfig, realm)), nil
}

func (s *OpenIDConfigurationService) GetOauthAuthorizationServerMetadata(ctx context.Context) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if realm == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to extract realm from context")
		return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: "failed to extract realm from context"}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Fetching OAuth authorization server metadata for realm: %s", realm)

	metadata, err := s.iamServices.FetchOAuthAuthorizationServerMetadata(ctx, realm)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to fetch OAuth authorization server metadata: %v", err)
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to fetch OAuth authorization server metadata"}
	}

	return api.Response(http.StatusOK, s.modifyConfiguration(ctx, metadata, realm)), nil
}

func (s *OpenIDConfigurationService) modifyConfiguration(ctx context.Context, keycloakConfig map[string]interface{}, realm string) map[string]interface{} {
	baseURL := s.config.AppHost
	if s.config.IsMultiTenant {
		rendered, err := utils.TemplateBaseUrl(s.config.TenantDomainTemplate, realm)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to render tenant domain template %q for realm %q: %v", s.config.TenantDomainTemplate, realm, err)
		} else {
			baseURL = rendered
		}
	}

	newTokenEndpoint := fmt.Sprintf("%s/api/v2/token", baseURL)

	keycloakConfig["token_endpoint"] = newTokenEndpoint

	if mtlsAliases, ok := keycloakConfig["mtls_endpoint_aliases"].(map[string]interface{}); ok {
		mtlsAliases["token_endpoint"] = newTokenEndpoint
	}

	keycloakConfig["grant_types_supported"] = models.OAuth2StandardGrantTypes

	return keycloakConfig
}
