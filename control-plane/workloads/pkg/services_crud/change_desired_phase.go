package services_crud

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/workloads/pkg/config"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/pkg/batch"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	workloads_mgr_api "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"

	workloads_mgr_client "github.com/run-ai/runai/control-plane/workloads/pkg/clients/workloads_mgr"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_integration"
	"github.com/run-ai/runai/control-plane/workloads/pkg/services_utils"
)

var unstoppablePhases = map[api.Phase]bool{
	api.PHASE_CREATING:    true,
	api.PHASE_STOPPING:    true,
	api.PHASE_DELETING:    true,
	api.PHASE_STOPPED:     true,
	api.PHASE_FAILED:      true,
	api.PHASE_COMPLETED:   true,
	api.PHASE_TERMINATING: true,
}

type WorkloadBatchProcessor struct {
	permittedScopes *authorization_api.PermittedScopes
	desiredPhase    api.WorkloadDesiredPhase
	tenantId        uint
}

func newWorkloadBatchProcessor(permittedScopes *authorization_api.PermittedScopes, desiredPhase api.WorkloadDesiredPhase, tenantId uint) *WorkloadBatchProcessor {
	return &WorkloadBatchProcessor{
		permittedScopes: permittedScopes,
		desiredPhase:    desiredPhase,
		tenantId:        tenantId,
	}
}

func ChangeDesiredStates(ctx context.Context, workloadIds []string,
	desiredPhase api.WorkloadDesiredPhase,
) (api.ImplResponse, error) {
	tenantId, parseErr := services_utils.ParseTenantId(ctx)
	if parseErr != nil {
		return api.Response(int(parseErr.Code), parseErr), nil
	}

	permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	workloadBatchProcessor := newWorkloadBatchProcessor(permittedScopes, desiredPhase, tenantId)
	batchSize := runai_config.Get[*config.WorkloadsBatchConfig]().BatchSize
	return workloadBatchProcessor.processMultipleWorkloads(ctx, workloadIds, batchSize), nil
}

func (wbp *WorkloadBatchProcessor) processMultipleWorkloads(ctx context.Context, workloadIds []string, batchSize int) api.ImplResponse {
	result := openapi.BatchResponse{
		Succeeded: []string{},
		Failed:    []openapi.BatchResponseFailureItem{},
	}

	// Optimization for a single workload
	if len(workloadIds) == 1 {
		workloads, failures := wbp.getWorkloadsWithFailures(ctx, workloadIds)
		if len(workloads) == 0 {
			return api.Response(http.StatusNotFound, &api.Error{
				Code:    http.StatusNotFound,
				Message: fmt.Sprintf("Failed to get workload %v", workloadIds[0]),
				Details: "requested resource does not exist",
			})
		}
		if failures != nil {
			result.Failed = append(result.Failed, failures...)
		} else {
			result = wbp.processWorkloadsBatch(ctx, workloads)
		}

		if len(result.Failed) > 0 {
			return api.Response(int(result.Failed[0].Code), &api.Error{
				Code:    result.Failed[0].Code,
				Message: result.Failed[0].Message,
			})
		}
		return api.Response(http.StatusAccepted, result)
	}

	workloadIdsBatches := batch.BatchIds(workloadIds, batchSize)
	resultChan := make(chan openapi.BatchResponse, len(workloadIdsBatches))
	var wg sync.WaitGroup

	for _, workloadIdsBatch := range workloadIdsBatches {
		wg.Add(1)
		go func(workloadIdsBatch []string) {
			defer wg.Done()

			workloads, failures := wbp.getWorkloadsWithFailures(ctx, workloadIdsBatch)
			if failures != nil {
				result.Failed = append(result.Failed, failures...)
				return
			}

			result.Failed = append(result.Failed, wbp.collectMissingWorkloadsFailures(workloadIdsBatch, workloads)...)

			batchResult := wbp.processWorkloadsBatch(ctx, workloads)
			resultChan <- batchResult
		}(workloadIdsBatch)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(resultChan)

	// Collect results from all goroutines
	for batchResult := range resultChan {
		result.Succeeded = append(result.Succeeded, batchResult.Succeeded...)
		result.Failed = append(result.Failed, batchResult.Failed...)
	}

	result.Id = uuid.NewString()
	return api.Response(http.StatusAccepted, result)
}

func (wbp *WorkloadBatchProcessor) processWorkloadsBatch(ctx context.Context, workloads []*models.WorkloadView) openapi.BatchResponse {
	result := openapi.BatchResponse{
		Succeeded: []string{},
		Failed:    []openapi.BatchResponseFailureItem{},
	}

	genericWorkloads := make(map[string]*models.WorkloadView)

	for _, workload := range workloads {
		isGeneric, err := wbp.processSingleWorkload(ctx, workload)
		if err != nil {
			result.Failed = append(result.Failed, batch.NewBatchResponseFailure(
				workload.ID.String(),
				err.Message,
				err.Code,
			))
		} else if isGeneric {
			genericWorkloads[workload.ID.String()] = workload
		} else {
			result.Succeeded = append(result.Succeeded, workload.ID.String())
		}
	}

	wbp.processGenericWorkloadsBatch(ctx, genericWorkloads, &result)

	return result
}

func (wbp *WorkloadBatchProcessor) processSingleWorkload(ctx context.Context, workload *models.WorkloadView) (bool, *api.Error) {
	isGeneric := convutils.Deref(workload.SourceApi, "") == api.SOURCEAPI_WORKLOADS_V2

	if !isPermittedForWorkload(workload, wbp.permittedScopes) {
		return isGeneric, api.NewError(http.StatusForbidden, auth.ForbiddenError.Message, nil)
	}

	//
	//  check cluster capability according to the workload source (generic/non-generic)
	//
	var capabilityErr *openapi.SubmissionErrorV2
	if isGeneric {
		capabilityErr = services_utils.ValidateClusterSupportsGenericWorkloads(ctx, workload.ClusterID.String())
	}

	if capabilityErr != nil {
		return isGeneric, api.NewError(http.StatusBadRequest, capabilityErr.Message, nil)
	}

	opErr := services_utils.CheckWorkloadOwnershipProtection(ctx, workload.ClusterID.String(), workload.DepartmentID, workload.SubmittedBy)
	if opErr != nil {
		return isGeneric, opErr
	}

	if workload.Type == Distributed {
		if distErr := wbp.validateDistributedFramework(ctx, workload); distErr != nil {
			return isGeneric, distErr
		}
	}

	if wbp.desiredPhase != api.WORKLOADDESIREDPHASE_DELETED {
		if isGeneric {
			if !workload.SupportSuspend {
				return isGeneric, api.NewValidationError("suspend/resume is not supported for this workload type")
			}
			return isGeneric, nil
		}

		actualPhase := workload.Phase
		var validationError *api.Error

		if wbp.desiredPhase == api.WORKLOADDESIREDPHASE_RUNNING && actualPhase != api.PHASE_STOPPED {
			validationError = api.NewValidationError(fmt.Sprintf("resume request is permitted only for stopped workloads (currently the workload is in %s phase)", actualPhase))
		}

		_, unstoppablePhase := unstoppablePhases[actualPhase]

		if wbp.desiredPhase == api.WORKLOADDESIREDPHASE_STOPPED && unstoppablePhase {
			validationError = api.NewValidationError(fmt.Sprintf("suspend request is not permitted for %s workloads", actualPhase))
		}

		if validationError != nil {
			return isGeneric, validationError
		}
	}

	if wbp.desiredPhase == api.WORKLOADDESIREDPHASE_DELETED {
		// When trying to delete a workload that wasn't sync with the cluster, only delete it from the db, without going through the cluster
		if workload.Phase == api.PHASE_FAILED && workload.PhaseMessage == services_utils.WorkloadFailedPhaseMsg_StuckInCreation {
			_, err := services_integration.DbDeleteWorkload(ctx, workload.ID.String())
			if err != nil {
				return isGeneric, api.NewError(http.StatusServiceUnavailable, err.Error(), err)
			}
		}
	}

	// generic records are not handled by workloads service. a batch request for all generic workoads
	// should be sent to workloads-manager service, and only after we can change the phase
	if isGeneric {
		return isGeneric, nil
	}

	updateTime, dbErr := dao_crud.DbUpdateDesiredPhase(ctx, wbp.tenantId, workload.Type, workload.ID.String(), wbp.permittedScopes, wbp.desiredPhase)
	if dbErr != nil {
		if dbErr.Code == http.StatusNotFound {
			updateTime, dbErr = dao_crud.DbAddChangeStateRecord(ctx, workload.Type, workload, wbp.desiredPhase)
		}
		if dbErr != nil {
			return isGeneric, dbErr
		}
	}

	// time should equal the timestamp of update time to the DB
	err := services_integration.UpdateDesiredPhaseOfWorkload(ctx, workload.ID.String(), wbp.desiredPhase, *updateTime)
	if err != nil {
		return isGeneric, err
	}

	return isGeneric, nil
}

func isPermittedForWorkload(workload *models.WorkloadView, permittedScopes *authorization_api.PermittedScopes) bool {
	if convutils.BoolIsTrue(permittedScopes.System) {
		return true
	}

	if datastructs.NewSet(permittedScopes.Tenants...).Has(strconv.Itoa(int(workload.TenantID))) {
		return true
	}

	if datastructs.NewSet(permittedScopes.Clusters...).Has(workload.ClusterID.String()) {
		return true
	}

	if datastructs.NewSet(permittedScopes.Departments...).Has(workload.DepartmentID) {
		return true
	}

	return datastructs.NewSet(permittedScopes.Projects...).Has(workload.ProjectID)
}

func (wbp *WorkloadBatchProcessor) validateDistributedFramework(ctx context.Context, workload *models.WorkloadView) *api.Error {
	var distErr *api.Error

	switch wbp.desiredPhase {
	case api.WORKLOADDESIREDPHASE_DELETED:
		distErr = services_utils.CheckDistributedFrameworkAvailable(ctx, workload.ClusterID.String(), workload.DistributedFramework)
	case api.WORKLOADDESIREDPHASE_STOPPED, api.WORKLOADDESIREDPHASE_RUNNING:
		distErr = services_utils.CheckDistributedFrameworkSuspendable(ctx, workload.ClusterID.String(), workload.DistributedFramework)
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("unsupported desired phase %s", wbp.desiredPhase)
	}

	return distErr
}

// collectMissingWorkloadsFailures returns a list of failures for workloads that were not found in the DB
func (WorkloadBatchProcessor) collectMissingWorkloadsFailures(workloadIds []string, workloads []*models.WorkloadView) []openapi.BatchResponseFailureItem {
	workloadsIdSet := datastructs.NewSet[string]()
	for _, workload := range workloads {
		workloadsIdSet.Add(workload.ID.String())
	}

	var result []openapi.BatchResponseFailureItem
	for _, workloadId := range workloadIds {
		if !workloadsIdSet.Has(workloadId) {
			result = append(result, batch.NewBatchResponseFailure(workloadId, "workload not found", http.StatusNotFound))
		}
	}

	return result
}

func (WorkloadBatchProcessor) getWorkloadsWithFailures(ctx context.Context, workloadIds []string) ([]*models.WorkloadView, []openapi.BatchResponseFailureItem) {
	workloads, getErr := services_integration.GetWorkloadsDetails(ctx, workloadIds, convutils.Ref(true))
	if getErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("failed to get batch of workloads")
		var failures []openapi.BatchResponseFailureItem
		for _, workloadId := range workloadIds {
			failures = append(failures, batch.NewBatchResponseFailure(workloadId, getErr.Message, getErr.Code))
		}
		return nil, failures
	}

	return workloads, nil
}

func (wbp *WorkloadBatchProcessor) processGenericWorkloadsBatch(ctx context.Context, workloads map[string]*models.WorkloadView, result *openapi.BatchResponse) {
	if len(workloads) == 0 {
		return
	}

	//
	//  build the request to workloads manager
	//
	action := workloads_mgr_api.WORKLOADSV2BATCHACTION_DELETE
	switch wbp.desiredPhase {
	case api.WORKLOADDESIREDPHASE_STOPPED:
		action = workloads_mgr_api.WORKLOADSV2BATCHACTION_SUSPEND
	case api.WORKLOADDESIREDPHASE_RUNNING:
		action = workloads_mgr_api.WORKLOADSV2BATCHACTION_RESUME
	}

	workloadsMgrBatchReq := workloads_mgr_api.WorkloadsV2Batch{
		Action: action,
		Ids:    make([]string, 0, len(workloads)),
	}

	for id := range workloads {
		workloadsMgrBatchReq.Ids = append(workloadsMgrBatchReq.Ids, id)
	}

	//
	//  sent it to workloads manager
	//
	resp, err := workloads_mgr_client.Get().BatchWorkloads(ctx, workloadsMgrBatchReq)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to process generic workloads batch: %s: %v", err.Message, err.Details)
		return
	}

	//
	//   add workloads that failed in workloads manager to the list of errors
	//
	for _, failure := range resp.Failed {
		result.Failed = append(result.Failed, openapi.BatchResponseFailureItem{
			Id:      failure.Id,
			Code:    failure.Code,
			Message: failure.Message,
		})
	}

	//
	//   update the succeeded ones in the workloads table only.
	//   Do NOT write to crud_workloads: V2 reconciliation is owned by workloads-manager
	//   and the V2 agent. Creating a crud row would make the legacy agent create/update
	//   a parallel InteractiveWorkload/TrainingWorkload CRD.
	//
	for _, workloadId := range resp.Succeeded {
		workload, found := workloads[workloadId]
		if !found || workload == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("unexpected workload %s reported by workloads manager", workloadId)
			continue
		}

		// Mirror the V1 behaviour: immediately patch the workloads table phase to
		// Stopping/Resuming/Deleting so the UI shows the progress state right away,
		// without waiting for the cluster-side reconciliation to report back.
		updateTime := time.Now()
		if phaseErr := services_integration.UpdateDesiredPhaseOfWorkload(ctx, workloadId, wbp.desiredPhase, updateTime); phaseErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to update intermediate phase for workload %s: %v", workloadId, phaseErr)
		}

		result.Succeeded = append(result.Succeeded, workloadId)
	}
}
