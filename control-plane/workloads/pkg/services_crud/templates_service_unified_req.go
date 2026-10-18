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
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/patching"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

const (
	specSection  = "spec"
	systemChange = "system"
)

// UnifiedRequest allows us to implement one CRUD for all workload types
// it holds the data as a superset template and provides many useful utility functions
type UnifiedRequest struct {
	tenantId      uint
	workloadType  api.WorkloadType
	template      *api.SupersetTemplate
	migrating     bool
	routingMethod string
}

func NewUnifiedRequest(ctx context.Context, workloadType api.WorkloadType, template interface{}) (*UnifiedRequest, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, apiErr
	}

	result := UnifiedRequest{
		workloadType: workloadType,
		template:     &api.SupersetTemplate{},
		tenantId:     tenantId,
	}

	copyErr := convutils.CopySpecs(result.template, template)
	if copyErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, copyErr.Error(), copyErr)
	}

	values, inputErr := gojsonq.LoadJson(result.template)
	if inputErr != nil {
		return nil, api.NewValidationError(inputErr.Error())
	}
	if inputErr := gpucompat.ValidateMemoryRequestInput(values); inputErr != nil {
		return nil, api.NewValidationError(inputErr.Error())
	}

	var aliasErr *api.Error
	result.template, aliasErr = normalizeTemplateGPUMemory(result.template)
	if aliasErr != nil {
		return nil, aliasErr
	}

	return &result, nil
}

func (ur *UnifiedRequest) EntityKind() string {
	return fmt.Sprintf("%s-template", ur.workloadType)
}

func (ur *UnifiedRequest) ScopeInfo() *scopeinfo.ScopeInfo {
	return &scopeinfo.ScopeInfo{
		Scope:   scopeinfo.Scope(ur.template.Meta.ScopeType),
		ScopeId: ur.template.Meta.ScopeId,
	}
}

func (ur *UnifiedRequest) ScopeInfoForAuthValidation(requestingTenantId uint) *scopeinfo.ScopeInfo {
	if ur.template.Meta.ScopeType == api.SCOPETYPE_SYSTEM {
		return &scopeinfo.ScopeInfo{
			Scope:   scopeinfo.Scope(api.SCOPETYPE_TENANT),
			ScopeId: strconv.FormatUint(uint64(requestingTenantId), 10),
		}
	}

	return ur.ScopeInfo()
}

func (ur *UnifiedRequest) ValidateSpecs(ctx context.Context) (*assetsApi.AssetFlatResponse, mapAssetIdToReplicas, *api.Error) {
	linkedAssets := make(mapAssetIdToReplicas)

	collectAssets(linkedAssets, ur.template.Assets, workerReplica)

	if ur.workloadType != api.WORKLOADTYPE_DISTRIBUTED {
		if ur.template.MasterSpecSameAsWorker != nil {
			return nil, nil, api.NewValidationError("masterSpecSameAsWorker can be provided for distributed templates only")
		}

		if ur.template.Master != nil {
			return nil, nil, api.NewValidationError("master can be provided for distributed templates only")
		}
	} else {
		if validationErr := validateDistributedTemplateSegmentTopology(ur.template); validationErr != nil {
			return nil, nil, validationErr
		}

		if convutils.BoolIsTrue(ur.template.MasterSpecSameAsWorker) {
			if ur.template.Master != nil {
				return nil, nil, api.NewValidationError("master should not be provided when masterSpecSameAsWorker=true")
			}

			collectAssets(linkedAssets, ur.template.Assets, workerAndMaster)
		}

		masterProvided := convutils.BoolIsTrue(ur.template.MasterSpecSameAsWorker) || ur.template.Master != nil
		if masterProvided {
			if ur.template.Spec.DistributedFramework != nil && *ur.template.Spec.DistributedFramework == api.DISTRIBUTEDFRAMEWORK_JAX {
				return nil, nil, api.NewValidationError("master should not be provided when distributedFramework=JAX")
			}
		}

		if ur.template.Master != nil {
			collectAssets(linkedAssets, ur.template.Master.Assets, masterReplica)
		}
	}

	assetAsFlat, assetInfoMap, loadErr := requestAssetInfo(ctx, ur.workloadType, linkedAssets)
	if loadErr != nil {
		return nil, nil, loadErr
	}

	for assetId, assetMeta := range assetInfoMap {
		replicas := linkedAssets[assetId]

		if kindToCategory(assetMeta.Kind) != replicas.category {
			return nil, nil, api.NewValidationError(fmt.Sprintf("kind of asset '%s' is not %s", assetId, replicas.category))
		}

		//
		// template of scope X can only inherit from scope X or above, and its scope id must match
		// the scope id of the template or its ancestors
		//
		scopeTypeOk := true

		templateScopeId := ""
		assetScopeId := ""

		switch assetMeta.Scope {
		case assetsApi.SCOPE_CLUSTER:
			if ur.template.Meta.ClusterId == nil {
				scopeTypeOk = false
			} else {
				templateScopeId = *ur.template.Meta.ClusterId
				assetScopeId = convutils.Deref(assetMeta.ClusterId.Get(), "")
			}
		case assetsApi.SCOPE_DEPARTMENT:
			if ur.template.Meta.DepartmentId == nil {
				scopeTypeOk = false
			} else if *ur.template.Meta.DepartmentId != convutils.Deref(assetMeta.DepartmentId.Get(), "") {
				templateScopeId = *ur.template.Meta.DepartmentId
				assetScopeId = convutils.Deref(assetMeta.DepartmentId.Get(), "")
			}
		case assetsApi.SCOPE_PROJECT:
			if ur.template.Meta.ProjectId == nil {
				scopeTypeOk = false
			} else if *ur.template.Meta.DepartmentId != convutils.Deref(assetMeta.DepartmentId.Get(), "") {
				templateScopeId = fmt.Sprintf("%d", *ur.template.Meta.ProjectId)
				assetScopeId = fmt.Sprintf("%d", convutils.Deref(assetMeta.ProjectId.Get(), 0))
			}
		}

		if !scopeTypeOk {
			return nil, nil, api.NewValidationError(fmt.Sprintf("template of scope %s cannot link to an asset %s of scope %s",
				ur.template.Meta.ScopeType, assetId, assetMeta.Scope))
		}

		if templateScopeId != assetScopeId {
			return nil, nil, api.NewValidationError(fmt.Sprintf("template of %s %s cannot link to an asset of %s %s", assetMeta.Scope, templateScopeId, assetMeta.Scope, assetScopeId))
		}

		//
		//   for environment assets, need to match asset supported workload types
		//
		if assetMeta.Kind == assetsApi.ASSETKIND_ENVIRONMENT {
			workloadTypeErr := validateAssetVsWorkloadType(ctx, assetMeta, string(ur.workloadType))
			if workloadTypeErr != nil {
				return nil, nil, workloadTypeErr
			}
		}
	}

	return assetAsFlat, linkedAssets, nil
}

func validateDistributedTemplateSegmentTopology(template *api.SupersetTemplate) *api.Error {
	if template == nil {
		return nil
	}

	nativeSpec := &workloadApi.SupersetSpec{}
	if copyErr := convutils.CopySpecs(nativeSpec, &template.Spec); copyErr != nil {
		return api.NewError(http.StatusInternalServerError, "failed to validate distributed template segment topology", copyErr)
	}

	// A template may intentionally defer its framework to the workload request.
	// Validate segment semantics as PyTorch on this copy without making the
	// partial template runnable or changing the stored template.
	if hasDistributedSegmentIntent(nativeSpec) && template.Spec.DistributedFramework == nil {
		nativeSpec.SetDistributedFramework(workloadApi.DISTRIBUTEDFRAMEWORK_PY_TORCH)
	}

	if _, validationErr := normalizeDistributedSegmentTopology(nativeSpec); validationErr != nil {
		return api.NewValidationError(validationErr.Message)
	}

	if template.Master == nil || template.Master.Spec == nil {
		return nil
	}

	nativeMasterSpec := &workloadApi.MasterSpec{}
	if copyErr := convutils.CopySpecs(nativeMasterSpec, template.Master.Spec); copyErr != nil {
		return api.NewError(http.StatusInternalServerError, "failed to validate distributed template master segment topology", copyErr)
	}

	if validationErr := validateSeparateMasterSegmentAnnotations(
		nativeMasterSpec,
		convutils.BoolIsTrue(template.MasterSpecSameAsWorker),
	); validationErr != nil {
		return api.NewValidationError(validationErr.Message)
	}

	return nil
}

func validateAssetVsWorkloadType(ctx context.Context, assetMeta *assetsApi.AssetMeta, workloadType string) *api.Error {
	if !convutils.HasValue(assetMeta.WorkloadSupportedTypes) {
		return nil
	}

	//
	//  convert the types to a map containing workspace/training/distributed/inference as key and true/false as value
	//
	typesAsJson, convErr := gojsonq.LoadJson(assetMeta.WorkloadSupportedTypes.Get())
	if convErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("load workload supported types failed: %v", convErr)
		return nil
	}

	if typesAsJson[strings.ToLower(workloadType)] != true {
		return api.NewValidationError(fmt.Sprintf("asset %s is not supported for workload of type %s", assetMeta.Id, workloadType))
	}

	return nil
}

func (ur *UnifiedRequest) ToDbRecord() (*models_crud.Template, *api.Error) {
	result := models_crud.Template{
		ID: uuid.MustParse(ur.template.Meta.Id),

		TemplateMeta: models_crud.TemplateMeta{

			CreatedAt: ur.template.Meta.CreatedAt,
			CreatedBy: ur.template.Meta.CreatedBy,

			UpdatedAt: *ur.template.Meta.UpdatedAt,
			UpdatedBy: *ur.template.Meta.UpdatedBy,

			WorkloadType: string(ur.workloadType),
			Name:         ur.template.Meta.Name,
			Description:  ur.template.Meta.Description,

			TenantId:  ur.tenantId,
			ScopeType: string(ur.template.Meta.ScopeType),
			ScopeId:   ur.template.Meta.ScopeId,

			ClusterId:    ur.template.Meta.ClusterId,
			DepartmentId: ur.template.Meta.DepartmentId,

			RoutingMethod: ur.routingMethod,
		},

		Runnable:      ur.template.Runnable.IsRunnable,
		RunnableIssue: ur.template.Runnable.Issue,
	}

	if ur.template.Meta.ProjectId != nil {
		result.ProjectId = convutils.Ref(fmt.Sprintf("%d", *ur.template.Meta.ProjectId))
	}

	var jsonErr error

	result.Spec, jsonErr = gojsonq.LoadJson(ur.template.Spec)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse spec", jsonErr)
	}

	result.CombinedSpec, jsonErr = gojsonq.LoadJson(ur.template.CombinedSpec)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse spec", jsonErr)
	}

	result.MasterSameAsWorker = ur.template.MasterSpecSameAsWorker

	if ur.template.Master != nil && ur.template.Master.Spec != nil {
		result.MasterSpec, jsonErr = gojsonq.LoadJson(ur.template.Master.Spec)
		if jsonErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to parse master spec", jsonErr)
		}
	}

	if ur.template.CombinedMasterSpec != nil {
		result.CombinedMasterSpec, jsonErr = gojsonq.LoadJson(ur.template.CombinedMasterSpec)
		if jsonErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to parse master spec", jsonErr)
		}
	}

	for _, spec := range []*dbmodels.JsonBMap{&result.Spec, &result.MasterSpec, &result.CombinedSpec, &result.CombinedMasterSpec} {
		stored, err := gpucompat.LegacyGPUFields(*spec)
		if err != nil {
			return nil, api.NewValidationError(err.Error())
		}
		*spec = stored
	}

	return &result, nil
}

func NewUnifiedRequestFromRec(rec *models_crud.LoadedTemplate) (*UnifiedRequest, *api.Error) {
	_, workerAssets := rec.ContainedAssetsOfReplica(api.REPLICATYPE_WORKER)

	result := UnifiedRequest{
		workloadType: api.WorkloadType(rec.WorkloadType),

		template: &api.SupersetTemplate{
			Meta: api.TemplateMetaFields{
				Id: rec.ID.String(),

				Name:        rec.Name,
				Description: rec.Description,

				WorkloadType: api.WorkloadType(rec.WorkloadType),

				ScopeType: api.ScopeType(rec.ScopeType),
				ScopeId:   rec.ScopeId,

				ClusterId:    rec.ClusterId,
				DepartmentId: rec.DepartmentId,

				CreatedAt: rec.CreatedAt,
				CreatedBy: rec.CreatedBy,

				UpdatedAt: convutils.Ref(rec.UpdatedAt),
				UpdatedBy: convutils.Ref(rec.UpdatedBy),
			},

			Runnable: api.RunnableInfo{
				IsRunnable: rec.Runnable,
				Issue:      rec.RunnableIssue,
			},

			Spec:         rec.Spec,
			CombinedSpec: rec.CombinedSpec,

			MasterSpecSameAsWorker: rec.MasterSameAsWorker,

			Assets: workerAssets,
		},

		tenantId:      rec.TenantId,
		routingMethod: rec.RoutingMethod,
	}

	if rec.ProjectId != nil {
		projectIdInt, convErr := strconv.ParseInt(*rec.ProjectId, 10, 32)
		if convErr != nil {
			return nil, api.NewError(http.StatusBadRequest, "failed to load template: malformed project id", convErr)
		}

		result.template.Meta.ProjectId = convutils.Ref(int32(projectIdInt))
	}

	if rec.MasterSpec != nil {
		result.template.Master = &api.TemplateMasterSection{
			Spec: rec.MasterSpec,
		}
	}

	hasAssets, masterAssets := rec.ContainedAssetsOfReplica(api.REPLICATYPE_MASTER)

	if hasAssets {
		if result.template.Master == nil {
			result.template.Master = &api.TemplateMasterSection{}
		}
		result.template.Master.Assets = masterAssets
	}

	if rec.CombinedMasterSpec != nil {
		result.template.CombinedMasterSpec = rec.CombinedMasterSpec
	}

	var aliasErr *api.Error
	result.template, aliasErr = normalizeTemplateGPUMemory(result.template)
	if aliasErr != nil {
		return nil, aliasErr
	}

	return &result, nil
}

// RunnableInfo composes and returns the "is template runnable" information
func (ur *UnifiedRequest) RunnableInfo() api.RunnableInfo {
	// for workload type-specific validation
	switch ur.workloadType {
	case DistributedInference:
		// For distributed inference, leader image must be provided
		if ur.template.CombinedSpec.Leader == nil || convutils.IsEmptyOrNull(ur.template.CombinedSpec.Leader.Image) {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("leader.image is not provided"),
			}
		}

		// serving port needs to be provided
		if ur.template.CombinedSpec.ServingPort == nil || ur.template.CombinedSpec.ServingPort.Port == nil {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("servingPort is not provided"),
			}
		}
	case Distributed:
		// image must always be provided
		if convutils.IsEmptyOrNull(ur.template.CombinedSpec.Image) {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("image is not provided"),
			}
		}

		// distributed framework needs to be provided
		if ur.template.CombinedSpec.DistributedFramework == nil {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("distributedFramework is not provided"),
			}
		} else {
			distributedFramework := workloadApi.DistributedFramework(*ur.template.CombinedSpec.DistributedFramework)

			// for distributed framework that do not support no-master mode, master needs to be provided for
			// the template to be runnable
			if !convutils.BoolIsTrue(ur.template.MasterSpecSameAsWorker) {
				if ur.template.CombinedMasterSpec == nil && !frameworksSupportingNoMaster.Has(distributedFramework) {
					return api.RunnableInfo{
						IsRunnable: false,
						Issue:      convutils.Ref(fmt.Sprintf("master is not provided for %s framework", distributedFramework)),
					}
				}
			}
		}

		// if master is provided, image needs to be provided for the master in order for the template to be runnable
		if ur.template.CombinedMasterSpec != nil && convutils.IsEmptyOrNull(ur.template.CombinedMasterSpec.Image) {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("master.image is not provided"),
			}
		}
	case Inference:
		// image must always be provided
		if convutils.IsEmptyOrNull(ur.template.CombinedSpec.Image) {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("image is not provided"),
			}
		}

		// serving port needs to be provided
		if ur.template.CombinedSpec.ServingPort == nil ||
			(ur.template.CombinedSpec.ServingPort.Container == nil && ur.template.CombinedSpec.ServingPort.Port == nil) {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("servingPort is not provided"),
			}
		}
	default:
		// image must always be provided for other workload types (workspace, training)
		if convutils.IsEmptyOrNull(ur.template.CombinedSpec.Image) {
			return api.RunnableInfo{
				IsRunnable: false,
				Issue:      convutils.Ref("image is not provided"),
			}
		}
	}

	return api.RunnableInfo{
		IsRunnable: true,
	}
}

func (ur *UnifiedRequest) IsSystem() bool {
	return string(ur.template.Meta.ScopeType) == string(api.SCOPETYPE_SYSTEM)
}

func (ur *UnifiedRequest) UpdateTemplateFrom(ctx context.Context, srcRequest *UnifiedRequest) (mapAssetIdToReplicas, *api.Error) {
	src := srcRequest.template
	dest := ur.template

	dest.Meta.Name = src.Meta.Name
	dest.Meta.Description = src.Meta.Description
	ur.routingMethod = srcRequest.routingMethod

	dest.Meta.UpdatedAt = convutils.Ref(time.Now())

	// When migrating respect the provided UpdatedBy value
	// Otherwise, use value from context
	if ur.migrating {
		if dest.Meta.UpdatedBy == nil || *dest.Meta.UpdatedBy == "" {
			dest.Meta.UpdatedBy = runaictx.GetString(ctx, runaictx.CfSubjectID)
		}
	} else {
		dest.Meta.UpdatedBy = runaictx.GetString(ctx, runaictx.CfSubjectID)
	}

	dest.Assets = src.Assets
	dest.Spec = src.Spec
	dest.Master = src.Master
	dest.MasterSpecSameAsWorker = src.MasterSpecSameAsWorker

	assetSpecAsFlat, linkedAssets, specErr := ur.ValidateSpecs(ctx)
	if specErr != nil {
		return nil, specErr
	}

	if assetsGiven(dest) {
		mutateErr := mutateTemplate(dest, srcRequest.workloadType, assetSpecAsFlat)
		if mutateErr != nil {
			return nil, mutateErr
		}
	} else {
		dest.CombinedSpec = dest.Spec
		if dest.Master != nil {
			dest.CombinedMasterSpec = dest.Master.Spec
		} else {
			dest.CombinedMasterSpec = nil
		}
	}

	dest.Runnable = ur.RunnableInfo()

	return linkedAssets, nil
}

func (ur *UnifiedRequest) PatchTemplateFrom(ctx context.Context, srcRequest *UnifiedRequest, options updateOptions) (mapAssetIdToReplicas, *api.Error) {
	src := srcRequest.template
	dest := ur.template

	if options.patchMetaFields != nil {
		if options.patchMetaFields.Name != nil {
			dest.Meta.Name = *options.patchMetaFields.Name
		}

		patching.PatchField(&dest.Meta.Description, options.patchMetaFields.Description)
	}

	dest.Meta.UpdatedAt = convutils.Ref(time.Now())

	// When migrating respect the provided UpdatedBy value
	// Otherwise, use value from context
	if ur.migrating {
		if dest.Meta.UpdatedBy == nil || *dest.Meta.UpdatedBy == "" {
			dest.Meta.UpdatedBy = runaictx.GetString(ctx, runaictx.CfSubjectID)
			if dest.Meta.UpdatedBy == nil {
				dest.Meta.UpdatedBy = convutils.Ref(systemChange)
			}
		}
	} else {
		dest.Meta.UpdatedBy = runaictx.GetString(ctx, runaictx.CfSubjectID)
		if dest.Meta.UpdatedBy == nil {
			dest.Meta.UpdatedBy = convutils.Ref(systemChange)
		}
	}

	if src != nil {
		if src.MasterSpecSameAsWorker != nil {
			dest.MasterSpecSameAsWorker = src.MasterSpecSameAsWorker
		}

		patching.ApplyPatch(&dest.Spec, &src.Spec)

		// the assets field is patched as a whole
		if src.Assets != nil {
			dest.Assets = src.Assets
		}

		if dest.Meta.WorkloadType == Distributed && src.Master != nil {
			if src.Master.Spec != nil {
				if dest.Master == nil {
					dest.Master = src.Master
				} else if dest.Master.Spec == nil {
					dest.Master.Spec = src.Master.Spec
				} else {
					patching.ApplyPatch(dest.Master.Spec, src.Master.Spec)
				}
			}

			if src.Master.Assets != nil {
				if dest.Master == nil {
					dest.Master = &api.TemplateMasterSection{}
				}
				dest.Master.Assets = src.Master.Assets
			}
		}
	}

	clearErr := ur.ClearFields(options)
	if clearErr != nil {
		return nil, clearErr
	}

	assetSpecAsFlat, linkedAssets, specErr := ur.ValidateSpecs(ctx)
	if specErr != nil {
		return nil, specErr
	}

	if assetsGiven(dest) {
		mutateErr := mutateTemplate(dest, srcRequest.workloadType, assetSpecAsFlat)
		if mutateErr != nil {
			return nil, mutateErr
		}
	} else {
		dest.CombinedSpec = dest.Spec
		if dest.Master != nil {
			dest.CombinedMasterSpec = dest.Master.Spec
		} else {
			dest.CombinedMasterSpec = nil
		}
	}

	dest.Runnable = ur.RunnableInfo()

	return linkedAssets, nil
}

func (ur *UnifiedRequest) ClearFields(options updateOptions) *api.Error {
	if len(options.fieldsToClear) == 0 {
		return nil
	}

	specAsJson, jsonErr := gojsonq.LoadJson(ur.template.Spec)
	if jsonErr != nil {
		return api.NewError(http.StatusInternalServerError, "failed to parse spec", jsonErr)
	}

	// we want to mimic this payload:
	//   spec:
	//     <worker spec fields>
	//   master:
	//     spec:
	//       <master spec fields>
	templateAsJson := map[string]interface{}{
		specSection: specAsJson,
	}

	hasMaster := false

	if ur.template.Master != nil && ur.template.Master.Spec != nil {
		masterSpecAsJson, jsonErr := gojsonq.LoadJson(ur.template.Master.Spec)
		if jsonErr != nil {
			return api.NewError(http.StatusInternalServerError, "failed to parse spec", jsonErr)
		}

		templateAsJson[masterSection] = map[string]interface{}{
			specSection: masterSpecAsJson,
		}

		hasMaster = true
	}

	fieldsToClear := make([]string, len(options.fieldsToClear))
	for i, field := range options.fieldsToClear {
		fieldsToClear[i] = gpucompat.CanonicalGPUFieldPath(field)
	}
	clearErr := patching.ClearFields(templateAsJson, fieldsToClear, patching.ClearFieldsOptions{})
	if clearErr != nil {
		return api.NewValidationError(clearErr.Error())
	}

	specAsJson = convutils.MapMe(templateAsJson[specSection])

	var result api.SupersetSpec
	jsonErr = json.Unmarshal(gojsonq.ToBytes(specAsJson), &result)
	if jsonErr != nil {
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to clear spec fields: %s", jsonErr.Error()), jsonErr)
	}

	ur.template.Spec = result

	if hasMaster {
		masterAsJson := convutils.MapMe(templateAsJson[masterSection])
		masterSpecAsJson := convutils.MapMe(masterAsJson[specSection])

		var masterResult api.MasterTemplateSpec

		jsonErr = json.Unmarshal(gojsonq.ToBytes(masterSpecAsJson), &masterResult)
		if jsonErr != nil {
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to clear master.spec fields: %s", jsonErr.Error()), jsonErr)
		}

		ur.template.Master.Spec = &masterResult
	}

	return nil
}
