package services

import (
	"context"
	"fmt"
	"net/http"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"
)

type RevisionsService struct {
}

func NewRevisionsService() api.RevisionsAPIServicer {
	return &RevisionsService{}
}

func (s *RevisionsService) SyncRevisions(ctx context.Context, revisionsSyncRequest api.RevisionsSyncRequest) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting revisions sync in tenant <%d> and project <%s>", tenantId, revisionsSyncRequest.ProjectId)
	from, to, apiErr := parseFromToUuids(revisionsSyncRequest.From, revisionsSyncRequest.To)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	syncedRevisions, failedRevisions, errResponse := s.syncGroupOfRevisions(ctx, revisionsSyncRequest.Revisions,
		revisionsSyncRequest.ClusterId, revisionsSyncRequest.ProjectId, revisionsSyncRequest.IdsToIgnore, tenantId, to, from)
	if errResponse != nil {
		return *errResponse, nil
	}

	if len(failedRevisions) != 0 {
		return api.Response(http.StatusMultiStatus, api.SyncRevisions207Response{
			SuccessfullySyncedRevisions: syncedRevisions,
			FailedRevisions:             failedRevisions,
		}), nil
	}

	return NoContentResponse(), nil
}

func (s *RevisionsService) syncGroupOfRevisions(ctx context.Context, revisions []api.RevisionSyncFields, clusterId string, projectId string,
	idsToIgnore []string, tenantId uint, to *uuid.UUID, from *uuid.UUID) ([]string, []string, *api.ImplResponse) {

	revisionModels := []models.Revision{}
	children := []models.RevisionChild{}

	for _, revision := range revisions {
		revisionModel, revisionChildren, err := CreationRequestToRevisionModel(ctx, tenantId, projectId, revision)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Revision <%s> is invalid within project with id <%s>. Err: %+v", revision.Name, revision.ProjectId, err)
			return nil, nil, ptr.To(NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("Revision <%s> is invalid in sync request of project <%s>", revision.Name, projectId), err)))
		}

		revisionModels = append(revisionModels, revisionModel)
		children = append(children, revisionChildren...)
	}

	syncedRevisions, failedRevisions, dbErr := dao.DbSyncRevisions(ctx, tenantId, clusterId, projectId, revisionModels, to, from, idsToIgnore)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync revisions of project with id <%s>. Err: %+v", projectId, dbErr)
		return nil, nil, ptr.To(NewErrorResponse(dbErr))
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof(
		"Successfully synced revisions with ids: %v in project <%s>. Revisions excluded from deletion: %v. From: <%s>. To: <%s>",
		syncedRevisions, projectId, idsToIgnore, from, to)

	if len(syncedRevisions) > 0 {
		dbErr = dao.DbUpsertRevisionsChildren(ctx, children, failedRevisions)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync workloads revisions children of project with id <%s>. Err: %+v", projectId, dbErr)
			return nil, nil, ptr.To(NewErrorResponse(dbErr))
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully synced revisions as workload children for revisions with ids: %v", syncedRevisions)
	}
	return syncedRevisions, failedRevisions, nil
}

func (s *RevisionsService) GetRevision(ctx context.Context, revisionId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	revision, dbErr := dao.DbGetRevision(ctx, tenantId, revisionId)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := RevisionModelToApi(revision)
	return api.Response(http.StatusOK, response), nil
}

func (s *RevisionsService) DeleteRevision(ctx context.Context, revisionId string) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Starting revision <%s> delete request in tenant <%d>", revisionId, tenantId)

	id, parseErr := uuid.Parse(revisionId)
	if parseErr != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, fmt.Sprintf("revision Id must be a valid uuid. <%s>", revisionId), parseErr)), nil
	}

	dbErr := dao.DbDeleteRevision(ctx, tenantId, id)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete revision <%s>. Err: %+v", revisionId, dbErr)
		return NewErrorResponse(dbErr), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Successfully deleted revision <%s> in tenant <%d>", revisionId, tenantId)
	return NoContentResponse(), nil
}

func (s *RevisionsService) GetInferenceWorkloadRevisions(ctx context.Context, workloadId string, deleted bool) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workload, apiErr := dao.GetWorkloadViewFromDb(ctx, workloadId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = validateRevisionSupported(ctx, workload.ClusterID)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	revisions, dbErr := dao.DbGetWorkloadRevisions(ctx, tenantId, workloadId, permittedScopes, deleted)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	apiRevisions := []api.Revision{}
	for _, revision := range *revisions {
		apiRevisions = append(apiRevisions, RevisionModelToApi(&revision))
	}

	response := api.GetInferenceWorkloadRevisions200Response{
		Revisions: apiRevisions,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *RevisionsService) CountInferenceWorkloadRevisions(ctx context.Context, workloadId string, deleted bool) (api.ImplResponse, error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	workload, apiErr := dao.GetWorkloadViewFromDb(ctx, workloadId)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = validateRevisionSupported(ctx, workload.ClusterID)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	count, dbErr := dao.DbCountWorkloadRevisions(ctx, tenantId, workloadId, permittedScopes, deleted)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountWorkloads200Response{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}
