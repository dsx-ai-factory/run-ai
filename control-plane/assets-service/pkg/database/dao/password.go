package dao

import (
	"context"
	"net/http"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/encryption"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/utils"
)

const wherePrivateUseOnlyQuery = "private_use_only = ?"

type CredentialsOptions struct {
	MaskCredentials            bool
	IncludeDeleted             bool
	IncludePrivateUsePasswords bool
}

func CreatePassword(ctx context.Context, request api.PasswordCreationRequest, keyId int32, nameValidationQuery *database.ParametrizedQuery) (*api.Password, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	password, err := CreatePasswordWithTransaction(ctx, tx, request, keyId, false, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_PASSWORD).WithDetails(err)
	}
	committed = true

	return password, nil
}

func CreatePasswordWithTransaction(ctx context.Context, tx orm.DB, request api.PasswordCreationRequest,
	keyId int32, privateUseOnly bool, nameValidationQuery *database.ParametrizedQuery,
) (*api.Password, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_PASSWORD, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_PASSWORD, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_PASSWORD).WithDetails(err)
	}

	passwordRecord := dbmodel.PasswordRecord{
		AssetId:        id,
		PrivateUseOnly: privateUseOnly,
		PasswordSpec: api.PasswordSpec{
			ExistingSecretName: request.Spec.ExistingSecretName,
			User:               request.Spec.User,
			Password:           request.Spec.Password,
			KeyId:              keyId,
		},
	}

	_, err = tx.Create(passwordRecord)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_PASSWORD).WithDetails(err)
	}

	if utils.StringHasValue(passwordRecord.Password) {
		passwordRecord.Password = convutils.Ref(encryption.SecretMask)
	}

	return &api.Password{
		Meta:        *assetMeta,
		Spec:        passwordRecord.PasswordSpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(assetMeta, passwordRecord.ExistingSecretName),
	}, nil
}

func GetPasswordById(ctx context.Context, id string, options CredentialsOptions) (*api.Password, *api.Error) {
	var combinedRecord dbmodel.PasswordFullRecord

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_PASSWORD,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsPassword)
	if err != nil {
		return nil, err
	}

	if !options.IncludePrivateUsePasswords {
		result = result.Where(wherePrivateUseOnlyQuery, false)
	}

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	scanRows, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedRecord)

	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_PASSWORD).WithDetailsFromDbResult(scanRows, scanErr)
	}

	if utils.StringHasValue(combinedRecord.Password) {
		combinedRecord.Password = convutils.Ref(encryption.SecretMask)
	}

	response := api.Password{
		Meta:        combinedRecord.AssetMeta,
		Spec:        combinedRecord.PasswordSpec,
		ClusterInfo: dbmodel.GetSecretClusterInfo(&combinedRecord.AssetMeta, combinedRecord.ExistingSecretName),
	}
	return &response, nil
}

func ListPasswords(ctx context.Context, queryParams *query_types.QueryParams, options CredentialsOptions) (*api.PasswordListResponse, *api.Error) {
	combinedList := make([]dbmodel.PasswordFullRecord, 0)

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_PASSWORD,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsPassword)
	if err != nil {
		return nil, err
	}

	if !options.IncludePrivateUsePasswords {
		result = result.Where(wherePrivateUseOnlyQuery, false)
	}

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)

	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list password credentials", scanErr)
	}

	response := api.PasswordListResponse{
		Entries: make([]api.Password, 0),
	}

	for _, row := range combinedList {
		if options.MaskCredentials && utils.StringHasValue(row.Password) {
			row.Password = convutils.Ref(encryption.SecretMask)
		}

		response.Entries = append(response.Entries, api.Password{
			Meta:        row.AssetMeta,
			Spec:        row.PasswordSpec,
			ClusterInfo: dbmodel.GetSecretClusterInfo(&row.AssetMeta, row.ExistingSecretName),
		})
	}

	return &response, nil
}

func UpdatePassword(ctx context.Context, id string, request api.PasswordUpdateRequest, keyId int32) (*api.Password, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	password, err := updatePasswordInTransaction(ctx, tx, id, request, keyId)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_PASSWORD).WithDetails(err)
	}
	committed = true

	return password, nil
}

func updatePasswordInTransaction(ctx context.Context, tx orm.DB, id string, request api.PasswordUpdateRequest, keyId int32) (*api.Password, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateCredentialAssetMeta(ctx, tx, id, request.Meta)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_PASSWORD).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	//
	//   for credentials, it is possible that spec is nil, and in this case it will only rename without changing the spec
	//
	if true /* At this point we are not supporting spec update -> request.Spec == nil */ {
		updated, err := GetPasswordById(ctx, id, CredentialsOptions{})
		if err != nil {
			return nil, err
		}
		updated.Meta = assetMeta.ToApiAssetMeta()

		// no need to notify asset modify because spec was not updated

		return updated, nil
	}

	/* At this point we are not supporting spec update
	   recId := dbmodel.PasswordRecord{AssetId: id}
	   passwordUpdate := dbmodel.PasswordUpdate{}

	   if options.IncludePrivateUsePasswords {
	   	result = tx.Model(recId).Where(recId).First(&passwordUpdate)
	   } else {
	   	result = tx.Model(recId).Where(recId).First(&passwordUpdate, wherePrivateUseOnlyQuery, false)
	   }

	   if result.Error != nil || result.RowsAffected == 0 {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_PASSWORD).WithDetailsFromDb(result)
	   }

	   passwordUpdate.PasswordSpec = api.PasswordSpec{
	   	User:     request.Spec.User,
	   	Password: request.Spec.Password,
	   	KeyId:    keyId,
	   }
	   result = tx.Where(recId).Save(&passwordUpdate)
	   if result.Error != nil {
	   	return nil, database.Error(database.UPDATE, api.ASSETKIND_PASSWORD).WithDetailsFromDb(result)
	   }

	   if utils.StringHasValue(passwordUpdate.Password) {
	   	passwordUpdate.Password = convutils.Ref(encryption.SecretMask)
	   }

	   return &api.Password{
	   	Meta: assetMeta.ToApiAssetMeta(),
	   	Spec: passwordUpdate.PasswordSpec,
	   }, nil
	*/
	return nil, nil
}

func DeletePassword(ctx context.Context, assetMeta *api.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	err := deletePasswordInTransaction(ctx, tx, assetMeta.Id, deletionType, CredentialsOptions{})
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_PASSWORD).WithDetails(err)
	}
	committed = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfAssetMeta(assetMeta))

	return nil
}

func deletePasswordInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType, options CredentialsOptions) *api.Error {
	inuse, err := checkPasswordInUse(ctx, tx, id)
	if err != nil {
		return err
	}
	if inuse {
		return api.NewError(http.StatusPreconditionFailed, "Cannot delete password since it is being inuse", nil)
	}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_PASSWORD).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == HardDelete {
		recId := dbmodel.PasswordRecord{AssetId: id}

		if !options.IncludePrivateUsePasswords {
			recId.PrivateUseOnly = false
		}

		delRows, delErr := tx.Where(&recId).Delete(&recId)
		if delErr != nil || delRows == 0 {
			return database.Error(database.DELETE, api.ASSETKIND_PASSWORD).WithDetailsFromDbResult(delRows, delErr)
		}
	}

	return nil
}

func GetPasswordByIdVisible(ctx context.Context, id string) (*api.Password, *api.Error) {
	var combinedRecord dbmodel.PasswordFullRecord

	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_PASSWORD,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsPassword)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.
		Where(wherePrivateUseOnlyQuery, true).
		Order(&query.AssetMeta.Id).
		Scan(&combinedRecord)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_PASSWORD).WithDetailsFromDbResult(scanRows, scanErr)
	}

	return &api.Password{
		Meta: combinedRecord.AssetMeta,
		Spec: combinedRecord.PasswordSpec,
	}, nil
}

func checkPasswordInUse(ctx context.Context, tx orm.DB, passwordAssetId string) (bool, *api.Error) {
	count, err := countGitAssetsUsingPassword(tx, passwordAssetId)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	count, err = countSecretVolumeAssetsUsingCredentialId(tx, passwordAssetId)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	count, err = countEnvironmentAssetsUsingCredentialAsEnvVar(ctx, passwordAssetId)
	if err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	return false, nil
}
