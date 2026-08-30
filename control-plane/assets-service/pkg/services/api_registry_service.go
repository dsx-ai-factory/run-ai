package services

import (
	"context"
	"errors"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/scope"

	audit "github.com/run-ai/runai/control-plane/audit-service/pkg/client"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	backend "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

//   ----------- NOTE --------------
//	The registry asset is not a regular asset, since it is not part of the workload creation.
//	We decided to implement it as a "fake" asset inorder to benefit credentials, scopes, and other asset properties
//   _______________________________

// RegistryApiService is a service that implements the logic for the registry service
// This service should implement the business logic for every endpoint for the RegistryApi API.
// Include any external packages or services that will be required by this service.
type RegistryApiService struct {
	httpClient http.Client
}

func NewRegistryApiService() api.RegistryAPIServicer {
	return &RegistryApiService{
		httpClient: http.Client{},
	}
}

func (s *RegistryApiService) CreateRegistry(ctx context.Context, request api.RegistryCreationRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_REGISTRY
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityName(request.Meta.Name).WithClusterId(safeStrPtr(request.Meta.ClusterId))

	if request.Meta.Scope != api.SCOPE_TENANT {
		return NewErrorResponse(api.NewValidationError("Registry is supported for tenant scope only")), nil
	}

	validationError := ValidateAssetCreateRequestMeta(ctx, &request.Meta, assetKind)
	if validationError != nil {
		return NewErrorResponse(validationError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfCreateRequest(&request.Meta), string(assetKind), backend.ACTION_CREATE, convutils.Deref(request.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	err := s.validateRegistryRequest(ctx, request.Spec)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	var activeKeyId int32
	switch request.Spec.CredentialKind {
	case api.CREDENTIALKIND_PASSWORD:
		activeKeyId, err = encryptPassword(ctx, &request.Spec)
		if err != nil {
			return NewErrorResponse(err), nil
		}
	default:
		return NewErrorResponse(api.NewError(http.StatusNotImplemented, "credential type not supported", nil)), nil
	}

	nameValidationQuery, err := scope.GetNameValidationQuery(
		ctx,
		request.Meta.Name,
		request.Meta.Scope,
		request.Meta.ProjectId,
		request.Meta.DepartmentId,
		request.Meta.ClusterId,
		request.Meta.Private,
		api.ASSETKIND_REGISTRY,
		nil,
	)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	registry, err := dao.CreateRegistry(ctx, request, activeKeyId, nameValidationQuery)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	audit.AuditFromCtx(ctx).WithEntityId(registry.Meta.Id)
	return api.Response(http.StatusCreated, registry), nil
}

func (s *RegistryApiService) DeleteRegistry(ctx context.Context, id string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_REGISTRY
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id)

	assetMeta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithEntityName(assetMeta.Name).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_DELETE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	actionError := dao.DeleteRegistry(ctx, id, dao.HardDelete)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return api.Response(http.StatusNoContent, nil), nil
}

func (s *RegistryApiService) GetRegistryById(ctx context.Context, id string) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_REGISTRY

	asset, actionError := dao.GetRegistryById(ctx, id)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfMeta(&asset.Meta), string(assetKind), backend.ACTION_READ, convutils.Deref(asset.Meta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	return api.Response(http.StatusOK, asset), nil
}

func (s *RegistryApiService) ListRegistries(ctx context.Context, nameParam string) (api.ImplResponse, error) {
	queryParams, queryError := query_types.NewQueryParams("", nameParam, "", false, 0, "", "", nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	if _, authError := query_types.ApplyRbacRules(ctx, queryParams); authError != nil {
		return NewErrorResponse(authError), nil
	}

	result, actionError := dao.ListRegistries(ctx, queryParams)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (s *RegistryApiService) UpdateRegistry(ctx context.Context, id string, request api.RegistryUpdateRequest) (api.ImplResponse, error) {
	assetKind := api.ASSETKIND_REGISTRY
	audit.AuditFromCtx(ctx).WithEntityType(string(assetKind)).WithEntityId(id).WithEntityName(request.Meta.Name)

	assetMeta, getError := dao.GetAssetMetaById(ctx, id)
	if getError != nil {
		return NewErrorResponse(getError), nil
	}
	audit.AuditFromCtx(ctx).WithClusterId(safeStrPtr(assetMeta.ClusterId))

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(assetKind), backend.ACTION_UPDATE, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	var activeKeyId int32 = -1

	/* At this point we are not supporting spec update
	   validationError := s.validateRegistryRequest(request.Spec)
	   if validationError != nil {
	   	return NewErrorResponse(validationError), nil
	   }

	   activeKeyId, err = encryptPassword(ctx, &request.Spec)
	   if err != nil {
	   	return NewErrorResponse(err), nil
	   }
	*/

	nameValidationQuery, err := scope.GetNameValidationQuery(ctx, request.Meta.Name, assetMeta.Scope, assetMeta.ProjectId, assetMeta.DepartmentId, assetMeta.ClusterId, assetMeta.Private, api.ASSETKIND_REGISTRY, &id)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	result, actionError := dao.UpdateRegistry(ctx, id, request, activeKeyId, nameValidationQuery)
	if actionError != nil {
		return NewErrorResponse(actionError), nil
	}
	return api.Response(http.StatusOK, result), nil
}

// GetRepositories - Get the repositories in the registry.
func (s *RegistryApiService) GetRepositories(ctx context.Context, id string, searchName string) (api.ImplResponse, error) {
	assetMeta, metaErr := dao.GetAssetMetaById(ctx, id)
	if metaErr != nil {
		return NewErrorResponse(metaErr), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(api.ASSETKIND_REGISTRY), backend.ACTION_READ, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	registry, err := dao.GetRegistryVisiblePassword(ctx, id)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	repositories, err := s.getRepositories(ctx, registry.Spec, searchName)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, repositories), nil
}

// GetRepositoryTags - Get the repositories in the registry.
func (s *RegistryApiService) GetRepositoryTags(ctx context.Context, id string, repository string, searchName string) (api.ImplResponse, error) {
	assetMeta, metaErr := dao.GetAssetMetaById(ctx, id)
	if metaErr != nil {
		return NewErrorResponse(metaErr), nil
	}

	rbacErr := auth.ValidateAccessByActionAndAssetKind(ctx, ScopeInfoOfDbMeta(assetMeta), string(api.ASSETKIND_REGISTRY), backend.ACTION_READ, convutils.Deref(assetMeta.Private, false))
	if rbacErr != nil {
		return NewErrorResponse(api.NewError(int(rbacErr.Code), rbacErr.Message, errors.New(rbacErr.Details))), nil
	}

	registry, err := dao.GetRegistryVisiblePassword(ctx, id)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	tags, err := s.getRepositoryTags(ctx, registry.Spec, repository, searchName)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, tags), nil
}
