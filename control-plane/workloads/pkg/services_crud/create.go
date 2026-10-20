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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/notifications"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_crud/validators"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/datavolumes"

	k8s_objects_tracker "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/k8s-objects-tracker"
	k8sObjectsClient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client"
	k8sObjectsApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"

	assetsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"

	clients "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/backend"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"

	logging "github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_integration"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/workloadutils"
)

const (
	Workspace            = "Workspace"
	Training             = "Training"
	Distributed          = "Distributed"
	Inference            = "Inference"
	DistributedInference = "DistributedInference"

	RunaiNsPrefix = "runai-"

	NoneStorageClass = "runai-none"
)

// MaxWorkloadLength defines the limit for workload names due to k8s label length constraint- pod name must be no more than 63 characters.
const (
	MaxWorkloadLength      = 50
	priorityClassNameLabel = "priorityClassName"
)

const (
	MinNodePort = 30000
	MaxNodePort = 32767
)

var (
	supersetWorkloadJsonCracker = jsonutils.NewJsonCracker(workloadsApi.DistributedSpecSpec{})
	distributedFrameworkField   = supersetWorkloadJsonCracker.NameOf(&supersetWorkloadJsonCracker.Struct().DistributedFramework)
)

var permissionError = workloadsApi.NewSubmissionErrorV2(http.StatusForbidden, "insufficient permissions")

var (
	computeJsonCracker = jsonutils.NewJsonCracker(workloadsApi.WorkspaceSpecSpecAllOfCompute{})

	fldNameGpuDevicesRequest       = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuDevicesRequest)
	fldNameGpuMemoryPortionRequest = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuMemoryPortionRequest)
	fldNameGpuMemoryPortionLimit   = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuMemoryPortionLimit)
	fldNameGpuMemoryRequest        = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuMemoryRequest)
	fldNameGpuMemoryLimit          = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuMemoryLimit)
	fldNameGpuRequestType          = computeJsonCracker.NameOf(&computeJsonCracker.Struct().GpuRequestType)
)

var frameworksSupportingNoMaster = datastructs.NewSet(workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH, workloadsApi.DISTRIBUTEDFRAMEWORK_TF, workloadsApi.DISTRIBUTEDFRAMEWORK_JAX)

type createWorkloadParams struct {
	AcceptedAt   time.Time
	WorkloadType string
	WorkloadId   uuid.UUID

	TemplateId *string

	// for submisison with assets, we need the asset ids
	// for workspace, training and inference - AssetIds holds the (optional) set of assets
	// for distributed - AssetIds holds the assets of the worker, masterAssetsId the assets to use for the master
	// for distributedInference - AssetIds holds the assets of the worker, LeaderAssetsIds the assets for the leader
	AssetIds       *api.AssetSubmissionIds
	MasterAssetIds *api.AssetSubmissionIds
	LeaderAssetIds *api.AssetSubmissionIds

	UseGivenNameAsPrefix bool

	TenantId int32

	Spec map[string]interface{}

	Response workloadsApi.SupersetWorkload
}

func createWorkload(ctx context.Context, backendClient clients.TenantSettingsClient, params *createWorkloadParams, workloadType string) *workloadsApi.SubmissionErrorV2 {
	if err := gpucompat.ValidateMemoryRequestInput(params.Spec); err != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, err.Error())
	}
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return ToSubmissionError(apiErr)
	}

	project, orgUnitErr := orgUnitClient.Get().GetProjectById(ctx, params.Response.ProjectId)
	if orgUnitErr != nil {
		return NewSubmissionErrorWithDetails(orgUnitErr.Code, orgUnitErr.Message, nil)
	}

	namespace := convutils.Deref(project.RequestedNamespace.Get(), "")
	if namespace == "" {
		namespace = RunaiNsPrefix + project.Name
	}

	ancestors := orgUnitApi.NewAncestors(strconv.Itoa(int(tenantId)))
	ancestors.SetClusterUuid(project.ClusterId)
	ancestors.SetDepartmentId(project.ParentId)
	ancestors.SetProjectId(project.Id)

	_, authErr := validateCreationAuth(ctx, ancestors, params.Response.ProjectId)
	if authErr != nil {
		return authErr
	}

	//
	//   make sure the cluster provided by the user is correct. in the future the cluster id may be optional
	//   but for now we are requesting the user to provide it and need to make sure it is correct.
	//
	if ancestors.ClusterUuid.IsSet() {
		clusterId := *ancestors.ClusterUuid.Get()
		if clusterId != params.Response.ClusterId {
			return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("project %s is not associated with cluster %s",
				params.Response.ProjectId, params.Response.ClusterId))
		}
	}

	params.TenantId = int32(tenantId)
	params.WorkloadId = uuid.New()

	params.Response.CreatedAt = time.Now().Truncate(time.Microsecond)
	params.Response.CreatedBy = convutils.Deref(runaictx.GetString(ctx, runaictx.CfSubjectID), "")
	params.Response.WorkloadId = params.WorkloadId.String()

	if ancestors.DepartmentId.IsSet() {
		params.Response.DepartmentId = ancestors.DepartmentId.Get()
	}

	if params.UseGivenNameAsPrefix {
		params.Response.Name = GenerateJobNameByPrefix(params.Response.RequestedName)
	} else {
		params.Response.Name = params.Response.RequestedName
	}

	namingErrs := validation.IsDNS1035Label(params.Response.Name)
	if len(namingErrs) > 0 {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "The workload name must consist of lower case alphanumeric characters or '-', and start with an alphabetic character.")
	}
	if len(params.Response.Name) > MaxWorkloadLength {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("The length of a workload name must not exceed %d characters", MaxWorkloadLength))
	}

	//
	//   check if such name already exists in the cluster. if cannot check it, do not fail the workload just log
	//
	nameInUse, validationErr := services_integration.IsWorkloadNameInUse(ctx, params.Response.ClusterId, params.Response.ProjectId, params.WorkloadType, params.Response.Name)
	if validationErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to validate if name %s is taken: %v", params.Response.Name, validationErr.Message)
	} else if nameInUse {
		return workloadsApi.NewSubmissionErrorV2(http.StatusConflict, fmt.Sprintf("%s named '%s' already exists in the project",
			params.WorkloadType, params.Response.Name))
	}

	effectivePolicy, policyErr := policy_compliance.GetEffectivePolicyOfProject(ctx, policies.Get().Client(), params.Response.ProjectId, workloadsApi.WorkloadType(params.WorkloadType))
	if policyErr != nil {
		return &workloadsApi.SubmissionErrorV2{
			Code:    policyErr.Code,
			Message: policyErr.Message,
		}
	}

	masterSameAsWorker := convutils.BoolIsTrue(params.Response.MasterSpecSameAsWorker.Get())

	// the permittedScoped we already have are permitted scope for creating a template.
	// they are totally different from the permitted scopes for reading a template, so
	// we have to load the read permissions specifically
	permittedScopesForRead, rbacErr := auth.LoadAllowedOrgUnits(ctx)
	if rbacErr != nil {
		return NewSubmissionErrorWithDetails(rbacErr.Code, rbacErr.Message, nil)
	}

	assetsSpec, assetsErr := loadAssets(ctx, params, masterSameAsWorker)
	if assetsErr != nil {
		return ToSubmissionError(assetsErr)
	}

	templateSpec, templateErr := loadTemplate(ctx, params, masterSameAsWorker, permittedScopesForRead.Raw(), ancestors)
	if templateErr != nil {
		return ToSubmissionError(templateErr)
	}

	// Mutate the workload (apply template + assets + policy defaults)
	mutatedSpec, mutateErr := MutateWorkload(params.Spec, effectivePolicy.DefaultsV2, masterSameAsWorker, templateSpec, assetsSpec)
	if mutateErr != nil {
		return mutateErr
	}

	// Even though the prepare for mutation is primarily needed for mutation function, we also need to call it before validateSpecAgainstPolicy.
	// - Transforms itemized fields from policy format (with .instances and .attributes sections) to workload format (arrays)
	// - Adds policy keys (nicknames) for itemized instances without keys
	// - Adds field name as a dummy instance for merging purposes
	// This transformation is critical because compliance validation (itemKeys function) expects both
	// spec and defaults to have the same structure when comparing keys for canAdd/locked validation.
	policy_compliance.PrepareDefaultsForMutation(effectivePolicy.DefaultsV2)

	// Validate policy compliance on the mutated workload
	complianceErr := validateSpecAgainstPolicy(params.WorkloadType, effectivePolicy.RulesV2, effectivePolicy.DefaultsV2, mutatedSpec)
	if complianceErr != nil {
		return complianceErr
	}

	// Build responseSpec from mutatedSpec (includes template + policy defaults)
	responseSpec, responseMasterSpec, rebuildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](workloadType, mutatedSpec)
	if rebuildErr != nil {
		return rebuildErr
	}

	responseSpec, priorityClassError := applyPriorityClassLabelWorkaround(ctx, responseSpec, workloadType, params.Response.ClusterId)
	if priorityClassError != nil {
		return workloadsApi.NewSubmissionErrorV2(priorityClassError.Code, priorityClassError.Message)
	}

	priorityValues, customPriorityErr := validators.ValidateCustomPriorityClass(ctx, params.Response.ClusterId, responseSpec.PriorityClass)
	if customPriorityErr != nil {
		return customPriorityErr
	}

	categoryValidationErr := validators.ValidateWorkloadCategory(ctx, responseSpec.Category)
	if categoryValidationErr != nil {
		return categoryValidationErr
	}

	//
	//   assign default category
	//
	if !convutils.HasValue(responseSpec.Category) || (responseMasterSpec != nil && !convutils.HasValue(responseMasterSpec.Category)) {
		categoryId, categoryName, catErr := services.GetCategoryIdByWorkloadTypeAndGroup(ctx, tenantId, common.NewWorkloadType(workloadType), services.RunaiCrdWorkloadTypeGroup)
		if catErr != nil || categoryId == nil {
			return ToSubmissionError(catErr)
		}

		if !convutils.HasValue(responseSpec.Category) {
			responseSpec.Category.Set(&categoryName)

			if masterSameAsWorker {
				responseMasterSpec.Category.Set(&categoryName)
			}
		}

		if responseMasterSpec != nil && !convutils.HasValue(responseMasterSpec.Category) {
			responseMasterSpec.Category.Set(&categoryName)
		}
	}

	storageClassConfigs, claimPerDvId, preValidationErr := preValidateSpec(ctx, params, responseSpec, responseMasterSpec)
	if preValidationErr != nil {
		return preValidationErr
	}

	scErr := workloadutils.ValidateAndHandelServingConfigurationFields(ctx, responseSpec, params.Response.ClusterId)
	if scErr != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, scErr.Error())
	}

	assetsMapping, credentialsProcessErr := processCredentials(ctx, params.Response.ClusterId, responseSpec, responseMasterSpec)
	if credentialsProcessErr != nil {
		return credentialsProcessErr
	}

	params.Response.Spec = responseSpec

	if workloadType == DistributedInference {
		if setTokenErr := setTokenUIDGIDForDistributedInference(ctx, responseSpec); setTokenErr != nil {
			return workloadsApi.NewSubmissionErrorV2(setTokenErr.Code, setTokenErr.Message)
		}
	}

	security, setTokenErr := setTokenUIDGIDForSecurity(ctx, params.Response.Spec.GetSecurity())
	if setTokenErr != nil {
		return workloadsApi.NewSubmissionErrorV2(setTokenErr.Code, setTokenErr.Message)
	}
	if security != nil {
		params.Response.Spec.SetSecurity(*security)
	}

	if responseMasterSpec != nil {
		params.Response.MasterSpec = *workloadsApi.NewNullableMasterSpec(responseMasterSpec)
		security, setTokenErr = setTokenUIDGIDForSecurity(ctx, params.Response.MasterSpec.Get().GetSecurity())
		if setTokenErr != nil {
			return workloadsApi.NewSubmissionErrorV2(setTokenErr.Code, setTokenErr.Message)
		}
		if security != nil {
			masterSpec := params.Response.MasterSpec.Get()
			masterSpec.SetSecurity(*security)
			params.Response.MasterSpec = *workloadsApi.NewNullableMasterSpec(masterSpec)
		}
	}

	if scopeErr := validateSegmentTopologyScope(workloadType, responseSpec, responseMasterSpec); scopeErr != nil {
		return NewSubmissionErrorWithDetails(scopeErr.Code, scopeErr.Message, scopeErr.Details)
	}

	client := clusterservice.Get()

	actualNodePoolsResponse, getErr := client.GetNodePoolsInternalByCluster(ctx, params.Response.ClusterId)
	if getErr != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, getErr.Error())
	}

	actualNodePools := []clusterApi.Nodepool{}
	if actualNodePoolsResponse != nil {
		actualNodePools = actualNodePoolsResponse.Nodepools
	}

	setDefaultNodepoolsIfNecessary(ctx, &params.Response, project)

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(params.Response.ClusterId)
	segmentRouting := segmentedDistributedRoutingDecision{}
	var segmentResponseState *segmentResponseSanitizationState
	if workloadType == Distributed {
		segmentRouting = segmentedDistributedSubmissionRouting(ctx, params.TenantId, responseSpec, clusterCapabilityChecker)
		if segmentRouting.err != nil {
			return NewSubmissionErrorWithDetails(segmentRouting.err.Code, segmentRouting.err.Message, segmentRouting.err.Details)
		}
		segmentResponseState = captureSegmentResponseSanitizationState(responseSpec, responseMasterSpec)
	}

	nodePoolsErr := handleNodePoolsAndTopology(ctx, workloadType, actualNodePools, responseSpec, responseMasterSpec, masterSameAsWorker, segmentRouting)
	if nodePoolsErr != nil {
		return NewSubmissionErrorWithDetails(http.StatusBadRequest, nodePoolsErr.Message, nodePoolsErr.Details)
	}

	options, checkerErr := determineComposerOptions(ctx, clusterCapabilityChecker)
	if checkerErr != nil {
		return checkerErr
	}

	complimentaryAssetInfo, err := composer.LoadComplimentaryAssetInfo(params.Response.Spec.Annotations, true)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to load asset info: %v", err)
	}

	notificationsComposer := notifications.NewNotificationsComposer(
		workloadType,
		params.Response.Name,
		params.Response.WorkloadId,
		params.Response.CreatedBy,
		params.Response.ClusterId,
		params.Response.ProjectId,
		convutils.Deref(params.Response.DepartmentId, ""),
		params.Response.CreatedAt.Truncate(time.Microsecond),
		nil,
		options).WithConnectionsPreProcessor(project.Name, "")

	wc := &workloadContext{
		// Common fields
		workloadType:         workloadType,
		workloadName:         params.Response.Name,
		createdBy:            params.Response.CreatedBy,
		tenantId:             params.TenantId,
		spec:                 responseSpec,
		options:              options,
		clusterCapabilities:  clusterCapabilityChecker,
		segmentRouting:       &segmentRouting,
		storageClassConfigs:  storageClassConfigs,
		dataVolumeClaimNames: claimPerDvId,
		masterSpec:           responseMasterSpec,
		masterSameAsWorker:   masterSameAsWorker,
		// Create-specific context
		create: &workloadCreateContext{
			createParams:           params,
			project:                project,
			namespace:              namespace,
			assetsMapping:          assetsMapping,
			complimentaryAssetInfo: complimentaryAssetInfo,
			notificationsComposer:  notificationsComposer,
			actualNodePools:        actualNodePools,
		},
	}

	if params.WorkloadType == string(api.WORKLOADTYPE_INFERENCE) {
		wc.inferenceSpec = &workloadsApi.InferenceSpecSpec{}
		err := convutils.CopySpecs(wc.inferenceSpec, params.Response.Spec)
		if err != nil {
			return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to copy inference spec: %v", err))
		}
	}

	if params.WorkloadType == string(api.WORKLOADTYPE_DISTRIBUTED_INFERENCE) {
		wc.distributedInferenceSpec = &workloadsApi.DistributedInferenceSpecSpec{}
		err := convutils.CopySpecs(wc.distributedInferenceSpec, params.Response.Spec)
		if err != nil {
			return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to copy distributed inference spec: %v", err))
		}
	}

	// Try the any-workload submitter for this workload type. If the submitter can handle create
	// and succeeds, submit via workloads-manager. Otherwise fall back to native submission.
	var anyWorkloadSubmission *Submission
	var workloadRequirements composer.WorkloadRequirements

	// Determine which submission path this cluster/tenant/workload-type combination requires.
	// When no submitter is registered for the type, keep the historical behavior of the native path.
	submitter := getAnyWorkloadSubmitter(wc)
	if submitter == nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("workload type %s is not supported for submission", params.WorkloadType))
	}

	method := submitter.DetermineSubmissionMethods(ctx, wc)

	var submitterErr *workloadsApi.SubmissionErrorV2

	if method.AnyWorkload {
		anyWorkloadSubmission, submitterErr = submitter.Create(ctx, wc)
		if submitterErr != nil && submitterErr.Code != http.StatusNotImplemented {
			return submitterErr
		}
	}

	switch {
	case anyWorkloadSubmission != nil:
		// The any-workload path produced a submission - use its requirements.
		workloadRequirements = anyWorkloadSubmission.Requirements
	case method.Native:
		// No any-workload submission was produced - either the cluster/tenant is not on the
		// any-workload path, or the submitter cannot (yet) handle this spec. Fall back to
		// native submission, which is still available on this cluster.
		var nativeErr *workloadsApi.SubmissionErrorV2
		workloadRequirements, nativeErr = composeNative(ctx, wc)
		if nativeErr != nil {
			return nativeErr
		}
	default:
		// Native submission is not available on this cluster (>= 3.0) and the any-workload
		// path did not produce a submission. Do not silently bypass the composer - fail
		// instead of persisting an invalid workload.
		if submitterErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s workload %s is not supported for any-workload submission: %s: %v",
				params.WorkloadType, params.WorkloadId, submitterErr.Message, submitterErr.Details)
			return submitterErr
		}
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to submit %s workload", params.WorkloadType))
	}

	if requirementsErr := validateRequirements(ctx, backendClient, &workloadRequirements); requirementsErr != nil {
		return requirementsErr
	}

	// Update the response with the proper preemptibility value
	if priorityValues != nil {
		setPreemptibilityFromPriority(responseSpec, priorityValues)
		if responseMasterSpec != nil {
			setMasterPreemptibilityFromPriority(responseMasterSpec, priorityValues)
		}
	}

	if anyWorkloadSubmission != nil {
		if submissionErr := sendSubmissionToWorkloadMgr(ctx, params, anyWorkloadSubmission); submissionErr != nil {
			return submissionErr
		}
		sanitizeSegmentedNativeResponse(responseSpec, responseMasterSpec, segmentResponseState)
		persistWorkloadV2Snapshot(ctx, wc)
		return nil
	}

	return persistAndRespondNative(ctx, wc, assetsMapping)
}

func persistWorkloadV2Snapshot(ctx context.Context, wc *workloadContext) {
	if wc == nil || wc.create == nil || wc.create.createParams == nil || wc.create.createParams.Response.Spec == nil {
		return
	}

	snapshot := NewWorkloadV2SnapshotFromWorkloadParams(wc.create.createParams, wc.create.assetsMapping, wc.create.namespace)
	// Note: Failing to persist the V2 workload snapshot here does not block or impact the main workload creation flow;
	// the error is logged but not returned. However, it may have downstream impact: users who later attempt to make
	// a copy of this V2 workload will rely on the snapshot's existence. If snapshot persistence fails, subsequent copy
	// operations for this workload may not succeed as expected.
	if dbErr := dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot); dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to persist V2 workload snapshot for workload %s: %v", snapshot.ID, dbErr.Message)
	}
}

func workloadsClientFieldToServerFields(clientFields workloadsApi.WorkloadCreationFields) (api.WorkloadCreationFields, error) {
	requestedResources := api.WorkloadCreationFields{}
	bytes, err := json.Marshal(clientFields)
	if err != nil {
		logging.Errorf("Failed to marshal client fileds struct. err: %v", err)
		return requestedResources, err
	}
	err = json.Unmarshal(bytes, &requestedResources)
	if err != nil {
		logging.Errorf("Failed to unmarshal client fileds struct. err: %v", err)
		return requestedResources, err
	}

	return requestedResources, nil
}

func validateCreationAuth(ctx context.Context, ancestors *orgUnitApi.Ancestors, projectId string) (*authorization_api.PermittedScopes, *workloadsApi.SubmissionErrorV2) {
	//   get the permitted scopes from the context
	permittedScopes, scopesErr := authorization_client.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		return nil, permissionError
	}

	if slices.Contains(permittedScopes.Projects, projectId) {
		return permittedScopes, nil
	}

	if permittedScopes.System != nil && *permittedScopes.System {
		return permittedScopes, nil // get permission for the entire system
	}

	if permittedScopes.Tenants != nil {
		if slices.Contains(permittedScopes.Tenants, ancestors.TenantId) {
			return permittedScopes, nil // get permission for the entire tenant
		}
	}

	if ancestors.ClusterUuid.IsSet() {
		if slices.Contains(permittedScopes.Clusters, *ancestors.ClusterUuid.Get()) {
			return permittedScopes, nil // get permission to the entire cluster
		}
	}

	if ancestors.DepartmentId.IsSet() {
		if slices.Contains(permittedScopes.Departments, *ancestors.DepartmentId.Get()) {
			return permittedScopes, nil // get permission to the entire department
		}
	}

	//  no permission to any of the above, reject the request
	return permittedScopes, permissionError
}

func validateRequirements(ctx context.Context, backendClient clients.TenantSettingsClient, req *composer.WorkloadRequirements) *workloadsApi.SubmissionErrorV2 {
	if req.GpuOptimization {
		enabled, getErr := backendClient.GetGpuResourceOptimization(ctx)
		if getErr != nil {
			// since this is only validation, we don't want to fail the submission if we could not do the validation
			// we allow the workload to continue, and worst case it will possibly fail later on
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to check tenant settings: %v", getErr)
			return nil
		}

		if !enabled {
			return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "gpuPortionLimit and gpuMemoryLimit require GPU Optimization settings to be enabled")
		}
	}

	return nil
}

// this workaround supports the old composer validation that requires the priority class to be set with new training values
// that old composer does not support. we add the priority class to the labels and unset it from the spec to avoid the cluster composer validation
func applyPriorityClassLabelWorkaround(ctx context.Context, responseSpec *workloadsApi.SupersetSpec, workloadType, clusterId string) (*workloadsApi.SupersetSpec, *api.Error) {
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	customPriorityClusterSupport, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfFlexiblePriorityClasses)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, err.Error(), nil)
	}
	if customPriorityClusterSupport {
		return responseSpec, nil
	}

	if workloadType != Training || !convutils.HasValue(responseSpec.PriorityClass) {
		return responseSpec, nil
	}

	// MinVersionTrainingPriorityClassPatchFix (2.18) is below the minimum supported cluster
	// version, so the patch-fix is always available — apply the priority-class label hack
	// unconditionally for clusters that lack flexible priority classes.
	value := responseSpec.PriorityClass.Get()
	valErr := composer.ValidateStaticPriorityClass(*value, workloadType) // todo: remove in new clusters
	if valErr != nil {
		return nil, api.NewError(int(valErr.Code), valErr.Message, nil)
	}

	// add the priority class to the labels
	responseSpec.Labels = append(responseSpec.Labels, workloadsApi.Label{
		Name:  *workloadsApi.NewNullableString(convutils.Ref(priorityClassNameLabel)),
		Value: *workloadsApi.NewNullableString((*string)(value)),
	})

	// unset the priority class (to prevent the composer from failing - from the cluster old composer validation)
	responseSpec.PriorityClass.Unset()

	return responseSpec, nil
}

func setPreemptibilityFromPriority(spec *workloadsApi.SupersetSpec, priorityValues map[string]int32) {
	if spec == nil || spec.Preemptibility.IsSet() || !convutils.HasValue(spec.PriorityClass) {
		return
	}
	if value, found := priorityValues[*spec.PriorityClass.Get()]; found {
		if value < services.PriorityPreemptionThreshold {
			spec.SetPreemptibility(workloadsApi.PREEMPTIBILITY_PREEMPTIBLE)
		} else {
			spec.SetPreemptibility(workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE)
		}
	}
}

func setMasterPreemptibilityFromPriority(spec *workloadsApi.MasterSpec, priorityValues map[string]int32) {
	if spec == nil || spec.Preemptibility.IsSet() || !convutils.HasValue(spec.PriorityClass) {
		return
	}
	if value, found := priorityValues[*spec.PriorityClass.Get()]; found {
		if value < services.PriorityPreemptionThreshold {
			spec.SetPreemptibility(workloadsApi.PREEMPTIBILITY_PREEMPTIBLE)
		} else {
			spec.SetPreemptibility(workloadsApi.PREEMPTIBILITY_NON_PREEMPTIBLE)
		}
	}
}

// Distributed inference stores security inside each replica rather than at spec.security.
func setTokenUIDGIDForDistributedInference(ctx context.Context, spec *workloadsApi.SupersetSpec) *api.Error {
	for _, replica := range []*workloadsApi.DistributedInferenceLeaderWorkerSpec{spec.Leader.Get(), spec.Worker.Get()} {
		if replica == nil || replica.Security.Get() == nil {
			continue
		}
		security := replica.Security.Get()
		resolved, err := setTokenUIDGIDForSecurity(ctx, workloadsApi.WorkspaceSpecSpecAllOfSecurity{
			UidGidSource:       security.UidGidSource,
			RunAsUid:           security.RunAsUid,
			RunAsGid:           security.RunAsGid,
			SupplementalGroups: security.SupplementalGroups,
		})
		if err != nil {
			return err
		}
		if resolved != nil {
			security.RunAsUid = resolved.RunAsUid
			security.RunAsGid = resolved.RunAsGid
			security.SupplementalGroups = resolved.SupplementalGroups
		}
	}
	return nil
}

func setTokenUIDGIDForSecurity(ctx context.Context, security workloadsApi.WorkspaceSpecSpecAllOfSecurity) (*workloadsApi.WorkspaceSpecSpecAllOfSecurity, *api.Error) {
	if !convutils.HasValue(security.UidGidSource) || *security.UidGidSource.Get() != workloadsApi.UIDGIDSOURCE_FROM_IDP_TOKEN {
		// Only return security if it has any values set
		if security.RunAsUid.IsSet() || security.RunAsGid.IsSet() || security.SupplementalGroups.IsSet() {
			return &security, nil
		}
		return nil, nil
	}

	// Handle UID
	uidFromContext := runaictx.GetString(ctx, runaictx.CfUId)
	if uidFromContext == nil {
		if security.RunAsUid.IsSet() {
			security.RunAsUid = *workloadsApi.NewNullableInt64(security.RunAsUid.Get())
		}
	} else {
		if uidInt, err := strconv.ParseInt(*uidFromContext, 10, 64); err == nil {
			security.RunAsUid = *workloadsApi.NewNullableInt64(convutils.Ref(uidInt))
		} else {
			logging.Errorf("failed to parse uid claim: %v", err)
		}
	}

	// Handle GID
	gidFromContext := runaictx.GetString(ctx, runaictx.CfGId)
	if gidFromContext == nil {
		if security.RunAsGid.IsSet() {
			security.RunAsGid = *workloadsApi.NewNullableInt64(security.RunAsGid.Get())
		}
	} else {
		if gidInt, err := strconv.ParseInt(*gidFromContext, 10, 64); err == nil {
			security.RunAsGid = *workloadsApi.NewNullableInt64(convutils.Ref(gidInt))
		} else {
			logging.Errorf("failed to parse gid claim: %v", err)
		}
	}

	// Handle supplemental groups
	suppGroupsFromContext := runaictx.GetString(ctx, runaictx.CfSupplementalGroups)
	if suppGroupsFromContext != nil && len(*suppGroupsFromContext) > 0 {
		groups := strings.Split(*suppGroupsFromContext, ",")
		validGroups := make([]string, 0, len(groups))
		for _, group := range groups {
			group = strings.TrimSpace(group)
			if _, err := strconv.ParseInt(group, 10, 64); err == nil {
				validGroups = append(validGroups, group)
			} else {
				logging.Errorf("failed to parse supplemental group: %v", err)
			}
		}
		if len(validGroups) > 0 {
			security.SupplementalGroups = *workloadsApi.NewNullableString(convutils.Ref(strings.Join(validGroups, ",")))
		}
	}

	return &security, nil
}

type Spec interface {
	workloadsApi.SupersetSpec | api.SupersetSpec
}

type MasterSpec interface {
	workloadsApi.MasterSpec | api.MasterSpec
}

// MutateWorkload applies template and policy defaults to a workload spec.
// It merges inputSpec + templateSpec + policyDefaults and returns the mutated workload.
//
// Returns:
//   - mutatedSpec: The final mutated spec (includes user input + template + policy defaults)
//   - error: SubmissionErrorV2 if mutation fails
func MutateWorkload(
	inputSpec map[string]interface{},
	policyDefaults map[string]interface{},
	masterSameAsWorker bool,
	templateSpec map[string]interface{},
	assetsSpec map[string]interface{},
) (map[string]interface{}, *workloadsApi.SubmissionErrorV2) {

	for _, source := range []struct {
		name  string
		value *map[string]interface{}
	}{
		{"request", &inputSpec}, {"policy defaults", &policyDefaults},
		{"template", &templateSpec}, {"assets", &assetsSpec},
	} {
		normalized, err := gpucompat.NormalizeLegacyGPUFields(*source.value)
		if err != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("%s: %v", source.name, err))
		}
		*source.value = normalized
	}

	var inputSpecAsMap map[string]interface{}
	if inputSpec == nil {
		inputSpecAsMap = make(map[string]interface{})
	} else {
		inputSpecAsMap = inputSpec
	}

	// Phase 1: Assets = Spec -> Assets
	if len(assetsSpec) == 0 {
		assetsSpec = inputSpecAsMap
	} else {
		policy_compliance.PrepareDefaultsForMutation(assetsSpec)

		err := policy_compliance.ApplyDefaults(inputSpecAsMap, assetsSpec, map[policy_compliance.PerReplicaField]interface{}{})
		if err != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(err.Code, err.Message)
		}
	}

	// Phase 2: Template = (Spec -> Assets) -> Template
	if len(templateSpec) == 0 {
		templateSpec = assetsSpec
	} else {
		policy_compliance.PrepareTemplateForMutation(templateSpec)

		err := policy_compliance.ApplyDefaults(assetsSpec, templateSpec, map[policy_compliance.PerReplicaField]interface{}{})
		if err != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(err.Code, err.Message)
		}
	}

	// Phase 3: Mutated = (Spec -> Assets -> Template) -> Policy
	var mutatedSpec map[string]interface{}

	if len(policyDefaults) == 0 {
		mutatedSpec = templateSpec
	} else {
		// Use gojsonq.LoadJson (not CopySpecs) to preserve numeric types via decoder.UseNumber()
		mutatedSpec, _ = gojsonq.LoadJson(policyDefaults)

		// per product decision, in case of master same as worker - the master will also get the defaults
		// of the worker.
		if masterSameAsWorker {
			mutatedSpec[masterSection] = mutatedSpec[workerSection]
		}

		attributesDefaults := policy_compliance.PrepareDefaultsForMutation(mutatedSpec)

		err := policy_compliance.ApplyDefaults(templateSpec, mutatedSpec, attributesDefaults)
		if err != nil {
			return nil, workloadsApi.NewSubmissionErrorV2(err.Code, err.Message)
		}
	}

	cleanEmptyArraysMap(mutatedSpec)

	return mutatedSpec, nil
}

// buildResponseSpecFromMap unmarshals a map into responseSpec
func buildResponseSpecFromMap[S Spec, MS MasterSpec](workloadType string, policyDefaults map[string]interface{}) (*S, *MS, *workloadsApi.SubmissionErrorV2) {
	var responseSpec *S
	var responseMasterSpec *MS

	switch workloadType {
	case Distributed:
		if workerSpec, hasWorkerSpec := policyDefaults[workerSection]; hasWorkerSpec {
			marshalErr := json.Unmarshal(gojsonq.ToBytes(convutils.MapMe(workerSpec)), &responseSpec)
			if marshalErr != nil {
				return nil, nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to load the workload spec: %v", marshalErr))
			}
		}

		if masterSpec, hasMasterSpec := policyDefaults[masterSection]; hasMasterSpec {
			marshalErr := json.Unmarshal(gojsonq.ToBytes(convutils.MapMe(masterSpec)), &responseMasterSpec)
			if marshalErr != nil {
				return nil, nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to load the workload masterSpec: %v", marshalErr))
			}
		}
	case DistributedInference:
		if len(policyDefaults) != 0 {
			marshalErr := json.Unmarshal(gojsonq.ToBytes(policyDefaults), &responseSpec)
			if marshalErr != nil {
				return nil, nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to load the workload spec: %v", marshalErr))
			}
		}
	default:
		if len(policyDefaults) != 0 {
			marshalErr := json.Unmarshal(gojsonq.ToBytes(policyDefaults), &responseSpec)
			if marshalErr != nil {
				return nil, nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, fmt.Sprintf("failed to load the workload spec: %v", marshalErr))
			}
		}
	}

	return responseSpec, responseMasterSpec, nil
}

func NewSubmissionErrorWithDetails(code int32, message string, details *string) *workloadsApi.SubmissionErrorV2 {
	return &workloadsApi.SubmissionErrorV2{
		Code:    code,
		Message: message,
		Details: details,
	}
}

func ToSubmissionError(err *api.Error) *workloadsApi.SubmissionErrorV2 {
	result := workloadsApi.NewSubmissionErrorV2(err.Code, err.Message)
	result.SetDetails(err.Details)
	return result
}

func cleanEmptyArraysMap(aMap map[string]interface{}) {
	for key, value := range aMap {
		if arrayInMap, ok := value.([]interface{}); ok {
			if len(arrayInMap) == 0 {
				delete(aMap, key)
				continue
			}
		}

		//  handle only entries which are maps
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// recursively cleans any nested map
			cleanEmptyArraysMap(nestedMap)

			// remove nil entries
			for innerKey, innerValue := range nestedMap {
				if innerValue == nil {
					delete(nestedMap, innerKey)
				}
			}

			// if nested map is empty remove it
			if len(nestedMap) == 0 {
				delete(aMap, key)
			}
		}
	}
}

// most validations are done as part of workload composer, but there are some validations
// that we have to do before sending the specs to the composer
func preValidateSpec(ctx context.Context, params *createWorkloadParams,
	spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec,
) (map[string]k8sObjectsApi.StorageClassConfiguration, map[string]string, *workloadsApi.SubmissionErrorV2) {

	storageClassConfig, storageClassErr := preValidateStorageClasses(ctx, params.Response.ClusterId, spec, masterSpec)
	if storageClassErr != nil {
		return nil, nil, storageClassErr
	}

	if convutils.HasValue(spec.Storage) && spec.Storage.Get().Pvc != nil {
		pvcErr := validateSingleAccessModePerPvc(spec.Storage.Get().Pvc)
		if pvcErr != nil {
			return nil, nil, pvcErr
		}
	}

	claimPerDvId, dataVolumeErr := preValidateDataVolumes(ctx, params.Response.ProjectId, spec, masterSpec)
	if dataVolumeErr != nil {
		return nil, nil, dataVolumeErr
	}

	imagePullSecretErr := preValidateImagePullSecrets(ctx, params.Response.ClusterId, params.Response.ProjectId, spec, masterSpec)
	if imagePullSecretErr != nil {
		return nil, nil, imagePullSecretErr
	}

	//ServiceTypeErr := preValidateServiceType(params)
	//if ServiceTypeErr != nil {
	//	return nil, ServiceTypeErr
	//}

	if params.WorkloadType == string(api.WORKLOADTYPE_DISTRIBUTED) {
		switch spec.GetDistributedFramework() {
		case workloadsApi.DISTRIBUTEDFRAMEWORK_MPI:
			mpiError := preValidateMpiGpuFields(ctx, spec, convutils.BoolIsTrue(params.Response.MasterSpecSameAsWorker.Get()), masterSpec)
			if mpiError != nil {
				return nil, nil, mpiError
			}
		case workloadsApi.DISTRIBUTEDFRAMEWORK_JAX:
			jaxError := preValidateJaxMasterSpec(ctx, masterSpec)
			if jaxError != nil {
				return nil, nil, jaxError
			}
		}
	}

	// this is not a proper validation but a fix for a validation error that happens in the workload-controller webhook

	if spec.GetCompute().GpuMemoryPortionRequest.Get() != nil && spec.GetCompute().GpuMemoryPortionLimit.Get() != nil {
		if *spec.GetCompute().GpuMemoryPortionRequest.Get() == 1 && *spec.GetCompute().GpuMemoryPortionLimit.Get() == 1 {
			spec.Compute.Get().GpuMemoryPortionLimit.Unset()
		}
	}
	if masterSpec != nil && masterSpec.GetCompute().GpuMemoryPortionRequest.Get() != nil && masterSpec.GetCompute().GpuMemoryPortionLimit.Get() != nil {
		if *masterSpec.GetCompute().GpuMemoryPortionRequest.Get() == 1 && *masterSpec.GetCompute().GpuMemoryPortionLimit.Get() == 1 {
			masterSpec.Compute.Get().GpuMemoryPortionLimit.Unset()
		}
	}

	return storageClassConfig, claimPerDvId, nil
}

// preValidateDataVolumes validates authorized data volumes
func preValidateDataVolumes(ctx context.Context, projectId string, spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec) (map[string]string, *workloadsApi.SubmissionErrorV2) {
	if !convutils.HasValue(spec.Storage) || spec.Storage.Get().DataVolume == nil || len(spec.Storage.Get().DataVolume) == 0 {
		return nil, nil
	}

	dataVolumeIds := collectDataVolumeIds(spec.Storage.Get().DataVolume)
	if masterSpec != nil {
		dataVolumeIds = append(dataVolumeIds, collectDataVolumeIds(masterSpec.Storage.Get().DataVolume)...)
	}

	if len(dataVolumeIds) == 0 {
		return nil, nil
	}

	dataVolumeIdsSet := datastructs.NewSet(dataVolumeIds...)

	dvs, err := datavolumes.Get().GetDataVolumesNotPaginated(ctx, "UsableInProject", projectId, make([]string, 0), "", "")
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusServiceUnavailable, err.Error())
	}

	dataVolumeClaimName := make(map[string]string)

	for _, dv := range dvs {
		if len(dataVolumeIds) == 0 {
			break
		}

		if dataVolumeIdsSet.Has(dv.Id) {
			dataVolumeClaimName[dv.Id] = dv.Status.GetDatavolumePvcName()
			delete(dataVolumeIdsSet, dv.Id)
		}
	}

	if len(dataVolumeIdsSet) > 0 {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusNotFound, fmt.Sprintf("data volumes with ids: %v were not found", strings.Join(dataVolumeIdsSet.Keys(), ", ")))
	}

	return dataVolumeClaimName, nil
}

// preValidateImagePullSecrets validates that the image pull secrets are supported, valid and exist in the cluster
func preValidateImagePullSecrets(ctx context.Context, clusterId string, projectId string, spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec) *workloadsApi.SubmissionErrorV2 {
	// Collect image pull secrets from both spec and masterSpec
	workloadSecretNames := collectImagePullSecretNames(spec, masterSpec)
	if len(workloadSecretNames) == 0 {
		return nil
	}

	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	supported, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfImagePullSecrets)
	if err != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	if !supported {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "'ImagePullSecrets' field is not supported in this cluster version")
	}

	existingSecretsInCluster, k8sApiErr := k8s_objects_tracker.Get().ListSecretsByScope(ctx, "project", projectId)
	if k8sApiErr != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, k8sApiErr.Message)
	}

	existingSecretsSet := datastructs.NewSet[string]()
	for _, secret := range existingSecretsInCluster.Secrets {
		existingSecretsSet.Add(secret.Name)
	}

	for _, secretName := range workloadSecretNames {
		if !existingSecretsSet.Has(secretName) {
			return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("secret named '%s' does not exist", secretName))
		}
	}

	return nil
}

func collectDataVolumeIds(instances []workloadsApi.DataVolumeInstance) []string {
	ids := make([]string, 0)

	for _, dv := range instances {
		if dv.Id.Get() != nil {
			ids = append(ids, *dv.Id.Get())
		}
	}

	return ids
}

// collectImagePullSecretNames collects and deduplicates image pull secret names from both spec and masterSpec
func collectImagePullSecretNames(spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec) []string {
	secretNamesSet := datastructs.NewSet[string]()

	// Collect from spec
	if spec != nil && spec.ImagePullSecrets != nil {
		for _, secret := range spec.ImagePullSecrets {
			// No need to collect imagePullSecrets that are using user credential since we don't know their cluster resource name.
			if secret.Name != nil && convutils.BoolIsFalse(secret.UserCredential.Get()) {
				secretNamesSet.Add(*secret.Name)
			}
		}
	}

	// Collect from masterSpec
	if masterSpec != nil && masterSpec.ImagePullSecrets != nil {
		for _, secret := range masterSpec.ImagePullSecrets {
			if secret.Name != nil && convutils.BoolIsFalse(secret.UserCredential.Get()) {
				secretNamesSet.Add(*secret.Name)
			}
		}
	}

	return secretNamesSet.Keys()
}

// preValidateStorageClasses validates that all storage classes are allowed to use according to the storage
// configuration stored in k8s objects tracker
func preValidateStorageClasses(ctx context.Context, clusterId string,
	spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec) (map[string]k8sObjectsApi.StorageClassConfiguration, *workloadsApi.SubmissionErrorV2) {
	if !convutils.HasValue(spec.Storage) || spec.Storage.Get().Pvc == nil || len(spec.Storage.Get().Pvc) == 0 {
		return nil, nil
	}

	//
	//  collect the storage classes used by new PVCs
	//
	storageClasses := make([]string, 0)
	if convutils.HasValue(spec.Storage) {
		storageClasses = append(storageClasses, collectStorageClasses(spec.Storage.Get().Pvc)...)
	}
	if masterSpec != nil && convutils.HasValue(masterSpec.Storage) {
		storageClasses = append(storageClasses, collectStorageClasses(masterSpec.Storage.Get().Pvc)...)
	}

	if len(storageClasses) == 0 {
		return nil, nil
	}

	storageClassesUnique := datastructs.NewSet(storageClasses...).Keys()

	//
	//  need to validate that the storage classes are allowed
	//
	configs, err := k8s_objects_tracker.Get().GetStorageClassesConfigInt(ctx, clusterId, storageClassesUnique)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(err.Code, err.Message)
	}

	for _, pvc := range spec.Storage.Get().Pvc {
		if !convutils.HasValue(pvc.ClaimInfo) || !convutils.HasValue(pvc.ClaimInfo.Get().StorageClass) {
			continue
		}

		storageClassName := *pvc.ClaimInfo.Get().StorageClass.Get()
		storageClassNameForMessage := fmt.Sprintf("storage class %s", storageClassName)

		if storageClassName == "" {
			storageClassName = NoneStorageClass
			storageClassNameForMessage = "static provisioning"
		}

		config, configDefined := configs[storageClassName]

		// if no permissions defined for this storage class, the assumption is not restricted
		if !configDefined || !convutils.HasValue(config.Permissions) {
			continue
		}

		// in existing pvcs we typically won't get the storage class name, but if the user did specify it - we can
		// at least verify that running workloads on this storage class is allowed
		if convutils.BoolIsTrue(pvc.ExistingPvc.Get()) {
			if !k8sObjectsClient.IsStorageClassActionPermitted(config.Permissions.Get().AllowedForWorkloads) {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("%s is forbidden for running workloads", storageClassNameForMessage))
			}
			continue
		}

		// non-existing pvc instances are treated as volumes. we distinguish ephemeral from persistent volumes
		if convutils.BoolIsTrue(pvc.Ephemeral.Get()) {
			if !k8sObjectsClient.IsStorageClassActionPermitted(config.Permissions.Get().AllowedForEphemeralVolumes) {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("creating ephemeral volumes on %s is forbidden", storageClassNameForMessage))
			}
		} else {
			if !k8sObjectsClient.IsStorageClassActionPermitted(config.Permissions.Get().AllowedForPersistentVolumes) {
				return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("creating persistent volumes on %s is forbidden", storageClassNameForMessage))
			}
		}
	}

	return configs, nil
}

// validateSingleAccessModePerPvc ensures no PVC is mounted with mixed read-only and read-write modes
func validateSingleAccessModePerPvc(pvcs []workloadsApi.PvcInstance) *workloadsApi.SubmissionErrorV2 {
	seen := map[string]bool{}

	for _, pvc := range pvcs {
		if !convutils.HasValue(pvc.ClaimName) {
			continue
		}
		name := *pvc.ClaimName.Get()
		readOnly := convutils.Deref(pvc.ReadOnly.Get(), false)

		if prev, ok := seen[name]; ok && prev != readOnly {
			return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("PVC %s cannot be mounted with both readOnly: true and readOnly: false in the same pod", name))
		}
		seen[name] = readOnly
	}
	return nil
}

func collectStorageClasses(pvcs []workloadsApi.PvcInstance) []string {
	result := make([]string, 0)

	if len(pvcs) == 0 {
		return result
	}

	for _, pvc := range pvcs {
		if convutils.HasValue(pvc.ClaimInfo) {
			if convutils.IsEmptyOrNull(pvc.ClaimInfo.Get().StorageClass.Get()) {
				result = append(result, NoneStorageClass)
			} else {
				result = append(result, *pvc.ClaimInfo.Get().StorageClass.Get())
			}
		}
	}

	return result
}

// preValidateMpiGpuFields validates the GPU fields in case MPI framework is used
// If master-gpu field was specified and is a value other than 0 - return an error.
// If master-gpu field wasn't specified, BUT worker-gpu was specified AND masterSameAsWorker is true, master-gpu should be
// set to 0 (even if the worker-gpu has a different value)
// the function assumes that distributed=true and framework=MPI validation is already done prior to calling this function
func preValidateMpiGpuFields(ctx context.Context, spec *workloadsApi.SupersetSpec, masterSameAsWorker bool, masterSpec *workloadsApi.MasterSpec) *workloadsApi.SubmissionErrorV2 {
	if masterSpec == nil || !convutils.HasValue(masterSpec.Compute) {
		return nil
	}

	computeFields := masterSpec.Compute.Get()

	var err *workloadsApi.SubmissionErrorV2

	if !masterSameAsWorker {
		validateForbiddenFieldForMpi(&err, fldNameGpuDevicesRequest, computeFields.GpuDevicesRequest.Get())
		validateForbiddenFieldForMpi(&err, fldNameGpuRequestType, computeFields.GpuRequestType.Get())
		validateForbiddenFieldForMpi(&err, fldNameGpuMemoryPortionRequest, computeFields.GpuMemoryPortionRequest.Get())
		validateForbiddenFieldForMpi(&err, fldNameGpuMemoryPortionLimit, computeFields.GpuMemoryPortionLimit.Get())
		validateForbiddenFieldForMpi(&err, fldNameGpuMemoryRequest, computeFields.GpuMemoryRequest.Get())
		validateForbiddenFieldForMpi(&err, fldNameGpuMemoryLimit, computeFields.GpuMemoryLimit.Get())
	} else {
		computeFields.GpuDevicesRequest = workloadsApi.NullableInt32{}
		computeFields.GpuRequestType = workloadsApi.NullableGpuRequestType{}
		computeFields.GpuMemoryPortionRequest = workloadsApi.NullableFloat64{}
		computeFields.GpuMemoryPortionLimit = workloadsApi.NullableFloat64{}
		computeFields.GpuMemoryRequest = workloadsApi.NullableString{}
		computeFields.GpuMemoryLimit = workloadsApi.NullableString{}
	}

	return err
}

func preValidateJaxMasterSpec(_ context.Context, masterSpec *workloadsApi.MasterSpec) *workloadsApi.SubmissionErrorV2 {
	if masterSpec != nil {
		return workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, "JAX submission cannot have master spec")
	}
	return nil
}

func validateForbiddenFieldForMpi[T comparable](err **workloadsApi.SubmissionErrorV2, name string, field *T) {
	//
	//  already found error in the set of fields we're checking
	//
	if *err != nil {
		return
	}

	var zero T

	if field != nil && (*field) != zero {
		*err = workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("%s should not be provided for master replica of MPI workload", name))
	}
}

// setDefaultNodepoolsIfNecessary populates the NodePools field with the default nodePools if none were provided.
// default nodePools are retrieved from the project if exist, or set to the "default" nodePool
func setDefaultNodepoolsIfNecessary(ctx context.Context, response *workloadsApi.SupersetWorkload, project *orgUnitApi.Project) {
	if len(response.Spec.NodePools) == 0 {
		defaultNodepools := getDefaultNodepools(project)

		response.Spec.NodePools = defaultNodepools

		if response.MasterSpec.IsSet() {
			response.MasterSpec.Get().NodePools = defaultNodepools
		}

		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("set default nodepools for workload %s. value: %v", response.Name, defaultNodepools)
	}
}

func getDefaultNodepools(project *orgUnitApi.Project) []string {
	if project != nil && len(project.DefaultNodePools) > 0 {
		return project.DefaultNodePools
	}

	return []string{runv1.DefaultNodePoolName}
}

// extractUserCredentialsFromSpec returns a deduplicated list of user credentials for spec and masterSpec.
func extractUserCredentialsFromSpec(spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec) ([]string, error) {
	// Helper function to collect credential names from environment variables
	checkEnvVars := func(envVars []workloadsApi.EnvironmentVariable) ([]string, error) {
		var names []string
		for _, env := range envVars {
			if env.UserCredential.IsSet() {
				if !env.UserCredential.Get().Name.IsSet() {
					return nil, fmt.Errorf("credential name is required for environment variable %s", *env.Name.Get())
				}
				names = append(names, *env.UserCredential.Get().Name.Get())
			}
		}
		return names, nil
	}

	checkImagePullSecrets := func(imagePullSecrets []workloadsApi.ImagePullSecret) []string {
		var names []string
		for _, secret := range imagePullSecrets {
			if secret.Name != nil && convutils.BoolIsTrue(secret.UserCredential.Get()) {
				names = append(names, *secret.Name)
			}
		}

		return names
	}

	// Collect credential names from both worker and master specs
	var names []string
	if spec != nil && (spec.EnvironmentVariables != nil || spec.ImagePullSecrets != nil) {
		specCreds, err := checkEnvVars(spec.EnvironmentVariables)
		if spec.ImagePullSecrets != nil {
			specCreds = append(specCreds, checkImagePullSecrets(spec.ImagePullSecrets)...)
		}
		if err != nil {
			return nil, err
		}
		names = append(names, specCreds...)
	}
	if masterSpec != nil && (masterSpec.EnvironmentVariables != nil || masterSpec.ImagePullSecrets != nil) {
		masterCreds, err := checkEnvVars(masterSpec.EnvironmentVariables)
		if masterSpec.ImagePullSecrets != nil {
			masterCreds = append(masterCreds, checkImagePullSecrets(masterSpec.ImagePullSecrets)...)
		}
		if err != nil {
			return nil, err
		}
		names = append(names, masterCreds...)
	}

	if names == nil {
		return nil, nil
	}

	// Deduplicate credential names
	dedupedNames := datastructs.Set[string]{}
	dedupedNames.Add(names...)
	return dedupedNames.Keys(), nil
}

func processCredentials(ctx context.Context, clusterId string, spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec) ([]workloadsApi.AssetsMapping, *workloadsApi.SubmissionErrorV2) {
	credentials, err := extractUserCredentialsFromSpec(spec, masterSpec)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, err.Error())
	}

	// If no credentials are used, return early
	if len(credentials) == 0 {
		return nil, nil
	}

	// Check if cluster supports credential environment variables
	clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	supported, err := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionOfUserCredentials)
	if err != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, err.Error())
	}
	if !supported {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusBadRequest, fmt.Sprintf("Cluster %s does not support user credentials in environment variables", clusterId))
	}

	// requesting only private credential
	credentialsAssets, clientErr := assetsClient.Get().ListCredentialsAssets(ctx, true)
	if clientErr != nil {
		return nil, workloadsApi.NewSubmissionErrorV2(http.StatusInternalServerError, clientErr.Message)
	}

	assetsMapping := make([]workloadsApi.AssetsMapping, 0)
	for _, credName := range credentials {
		assetMeta := func(cred string) *assetsApi.AssetMeta {
			for _, asset := range credentialsAssets.Entries {
				if asset.Meta.Name == cred {
					return &asset.Meta
				}
			}
			return nil
		}(credName)
		if assetMeta == nil {
			return nil, workloadsApi.NewSubmissionErrorV2(http.StatusNotFound, fmt.Sprintf("credential with name '%s' not found", credName))
		}
		assetsMapping = append(assetsMapping, workloadsApi.AssetsMapping{
			Id:   convutils.Ref(assetMeta.Id),
			Name: convutils.Ref(credName),
			Kind: convutils.Ref(workloadsApi.AssetKind(assetMeta.Kind)),
		})
	}

	return assetsMapping, nil
}

func determineComposerOptions(ctx context.Context, ClusterCapabilitiesChecker *checker.ClusterCapabilitiesChecker) (composer.FieldsComposerOptions, *workloadsApi.SubmissionErrorV2) {
	options := composer.FieldsComposerOptions{
		UserCredentialsSupported: true,
	}

	checkerErr := composer_v2.SetClusterDependentOptions(ctx, &options, ClusterCapabilitiesChecker)
	if checkerErr != nil {
		return options, NewSubmissionErrorWithDetails(http.StatusServiceUnavailable, checkerErr.Error(), nil)
	}

	return options, nil
}
