/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tree

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"

	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/itemized"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/scopes"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/conjungo"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	assetsService "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

var (
	nonDistributedCracker = jsonutils.NewJsonCracker(api.SupersetDefaults{})
	storageSection        = nonDistributedCracker.NameOf(&nonDistributedCracker.Struct().Storage)
	securitySection       = nonDistributedCracker.NameOf(&nonDistributedCracker.Struct().Security)
	computeSection        = nonDistributedCracker.NameOf(&nonDistributedCracker.Struct().Compute)
)

var (
	distributedJsonCracker = jsonutils.NewJsonCracker(api.DistributedPolicyDefaultsV2{})
	masterSection          = distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Master)
	workerSection          = distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Worker)
)

var (
	distributedInferenceJsonCracker = jsonutils.NewJsonCracker(api.DistributedInferenceDefaultsV2{})
	leaderSection                   = distributedInferenceJsonCracker.NameOf(&distributedInferenceJsonCracker.Struct().Leader)
	diWorkerSection                 = distributedInferenceJsonCracker.NameOf(&distributedInferenceJsonCracker.Struct().Worker)
)

// we set the name field according to the asset id
const storageNameField = "name"

// validate correctness of imposed assets. returns nil or error, and in addition return the spec of those assets ready to add to
// the spec of the policy (defaults and rules)
func (ht *HierarchyTree) validateImposedAssets(ctx context.Context, policyScope scopes.OrgUnit, imposedAssetIds *[]string) (map[string]interface{}, map[string]interface{}, *api.Error) {
	tenantId := auth.TenantIdOf(ctx)

	//
	//    prepare the result arrays - one for defaults, one for rules.
	//
	resultDefaults := make(map[string]interface{})
	resultRules := make(map[string]interface{})

	if imposedAssetIds == nil || len(*imposedAssetIds) == 0 {
		// Empty maps would introduce replica sections into non-distributed policies.
		return nil, nil, nil
	}

	//
	//  make a list of the assets. validate the ids are ok and no duplicates
	//
	allAssetIds := datastructs.NewSet[string]()

	for _, assetId := range *imposedAssetIds {
		if assetId == "" {
			return nil, nil, api.NewValidationError("imposed asset id should not be empty")
		}
		_, err := uuid.Parse(assetId)
		if err != nil {
			return nil, nil, api.NewValidationError(fmt.Sprintf("malformed asset id '%s': %v", assetId, err))
		}
		if allAssetIds.Has(assetId) {
			return nil, nil, api.NewValidationError(fmt.Sprintf("asset with id '%s' appears more than once", assetId))
		}
		allAssetIds.Add(assetId)
	}

	//
	//   get the details of the assets
	//
	assetsDetails, listError := assets.Get().Client().ListDatasourcesByIds(ctx, *imposedAssetIds)
	if listError != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] list datasources failed: %s: %v", ht.treeId.ToString(), listError.Message, convutils.Deref(listError.Details, ""))
		return nil, nil, api.NewError(int(listError.Code), listError.Message, nil)
	}

	//
	//   check the scope
	//
	for _, asset := range assetsDetails.Entries {
		delete(allAssetIds, asset.Meta.Id) // remove it from the set

		assetScope, scopeErr := scopes.NewOrgUnitFromAssetMeta(&asset.Meta, tenantId)
		if scopeErr != nil {
			return nil, nil, scopeErr
		}
		if asset.Meta.Scope == assetsService.SCOPE_TENANT {
			assetScope.Id = fmt.Sprintf("%d", auth.TenantIdOf(ctx))
		}

		if !assetScope.IsSameScopeAs(policyScope) {
			return nil, nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("scope of asset %s (%s) does not match to to the scope of the policy (%s)",
				asset.Meta.Id, assetScope.String(), policyScope.String()), nil)
		}

		assetKind := string(asset.Meta.Kind)

		datasourceField, datasourceOk := composer_v2.MapAssetKindToDatasourcesList[assetKind]
		if !datasourceOk {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] no mapping b/w asset kind %s to datasource list spec", ht.treeId.ToString(), assetKind)
			return nil, nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("cannot use asset of kind %s as imposedAsset", assetKind), nil)
		}

		//
		//   the spec will hold one item, with the key = the data source field, and the value is the spec
		//
		specAsJson, jsonErr := gojsonq.LoadJson(asset.Spec)
		if jsonErr != nil {
			return nil, nil, api.NewError(http.StatusInternalServerError, "failed to parse the details of the imposedAssets: %v", jsonErr)
		}

		//
		//    we want to push it to the result according to the field names of the workload payload
		//    (typically all three are the same: assetKind = nfs, datasource list response has nfs field
		//    and workload has nfs storage). but for configMapVolume the asset kind is config-map, the
		//    field in the datasource list will appear as "config_map", and in the workload payload it
		//    is expected as "configMapVolume". guys, it is merged already, so we cannot change it...
		//    but please... next time, let's keep it consistent, ok?
		//
		workloadFieldName, workloadFieldOk := composer_v2.MapAssetKindToWorkloadSpec[assetKind]
		if !workloadFieldOk {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("[%s] no mapping b/w asset kind %s to workload spec", ht.treeId.ToString(), assetKind)
			return nil, nil, api.NewError(http.StatusInternalServerError, "failed to process the list imposed assets", nil)
		}

		//   the name of the datadource is set to be the id of the asset
		storageAsMap := convutils.MapMe(specAsJson[datasourceField])
		storageAsMap[storageNameField] = asset.Meta.Id

		//
		//   Add the instance to defaults
		//
		if resultDefaults[workloadFieldName] == nil {
			resultDefaults[workloadFieldName] = map[string]interface{}{
				itemized.PolicyInstances: []interface{}{specAsJson[datasourceField]},
			}
		} else {
			workloadField := convutils.MapMe(resultDefaults[workloadFieldName])

			if workloadField[itemized.PolicyInstances] == nil {
				workloadField[itemized.PolicyInstances] = []interface{}{specAsJson[datasourceField]}
			} else {
				instancesArr := workloadField[itemized.PolicyInstances].([]interface{})
				instancesArr = append(instancesArr, specAsJson[datasourceField])
				workloadField[itemized.PolicyInstances] = instancesArr
			}
		}

		//
		//   Add the instance name to locked rules
		//
		if resultRules[workloadFieldName] == nil {
			resultRules[workloadFieldName] = map[string]interface{}{
				itemized.PolicyInstances: map[string]interface{}{
					"locked": []interface{}{asset.Meta.Id},
				},
			}
		} else {
			workloadFieldRules := convutils.MapMe(resultRules[workloadFieldName])

			if workloadFieldRules[itemized.PolicyInstances] == nil {
				workloadFieldRules[itemized.PolicyInstances] = map[string]interface{}{
					"locked": []interface{}{asset.Meta.Id},
				}
			} else {
				instancesRules := convutils.MapMe(workloadFieldRules[itemized.PolicyInstances])
				if instancesRules["locked"] == nil {
					instancesRules["locked"] = []interface{}{asset.Meta.Id}
				} else {
					lockedArr := instancesRules["locked"].([]interface{})
					lockedArr = append(lockedArr, asset.Meta.Id)
					instancesRules["locked"] = lockedArr
				}
			}
		}
	}

	//
	//   if there are assets that were not removed from the list, it means they do not exist
	//
	if len(allAssetIds) != 0 {
		var missingIds string
		var missingAssets int

		for missingId := range allAssetIds {
			missingIds += errorDelimiter + missingId
			missingAssets += 1

			if missingAssets > maxMissingIdsPerErrorMessage {
				missingIds += fmt.Sprintf(", and additional %d assets", len(allAssetIds)-maxMissingIdsPerErrorMessage)
				break
			}
		}

		return nil, nil, api.NewError(http.StatusNotFound, fmt.Sprintf("asset(s) %s do not exist", missingIds[len(errorDelimiter):]), nil)
	}

	return resultDefaults, resultRules, nil
}

func (ht *HierarchyTree) enrichDefaultsWithImposed(defaults dbmodels.JsonBMap,
	imposedWorkerAssets map[string]interface{}, imposedMasterAssets map[string]interface{},
) (dbmodels.JsonBMap, *api.Error) {
	if imposedWorkerAssets == nil && imposedMasterAssets == nil {
		return defaults, nil
	}

	result := defaults
	if result == nil {
		result = dbmodels.JsonBMap{}
	}

	if imposedWorkerAssets != nil {
		var workerTarget map[string]interface{}

		switch ht.treeId.WorkloadType {
		case WorkloadTypeDistributed:
			workerTarget = findOrCreateSection(result, workerSection, storageSection)
		case WorkloadTypeDistributedInference:
			workerTarget = findOrCreateSection(result, diWorkerSection, storageSection)
		default:
			workerTarget = findOrCreateSection(result, storageSection)
		}

		err := conjungo.Merge(&workerTarget, imposedWorkerAssets, conjungo.NewOptions())
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, "failed to import the imposedAssets to the policy", err)
		}
	}

	if imposedMasterAssets != nil {
		var masterTarget map[string]interface{}

		if ht.treeId.WorkloadType == WorkloadTypeDistributedInference {
			// For distributed inference, master assets go to leader.storage
			masterTarget = findOrCreateSection(result, leaderSection, storageSection)
		} else {
			masterTarget = findOrCreateSection(result, masterSection, storageSection)
		}

		err := conjungo.Merge(&masterTarget, imposedMasterAssets, conjungo.NewOptions())
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, "failed to import the imposedAssets to the master policy", err)
		}
	}

	return result, nil
}

func (ht *HierarchyTree) enrichRulesWithImposed(rules dbmodels.JsonBMap,
	imposedWorkerRules map[string]interface{}, imposedMasterRules map[string]interface{},
) (dbmodels.JsonBMap, *api.Error) {
	if imposedWorkerRules == nil && imposedMasterRules == nil {
		return rules, nil
	}

	result := rules
	if result == nil {
		result = dbmodels.JsonBMap{}
	}

	if imposedWorkerRules != nil {
		var workerTarget map[string]interface{}

		switch ht.treeId.WorkloadType {
		case WorkloadTypeDistributed:
			workerTarget = findOrCreateSection(result, workerSection, storageSection)
		case WorkloadTypeDistributedInference:
			workerTarget = findOrCreateSection(result, diWorkerSection, storageSection)
		default:
			workerTarget = findOrCreateSection(result, storageSection)
		}

		err := conjungo.Merge(&workerTarget, imposedWorkerRules, conjungo.NewOptions())
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, "failed to import the imposedAssets rules to the policy", err)
		}
	}

	if imposedMasterRules != nil {
		var masterTarget map[string]interface{}

		if ht.treeId.WorkloadType == WorkloadTypeDistributedInference {
			// For distributed inference, master rules go to leader.storage
			masterTarget = findOrCreateSection(result, leaderSection, storageSection)
		} else {
			masterTarget = findOrCreateSection(result, masterSection, storageSection)
		}

		err := conjungo.Merge(&masterTarget, imposedMasterRules, conjungo.NewOptions())
		if err != nil {
			return nil, api.NewError(http.StatusServiceUnavailable, "failed to import the imposedAssets rules to the master policy", err)
		}
	}

	return result, nil
}

func findOrCreateSection(root map[string]interface{}, sections ...string) map[string]interface{} {
	curElement := root

	for _, section := range sections {
		if curElement[section] == nil {
			curElement[section] = make(map[string]interface{})
		}
		curElement = convutils.MapMe(curElement[section])
	}

	return curElement
}
