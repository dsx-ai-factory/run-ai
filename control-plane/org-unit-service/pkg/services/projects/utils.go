/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package projects

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"sync"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"k8s.io/utils/ptr"
)

const (
	// MaxBatchPatchProjectRequests limits the number of project patch operations per request
	MaxBatchPatchProjectRequests = 100
)

// dbProjectWithPatchRequests ties a DB project instance with its corresponding patch requests.
type dbProjectWithPatchRequests struct {
	dbProject     *models.Project
	patchRequests []api.NodePoolResourcesPatchRequest
}

// departmentPatchGroup represents a single department batch group
type departmentPatchGroup struct {
	clusterId string
	patches   []dbProjectWithPatchRequests
}

// validateResetResourcesPermitted
// permittedScopes - the permitted scopes of the API request - scopes in which it is allowed to Update projects.
// The func will count all projects in the cluster (and in the parent department if provided),
// and compare to the count of projects (in the same scope) that are allowed by the permittedScopes.
// If the counts are equal - then the user is authorized to reset resources for all projects in the scope,
// else - the user is not authorized.
func validateResetResourcesPermitted(ctx context.Context, tenantId, clusterId, parentId string, permittedScopes *authorization_api.PermittedScopes) *api.Error {
	scopeFilters := []queryapi.Filter{
		{Field: "cluster_id", Operator: "=", Values: []string{clusterId}},
	}

	if parentId != "" {
		scopeFilters = append(scopeFilters, queryapi.Filter{Field: "parent_id", Operator: "=", Values: []string{parentId}})
	}

	permittedCount, dbErr := dao.CountProjects(ctx, tenantId, permittedScopes, scopeFilters, "", models.MetaDataVerbosity)
	if dbErr != nil {
		return api.NewError(http.StatusInternalServerError, "failed counting projects for reset resources authorization check with permitted scopes", dbErr)
	}

	allCount, dbErr := dao.CountProjects(ctx, tenantId, nil, scopeFilters, "", models.MetaDataVerbosity)
	if dbErr != nil {
		return api.NewError(http.StatusInternalServerError, "failed counting projects for reset resources authorization check", dbErr)
	}

	if permittedCount != allCount || allCount == 0 {
		scopeLogStr := "cluster"
		if parentId != "" {
			scopeLogStr = "department"
		}
		return api.NewError(http.StatusForbidden, fmt.Sprintf("not authorized to reset resources for all projects in the %s", scopeLogStr), nil)
	}
	return nil
}

/////////////////// batch patch projects utils ///////////////////

// validateAndFetchProjectsFromDb filters out requests with invalid project IDs (non-integer),
// collects the valid ones and failures for the rest,
// then batch fetches the valid projects from DB and returns them indexed by ID.
func validateAndFetchProjectsFromDb(
	ctx context.Context,
	tenantId string,
	permittedScopes *authorization_api.PermittedScopes,
	request api.BatchPatchNodePoolResourcesRequest,
) (map[string][]api.NodePoolResourcesPatchRequest, []api.FailedNodePoolResourcesPatchRequest, map[string]*models.Project, *api.Error) {
	projectIds := datastructs.Set[string]{}
	patchRequestsWithValidIdPerProject := map[string][]api.NodePoolResourcesPatchRequest{}
	failed := []api.FailedNodePoolResourcesPatchRequest{}

	for _, patchRequest := range request.PatchRequests {
		orgUnitId := patchRequest.OrgUnitId
		_, parseErr := strconv.ParseInt(orgUnitId, 10, 64)
		if parseErr != nil {
			failed = append(failed, api.FailedNodePoolResourcesPatchRequest{
				Payload: patchRequest,
				Code:    http.StatusBadRequest,
				Message: "failed to parse project id",
			})
			continue
		}
		projectIds.Add(orgUnitId)
		_, ok := patchRequestsWithValidIdPerProject[orgUnitId]
		if !ok {
			patchRequestsWithValidIdPerProject[orgUnitId] = []api.NodePoolResourcesPatchRequest{}
		}
		patchRequestsWithValidIdPerProject[orgUnitId] = append(patchRequestsWithValidIdPerProject[orgUnitId], patchRequest)
	}

	dbProjects, err := dao.GetProjectsByIds(ctx, tenantId, nil, projectIds.Keys(), permittedScopes)
	if err != nil {
		return nil, nil, nil, api.NewError(http.StatusInternalServerError, "failed to retrieve projects", err)
	}
	dbProjectById := map[string]*models.Project{}
	for _, p := range dbProjects {
		dbProjectById[strconv.FormatInt(p.ID, 10)] = p
	}

	return patchRequestsWithValidIdPerProject, failed, dbProjectById, nil
}

// buildDepartmentPatchGroups groups requests by department and cluster,
// and returns failures for missing/not-permitted cases.
func buildDepartmentPatchGroups(
	dbProjectById map[string]*models.Project,
	patchRequestsWithValidIdPerProject map[string][]api.NodePoolResourcesPatchRequest,
) (map[int64]*departmentPatchGroup, []api.FailedNodePoolResourcesPatchRequest) {
	departmentIdToPatchGroup := map[int64]*departmentPatchGroup{}
	failed := []api.FailedNodePoolResourcesPatchRequest{}

	for projectId, patchRequests := range patchRequestsWithValidIdPerProject {
		project, ok := dbProjectById[projectId]
		if !ok || project == nil || project.ParentID == nil {
			// not found or not permitted in current scopes
			apiErr := &api.Error{Code: http.StatusNotFound, Message: "failed to retrieve project"}
			appendFailedPatchRequests(patchRequests, apiErr, &failed)
			continue
		}

		departmentId := *project.ParentID
		patchGroup, ok := departmentIdToPatchGroup[departmentId]
		if !ok {
			patchGroup = &departmentPatchGroup{
				clusterId: project.ClusterId.String(),
				patches:   []dbProjectWithPatchRequests{},
			}
			departmentIdToPatchGroup[departmentId] = patchGroup
		}
		patchGroup.patches = append(patchGroup.patches, dbProjectWithPatchRequests{
			dbProject:     project,
			patchRequests: patchRequests,
		})
	}
	return departmentIdToPatchGroup, failed
}

// prepareAndValidatePatchResourcesForDepartment prepares and validates a batch of project resource patches
// for projects that belong to the same department. It performs per-project mapping/validation against nodepools,
// then validates that the entire batch does not cause over-subscription of department (if needed),
// and returns the staged, mapped projects for later commit.
func prepareAndValidatePatchResourcesForDepartment(
	ctx context.Context,
	tenantId string,
	departmentId int64,
	patchGroup *departmentPatchGroup,
	nodePoolsPerCluster *sync.Map,
	clusterClient cluster_service_client.IClusterExternalClient,
) ([]dbProjectWithPatchRequests, []api.FailedNodePoolResourcesPatchRequest, *api.Error) {
	failed := []api.FailedNodePoolResourcesPatchRequest{}

	if len(patchGroup.patches) == 0 {
		return nil, failed, api.NewError(http.StatusBadRequest, "no patch requests provided", nil)
	}

	// Resolve nodePools once for the department's single cluster
	nodePools, apiErr := services.GetNodePoolsForClusterFromServiceOrCache(ctx, clusterClient, patchGroup.clusterId, nodePoolsPerCluster)
	if apiErr != nil {
		return nil, failed, apiErr
	}

	batchUpdateRequests := make(map[int64]services.BatchProjectUpdateRequest)
	stagedGroup := []dbProjectWithPatchRequests{}
	// Prepare per-project data using already-fetched projects
	for _, projectPatches := range patchGroup.patches {
		stagedGroupPatches := []api.NodePoolResourcesPatchRequest{}

		requestedPatchNodePoolResourcesPerProject := []api.ResourcesNullable{}
		for _, patchRequest := range projectPatches.patchRequests {
			resources := services.ConvertNodePoolResourcesPatchRequest(patchRequest,
				services.StoredGpuDeserved(projectPatches.dbProject.OrgUnitData.Resources, patchRequest.NodePoolId))
			err := services.ValidateNodePoolResourcesPatchRequest(ctx, resources, nodePools)
			if err != nil {
				failed = append(failed, api.FailedNodePoolResourcesPatchRequest{
					Payload: patchRequest,
					Code:    http.StatusBadRequest,
					Message: err.Error(),
				})
				continue
			}
			requestedPatchNodePoolResourcesPerProject = append(requestedPatchNodePoolResourcesPerProject, resources)
			stagedGroupPatches = append(stagedGroupPatches, patchRequest)
		}

		project := projectPatches.dbProject
		originalResources := services.CopyResources(project.OrgUnitData.Resources)
		patchedResources, err := services.MapNodePoolPatchResourcesRequestToModel(ctx, requestedPatchNodePoolResourcesPerProject, project.OrgUnitData.Resources, nodePools, stagedGroupPatches)
		if err != nil {
			apiErr = &api.Error{Code: http.StatusBadRequest, Message: err.Error()}
			appendFailedPatchRequests(projectPatches.patchRequests, apiErr, &failed)
			continue
		}
		project.OrgUnitData.Resources = patchedResources

		apiErr = services.ValidateProjectUpdateNotCausingOverSubscriptionOfAllocatedNonPreemtible(ctx, project.OrgUnitData.Resources, *project)
		if apiErr != nil {
			appendFailedPatchRequests(projectPatches.patchRequests, apiErr, &failed)
			continue
		}

		if !resourcesEqual(originalResources, patchedResources) {
			project.ProjectStatus = models.ProjectStatus{
				Phase: ptr.To(UpdatingPhase),
			}
		}

		batchUpdateRequests[project.ID] = services.BatchProjectUpdateRequest{
			DbProject: project,
			Requested: project.OrgUnitData.Resources,
			Original:  originalResources,
		}

		// store the modified project instance with only the valid requests
		stagedGroup = append(stagedGroup, dbProjectWithPatchRequests{
			dbProject:     project,
			patchRequests: stagedGroupPatches,
		})
	}

	// Validate batch against department quotas
	apiErr = services.ValidateBatchProjectUpdateNotCausingOverSubscriptionOfDepartment(ctx, tenantId, departmentId, batchUpdateRequests)
	if apiErr != nil {
		return nil, failed, apiErr
	}

	return stagedGroup, failed, nil
}

// commitProjectPatchesPerDepartmentInParallel runs an all-or-nothing transaction per department group, in parallel.
// Each group's updates are committed atomically; groups run concurrently.
func commitProjectPatchesPerDepartmentInParallel(
	ctx context.Context,
	tenantId string,
	stagedGroupsPerDepartment [][]dbProjectWithPatchRequests,
	subjectId *string,
	nodePoolsPerCluster *sync.Map,
	clusterClient cluster_service_client.IClusterExternalClient,
) ([]api.ProjectNodePoolResources, []api.FailedNodePoolResourcesPatchRequest) {
	numGroups := len(stagedGroupsPerDepartment)
	if numGroups == 0 {
		return []api.ProjectNodePoolResources{}, []api.FailedNodePoolResourcesPatchRequest{}
	}

	numWorkers := min(4*runtime.NumCPU(), numGroups)
	workChan := make(chan []dbProjectWithPatchRequests, numGroups)
	resultChan := make(chan struct{}, numGroups)
	succeededSync := services.SyncSlice[api.ProjectNodePoolResources]{}
	failedSync := services.SyncSlice[api.FailedNodePoolResourcesPatchRequest]{}

	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		for group := range workChan {
			groupSucceeded, apiErr := commitProjectPatchesForDepartmentAtomic(ctx, tenantId, group, subjectId, nodePoolsPerCluster, clusterClient)
			if apiErr != nil {
				for _, staged := range group {
					for _, patchRequest := range staged.patchRequests {
						failedSync.Append(api.FailedNodePoolResourcesPatchRequest{
							Payload: patchRequest,
							Code:    apiErr.Code,
							Message: apiErr.Message,
						})
					}
				}
			} else {
				for _, s := range groupSucceeded {
					succeededSync.Append(s)
				}
			}
			resultChan <- struct{}{}
		}
	}

	// start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker()
	}
	// send work to workers
	go func() {
		for _, group := range stagedGroupsPerDepartment {
			workChan <- group
		}
		close(workChan)
	}()

	// wait for all work to complete
	for range stagedGroupsPerDepartment {
		<-resultChan
	}
	// wait for all workers to finish
	wg.Wait()

	audit.AuditFromCtx(ctx).Drop()

	return (&succeededSync).GetAll(), (&failedSync).GetAll()
}

// commitProjectPatchesForDepartmentAtomic commits all staged updates for a single department in one DB transaction.
// If any update fails, the entire transaction is rolled back.
// NOTE: on success, the the patch requests are reported as succeeded in the return param,
// but if failed - no per-request failure info is returned - the caller must build that.
func commitProjectPatchesForDepartmentAtomic(
	ctx context.Context,
	tenantId string,
	stagedGroup []dbProjectWithPatchRequests,
	subjectId *string,
	nodePoolsPerCluster *sync.Map,
	clusterClient cluster_service_client.IClusterExternalClient,
) ([]api.ProjectNodePoolResources, *api.Error) {
	if len(stagedGroup) == 0 {
		return []api.ProjectNodePoolResources{}, nil
	}

	// All projects in the group share the same cluster (by construction)
	clusterId := stagedGroup[0].dbProject.ClusterId.String()
	nodePools, apiErr := services.GetNodePoolsForClusterFromServiceOrCache(ctx, clusterClient, clusterId, nodePoolsPerCluster)
	if apiErr != nil {
		return nil, apiErr
	}

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	// Apply updates within the transaction
	for _, staged := range stagedGroup {
		if subjectId != nil {
			staged.dbProject.UpdatedBy = *subjectId
		}
		err := dao.UpdateProjectTx(ctx, tx, staged.dbProject, false, false)
		if err != nil {
			tx.Rollback()
			return nil, api.NewError(http.StatusInternalServerError, "Failed to update project", err)
		}
	}

	// Commit the transaction
	err := tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, api.NewError(http.StatusInternalServerError, "Failed to commit transaction", err)
	}
	committed = true

	// Invalidate caches once per cluster (tenant the same across group)
	dao.InvalidateAllCache(ctx, stagedGroup[0].dbProject.TenantId, clusterId)

	// Build success responses and audit
	succeeded := make([]api.ProjectNodePoolResources, 0, len(stagedGroup))
	for _, staged := range stagedGroup {
		for _, patchRequest := range staged.patchRequests {
			nodePoolResources, getErr := dao.GetProjectNodePoolResourcesPerDepartmentNodePool(ctx, tenantId, patchRequest.OrgUnitId, patchRequest.NodePoolId)
			if getErr != nil {
				return nil, api.NewError(http.StatusInternalServerError, "Failed to get updated project node pool resources", getErr)
			}
			result := services.MapModelProjectNodePoolResourcesToApi(*nodePoolResources, nodePools)
			succeeded = append(succeeded, *result)
		}

		projectId := strconv.Itoa(int(staged.dbProject.ID))
		auditLogError := audit.AuditFromAudit(audit.AuditFromCtx(ctx)).
			WithAction("Patch").
			WithHttpMethod(http.MethodPatch).
			WithEntityName(staged.dbProject.Name).
			WithEntityType(AuditEntityType).
			WithEntityId(projectId).
			WithStatusCode(http.StatusOK).
			Write()
		if auditLogError != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to write audit log for patch resources of project %s (id %s): %v",
				staged.dbProject.Name, projectId, auditLogError)
		}
	}

	return succeeded, nil
}

func appendFailedPatchRequests(patchRequests []api.NodePoolResourcesPatchRequest, apiErr *api.Error, failedOutput *[]api.FailedNodePoolResourcesPatchRequest) {
	if apiErr == nil || failedOutput == nil {
		return
	}
	for _, patchRequest := range patchRequests {
		*failedOutput = append(*failedOutput, api.FailedNodePoolResourcesPatchRequest{
			Payload: patchRequest,
			Code:    apiErr.Code,
			Message: apiErr.Message,
		})
	}
}
