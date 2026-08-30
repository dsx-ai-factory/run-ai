package services

import (
	"context"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type helmReposAPIService struct {
	assetsClient AssetsClient
}

func NewHelmReposAPIService(assetsClient ...AssetsClient) *helmReposAPIService {
	var client AssetsClient
	if len(assetsClient) > 0 {
		client = assetsClient[0]
	}
	return &helmReposAPIService{
		assetsClient: client,
	}
}

func (s *helmReposAPIService) ListHelmRepos(ctx context.Context) (api.ImplResponse, error) {
	helmRepos, apiErr := dao.ListHelmRepos(ctx)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: list helm repos: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	helmRepoList := make([]api.HelmRepo, len(helmRepos))
	for i, helmRepo := range helmRepos {
		helmRepoList[i] = modelToApiHelmRepo(ctx, helmRepo)
	}

	return api.Response(http.StatusOK, api.HelmRepoListResponse{
		HelmRepos: helmRepoList,
	}), nil
}

func (s *helmReposAPIService) GetHelmRepoById(ctx context.Context, helmRepoId string) (api.ImplResponse, error) {
	helmRepo, apiErr := dao.GetHelmRepoById(ctx, helmRepoId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: get helm repo: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	return api.Response(http.StatusOK, modelToApiHelmRepo(ctx, *helmRepo)), nil
}

func (s *helmReposAPIService) CreateHelmRepo(ctx context.Context, request api.HelmRepoCreationRequest) (api.ImplResponse, error) {
	helmRepo, apiErr := s.createHelmRepo(ctx, request, false)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: create helm repo: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	return api.Response(http.StatusOK, helmRepo), nil
}

func (s *helmReposAPIService) UpdateHelmRepo(ctx context.Context, helmRepoId string, request api.HelmRepoUpdateRequest) (api.ImplResponse, error) {
	helmRepo, apiErr := s.updateHelmRepo(ctx, helmRepoId, request)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: update helm repo: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	return api.Response(http.StatusOK, helmRepo), nil
}

func (s *helmReposAPIService) DeleteHelmRepo(ctx context.Context, helmRepoId string) (api.ImplResponse, error) {
	apiErr := s.deleteHelmRepo(ctx, helmRepoId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: delete helm repo: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	return api.Response(http.StatusAccepted, nil), nil
}

func modelToApiHelmRepo(ctx context.Context, model models.HelmRepo) api.HelmRepo {
	var tenantId int32
	if model.TenantId != nil {
		tenantIdInt, _ := strconv.ParseInt(*model.TenantId, 10, 32)
		tenantId = int32(tenantIdInt)
	} else {
		tenantId = auth.TenantIdOf(ctx)
	}

	var description string
	if model.Description != nil {
		description = *model.Description
	}

	return api.HelmRepo{
		Id:          model.ID.String(),
		ScopeType:   api.ScopeType(model.ScopeType),
		ScopeId:     model.ScopeId,
		TenantId:    tenantId,
		Name:        model.Name,
		Description: description,
		Url:         model.Url,
		IsPrivate:   model.IsPrivate,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
		CreatedBy:   model.CreatedBy,
		UpdatedBy:   model.UpdatedBy,
	}
}
