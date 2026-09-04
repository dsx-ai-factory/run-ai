/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business"
	auth_manager "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/authorization"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/identity_provider"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/reconciler"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/services"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
	runaictx "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	tracing "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/http"
)

func CreateRouter(iamServices iam.IamServiceContainer,
	tenantRealmArgs *config.TenantRealm,
	iamArgs *config.Iam,
	notifyArgs *config.Notify,
	masterReconciler reconciler.MasterReconcilerI,
	idpFactory identity_provider.IdpRepresentationFactory,
	realmDal business.RealmDal,
	idpDal business.IdpDal,
	appHost string,
	authManager auth_manager.AuthManagerInterface,
) chi.Router {
	HealthApiService := services.NewHealthApiService()
	HealthApiController := api.NewHealthAPIController(HealthApiService)

	ConfigApiService := services.NewConfigApiService(tenantRealmArgs)
	ConfigApiController := api.NewConfigAPIController(ConfigApiService)

	IdpApiService := services.NewIdpsApiService(&iamServices, idpFactory, idpDal, &authManager)
	IdpApiController := api.NewIdpsAPIController(IdpApiService)

	RealmApiService := services.NewRealmApiService(iamServices, tenantRealmArgs, &masterReconciler, realmDal)
	RealmApiController := api.NewRealmAPIController(RealmApiService)

	UsersApiService := services.NewUsersApiService(iamServices, iamArgs, notifyArgs, tenantRealmArgs.TenantDomainTemplate, authManager)
	UsersApiController := api.NewUsersAPIController(UsersApiService)

	ServiceAccountsApiService := services.NewServiceAccountsApiService(&iamServices, authManager, realmDal)
	ServiceAccountsApiController := api.NewServiceAccountsAPIController(ServiceAccountsApiService)

	InternalClientsApiService := services.NewClientsApiService(&iamServices, realmDal)
	InternalClientsApiController := api.NewClientsAPIController(InternalClientsApiService)

	TokensApiService := services.NewTokensApiService(tenantRealmArgs, &iamServices)
	TokensApiController := api.NewTokensAPIController(TokensApiService)

	OpenIDConfigService := services.NewOpenIDConfigurationService(tenantRealmArgs, &iamServices)
	OpenIDConfigController := api.NewOpenIDConnectAPIController(OpenIDConfigService)

	LoginsApiService := services.NewLoginsApiService(appHost)
	LoginsApiController := api.NewLoginsAPIController(LoginsApiService)

	AccessKeysApiService := services.NewAccessKeysAPIService(&iamServices)
	AccessKeysApiController := api.NewAccessKeysAPIController(AccessKeysApiService)

	SettingsApiService := services.NewSettingsAPIService(&iamServices, &masterReconciler, &authManager)
	SettingsController := api.NewSettingsAPIController(SettingsApiService)

	LogoutApiService := services.NewLogoutAPIService(&iamServices)
	LogoutApiController := api.NewLogoutAPIController(LogoutApiService)

	// Add routes to the router for each of the controllers
	router := api.NewRouter()

	api.AddRoutes(router, HealthApiController)

	// no setAuthRealmFromTenantId
	router.Group(func(r chi.Router) {
		r.Use(tracing.HttpMiddleware)
		r.Use(runaictx.ContextMiddleware)
		r.Use(runaimetrics.HttpMiddleware)
		api.AddRoutes(r, RealmApiController)
	})

	// Authorized routes
	router.Group(func(r chi.Router) {
		r.Use(runaimetrics.HttpMiddleware)
		r.Use(tracing.HttpMiddleware)
		r.Use(runaictx.ContextMiddleware, setAuthRealmFromTenantId(realmDal))
		api.AddRoutes(r,
			IdpApiController,
			UsersApiController,
			ServiceAccountsApiController,
			InternalClientsApiController,
			AccessKeysApiController,
			SettingsController,
			LogoutApiController,
		)
	})
	setupUnauthenticatedRealmRoute := func(r chi.Router) {
		r.Use(setRequestHost(tenantRealmArgs.HostHeader))
		r.Use(runaictx.ContextMiddleware)
		if tenantRealmArgs.Name != "" {
			r.Use(extractRealmNameUnauthenticated(tenantRealmArgs.Name))
		} else {
			r.Use(extractRealmNameUnauthenticatedMultiTenant(tenantRealmArgs.HostHeader, realmDal))
		}
	}
	setupUnauthenticatedRoute := func(r chi.Router) {
		setupUnauthenticatedRealmRoute(r)
		r.Use(extractTenantIdUnauthenticated(realmDal))
	}

	router.Group(func(r chi.Router) {
		setupUnauthenticatedRealmRoute(r)
		api.AddRoutes(r, ConfigApiController)
	})

	router.Group(func(r chi.Router) {
		setupUnauthenticatedRoute(r)
		api.AddRoutes(r, OpenIDConfigController, LoginsApiController)
	})

	router.Group(func(r chi.Router) {
		setupUnauthenticatedRoute(r)
		r.Use(cookieToHeaderMiddleware)
		r.Use(basicAuthClientCredsMiddleware)
		api.AddRoutes(r, TokensApiController)
	})

	return router
}
