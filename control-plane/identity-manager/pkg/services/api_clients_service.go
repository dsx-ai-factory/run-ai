/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	iamServices "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/services"
	iamUtils "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type ClientsApiService struct {
	iamClients iamServices.ClientService
	realmDal   business.RealmDal
}

func NewClientsApiService(iamServices *iam.IamServiceContainer, realmDal business.RealmDal) api.ClientsAPIServicer {
	return &ClientsApiService{
		iamClients: (*iamServices).ClientService,
		realmDal:   realmDal,
	}
}

func (s *ClientsApiService) CreateInternalClient(ctx context.Context, clientCreationRequest api.ClientCreationRequest) (api.ImplResponse, error) {
	clientId := clientCreationRequest.ClientId
	realmName := s.getRealmName(ctx, clientCreationRequest.TenantId)

	client, err := generateClientByType(ctx, clientCreationRequest)
	if err != nil {
		errMsg := fmt.Sprintf(`failed creating client. error to generate client by type: %s`, clientCreationRequest.Type)
		return utils.HandleError(ctx, "", "", "", err, errMsg)
	}
	id, err := s.iamClients.CreateClient(ctx, realmName, client)
	if err != nil {
		errMsg := fmt.Sprintf(`failed creating internal client with client id: %s`, clientId)
		return utils.HandleError(ctx, utils.InternalClient, utils.ClientId, clientId, err, errMsg)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Client with clientId %s was successfully created", clientId)
	return api.Response(http.StatusCreated, api.CreateInternalClient201Response{Id: id}), nil
}

func (s *ClientsApiService) DeleteClient(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	err := s.iamClients.DeleteClient(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`Failed deleting client with id: " %s`, id)
		return utils.HandleError(ctx, utils.InternalClient, utils.ID, id, err, errMsg)
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *ClientsApiService) GetAllClients(ctx context.Context, filter string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("calling get clients with filter: %s", filter)
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	filters := s.getFilters(filter)
	clients, err := s.iamClients.GetClients(ctx, realm, filters)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting clients with filter: %s`, filter)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s. err: %s", errMsg, err.Error())
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: errMsg}
	}
	return api.Response(http.StatusOK, clients), nil
}

func (s *ClientsApiService) GetClientSecret(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	secret, err := s.iamClients.GetClientSecret(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting secret for client with id: %s`, id)
		return utils.HandleError(ctx, utils.InternalClient, utils.ID, id, err, errMsg)
	}
	return api.Response(http.StatusOK, api.GetClientSecret200Response{Secret: secret}), nil
}

func (s *ClientsApiService) UpdateRedirectUris(ctx context.Context, clientId string, updateRedirectUrisRequest api.ClientRedirectUrisUpdateRequest) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)

	client, err := s.iamClients.GetClientByClientId(ctx, realm, clientId)
	if err != nil {
		errMsg := fmt.Sprintf("Update client: failed to find client for realm: %s, clientId: %s.", realm, clientId)
		return utils.HandleError(ctx, utils.InternalClient, utils.ID, clientId, err, errMsg)
	}
	if client == nil || client.ID == nil {
		errMsg := fmt.Sprintf("Update client: client not found for realm: %s, clientId: %s.", realm, clientId)
		return utils.HandleError(ctx, utils.InternalClient, utils.ID, clientId, utils.ErrNotFound, errMsg)
	}
	err = s.iamClients.UpdateRedirectUris(ctx, realm, *client.ID, updateRedirectUrisRequest.RedirectUris)
	if err != nil {
		errMsg := fmt.Sprintf("Update client: failed to update client for realm: %s, clientId: %s", realm, clientId)
		return utils.HandleError(ctx, utils.InternalClient, utils.ID, clientId, err, errMsg)
	}
	return api.Response(http.StatusOK, nil), nil
}

func (s *ClientsApiService) getRealmName(ctx context.Context, tenantId string) string {
	authRelamFromCtx := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if authRelamFromCtx != "" {
		return authRelamFromCtx
	}
	realm, _ := s.realmDal.GetRealmByTenantId(ctx, tenantId)
	if realm != nil {
		return realm.Realm
	}
	return ""
}

func (s *ClientsApiService) getFilters(input string) *iamModels.ClientFilter {
	clientFilters := iamModels.ClientFilter{}
	pairs := strings.Split(input, " ")

	for _, pair := range pairs {
		parts := strings.Split(pair, ":")
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			switch key {
			case iamUtils.ClusterIdAttr:
				clientFilters.ClusterId = &value
			case iamUtils.ClientTypeAttr:
				clientFilters.ClientType = &value
			case iamUtils.IsExternal:
				switch value {
				case "true":
					clientFilters.IsExternal = ptr.To(true)
				case "false":
					clientFilters.IsExternal = ptr.To(false)
				}
			}
		}
	}
	return &clientFilters
}

func generateClientByType(ctx context.Context, clientCreationRequest api.ClientCreationRequest) (*iamModels.Client, error) {
	clientType := clientCreationRequest.Type
	switch clientType {
	case api.CLIENTTYPE_AGENT:
		if clientCreationRequest.TenantId == "" {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Creating agent client failed. TenantId is missing.")
			return nil, utils.ErrInternal
		}
		client := createAgentClient(clientCreationRequest)
		return client, nil

	case api.CLIENTTYPE_WORKLOADS_AUTH:
		client := createWorkloadAuthClient(clientCreationRequest)
		return client, nil

	case api.CLIENTTYPE_APPLICATION:
		client := createExternalApplicationClient(clientCreationRequest)
		return client, nil

	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Client type %s is not supported or invalid.", clientType)
		return nil, utils.ErrInternal
	}
}

func createExternalApplicationClient(clientCreationRequest api.ClientCreationRequest) *iamModels.Client {
	clientHandler := realm_client.NewRunaiServiceAccountClient(clientCreationRequest.ClientId, clientCreationRequest.TenantId, nil, realm_client.ClientTypeTenantApp)
	return clientHandler.GetClientData(context.Background(), nil)
}

func createWorkloadAuthClient(clientCreationRequest api.ClientCreationRequest) *iamModels.Client {
	tenantRealm := *runai_config.Get[*config.TenantRealm]()
	clientHandler := realm_client.NewWorkloadsAuthClient(
		clientCreationRequest.ClientId,
		clientCreationRequest.TenantId,
		clientCreationRequest.Attributes.RunaiClusterId,
		tenantRealm.AppHost,
		tenantRealm.TenantDomainTemplate)
	return clientHandler.GetClientData(context.Background(), nil)
}

func createAgentClient(clientCreationRequest api.ClientCreationRequest) *iamModels.Client {
	clientHandler := realm_client.NewRunaiServiceAccountClient(
		clientCreationRequest.ClientId,
		clientCreationRequest.TenantId,
		&clientCreationRequest.Attributes.RunaiClusterId,
		realm_client.ClientTypeAgent)
	return clientHandler.GetClientData(context.Background(), nil)
}
