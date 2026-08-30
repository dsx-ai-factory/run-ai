package dao

import (
	"github.com/iancoleman/strcase"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

// ListCredentialsAssets receives a list of credentials asset ids. it returns a map from credentials asset id to its spec
func ListCredentialsAssets(assetIds []string, db orm.DB) (map[string]dbmodel.CredentialsAssetSpecs, *api.Error) {
	credentialsAssets := make([]dbmodel.CredentialsAssetSpecs, 0)

	model := dbmodel.AssetMetaRec{}

	result := db.
		Model(model).
		Select("*").
		Order(&model.Id).
		Where("id IN ?", assetIds)

	for _, credentialsAssetTable := range dbtables.TableOfCredentials {
		result.Preload(strcase.ToCamel(credentialsAssetTable))
	}

	if findErr := result.Find(&credentialsAssets); findErr != nil {
		return nil, database.Error(database.LIST, "credentials").WithDetails(findErr)
	}

	idToAssetSpec := make(map[string]dbmodel.CredentialsAssetSpecs)
	for _, credentialsAsset := range credentialsAssets {
		idToAssetSpec[credentialsAsset.Id] = credentialsAsset
	}
	return idToAssetSpec, nil
}

// IdentifyCredentialAssets receives a list of assets and returns the ids
// of all the credentials assets referenced by those assets
func IdentifyCredentialAssets(assets []dbmodel.AssetsContainedInWorkload) []string {
	credentialIds := datastructs.Set[string]{}

	for _, asset := range assets {
		switch asset.Kind {
		case api.ASSETKIND_S3:
			if asset.AssetsS3 != nil && asset.AssetsS3.AccessKeyAssetId != nil {
				credentialIds.Add(*asset.AssetsS3.AccessKeyAssetId)
			}
		case api.ASSETKIND_GIT:
			if asset.AssetsGit != nil && asset.AssetsGit.CredentialId != nil {
				credentialIds.Add(*asset.AssetsGit.CredentialId)
			}
			if asset.AssetsGit != nil && asset.AssetsGit.PasswordAssetId != nil {
				credentialIds.Add(*asset.AssetsGit.PasswordAssetId)
			}
		case api.ASSETKIND_SECRET_VOLUME:
			if asset.AssetsSecretVolume != nil && asset.AssetsSecretVolume.CredentialAssetId != nil {
				credentialIds.Add(*asset.AssetsSecretVolume.CredentialAssetId)
			}
		case api.ASSETKIND_ENVIRONMENT:
			environmentCredentialIds := GetCredentialIdsFromEnvironmentVariables(asset.AssetsEnvironment.EnvironmentVariables)
			for _, id := range environmentCredentialIds {
				credentialIds.Add(id)
			}
		}
	}

	result := make([]string, 0)

	for id := range credentialIds {
		result = append(result, id)
	}

	return result
}

// GetCredentialIdsFromEnvironmentVariables returns credential-assets ids that are in use for environment variable
func GetCredentialIdsFromEnvironmentVariables(environmentVariables *[]api.EnvironmentVariableOfAsset) []string {
	if environmentVariables == nil {
		return nil
	}

	var assetIds []string
	for _, envVar := range *environmentVariables {
		if envVar.Credential == nil {
			continue
		}
		assetIds = append(assetIds, *envVar.Credential.AssetId)
	}

	return assetIds
}
