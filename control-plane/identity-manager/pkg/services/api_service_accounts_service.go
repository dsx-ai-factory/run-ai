/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"time"

	"k8s.io/utils/ptr"

	cache "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
	async_events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"

	"github.com/pkg/errors"

	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business"
	authBusiness "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	internalCache "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/cache"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam"
	iamServices "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	DefaultServiceAccountTokenLifespan = 30 * time.Minute
	installerBootstrapLockTTL          = 30 * time.Second
	installerBootstrapUnlockTimeout    = 5 * time.Second
)

func installerBootstrapLockName(realmKey string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(realmKey))
	return fmt.Sprintf("inst-boot-%08x", h.Sum32())
}

type ServiceAccountsApiService struct {
	authzClient          auth_client.IAuthorizationClient
	iamClients           iamServices.ClientService
	iamAuth              iamServices.AuthService
	realmDal             business.RealmDal
	authManager          authBusiness.AuthManagerInterface
	serviceAccountsCache internalCache.ServiceAccountsCache
}

func NewServiceAccountsApiService(iamServices *iam.IamServiceContainer, authManager authBusiness.AuthManagerInterface, realmDal business.RealmDal) *ServiceAccountsApiService {
	client := auth_client.NewAuthorizationClient()
	return &ServiceAccountsApiService{
		authzClient:          client,
		iamClients:           (*iamServices).ClientService,
		iamAuth:              (*iamServices).AuthService,
		authManager:          authManager,
		realmDal:             realmDal,
		serviceAccountsCache: internalCache.NewServiceAccountsCacheManager(),
	}
}

func (s *ServiceAccountsApiService) CreateServiceAccount(ctx context.Context, serviceAccountCreationRequest api.ServiceAccountCreationRequest) (api.ImplResponse, error) {
	name := serviceAccountCreationRequest.Name
	createdBy := runaictx.GetString(ctx, runaictx.CfSubjectID)
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	baseErrMsg := fmt.Sprintf(`failed creating service account with name: %s`, name)

	serviceAccountClient := &iamModels.Client{
		Name:     &name,
		ClientID: &name,
		Enabled:  ptr.To(true),
		TenantId: tenantId,
		Attributes: &iamModels.ClientAttributes{
			CreatedBy: createdBy,
		},
	}
	realm_client.ApplyClientProfile(serviceAccountClient, realm_client.ClientTypeTenantApp)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeServiceAccount).WithEntityName(name)
	serviceAccountId, err := s.iamClients.CreateClient(ctx, realm, serviceAccountClient)
	if err != nil {
		return utils.HandleError(ctx, utils.ServiceAccount, utils.Name, name, err, baseErrMsg)
	}
	defer s.deleteServiceAccountsFromCache(ctx, realm)
	serviceAccountResp, err := s.iamClients.GetClient(ctx, realm, serviceAccountId)
	if err != nil {
		errMsg := fmt.Sprintf(`%s. service account not found`, baseErrMsg)
		return utils.HandleError(ctx, utils.ServiceAccount, utils.Name, name, err, errMsg)
	}
	serviceAccountCreationResponse := buildServiceAccountResponse(serviceAccountResp)
	return api.Response(http.StatusCreated, serviceAccountCreationResponse), nil
}

func (s *ServiceAccountsApiService) CreateServiceAccountInternally(ctx context.Context, internalServiceAccountCreationRequest api.InternalServiceAccountCreationRequest) (api.ImplResponse, error) {
	ctx = context.WithValue(ctx, utils.AuthRealmKey, internalServiceAccountCreationRequest.Realm)
	return s.CreateServiceAccount(ctx, api.ServiceAccountCreationRequest{
		Name: internalServiceAccountCreationRequest.Name,
	})
}

func (s *ServiceAccountsApiService) ClusterDeleteCallback(ctx context.Context, event *async_events.ClusterDelete) error {
	tenantID := runaictx.GetString(ctx, runaictx.CfTenantId)
	clusterID := event.ClusterId
	if tenantID == nil {
		return fmt.Errorf("delete cluster event callback for cluster id: %s failed. invalid tenant", clusterID)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("delete cluster event callback called for cluster with id: %s in tenant: %s", clusterID, *tenantID)
	realm, err := s.realmDal.GetRealmByTenantId(ctx, *tenantID)
	if err != nil {
		return fmt.Errorf("failed getting realm name for tenant :%s on delete cluster request for cluster id: %s. err: %v", *tenantID, clusterID, err.Error())
	}
	return s.deleteClusterClients(ctx, realm.Realm, *tenantID, clusterID)
}

func (s *ServiceAccountsApiService) deleteClusterClients(ctx context.Context, realmName, tenantID, clusterID string) error {
	defer s.deleteServiceAccountsFromCache(ctx, realmName)
	clients, err := s.iamClients.GetClients(ctx, realmName, &iamModels.ClientFilter{ClusterId: &clusterID})
	if err != nil {
		return fmt.Errorf("ClusterDeleteCallback failed. failed getting internal service accounts in realm: %s cluster id: %s. err: %v", realmName, clusterID, err.Error())
	}
	for _, client := range clients {
		err := s.iamClients.DeleteClient(ctx, realmName, *client.ID)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed deleting cluster service account: %s, in realm: %s. err: %v", *client.ClientID, realmName, err.Error())
			return fmt.Errorf("error on deleting service account in realm: %s, for tenant %s", realmName, tenantID)
		}
	}
	return nil
}

func (s *ServiceAccountsApiService) DeleteServiceAccountById(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeServiceAccount).WithEntityId(id)
	baseErrMsg := fmt.Sprintf(`failed deleting service account with id: %s`, id)

	serviceAccountClient, err := s.iamClients.GetClient(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`%s. failed to get service account by id`, baseErrMsg)
		return utils.HandleError(ctx, utils.ServiceAccount, utils.ID, id, err, errMsg)
	}
	defer s.deleteServiceAccountsFromCache(ctx, realm)
	if serviceAccountClient.Name == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Can't delete access rules for service account with id: %s. Missing service account name", id)
	} else {
		serviceAccountName := *serviceAccountClient.Name
		audit.AuditFromCtx(ctx).WithEntityName(serviceAccountName)
		err = s.deleteServiceAccountAccessRules(ctx, serviceAccountName)
		if err != nil {
			errMsg := fmt.Sprintf(`%s. failed deleting access rules`, baseErrMsg)
			return utils.HandleError(ctx, utils.ServiceAccount, utils.ID, id, err, errMsg)
		}
	}
	err = s.iamClients.DeleteClient(ctx, realm, id)
	if err != nil {
		return utils.HandleError(ctx, utils.ServiceAccount, utils.ID, id, err, baseErrMsg)
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *ServiceAccountsApiService) GetServiceAccountById(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	serviceAccountClient, err := s.iamClients.GetClient(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting service account with id: %s`, id)
		return utils.HandleError(ctx, utils.ServiceAccount, utils.ID, id, err, errMsg)
	}
	serviceAccount := serviceAccountModelToApiResponse(ctx, serviceAccountClient)
	return api.Response(http.StatusOK, serviceAccount), nil
}

func (s *ServiceAccountsApiService) GetServiceAccounts(ctx context.Context) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	tenantServiceAccounts, err := s.serviceAccountsCache.GetTenantServiceAccountsForRealm(ctx, realm, func() (*internalCache.TenantServiceAccounts, error) {
		serviceAccounts, err := s.iamClients.GetClients(ctx, realm, &iamModels.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeTenantApp)})
		if err != nil {
			return nil, err
		}
		return &internalCache.TenantServiceAccounts{ServiceAccounts: serviceAccounts}, nil
	})
	if err != nil {
		errMsg := `failed getting service accounts`
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s. err: %s", errMsg, err.Error())
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: errMsg}
	}
	serviceAccounts := tenantServiceAccounts.ServiceAccounts
	serviceAccountsArray := []api.ServiceAccount{}
	for i := range serviceAccounts {
		serviceAccount := serviceAccountModelToApiResponse(ctx, &serviceAccounts[i])
		serviceAccountsArray = append(serviceAccountsArray, serviceAccount)
	}
	return api.Response(http.StatusOK, serviceAccountsArray), nil
}

func (s *ServiceAccountsApiService) UpdateServiceAccountById(ctx context.Context, id string, serviceAccountPatchRequest api.ServiceAccountPatchRequest) (api.ImplResponse, error) {
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeServiceAccount).WithEntityId(id)
	isEnabled := serviceAccountPatchRequest.Enabled
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	err := s.iamClients.UpdateClientStatus(ctx, realm, id, isEnabled)
	if err != nil {
		errMsg := fmt.Sprintf(`failed updating service account with id: %s`, id)
		return utils.HandleError(ctx, utils.ServiceAccount, utils.Name, id, err, errMsg)
	}
	s.deleteServiceAccountsFromCache(ctx, realm)
	return api.Response(http.StatusOK, nil), nil
}

func (s *ServiceAccountsApiService) RegenerateServiceAccountSecret(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeServiceAccount).WithEntityId(id)

	clientID, implResp, err := s.validatePrivilegeEscalation(ctx, realm, id)
	if err != nil {
		return implResp, err
	}

	secret, err := s.iamClients.RegenerateClientSecret(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed regenerating secret for service account with id: %s`, id)
		return utils.HandleError(ctx, utils.ServiceAccount, utils.Name, id, err, errMsg)
	}
	s.deleteServiceAccountsFromCache(ctx, realm)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully regenerate service account secret for service account with id: %s (clientID: %s)", id, *clientID)
	return api.Response(http.StatusOK, api.RegenerateServiceAccountSecretResponse{ClientSecret: secret}), nil
}

func (s *ServiceAccountsApiService) GetInstallerServiceAccount(ctx context.Context, tenantId string) (api.ImplResponse, error) {
	clientId := utils.RunaiInstallerClientId
	realmKey, err := s.getRealmKey(ctx, tenantId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "unauthorized") {
			return api.Response(http.StatusUnauthorized, err.Error()), utils.Error{Code: http.StatusUnauthorized, Message: `failed getting installer service account`}
		}
		return utils.HandleError(ctx, utils.ServiceAccount, utils.ClientId, clientId, err, `failed getting installer service account`)
	}
	serviceAccountClient, err := s.iamClients.GetClientByClientId(ctx, realmKey, clientId)
	if err != nil {
		errMsg := `failed getting installer service account`
		return utils.HandleError(ctx, utils.ServiceAccount, utils.ClientId, clientId, err, errMsg)
	}
	if serviceAccountClient == nil {
		return api.Response(http.StatusNotFound, nil), utils.Error{Code: http.StatusNotFound, Message: fmt.Sprintf(`installer service account with clientId: %s not found`, clientId)}
	}
	serviceAccount := buildServiceAccountResponse(serviceAccountClient)
	return api.Response(http.StatusOK, serviceAccount), nil
}

func (s *ServiceAccountsApiService) MintInstallerBootstrapToken(ctx context.Context, clusterId string, tenantId string) (api.ImplResponse, error) {
	realmKey, err := s.getRealmKey(ctx, tenantId)
	if err != nil {
		if strings.HasPrefix(err.Error(), "unauthorized") {
			return api.Response(http.StatusUnauthorized, err.Error()), utils.Error{Code: http.StatusUnauthorized, Message: "unauthorized to mint installer bootstrap token"}
		}
		return utils.HandleError(ctx, utils.ServiceAccount, utils.ClientId, clusterId, err, "failed resolving realm for installer bootstrap token")
	}

	lockKey := realmKey + ":" + clusterId

	rotationSafe := cache.IsInitialized()
	if rotationSafe {
		installerLock := cache.Lock(installerBootstrapLockName(lockKey))
		if err := installerLock.Lock(ctx, installerBootstrapLockTTL); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("MintInstallerBootstrapToken: failed to acquire installer secret rotation lock in realm %s: %v", realmKey, err)
			return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to mint bootstrap token"}
		}
		defer func() {
			unlockCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), installerBootstrapUnlockTimeout)
			defer cancel()
			if unlockErr := installerLock.Unlock(unlockCtx); unlockErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("MintInstallerBootstrapToken: failed to release installer secret rotation lock in realm %s: %v", realmKey, unlockErr)
			}
		}()
	}

	clientId, err := s.getOrCreateInstallerClient(ctx, realmKey, tenantId, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("MintInstallerBootstrapToken: failed to resolve installer client for cluster %s in realm %s: %v", clusterId, realmKey, err)
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to resolve installer client"}
	}

	client, err := s.iamClients.GetClientByClientId(ctx, realmKey, clientId)
	if err != nil || client == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("MintInstallerBootstrapToken: failed to get installer client in realm %s: %v", realmKey, err)
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to get installer client"}
	}
	if client.ClientID == nil || client.ID == nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "installer client is missing required fields"}
	}

	var secret string
	if rotationSafe {
		secret, err = s.iamClients.RegenerateClientSecret(ctx, realmKey, *client.ID)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("MintInstallerBootstrapToken: failed to rotate installer client secret in realm %s: %v", realmKey, err)
			return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to rotate installer client secret"}
		}
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
			"MintInstallerBootstrapToken: skipping secret rotation in realm %s, cache not initialized so rotation can't be safely serialized",
			realmKey,
		)
		secret, err = s.iamClients.GetClientSecret(ctx, realmKey, *client.ID)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("MintInstallerBootstrapToken: failed to get installer client secret in realm %s: %v", realmKey, err)
			return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to mint bootstrap token"}
		}
	}

	tokens, err := s.iamAuth.AuthenticateClient(ctx, realmKey, *client.ClientID, secret)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("MintInstallerBootstrapToken: failed to authenticate installer client in realm %s: %v", realmKey, err)
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "failed to mint bootstrap token"}
	}

	return api.Response(http.StatusOK, api.InstallerBootstrapResponse{Token: tokens.AccessToken}), nil
}

func (s *ServiceAccountsApiService) getOrCreateInstallerClient(ctx context.Context, realm string, tenantId string, clusterId string) (string, error) {
	grantTenantId := tenantId
	if grantTenantId == "" {
		if ctxTenantId := runaictx.GetString(ctx, runaictx.CfTenantId); ctxTenantId != nil {
			grantTenantId = *ctxTenantId
		}
	}

	clients, err := s.iamClients.GetClients(ctx, realm, &iamModels.ClientFilter{
		ClusterId:  &clusterId,
		ClientType: ptr.To(realm_client.ClientTypeInstaller),
	})
	if err != nil {
		return "", fmt.Errorf("failed to look up installer client for cluster %s: %w", clusterId, err)
	}

	clientId := ""
	if len(clients) > 0 {
		if clients[0].ClientID == nil {
			return "", fmt.Errorf("installer client for cluster %s is missing its clientId", clusterId)
		}
		clientId = *clients[0].ClientID
	} else {
		clientId = fmt.Sprintf("runai-internal.%s.installer", clusterId)
		newClient := &iamModels.Client{
			ClientID: ptr.To(clientId),
			Name:     ptr.To(clientId),
			Enabled:  ptr.To(true),
			TenantId: ptr.To(grantTenantId),
			Attributes: &iamModels.ClientAttributes{
				ClusterId: ptr.To(clusterId),
			},
		}
		realm_client.ApplyClientProfile(newClient, realm_client.ClientTypeInstaller)

		if _, err := s.iamClients.CreateClient(ctx, realm, newClient); err != nil && !errors.Is(err, iamError.ErrConflict) {
			return "", fmt.Errorf("failed to create installer client for cluster %s: %w", clusterId, err)
		}
	}

	scope := authorization_api.InternalCreateAccessRuleRequestScope{
		Id:   clusterId,
		Type: authorization_api.SCOPETYPE_CLUSTER,
	}
	if err := s.authzClient.SetServiceAccountPermissions(ctx, grantTenantId, clientId, scope, authorization_api.ALLOWEDINTERNALCREATEROLES_INSTALLER_SERVICE_ACCOUNT); err != nil {
		return "", fmt.Errorf("failed to grant access rule for installer client %s: %w", clientId, err)
	}

	return clientId, nil
}

func (s *ServiceAccountsApiService) deleteServiceAccountAccessRules(ctx context.Context, serviceAccountName string) error {
	request := s.authzClient.DeleteSubjectAccessRules(ctx)
	request = request.SubjectType(string(authorization_api.SUBJECTTYPE_SERVICE_ACCOUNT)).SubjectId(serviceAccountName)
	_, err := s.authzClient.DeleteSubjectAccessRulesExecute(request)
	if err != nil {
		return err
	}
	return nil
}

func serviceAccountModelToApiResponse(ctx context.Context, serviceAccount *iamModels.Client) api.ServiceAccount {
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	var creationTime *time.Time
	var lastUpdateTime *time.Time
	var lastLoginTime *time.Time
	var createdBy = ""
	if serviceAccount.Attributes != nil {
		creationTime = utils.FormatDate(ctx, serviceAccount.Attributes.CreatedAt)
		lastUpdateTime = utils.FormatDate(ctx, serviceAccount.Attributes.UpdatedAt)
		lastLoginTime = utils.FormatDate(ctx, serviceAccount.Attributes.LastLogin)
		if serviceAccount.Attributes.CreatedBy != nil {
			createdBy = *serviceAccount.Attributes.CreatedBy
		}
	}
	return api.ServiceAccount{
		Name:      *serviceAccount.Name,
		Id:        *serviceAccount.ID,
		ClientId:  *serviceAccount.ClientID,
		Enabled:   *serviceAccount.Enabled,
		TenantId:  *tenantId,
		CreatedBy: createdBy,
		CreatedAt: creationTime,
		UpdatedAt: lastUpdateTime,
		LastLogin: lastLoginTime,
	}
}

func buildServiceAccountResponse(serviceAccount *iamModels.Client) api.ServiceAccountPostResponse {
	if serviceAccount == nil {
		logging.Warnf("buildServiceAccountResponse: service account is nil, cannot build service account response")
		return api.ServiceAccountPostResponse{}
	}
	return api.ServiceAccountPostResponse{
		Id:           *serviceAccount.ID,
		Name:         *serviceAccount.Name,
		ClientSecret: *serviceAccount.Secret,
		ClientId:     *serviceAccount.ClientID,
	}
}

func (s *ServiceAccountsApiService) validatePrivilegeEscalation(ctx context.Context, realm, serviceAccountID string) (*string, api.ImplResponse, error) {
	serviceAccount, err := s.iamClients.GetClient(ctx, realm, serviceAccountID)
	if err != nil {
		errMsg := "failed getting service account for privilege escalation check"
		implResp, utilErr := utils.HandleError(ctx, utils.ServiceAccount, utils.ID, serviceAccountID, err, errMsg)
		return nil, implResp, utilErr
	}

	if serviceAccount.ClientID == nil {
		errMsg := "Service account client ID is required for privilege escalation check"
		implResp, utilErr := utils.HandleError(ctx, utils.ServiceAccount, utils.ID, serviceAccountID, utils.ErrInvalidInput, errMsg)
		return nil, implResp, utilErr
	}

	clientID := *serviceAccount.ClientID
	err = s.authManager.ValidatePrivilegeEscalation(ctx, s.authzClient, clientID)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Privilege escalation check failed for %s %s (subjectId: %s): %v", utils.ServiceAccount, serviceAccountID, clientID, err)

		if utils.IsForbidden(err) {
			errMsg := "You're not authorized to regenerate client secret for the selected service account"
			implResp, utilErr := utils.HandleError(ctx, utils.ServiceAccount, utils.ID, serviceAccountID, utils.ErrForbidden, errMsg)
			return nil, implResp, utilErr
		}

		errMsg := "Failed to validate permissions for this service account"
		implResp, utilErr := utils.HandleError(ctx, utils.ServiceAccount, utils.ID, serviceAccountID, err, errMsg)
		return nil, implResp, utilErr
	}

	return &clientID, api.ImplResponse{}, nil
}

func (s *ServiceAccountsApiService) getRealmKey(ctx context.Context, tenantId string) (string, error) {
	if tenantId == "" {
		return utils.GetStringFromCtx(ctx, utils.AuthRealmKey), nil
	}
	if !authorization_utils.HasSystemPermissions(ctx) {
		return "", errors.New("unauthorized attempt to call GetInstallerServiceAccount with tenantId")
	}
	realm, err := s.realmDal.GetRealmByTenantId(ctx, tenantId)
	if err == nil && realm != nil {
		return realm.Realm, nil
	}
	return "", err
}

func (s *ServiceAccountsApiService) deleteServiceAccountsFromCache(ctx context.Context, realm string) {
	if err := s.serviceAccountsCache.DeleteTenantServiceAccountsFromRealm(ctx, realm); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to delete service accounts from cache in realm %s. err: %v", realm, err)
	}
}
