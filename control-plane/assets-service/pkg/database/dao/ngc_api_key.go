package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/composer"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/encryption"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

func CreateNgcApiKey(ctx context.Context, request api.NgcApiKeyCreationRequest, keyId int32, nameValidationQuery *database.ParametrizedQuery) (*api.NgcApiKey, *api.Error) {
	id := uuid.New().String()

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_NGC_API_KEY, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_NGC_API_KEY, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_NGC_API_KEY).WithDetails(err)
	}

	ngcRecord := dbmodel.NgcApiKeyRecord{
		AssetId: id,
		NgcApiKeySpec: api.NgcApiKeySpec{
			NgcApiKey: request.Spec.NgcApiKey,
			KeyId:     keyId,
		},
	}

	_, err = tx.Create(ngcRecord)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_NGC_API_KEY).WithDetails(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_NGC_API_KEY).WithDetails(err)
	}

	committed = true
	ngcApiKeySecretReduction(&ngcRecord.NgcApiKeySpec)

	return &api.NgcApiKey{
		Meta:        *assetMeta,
		Spec:        ngcRecord.NgcApiKeySpec,
		ClusterInfo: getNgcApiKeyClusterInfo(assetMeta, nil),
	}, nil
}

func GetNgcApiKeyById(ctx context.Context, id string, options CredentialsOptions) (*api.NgcApiKey, *api.Error) {
	var combinedRecord dbmodel.NgcApiKeyFullRecord

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_NGC_API_KEY,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsNgcApiKey)
	if err != nil {
		return nil, err
	}
	// Either it's a non-private or owned by the subject
	result.Where(fmt.Sprintf("COALESCE(%v.%v, FALSE) = false OR %s.%s = ?",
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnPrivate,
		dbtables.TableNameAssetsMeta, dbtables.AssetMetaColumnCreatedBy), auth.SubjectIdOf(ctx))

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	scanRows, scanErr := result.Scan(&combinedRecord)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_NGC_API_KEY).WithDetailsFromDbResult(scanRows, scanErr)
	}

	if options.MaskCredentials {
		ngcApiKeySecretReduction(&combinedRecord.NgcApiKeySpec)
	}

	apiAssetMeta := combinedRecord.ToApiAssetMeta()

	response := api.NgcApiKey{
		Meta:        apiAssetMeta,
		Spec:        combinedRecord.NgcApiKeySpec,
		ClusterInfo: getNgcApiKeyClusterInfo(&apiAssetMeta, nil),
	}

	return &response, nil
}

func ListNgcApiKeys(ctx context.Context, queryParams *query_types.QueryParams, options CredentialsOptions) (*api.NgcApiKeyListResponse, *api.Error) {
	combinedList := make([]dbmodel.NgcApiKeyFullRecord, 0)

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_NGC_API_KEY,
		},
	}
	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsNgcApiKey)
	if err != nil {
		return nil, err
	}

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list ngc api keys", scanErr)
	}

	var response api.NgcApiKeyListResponse

	for _, row := range combinedList {
		apiAssetMeta := row.ToApiAssetMeta()

		if options.MaskCredentials {
			ngcApiKeySecretReduction(&row.NgcApiKeySpec)
		}
		response.Entries = append(response.Entries, api.NgcApiKey{
			Meta:        apiAssetMeta,
			Spec:        row.NgcApiKeySpec,
			ClusterInfo: getNgcApiKeyClusterInfo(&apiAssetMeta, nil),
		})
	}

	return &response, nil
}

func UpdateNgcApiKey(ctx context.Context, id string, request api.NgcApiKeyUpdateRequest, keyId int32) (*api.NgcApiKey, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	_, rowsAffected, apiErr, dbErr := updateCredentialAssetMeta(ctx, tx, id, request.Meta)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_NGC_API_KEY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	//
	//   for credentials, it is possible that spec is nil, and in this case it will only rename without changing the spec
	//
	/* RUN-20970: At this point we are not supporting spec update -> request.Spec == nil */
	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_NGC_API_KEY).WithDetails(err)
	}
	committed = true

	// no need to notify asset modified because spec remained unchanged

	return GetNgcApiKeyById(ctx, id, CredentialsOptions{MaskCredentials: true})
}

func DeleteNgcApiKey(ctx context.Context, assetMeta *api.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	inuse, err := checkNgcApiKeyInUse(ctx, tx, assetMeta.Id)
	if err != nil {
		return err
	}
	if inuse {
		return api.NewError(http.StatusPreconditionFailed, "Cannot delete ngc api key since it is being used by other resources", nil)
	}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, assetMeta.Id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_NGC_API_KEY).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == HardDelete {
		recId := dbmodel.NgcApiKeyRecord{AssetId: assetMeta.Id}

		delRows, delErr := tx.Where(&recId).Delete(&recId)
		if delErr != nil || delRows == 0 {
			return database.Error(database.DELETE, api.ASSETKIND_NGC_API_KEY).WithDetailsFromDbResult(delRows, delErr)
		}
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_NGC_API_KEY).WithDetails(err)
	}
	committed = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfAssetMeta(assetMeta))

	return nil
}

func ngcApiKeySecretReduction(ngcRecord *api.NgcApiKeySpec) {
	//  secret reduction from the response
	ngcRecord.NgcApiKey = encryption.SecretMask
}

func checkNgcApiKeyInUse(ctx context.Context, tx orm.DB, Id string) (bool, *api.Error) {
	return false, nil
}

func getNgcApiKeyClusterInfo(assetMeta *api.AssetMeta, existingSecret *string) *api.ClusterInfo {
	return &api.ClusterInfo{
		Resources: []api.ClusterResourceInfo{
			{
				Name: convutils.Ref(strings.ToLower(fmt.Sprintf("%s-%s-%s", api.ASSETKIND_GENERIC_SECRET, composer.NgcApiKeyGenericSecretPrefix, assetMeta.Name))),
			},
			{
				Name: convutils.Ref(strings.ToLower(fmt.Sprintf("%s-%s-%s", api.ASSETKIND_DOCKER_REGISTRY, composer.NgcApiKeyDockerRegistryPrefix, assetMeta.Name))),
			},
		},
	}
}
