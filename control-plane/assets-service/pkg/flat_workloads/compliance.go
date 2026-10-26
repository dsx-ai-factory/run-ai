/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/status_info"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"

	authApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	k8sTrackerApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

const (
	StorageInstanceName       = "name"
	ComplianceTasksInParallel = 10
	Distributed               = "Distributed"
	DistributedInference      = "DistributedInference"
	ClusterTBD                = ""

	datasourceMsg           = "This data source can't be selected because its status is \"Issues found\". Contact your administrator"
	dependentCredentialsMsg = "This data source can't be selected because its credentials' status is \"Issues found\". Contact your administrator"
	syncFailedMsg           = "This data source can't be selected because it could not synchronize to the cluster. Contact your administrator"
	syncApplyInProgressMsg  = "This data source can't be selected yet, because it is being synchronized to the cluster."
	syncDeleteInProgressMsg = "This data source can't be selected, because it is being deleted."
)

// if tenant is not working with policies, do not attempt to create them

var (
	defaultsJsonCracker = jsonutils.NewJsonCracker(policyApi.WorkspacePolicyDefaultsV2{})
	storageSection      = defaultsJsonCracker.NameOf(&defaultsJsonCracker.Struct().Storage)
	securitySection     = defaultsJsonCracker.NameOf(&defaultsJsonCracker.Struct().Security)
)

var (
	distributedJsonCracker          = jsonutils.NewJsonCracker(policyApi.DistributedPolicyRulesV2{})
	distributedInferenceJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedInferenceRulesV2{})
	sectionPerReplicaType           = map[workloadsApi.ReplicaType]string{
		workloadsApi.REPLICATYPE_MASTER: distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Master),
		workloadsApi.REPLICATYPE_WORKER: distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Worker),
		workloadsApi.REPLICATYPE_LEADER: distributedInferenceJsonCracker.NameOf(&distributedInferenceJsonCracker.Struct().Leader),
	}
)

// workloadClusterAssets are assets in workload that have an underlying cluster resource
var workloadClusterAssets = map[api.AssetKind]struct{}{
	api.ASSETKIND_PVC:           {},
	api.ASSETKIND_S3:            {},
	api.ASSETKIND_GIT:           {},
	api.ASSETKIND_CONFIG_MAP:    {},
	api.ASSETKIND_SECRET_VOLUME: {},
}

var (
	environmentSpecJsonCracker = jsonutils.NewJsonCracker(api.EnvironmentAssetSpec{})
	environmentSpecFieldNames  = datastructs.NewSet(environmentSpecJsonCracker.AllNames()...)
)

var (
	environmentOverridable = jsonutils.NewJsonCracker(api.OverridableSpecFields{})
	overridableFields      = datastructs.NewSet(environmentOverridable.AllNames()...)
)

var uuidLength = len(uuid.NewString())

const (
	valueForMandatoryError = "required field"
)

// ClusterResourceId represents a resource (pvc or secret) in the cluster.
type ClusterResourceId struct {
	ScopeType string
	ScopeId   string
	Name      string
	Type      string               // this one is relevant only for secrets
	Origin    k8sTrackerApi.Origin // assets resources or existing customer resources
}

type assetChanInfo struct {
	isImposed        bool
	patchingPtrs     response_patching.ResponsePatchingPtrs
	defaults         map[string]interface{}
	rules            map[string]interface{}
	defaultFromOrder policyApi.DefaultFromOrder
	workloadType     string
	replicaType      string
	policyInfo       dao.PolicyInfo
}

type ComplianceInfoPatcher struct {
	projectId    int32
	projectIdStr string
	workloadType string
	replicaType  string

	permittedScopes *auth.PermittedScopes

	assetClusterStatus *asset_cluster_status.AssetClusterStatus

	assetIdsToPatch datastructs.Set[string]
	patchingInfo    map[string]response_patching.ResponsePatchingPtrs
}

func NewComplianceInfoPatcher(ctx context.Context, projectId int32, workloadType string, replicaType string,
	patchingInfo map[string]response_patching.ResponsePatchingPtrs, permittedScopes *auth.PermittedScopes,
	assetClusterStatus *asset_cluster_status.AssetClusterStatus,
) (*ComplianceInfoPatcher, *api.Error) {
	//
	//   some sanity validations
	//
	if patchingInfo == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("patchingInfo nil, cannot patch compliance information")
		return nil, api.NewError(http.StatusInternalServerError, "patching invoked with no patchingInfo to update", nil)
	}

	if projectId <= 0 || workloadType == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("projectId and workloadType are required for compliance patching")
		return nil, api.NewError(http.StatusInternalServerError, "patching invoked with missing projectId nd workloadType", nil)
	}

	if (workloadType == Distributed || workloadType == DistributedInference) && replicaType == "" {
		return nil, api.NewValidationError(fmt.Sprintf("copmlyToReplicaType is mandatory when complyToWorkloadType=%s", workloadType))
	}

	if workloadType != Distributed && workloadType != DistributedInference && replicaType != "" {
		return nil, api.NewValidationError(fmt.Sprintf("copmlyToReplicaType is not valid when complyToWorkloadType=%s", workloadType))
	}

	result := &ComplianceInfoPatcher{
		projectId:    projectId,
		projectIdStr: strconv.Itoa(int(projectId)),
		workloadType: workloadType,
		replicaType:  replicaType,

		patchingInfo: patchingInfo,

		assetIdsToPatch: datastructs.NewSet[string](),

		permittedScopes:    permittedScopes,
		assetClusterStatus: assetClusterStatus,
	}

	//
	//   make a list of all the assets we want to validate and initialize their compliance to true,
	//   assuming everything is ok unless something found
	//
	for assetId := range patchingInfo {
		result.assetIdsToPatch.Add(assetId)

		*result.patchingInfo[assetId].ComplianceInfo = &api.ComplianceInfo{
			Compliance: true,
		}
	}

	return result, nil
}

func (cip *ComplianceInfoPatcher) PatchComplianceInfo(ctx context.Context) *api.Error {
	if len(cip.assetIdsToPatch) == 0 {
		return nil
	}

	//
	//   first thing to do - get the effective policy of this project/workload type
	//
	policiesClient := policies.Get().Client()
	effectivePolicy, err := policy_compliance.GetEffectivePolicyOfProject(ctx, policiesClient, cip.projectIdStr,
		workloadsApi.WorkloadType(cip.workloadType))
	if err == nil {
		policyErr := cip.policyBasedStatus(ctx, effectivePolicy)
		if policyErr != nil {
			return policyErr
		}
	} else if err.Code != http.StatusNotFound {
		// status not found == does not work with policy, this is ok. any other error - fail the compliance calculation
		return api.NewError(int(err.Code), err.Message, nil)
	}

	cip.patchDataVolumeCompliance(ctx)

	cip.patchPvcStorageClassCompliance(ctx)

	// cluster based status
	if len(cip.assetIdsToPatch) != 0 {
		clusterErr := cip.clusterBasedStatus(ctx)
		if clusterErr != nil {
			return clusterErr
		}
	}

	return nil
}

func (cip *ComplianceInfoPatcher) policyBasedStatus(ctx context.Context, effectivePolicy *policy_compliance.EffectivePolicyOfProject) *api.Error {
	policyInfo := dao.PolicyInfo{
		TenantId:     auth.TenantIdOf(ctx),
		WorkloadType: cip.workloadType,
		ReplicaType:  cip.replicaType,
		ProjectId:    cip.projectId,
		PolicyTime:   effectivePolicy.UpdatedAt,
	}

	//
	//  try to load information from the policy_compliance table first. any asset found there we will not have to
	//  validate from scratch
	//
	complianceResults, dbErr := dao.ListPolicyComplianceOfAssets(ctx, cip.assetIdsToPatch.Keys(), policyInfo)
	if dbErr != nil {
		return dbErr
	}

	//
	//   patch all the information that was loaded from the DB, and remove any patched asset from the map so after
	//   the loop only assets which were not patched from DB remain
	//
	imposedAssets := getImposedAssetIds(effectivePolicy)

	for _, complianceResult := range complianceResults {
		patchPtr := cip.patchingInfo[complianceResult.AssetId]

		*patchPtr.ComplianceInfo = &api.ComplianceInfo{
			Compliance: complianceResult.Aligned,
			Reason:     complianceResult.Reason,
			Imposed:    imposedAssets.Has(patchPtr.Meta.Id),
		}

		//
		//   if asset has been disqualified due to policy, no need to check other compliance issues
		//
		if !complianceResult.Aligned {
			delete(cip.assetIdsToPatch, complianceResult.AssetId)
		}
	}

	//
	//   if no remaining assets to validate it means that we were able to find all the information in the compliance table
	//   otherwise, we will have to calculate it
	//
	if len(cip.assetIdsToPatch) == 0 {
		return nil
	}

	//
	//   the rest we will do in parallel (like a thread pool) so create a wait group that will be responsible for
	//   orchestrating the validation all the remaining assets
	//
	var wg sync.WaitGroup
	wg.Add(len(cip.assetIdsToPatch))

	//
	//   create a channel that can buffer the number of assets that we intend to validate in parallel
	//
	assetsChan := make(chan assetChanInfo, ComplianceTasksInParallel)

	//
	//   spawn go routines to handle the compliance validation tasks
	//
	for i := 0; i < ComplianceTasksInParallel; i++ {
		go ValidateComplianceOfAssets(ctx, assetsChan, &wg)
	}

	//
	//   Send the tasks to the channel for validation
	//
	for assetId := range cip.assetIdsToPatch {
		assetChanInfo := assetChanInfo{
			patchingPtrs:     cip.patchingInfo[assetId],
			isImposed:        imposedAssets.Has(assetId),
			rules:            effectivePolicy.RulesV2,
			defaults:         effectivePolicy.DefaultsV2,
			defaultFromOrder: effectivePolicy.DefaultFromOrder,
			workloadType:     cip.workloadType,
			replicaType:      cip.replicaType,
			policyInfo:       policyInfo,
		}

		assetsChan <- assetChanInfo
	}

	//
	//   Wait for all the tasks to be completed
	//
	wg.Wait()

	cip.updateRemainingAssets()

	return nil
}

func (cip *ComplianceInfoPatcher) clusterBasedStatus(ctx context.Context) *api.Error {
	//
	//   before calculating the status based on k8sObjectsTracker, try to check the sync info table
	//   for sync-based status
	//
	syncBasedStatus, _ := status_info.LoadSyncStatus(ctx, cip.assetIdsToPatch, cip.patchingInfo)

	cip.updateRemainingAssets()

	if len(cip.assetIdsToPatch) == 0 {
		return nil
	}

	cip.assetClusterStatus.StoreSyncBasedInfo(syncBasedStatus)

	relevantAccessKinds := datastructs.NewSet[api.AssetKind]()
	for assetId := range cip.assetIdsToPatch {
		curPatch := cip.patchingInfo[assetId]

		// Only if the asset is base on a cluster load the status
		if _, ok := workloadClusterAssets[curPatch.Meta.Kind]; ok {
			relevantAccessKinds.Add(curPatch.Meta.Kind)
		}
	}

	for assetKind := range relevantAccessKinds {
		err := cip.assetClusterStatus.LoadAssetsStatus(ctx, assetKind)
		if err != nil {
			return err
		}
	}

	assetIdToClusterResourceStatus, err := cip.assetClusterStatus.GetAssetsStatus()
	if err != nil {
		return err
	}

	for assetId := range cip.assetIdsToPatch {
		assetInfo := cip.patchingInfo[assetId]

		assetClusterStatus := assetIdToClusterResourceStatus[assetId]
		if assetClusterStatus == nil {
			continue
		}

		if *assetInfo.ComplianceInfo != nil && !(*assetInfo.ComplianceInfo).Compliance {
			continue
		}
		// If the asset is marked for auto delete, we don't expect it to be in the cluster
		if convutils.BoolIsTrue(assetInfo.Meta.AutoDelete) {
			continue
		}

		// during delete, don't list issues just indicate asset being deleted
		if !assetInfo.Meta.DeletedAt.IsZero() {
			*assetInfo.ComplianceInfo = cip.SingleReasonCompliance(syncDeleteInProgressMsg)
			continue
		}

		switch assetClusterStatus.Status {
		case api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND, api.ASSETSYNCSTATUSENUM_NO_STATUS, api.ASSETSYNCSTATUSENUM_DELETION_FAILED:
			continue
		case api.ASSETSYNCSTATUSENUM_FAILED:
			*assetInfo.ComplianceInfo = cip.SingleReasonCompliance(syncFailedMsg)
			continue
		case api.ASSETSYNCSTATUSENUM_CREATING:
			*assetInfo.ComplianceInfo = cip.SingleReasonCompliance(syncApplyInProgressMsg)
			continue
		case api.ASSETSYNCSTATUSENUM_DELETING:
			*assetInfo.ComplianceInfo = cip.SingleReasonCompliance(syncDeleteInProgressMsg)
			continue
		}

		foundIssuesInCluster := false

		for _, issue := range assetClusterStatus.Issues {
			// If there is an issue in the specific scope checking compliance of,
			// or the asset is not even created in the cluster then we need to patch compliance false
			if string(issue.ScopeType) == string(authApi.SCOPETYPE_PROJECT) && issue.ScopeId == fmt.Sprintf("%d", cip.projectId) || issue.Issue == api.ASSETCLUSTERSTATUSISSUEENUM_NOT_CREATED {
				foundIssuesInCluster = true
			}
		}

		if !foundIssuesInCluster {
			continue
		}

		logger := logging.WithCtx(runaictx.LoggerHook(ctx))
		var message string
		switch assetInfo.Meta.Kind {
		case api.ASSETKIND_PVC, api.ASSETKIND_CONFIG_MAP:
			message = datasourceMsg
		case api.ASSETKIND_S3, api.ASSETKIND_GIT, api.ASSETKIND_SECRET_VOLUME:
			message = dependentCredentialsMsg
		default:
			logger.Errorf("unknown asset kind: %s", assetInfo.Meta.Kind)
		}

		*assetInfo.ComplianceInfo = cip.SingleReasonCompliance(message)
	}

	return nil
}

func (cip *ComplianceInfoPatcher) SingleReasonCompliance(message string) *api.ComplianceInfo {
	return &api.ComplianceInfo{
		Compliance: false,
		Reason: []api.ComplianceInfoReason{
			{
				Details: message,
			},
		},
	}
}

func (cip *ComplianceInfoPatcher) updateRemainingAssets() {
	//
	//  remove from the assets to patch set all those that were disqualified
	//
	for assetId := range cip.assetIdsToPatch {
		patchTo := cip.patchingInfo[assetId]
		if !(*patchTo.ComplianceInfo).Compliance {
			delete(cip.assetIdsToPatch, assetId)
		}
	}
}

func ValidateComplianceOfAssets(ctx context.Context, assetsChan chan assetChanInfo, wg *sync.WaitGroup) {
	for assetInfo := range assetsChan {
		validateComplianceOfAsset(ctx, assetInfo)
		wg.Done()
	}
}

func validateComplianceOfAsset(ctx context.Context, assetInfo assetChanInfo) {
	complianceErrors := make([]api.ComplianceInfoReason, 0)

	//
	//   if the asset can be used for a workload, we have to check if it compliant to the policy
	//
	if len(complianceErrors) == 0 {
		assetKind := assetInfo.patchingPtrs.Meta.Kind
		assetKindAsString := string(assetKind)

		var err *api.Error
		specAsJson, jpErr := gojsonq.LoadJson(assetInfo.patchingPtrs.Spec)
		if jpErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("fail to load spec of %s %s: %v",
				assetInfo.patchingPtrs.Meta.Kind, assetInfo.patchingPtrs.Meta.Name, jpErr)
			return
		}

		payloadMap, err := NewPayloadMap(ctx, nil, nil)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("fail to initialize payload map: %v", err)
			return
		}

		//
		//    since the code supports both get and list, there are two cases for specAsJson. consider
		//    hostPath as an example, there are two cases
		//          (1) GET: specAsJson directly contains the attributes of the hostPath
		//          (2) LIST: specAsJson contains "hostPath" and under this key there are the attributes of that hostPath
		//
		var assetAttributes interface{}
		if specAsJson[assetKindAsString] != nil {
			assetAttributes = specAsJson[assetKindAsString] // case 2, need to find the fields under HostPath
		} else {
			assetAttributes = specAsJson // case 1, specAsJson directly contains the fields
		}

		//
		//   determine the section to which we need to add the asset. environment/compute ->
		//   <top>.<asset-kind>, datasource -> <top>.storage.<assetKind> as array
		//
		if assetKind == api.ASSETKIND_ENVIRONMENT || assetKind == api.ASSETKIND_COMPUTE {
			payloadMap[assetKindAsString] = assetAttributes
		} else {
			destSection := convutils.MapMe(payloadMap[workloadStorageField])
			destSection[assetKindAsString] = []interface{}{assetAttributes}
		}

		//
		//   rearrange whatever is needed in the payload to align it with the new API (for example runAsUid
		//   will move under "security"
		//
		RearrangePayload(payloadMap)
		canonicalPayload, aliasErr := gpucompat.NormalizeLegacyGPUFields(payloadMap)
		if aliasErr != nil {
			storeAssetCompliance(ctx, assetInfo, []api.ComplianceInfoReason{{Details: aliasErr.Error()}})
			return
		}

		payloadMap = canonicalPayload

		fieldsWithCantEditRule := make([]string, 0)
		complianceResults := policy_compliance.ComplianceErrors{}

		defaults := assetInfo.defaults
		rules := assetInfo.rules

		defaultFromOrder := assetInfo.defaultFromOrder.Worker

		if assetInfo.replicaType != "" {
			sectionOfReplica := sectionPerReplicaType[workloadsApi.ReplicaType(assetInfo.replicaType)]
			if sectionOfReplica != "" {
				defaults = convutils.MapMe(defaults[sectionOfReplica])
				rules = convutils.MapMe(rules[sectionOfReplica])

				if workloadsApi.ReplicaType(assetInfo.replicaType) == workloadsApi.REPLICATYPE_MASTER {
					defaultFromOrder = assetInfo.defaultFromOrder.Master
				}
			}
		}

		policy_compliance.ValidateCompliance("", rules, defaults, payloadMap,
			*defaultFromOrder, &fieldsWithCantEditRule, nil, &complianceResults)

		for _, complianceResult := range complianceResults {
			belongsToAsset, shortName := fieldBelongsToTheAsset(complianceResult.PolicyField, assetKind)
			if !belongsToAsset {
				continue
			}

			// skip also mandatory fields that are overridable, because there value can change when creating the workload
			if strings.Index(complianceResult.Details, valueForMandatoryError) > 0 {
				if overridableFields.Has(shortName) {
					continue
				}
			}

			err := api.ComplianceInfoReason{
				Field:   convutils.Ref(complianceResult.PolicyField),
				Details: complianceResult.Details,
			}

			if complianceResult.Rule != nil {
				err.Rule = convutils.Ref(api.PolicyRuleEnum(*complianceResult.Rule))
			}

			complianceErrors = append(complianceErrors, err)
		}
	}

	storeAssetCompliance(ctx, assetInfo, complianceErrors)
}

func storeAssetCompliance(ctx context.Context, assetInfo assetChanInfo, complianceErrors []api.ComplianceInfoReason) {
	assetId := assetInfo.patchingPtrs.Meta.Id
	isCompliant := len(complianceErrors) == 0

	complianceInfo := api.ComplianceInfo{
		Compliance: isCompliant,
		Reason:     complianceErrors,
		Imposed:    assetInfo.isImposed && isCompliant, /* we cannot impose an asset that does not comply to the policy */
	}
	*assetInfo.patchingPtrs.ComplianceInfo = &complianceInfo

	//
	// store the validation result in the DB for next time. no need to store the compliance result of workspace/training
	// which are only relevant for specific run params and not to real asset
	//
	if string(assetInfo.patchingPtrs.Meta.Kind) != string(auth.WorkloadTypeToAssetKind[assetInfo.workloadType]) {
		dbErr := dao.UpsertPolicyCompliance(ctx, assetId, assetInfo.policyInfo, complianceInfo)
		if dbErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to store compliance result of %s %s in the db: %v",
				assetInfo.patchingPtrs.Meta.Kind, assetInfo.patchingPtrs.Meta.Name, dbErr)
		}
	}
}

// get the ids of instances that we suspect to be imposed assets. we assume them as assets in case the name
// is a uuid. it is ok if this is not really an imposedAsset cause the final validation is done against the
// ids in the list, so if we have an id which is not on the list no harm done.
// note that we could NOT use the imposedAssets field of the policy, because it only contains the imposed assets
// of the current scope, while effective policy can have instances that were originated from imposed assets
// of parent scope (the policy have them as instances but not as imposed assets)
func getImposedAssetIds(effectivePolicy *policy_compliance.EffectivePolicyOfProject) datastructs.Set[string] {
	result := datastructs.NewSet[string]()

	storage := convutils.MapMe(effectivePolicy.DefaultsV2[workloadStorageField])

	for _, storageDetails := range storage {
		storageDetailsMap := convutils.MapMe(storageDetails)

		instances := storageDetailsMap[itemized.PolicyInstances]
		if instances == nil {
			continue
		}

		for _, instance := range instances.([]interface{}) {
			instanceAsMap := convutils.MapMe(instance)
			name := instanceAsMap[StorageInstanceName]

			if name != nil && len(name.(string)) == uuidLength {
				if query_types.AssetIdsRegexp.MatchString(name.(string)) {
					result.Add(name.(string))
				}
			}
		}
	}

	return result
}

func fieldBelongsToTheAsset(fieldName string, assetKind api.AssetKind) (bool, string) {
	if fieldName == "" {
		return false, ""
	}

	assetKindAsString := string(assetKind)

	// compute is 1:1 b/w the asset and the workload section
	if assetKind == api.ASSETKIND_COMPUTE {
		if strings.HasPrefix(fieldName, assetKindAsString+".") {
			return true, fieldName[len(assetKindAsString)+1:]
		}
		return false, fieldName
	}

	// datasource assets go to the storage section
	if assetKind != api.ASSETKIND_ENVIRONMENT {
		assetRelatedSection := composer_v2.MapAssetKindToWorkloadSpec[string(assetKind)]
		expectedPrefix := convutils.DotUs(storageSection, assetRelatedSection) + "."

		if strings.HasPrefix(fieldName, expectedPrefix) {
			return strings.HasPrefix(fieldName, expectedPrefix), fieldName[len(expectedPrefix):]
		}
		return false, fieldName
	}

	// reaching here, we are handling environment asset. for environment, we need to see if the field
	// belongs to the spec of the environment asset. we have to disregard security.xxx prefix
	if strings.HasPrefix(fieldName, securitySection+".") {
		fieldName = fieldName[len(securitySection)+1:]
	}

	// remove instances suffix from the field, we do not have instances in the
	// workload itself, only in the policy representation of the field
	if strings.HasSuffix(fieldName, itemized.InstancesSuffix) {
		fieldName = strings.ReplaceAll(fieldName, itemized.InstancesSuffix, "")
	}

	// field should be reported if it belongs to asset
	return environmentSpecFieldNames.Has(fieldName), fieldName
}
