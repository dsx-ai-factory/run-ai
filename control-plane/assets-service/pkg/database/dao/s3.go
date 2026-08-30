package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

func CreateS3Asset(ctx context.Context, request api.S3CreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.S3Asset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createS3AssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_S3).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createS3AssetInTransaction(ctx context.Context, tx orm.DB, request api.S3CreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.S3Asset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_S3, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_S3, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_S3).WithDetails(err)
	}

	s3AssetRec := dbmodel.S3AssetRec{
		AssetId: id,
		S3AssetSpec: api.S3AssetSpec{
			Url:              request.Spec.Url,
			Bucket:           request.Spec.Bucket,
			Path:             request.Spec.Path,
			AccessKeyAssetId: request.Spec.AccessKeyAssetId,
		},
	}

	_, err = tx.Create(s3AssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_S3).WithDetails(err)
	}

	return &api.S3Asset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

func GetS3AssetById(ctx context.Context, id string) (*api.S3Asset, *api.Error) {
	var combinedRec dbmodel.S3CombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_S3,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsS3)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_S3).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.S3Asset{
		Meta: combinedRec.AssetMeta,
		Spec: combinedRec.S3AssetSpec,
	}
	return &response, nil
}

func ListS3Assets(ctx context.Context, queryParams *query_types.QueryParams) (*api.S3ListResponse, *api.Error) {
	combinedList := make([]dbmodel.S3CombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_S3,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsS3)
	if err != nil {
		return nil, err
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list s3 assets", scanErr)
	}

	response := api.S3ListResponse{
		Entries: make([]api.S3Asset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.S3Asset{
			Meta: row.AssetMeta,
			Spec: row.S3AssetSpec,
		})
	}

	return &response, nil
}

func ListS3UsingAccessKey(ctx context.Context, specificId *string) (*api.S3ListResponse, *api.Error) {
	combinedList := make([]dbmodel.S3CombinedRec, 0)
	kindFilter := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_S3,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), kindFilter, nil, dbtables.TableNameAssetsS3)
	if err != nil {
		return nil, err
	}

	if specificId != nil {
		result.Where(fmt.Sprintf("%s.%s = ?", dbtables.TableNameAssetsS3, dbtables.S3ColumnAccessKeyAssetId), *specificId)
	} else {
		result.Where(fmt.Sprintf("%s.%s IS NOT NULL", dbtables.TableNameAssetsS3, dbtables.S3ColumnAccessKeyAssetId))
	}

	result.Order(fmt.Sprintf("%s.%s", dbtables.TableNameAssetsS3, dbtables.S3ColumnAccessKeyAssetId))

	_, scanErr := result.Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list s3 assets", scanErr)
	}

	response := api.S3ListResponse{
		Entries: make([]api.S3Asset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.S3Asset{
			Meta: row.AssetMeta,
			Spec: row.S3AssetSpec,
		})
	}

	return &response, nil
}

type AssetVsCredentialsValidationHook func(ctx context.Context, scope api.Scope, scopeId string, accessKeyId *string) *api.Error

func UpdateS3Asset(ctx context.Context, id string, request api.S3UpdateRequest,
	validationHook AssetVsCredentialsValidationHook, nameValidationQuery *database.ParametrizedQuery,
) (*api.S3Asset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, assetMeta, err := updateS3AssetInTransaction(ctx, tx, id, request, validationHook, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	policyConflict := ValidateAssetVsPolicy(ctx, id, ScopeInfoOfAssetMeta(&result.Meta))
	if policyConflict != nil {
		return nil, api.NewError(int(policyConflict.Code), policyConflict.Message, nil)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_S3).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return result, nil
}

func updateS3AssetInTransaction(ctx context.Context, tx orm.DB, id string, request api.S3UpdateRequest, validationHook AssetVsCredentialsValidationHook,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.S3Asset, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_S3).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if validationHook != nil {
		scopeInfo := ScopeInfoOfDbMeta(assetMeta)
		validationErr := validationHook(ctx, assetMeta.Scope, scopeInfo.ScopeId, request.Spec.AccessKeyAssetId)
		if validationErr != nil {
			return nil, assetMeta, validationErr
		}
	}

	recId := dbmodel.S3AssetRec{AssetId: id}
	s3AssetUpdate := dbmodel.S3AssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&s3AssetUpdate); err != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_S3).WithDetails(err)
	}

	s3AssetUpdate.S3AssetSpec = request.Spec

	saveRows, saveErr := tx.Where(recId).Save(&s3AssetUpdate)
	if saveErr != nil {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_S3).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, assetMeta, policyErr
	}

	return &api.S3Asset{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: request.Spec,
	}, assetMeta, nil
}

func DeleteS3Asset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteS3AssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_S3).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

func deleteS3AssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	recId := dbmodel.S3AssetRec{AssetId: id}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_S3).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_S3).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}

// countS3AssetsUsingAccessKeyId counts the number of S3 assets which uses a given access key
func countS3AssetsUsingAccessKeyId(tx orm.DB, accessKeyId string) (int64, *api.Error) {
	S3AssetRec := dbmodel.S3AssetRec{
		S3AssetSpec: api.S3AssetSpec{
			AccessKeyAssetId: &accessKeyId,
		},
	}

	count := int64(0)
	if countErr := tx.Model(&S3AssetRec).
		Where(api.S3AssetSpec{AccessKeyAssetId: &accessKeyId}).
		Count(&count); countErr != nil {
		return 0, database.Error(database.COUNT, api.ASSETKIND_S3).WithDetails(countErr)
	}

	return count, nil
}
