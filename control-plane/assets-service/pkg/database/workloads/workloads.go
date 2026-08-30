package workloads

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/jinzhu/copier"

	"github.com/run-ai/runai/runai-application-packages/jsonutils/gojsonq"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

	"github.com/run-ai/runai/runai-application-packages/datastructs"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	ColumnTenantId = "tenant_id"

	annotationClusterId   = "clusterId"
	annotationWorkloadId  = "workloadId"
	annotationWorkspaceId = "workspaceId" // leave this one for backward compatibility
)

func JoinWorkloadsAndWorkloadsAssets() string {
	return fmt.Sprintf("join %s on %s.id = %s.workload_id", dbtables.TableNameWorkloads, dbtables.TableNameWorkloads, dbtables.TableNameWorkloadsAssets)
}

func JoinWorkloadAssetsAndAssetMeta() string {
	return fmt.Sprintf("join %s on %s.id = %s.asset_id", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta, dbtables.TableNameWorkloadsAssets)
}

func CreateWorkload(ctx context.Context, workloadType string, request *api.WorkloadCreationRequest,
	workloadId string, containedAssets []dbmodel.AssetsContainedInWorkload, masterContainedAssets []dbmodel.AssetsContainedInWorkload,
	workloadPayload *jobs.JobWorkload, dataSourceOverridesById map[string]*api.DataSourceOverrides, masterDataSourceOverridesById map[string]*api.DataSourceOverrides,
) (*dbmodel.GenericWorkload, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	workloadsRec := dbmodel.WorkloadRec{
		WorkloadType: workloadType,

		WorkloadMetaV1: api.WorkloadMetaV1{
			Id: workloadId,

			TenantId:  auth.TenantIdOf(ctx),
			ClusterId: request.ClusterId,
			ProjectId: request.ProjectId,
			Name:      request.Name,

			//   at this stage the state is not defined because frontend did not yet create the workload
			//   in the cluster, and it is possible it may never be submitted. in that case we have to
			//   keep this entry in an un-initialized state. once frontend send it, it will change the
			//   desired state to Submitted in case it succeeded to send, or delete the entry in case
			//   it was not able to submit. in fully saas, when frontend will not be in the loop, we
			//   will improve this mechanism
			DesiredState: api.WORKLOADSTATE_EMPTY,
			CurrentState: api.WORKLOADSTATE_EMPTY,

			CreatedBy: auth.SubjectIdOf(ctx),
			UpdatedBy: auth.SubjectIdOf(ctx),
		},

		CreatedAt: time.Now(),

		SpecificEnv: request.SpecificEnv,
	}

	if workloadPayload != nil {
		workloadsRec.Workload = *workloadPayload
	}

	if workloadType == auth.WorkloadTypeDistributed {
		// request.Distributed was validated as not nil (workload builder)
		workloadsRec.DistSpecificEnv = &dbmodel.DistSpecificEnv{
			NumWorkers: request.Distributed.NumWorkers,
			NoMaster:   request.Distributed.NoMaster,
		}

		// handle MasterSpecificRunParams
		if !convutils.Deref(request.Distributed.NoMaster, false) {
			if request.Distributed.Master != nil {
				workloadsRec.DistSpecificEnv.MasterSpecificRunParams = request.Distributed.Master.SpecificEnv
			} else {
				workloadsRec.DistSpecificEnv.MasterSpecificRunParams = request.SpecificEnv
			}
		}

		// handle distFramework
		workloadsRec.DistFramework = request.Distributed.DistFramework

	}
	rowsAffected, err := tx.Create(workloadsRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.AssetKind(workloadType)).WithDetailsFromDbResult(rowsAffected, err)
	}

	for _, asset := range containedAssets {
		workloadAssetsRec := dbmodel.WorkloadsAssetsRec{
			WorkloadId: workloadId,
			AssetId:    asset.Id,
			AssetKind:  asset.Kind,
			AutoDelete: convutils.Deref(asset.AutoDelete, dbmodel.DefaultForAutoDelete),
		}

		assetOverrides, ok := dataSourceOverridesById[asset.Id]
		if ok {
			workloadAssetsRec.Overrides = assetOverrides
		}

		if workloadType == auth.WorkloadTypeDistributed {
			workloadAssetsRec.DistReplicaType = dbmodel.DistReplicaTypeWorker
		}

		rowsAffected, err := tx.Create(workloadAssetsRec)
		if err != nil {
			return nil, database.Error(database.CREATE, api.AssetKind(workloadType)).WithDetailsFromDbResult(rowsAffected, err)
		}
	}

	for _, asset := range masterContainedAssets {
		workloadAssetsRec := dbmodel.WorkloadsAssetsRec{
			WorkloadId:      workloadId,
			AssetId:         asset.Id,
			AssetKind:       asset.Kind,
			AutoDelete:      convutils.Deref(asset.AutoDelete, dbmodel.DefaultForAutoDelete),
			DistReplicaType: dbmodel.DistReplicaTypeMaster,
		}
		assetOverrides, ok := masterDataSourceOverridesById[asset.Id]
		if ok {
			workloadAssetsRec.Overrides = assetOverrides
		}

		rowsAffected, err := tx.Create(workloadAssetsRec)
		if err != nil {
			return nil, database.Error(database.CREATE, api.AssetKind(workloadType)).WithDetailsFromDbResult(rowsAffected, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.AssetKind(workloadType)).WithDetails(err)
	}
	committed = true

	// saving usage info of unique assets by the workload
	// this function only logs errors and we do not check if it returned errors because we dont want it to fail the API call
	_ = updateAssetsLastUsedByWorkload(ctx, workloadType, workloadId, workloadsRec.ProjectId, containedAssets, masterContainedAssets)

	return GetWorkloadById(ctx, workloadType, workloadId)
}

func DeleteWorkload(ctx context.Context, workloadType string, workloadId string) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	// delete workload
	rowsAffected, err := tx.Delete(&dbmodel.WorkloadRec{
		WorkloadMetaV1: api.WorkloadMetaV1{
			Id: workloadId,
		},
	})
	if err != nil {
		return database.Error(database.DELETE, api.AssetKind(workloadType)).WithDetailsFromDbResult(rowsAffected, err)
	}

	err = tx.Commit()
	if err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.AssetKind(workloadType)).WithDetails(err)
	}
	committed = true

	return nil
}

func GetWorkloadById(ctx context.Context, workloadType string, id string) (*dbmodel.GenericWorkload, *api.Error) {
	var workloadTypesArray []string

	if workloadType == auth.WorkloadTypeTraining {
		workloadTypesArray = append(workloadTypesArray, workloadType)
		workloadTypesArray = append(workloadTypesArray, auth.WorkloadTypeDistributed)
	} else {
		workloadTypesArray = append(workloadTypesArray, workloadType)
	}

	//
	//  get the workload assets. it also returns the assets metadata and specific env
	//
	workloadAssets, dbError := GetWorkloadAssetsByWorkloadId(ctx, workloadTypesArray, id)
	if dbError != nil {
		return nil, dbError
	}

	if len(workloadAssets) == 0 {
		return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("cannot find %s %s", workloadType, id), nil)
	}

	//
	//  initialize data sources array. the rest of the payload is self initialized
	//
	payload := dbmodel.GenericWorkload{
		Spec: api.SpecificRunInfoFields{
			Assets: api.AssetsRef{
				Datasources:     make([]api.DatasourceRef, 0),
				WorkloadVolumes: make([]string, 0),
			},
		},
	}

	//
	//  The metadata of the workload can be taken from any record in the result, they are all the same
	//  only the assets part of the record changes
	//
	workloadRec := workloadAssets[0].WorkloadRec
	payload.Meta = workloadRec.WorkloadMetaV1
	payload.Spec.SpecificEnv = workloadRec.SpecificEnv
	payload.CreatedAt = &workloadRec.CreatedAt

	distFramework := workloadRec.DistFramework
	if distFramework != nil {
		// init distributed spec
		payload.Spec.Distributed = &api.InfoDistributed{
			DistFramework: distFramework,
			NumWorkers:    workloadRec.DistSpecificEnv.NumWorkers,
			NoMaster:      workloadRec.DistSpecificEnv.NoMaster,
		}

		// init master spec
		if !convutils.Deref(workloadRec.DistSpecificEnv.NoMaster, false) {
			payload.Spec.Distributed.Master = &api.InfoDistMaster{}
		}
	}

	//
	//   prepare the workload for storing in the result
	//
	StoreWorkloadIdInWorkload(workloadType, &workloadRec.Workload, workloadRec.Id)
	StoreClusterIdInWorkload(workloadType, &workloadRec.Workload, workloadRec.ClusterId)

	workloadJson, convError := gojsonq.LoadJson(&workloadRec.Workload)
	if convError != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse the payload of the workload", convError)
	}
	payload.Workload = workloadJson

	// fill the assets part of the workload
	if payload.Spec.Distributed == nil {
		workloadErr := dao.WorkloadRecToAssetsRef(&payload.Spec.Assets, workloadAssets, nil)
		if workloadErr != nil {
			return nil, workloadErr
		}
	} else {
		distReplicaType := dbmodel.DistReplicaTypeWorker
		workloadErr := dao.WorkloadRecToAssetsRef(&payload.Spec.Assets, workloadAssets, &distReplicaType)
		if workloadErr != nil {
			return nil, workloadErr
		}

		if payload.Spec.Distributed.Master != nil {
			distReplicaType = dbmodel.DistReplicaTypeMaster
			var assetsRef api.AssetsRef
			workloadErr = dao.WorkloadRecToAssetsRef(&assetsRef, workloadAssets, &distReplicaType)
			if assetsRef.Environment.Id != "" {
				payload.Spec.Distributed.Master.Assets = &api.AssetsRefOptional{}
				copier.Copy(payload.Spec.Distributed.Master.Assets, &assetsRef)

				payload.Spec.Distributed.Master.SpecificEnv = workloadRec.DistSpecificEnv.MasterSpecificRunParams
			}
			if workloadErr != nil {
				return nil, workloadErr
			}
		}
	}
	return &payload, nil
}

// GetWorkloadAssetsByWorkloadId returns all the assets contained in workload. Each asset element record is a "join" of
// workloads and workloads_assets, so both the assets meta details and the contained assets is returned
func GetWorkloadAssetsByWorkloadId(ctx context.Context, workloadTypes []string, id string) ([]dbmodel.CombinedWorkloadsRec, *api.Error) {
	workloads := make([]dbmodel.CombinedWorkloadsRec, 0)

	query := dbmodel.WorkloadsAssetsRec{
		WorkloadId: id,
	}

	result := runaidb.DB().WithContext(ctx).
		Model(&dbmodel.CombinedWorkloadsRec{}).
		Where(&query).
		Select(fmt.Sprintf("%s.*, %s.workload_id, %s.asset_id, %s.asset_kind, %s.auto_delete, %s.dist_replica_type, %s.overrides, %s.name as asset_name",
			dbtables.TableNameWorkloads,
			dbtables.TableNameWorkloadsAssets,
			dbtables.TableNameWorkloadsAssets,
			dbtables.TableNameWorkloadsAssets,
			dbtables.TableNameWorkloadsAssets,
			dbtables.TableNameWorkloadsAssets,
			dbtables.TableNameWorkloadsAssets,
			dbtables.TableNameAssetsMeta)).
		Where(fmt.Sprintf("%s IN ?", dbtables.WorkloadsColumnWorkloadType), workloadTypes).
		Joins(JoinWorkloadsAndWorkloadsAssets()).
		Joins(JoinWorkloadAssetsAndAssetMeta())

	_, scanErr := result.Scan(&workloads)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to get workspace %s", id), scanErr)
	}

	return workloads, nil
}

// SearchWorkloads find workload assets by criteria
func SearchWorkloads(ctx context.Context, workloadType string,
	id *string, kind *api.AssetKind, allowedProjectIds *[]string,
) ([]string, []dbmodel.CombinedWorkloadsRec, *api.Error) {
	workloads := make([]dbmodel.CombinedWorkloadsRec, 0)

	query := dbmodel.CombinedWorkloadsRec{}
	if id != nil {
		query.WorkloadsAssetsRec.AssetId = *id
	}
	if kind != nil {
		query.WorkloadsAssetsRec.AssetKind = *kind
	}

	combinedRec := dbmodel.CombinedWorkloadsRec{}
	result := runaidb.DB().WithContext(ctx).Model(combinedRec).
		Select("*").
		Where(fmt.Sprintf("%s = ?", dbtables.WorkloadsColumnWorkloadType), workloadType).
		Joins(JoinWorkloadsAndWorkloadsAssets()).
		Where(query)

	if allowedProjectIds != nil {
		result.Where(fmt.Sprintf("%s IN (?)", dbtables.WorkloadsColumnProjectId), *allowedProjectIds)
	}

	result.Where(fmt.Sprintf("%s.%s = ?", dbtables.TableNameWorkloads, ColumnTenantId), auth.TenantIdOf(ctx))

	_, scanErr := result.
		Order(&combinedRec.WorkloadsAssetsRec.AssetKind).
		Order(&combinedRec.WorkloadsAssetsRec.AssetId).
		Order(&combinedRec.WorkloadRec.Id).
		Scan(&workloads)
	if scanErr != nil {
		return nil, nil, api.NewError(http.StatusServiceUnavailable, "failed to list workloads", scanErr)
	}

	names := make([]string, 0, len(workloads))
	for _, workspace := range workloads {
		names = append(names, workspace.WorkloadRec.Name)
	}
	return names, workloads, nil
}

// updates lastUsedByWorkload table for the assets used by a specific workload
func updateAssetsLastUsedByWorkload(ctx context.Context, workloadType string, workloadId string, projectId int32, workloadAssets []dbmodel.AssetsContainedInWorkload,
	masterWorkloadAssets []dbmodel.AssetsContainedInWorkload,
) *api.Error {
	var allWorkloadAssets []dbmodel.AssetsContainedInWorkload
	if len(masterWorkloadAssets) != 0 {
		allWorkloadAssets = append(workloadAssets, masterWorkloadAssets...)
	} else {
		allWorkloadAssets = workloadAssets
	}

	uniqueAssetsIds := datastructs.NewSet[string]()
	for _, asset := range allWorkloadAssets {
		if uniqueAssetsIds.Has(asset.Id) {
			continue
		}
		uniqueAssetsIds.Add(asset.Id)

		err := dao.UpsertAssetLastUsedByWorkload(ctx, asset.Id, workloadType, workloadId, projectId, time.Now())
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update %s table for asset of id :%s, Error: %v", dbtables.TableNameLastUsedByWorkload, asset.Id, err)
			return err
		}
	}

	return nil
}

func StoreClusterIdInWorkload(_ /*workloadType*/ string, workload *jobs.JobWorkload, id string) {
	StoreInAnnotation(workload, annotationClusterId, id)
}

func StoreWorkloadIdInWorkload(workloadType string, workload *jobs.JobWorkload, id string) {
	if workloadType == auth.WorkloadTypeWorkspace {
		StoreInAnnotation(workload, annotationWorkspaceId, id)
	}
	StoreInAnnotation(workload, annotationWorkloadId, id)
}

func StoreInAnnotation(workload *jobs.JobWorkload, annotationKey string, annotationValue string) {
	annotations := map[string]string{
		annotationKey: annotationValue,
	}

	if workload.Annotations == nil {
		workload.Annotations = map[string]string{}
	}
	workload.Annotations[annotationKey] = annotationValue

	if workload.Spec.Annotations == nil {
		workload.Spec.Annotations = field_types.NewDictionaryField(annotations)
	}
	(*workload.Spec.Annotations.Value)[annotationKey] = field_types.DictionaryItemAndRules{
		Value: convutils.Ref(annotationValue),
	}
}

// UpdateWorkloadCurrentState sets workloads.current_state for the given workload id (same as the legacy backend).
func UpdateWorkloadCurrentState(ctx context.Context, clusterId, workloadId, currentState string) *api.Error {
	db := runaidb.DB().WithContext(ctx)
	updRows, updErr := db.Model(&dbmodel.WorkloadRec{}).
		Where(fmt.Sprintf("%s = ? AND %s = ? AND %s = ?",
			dbtables.WorkloadsColumnTenantId,
			dbtables.WorkloadsColumnClusterId,
			dbtables.WorkloadsColumnId),
			auth.TenantIdOf(ctx), clusterId, workloadId).
		Updates(map[string]interface{}{
			dbtables.WorkloadsColumnCurrentState: currentState,
			dbtables.WorkloadsColumnUpdatedBy:    auth.SubjectIdOf(ctx),
		})
	if updErr != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to update current state of %s to %s per update from cluster %s", workloadId, currentState, clusterId),
			updErr)
	}
	if updRows == 0 {
		return api.NewError(http.StatusNotFound,
			fmt.Sprintf("%s does not exist, handling status update from %s failed", workloadId, clusterId), nil)
	}
	return nil
}

// DeleteWorkloadAndAssetsById deletes the workload row by id, then all workloads_assets rows for that workload id,
// matching the legacy backend functions.
func DeleteWorkloadAndAssetsById(ctx context.Context, clusterId string, workloadId string) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	rowsAffected, err := tx.Where(fmt.Sprintf("%s = ? AND %s = ? AND %s = ?",
		dbtables.WorkloadsColumnTenantId,
		dbtables.WorkloadsColumnClusterId,
		dbtables.WorkloadsColumnId),
		auth.TenantIdOf(ctx), clusterId, workloadId).
		Delete(&dbmodel.WorkloadRec{})
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to delete %s per update from %s", workloadId, clusterId), err)
	}
	if rowsAffected == 0 {
		return api.NewError(http.StatusNotFound,
			fmt.Sprintf("%s does not exist, ignore update from %s", workloadId, clusterId), nil)
	}

	_, err = tx.Where(fmt.Sprintf("%s = ?", dbtables.WorkloadsAssetsColumnWorkloadId), workloadId).
		Delete(&dbmodel.WorkloadsAssetsRec{})
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to delete assets of %s per update from %s", workloadId, clusterId), err)
	}
	if err = tx.Commit(); err != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to commit delete of %s following update from %s", workloadId, clusterId), err)
	}

	committed = true
	return nil
}
