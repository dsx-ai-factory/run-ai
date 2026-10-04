/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	DefaultAutoDelete        = false
	clusterPlaceHolderArg    = "clusterId"
	departmentPlaceHolderArg = "departmentId"
	projectPlaceHolderArg    = "projectId"
	alreadyExistsError       = "already exists"
)

type DeletionType int

const (
	SoftDelete = iota
	HardDelete
)

var SoftDeleteAssets = datastructs.NewSet[api.AssetKind](
	api.ASSETKIND_ACCESS_KEY,
	api.ASSETKIND_DOCKER_REGISTRY,
	api.ASSETKIND_GENERIC_SECRET,
	api.ASSETKIND_PASSWORD,
	api.ASSETKIND_PVC,
	api.ASSETKIND_NGC_API_KEY,
	api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
)

func JoinAssetMetaWorkloadSupportedTypes() string {
	return fmt.Sprintf("left join %s on %s.asset_id = %s.id",
		dbtables.TableNameWorkloadSupportedTypes, dbtables.TableNameWorkloadSupportedTypes, dbtables.TableNameAssetsMeta)
}

func createAssetMeta(ctx context.Context, tx orm.DB, id string, kind api.AssetKind, assetMeta api.AssetCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.AssetMeta, error) {
	//
	// autoDeletion is considered internal feature supported only for PVC. ignore it with logging in case it is
	// requested for other kind of assets. no need for user facing error because it is internal feature.
	//
	lockRec, lockErr := createAssetNameLock(ctx, assetMeta.Name, kind)
	if lockErr != nil {
		return nil, fmt.Errorf("creating an asset of kind '%s' with the name '%s' is temporarily unavailable, try again later", kind, assetMeta.Name)
	}
	defer deleteAssetNameLock(ctx, lockRec.Id)

	if convutils.Deref(assetMeta.AutoDelete, DefaultAutoDelete) && kind != api.ASSETKIND_PVC {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("autoDelete is supported only for PVC assets")
		assetMeta.AutoDelete = nil
	}

	if err := ValidateAssetNameUniqueness(ctx, tx, kind, assetMeta.Name, nameValidationQuery); err != nil {
		return nil, err
	}

	assetMetaRec := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Id:           id,
			Kind:         kind,
			Scope:        assetMeta.Scope,
			DepartmentId: assetMeta.DepartmentId,
			ClusterId:    assetMeta.ClusterId,
			ProjectId:    assetMeta.ProjectId,
			Name:         assetMeta.Name,
			CreatedBy:    auth.SubjectIdOf(ctx),
			UpdatedBy:    auth.SubjectIdOf(ctx),
			AutoDelete:   assetMeta.AutoDelete,
			Description:  assetMeta.Description,
			Private:      assetMeta.Private,
		},
	}

	if assetMeta.Scope != api.SCOPE_SYSTEM {
		assetMetaRec.TenantId = convutils.Ref(auth.TenantIdOf(ctx))
	}

	_, err := tx.Create(&assetMetaRec)
	if err != nil {
		return nil, err
	}

	recId := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Id: id,
		},
	}

	_, err = tx.Model(recId).Where(recId).Select("*").Scan(&assetMetaRec)
	if err != nil {
		return nil, err
	}

	meta := api.AssetMeta{}

	if kind == api.ASSETKIND_ENVIRONMENT {
		workloadSupportedTypes, wstErr := createWorkloadSupportedTypes(tx, id, kind, assetMeta)
		if wstErr != nil {
			return nil, wstErr
		}

		meta = assetMetaRec.ToApiAssetMeta()
		meta.WorkloadSupportedTypes = workloadSupportedTypes.ToApiWorkloadSupportedTypes()

	} else {
		meta = assetMetaRec.ToApiAssetMeta()
	}

	return &meta, nil
}

type UpdateAssetMetaOptions struct {
	RenameNotAllowedDuringSync bool // for synced credentials rename is not allowed during sync
}

func updateAssetMeta(ctx context.Context, tx orm.DB, id string, updateRequest api.AssetUpdateRequest,
	nameValidationQuery *database.ParametrizedQuery, options *UpdateAssetMetaOptions) (*dbmodel.AssetMeta, int64, *api.Error, error) {
	assetMetaRec, rowsAffected, err, dbErr := getAssetMetaByIdInt(ctx, tx, id, false)
	if err != nil {
		return nil, 0, err, nil
	}
	if rowsAffected == 0 || dbErr != nil {
		return nil, rowsAffected, nil, dbErr
	}

	assetMetaRec.UpdatedAt = time.Now()
	assetMetaRec.UpdatedBy = convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "unknown")

	if assetMetaRec.Name != updateRequest.Name {
		if options != nil && options.RenameNotAllowedDuringSync {
			inSyncErr := validateNotInSync(ctx, tx, assetMetaRec)
			if inSyncErr != nil {
				return nil, 0, inSyncErr, nil
			}
		}

		assetMetaRec.PrevName = assetMetaRec.Name

		lockRec, lockErr := createAssetNameLock(ctx, updateRequest.Name, assetMetaRec.Kind)
		if lockErr != nil {
			return nil, 0, nil, fmt.Errorf("renaming an asset of kind '%s' with the name '%s' is temporarily unavailable, try again later", assetMetaRec.Kind, assetMetaRec.Name)
		}
		defer deleteAssetNameLock(ctx, lockRec.Id)
	}

	if dbErr := ValidateAssetNameUniqueness(ctx, tx, assetMetaRec.Kind, updateRequest.Name, nameValidationQuery); dbErr != nil {
		return nil, 0, nil, dbErr
	}

	assetMetaRec.Name = updateRequest.Name
	assetMetaRec.Description = updateRequest.Description
	assetMetaRec.UpdatedBy = auth.SubjectIdOf(ctx)

	rowsAffected, saveErr := tx.Save(assetMetaRec)
	if saveErr != nil {
		return nil, rowsAffected, nil, saveErr
	}

	assetMetaRec, rowsAffectedAfter, err, dbErrAfter := getAssetMetaByIdInt(ctx, tx, id, false)
	return &assetMetaRec.AssetMeta, rowsAffectedAfter, err, dbErrAfter
}

func deleteAssetMeta(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) (int64, *api.Error, error) {
	return deleteAssetMetaByIds(ctx, tx, []string{id}, deletionType)
}

func deleteAssetMetaByIds(ctx context.Context, tx orm.DB, ids []string, deletionType DeletionType) (int64, *api.Error, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	// Get all asset metadata records for the given IDs
	assetMetaRecs, err := getAssetMetaByIdsInt(ctx, tx, ids, true) /* allow deletion of currently deleted assets */
	if err != nil {
		return 0, err, nil
	}
	if len(assetMetaRecs) == 0 {
		return 0, nil, nil
	}

	// delete workload supported types first
	rowsAffected, delErr := DeleteWorkloadSupportedTypesOfByIds(tx, ids)
	if delErr != nil {
		return rowsAffected, nil, delErr
	}

	// Check if any of the assets are PVCs for hard delete
	hasPVC := false
	for _, assetMetaRec := range assetMetaRecs {
		if assetMetaRec.Kind == api.ASSETKIND_PVC && deletionType == HardDelete {
			hasPVC = true
			break
		}
	}

	if hasPVC && deletionType == HardDelete {
		// Delete all mappings of the assets to resource ids
		if apiErr := DeleteAssetResourceIdsByIds(ctx, tx, ids); apiErr != nil {
			return rowsAffected, apiErr, nil
		}

		// Delete all events of the assets
		if apiErr := DeleteAssetEventsByIds(ctx, tx, ids); apiErr != nil {
			return rowsAffected, apiErr, nil
		}
	}

	if deletionType == SoftDelete {
		// Check if all asset kinds support soft deletion
		for _, assetMetaRec := range assetMetaRecs {
			if !SoftDeleteAssets.Has(assetMetaRec.Kind) {
				return 0, api.NewError(http.StatusInternalServerError, fmt.Sprintf("asset of kind '%s' does not support soft deletion", assetMetaRec.Kind), nil), nil
			}
		}

		// Get tenant ID from the first record (assuming all are from the same tenant)
		tenantId := assetMetaRecs[0].TenantId
		if _, byErr := tx.Model(&dbmodel.AssetMetaRec{}).Unscoped().
			Where(fmt.Sprintf("%s = ?", dbtables.AssetMetaColumnTenantId), tenantId).
			Where(fmt.Sprintf("%s IN ?", dbtables.AssetMetaColumnId), ids).
			UpdateColumn(dbtables.AssetMetaColumnDeletedBy, convutils.Ref(auth.SubjectIdOf(ctx))); byErr != nil {
			return 0, nil, byErr
		}

		atRows, atErr := tx.Model(&dbmodel.AssetMetaRec{}).Unscoped().
			Where(fmt.Sprintf("%s = ?", dbtables.AssetMetaColumnTenantId), tenantId).
			Where(fmt.Sprintf("%s IN ?", dbtables.AssetMetaColumnId), ids).
			UpdateColumn(dbtables.AssetMetaColumnDeletedAt, time.Now())
		return atRows, nil, atErr
	} else {
		delRows, delErr := tx.Unscoped().Model(&dbmodel.AssetMetaRec{}).Where(fmt.Sprintf("%s IN ?", dbtables.AssetMetaColumnId), ids).Delete(&dbmodel.AssetMetaRec{})
		return delRows, nil, delErr
	}
}

func ListContainedAssetsByIds(ctx context.Context, assetIds []string, queryParams *query_types.QueryParams) (dbmodel.AssetIdToMetaMap, *api.Error) {
	return ListAssetsMetaByIds(ctx, assetIds, true, queryParams) /* for content of workload/template we want to see also auto-delete items*/
}

// ListAssetsMetaByIds returns metadata information of all assets based on their Ids.
// it is being used by workload/template related functions and also by list policies and by list assets by ids
func ListAssetsMetaByIds(ctx context.Context, assetIds []string, includeAutoDeleteAssets bool, queryParams *query_types.QueryParams) (dbmodel.AssetIdToMetaMap, *api.Error) {
	entries := make([]dbmodel.CombinedRec, 0)

	model := dbmodel.AssetMetaRec{}

	if queryParams == nil {
		queryParams = &query_types.QueryParams{}
	}
	queryParams.IncludeAutoDeleteAssets = includeAutoDeleteAssets

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), model, queryParams)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Where("id IN ?", assetIds).
		Order(&model.AssetMeta.Id).
		Scan(&entries)
	if scanErr != nil {
		return nil, database.Error(database.LIST, "metadata").WithDetails(scanErr)
	}

	allNil := api.WorkloadSupportedTypes{}

	assetMetaById := make(dbmodel.AssetIdToMetaMap)
	for _, entry := range entries {
		if entry.WorkloadSupportedTypes != allNil {
			entry.AssetMeta.WorkloadSupportedTypes = convutils.Ref(entry.WorkloadSupportedTypes)
		}
		assetMetaById[entry.Id] = entry
	}

	return assetMetaById, nil
}

func ListAssetMeta(ctx context.Context, query dbmodel.AssetMetaRec) (*[]dbmodel.AssetMeta, *api.Error) {
	result, err := getAssetMetaInt(ctx, runaidb.DB().WithContext(ctx), query)
	if err != nil {
		return nil, err
	}

	metadataRecords := make([]dbmodel.AssetMeta, 0)

	_, scanErr := result.Scan(&metadataRecords)
	if scanErr != nil {
		return nil, database.Error(database.LIST, "metadata").WithDetails(scanErr)
	}

	return &metadataRecords, nil
}

func GetAssetMetaById(ctx context.Context, id string) (*dbmodel.AssetMeta, *api.Error) {
	var rec dbmodel.AssetMetaRec
	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), dbmodel.AssetMetaRec{}, nil)
	if err != nil {
		return nil, err
	}

	result = result.Unscoped()

	// Either it's a non-private or owned by the subject
	result.Where(fmt.Sprintf("COALESCE(%v.%v, FALSE) = false OR %s.%s = ?",
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate,
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnCreatedBy), auth.SubjectIdOf(ctx))

	scanRows, scanErr := result.
		Where(dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Id: id}}).
		Scan(&rec)
	if scanErr != nil || scanRows == 0 {
		return nil, api.NewError(http.StatusNotFound, "requested asset does not exist", scanErr)
	}
	return &rec.AssetMeta, nil
}

// GetAssetMeta is a general function for getting/listing/fetching an asset. it complements the original
// gorm.DB object with the necessary select clauses according to:
//   - query: typically containing asset type and optional id
//   - authInfo: containing project/cluster/tenant access information. note that auth
//     info can be nil. this can happen in create/update functions in which we reload
//     the complete asset for returning to the requester. in this case all the filtering
//     has been done already, and no need to repeat them.
//   - queryParams: search filters like name, projectId, scope etc. this is typically sent
//     as part of list asset. in other cases it can be sent as nil.
//   - joinWith: join with additional tables. supported name of table or table.column
func GetAssetMeta(ctx context.Context, tx orm.DB, query dbmodel.AssetMetaRec,
	queryParams *query_types.QueryParams, joinWith ...string,
) (orm.DB, *api.Error) {
	result := tx.Model(dbmodel.AssetMetaRec{})

	selectedColumns := fmt.Sprintf("%s.*, %s.*", dbtables.TableNameAssetsMeta, dbtables.TableNameWorkloadSupportedTypes)
	for _, currentJoin := range joinWith {
		selectedColumns += fmt.Sprintf(",%s", currentJoin)
		// if column is not specified in the parameter, join all the columns
		if !strings.Contains(currentJoin, ".") {
			selectedColumns += ".*"
		}
	}
	result.Select(selectedColumns)

	var err *api.Error
	result, err = filterAssetMeta(ctx, result, query, queryParams)
	if err != nil {
		return nil, err
	}

	result.Joins(JoinAssetMetaWorkloadSupportedTypes())

	for _, currentJoin := range joinWith {
		currentTable := strings.Split(currentJoin, ".")[0]
		result.Joins(database.JoinAssetMetaAndSpec(currentTable))
	}

	return result, nil
}

func FindAssetMetaOfPVCEvent(ctx context.Context, event *api.EventSyncFields) (*dbmodel.AssetMeta, *api.Error) {
	db := runaidb.DB().WithContext(ctx)

	var assetId string

	filters := dbmodel.AssetResourceIdsRecord{
		ResourceId: event.InvolvedObject.Uid,
	}

	err := db.Model(filters).
		Select(dbtables.AssetIdColumnName).
		Where(filters).
		First(&assetId)
	if err != nil {
		if errors.Is(err, orm.ErrRecordNotFound) {
			return nil, api.NewError(http.StatusInternalServerError, "no asset id found for the event with id "+event.Id, nil)
		}
		return nil, database.Error(database.GET, dbtables.TableNameAssetsResourcesIds).WithDetails(err)
	}

	rec := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Id: assetId,
		},
	}

	if err := db.Model(rec).
		Where(rec).
		First(&rec); err != nil {
		return nil, database.Error(database.GET, dbtables.TableNameAssetsMeta).WithDetails(err)
	}

	return &rec.AssetMeta, nil
}

// minimized version of GetAssetMeta, with no joins, just filtering. used for internal
// purposes, for example before deletion when we want to look at asset meta without any joins
func getAssetMetaInt(ctx context.Context, tx orm.DB, query dbmodel.AssetMetaRec) (orm.DB, *api.Error) {
	result := tx.Model(dbmodel.AssetMetaRec{})

	var err *api.Error
	result, err = filterAssetMeta(ctx, result, query, nil)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func getAssetMetaByIdInt(ctx context.Context, tx orm.DB, id string, includeDeleted bool) (*dbmodel.AssetMetaRec, int64, *api.Error, error) {
	assetMetaRec := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{Id: id},
	}

	result, err := getAssetMetaInt(ctx, tx, assetMetaRec)
	if err != nil {
		return nil, 0, err, nil
	}

	if includeDeleted {
		result = result.Unscoped()
	}

	rowsAffected, scanErr := result.Scan(&assetMetaRec)
	if rowsAffected == 0 || scanErr != nil {
		return nil, rowsAffected, nil, scanErr
	}

	return &assetMetaRec, rowsAffected, nil, nil
}

func getAssetMetaByIdsInt(ctx context.Context, tx orm.DB, ids []string, includeDeleted bool) ([]dbmodel.AssetMetaRec, *api.Error) {
	if len(ids) == 0 {
		return []dbmodel.AssetMetaRec{}, nil
	}

	assetMetaRecs := make([]dbmodel.AssetMetaRec, 0)

	result := tx.Model(&dbmodel.AssetMetaRec{})
	if includeDeleted {
		result = result.Unscoped()
	}

	result = result.Where(fmt.Sprintf("%s IN ?", dbtables.AssetMetaColumnId), ids)
	_, scanErr := result.Scan(&assetMetaRecs)
	if scanErr != nil {
		return nil, database.Error(database.LIST, "metadata").WithDetails(scanErr)
	}

	return assetMetaRecs, nil
}

// CountAssetMeta works the same as GetAssetMeta, but returns count of the results instead of the
// results themselves. see GetAssetMeta for the parameters
func CountAssetMeta(ctx context.Context, kind *api.AssetKind, queryParams *query_types.QueryParams) (orm.DB, *api.Error) {
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{},
	}

	if kind != nil {
		query.Kind = *kind
	}

	result := runaidb.DB().WithContext(ctx).
		Model(dbmodel.AssetMetaRec{}).
		Select("count(*) as count")

	result, err := filterAssetMeta(ctx, result, query, queryParams)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func filterAssetMeta(ctx context.Context, tx orm.DB, query dbmodel.AssetMetaRec, queryParams *query_types.QueryParams) (orm.DB, *api.Error) {
	result, filterError := filterResultsBasedOnQueryParams(ctx, tx, &query, queryParams)
	if filterError != nil {
		return nil, filterError
	}

	filterAllowedOrgUnits(ctx, result, queryParams)

	return result, nil
}

// filterResultsBasedOnQueryParams enrich the "get" query based on query and queryParams
func filterResultsBasedOnQueryParams(ctx context.Context, tx orm.DB, query *dbmodel.AssetMetaRec, queryParams *query_types.QueryParams) (orm.DB, *api.Error) {
	if queryParams != nil {
		if queryParams.AssetIds != nil {
			tx.Where(fmt.Sprintf("%s.%s in ?", dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnId), *queryParams.AssetIds)
		}

		if queryParams.Name != nil {
			query.Name = *queryParams.Name
		}

		if queryParams.Scope != nil {
			if *queryParams.Scope != api.SCOPE_TENANT {
				query.Scope = *queryParams.Scope
			} else {
				// for tenant scope we need to filter by both system and tenant scopes
				tx.Where(fmt.Sprintf("(%s.scope = ?) or (%s.scope = ?)", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta),
					api.SCOPE_SYSTEM, api.SCOPE_TENANT)
			}
		}

		HandleDistributedQueryParams(tx, queryParams)

		//   currently filtering by cluster ID is only used by asset sync. it provides cluster id only for project-specific
		//   assets, and in this case projectId will also be provided (see BuildQueryParamsForSyncRequest).
		if queryParams.ClusterId != nil {
			tx.Where(fmt.Sprintf("%s.cluster_id = ? or %s.cluster_id is null", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), *queryParams.ClusterId)
		}

		//   typically, we do not want to list auto-delete assets because they are not user-facing. we only list them if specifically requested
		if !toIncludeAutoDeleteAssets(queryParams) {
			tx.Where(fmt.Sprintf("%s.auto_delete is null or %s.auto_delete = ?", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), false)
		}

		// PrivateOnly and IncludePrivate are mutually exclusive.
		switch {
		case queryParams.PrivateOnly:
			// Return only assets that are private AND created by the subject
			tx.Where(fmt.Sprintf("COALESCE(%v.%v, FALSE) = true AND %s.%s = ?",
				dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate,
				dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnCreatedBy), auth.SubjectIdOf(ctx))
		case queryParams.IncludePrivate:
			// Return non-private OR created by the subject
			tx.Where(fmt.Sprintf("COALESCE(%v.%v, FALSE) = false OR %s.%s = ?",
				dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate,
				dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnCreatedBy), auth.SubjectIdOf(ctx))
		default:
			// Return only non-private
			tx.Where(fmt.Sprintf("COALESCE(%v.%v, FALSE) = false",
				dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate))
		}

		//   when searching for projectId we want all the assets that are relevant to the specified project.
		//   that includes both tenant-scoped assets accessible by the project/cluster and assets specific to this project/cluster
		//   to filter only project-scoped assets the query needs to include scope=Project
		byResource := database.NewParametrizedQuery()

		// prevent sql injection in or clause
		if queryParams.ProjectId != nil {
			if queryParams.Scope == nil {
				resp, err := org_unit.Get().GetProjectAncestors(ctx, fmt.Sprintf("%v", *queryParams.ProjectId))
				if err != nil {
					return nil, api.NewError(http.StatusServiceUnavailable, "failed to interrogate the ancestors of the project", errors.New(err.Message))
				}

				// when project is not specified, we want system assets + tenants assets + department assets from the department that
				// project belongs to + project specific records
				byResource.Add(fmt.Sprintf("(%s.scope = ?) or (%s.scope = ?) or ",
					dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_SYSTEM, api.SCOPE_TENANT)

				departmentId := resp.DepartmentId.Get()
				if departmentId != nil {

					byResource.Add(fmt.Sprintf("(%s.scope = ? and %s.department_id = ?) or ",
						dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_DEPARTMENT, departmentId)
				}

				clusterId := resp.ClusterUuid.Get()
				if clusterId != nil {
					byResource.Add(fmt.Sprintf("(%s.scope = ? and %s.cluster_id = ?) or ",
						dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_CLUSTER, clusterId)
				}
			}

			byResource.Add(fmt.Sprintf("(%s.project_id is not null and %s.project_id = ?)", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), *queryParams.ProjectId)

			tx.Where(byResource.Query, byResource.Params...)
		} else if queryParams.DepartmentId != nil {
			if queryParams.Scope == nil {
				resp, err := org_unit.Get().GetDepartmentAncestors(ctx, fmt.Sprintf("%v", *queryParams.DepartmentId))
				if err != nil {
					return nil, api.NewError(int(err.GetCode()), "failed to interrogate the ancestors of the department", errors.New(err.Message))
				}
				clusterId := resp.ClusterUuid.Get()
				if clusterId != nil {
					byResource.Add(fmt.Sprintf("(%s.scope = ? and %s.cluster_id = ?) or ",
						dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_CLUSTER, *clusterId)
				}

				byResource.Add(fmt.Sprintf("(%s.scope = ?) or (%s.scope = ?) or ",
					dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_SYSTEM, api.SCOPE_TENANT)
			}

			if queryParams.IncludeDescendants {
				byResource.Add(fmt.Sprintf("(%s.department_id = ?)", dbtables.TableNameAssetsMeta), *queryParams.DepartmentId)
			} else {
				byResource.Add(fmt.Sprintf("(%s.scope = ? and %s.department_id = ?)",
					dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_DEPARTMENT, *queryParams.DepartmentId)
			}
			tx.Where(byResource.Query, byResource.Params...)
		} else if queryParams.ClusterId != nil {
			if queryParams.Scope == nil {
				// If scope isn't specified, we want to include system assets and tenants assets
				byResource.Add(fmt.Sprintf("(%s.scope = ?) or (%s.scope = ?) or ",
					dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_SYSTEM, api.SCOPE_TENANT)
			}
			if queryParams.IncludeDescendants {
				byResource.Add(fmt.Sprintf("(%s.cluster_id = ?)", dbtables.TableNameAssetsMeta), *queryParams.ClusterId)
			} else {
				// Additionally we want to include cluster-scoped assets from the specified cluster
				byResource.Add(fmt.Sprintf("(%s.scope = ? and %s.cluster_id = ?)", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), api.SCOPE_CLUSTER, *queryParams.ClusterId)
			}
			tx.Where(byResource.Query, byResource.Params...)
		}

	}

	return tx.Where(query), nil
}

func HandleDistributedQueryParams(tx orm.DB, queryParams *query_types.QueryParams) {
	// UI sends the following query params per workload type:
	//   Training:             ?isTraining=true
	//   Inference:            ?isInference=true
	//   Distributed training: ?isDistributed=true&distributedFramework=<X>&complyToReplicaType=Worker
	//   Distributed inference:?isDistributedInference=true
	//
	// IsDistributed condition rationale:
	//   When distributed inference was introduced alongside the existing distributed
	//   training support, IsDistributedInference was added to target envs with inference=true.
	//   To avoid those envs leaking into the distributed training list, IsDistributed was
	//   restricted to distributed=true AND inference=false/NULL.
	//
	//   Side effect: envs that support ALL workload types (training=true, inference=true,
	//   distributed=true) were also excluded, even though they are valid for distributed training.
	//   The inference=true value caused them to be treated as inference-only.
	//   Fix: if training=true is explicitly set, include the env regardless of inference value —
	//   an env that supports training must appear in the distributed training list.
	//   Envs that are inference-only (training=false, inference=true) remain excluded.

	if queryParams.DistributedQueryParams != nil {
		if queryParams.DistributedQueryParams.IsWorkspace {
			tx.Where(fmt.Sprintf("%s.workspace = ?", dbtables.TableNameWorkloadSupportedTypes), queryParams.DistributedQueryParams.IsWorkspace)
		}

		if queryParams.DistributedQueryParams.IsTraining {
			tx.Where(fmt.Sprintf("%s.training = ?", dbtables.TableNameWorkloadSupportedTypes), queryParams.DistributedQueryParams.IsTraining)
		}

		if queryParams.DistributedQueryParams.IsInference {
			tx.Where(fmt.Sprintf("%s.inference = ?", dbtables.TableNameWorkloadSupportedTypes), queryParams.DistributedQueryParams.IsInference)
		}

		if queryParams.DistributedQueryParams.IsDistributed {
			// Match distributed training envs (inference=false/NULL) OR universal envs (training=true).
			// Excludes pure distributed inference envs (training=false, inference=true).
			tx.Where(
				fmt.Sprintf(
					"%s.distributed = ? AND ((%s.inference = ? OR %s.inference IS NULL) OR %s.training = ?)",
					dbtables.TableNameWorkloadSupportedTypes,
					dbtables.TableNameWorkloadSupportedTypes,
					dbtables.TableNameWorkloadSupportedTypes,
					dbtables.TableNameWorkloadSupportedTypes,
				),
				true,
				false,
				true,
			)
			if queryParams.DistributedQueryParams.DistributedFramework != "" {
				tx.Where(fmt.Sprintf("%s.dist_framework = ?", dbtables.TableNameWorkloadSupportedTypes), queryParams.DistributedQueryParams.DistributedFramework)
			}
		} else if queryParams.DistributedQueryParams.IsDistributedInference {
			tx.Where(fmt.Sprintf("%s.distributed = ? AND %s.inference = ?", dbtables.TableNameWorkloadSupportedTypes, dbtables.TableNameWorkloadSupportedTypes), true, true)
		} else if queryParams.DistributedQueryParams.IsWorkspace || queryParams.DistributedQueryParams.IsTraining || queryParams.DistributedQueryParams.IsInference {
			tx.Where(fmt.Sprintf("%s.distributed = ? or %s.distributed is null", dbtables.TableNameWorkloadSupportedTypes, dbtables.TableNameWorkloadSupportedTypes), false)
		}
	}
}

func filterAllowedOrgUnits(ctx context.Context, tx orm.DB, queryParams *query_types.QueryParams) {
	tx.Where(fmt.Sprintf("%s.%s = ? OR %s.%s = ?",
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnScope,
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnTenantId),
		api.SCOPE_SYSTEM, auth.TenantIdOf(ctx))

	if queryParams == nil {
		return
	}

	if queryParams.AllowedOrgUnits != nil {
		tx.Where(fmt.Sprintf("%s.department_id IS NULL or %s.department_id IN ?", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), queryParams.AllowedOrgUnits.AllowedDepartments)
		tx.Where(fmt.Sprintf("%s.project_id IS NULL or CAST(%s.project_id as text) IN ?", dbtables.TableNameAssetsMeta, dbtables.TableNameAssetsMeta), queryParams.AllowedOrgUnits.AllowedProjects)
	}
}

func toIncludeAutoDeleteAssets(qp *query_types.QueryParams) bool {
	return qp != nil && qp.IncludeAutoDeleteAssets
}

// AssetNameAvailability validates if a given asset name is available under the given conditions.
// note that CountAsset cannot be used for this because asset count, similar to list assets (both must return same result!)
// is intended for listing/counting assets that can be USED in a given scope, and the rules of asset usage is different from
// the rules of name conflicts, which are:
//
// An asset name must be unique within the same hierarchical branch for a given asset type.
//
// Example 1:
// If an 'environment' asset named 'existing' is defined at the system scope, no other scope
// can have an 'environment' asset with the same name.
//
// Example 2:
// If an 'environment' asset named 'existing' is defined at the department scope, then:
// - None of the department's projects can have an 'environment' asset with the same name.
// - The department's cluster and the tenant level also cannot have an 'environment' asset with the same name.
func AssetNameAvailability(ctx context.Context, assetKind api.AssetKind, nameValidationQuery *database.ParametrizedQuery) (bool, *api.Error) {
	var count int64
	countErr := runaidb.DB().WithContext(ctx).
		Model(dbmodel.AssetMetaRec{}).
		Where(nameValidationQuery.Query, nameValidationQuery.Params...).
		Count(&count)

	if countErr != nil {
		return false, database.Error(database.LIST, assetKind).WithDetails(countErr)
	}

	return count == 0, nil
}

func ValidateAssetNameUniqueness(ctx context.Context, tx orm.DB, assetKind api.AssetKind, assetName string, nameValidationQuery *database.ParametrizedQuery) error {
	if nameValidationQuery == nil {
		return nil
	}

	var count int64
	countErr := tx.
		Model(dbmodel.AssetMetaRec{}).
		Where(nameValidationQuery.Query, nameValidationQuery.Params...).
		Count(&count)

	if countErr != nil {
		return countErr
	}
	if count > 0 {
		// Return a user-friendly or logged error when the name already exists
		err := fmt.Errorf("%s asset with name '%s' already exists", assetKind, assetName)
		logging.WithCtx(runaictx.LoggerHook(ctx))
		return err
	}

	return nil
}

// ValidateClaimNameUniqueness validates that a PVC claim name is unique within the scope hierarchy.
// It checks the assets_pvc table joined with assets_metadata using the provided validation query.
func ValidateClaimNameUniqueness(ctx context.Context, tx orm.DB, claimName string, claimNameValidationQuery *database.ParametrizedQuery) error {
	if claimNameValidationQuery == nil {
		return nil
	}

	var count int64
	countErr := tx.
		Table(dbtables.TableNameAssetsPVC).
		Joins(fmt.Sprintf("INNER JOIN %s ON %s.%s = %s.%s",
			dbtables.TableNameAssetsMeta,
			dbtables.TableNameAssetsMeta,
			dbtables.AssetMetaColumnId,
			dbtables.TableNameAssetsPVC,
			dbtables.AssetIdColumnName)).
		Where(claimNameValidationQuery.Query, claimNameValidationQuery.Params...).
		Count(&count)

	if countErr != nil {
		return countErr
	}
	if count > 0 {
		// Return a user-friendly or logged error when the claim name already exists
		err := fmt.Errorf("PVC asset with claim name '%s' already exists in the scope hierarchy", claimName)
		logging.WithCtx(runaictx.LoggerHook(ctx))
		return err
	}

	return nil
}

func updateCredentialAssetMeta(ctx context.Context, tx orm.DB, id string, updateRequest api.CredentialUpdateRequest) (*dbmodel.AssetMeta, int64, *api.Error, error) {
	assetMetaRec, rowsAffected, err, dbErr := getAssetMetaByIdInt(ctx, tx, id, false)
	if err != nil {
		return nil, 0, err, nil
	}
	if rowsAffected == 0 || dbErr != nil {
		return nil, rowsAffected, nil, dbErr
	}

	assetMetaRec.UpdatedAt = time.Now()
	assetMetaRec.UpdatedBy = convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "unknown")

	assetMetaRec.Description = updateRequest.Description

	assetMetaRec.UpdatedBy = auth.SubjectIdOf(ctx)

	rowsAffected, saveErr := tx.Save(assetMetaRec)
	if saveErr != nil {
		return nil, rowsAffected, nil, saveErr
	}

	assetMetaRec, rowsAffectedAfter, err, dbErrAfter := getAssetMetaByIdInt(ctx, tx, id, false)
	return &assetMetaRec.AssetMeta, rowsAffectedAfter, err, dbErrAfter
}
