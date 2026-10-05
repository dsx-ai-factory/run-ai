/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/builders"
	tm_config "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/domain/models"
	tenant_lifecycle "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/pubsub/tenant-lifecycle"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
	"github.com/google/uuid"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	identity_manager_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

const (
	pollInterval = 2 * time.Second
	pollTimeout  = 5 * time.Minute
)

type ProvisionConfig struct {
	Email       string
	Password    *string
	Role        authorization_api.AllowedInternalCreateRoles
	MetricStore *models.MetricStoreDetails
}

type TenantProvider interface {
	CreateTenant(ctx context.Context, tenantCreationRequest api.TenantCreationRequest) (*api.TenantCreationResponse, *api.Error)
	DeleteTenantById(ctx context.Context, tenantId int32, isHardDelete bool) (*string, *api.Error)
	GetTenantById(ctx context.Context, tenantId int32) (*models.Tenant, *api.Error)
	GetAllTenants(ctx context.Context, permittedScopes *authorization_api.PermittedScopes, filters *dao.TenantFilters) (*[]models.Tenant, *api.Error)
	GetTenantByName(ctx context.Context, name string) (*models.Tenant, *api.Error)
	UpdateTenantAlias(ctx context.Context, tenantId int32, alias *string) (*models.Tenant, *api.Error)
	RequeueProvision(ctx context.Context, tenant *models.Tenant, cfg ProvisionConfig) *api.Error
}

type TenantProviderImp struct {
	tenantDao      dao.Tenant
	idpClient      identity_manager_client.IIdm
	builder        builders.Interface
	serviceUtils   utils.ErrorUtilser
	domainTemplate string
}

func NewTenantProvider(tenantDao dao.Tenant, idpClient identity_manager_client.IIdm, builder builders.Interface) TenantProvider {
	tc := config.Get[*tm_config.Tenant]()
	return &TenantProviderImp{
		tenantDao:      tenantDao,
		idpClient:      idpClient,
		builder:        builder,
		serviceUtils:   utils.NewErrorUtils(),
		domainTemplate: tc.DomainTemplate,
	}
}

// CreateTenant implements TenantProvider.
func (t *TenantProviderImp) CreateTenant(ctx context.Context, tenantCreationRequest api.TenantCreationRequest) (*api.TenantCreationResponse, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Create tenant request. name: %s, email: %s", tenantCreationRequest.Name, tenantCreationRequest.Email)

	if tenantCreationRequest.Password != nil && *tenantCreationRequest.Password != "" {
		if err := utils.ValidatePassword(*tenantCreationRequest.Password); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Password validation failed: %v", err)
			return nil, &api.Error{Code: http.StatusBadRequest, Message: err.Error()}
		}
	}

	tenant := t.getTenantModelFromRequest(tenantCreationRequest)

	if apiErr := validateNameAndAliasDiffer(ctx, tenant.Name, tenant.Alias); apiErr != nil {
		return nil, apiErr
	}

	if apiErr := t.validateTenantNameAndAliasUniqueness(ctx, tenant.Name, tenant.Alias); apiErr != nil {
		return nil, apiErr
	}
	firstUserRole, err := t.GetFirstUserRole(ctx, tenantCreationRequest.Role)
	if firstUserRole == nil || err != nil {
		errorResponse := t.serviceUtils.ErrorToApiError(err, http.StatusBadRequest)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("First user role %v is not supported", tenantCreationRequest.Role)
		return nil, errorResponse
	}
	tenantData, actionError := t.tenantDao.Create(ctx, tenant)
	if actionError != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(actionError)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error creating tenant %s in database, %s, %v", tenant.DisplayName, errorResponse.Message, actionError.Err)
		return nil, errorResponse
	}

	customerPassword := t.resolveCustomerPassword(tenantCreationRequest)
	additionalData, buildErr := t.buildAdditionalData(tenantCreationRequest.Email, customerPassword, tenantData.Name)
	if buildErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Error building additional data for tenant %s: %v", tenantCreationRequest.Name, buildErr)
	}

	response := api.TenantCreationResponse{
		Tenant:         tenantData.ToApiModel(ctx),
		AdditionalData: &additionalData,
	}

	metricStoreDetails := (*models.MetricStoreDetails)(tenantCreationRequest.MetricStore)
	tenantBuilderInfo := builders.TenantInfo{
		Name:         tenantData.Name,
		Id:           tenantData.ID,
		MimirId:      tenantData.MimirId,
		Email:        tenantCreationRequest.Email,
		Password:     &customerPassword,
		ContractType: &tenantCreationRequest.ContractType,
		Role:         *firstUserRole,
		MetricStore:  metricStoreDetails,
	}

	provisionEvent := &tenant_lifecycle.TenantProvision{
		TenantId:     tenantBuilderInfo.Id,
		TenantName:   tenantBuilderInfo.Name,
		MimirId:      tenantBuilderInfo.MimirId,
		Email:        tenantBuilderInfo.Email,
		Password:     tenantBuilderInfo.Password,
		ContractType: tenantBuilderInfo.ContractType,
		Role:         tenantBuilderInfo.Role,
		MetricStore:  tenantBuilderInfo.MetricStore,
	}
	if err := runaipubsub.Publish[tenant_lifecycle.TenantProvision](ctx, provisionEvent); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to publish TenantProvision for tenant %s: %v", tenant.DisplayName, err)
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if rollbackErr := t.tenantDao.HardDeleteById(cleanupCtx, tenantBuilderInfo.Id); rollbackErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to rollback tenant %s after publish failure: %v", tenant.DisplayName, rollbackErr)
		}
		return nil, &api.Error{Code: http.StatusInternalServerError, Message: "failed to enqueue tenant provisioning; cleanup attempted"}
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Tenant %s creation accepted, TenantProvision event published", tenant.DisplayName)
	return &response, nil
}

func (t *TenantProviderImp) resolveCustomerPassword(request api.TenantCreationRequest) string {
	switch {
	case request.Password != nil && *request.Password != "":
		return *request.Password
	default:
		return utils.GenerateTemporaryUserPassword()
	}
}

func (t *TenantProviderImp) buildAdditionalData(customerEmail, customerPassword, tenantName string) (map[string]interface{}, error) {
	tenantDomain, err := utils.TemplateBaseUrl(t.domainTemplate, tenantName)
	if err != nil {
		return nil, fmt.Errorf("tenant domain template: %w", err)
	}
	return map[string]interface{}{
		"tenant_domain": tenantDomain,
		customerEmail:   customerPassword,
	}, nil
}

// DeleteTenantByIdInternal implements TenantProvider.
func (t *TenantProviderImp) DeleteTenantById(ctx context.Context, tenantId int32, isHardDelete bool) (*string, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Deleting tenant: %d, isHardDelete: %t", tenantId, isHardDelete)

	var apiErr *api.Error
	tenantModel, daoErr := t.tenantDao.GetByIdIncludeDeleted(ctx, uint(tenantId), isHardDelete)
	if daoErr != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(daoErr)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("error getting tenant with id: %d. skipping delete", tenantId)
		return nil, errorResponse
	}
	if isHardDelete {
		apiErr = t.hardDeleteTenant(ctx, tenantId, tenantModel)
	} else {
		apiErr = t.softDeleteTenant(ctx, tenantId, tenantModel)
	}
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Failed deleting tenant with id %d. err: %v", tenantId, apiErr)
		return nil, apiErr
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("tenant with id %d deleted successfully.", tenantId)
	return &tenantModel.Uuid, nil
}

// GetTenantByIdInternal implements TenantProvider.
func (t *TenantProviderImp) GetTenantById(ctx context.Context, tenantId int32) (*models.Tenant, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get tenant for id %d", tenantId)
	tenantData, daoError := t.tenantDao.GetById(ctx, uint(tenantId))
	if daoError != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(daoError)
		errMsg := fmt.Sprintf("failed to get from db tenant with id: %d. err: %v", tenantId, daoError.Error())
		if errorResponse.Code == http.StatusInternalServerError {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMsg)
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", errMsg)
		}
		return nil, errorResponse
	}
	return tenantData, nil
}

func (t *TenantProviderImp) GetAllTenants(ctx context.Context, permittedScopes *authorization_api.PermittedScopes, filters *dao.TenantFilters) (*[]models.Tenant, *api.Error) {
	deleted := filters != nil && filters.Deleted
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get tenants list with deleted: %t", deleted)

	tenantsData, daoError := t.tenantDao.GetAll(ctx, permittedScopes, filters)
	if daoError != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(daoError)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get tenants list: %s", daoError.Error())
		return nil, errorResponse
	}
	return tenantsData, nil
}

func (t *TenantProviderImp) GetTenantByName(ctx context.Context, name string) (*models.Tenant, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get tenant by name %s", name)
	tenantData, actionError := t.tenantDao.GetByName(ctx, name)
	if actionError != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(actionError)
		if errorResponse != nil {
			return nil, errorResponse
		}
	}
	return tenantData, nil
}

func (t *TenantProviderImp) GetContractType(contractType string) string {
	if contractType == "" {
		return models.ContractTypeNormal
	}
	return contractType
}

func (t *TenantProviderImp) GetFirstUserRole(ctx context.Context, roleName *api.Roles) (*authorization_api.AllowedInternalCreateRoles, error) {
	if roleName == nil || *roleName == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("empty role. setting first user role to 'system admin'")
		return authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR.Ptr(), nil
	}

	switch *roleName {
	case api.ROLES_SYSTEM_ADMINISTRATOR:
		return authorization_api.ALLOWEDINTERNALCREATEROLES_SYSTEM_ADMINISTRATOR.Ptr(), nil
	case api.ROLES_CLOUD_OPERATOR:
		return authorization_api.ALLOWEDINTERNALCREATEROLES_CLOUD_OPERATOR.Ptr(), nil
	}
	return nil, fmt.Errorf("first user role %s is not supported", *roleName)
}

func (t *TenantProviderImp) UpdateTenantAlias(ctx context.Context, tenantId int32, alias *string) (*models.Tenant, *api.Error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Update tenant alias for tenant id %d", tenantId)

	lowercaseAlias := alias
	if alias != nil && *alias != "" {
		lowered := strings.ToLower(*alias)
		lowercaseAlias = &lowered
	}

	existingTenant, daoErr := t.tenantDao.GetById(ctx, uint(tenantId))
	if daoErr != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(daoErr)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant with id: %d. err: %v", tenantId, daoErr.Error())
		return nil, errorResponse
	}

	if apiErr := validateNameAndAliasDiffer(ctx, existingTenant.Name, lowercaseAlias); apiErr != nil {
		return nil, apiErr
	}

	if lowercaseAlias != nil && *lowercaseAlias != "" {
		conflictingTenant, daoErr := t.tenantDao.GetByNameOrAlias(ctx, *lowercaseAlias)
		if daoErr == nil && conflictingTenant != nil && conflictingTenant.ID != existingTenant.ID {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Alias: %s conflicts with existing tenant name or alias", *lowercaseAlias)
			return nil, &api.Error{Code: http.StatusConflict, Message: fmt.Sprintf("tenant with name or alias: '%s' already exists", *lowercaseAlias)}
		}
		if daoErr != nil && daoErr.Reason != dao.NotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error checking alias uniqueness: %v", daoErr.Error())
			return nil, t.serviceUtils.TenantErrorToResponse(daoErr)
		}
	}
	updatedTenant, daoErr := t.tenantDao.UpdateByField(ctx, uint(tenantId), "Alias", lowercaseAlias)
	if daoErr != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(daoErr)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update tenant alias for tenant id: %d. err: %v", tenantId, daoErr.Error())
		return nil, errorResponse
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Tenant alias updated successfully for tenant id %d", tenantId)
	return updatedTenant, nil
}

func (t *TenantProviderImp) hardDeleteTenant(ctx context.Context, tenantId int32, tenantModel *models.Tenant) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("hard deleting tenant by id. tenantId: %d", tenantId)

	if _, daoErr := t.tenantDao.UpdateStatus(ctx, uint(tenantId), models.TenantDeleting); daoErr != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(daoErr)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update tenant %d status to Deleting: %v", tenantId, daoErr.Error())
		return errorResponse
	}

	destructEvent := &tenant_lifecycle.TenantDestruct{
		TenantId:   tenantModel.ID,
		TenantName: tenantModel.Name,
		MimirId:    tenantModel.MimirId,
	}
	if err := runaipubsub.Publish[tenant_lifecycle.TenantDestruct](ctx, destructEvent); err != nil {
		errorResponse := t.serviceUtils.ErrorToApiError(err, http.StatusInternalServerError)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to publish TenantDestruct for tenant %d: %v", tenantId, err)
		return errorResponse
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("TenantDestruct event published for tenant %d, polling until deleted", tenantId)
	return t.pollUntilDeleted(ctx, tenantId)
}

func (t *TenantProviderImp) pollUntilDeleted(ctx context.Context, tenantId int32) *api.Error {
	deadline := time.Now().Add(pollTimeout)
	for {
		select {
		case <-ctx.Done():
			return t.serviceUtils.ErrorToApiError(ctx.Err(), http.StatusGatewayTimeout)
		default:
		}
		if time.Now().After(deadline) {
			return t.serviceUtils.ErrorToApiError(
				fmt.Errorf("timed out waiting for tenant %d to be deleted", tenantId),
				http.StatusGatewayTimeout,
			)
		}
		_, daoErr := t.tenantDao.GetById(ctx, uint(tenantId))
		if daoErr != nil {
			if daoErr.Reason == dao.NotFound {
				return nil
			}
			return t.serviceUtils.ErrorToApiError(daoErr, http.StatusInternalServerError)
		}
		time.Sleep(pollInterval)
	}
}

func (t *TenantProviderImp) softDeleteTenant(ctx context.Context, tenantId int32, tenantModel *models.Tenant) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("soft deleting tenant by id. tenantId: %d", tenantId)

	_, err := t.tenantDao.UpdateStatus(ctx, uint(tenantId), models.TenantDeleting)
	if err != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update tenant status to deleting: %d. err: %v", tenantId, err.Error())
		return errorResponse
	}
	resp, iamErr := t.idpClient.DeleteRealm(ctx, tenantModel.Name, false)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if iamErr != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Realm %s does not exist for tenant id: %d, skipping. %v", tenantModel.Name, tenantId, iamErr)
		} else {
			errorResponse := t.serviceUtils.ErrorToApiError(iamErr, http.StatusInternalServerError)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to disable realm %s for tenant id: %d. err: %v", tenantModel.Name, tenantId, iamErr.Error())
			return errorResponse
		}
	}
	daoErr := t.tenantDao.Delete(ctx, *tenantModel)
	if daoErr != nil {
		errorResponse := t.serviceUtils.TenantErrorToResponse(daoErr)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete from db tenant with id: %d. err: %v", tenantId, daoErr.Error())
		return errorResponse
	}
	return nil
}

func (t *TenantProviderImp) getTenantModelFromRequest(tenantCreationRequest api.TenantCreationRequest) models.Tenant {
	tenantId := uuid.New().String()
	tenant := models.Tenant{
		Uuid:         tenantId,
		MimirId:      tenantId,
		DisplayName:  tenantCreationRequest.Name,
		Name:         utils.DisplayNameToRealm(tenantCreationRequest.Name),
		ContractType: t.GetContractType(tenantCreationRequest.ContractType),
		Status:       models.TenantCreating,
		Smg:          true,
	}
	if tenantCreationRequest.Alias != nil && *tenantCreationRequest.Alias != "" {
		lowercaseAlias := strings.ToLower(*tenantCreationRequest.Alias)
		tenant.Alias = &lowercaseAlias
	}
	return tenant
}

func validateNameAndAliasDiffer(ctx context.Context, name string, alias *string) *api.Error {
	if alias != nil && *alias != "" && name == *alias {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Tenant name and alias must not be the same: %s", name)
		return &api.Error{Code: http.StatusBadRequest, Message: "tenant name and alias must not be the same"}
	}
	return nil
}

func (t *TenantProviderImp) RequeueProvision(ctx context.Context, tenant *models.Tenant, cfg ProvisionConfig) *api.Error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Requeuing TenantProvision for failed tenant %s (id: %d)", tenant.Name, tenant.ID)

	if _, daoErr := t.tenantDao.UpdateStatus(ctx, tenant.ID, models.TenantCreating); daoErr != nil {
		errMsg := fmt.Sprintf("failed to reset tenant %s to Creating: %v", tenant.Name, daoErr.Error())
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMsg)
		return &api.Error{Code: http.StatusInternalServerError, Message: errMsg}
	}

	provisionEvent := &tenant_lifecycle.TenantProvision{
		TenantId:     tenant.ID,
		TenantName:   tenant.Name,
		MimirId:      tenant.MimirId,
		Email:        cfg.Email,
		Password:     cfg.Password,
		ContractType: &tenant.ContractType,
		Role:         cfg.Role,
		MetricStore:  cfg.MetricStore,
	}
	if err := runaipubsub.Publish[tenant_lifecycle.TenantProvision](ctx, provisionEvent); err != nil {
		errMsg := fmt.Sprintf("failed to publish TenantProvision for tenant %s: %v", tenant.Name, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errMsg)
		if _, daoErr := t.tenantDao.UpdateStatus(ctx, tenant.ID, models.TenantFailed); daoErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to restore tenant %s to Failed after publish error: %v", tenant.Name, daoErr.Error())
		}
		return &api.Error{Code: http.StatusInternalServerError, Message: errMsg}
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("TenantProvision requeued for tenant %s", tenant.Name)
	return nil
}

func (t *TenantProviderImp) validateTenantNameAndAliasUniqueness(ctx context.Context, name string, alias *string) *api.Error {
	existingTenant, daoErr := t.tenantDao.GetByNameOrAlias(ctx, name)
	if daoErr == nil && existingTenant != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Tenant name: %s conflicts with existing tenant name or alias", name)
		return &api.Error{Code: http.StatusConflict, Message: fmt.Sprintf("tenant with name or alias: '%s' already exists", name)}
	}
	if daoErr != nil && daoErr.Reason != dao.NotFound {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error checking tenant name uniqueness: %v", daoErr.Error())
		return t.serviceUtils.TenantErrorToResponse(daoErr)
	}

	if alias != nil && *alias != "" {
		existingTenant, daoErr = t.tenantDao.GetByNameOrAlias(ctx, *alias)
		if daoErr == nil && existingTenant != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Tenant alias: %s conflicts with existing tenant name or alias", *alias)
			return &api.Error{Code: http.StatusConflict, Message: fmt.Sprintf("tenant with name or alias: '%s' already exists", *alias)}
		}
		if daoErr != nil && daoErr.Reason != dao.NotFound {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error checking tenant alias uniqueness: %v", daoErr.Error())
			return t.serviceUtils.TenantErrorToResponse(daoErr)
		}
	}
	return nil
}
