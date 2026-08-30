package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"k8s.io/utils/ptr"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	iamServices "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/services"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const accessKeysLimit = 20

type AccessKeysAPIService struct {
	iamClients iamServices.ClientService
}

func NewAccessKeysAPIService(iamServices *iam.IamServiceContainer) api.AccessKeysAPIServicer {
	return &AccessKeysAPIService{
		iamClients: (*iamServices).ClientService,
	}
}

func (s *AccessKeysAPIService) CreateAccessKey(ctx context.Context, accessKeyCreationRequest api.AccessKeyCreationRequest) (api.ImplResponse, error) {
	name := accessKeyCreationRequest.Name

	createdBy := runaictx.GetString(ctx, runaictx.CfSubjectID)
	tenantId := runaictx.GetString(ctx, runaictx.CfTenantId)
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	filters := getAccessKeyFilter(false, createdBy)

	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeAccessKey).WithEntityName(name)
	accessKeysCount, err := s.iamClients.GetClientsCount(ctx, realm, &filters)
	if err != nil {
		errMsg := fmt.Sprintf(`failed creating access key with name: %s for user: %s. access keys count error.`, name, ptr.Deref(createdBy, "unknown"))
		return utils.HandleError(ctx, utils.AccessKey, utils.Name, name, err, errMsg)
	}
	if !isAccessKeysCountValid(&accessKeysCount) {
		return api.Response(http.StatusBadRequest, nil), utils.Error{Code: http.StatusBadRequest, Message: fmt.Sprintf(`access key limit exceeded. You can only create up to %d access keys`, accessKeysLimit)}
	}

	accessKeyToCreate := &models.Client{
		Name:     &name,
		TenantId: tenantId,
		ClientID: utils.GenerateClientID(),
		Enabled:  ptr.To(true),
		ClientConfig: &models.ClientConfig{
			ServiceAccountsEnabled:    ptr.To(true),
			PublicClient:              ptr.To(false),
			DirectAccessGrantsEnabled: ptr.To(false),
			StandardFlowEnabled:       ptr.To(false),
		},
		Attributes: &models.ClientAttributes{
			CreatedBy:     createdBy,
			ClientType:    ptr.To(realm_client.ClientTypeUserApp),
			TokenLifespan: ptr.To(DefaultServiceAccountTokenLifespan),
		},
	}

	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeAccessKey).WithEntityName(name)
	accessKey, err := s.iamClients.CreateClient(ctx, realm, accessKeyToCreate)
	if err != nil {
		errMsg := fmt.Sprintf(`failed creating access key with name: %s for user: %s`, name, ptr.Deref(createdBy, "unknown"))
		return utils.HandleError(ctx, utils.AccessKey, utils.Name, name, err, errMsg)
	}
	accessKeyResp, err := s.iamClients.GetClient(ctx, realm, accessKey)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting access key by id %s`, accessKey)
		return utils.HandleError(ctx, utils.AccessKey, utils.ID, accessKey, err, errMsg)
	}
	appCreationResponse := buildAccessKeyResponse(ctx, accessKeyResp)
	if appCreationResponse.CreatedAt == nil {
		currentTime := time.Now()
		appCreationResponse.CreatedAt = &currentTime
	}
	return api.Response(http.StatusCreated, appCreationResponse), nil
}

func (s *AccessKeysAPIService) DeleteAccessKeyById(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	createdBy := runaictx.GetString(ctx, runaictx.CfSubjectID)

	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeAccessKey).WithEntityId(id)
	_, err := s.iamClients.GetAccessKeyClientById(ctx, realm, *createdBy, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting access key for deletion. access key id: %s`, id)
		return utils.HandleError(ctx, utils.AccessKey, utils.ID, id, err, errMsg)
	}
	err = s.iamClients.DeleteClient(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed deleting access key by id: %s`, id)
		return utils.HandleError(ctx, utils.AccessKey, utils.ID, id, err, errMsg)
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *AccessKeysAPIService) GetAccessKeyById(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	createdBy := runaictx.GetString(ctx, runaictx.CfSubjectID)

	accessKeyResp, err := s.iamClients.GetAccessKeyClientById(ctx, realm, *createdBy, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting access key by id %s`, id)
		return utils.HandleError(ctx, utils.AccessKey, utils.ID, id, err, errMsg)
	}
	accessKey := accessKeyModelToApiResponse(ctx, accessKeyResp)
	return api.Response(http.StatusOK, accessKey), nil
}

func (s *AccessKeysAPIService) RegenerateAccessKeySecret(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	createdBy := runaictx.GetString(ctx, runaictx.CfSubjectID)

	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeAccessKey).WithEntityId(id)
	_, err := s.iamClients.GetAccessKeyClientById(ctx, realm, *createdBy, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed regenerating secret for access key. failed to get access key by id: %s.`, id)
		return utils.HandleError(ctx, utils.AccessKey, utils.ID, id, err, errMsg)
	}
	secret, err := s.iamClients.RegenerateClientSecret(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed regenerating secret for access key with id: %s`, id)
		return utils.HandleError(ctx, utils.AccessKey, utils.Name, id, err, errMsg)
	}
	return api.Response(http.StatusOK, map[string]string{"clientSecret": secret}), nil
}

func (s *AccessKeysAPIService) GetAccessKeys(ctx context.Context) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	subId := runaictx.GetString(ctx, runaictx.CfSubjectID)
	filters := getAccessKeyFilter(false, subId)

	accessKeys, err := s.iamClients.GetClients(ctx, realm, &filters)
	if err != nil {
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: `failed getting access keys.`}
	}
	accessKeysArray := []api.AccessKey{}
	for i := range accessKeys {
		accessKey := accessKeyModelToApiResponse(ctx, &accessKeys[i])
		accessKeysArray = append(accessKeysArray, accessKey)
	}
	accessKeysResp := api.AccessKeys{
		AccessKeys: accessKeysArray,
	}
	return api.Response(http.StatusOK, accessKeysResp), nil
}

func (s *AccessKeysAPIService) GetAccessKeysAdministration(ctx context.Context, clientId string, createdBy string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	if clientId != "" && createdBy != "" {
		return api.Response(http.StatusBadRequest, "The request support only one filter"), utils.Error{Code: http.StatusBadRequest, Message: ""}
	}
	accessKeys, err := s.getAccessKeys(ctx, realm, clientId, createdBy)
	if err != nil {
		errMsg := fmt.Sprintf(`failed getting access keys for administrator. clientId: %s, createdBy: %s`, clientId, createdBy)
		return utils.HandleError(ctx, utils.AccessKey, utils.Realm, realm, err, errMsg)
	}
	accessKeysArray := []api.AccessKeyAdministration{}
	for i := range accessKeys {
		if isAccessKey(&accessKeys[i]) {
			accessKey := accessKeyModelToApiResponseAdministrator(ctx, &accessKeys[i])
			accessKeysArray = append(accessKeysArray, accessKey)
		}
	}
	accessKeysAdminResp := api.AccessKeysAdministration{
		AccessKeys: accessKeysArray,
	}
	return api.Response(http.StatusOK, accessKeysAdminResp), nil
}

func (s *AccessKeysAPIService) DeleteAccessKeyByIdAdministration(ctx context.Context, id string) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	audit.AuditFromCtx(ctx).WithEntityType(utils.AuditEntityTypeAccessKey).WithEntityId(id)

	accessKey, err := s.iamClients.GetClient(ctx, realm, id)
	if err != nil {
		errMsg := fmt.Sprintf(`failed deleting access key for administrator. failed to get access key by id: %s`, id)
		return utils.HandleError(ctx, utils.AccessKey, utils.ID, id, err, errMsg)
	}
	if isAccessKey(accessKey) {
		err = s.iamClients.DeleteClient(ctx, realm, id)
		if err != nil {
			errMsg := fmt.Sprintf(`failed deleting access key for administrator with access key id: %s`, id)
			return utils.HandleError(ctx, utils.AccessKey, utils.ID, id, err, errMsg)
		}
		return api.Response(http.StatusNoContent, nil), nil
	}
	return api.Response(http.StatusNotFound, nil), utils.Error{Code: http.StatusNotFound, Message: ""}
}

func (s *AccessKeysAPIService) getAccessKeys(ctx context.Context, realm, clientId, createdBy string) ([]models.Client, error) {
	accessKeys := []models.Client{}
	var accessKey *models.Client
	var err error

	if clientId != "" {
		accessKey, err = s.iamClients.GetClientByClientId(ctx, realm, clientId)
		if err != nil {
			return nil, err
		}
		if accessKey == nil {
			return nil, fmt.Errorf("access key with clientId %s not found: %w", clientId, utils.ErrNotFound)
		}
		if !isAccessKey(accessKey) {
			return nil, fmt.Errorf("client with clientId %s is not an access key: %w", clientId, utils.ErrNotFound)
		}
		accessKeys = append(accessKeys, *accessKey)
		return accessKeys, nil
	}
	filters := getAccessKeyFilter(true, &createdBy)
	accessKeysResp, err := s.iamClients.GetClients(ctx, realm, &filters)
	if err != nil {
		return nil, err
	}
	return accessKeysResp, nil
}

func buildAccessKeyResponse(ctx context.Context, accessKey *models.Client) api.AccessKeyPostResponse {
	var creationTime *time.Time
	if accessKey == nil {
		logging.Warnf("buildAccessKeyResponse: access key model is nil, cannot build response")
		return api.AccessKeyPostResponse{}
	}
	if accessKey.Attributes != nil {
		creationTime = utils.FormatDate(ctx, accessKey.Attributes.CreatedAt)
	}
	return api.AccessKeyPostResponse{
		Id:           *accessKey.ID,
		ClientId:     *accessKey.ClientID,
		Name:         *accessKey.Name,
		ClientSecret: *accessKey.Secret,
		CreatedAt:    creationTime,
	}
}

func accessKeyModelToApiResponse(ctx context.Context, accessKey *models.Client) api.AccessKey {
	var creationTime *time.Time
	var lastLoginTime *time.Time

	if accessKey.Attributes != nil {
		creationTime = utils.FormatDate(ctx, accessKey.Attributes.CreatedAt)
		lastLoginTime = utils.FormatDate(ctx, accessKey.Attributes.LastLogin)
	}
	return api.AccessKey{
		Name:      *accessKey.Name,
		Id:        *accessKey.ID,
		ClientId:  *accessKey.ClientID,
		CreatedAt: creationTime,
		LastLogin: lastLoginTime,
	}
}

func accessKeyModelToApiResponseAdministrator(ctx context.Context, accessKey *models.Client) api.AccessKeyAdministration {
	var creationTime *time.Time
	var lastLoginTime *time.Time

	if accessKey.Attributes != nil {
		creationTime = utils.FormatDate(ctx, accessKey.Attributes.CreatedAt)
		lastLoginTime = utils.FormatDate(ctx, accessKey.Attributes.LastLogin)
	}
	return api.AccessKeyAdministration{
		Name:      *accessKey.Name,
		Id:        *accessKey.ID,
		ClientId:  *accessKey.ClientID,
		CreatedAt: creationTime,
		LastLogin: lastLoginTime,
		CreatedBy: *accessKey.Attributes.CreatedBy,
	}
}

func isAccessKey(app *models.Client) bool {
	return app != nil && app.Attributes != nil && *app.Attributes.ClientType == realm_client.ClientTypeUserApp
}

func getAccessKeyFilter(isTenantAdmin bool, createdBy *string) models.ClientFilter {
	return models.ClientFilter{ClientType: ptr.To(realm_client.ClientTypeUserApp), IsTenantAdmin: isTenantAdmin, CreatedBy: createdBy}
}

func isAccessKeysCountValid(accessKeysCount *int) bool {
	return accessKeysCount != nil && *accessKeysCount < accessKeysLimit
}
