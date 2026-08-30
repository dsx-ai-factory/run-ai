package flat_workloads

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/jsonutils"
	policyconstants "github.com/run-ai/runai/runai-clients/policy-service-client/constants"
)

var (
	assetsWorkloadCracker    = jsonutils.NewJsonCracker(api.TrainingV2{})
	fldNamePolicyInfo        = assetsWorkloadCracker.NameOf(&assetsWorkloadCracker.Struct().PolicyInfo)        // policyInfo
	fldNameSpecificRunParams = assetsWorkloadCracker.NameOf(&assetsWorkloadCracker.Struct().SpecificRunParams) // specificRunParams
)

var (
	specificRunJsonCracker = jsonutils.NewJsonCracker(api.TrainingSpecificRunParams{})
	fldNameAllowOverQuota  = specificRunJsonCracker.NameOf(&specificRunJsonCracker.Struct().AllowOverQuota) // specificRunParams.allowOverQuota
)

var (
	policyInfoCracker             = jsonutils.NewJsonCracker(api.PolicyInfo{})
	fldNamePolicyInfoRules        = policyInfoCracker.NameOf(&policyInfoCracker.Struct().Rules)             // policyInfo.rules
	fldNamePolicyInfoAttrDefaults = policyInfoCracker.NameOf(&policyInfoCracker.Struct().AttributeDefaults) // policyInfo.attributeDefaults
)

var (
	policyDefaultsCracker      = jsonutils.NewJsonCracker(api.AttributeDefaults{})
	fldNameAssetsDefaults      = policyDefaultsCracker.NameOf(&policyDefaultsCracker.Struct().Assets)            // policyInfo.attributeDefaults.assets
	fldNameSpecificRunDefaults = policyDefaultsCracker.NameOf(&policyDefaultsCracker.Struct().SpecificRunParams) // policyInfo.attributeDefaults.specificRunParams
)

var (
	assetDefaultsCracker       = jsonutils.NewJsonCracker(api.AssetsDefaults{})
	fldNameDatasourcesDefaults = assetDefaultsCracker.NameOf(&assetDefaultsCracker.Struct().Datasources) // policyInfo.rules.assets.datasources
)

var (
	rulesInfoCracker        = jsonutils.NewJsonCracker(api.Rules{})
	fldNameAssetsRules      = rulesInfoCracker.NameOf(&rulesInfoCracker.Struct().Assets)            // policyInfo.rules.assets
	fldNameSpecificRunRules = rulesInfoCracker.NameOf(&rulesInfoCracker.Struct().SpecificRunParams) // policyInfo.rules.specificRunParams
)

var (
	assetRulesCracker       = jsonutils.NewJsonCracker(api.AssetsRules{})
	fldNameDatasourcesRules = assetRulesCracker.NameOf(&assetRulesCracker.Struct().Datasources) // policyInfo.rules.assets.datasources
)

var fldNamePolicyPriorityClass = defaultsJsonCracker.NameOf(&defaultsJsonCracker.Struct().PriorityClass)

var (
	specificRunParamsJsonCracker = jsonutils.NewJsonCracker(api.SpecificRunParamsRules{})
	fldNameAutoscaling           = specificRunParamsJsonCracker.NameOf(&specificRunParamsJsonCracker.Struct().Autoscaling) // specificRunParams.autoscaling

	autoScalingFlatJsonCracker            = jsonutils.NewJsonCracker(api.AutoScalingRules{})
	fldNameFlatAutoscalingMetric          = autoScalingFlatJsonCracker.NameOf(&autoScalingFlatJsonCracker.Struct().Metric)
	fldNameFlatAutoscalingMetricThreshold = autoScalingFlatJsonCracker.NameOf(&autoScalingFlatJsonCracker.Struct().MetricThreshold)

	autoScalingAssetJsonCracker            = jsonutils.NewJsonCracker(api.SpecificRunAutoScalingAutoScaling{})
	fldNameAssetAutoscalingMetric          = autoScalingAssetJsonCracker.NameOf(&autoScalingAssetJsonCracker.Struct().ThresholdMetric)
	fldNameAssetAutoscalingMetricThreshold = autoScalingAssetJsonCracker.NameOf(&autoScalingAssetJsonCracker.Struct().ThresholdValue)

	inferenceParamsJsonCracker = jsonutils.NewJsonCracker(api.SupersetSpec{})
	fldNameServingPort         = inferenceParamsJsonCracker.NameOf(&inferenceParamsJsonCracker.Struct().ServingPort)

	specificRunParamsAssetJsonCracker = jsonutils.NewJsonCracker(api.SpecificRunServingPortAccess{})
	fldNameAssetServingPortAccess     = specificRunParamsAssetJsonCracker.NameOf(&specificRunParamsAssetJsonCracker.Struct().ServingPortAccess)

	servingPortAccessFlatJsonCracker    = jsonutils.NewJsonCracker(api.ServingPortAccess{})
	fldNameServingPortAuthorizationType = servingPortAccessFlatJsonCracker.NameOf(&servingPortAccessFlatJsonCracker.Struct().AuthorizationType)
	fldNameServingPortAuthorizedUsers   = servingPortAccessFlatJsonCracker.NameOf(&servingPortAccessFlatJsonCracker.Struct().AuthorizedUsers)
	fldNameServingPortAuthorizedGroups  = servingPortAccessFlatJsonCracker.NameOf(&servingPortAccessFlatJsonCracker.Struct().AuthorizedGroups)
	fldNameServingPortLocalAccess       = servingPortAccessFlatJsonCracker.NameOf(&servingPortAccessFlatJsonCracker.Struct().ClusterLocalAccessOnly)
)

// FlatToAssetsBased does a reverse translation from flat policy to assets-based workload struct.
func FlatToAssetsBased(workloadAsMap map[string]interface{}, defaults map[string]interface{}, rules map[string]interface{}) (map[string]interface{}, *api.Error) {
	specificRunParamsAsMap := convutils.MapMe(workloadAsMap[fldNameSpecificRunParams])

	//
	//   also make sure all itemized instances have keys. this is relevant
	//   only for fields from copy-and-edit workload, because in policy we make
	//   sure every itemized have nickname
	//
	policy_compliance.MakeSureAllItemizedHaveKey(specificRunParamsAsMap)

	//
	//  handle exotic fields
	//
	if defaults[fldNamePolicyPriorityClass] != nil {
		// allow over quota is based on priority class field
		priorityClass := fmt.Sprintf("%v", defaults[fldNamePolicyPriorityClass])
		specificRunParamsAsMap[fldNameAllowOverQuota] = priorityClass == policyconstants.PriorityClassInteractivePreemptible
	}

	flatToAssetsBasedInferenceFields(defaults, rules)

	//
	//   move all the security fields to the top level
	//
	moveSection(defaults, securitySection, defaults, true)
	moveSection(rules, securitySection, rules, true)

	//
	//   copy the rules to the PolicyInfo, as follows:
	//      everything under storage goes to assets.datasources
	//      all the rest goes to specificRunParams
	//
	attributeDefaults := make(map[policy_compliance.PerReplicaField]interface{})
	policyInfoSection := make(map[string]interface{})

	if len(defaults) != 0 || len(rules) != 0 {
		policyInfoSection = jsonutils.AddSection(workloadAsMap, fldNamePolicyInfo)
		policyInfoRulesSection := jsonutils.AddSection(policyInfoSection, fldNamePolicyInfoRules)
		assetsRulesSection := jsonutils.AddSection(policyInfoRulesSection, fldNameAssetsRules)

		assetsRulesSection[fldNameDatasourcesRules] = rules[storageSection]
		delete(rules, storageSection)

		policyInfoRulesSection[fldNameSpecificRunRules] = rules

		//
		//   prepare the default for merging with specific run params. this will
		//   split instances and attribute defaults. the instances will remain in
		//   the map as ordinary items array, the defaults are returned as a map
		//
		attributeDefaults = policy_compliance.PrepareDefaultsForMutation(defaults)

		//
		//   we do not need instances of storage section. instances we need only for
		//   the specific run params
		//
		delete(defaults, storageSection)

		policy_compliance.ApplyDefaults(specificRunParamsAsMap, defaults, attributeDefaults)

		//
		//   the merged set of ApplyDefaults is written over defaults, so use this
		//   as the complete set of defaults for the result
		//
		workloadAsMap[fldNameSpecificRunParams] = defaults
	}

	//
	//   finally we have to translate the attributeDefaults to the attribute defaults
	//   section of the policyInfo. the map contains keys like this:
	//      tolerations
	//      storage.S3
	//      storage.Nfs
	//   all the storage fields we have to push under datasources, all the rest goes to
	//   the top level of specificRunParams
	//
	if len(attributeDefaults) != 0 {
		attributeDefaultsSection := jsonutils.AddSection(policyInfoSection, fldNamePolicyInfoAttrDefaults)
		assetsDefaultsSection := jsonutils.AddSection(attributeDefaultsSection, fldNameAssetsDefaults)
		datasourceDefaultsSection := jsonutils.AddSection(assetsDefaultsSection, fldNameDatasourcesDefaults)
		specificRunDefaultsSection := jsonutils.AddSection(attributeDefaultsSection, fldNameSpecificRunDefaults)

		for fldInfo, defaults := range attributeDefaults {
			if strings.HasPrefix(fldInfo.FieldName, storageSection) {
				datasourceDefaultsSection[fldInfo.FieldName[len(storageSection)+1:]] = defaults
			} else {
				specificRunDefaultsSection[fldInfo.FieldName] = defaults
			}
		}
	}

	return workloadAsMap, nil
}

// flatToAssetsBasedInferenceFields translates the inference fields from flat to asset-based policy.
func flatToAssetsBasedInferenceFields(defaults, rules map[string]interface{}) {
	for _, m := range []map[string]interface{}{defaults, rules} {
		// autoscaling fields metric and metricThreshold have different names in flat and asset-based policy:
		// metric -> thresholdMetric
		// metricThreshold -> thresholdValue
		if autoscalingAny, ok := m[fldNameAutoscaling]; ok {
			if autoscalingMap, isMap := autoscalingAny.(map[string]interface{}); isMap {
				if metricValue, exists := autoscalingMap[fldNameFlatAutoscalingMetric]; exists {
					autoscalingMap[fldNameAssetAutoscalingMetric] = metricValue
					delete(autoscalingMap, fldNameFlatAutoscalingMetric)
				}
				if metricThresholdValue, exists := autoscalingMap[fldNameFlatAutoscalingMetricThreshold]; exists {
					autoscalingMap[fldNameAssetAutoscalingMetricThreshold] = metricThresholdValue
					delete(autoscalingMap, fldNameFlatAutoscalingMetricThreshold)
				}
			}
		}

		// servingPortAccess fields have different locations in flat and asset-based policy:
		// servingPort.someAuthField -> servingPortAccess.someAuthField
		if m[fldNameServingPort] != nil {
			servingPortFlat := m[fldNameServingPort].(map[string]interface{})
			// if one of the fields is set
			if servingPortFlat[fldNameServingPortAuthorizationType] != nil ||
				servingPortFlat[fldNameServingPortAuthorizedUsers] != nil ||
				servingPortFlat[fldNameServingPortAuthorizedGroups] != nil ||
				servingPortFlat[fldNameServingPortLocalAccess] != nil {

				servingPortAccessAsset := make(map[string]interface{})

				if servingPortFlat[fldNameServingPortAuthorizationType] != nil {
					servingPortAccessAsset[fldNameServingPortAuthorizationType] = servingPortFlat[fldNameServingPortAuthorizationType]
					delete(servingPortFlat, fldNameServingPortAuthorizationType)
				}

				if servingPortFlat[fldNameServingPortAuthorizedUsers] != nil {
					servingPortAccessAsset[fldNameServingPortAuthorizedUsers] = servingPortFlat[fldNameServingPortAuthorizedUsers]
					delete(servingPortFlat, fldNameServingPortAuthorizedUsers)
				}

				if servingPortFlat[fldNameServingPortAuthorizedGroups] != nil {
					servingPortAccessAsset[fldNameServingPortAuthorizedGroups] = servingPortFlat[fldNameServingPortAuthorizedGroups]
					delete(servingPortFlat, fldNameServingPortAuthorizedGroups)
				}

				if servingPortFlat[fldNameServingPortLocalAccess] != nil {
					servingPortAccessAsset[fldNameServingPortLocalAccess] = servingPortFlat[fldNameServingPortLocalAccess]
					delete(servingPortFlat, fldNameServingPortLocalAccess)
				}

				m[fldNameAssetServingPortAccess] = servingPortAccessAsset
			}
		}
	}
}
