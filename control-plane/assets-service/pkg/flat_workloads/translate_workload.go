/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/jinzhu/copier"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	utils2 "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/utils"
	workloadsService "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	policyconstants "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/constants"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var securityFields = jsonutils.NewJsonCracker(workloadsService.WorkspaceSpecSpecAllOfSecurity{}).AllNames()

var (
	workloadCracker            = jsonutils.NewJsonCracker(workloadsService.SupersetSpec{})
	workloadComputeField       = workloadCracker.NameOf(&workloadCracker.Struct().Compute)
	workloadStorageField       = workloadCracker.NameOf(&workloadCracker.Struct().Storage)
	workloadSecurityField      = workloadCracker.NameOf(&workloadCracker.Struct().Security)
	workloadPortsField         = workloadCracker.NameOf(&workloadCracker.Struct().Ports)
	workloadServingPortField   = workloadCracker.NameOf(&workloadCracker.Struct().ServingPort)
	workloadExposedUrlField    = workloadCracker.NameOf(&workloadCracker.Struct().ExposedUrls)
	workloadRelatedUrlField    = workloadCracker.NameOf(&workloadCracker.Struct().RelatedUrls)
	workloadPriorityClassField = workloadCracker.NameOf(&workloadCracker.Struct().PriorityClass)
	workloadAnnotationsField   = workloadCracker.NameOf(&workloadCracker.Struct().Annotations)
	workloadAutoScalingField   = workloadCracker.NameOf(&workloadCracker.Struct().Autoscaling)
)

var (
	gitSecretRefCracker         = jsonutils.NewJsonCracker(workloadsService.GitSecretRef{})
	gitSecretRefFieldName       = gitSecretRefCracker.NameOf(&gitSecretRefCracker.Struct().Name)
	gitSecretRefFieldAuthMethod = gitSecretRefCracker.NameOf(&gitSecretRefCracker.Struct().AuthenticationMethod)
	gitSecretRefFieldKeyUser    = gitSecretRefCracker.NameOf(&gitSecretRefCracker.Struct().SecretKeyOfUser)
	gitSecretRefFieldKeyPass    = gitSecretRefCracker.NameOf(&gitSecretRefCracker.Struct().SecretKeyOfPassword)
)

var (
	accessKeyCracker        = jsonutils.NewJsonCracker(workloadsService.S3AccessKey{})
	accessKeyFieldSecret    = accessKeyCracker.NameOf(&accessKeyCracker.Struct().AccessKeySecret)
	accessKeyFieldAccessKey = accessKeyCracker.NameOf(&accessKeyCracker.Struct().SecretKeyOfAccessKeyId)
	accessKeyFieldSecretKey = accessKeyCracker.NameOf(&accessKeyCracker.Struct().SecretKeyOfSecretKey)
)

var (
	secretVolumeCracker    = jsonutils.NewJsonCracker(workloadsService.Secret{})
	secretVolumSecretField = secretVolumeCracker.NameOf(&secretVolumeCracker.Struct().Secret)
)

var (
	hostPathSpecCracker       = jsonutils.NewJsonCracker(api.HostPathSpec{})
	s3AssetSpecCracker        = jsonutils.NewJsonCracker(api.S3AssetSpec{})
	gitAssetSpecCracker       = jsonutils.NewJsonCracker(api.GitAssetSpec{})
	nfsAssetSpecCracker       = jsonutils.NewJsonCracker(api.NfsAssetSpec{})
	pvcAssetSpecCracker       = jsonutils.NewJsonCracker(api.PvcAssetSpec{})
	secretAssetSpecCracker    = jsonutils.NewJsonCracker(api.SecretAssetSpec{})
	configMapAssetSpecCracker = jsonutils.NewJsonCracker(api.ConfigMapAssetSpec{})
)

var (
	commonItemizedFieldsCracker = jsonutils.NewJsonCracker(workloadsService.SupersetSpec{})
	EnvironmentVariablesField   = commonItemizedFieldsCracker.NameOf(&commonItemizedFieldsCracker.Struct().EnvironmentVariables)
)

var assetKindToContainerPathKey = map[api.AssetKind]string{
	api.ASSETKIND_HOST_PATH:     hostPathSpecCracker.NameOf(&hostPathSpecCracker.Struct().MountPath),
	api.ASSETKIND_NFS:           nfsAssetSpecCracker.NameOf(&nfsAssetSpecCracker.Struct().MountPath),
	api.ASSETKIND_PVC:           pvcAssetSpecCracker.NameOf(&pvcAssetSpecCracker.Struct().Path),
	api.ASSETKIND_S3:            s3AssetSpecCracker.NameOf(&s3AssetSpecCracker.Struct().Path),
	api.ASSETKIND_GIT:           gitAssetSpecCracker.NameOf(&gitAssetSpecCracker.Struct().Path),
	api.ASSETKIND_CONFIG_MAP:    configMapAssetSpecCracker.NameOf(&configMapAssetSpecCracker.Struct().MountPath),
	api.ASSETKIND_SECRET_VOLUME: secretAssetSpecCracker.NameOf(&secretAssetSpecCracker.Struct().MountPath),
}

const (
	environmentFieldName = "environment"

	// NOTE! The following constants must be aligned with assets sync
	AssetSyncSecretDataKeyUser            = "User"
	AssetSyncSecretDataKeyPassword        = "Password"
	AssetSyncSecretDataKeyAccessKeyId     = "AccessKeyId"
	AssetSyncSecretDataKeySecretAccessKey = "SecretKey"
)

type TranslateFlatWorkloadOptions struct {
	DisableComplimentaryInfo bool
}

// TranslateToFlatWorkload translates payload of asset-based requests for creating a workload to the payload
// of the create-workload API of workloads service
func TranslateToFlatWorkload(ctx context.Context, workloadType string, workloadAssets []dbmodel.AssetsContainedInWorkload,
	credentialsAssets map[string]dbmodel.CredentialsAssetSpecs, specificRunParams *api.SpecificRunParams, dataSourceOverridesById map[string]*api.DataSourceOverrides,
	options *TranslateFlatWorkloadOptions,
) (map[string]interface{}, []byte, *api.Error) {
	if options == nil {
		options = &TranslateFlatWorkloadOptions{}
	}

	payloadMap, err := NewPayloadMap(ctx, specificRunParams, credentialsAssets)
	if err != nil {
		return nil, nil, err
	}

	var jsonErr error

	//
	//   the API of workload service is missing fields for sending assets-specific information
	//   we send this information encoded in the annotations in the complimentary asset info struct
	//
	complimentaryAssetInfo := &composer.ComplimentaryAssetInfo{
		Datasources: make(composer.MapAssetKindToResources),
	}

	//
	//   load the assets to the payload map
	//
	storageMap := make(map[string]interface{})

	for _, workloadAsset := range workloadAssets {
		kindAsString := string(workloadAsset.Kind)

		var assetAsJson map[string]interface{}

		storageAsset := true

		assetOverrides := dataSourceOverridesById[workloadAsset.Id]

		switch workloadAsset.Kind {
		case api.ASSETKIND_ENVIRONMENT:
			storageAsset = false

			//
			//   before loading the environment asset to the payload we need to complement it with
			//   the overridable fields from the specific run params
			//
			if specificRunParams != nil {
				ComplementEnvironmentWithSpecificEnv(workloadAsset.AssetsEnvironment, specificRunParams)
			}

			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsEnvironment.EnvironmentAssetSpec)

			//
			//   for environment, complimentary info needs to contain the id and name of the asset.
			//   no need to send details about the connections, they are taken from the ports/urls fields.
			//
			complimentaryAssetInfo.Environment = workloadsService.Environment{
				Name: workloadAsset.Name,
				Id:   workloadAsset.Id,
			}

			//
			//   translate the assets-based connections to the ports/urls list of the workload service API
			//
			patchErr := patchConnectionInfo(ctx, workloadType, payloadMap, &workloadAsset.AssetsEnvironment.EnvironmentAssetSpec, specificRunParams)
			if patchErr != nil {
				return nil, nil, patchErr
			}
			// translate the environment variables to the payload
			patchEnvironmentVariableCredentials(ctx, payloadMap, workloadAsset.AssetsEnvironment.EnvironmentVariables, credentialsAssets)
		case api.ASSETKIND_COMPUTE:
			storageAsset = false

			//
			//  aligning non-standard asset behavior of sending requestType portion for 0 or 1 GPU devices.
			//  this is rejected by workload service. to avoid this make sure the combination of
			//  fields is correct (requestType = portion, no portion requested, gpuDeviceRequest = 0/1 and then
			//  send the request without requestType, using only gpuDeviceRequest
			//
			requestTypePortion := workloadAsset.AssetsCompute.GpuRequestType != nil && *workloadAsset.AssetsCompute.GpuRequestType == api.GPUREQUESTTYPE_PORTION

			if requestTypePortion && workloadAsset.AssetsCompute.GpuPortionRequest == nil {
				if workloadAsset.AssetsCompute.GpuDevicesRequest == nil {
					return nil, nil, api.NewError(http.StatusBadRequest, "gpuMemoryPortionRequest (deprecated alias: gpuPortionRequest) must be provided for fractional workload of type 'portion'", nil)
				}
				if *workloadAsset.AssetsCompute.GpuDevicesRequest != 0 && *workloadAsset.AssetsCompute.GpuDevicesRequest != 1 {
					return nil, nil, api.NewError(http.StatusBadRequest, "gpuDeviceRequest can be 0 or 1 for workload of type 'portion'", nil)
				}
				workloadAsset.AssetsCompute.GpuRequestType = nil
			}

			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsCompute.ComputeAssetSpec)

		case api.ASSETKIND_HOST_PATH:
			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsHostPath.HostPathSpec)
			patchDatasourceOverrides(assetAsJson, workloadAsset.Kind, assetOverrides)

		case api.ASSETKIND_PVC:
			//
			//   pvc assets that are not auto-deleted must be handled like existing-pvcs. previously section creator was
			//   doing that, but now the translation to cluster is done by the workload composer that is not aware of
			//   asset being auto-deleted or not, so we have to align the assets to it
			//
			if workloadAsset.Kind == api.ASSETKIND_PVC && !convutils.BoolIsTrue(workloadAsset.AutoDelete) {
				var modifiedPvc api.PvcAssetSpec
				copier.Copy(&modifiedPvc, workloadAsset.AssetsPvc)

				modifiedPvc.ExistingPvc = convutils.Ref(true)
				modifiedPvc.ClaimInfo = nil

				assetAsJson, jsonErr = gojsonq.LoadJson(modifiedPvc)
			} else {
				assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsPvc)
			}

			patchDatasourceOverrides(assetAsJson, workloadAsset.Kind, assetOverrides)

		case api.ASSETKIND_NFS:
			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsNfs.NfsAssetSpec)
			patchDatasourceOverrides(assetAsJson, workloadAsset.Kind, assetOverrides)

		case api.ASSETKIND_GIT:
			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsGit.GitAssetSpec)
			patchDatasourceOverrides(assetAsJson, workloadAsset.Kind, assetOverrides)

			// the git asset may contain credentials asset id, but the workload service API needs
			// the explicit secret details from this asset
			patchSecretDetails(ctx, workloadAsset.Kind, dbmodel.GetGitCredentialId(workloadAsset.AssetsGit.GitAssetSpec), credentialsAssets, assetAsJson)

		case api.ASSETKIND_S3:
			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsS3.S3AssetSpec)
			patchDatasourceOverrides(assetAsJson, workloadAsset.Kind, assetOverrides)

			// the s3 asset may contain credentials asset id, but the workload service API needs
			// the explicit secret details from this asset
			patchSecretDetails(ctx, workloadAsset.Kind, workloadAsset.AssetsS3.AccessKeyAssetId, credentialsAssets, assetAsJson)

		case api.ASSETKIND_CONFIG_MAP:
			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsConfigMap.ConfigMapAssetSpec)
			patchDatasourceOverrides(assetAsJson, workloadAsset.Kind, assetOverrides)

		case api.ASSETKIND_SECRET_VOLUME:
			assetAsJson, jsonErr = gojsonq.LoadJson(workloadAsset.AssetsSecretVolume.SecretAssetSpec)
			patchDatasourceOverrides(assetAsJson, workloadAsset.Kind, assetOverrides)
			patchSecretDetails(ctx, workloadAsset.Kind, workloadAsset.AssetsSecretVolume.CredentialAssetId, credentialsAssets, assetAsJson)

		default:
			return nil, nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("unsupported asset kind %s", workloadAsset.Kind), nil)
		}

		if jsonErr != nil {
			return nil, nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf(
				"failed to parse %s asset %s", workloadAsset.Kind, workloadAsset.Id), jsonErr)
		}

		if storageAsset {
			workloadStorageField, datasourceKnown := composer_v2.MapAssetKindToWorkloadSpec[kindAsString]
			if !datasourceKnown {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("missing mapping from assetKind %s to workload spec", kindAsString)
				return nil, nil, api.NewError(http.StatusInternalServerError, "failed to compose workload", nil)
			}

			//
			//   add the asset to the list of datasource(s)
			//
			if storageMap[workloadStorageField] == nil {
				storageMap[workloadStorageField] = []map[string]interface{}{assetAsJson}
			} else {
				storageMap[workloadStorageField] = append(storageMap[workloadStorageField].([]map[string]interface{}), assetAsJson)
			}

			//
			// add also to the notification and complimentary info, except if it is auto-deleted asset
			// which is a volume and not datasource
			//
			if convutils.BoolIsFalse(workloadAsset.AutoDelete) {
				//
				//   push the details of the datasource to the complimentary info struct
				//
				if complimentaryAssetInfo.Datasources[kindAsString] == nil {
					complimentaryAssetInfo.Datasources[kindAsString] = []workloadsService.Datasource{}
				}

				complimentaryAssetInfo.Datasources[kindAsString] = append(complimentaryAssetInfo.Datasources[kindAsString], workloadsService.Datasource{
					Type: &kindAsString,
					Name: workloadAsset.Name,
					Id:   workloadAsset.Id,
				})
			}
		} else {
			payloadMap[kindAsString] = assetAsJson
		}
	}

	payloadMap[workloadStorageField] = storageMap

	//
	//   re-arrange the fields to the payload of the API
	//
	RearrangePayload(payloadMap)

	//
	//   translate allowOverQuota to priority class
	//
	if specificRunParams != nil {
		handleOverQuota(specificRunParams.AllowOverQuota, workloadType, payloadMap)
	}

	//
	//   translate autoscaling params
	//
	if specificRunParams != nil {
		handleAutoScaling(specificRunParams.AutoScaling, payloadMap)
	}

	//
	//   encode complimentary info to the annotations of the workload
	//
	if !options.DisableComplimentaryInfo {
		err = storeComplimentaryInfo(complimentaryAssetInfo, payloadMap)
		if err != nil {
			return nil, nil, err
		}
	}

	payloadAsBytes, jsonErr := json.Marshal(payloadMap)
	if jsonErr != nil {
		return nil, nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to compose workload payload: %v", jsonErr), nil)
	}

	return payloadMap, payloadAsBytes, nil
}

func NewPayloadMap(ctx context.Context, specificRunParams *api.SpecificRunParams, credentialsAssets map[string]dbmodel.CredentialsAssetSpecs) (map[string]interface{}, *api.Error) {
	//
	//   we create the payload of the API in a JSON map because it is easier to manipulate than go structures.
	//   the first thing we load is the specific run params. we load all of them, although only the non-overridable
	//   fields will eventually remain in the payload. the overridable fields are treated separately (see below)
	//
	var payloadMap map[string]interface{}

	if specificRunParams != nil {
		var runParamsErr error

		payloadMap, runParamsErr = gojsonq.LoadJson(specificRunParams)
		if runParamsErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to parse workload fields", runParamsErr)
		}
		patchEnvironmentVariableCredentials(ctx, payloadMap, specificRunParams.EnvironmentVariables, credentialsAssets)
	} else {
		payloadMap = make(map[string]interface{})
	}

	//
	//   prepare empty sections for security and compute. we need it for the API of the new payload
	//   and there are no such sections in the assets based payload
	//
	payloadMap[workloadSecurityField] = make(map[string]interface{})
	payloadMap[workloadComputeField] = make(map[string]interface{})
	payloadMap[workloadStorageField] = make(map[string]interface{})

	return payloadMap, nil
}

func RearrangePayload(payloadMap map[string]interface{}) {
	//
	//   move all the security fields which we loaded from the environment asset to the security section
	//
	moveFieldsByName(convutils.MapMe(payloadMap[environmentFieldName]), convutils.MapMe(payloadMap[workloadSecurityField]), securityFields, false /*do not overwrite*/)

	//
	//   move the rest of the fields from the environment asset to the top of the payload
	//
	moveSection(payloadMap, environmentFieldName, payloadMap, false)
}

// load the secret details from a credential asset and store them in the payload of the datasource which use them
func patchSecretDetails(ctx context.Context, targetKind api.AssetKind, credentialsAssetId *string, credentialsAssets map[string]dbmodel.CredentialsAssetSpecs,
	assetAsJson map[string]interface{},
) {
	//
	//   if no credential asset specified, this is a public git/s3 so no need for secret info
	//
	if credentialsAssetId == nil {
		return
	}

	//
	//  get the details of the credentials from the map of credentials asset. we do a validation to make
	//  sure the asset is on the map, but it is only theoretical case because we already took care for this
	//  in the function that builds this map
	//
	credentialAsset, found := credentialsAssets[*credentialsAssetId]
	if !found {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("credentials asset %s unknown", *credentialsAssetId)
		return
	}

	switch targetKind {
	case api.ASSETKIND_SECRET_VOLUME:
		secretName := credentialAsset.SecretName()
		assetAsJson[secretVolumSecretField] = *secretName
		return
	case api.ASSETKIND_GIT:
		secretRefAsJson := make(map[string]interface{})
		assetAsJson["secretRef"] = secretRefAsJson
		secretRefAsJson[gitSecretRefFieldName] = *credentialAsset.SecretName()

		// git data source supports two types of credentials:
		// 1. password - for git repositories that require username and password
		// 2. generic secret - that holds an SSH key for git repositories that require SSH authentication
		if credentialAsset.AssetsPassword != nil {
			secretRefAsJson[gitSecretRefFieldAuthMethod] = workloadsService.GITAUTHENTICATIONMETHOD_PASSWORD
			passwordSpec := credentialAsset.AssetsPassword.PasswordSpec
			if convutils.IsEmptyOrNull(passwordSpec.ExistingSecretName) {
				secretRefAsJson[gitSecretRefFieldKeyUser] = AssetSyncSecretDataKeyUser
				secretRefAsJson[gitSecretRefFieldKeyPass] = AssetSyncSecretDataKeyPassword
			} else {
				sc := runai_config.Get[*config.SecretConfig]()
				secretRefAsJson[gitSecretRefFieldKeyUser] = sc.DataKeyUser
				secretRefAsJson[gitSecretRefFieldKeyPass] = sc.DataKeyPassword
			}
			return
		}
		if credentialAsset.AssetsGenericSecret != nil {
			secretRefAsJson[gitSecretRefFieldAuthMethod] = workloadsService.GITAUTHENTICATIONMETHOD_SSH_KEY
			return
		}
		return
	case api.ASSETKIND_S3:
		if credentialAsset.AssetsAccessKey != nil {
			accessKeySrc := credentialAsset.AssetsAccessKey.AccessKeySpec

			if convutils.IsEmptyOrNull(accessKeySrc.ExistingSecretName) {
				assetAsJson[accessKeyFieldSecret] = GetSecretName(credentialAsset.Kind, credentialAsset.Name)
				assetAsJson[accessKeyFieldAccessKey] = AssetSyncSecretDataKeyAccessKeyId
				assetAsJson[accessKeyFieldSecretKey] = AssetSyncSecretDataKeySecretAccessKey
			} else {
				sc := runai_config.Get[*config.SecretConfig]()
				assetAsJson[accessKeyFieldSecret] = accessKeySrc.ExistingSecretName
				assetAsJson[accessKeyFieldAccessKey] = sc.DataKeyAccessKeyId
				assetAsJson[accessKeyFieldSecretKey] = sc.DataKeySecretAccessKey
			}

			return
		}
	}
}

// translate connections from the environment asset to port/exposedUrl/relatedUrl
func patchConnectionInfo(ctx context.Context, workloadType string, payloadMap map[string]interface{},
	environmentSpec *api.EnvironmentAssetSpec, specificRunParams *api.SpecificRunParams,
) *api.Error {
	if environmentSpec.Connections == nil {
		return nil
	}

	//
	//  search for the connection info in the specific run params
	//
	connSpecificEnvInfo := map[string]api.SpecificRunConnectionInfo{}
	if specificRunParams != nil && specificRunParams.Connections != nil {
		for _, conn := range *specificRunParams.Connections {
			connSpecificEnvInfo[conn.Name] = conn
		}
	}

	//
	//  prepare the arrays containing the connections in the payloads
	//   ports - container ports exposed to host ports
	//   exposedUrls - container ports exposed to urls
	//   relatedUrls - urls to external servers (such as weights and bias)
	//
	ports := make([]workloadsService.Port, 0)
	exposedUrls := make([]workloadsService.ExposedUrl, 0)
	relatedUrls := make([]workloadsService.RelatedUrl, 0)
	var servingPort *workloadsService.ServingPort

	for _, connection := range environmentSpec.Connections {
		connInfo, connInfoFound := connSpecificEnvInfo[connection.Name]

		var connInfoPtr *api.SpecificRunConnectionInfo
		if connInfoFound {
			connInfoPtr = &connInfo
		}

		if connection.InternalToolInfo != nil {
			switch connection.InternalToolInfo.ConnectionType {
			case api.INTERNALCONNECTIONTYPE_EXTERNAL_URL:
				url := composeExposedUrl(connection.Name, connection.InternalToolInfo, connInfoPtr)
				exposedUrls = append(exposedUrls, url)

			case api.INTERNALCONNECTIONTYPE_SERVING_PORT:
				switch workloadType {
				case auth.WorkloadTypeInference, string(workloadsService.WORKLOADTYPE_DISTRIBUTED_INFERENCE):
					var err *api.Error
					servingPort, err = composeServingPort(connection.InternalToolInfo, specificRunParams)
					if err != nil {
						return err
					}
				}

			default:
				port, err := composePort(connection.Name, connection.InternalToolInfo, connInfoPtr)
				if err != nil {
					return err
				}
				ports = append(ports, port)
			}
		} else {
			relatedUrls = append(relatedUrls, workloadsService.RelatedUrl{
				Name: *workloadsService.NewNullableString(&connection.Name),
				Type: *workloadsService.NewNullableString(convutils.Ref(string(connection.ExternalToolInfo.ToolType))),
				Url:  *workloadsService.NewNullableString(&connection.ExternalToolInfo.ExternalUrl),
			})
		}
	}

	portsAndUrls := map[string]interface{}{
		workloadPortsField:      ports,
		workloadExposedUrlField: exposedUrls,
		workloadRelatedUrlField: relatedUrls,
	}

	if servingPort != nil {
		portsAndUrls[workloadServingPortField] = servingPort
	}

	portsAndUrlsAsJson, _ := gojsonq.LoadJson(portsAndUrls)

	payloadMap[workloadPortsField] = portsAndUrlsAsJson[workloadPortsField]
	payloadMap[workloadExposedUrlField] = portsAndUrlsAsJson[workloadExposedUrlField]
	payloadMap[workloadRelatedUrlField] = portsAndUrlsAsJson[workloadRelatedUrlField]

	if servingPort != nil {
		payloadMap[workloadServingPortField] = portsAndUrlsAsJson[workloadServingPortField]
	}

	return nil
}

func patchDatasourceOverrides(dataSourceAsJson map[string]interface{}, targetKind api.AssetKind, dataSourceOverrides *api.DataSourceOverrides) {
	if dataSourceOverrides == nil {
		return
	}

	if dataSourceOverrides.ContainerPath != nil {
		keyName := assetKindToContainerPathKey[targetKind]
		dataSourceAsJson[keyName] = *dataSourceOverrides.ContainerPath
	}
}

func composeExposedUrl(name string, connection *api.InternalToolInfo, specificEnvInfo *api.SpecificRunConnectionInfo) workloadsService.ExposedUrl {
	customUrl := connection.ExternalUrlInfo != nil && connection.ExternalUrlInfo.IsCustomUrl

	result := workloadsService.ExposedUrl{
		Container: *workloadsService.NewNullableInt32(&connection.ContainerPort),
		ToolType:  *workloadsService.NewNullableString(convutils.Ref(string(connection.ToolType))),
		ToolName:  *workloadsService.NewNullableString(&name),
	}

	if specificEnvInfo != nil && specificEnvInfo.AuthorizedUsers != nil {
		result.AuthorizedUsers = *specificEnvInfo.AuthorizedUsers
	}

	if specificEnvInfo != nil && specificEnvInfo.AuthorizedGroups != nil {
		result.AuthorizedGroups = *specificEnvInfo.AuthorizedGroups
	}

	if customUrl {
		if specificEnvInfo != nil && utils2.StringHasValue(specificEnvInfo.ExternalUrl) {
			result.Url = *workloadsService.NewNullableString(specificEnvInfo.ExternalUrl)
		} else if utils2.StringHasValue(connection.ExternalUrlInfo.ExternalUrl) {
			result.Url = *workloadsService.NewNullableString(connection.ExternalUrlInfo.ExternalUrl)
		}
	}

	return result
}

func composePort(name string, connection *api.InternalToolInfo, specificEnvInfo *api.SpecificRunConnectionInfo) (workloadsService.Port, *api.Error) {
	port := workloadsService.Port{
		Container: *workloadsService.NewNullableInt32(&connection.ContainerPort),
		ToolType:  *workloadsService.NewNullableString(convutils.Ref(string(connection.ToolType))),
		ToolName:  *workloadsService.NewNullableString(&name),
	}

	switch connection.ConnectionType {
	case api.INTERNALCONNECTIONTYPE_NODE_PORT:
		port.ServiceType = *workloadsService.NewNullablePortServiceType(convutils.Ref(workloadsService.PORTSERVICETYPE_NODE_PORT))

		isCustomPort := connection.NodePortInfo != nil && connection.NodePortInfo.IsCustomPort

		if isCustomPort && specificEnvInfo != nil && specificEnvInfo.NodePort != nil {
			providedPort := convutils.Deref(specificEnvInfo.NodePort, 0)
			port.External = *workloadsService.NewNullableInt32(&providedPort)
		}
	case api.INTERNALCONNECTIONTYPE_LOAD_BALANCER:
		port.ServiceType = *workloadsService.NewNullablePortServiceType(convutils.Ref(workloadsService.PORTSERVICETYPE_LOAD_BALANCER))
	}

	return port, nil
}

func composeServingPort(connection *api.InternalToolInfo, specificRunParams *api.SpecificRunParams) (*workloadsService.ServingPort, *api.Error) {
	var protocol workloadsService.NullableServingPortProtocol
	if connection.ServingPortInfo != nil {
		protocol.Set((*workloadsService.ServingPortProtocol)(connection.ServingPortInfo.Protocol))
	}

	servingPort := &workloadsService.ServingPort{
		Container: *workloadsService.NewNullableInt32(convutils.Ref(connection.ContainerPort)),
		Protocol:  protocol,
	}

	if specificRunParams != nil && specificRunParams.ServingPortAccess != nil {
		spa := specificRunParams.ServingPortAccess
		if spa.AuthorizationType != nil {
			authStr := string(*spa.AuthorizationType)
			enumVal, err := workloadsService.NewServingPortAccessAuthorizationTypeEnumFromValue(authStr)
			if err != nil {
				return nil, api.NewError(http.StatusBadRequest, "failed to compose serving port", err)
			}
			servingPort.AuthorizationType = *workloadsService.NewNullableServingPortAccessAuthorizationTypeEnum(enumVal)
		}

		if spa.AuthorizedUsers != nil {
			servingPort.AuthorizedUsers = *spa.AuthorizedUsers
		}

		if spa.AuthorizedGroups != nil {
			servingPort.AuthorizedGroups = *spa.AuthorizedGroups
		}

		if spa.ClusterLocalAccessOnly != nil {
			servingPort.ClusterLocalAccessOnly = *workloadsService.NewNullableBool(spa.ClusterLocalAccessOnly)
		}
	}

	return servingPort, nil
}

// receives a list of field names and move them from one place in the payload to another
// for examples receive the names of all the security fields and move them from the top
// of the payload to the security inner-section
func moveFieldsByName(srcMap map[string]interface{}, destMap map[string]interface{}, fieldNames []string, overwrite bool) {
	if len(srcMap) == 0 || len(fieldNames) == 0 {
		return
	}

	for _, column := range fieldNames {
		srcValue, srcExist := srcMap[column]
		if !srcExist {
			continue
		}
		_, destExist := destMap[column]
		if !destExist || overwrite {
			destMap[column] = srcValue
		}

		delete(srcMap, column)
	}
}

// move an entire section from one place in the payload to another place.
// option to overwrite fields if they already exist in the destination section
func moveSection(srcMap map[string]interface{}, section string, destMap map[string]interface{}, overwrite bool) {
	srcSection := convutils.MapMe(srcMap[section])

	if len(srcSection) == 0 {
		return
	}

	for column, value := range srcSection {
		_, destExist := destMap[column]
		if !destExist || overwrite {
			destMap[column] = value
		}

		delete(srcSection, column)
	}

	delete(srcMap, section)
}

func handleOverQuota(allowOverQuota *bool, workloadType string, destMap map[string]interface{}) {
	if allowOverQuota != nil && workloadType == auth.WorkloadTypeWorkspace {
		if *allowOverQuota {
			destMap[workloadPriorityClassField] = policyconstants.PriorityClassInteractivePreemptible
		} else {
			destMap[workloadPriorityClassField] = policyconstants.PriorityClassBuild
		}
	}
}

func handleAutoScaling(autoScaling *api.SpecificRunAutoScalingAutoScaling, destMap map[string]interface{}) {
	if autoScaling == nil {
		return
	}

	as := workloadsService.AutoScaling{}

	if autoScaling.ThresholdMetric != nil {
		as.Metric = *workloadsService.NewNullableString(autoScaling.ThresholdMetric)
	}

	if autoScaling.ThresholdValue != nil {
		as.MetricThreshold = *workloadsService.NewNullableInt32(autoScaling.ThresholdValue)
	}

	if autoScaling.MetricThresholdPercentage != nil {
		as.MetricThresholdPercentage = *workloadsService.NewNullableFloat32(autoScaling.MetricThresholdPercentage)
	}

	if autoScaling.ConcurrencyHardLimit != nil {
		as.ConcurrencyHardLimit = *workloadsService.NewNullableInt32(autoScaling.ConcurrencyHardLimit)
	}

	if autoScaling.MinReplicas != nil {
		as.MinReplicas = *workloadsService.NewNullableInt32(autoScaling.MinReplicas)
	}

	if autoScaling.MaxReplicas != nil {
		as.MaxReplicas = *workloadsService.NewNullableInt32(autoScaling.MaxReplicas)
	}

	if autoScaling.InitialReplicas != nil {
		as.InitialReplicas = *workloadsService.NewNullableInt32(autoScaling.InitialReplicas)
	}

	if autoScaling.ActivationReplicas != nil {
		as.ActivationReplicas = *workloadsService.NewNullableInt32(autoScaling.ActivationReplicas)
	}

	if autoScaling.ScaleToZeroRetentionSeconds != nil {
		as.ScaleToZeroRetentionSeconds = *workloadsService.NewNullableInt32(autoScaling.ScaleToZeroRetentionSeconds)
	}

	if autoScaling.ScaleDownDelaySeconds != nil {
		as.ScaleDownDelaySeconds = *workloadsService.NewNullableInt32(autoScaling.ScaleDownDelaySeconds)
	}

	if autoScaling.InitializationTimeoutSeconds != nil {
		as.InitializationTimeoutSeconds = *workloadsService.NewNullableInt32(autoScaling.InitializationTimeoutSeconds)
	}

	destMap[workloadAutoScalingField] = as
}

// store assets complimentary info in the workload annotation. the workload composer will use them
// for adding information to the payload of the created workload.
func storeComplimentaryInfo(info *composer.ComplimentaryAssetInfo, destMap map[string]interface{}) *api.Error {
	if info == nil {
		return nil
	}

	var annotations []interface{}

	if destMap[workloadAnnotationsField] != nil {
		annotations = destMap[workloadAnnotationsField].([]interface{})
	}

	assetsInfoAsBytes, err := json.Marshal(*info)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, "failed to create workload annotations", err)
	}

	assetsInfoEncoded := b64.StdEncoding.EncodeToString(assetsInfoAsBytes)

	annotations = append(annotations, workloadsService.Annotation{
		Name:  *workloadsService.NewNullableString(convutils.Ref(composer.ComplimentaryAssetInfoAnn)),
		Value: *workloadsService.NewNullableString(&assetsInfoEncoded),
	})

	destMap[workloadAnnotationsField] = annotations

	return nil
}

// GetSecretName -> NOTE! The following function is based on assets-sync
func GetSecretName(credentialsType api.AssetKind, credentialsName string) string {
	secretName := fmt.Sprintf("%s-%s", credentialsType, credentialsName)
	return strings.ToLower(secretName)
}

func patchEnvironmentVariableCredentials(ctx context.Context, payloadMap map[string]interface{}, environmentVariables *[]api.EnvironmentVariableOfAsset, credentialsAssets map[string]dbmodel.CredentialsAssetSpecs) {
	if environmentVariables == nil {
		return
	}

	newEnvironmentVariables := make([]workloadsService.EnvironmentVariable, 0, len(*environmentVariables))
	for _, envVar := range *environmentVariables {
		newVar := workloadsService.EnvironmentVariable{
			Name: *workloadsService.NewNullableString(envVar.Name),
		}

		if convutils.BoolIsTrue(envVar.Exclude) {
			newVar.Exclude = *workloadsService.NewNullableBool(envVar.Exclude)
			newEnvironmentVariables = append(newEnvironmentVariables, newVar)
			continue
		}

		if envVar.Description != nil {
			newVar.Description = *workloadsService.NewNullableString(envVar.Description)
		}

		if envVar.ConfigMap != nil {
			newVar.ConfigMap = *workloadsService.NewNullableEnvironmentVariableConfigMap(&workloadsService.EnvironmentVariableConfigMap{
				Name: *workloadsService.NewNullableString(envVar.ConfigMap.Name),
				Key:  *workloadsService.NewNullableString(envVar.ConfigMap.Key),
			})
		}

		if envVar.PodFieldRef != nil {
			newVar.PodFieldRef = *workloadsService.NewNullableEnvironmentVariablePodFieldReference(&workloadsService.EnvironmentVariablePodFieldReference{
				Path: *workloadsService.NewNullableString(envVar.PodFieldRef.Path),
			})
		}

		if envVar.Credential != nil {
			credentialAsset, found := credentialsAssets[*envVar.Credential.AssetId]
			if !found {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("credentials asset %v unknown", envVar.Credential.AssetId)
				continue
			}
			secret := workloadsService.NewEnvironmentVariableSecret()
			secret.SetName(*credentialAsset.SecretName())
			secret.SetKey(*envVar.Credential.Key)
			newVar.Secret = *workloadsService.NewNullableEnvironmentVariableSecret(secret)
		}

		if envVar.Value != nil {
			newVar.Value = *workloadsService.NewNullableString(envVar.Value)
		}

		newEnvironmentVariables = append(newEnvironmentVariables, newVar)
		continue
	}

	environment := map[string]interface{}{
		EnvironmentVariablesField: newEnvironmentVariables,
	}

	environmentVariablesAsJson, err := gojsonq.LoadJson(environment)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error: err %v", err)
		return
	}
	payloadMap[EnvironmentVariablesField] = environmentVariablesAsJson[EnvironmentVariablesField]
}
