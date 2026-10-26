/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/jinzhu/copier"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
)

var (
	unifiedJsonCracker = jsonutils.NewJsonCracker(api.UnifiedPolicyInfoPerReplicaType{})
	defaultsSection    = unifiedJsonCracker.NameOf(&unifiedJsonCracker.Struct().Defaults)
	rulesSection       = unifiedJsonCracker.NameOf(&unifiedJsonCracker.Struct().Rules)

	rulesJsonCracker   = jsonutils.NewJsonCracker(api.SupersetRules{})
	nodeTypeFieldName  = rulesJsonCracker.NameOf(&rulesJsonCracker.Struct().NodeType)
	nodePoolsFieldName = rulesJsonCracker.NameOf(&rulesJsonCracker.Struct().NodePools)

	stringOptionsCracker = jsonutils.NewJsonCracker(api.StringRules{})
	optionsRules         = stringOptionsCracker.NameOf(&stringOptionsCracker.Struct().Options)

	optionsCracker = jsonutils.NewJsonCracker(api.StringOption{})
	optionValue    = optionsCracker.NameOf(&optionsCracker.Struct().Value)
)

type UnifiedAPIService struct{}

func NewUnifiedAPIService() *UnifiedAPIService {
	return &UnifiedAPIService{}
}

type UnifiedWorkloadRequest api.UnifiedWorkloadRequest

func (s *UnifiedAPIService) GetUnifiedWorkload(ctx context.Context, req api.UnifiedWorkloadRequest) (api.ImplResponse, error) {
	result, err := (UnifiedWorkloadRequest)(req).Exec(ctx)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, result), nil
}

func (req UnifiedWorkloadRequest) Exec(ctx context.Context) (*workloadsApi.UnifiedWorkload, *workloadsApi.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, workloadsApi.NewError(apiErr.Code, apiErr.Message)
	}

	//
	//    asset and template mode require clusterId
	//
	var jsonErr error

	flatSpec := make(map[string]interface{})
	templateSpec := make(map[string]interface{})

	assetIds := make([]string, 0)

	if req.AssetId != nil {
		assetIds = append(assetIds, *req.AssetId)
	}
	if req.AssetIds != nil {
		assetIds = append(assetIds, req.AssetIds...)
	}

	dataVolumeIds := make([]string, 0)
	if len(assetIds) != 0 && req.WorkloadId != nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, "assetId and workloadId are mutual exclusive")
	}

	if len(req.DataVolumeIds) != 0 && req.WorkloadId != nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, "dataVolumeIds and workloadId are mutual exclusive")
	}

	if req.TemplateId != nil && req.WorkloadId != nil {
		return nil, workloadsApi.NewError(http.StatusBadRequest, "workloadId and templateId are mutual exclusive")
	}

	if req.DataVolumeIds != nil {
		dataVolumeIds = req.DataVolumeIds
	}
	sameAsWorker := false

	if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED && len(assetIds) == 0 && req.WorkloadId == nil {
		flatSpec[masterSection] = make(map[string]interface{}) // we need the section to exist, even if empty
	}

	var assetFlat *assetsApi.AssetFlatResponse

	if len(assetIds) != 0 {
		if req.ClusterId == nil {
			return nil, workloadsApi.NewError(http.StatusBadRequest, "clusterId is mandatory when specifying assetId/assetIds")
		}

		assetFlatReq := assetsApi.AssetFlatRequest{
			AssetIds:           assetIds,
			ClusterId:          *req.ClusterId,
			ProjectId:          req.ProjectId,
			WorkloadType:       assetsApi.WorkloadType(req.WorkloadType),
			MasterSameAsWorker: convutils.Ref(false),
		}

		var assetErr *assetsApi.Error

		assetFlat, assetErr = assets.Get().RequestAssetFlatFormat(ctx, assetFlatReq)
		if assetErr != nil {
			return nil, workloadsApi.NewError(assetErr.Code, assetErr.Message)
		}

		if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED_INFERENCE && assetFlat != nil && assetFlat.Worker != nil {
			if err := populateDistributedInferenceReplicas(assetFlat.Worker); err != nil {
				return nil, err
			}
		}
	}

	if len(dataVolumeIds) != 0 {
		if assetFlat == nil {
			assetFlat = &assetsApi.AssetFlatResponse{}
		}

		if assetFlat.Worker == nil {
			assetFlat.Worker = &assetsApi.SupersetSpec{}
			if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
				assetFlat.Master = *assetsApi.NewNullableMasterSpec(&assetsApi.MasterSpec{})
			}
		}

		if assetFlat.Worker.Storage.Get() == nil {
			assetFlat.Worker.Storage = *assetsApi.NewNullableSupersetSpecAllOfStorage(&assetsApi.SupersetSpecAllOfStorage{})
			if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
				assetFlat.Master.Get().Storage = *assetsApi.NewNullableSupersetSpecAllOfStorage(&assetsApi.SupersetSpecAllOfStorage{})
			}
		}

		dataVolumeInstances := make([]assetsApi.DataVolumeInstance, 0)
		for _, id := range dataVolumeIds {
			dataVolumeInstance := assetsApi.DataVolumeInstance{
				Id: *assetsApi.NewNullableString(&id),
			}
			dataVolumeInstances = append(dataVolumeInstances, dataVolumeInstance)
		}

		assetFlat.Worker.Storage.Get().DataVolume = dataVolumeInstances
		if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
			assetFlat.Master.Get().Storage.Get().DataVolume = dataVolumeInstances
		}
	}

	if assetFlat != nil {
		if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
			flatSpec, jsonErr = gojsonq.LoadJson(assetFlat) // we need both master and worker sections
		} else {
			flatSpec, jsonErr = gojsonq.LoadJson(assetFlat.Worker) // non-distributed we need only the worker sections
		}
		if jsonErr != nil {
			return nil, workloadsApi.NewError(http.StatusInternalServerError, jsonErr.Error())
		}
	}

	var workloadMeta *workloadsApi.WorkloadMeta

	if req.WorkloadId != nil {
		permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
		if scopesErr != nil {
			permittedScopes = nil
		}

		workload, dbErr := dao_crud.DbGetWorkload(ctx, tenantId, string(req.WorkloadType), *req.WorkloadId, permittedScopes)
		if dbErr != nil {
			if dbErr.Code != http.StatusNotFound {
				return nil, workloadsApi.NewError(dbErr.Code, dbErr.Message)
			}

			snapshot, snapshotErr := dao_crud.DbGetWorkloadV2Snapshot(ctx, tenantId, string(req.WorkloadType), *req.WorkloadId, permittedScopes)
			if snapshotErr != nil {
				return nil, workloadsApi.NewError(snapshotErr.Code, snapshotErr.Message)
			}
			workload = snapshot.ToWorkload()
		}

		// Always redact literal access tokens; caller-dependent credential sanitization is applied separately.
		sanitizeWorkload(ctx, workload)

		workloadMeta = &workloadsApi.WorkloadMeta{}
		copier.Copy(workloadMeta, &workload)

		// in case of worker, we want to mimic the sameAsWorker from the workload itself
		sameAsWorker = convutils.BoolIsTrue(workload.MasterSameAsWorker)

		if req.ClusterId != nil && *req.ClusterId != workload.ClusterId {
			return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("workload %s is associated with cluster %s, not %s", workload.ID, workload.ClusterId, *req.ClusterId))
		}
		if req.ProjectId != "" && req.ProjectId != workload.ProjectId {
			return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("workload %s is associated with project %s, not %s", workload.ID, workload.ProjectId, req.ProjectId))
		}

		if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
			flatSpec[workerSection], jsonErr = gojsonq.LoadJson(workload.Spec)
			if jsonErr == nil {
				if convutils.BoolIsTrue(workload.MasterSameAsWorker) {
					flatSpec[masterSection], jsonErr = gojsonq.LoadJson(workload.Spec)
				} else if workload.MasterSpec != nil {
					flatSpec[masterSection], jsonErr = gojsonq.LoadJson(workload.MasterSpec)
				}
			}
		} else {
			flatSpec, jsonErr = gojsonq.LoadJson(workload.Spec)
		}

		if jsonErr != nil {
			return nil, workloadsApi.NewError(http.StatusInternalServerError, jsonErr.Error())
		}
	}

	if req.TemplateId != nil {
		// the permittedScoped we already have are permitted scope for creating a template.
		// they are totally different from the permitted scopes for reading a template, so
		// we have to load the read permissions specifically
		permittedScopesForRead, rbacErr := auth.LoadAllowedOrgUnits(ctx)
		if rbacErr != nil {
			return nil, workloadsApi.NewError(rbacErr.Code, rbacErr.Message)
		}

		template, dbErr := dao_crud.GetTemplateById(ctx, tenantId, *req.TemplateId, permittedScopesForRead.Raw())
		if dbErr != nil {
			return nil, workloadsApi.NewError(dbErr.Code, dbErr.Message)
		}

		if string(req.WorkloadType) != template.WorkloadType {
			return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("template %s is of type %s, not %s", template.ID, template.WorkloadType, req.WorkloadType))
		}

		//
		//   doing a basic validation that template scope match request scope. this is not a 100% validation
		//   to do that need to use get ancestors and validate VS the ancestors, but since the unified is used
		//   by the frontend which does the filtering of the template, this partial validation is sufficient.
		//
		if req.ClusterId != nil && template.ClusterId != nil && *req.ClusterId != *template.ClusterId {
			return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("template %s is associated with cluster %s, not %s", template.ID, *template.ClusterId, *req.ClusterId))
		}
		if req.ProjectId != "" && template.ProjectId != nil && req.ProjectId != *template.ProjectId {
			return nil, workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("workload %s is associated with project %s, not %s", template.ID, *template.ProjectId, req.ProjectId))
		}

		if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
			templateSpec[workerSection], jsonErr = gojsonq.LoadJson(template.CombinedSpec)
			if jsonErr == nil {
				if convutils.BoolIsTrue(template.MasterSameAsWorker) {
					templateSpec[masterSection], jsonErr = gojsonq.LoadJson(template.CombinedSpec)
				} else if template.CombinedMasterSpec != nil {
					templateSpec[masterSection], jsonErr = gojsonq.LoadJson(template.CombinedMasterSpec)
				}
			}
		} else {
			templateSpec, jsonErr = gojsonq.LoadJson(template.CombinedSpec)
		}

		if jsonErr != nil {
			return nil, workloadsApi.NewError(http.StatusInternalServerError, jsonErr.Error())
		}
	}

	//
	//   get the policy of the project. the unified API incorporate policy and workload/asset field
	//
	var effectivePolicy *policyCompliance.EffectivePolicyOfProject
	if req.ProjectId != "" {
		var policyErr *policyApi.Error
		effectivePolicy, policyErr = policyCompliance.GetEffectivePolicyOfProject(ctx, policies.Get().Client(), req.ProjectId, workloadsApi.WorkloadType(req.WorkloadType))
		if policyErr != nil {
			return nil, workloadsApi.NewError(policyErr.Code, policyErr.Message)
		}
	} else {
		effectivePolicy = policyCompliance.NewEffectivePolicyOfProject()
	}

	result := workloadsApi.UnifiedWorkload{
		Meta: workloadMeta,
	}

	//
	//  enrich the policy with node types and node pools
	//
	if req.ProjectId != "" {
		projectInfo, enrichError := req.EnrichPolicyFromProject(ctx, effectivePolicy.DefaultsV2, effectivePolicy.RulesV2)
		if enrichError != nil {
			return nil, enrichError
		}

		result.ProjectInfo = *workloadsApi.NewNullableUnifiedProjectInfo(&workloadsApi.UnifiedProjectInfo{
			Name: convutils.Ref(projectInfo.Name),
		})
	}

	policyInfo, policyErr := req.composePolicyInfo(effectivePolicy.DefaultsV2, effectivePolicy.RulesV2)
	if policyErr != nil {
		return nil, policyErr
	}

	if policyInfo != nil {
		result.PolicyInfo = *workloadsApi.NewNullableUnifiedPolicyInfo(policyInfo)
	}

	mutatedSpec, mutateErr := MutateWorkload(flatSpec, effectivePolicy.DefaultsV2, sameAsWorker, templateSpec, nil)
	if mutateErr != nil {
		return nil, workloadsApi.NewError(mutateErr.Code, mutateErr.Message)
	}

	responseSpec, responseMasterSpec, buildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](string(req.WorkloadType), mutatedSpec)
	if buildErr != nil {
		return nil, workloadsApi.NewError(buildErr.Code, buildErr.Message)
	}

	if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED_INFERENCE && responseSpec != nil {
		clearContainerFieldsFromParent(responseSpec)
	}

	nonEmptySpec := false
	spec := workloadsApi.UnifiedSpec{}

	if responseSpec != nil {
		spec.Worker = *workloadsApi.NewNullableSupersetSpec(gpucompat.LegacyCompatibleWorkloadSpec(responseSpec))
		nonEmptySpec = true
	}
	if responseMasterSpec != nil {
		spec.Master = *workloadsApi.NewNullableMasterSpec(gpucompat.LegacyCompatibleMasterSpec(responseMasterSpec))
		nonEmptySpec = true
	}
	if sameAsWorker {
		spec.MasterSpecSameAsWorker = *workloadsApi.NewNullableBool(&sameAsWorker)
	}

	if nonEmptySpec {
		result.Spec = *workloadsApi.NewNullableUnifiedSpec(&spec)
	}

	redactUnifiedWorkloadPlaintextTokenValues(&result)

	return &result, nil
}

func (req UnifiedWorkloadRequest) composePolicyInfo(defaults map[string]interface{}, rules map[string]interface{}) (*workloadsApi.UnifiedPolicyInfo, *workloadsApi.Error) {
	defaults = gpucompat.LegacyCompatibleGPUFields(defaults)
	rules = gpucompat.LegacyCompatibleGPUPolicy(rules)
	workerSectionAsMap := make(map[string]interface{})
	masterSectionAsMap := make(map[string]interface{})

	switch req.WorkloadType {
	case Distributed:
		req.setDistributedPolicySections(workerSectionAsMap, masterSectionAsMap, defaults, rules)
	case DistributedInference:
		req.setDistributedInferencePolicySections(workerSectionAsMap, defaults, rules)
	default:
		req.setNonDistributedPolicySections(workerSectionAsMap, defaults, rules)
	}

	asMap := make(map[string]interface{})
	if len(workerSectionAsMap) != 0 {
		asMap[workerSection] = workerSectionAsMap
	}
	if len(masterSectionAsMap) != 0 {
		asMap[masterSection] = masterSectionAsMap
	}

	if len(asMap) == 0 {
		return nil, nil
	}

	result := workloadsApi.UnifiedPolicyInfo{}
	err := json.Unmarshal(gojsonq.ToBytes(asMap), &result)
	if err != nil {
		return nil, workloadsApi.NewError(int32(http.StatusInternalServerError), fmt.Sprintf("failed to parse policy: %v", err))
	}

	return &result, nil
}

func (req UnifiedWorkloadRequest) setDistributedPolicySections(workerSectionAsMap map[string]interface{}, masterSectionAsMap map[string]interface{}, defaults map[string]interface{}, rules map[string]interface{}) {
	if len(defaults) != 0 {
		workerSectionAsMap[defaultsSection] = defaults[workerSection]
		masterSectionAsMap[defaultsSection] = defaults[masterSection]
	}
	if len(rules) != 0 {
		workerSectionAsMap[rulesSection] = rules[workerSection]
		masterSectionAsMap[rulesSection] = rules[masterSection]
	}
}

func (req UnifiedWorkloadRequest) setDistributedInferencePolicySections(workerSectionAsMap map[string]interface{}, defaults map[string]interface{}, rules map[string]interface{}) {
	if len(defaults) > 0 {
		req.convertDefaultsForDistributedInference(workerSectionAsMap, defaults)
	}
	if len(rules) > 0 {
		req.convertRulesForDistributedInference(workerSectionAsMap, rules)
	}
}

func (req UnifiedWorkloadRequest) convertDefaultsForDistributedInference(workerSectionAsMap map[string]interface{}, defaults map[string]interface{}) {
	if len(defaults) > 0 {
		workerSectionAsMap[defaultsSection] = defaults
	}
}

func (req UnifiedWorkloadRequest) convertRulesForDistributedInference(workerSectionAsMap map[string]interface{}, rules map[string]interface{}) {
	if len(rules) > 0 {
		workerSectionAsMap[rulesSection] = rules
	}
}

func (req UnifiedWorkloadRequest) setNonDistributedPolicySections(workerSectionAsMap map[string]interface{}, defaults map[string]interface{}, rules map[string]interface{}) {
	if len(defaults) != 0 {
		workerSectionAsMap[defaultsSection] = defaults
	}
	if len(rules) != 0 {
		workerSectionAsMap[rulesSection] = rules
	}
}

func (req UnifiedWorkloadRequest) EnrichPolicyFromProject(ctx context.Context,
	defaults map[string]interface{},
	rules map[string]interface{},
) (*orgUnitApi.Project, *workloadsApi.Error) {

	projectInfo, err := orgUnitService.Get().GetProjectById(ctx, req.ProjectId)
	if err != nil {
		return nil, workloadsApi.NewError(err.Code, err.Message)
	}

	enrichErr := req.enrichNodeTypesFromProject(rules, projectInfo.NodeTypes)
	if enrichErr != nil {
		return nil, enrichErr
	}

	enrichErr = req.enrichNodePoolsFromProject(defaults, projectInfo.Effective.DefaultNodePools, rules, projectInfo.Resources)
	if enrichErr != nil {
		return nil, enrichErr
	}

	return projectInfo, nil
}

func (req UnifiedWorkloadRequest) enrichNodeTypesFromProject(
	rules map[string]interface{}, nodeTypesPerWorkload *orgUnitApi.NodeTypesPerWorkload,
) *workloadsApi.Error {
	if nodeTypesPerWorkload == nil {
		return nil
	}

	var nodeTypes []string
	switch req.WorkloadType {
	case api.WORKLOADTYPE_WORKSPACE:
		nodeTypes = nodeTypesPerWorkload.Workspace
	case api.WORKLOADTYPE_TRAINING, api.WORKLOADTYPE_DISTRIBUTED:
		nodeTypes = nodeTypesPerWorkload.Training
	}

	if len(nodeTypes) == 0 {
		return nil
	}

	options := make([]policyApi.StringOption, 0, len(nodeTypes))

	for _, nodeType := range nodeTypes {
		options = append(options, policyApi.StringOption{
			Value: nodeTypesPerWorkload.Names[nodeType],
		})
	}

	optionsAsBytes, _ := json.Marshal(options)

	var optionsAsMap []interface{}
	err := json.Unmarshal(optionsAsBytes, &optionsAsMap)
	if err != nil {
		return workloadsApi.NewError(http.StatusInternalServerError, err.Error())
	}

	if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
		req.enrichNodeTypesOfReplica(workerSection, rules, optionsAsMap)
		req.enrichNodeTypesOfReplica(masterSection, rules, optionsAsMap)
	} else {
		req.enrichNodeTypesOfReplica("", rules, optionsAsMap)
	}

	return nil
}

func (req UnifiedWorkloadRequest) enrichNodeTypesOfReplica(replicaType string,
	rules map[string]interface{}, optionsAsMap []interface{},
) {
	section := rules
	if replicaType != "" {
		section = jsonutils.AddSection(section, replicaType)
	}

	nodeTypeField := jsonutils.AddSection(section, nodeTypeFieldName)

	if nodeTypeField[optionsRules] == nil {
		nodeTypeField[optionsRules] = optionsAsMap
	}
}

func (req UnifiedWorkloadRequest) enrichNodePoolsFromProject(
	defaults map[string]interface{}, defaultNodePools []string,
	rules map[string]interface{}, resources []orgUnitApi.Resources,
) *workloadsApi.Error {
	//
	//  collect the complete set of node pools from the resources of the project
	//
	allNodePools := make([]string, 0, len(resources))

	for _, resource := range resources {
		if !convutils.HasValue(resource.NodePool) {
			continue
		}

		nodePool := resource.NodePool.Get()

		if convutils.IsStringHasValue(nodePool.Name) {
			allNodePools = append(allNodePools, *nodePool.Name)
		}
	}

	//
	// policy can contain non-existing node pools. we have to clean it before returning
	// it to the response of the unified API
	//
	allNodePoolsSet := datastructs.NewSet(allNodePools...)

	if defaults[nodePoolsFieldName] != nil {
		validNodePools := make([]string, 0, len(defaults))

		for _, nodePool := range convutils.ArrayMe(defaults[nodePoolsFieldName]) {
			if allNodePoolsSet.Has(nodePool.(string)) {
				validNodePools = append(validNodePools, nodePool.(string))
			}
		}

		defaults[nodePoolsFieldName] = validNodePools
	}

	if req.WorkloadType == api.WORKLOADTYPE_DISTRIBUTED {
		req.enrichNodePoolsOfReplica(workerSection, defaults, defaultNodePools, rules, allNodePools)
		req.enrichNodePoolsOfReplica(masterSection, defaults, defaultNodePools, rules, allNodePools)
	} else {
		req.enrichNodePoolsOfReplica("", defaults, defaultNodePools, rules, allNodePools)
	}

	return nil
}

func (req UnifiedWorkloadRequest) enrichNodePoolsOfReplica(replicaType string,
	defaults map[string]interface{}, defaultNodePools []string,
	rules map[string]interface{}, allNodePools []string,
) {
	//
	//  the default node pools contain the (partial) list of nodes that the
	//  project is associated with. it is stored in the default section of the policy
	//
	if len(defaultNodePools) != 0 {
		defaultSection := defaults
		if replicaType != "" {
			defaultSection = jsonutils.AddSection(defaultSection, replicaType)
		}

		if defaultSection[nodePoolsFieldName] == nil {
			nodePoolsAsInterface := make([]interface{}, len(defaultNodePools))
			for i, nodePool := range defaultNodePools {
				nodePoolsAsInterface[i] = nodePool
			}
			defaultSection[nodePoolsFieldName] = nodePoolsAsInterface
		}
	}

	//
	//  the all node pools is the complete list, and it is stored as options in
	//  the rules section of the policy
	//
	if len(allNodePools) != 0 {
		rulesSection := rules
		if replicaType != "" {
			rulesSection = jsonutils.AddSection(rulesSection, replicaType)
		}

		nodePoolsRules := jsonutils.AddSection(rulesSection, nodePoolsFieldName)

		if nodePoolsRules[optionsRules] == nil {
			nodesAsOptions := make([]map[string]interface{}, 0, len(allNodePools))

			for _, nodePool := range allNodePools {
				nodesAsOptions = append(nodesAsOptions, map[string]interface{}{
					optionValue: nodePool,
				})
			}

			nodePoolsRules[optionsRules] = nodesAsOptions
		}
	}
}

// clearContainerFieldsFromParent removes duplicate container-specific fields from the parent SupersetSpec
// for distributed inference workloads. Container fields (image, command, security, etc.) should only exist
// within the leader and worker subsections to maintain consistency with distributed training patterns.
//
// This function uses reflection to dynamically clear all fields except those explicitly preserved in the
// topLevelFields allowlist. This approach eliminates the need for manual maintenance when new container
// fields are added to SupersetSpec - they will be automatically cleared unless explicitly added to the
// allowlist.
//
// Safety guarantees:
//   - Panic recovery prevents runtime crashes from reflection operations
//   - Comprehensive unit tests validate behavior across field types
//   - Type and method validation before invoking reflection calls
func clearContainerFieldsFromParent(spec *workloadsApi.SupersetSpec) {
	if spec == nil {
		return
	}

	topLevelFields := datastructs.NewSet(
		"Leader", "Worker",
		"NodePools", "ServingPort", "Replicas", "Workers",
		"StartupPolicy", "RestartPolicy", "Category", "PriorityClass", "Preemptibility",
	)

	defer func() {
		if r := recover(); r != nil {
			logging.Errorf("reflection panic in clearContainerFieldsFromParent: %v", r)
		}
	}()

	specValue := reflect.ValueOf(spec)
	if specValue.Kind() != reflect.Ptr || specValue.IsNil() {
		return
	}

	specValue = specValue.Elem()
	if specValue.Kind() != reflect.Struct {
		return
	}

	specType := specValue.Type()

	for i := 0; i < specValue.NumField(); i++ {
		field := specValue.Field(i)
		fieldName := specType.Field(i).Name

		if topLevelFields.Has(fieldName) || !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct && field.CanAddr() {
			fieldAddr := field.Addr()
			unsetMethod := fieldAddr.MethodByName("Unset")
			if unsetMethod.IsValid() && unsetMethod.Type().NumIn() == 0 {
				unsetMethod.Call(nil)
			}
		} else if field.Kind() == reflect.Slice || field.Kind() == reflect.Ptr {
			if !field.IsNil() {
				field.Set(reflect.Zero(field.Type()))
			}
		}
	}
}

// populateDistributedInferenceReplicas converts the flat SupersetSpec from assets service into
// separate leader and worker subsections for distributed inference workloads.
func populateDistributedInferenceReplicas(spec *assetsApi.SupersetSpec) *workloadsApi.Error {
	asBytes, marshalErr := json.Marshal(spec)
	if marshalErr != nil {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to marshal asset spec: %v", marshalErr.Error()))
	}

	var replicaSpec assetsApi.DistributedInferenceLeaderWorkerSpec
	if unmarshalErr := json.Unmarshal(asBytes, &replicaSpec); unmarshalErr != nil {
		return workloadsApi.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to unmarshal replica spec: %v", unmarshalErr.Error()))
	}

	spec.Leader = *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpec(&replicaSpec)
	spec.Worker = *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpec(&replicaSpec)
	return nil
}
