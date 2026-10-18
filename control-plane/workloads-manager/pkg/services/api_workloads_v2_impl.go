/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/plugins"

	"github.com/google/uuid"

	"github.com/cenkalti/backoff/v4"
	ourConfig "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/config"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	cache "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cacheopts "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/karta/pkg/jq/execution"
	kartaresource "github.com/run-ai/karta/pkg/resource"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/middleware"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/apimachinery/pkg/util/validation"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authApi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	workloadsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/settings"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

const (
	MaxWorkloadNameLength        = 50
	MaxGeneratedNameSuffixLength = 12

	WorkloadNameFilter    = "name"
	WorkloadProjectFilter = "projectId"
	WorkloadClusterFilter = "clusterId"

	PreciseName                 = true
	NotToIncludeDeletedWorkload = false

	NonPreemptibleMinLevel = 100

	PolicyMaxErrors         = 5
	PolicyMaxErrorsInUpdate = 1
)

var permissionError = api.NewError(http.StatusForbidden, "insufficient permissions", nil)

// componentDefCacheEntry wraps a ComponentDefinition pointer to distinguish a cache miss (nil)
// from a cached "type has no component definition" result (&{nil}).
type componentDefCacheEntry struct {
	Def *kartav1alpha1.ComponentDefinition
}

func RegisterSuspendDefCacheStore() {
	servicesConfig := config.Get[*ourConfig.Services]()
	if err := cache.RegisterStore[componentDefCacheEntry](cacheopts.Ttl(time.Duration(servicesConfig.SuspendDefCacheTTLSec) * time.Second)); err != nil {
		logging.Warnf("Failed to register suspend definition cache store: %v", err)
	}
}

type WorkloadsV2Impl struct {
	k8sUtil          *K8sResourceUtil
	settingsManager  *settings.SettingsManager
	gvkPluginFactory *plugins.Factory
}

type CreateWorkloadV2Options struct {
	DryRun                  bool
	CheckYamlSubmissionFlag bool
	TranslatedFromV1        bool
}

func NewWorkloadsV2Impl() *WorkloadsV2Impl {
	return &WorkloadsV2Impl{
		k8sUtil:          NewK8sResourceUtil(),
		settingsManager:  settings.NewSettingsManager(),
		gvkPluginFactory: plugins.NewFactory(),
	}
}

func (w *WorkloadsV2Impl) CreateWorkloadV2(ctx context.Context, request api.WorkloadV2CreateRequest, checkYamlSubmissionFlag bool) (*api.WorkloadV2, []api.SubmissionErrorV2) {
	return w.CreateWorkloadV2WithDryRun(ctx, request, CreateWorkloadV2Options{DryRun: false, CheckYamlSubmissionFlag: checkYamlSubmissionFlag})
}

func (w *WorkloadsV2Impl) DryRunCreateWorkloadV2(ctx context.Context, request api.WorkloadV2CreateRequest) (*api.WorkloadV2, []api.SubmissionErrorV2) {
	return w.CreateWorkloadV2WithDryRun(ctx, request, CreateWorkloadV2Options{DryRun: true, CheckYamlSubmissionFlag: true})
}

func (w *WorkloadsV2Impl) CreateWorkloadV2WithDryRun(ctx context.Context, request api.WorkloadV2CreateRequest, options CreateWorkloadV2Options) (*api.WorkloadV2, []api.SubmissionErrorV2) {
	if options.CheckYamlSubmissionFlag {
		yamlSubmissionFlag, err := w.settingsManager.GetSettingByKey(api.WORKLOADSSETTINGSKEYENUM_YAML_SUBMISSION, strconv.Itoa(int(auth.TenantIdOf(ctx))))
		if err != nil {
			return nil, SingleError(NewSubmissionError(http.StatusInternalServerError, "failed to get yaml submission flag", err))
		}

		if !yamlSubmissionFlag.(*settings.YamlSubmission).Enabled {
			return nil, SingleError(NewSubmissionError(http.StatusForbidden, "yaml submission is not enabled. Please contact the administrator to enable it.", nil))
		}
	}

	validationErrors := NewValidationErrors()

	subjectId, apiErr := runaictx.ParseSubjectId(ctx)
	if apiErr != nil {
		validationErrors.Add(ToSubmissionErrorV2(api.NewValidationError(apiErr.Error())))
	}

	// Validate manifest first - if invalid, we can't proceed with other validations
	manifest, manifestErr := w.validateManifest(ctx, request.Manifest, request.ManifestBase64Encoded)
	if manifestErr != nil {
		validationErrors.Add(NewManifestError(manifestErr.Message))
	}

	// Strip the converted-from-native label on public requests to prevent clients from forging it
	if !options.TranslatedFromV1 && manifest != nil {
		stripConvertedFromNativeLabel(manifest)
	}

	// Validate projectId/clusterId - if invalid, we can't proceed with project-dependent validations
	projectIdOrClusterIdErr := validateProjectIdOrClusterId(request.Metadata.ProjectId, request.Metadata.ClusterId)
	if projectIdOrClusterIdErr != nil {
		validationErrors.Add(projectIdOrClusterIdErr)
	}

	// If we can't proceed without manifest or project info, return early with collected errors
	if manifest == nil || projectIdOrClusterIdErr != nil {
		return nil, validationErrors.ToSubmissionErrors()
	}

	// Resolve project - if we can't get the project, we can't continue with cluster-dependent validations
	// Skip auth validation for internal requests (TranslatedFromV1 implies trusted internal caller)
	project, ancestors, resolveErr := w.resolveProject(ctx, &request.Metadata, manifest, options.TranslatedFromV1)
	if resolveErr != nil {
		validationErrors.Add(resolveErr)
		return nil, validationErrors.ToSubmissionErrors()
	}

	// From here on, we can validate independently and collect all errors

	manifestResourceInfo, err := w.k8sUtil.GetResourceInfo(manifest)
	if err != nil {
		validationErrors.Add(NewManifestError(fmt.Sprintf("failed to extract resource info from manifest: %v", err)))
		return nil, validationErrors.ToSubmissionErrors()
	}
	trustedConvertedFromNative := options.TranslatedFromV1 && isConvertedFromNativeWorkload(manifestResourceInfo.Labels)

	// Validate GVK/workload type
	var workloadType *workloadsApi.WorkloadTypeConfig
	workloadType, gvkErr := w.getWorkloadTypeFromGVK(ctx, manifestResourceInfo, project.ClusterId)
	if gvkErr != nil {
		validationErrors.Add(gvkErr)
	}

	// Validate category and priority
	priorityNameToValue, catPrioErr := validateCategoryAndPriority(ctx, project.ClusterId, &request.Metadata)
	if catPrioErr != nil {
		validationErrors.Add(catPrioErr)
	}

	// Validate cluster supports v2 workloads
	clusterErr := validateClusterSupportsV2Workloads(ctx, project.ClusterId)
	validationErrors.Add(clusterErr)
	if clusterErr == nil && !trustedConvertedFromNative {
		validationErrors.Add(w.validateSegmentTopologyClusterSupport(ctx, project.ClusterId, manifest))
	}

	// Get and validate workload name
	name := w.GetWorkloadName(ctx, &request.Metadata, manifest)
	if nameErr := w.ValidateWorkloadName(ctx, name, request.Metadata.ProjectId, ancestors.ClusterUuid.Get()); nameErr != nil {
		validationErrors.Add(nameErr)
	}

	// If there are any validation errors, return them all
	if validationErrors.HasErrors() {
		return nil, validationErrors.ToSubmissionErrors()
	}

	// All validations passed - compute effective metadata
	mutatedMetadata := w.applyDefaultMutations(request.Metadata, workloadType, priorityNameToValue)
	mutatedMetadata.Name = name

	networkingProfile := networkingSubmissionProfileForRecord(&models.WorkloadsQueueRecord{
		TranslatedFromV1: options.TranslatedFromV1,
		InputRequestPath: middleware.GetRequestPath(ctx),
	})
	configuration, configErr := w.buildConfiguration(ctx, request.Metadata.Configuration, project, workloadType, manifestResourceInfo, mutatedMetadata.Name, networkingProfile)
	if configErr != nil {
		validationErrors.Add(configErr)
		return nil, validationErrors.ToSubmissionErrors()
	}

	// For dryRun, return effective metadata without creating workload
	if options.DryRun {
		validationErrors.Add(w.validateSegmentTopologyForDryRun(ctx, manifest, project, trustedConvertedFromNative))
		dryRunRecord := &models.WorkloadsQueueRecord{
			ID:                "dry-run",
			ResourceVersion:   manifestResourceInfo.Version,
			EffectiveManifest: w.k8sUtil.deepCopyMap(manifest),
			Configuration:     configuration,
		}
		enableNetworkingOwnershipContract(&dryRunRecord.Configuration)
		if bindingErr := applyNetworkingOwnershipBinding(ctx, dryRunRecord, workloadType); bindingErr != nil {
			validationErrors.Add(NewConfigurationError(fmt.Sprintf("failed to apply networking ownership binding: %v", bindingErr)))
		}
		if validationErrors.HasErrors() {
			return nil, validationErrors.ToSubmissionErrors()
		}
		return &api.WorkloadV2{
			Metadata: api.WorkloadV2Metadata{
				Name:           mutatedMetadata.Name,
				ProjectId:      mutatedMetadata.ProjectId,
				Priority:       mutatedMetadata.Priority,
				Category:       mutatedMetadata.Category,
				Preemptibility: mutatedMetadata.Preemptibility,
			},
		}, nil
	}

	// Proceed with creation
	workloadId := uuid.New().String()
	effectiveManifest, workloadId, editErr := w.editManifestWithMetadata(
		ctx,
		manifest,
		mutatedMetadata,
		project,
		workloadId,
		trustedConvertedFromNative,
		nodePoolsKarta(ctx, workloadType, manifestResourceInfo.Version),
	)
	if editErr != nil {
		return nil, SingleError(NewManifestError(fmt.Sprintf("failed to edit manifest: %v", editErr)))
	}

	resourceInfo, resourceErr := w.k8sUtil.GetResourceInfo(effectiveManifest)
	if resourceErr != nil {
		return nil, SingleError(NewManifestError(fmt.Sprintf("failed to extract resource info: %v", resourceErr)))
	}

	// Get request path and body from context
	requestPath := middleware.GetRequestPath(ctx)
	requestBody := middleware.GetRequestBody(ctx)

	// Convert request body to JsonBMap if present
	var requestBodyMap dbmodels.JsonBMap
	if len(requestBody) > 0 {
		// JsonBMap is a wrapper for map[string]interface{}, so we need to unmarshal the JSON
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(requestBody, &bodyMap); err != nil {
			// Log error but don't fail the request
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to unmarshal request body: %v", err)
		} else {
			requestBodyMap = bodyMap
		}
	}
	if requestPath == NimServicesAPIPath {
		stored, err := gpucompat.LegacyGPUFields(requestBodyMap)
		if err != nil {
			return nil, SingleError(NewSubmissionError(http.StatusBadRequest, err.Error(), err))
		}
		requestBodyMap = stored
	}

	qRecord := models.WorkloadsQueueRecord{
		ID: workloadId,

		Name: name,

		CreatedAt: time.Now().Truncate(time.Microsecond),
		CreatedBy: subjectId,

		TenantId:     auth.TenantIdOf(ctx),
		ClusterId:    project.ClusterId,
		DepartmentId: project.ParentId,
		ProjectId:    project.Id,

		Category:      mutatedMetadata.Category,
		PriorityClass: mutatedMetadata.Priority,

		ResourceGroup:   resourceInfo.Group,
		ResourceVersion: resourceInfo.Version,
		ResourceKind:    resourceInfo.Kind,

		RawManifest:       manifest,
		EffectiveManifest: effectiveManifest,

		DesiredPhase: string(api.DESIREDPHASE_RUNNING),

		TranslatedFromV1: options.TranslatedFromV1 && isConvertedFromNativeWorkload(manifestResourceInfo.Labels),

		InputRequestPath: requestPath,
		InputRequestBody: requestBodyMap,
	}

	if mutatedMetadata.Preemptibility != nil {
		qRecord.Preemptibility = convutils.Ref(string(*mutatedMetadata.Preemptibility))
	}

	qRecord.UpdatedBy = qRecord.CreatedBy
	qRecord.UpdatedAt = qRecord.CreatedAt

	qRecord.ProjectName = project.Name

	if project.Parent.IsSet() && project.Parent.Get() != nil {
		qRecord.DepartmentName = project.Parent.Get().Name
	}

	if ancestors.DepartmentId.IsSet() {
		qRecord.DepartmentId = convutils.Deref(ancestors.DepartmentId.Get(), "")
	}

	/*WAIT_FOR_OFER need to contact workloads for this validation
	customPriorityErr := validators.ValidateCustomPriorityClass(ctx, project.ClusterId, request.Metadata.Priority)
	if customPriorityErr != nil {
		return nil, customPriorityErr
	}

	categoryValidationErr := validators.ValidateWorkloadCategory(ctx, request.Metadata.Category)
	if categoryValidationErr != nil {
		return nil, categoryValidationErr
	}
	*/

	if convutils.IsEmptyOrNull(mutatedMetadata.Category) {
		qRecord.Category = convutils.Ref(DefaultCategory)
	}

	qRecord.Configuration = configuration

	polErrs := ApplyPolicy(ctx, &qRecord, PolicyMaxErrors)
	if len(polErrs) > 0 {
		submissionErrs := make([]api.SubmissionErrorV2, 0, len(polErrs))
		for _, e := range polErrs {
			submissionErrs = append(submissionErrs, *ToSubmissionErrorV2(e))
		}
		return nil, submissionErrs
	}

	// The contract is private server state. Setting it only after policy
	// mutation prevents request or policy input from selecting the protected
	// workload identity value.
	enableNetworkingOwnershipContract(&qRecord.Configuration)
	if bindingErr := applyNetworkingOwnershipBinding(ctx, &qRecord, workloadType); bindingErr != nil {
		return nil, SingleError(NewConfigurationError(fmt.Sprintf("failed to apply networking ownership binding: %v", bindingErr)))
	}

	notificationCtx := context.WithoutCancel(ctx)
	go notifyWorkloadCreated(notificationCtx, &qRecord)

	dbErr := dao.DbUpsertWorkload(ctx, &qRecord)
	if dbErr != nil {
		return nil, SingleError(ToSubmissionErrorV2(dbErr))
	}

	responsePayload := qRecord.ToWorkloadV2()
	if !options.TranslatedFromV1 {
		redactWorkloadV2SensitiveValues(&responsePayload)
	}

	return &responsePayload, nil
}

// buildConfiguration computes the persisted workload configuration from the request's configuration
// params. Shared by create and update so both paths apply MNNVL determination, networking validation,
// user credentials, and PVCs (cluster-capability gated) identically.
func (w *WorkloadsV2Impl) buildConfiguration(ctx context.Context, configParams api.WorkloadV2Configuration,
	project *orgUnitApi.Project, workloadType *workloadsApi.WorkloadTypeConfig, manifestResourceInfo *ResourceInfo,
	workloadName string, networkingProfile api.NetworkingSubmissionProfile,
) (*models.WorkloadConfiguration, *api.SubmissionErrorV2) {
	configuration, mnnvlErr := DetermineMNNVLConfiguration(ctx, configParams.Mnnvl, project, project.ClusterId, workloadType, manifestResourceInfo)
	if mnnvlErr != nil {
		return nil, mnnvlErr
	}

	if configParams.Networking != nil {
		if networkingErr := validateWorkloadV2NetworkingForCluster(ctx, project.ClusterId, workloadName, configParams.Networking, networkingProfile); networkingErr != nil {
			return nil, networkingErr
		}
		configuration.Networking = cloneNetworkingConfig(configParams.Networking)
	}

	if pc := configParams.UserCredentials; pc != nil && len(pc.Sources) > 0 {
		sources := make([]string, 0, len(pc.Sources))
		for _, s := range pc.Sources {
			sources = append(sources, s.Ref)
		}
		configuration.UserCredentials = &models.UserCredentialsConfiguration{
			Sources: sources,
		}
	}

	if pvcs := configParams.Pvcs; pvcs != nil {
		capChecker := checker.NewClusterCapabilitiesChecker(project.ClusterId)
		ok, err := capChecker.ValidateClusterCapability(ctx, checker.MinVersionWorkloadV2Pvcs)
		if err != nil {
			return nil, NewManifestError(fmt.Sprintf("cluster capability check: %v", err))
		}
		if !ok {
			errMsg := fmt.Sprintf("metadata.configuration.pvcs is supported for cluster version %s and above", checker.MinVersionForSupporting(checker.MinVersionWorkloadV2Pvcs))
			return nil, NewManifestError(errMsg)
		}
		configuration.Pvcs = *pvcs
		configuration.PvcsConfig = configParams.PvcsConfig
	}

	return configuration, nil
}

// validateProjectIdOrClusterId validates that either projectId or clusterId is provided, but not both
func validateProjectIdOrClusterId(projectId, clusterId string) *api.SubmissionErrorV2 {
	if projectId == "" && clusterId == "" {
		return NewProjectError("either projectId or clusterId must be provided")
	}
	if projectId != "" && clusterId != "" {
		return NewProjectError("projectId and clusterId are mutually exclusive, provide only one")
	}
	return nil
}

// resolveProject resolves the project from the metadata and optionally validates auth.
// It returns the project after updating metadata.ProjectId with the resolved value.
// When skipAuth is true (internal trusted requests), authorization validation is skipped.
func (w *WorkloadsV2Impl) resolveProject(ctx context.Context, metadata *api.WorkloadV2MetadataCreateParams,
	manifest map[string]interface{}, skipAuth bool,
) (*orgUnitApi.Project, *orgUnitApi.Ancestors, *api.SubmissionErrorV2) {
	tenantId, apiErr := runaictx.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, nil, NewProjectError(apiErr.Error())
	}

	if err := validateProjectIdOrClusterId(metadata.ProjectId, metadata.ClusterId); err != nil {
		return nil, nil, err
	}

	projectId := metadata.ProjectId

	if metadata.ClusterId != "" {
		projectMetadata, orgUnitErr := w.inferProjectFromClusterAndNamespace(ctx, metadata.ClusterId, manifest)
		if orgUnitErr != nil {
			return nil, nil, NewProjectErrorWithCode(int(orgUnitErr.Code), orgUnitErr.Message)
		}
		projectId = projectMetadata.Id
	}

	project, orgUnitErr := orgUnitClient.Get().GetProjectById(ctx, projectId)
	if orgUnitErr != nil {
		return nil, nil, NewProjectErrorWithCode(int(orgUnitErr.Code), orgUnitErr.Message)
	}

	ancestors := &orgUnitApi.Ancestors{
		TenantId:     tenantId,
		ClusterUuid:  *orgUnitApi.NewNullableString(&project.ClusterId),
		DepartmentId: *orgUnitApi.NewNullableString(&project.ParentId),
		ProjectId:    *orgUnitApi.NewNullableString(&project.Id),
	}

	if !skipAuth {
		_, authErr := validateCreationAuth(ctx, ancestors, projectId)
		if authErr != nil {
			return nil, nil, ToSubmissionErrorV2(authErr)
		}
	}

	// Update metadata with resolved projectId
	metadata.ProjectId = projectId

	return project, ancestors, nil
}

// ValidateWorkloadNameFormat validates only the name format without checking uniqueness
func (w *WorkloadsV2Impl) ValidateWorkloadNameFormat(name string) *api.SubmissionErrorV2 {
	if name == "" {
		return NewNameError("Workload name must not be empty")
	}
	if len(name) > MaxWorkloadNameLength {
		return NewNameError(fmt.Sprintf("The length of a workload name must not exceed %d characters", MaxWorkloadNameLength))
	}

	namingErrs := validation.IsDNS1035Label(name)
	if len(namingErrs) > 0 {
		return NewNameError("The workload name must consist of lower case alphanumeric characters or '-', and start with an alphabetic character.")
	}

	return nil
}

func (w *WorkloadsV2Impl) ValidateWorkloadName(ctx context.Context, name string, projectId string, clusterId *string) *api.SubmissionErrorV2 {
	if name == "" {
		return NewNameError("Workload name must not be empty")
	}
	if len(name) > MaxWorkloadNameLength {
		return NewNameError(fmt.Sprintf("The length of a workload name must not exceed %d characters", MaxWorkloadNameLength))
	}

	namingErrs := validation.IsDNS1035Label(name)
	if len(namingErrs) > 0 {
		return NewNameError("The workload name must consist of lower case alphanumeric characters or '-', and start with an alphabetic character.")
	}

	filters := client_utils.NewFilterBuilder().
		AddFilter(WorkloadNameFilter, client_utils.EqualOp, name).
		AddFilter(WorkloadProjectFilter, client_utils.EqualOp, projectId)

	if clusterId != nil {
		filters.AddFilter(WorkloadClusterFilter, client_utils.EqualOp, *clusterId)
	}

	count, workloadsErr := workloadsClient.Get().CountWorkloads(ctx, filters.Get(), "", NotToIncludeDeletedWorkload, PreciseName)

	// failure to validate the name should not fail the submission. we want the workload to be
	// created and in worst case it will fail an additional validation that we have in the cluster
	if workloadsErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("count workloads failed, workload %s may be duplicate in project %s: %v",
			name, projectId, workloadsErr)
		return nil
	}

	if count != 0 {
		return NewSubmissionErrorWithField(http.StatusConflict, fmt.Sprintf("workload with name '%s' already exists in this project", name), ErrorFieldName)
	}

	return nil
}

// generateNameSuffix generates a suffix for workload names
func (w *WorkloadsV2Impl) generateNameSuffix() string {
	newUuid := strings.ReplaceAll(uuid.NewString(), "-", "")
	return newUuid[len(newUuid)-MaxGeneratedNameSuffixLength:]
}

// generateNameByPrefix returns a name that starts with the given prefix and ends with a random suffix,
// which is based on the last part of a v4 uuid, separated with a '-'
func (w *WorkloadsV2Impl) generateNameByPrefix(prefix string) string {
	suffix := w.generateNameSuffix()
	return fmt.Sprintf("%s-%s", prefix, suffix)
}

// GetWorkloadName determines the workload name based on API metadata and YAML manifest
func (w *WorkloadsV2Impl) GetWorkloadName(ctx context.Context, metadata *api.WorkloadV2MetadataCreateParams, manifest map[string]interface{}) string {
	// API overrides YAML - check API metadata first
	if metadata.Name != "" {
		// If UseGivenNameAsPrefix is true, generate the name using the prefix
		if metadata.UseGivenNameAsPrefix {
			generatedName := w.generateNameByPrefix(metadata.Name)
			return generatedName
		}

		return metadata.Name
	}

	if manifest == nil {
		// No manifest, return empty name
		// will fail later on the name validation as the workload is invalid
		return ""
	}

	// API doesn't provide a name, check YAML manifest
	manifestMetadata, ok := manifest[constants.FieldMetadata].(map[string]interface{})
	if !ok {
		// No metadata in manifest, return empty name
		// will fail later on the name validation as the workload is invalid
		return ""
	}

	// Check if YAML has a name field
	if yamlName, exists := manifestMetadata[constants.FieldName].(string); exists && yamlName != "" {
		// YAML provides a name
		return yamlName
	}

	// Check if YAML has a generateName field (used as prefix)
	if yamlGenerateName, exists := manifestMetadata[constants.FieldGenerateName].(string); exists && yamlGenerateName != "" {
		// YAML provides a generateName prefix
		generatedName := w.generateNameByPrefix(yamlGenerateName)
		return generatedName
	}

	// No name or prefix found in either API or YAML
	// will fail later on the name validation as the workload is invalid
	return ""
}

func (w *WorkloadsV2Impl) GetWorkloadV2ById(ctx context.Context, workloadId string) (*api.WorkloadV2GetResponse, *api.Error) {
	permittedScopes, scopesErr := authClient.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	qRecord, dbErr := dao.DbGetWorkload(ctx, workloadId, permittedScopes)
	if dbErr != nil {
		return nil, dbErr
	}

	response := qRecord.ToGetResponse()
	redactWorkloadV2GetResponseSensitiveValues(&response)

	statusRecord, dbErr := dao.DbGetSyncStatus(ctx, workloadId)
	if dbErr != nil {
		if dbErr.Code != http.StatusNotFound {
			return nil, dbErr
		}

		response.SyncInfo = api.WorkloadV2SyncInfo{
			Status: api.WORKLOADV2SYNCSTATUS_QUEUED,
		}
	} else {
		response.SyncInfo = statusRecord.ToSyncInfo(qRecord.UpdateCount)
	}

	return &response, nil
}

type RecordModifier func(record *models.WorkloadsQueueRecord)

func (w *WorkloadsV2Impl) UpdateWorkloadV2(ctx context.Context, workloadId string, request api.WorkloadV2UpdateRequest, modifiers ...RecordModifier) (*api.WorkloadV2, *api.Error) {
	return w.updateWorkloadV2(ctx, workloadId, request, false, modifiers...)
}

func (w *WorkloadsV2Impl) UpdateWorkloadInternal(ctx context.Context, workloadId string, request api.WorkloadV2UpdateRequest) (*api.WorkloadV2, *api.Error) {
	// The originating typed-V1 API already authorized the request. This trusted path mirrors
	// CreateWorkloadInternal while retaining the provenance, GVK, and immutability checks below.
	return w.updateWorkloadV2(ctx, workloadId, request, true)
}

func (w *WorkloadsV2Impl) updateWorkloadV2(ctx context.Context, workloadId string, request api.WorkloadV2UpdateRequest, skipAuth bool, modifiers ...RecordModifier) (*api.WorkloadV2, *api.Error) {
	// Validate manifest
	manifest, err := w.validateManifest(ctx, request.Manifest, request.ManifestBase64Encoded)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("manifest validation failed: %v", err), nil)
	}

	tenantId, apiErr := runaictx.ParseTenantId(ctx)
	if apiErr != nil {
		return nil, api.NewValidationError(apiErr.Error())
	}

	subjectId, apiErr := runaictx.ParseSubjectId(ctx)
	if apiErr != nil {
		return nil, api.NewValidationError(apiErr.Error())
	}

	permittedScopes, scopesErr := authClient.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	// Get existing workload
	qRecord, dbErr := dao.DbGetWorkload(ctx, workloadId, permittedScopes)
	if dbErr != nil {
		return nil, dbErr
	}
	if restoreErr := restoreRedactedSensitiveManifestValues(manifest, qRecord.RawManifest); restoreErr != nil {
		return nil, api.NewError(http.StatusBadRequest, restoreErr.Error(), nil)
	}
	preserveNetworkingOwnershipContract := usesNetworkingOwnershipContract(qRecord.Configuration)

	// Get request path and body from context
	requestPath := middleware.GetRequestPath(ctx)
	requestBody := middleware.GetRequestBody(ctx)

	// Validate that the existing workload was created by the same request path, i.e. the same API type, e.g. nim-service.
	// Since the Update API contains the workload id, we are looking for the same prefix, not the exact path.
	if !strings.HasPrefix(requestPath, qRecord.InputRequestPath) {
		return nil, api.NewError(http.StatusForbidden, fmt.Sprintf("cannot update workload from a different API type: expected %s to start with %s", requestPath, qRecord.InputRequestPath), nil)
	}

	// Validate that the updated manifest's GVK matches the original workload's GVK
	err = w.validateSameGVK(manifest, qRecord.ResourceGroup, qRecord.ResourceVersion, qRecord.ResourceKind)
	if err != nil {
		return nil, err
	}

	// Reject updates to v1-cast-to-v2 workloads whose type was immutable in v1.
	// Suspend goes through a separate endpoint and is unaffected.
	if immutable, workloadKind := isImmutableV1CastWorkload(qRecord.TranslatedFromV1, qRecord.RawManifest, w.k8sUtil); immutable {
		return nil, api.NewError(http.StatusForbidden,
			fmt.Sprintf("workload submitted via the v1 API (workloadKind: %s) is immutable; use suspend or delete instead",
				workloadKind), nil)
	}

	project, orgUnitErr := orgUnitClient.Get().GetProjectById(ctx, qRecord.ProjectId)
	if orgUnitErr != nil {
		return nil, api.NewError(int(orgUnitErr.Code), orgUnitErr.Message, nil)
	}

	manifestResourceInfo, resourceErr := w.k8sUtil.GetResourceInfo(manifest)
	if resourceErr != nil {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("failed to extract resource info from manifest: %v", err), nil)
	}
	workloadType, gvkErr := w.getWorkloadTypeFromGVK(ctx, manifestResourceInfo, qRecord.ClusterId)
	if gvkErr != nil {
		return nil, SubmissionErrorToApiError(gvkErr)
	}
	if segmentErr := w.validateSegmentTopologyClusterSupport(ctx, qRecord.ClusterId, manifest); segmentErr != nil {
		return nil, SubmissionErrorToApiError(segmentErr)
	}

	ancestors := &orgUnitApi.Ancestors{
		TenantId:     tenantId,
		ClusterUuid:  *orgUnitApi.NewNullableString(&project.ClusterId),
		DepartmentId: *orgUnitApi.NewNullableString(&project.ParentId),
		ProjectId:    *orgUnitApi.NewNullableString(&project.Id),
	}

	if !skipAuth {
		_, authErr := validateCreationAuth(ctx, ancestors, project.Id)
		if authErr != nil {
			return nil, authErr
		}
	}

	// Merge metadata: use new values if provided, otherwise keep existing values
	mergedMetadata := api.WorkloadV2MetadataCreateParams{
		Name:     qRecord.Name,
		Priority: qRecord.PriorityClass,
		Category: qRecord.Category,
	}

	// Override with new values if provided
	if request.Metadata.Priority != nil {
		mergedMetadata.Priority = request.Metadata.Priority
	}
	if request.Metadata.Category != nil {
		mergedMetadata.Category = request.Metadata.Category
	}

	// Validate category and priority
	priorityNameToValue, subErr := validateCategoryAndPriority(ctx, project.ClusterId, &mergedMetadata)
	if subErr != nil {
		return nil, SubmissionErrorToApiError(subErr)
	}

	// Apply default mutations from workload type if not set in request.
	// In this case we are calling this function to handle preemptibility calculation, per the given priority or workload type.
	effectiveMetadata := w.applyDefaultMutations(mergedMetadata, workloadType, priorityNameToValue)

	// Edit manifest with metadata
	effectiveManifest, _, editErr := w.editManifestWithMetadata(
		ctx,
		manifest,
		effectiveMetadata,
		project,
		workloadId,
		qRecord.TranslatedFromV1,
		nodePoolsKarta(ctx, workloadType, manifestResourceInfo.Version),
	)
	if editErr != nil {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("failed to edit manifest: %v", editErr), nil)
	}

	resourceInfo, resourceErr := w.k8sUtil.GetResourceInfo(effectiveManifest)
	if resourceErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to extract resource info: %v", resourceErr), nil)
	}

	// Convert request body to JsonBMap if present
	var requestBodyMap dbmodels.JsonBMap
	if len(requestBody) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(requestBody, &bodyMap); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to unmarshal request body: %v", err)
		} else {
			requestBodyMap = bodyMap
		}
	}

	// Update the record
	qRecord.UpdatedAt = time.Now().Truncate(time.Microsecond)
	qRecord.UpdatedBy = subjectId

	qRecord.Category = effectiveMetadata.Category
	qRecord.PriorityClass = effectiveMetadata.Priority

	qRecord.ResourceGroup = resourceInfo.Group
	qRecord.ResourceVersion = resourceInfo.Version
	qRecord.ResourceKind = resourceInfo.Kind

	qRecord.RawManifest = manifest
	qRecord.EffectiveManifest = effectiveManifest

	// Recompute configuration only when the caller sent it; a priority/category- or manifest-only
	// update (configuration omitted) leaves the create-time configuration untouched. When the caller
	// does send configuration, it fully replaces the previous one, so an explicit empty configuration
	// clears networking/MNNVL/credentials/PVCs.
	if request.Metadata.Configuration != nil {
		configuration, configErr := w.buildConfiguration(ctx, *request.Metadata.Configuration, project, workloadType, manifestResourceInfo, qRecord.Name, networkingSubmissionProfileForRecord(qRecord))
		if configErr != nil {
			return nil, SubmissionErrorToApiError(configErr)
		}
		qRecord.Configuration = configuration
	}

	// Leaving qRecord.InputRequestPath as is, to keep the creation request path, not the update one.
	qRecord.InputRequestBody = requestBodyMap

	if convutils.IsEmptyOrNull(effectiveMetadata.Category) {
		qRecord.Category = convutils.Ref(DefaultCategory)
	}

	if effectiveMetadata.Preemptibility != nil {
		qRecord.Preemptibility = convutils.Ref(string(*effectiveMetadata.Preemptibility))
	}

	for _, modifier := range modifiers {
		modifier(qRecord)
	}

	polErrs := ApplyPolicy(ctx, qRecord, PolicyMaxErrorsInUpdate)
	if len(polErrs) > 0 {
		return nil, polErrs[0]
	}

	// Existing records created before this contract remain unchanged. Records
	// created with the contract keep it across full configuration replacement.
	if preserveNetworkingOwnershipContract {
		enableNetworkingOwnershipContract(&qRecord.Configuration)
		if bindingErr := applyNetworkingOwnershipBinding(ctx, qRecord, workloadType); bindingErr != nil {
			return nil, SubmissionErrorToApiError(NewConfigurationError(fmt.Sprintf("failed to apply networking ownership binding: %v", bindingErr)))
		}
	}

	dbErr = dao.DbUpsertWorkload(ctx, qRecord)
	if dbErr != nil {
		return nil, dbErr
	}

	responsePayload := qRecord.ToWorkloadV2()
	if !skipAuth {
		redactWorkloadV2SensitiveValues(&responsePayload)
	}

	return &responsePayload, nil
}

func (w *WorkloadsV2Impl) DeleteWorkloadV2(ctx context.Context, workloadId string) (*api.WorkloadV2, *api.Error) {
	permittedScopes, scopesErr := authClient.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	subjectId, apiErr := runaictx.ParseSubjectId(ctx)
	if apiErr != nil {
		return nil, api.NewValidationError(apiErr.Error())
	}

	qRecord, dbErr := dao.DbGetWorkload(ctx, workloadId, permittedScopes)
	if dbErr != nil {
		return nil, dbErr
	}

	qRecord.DesiredPhase = string(api.DESIREDPHASE_DELETED)
	qRecord.UpdatedBy = subjectId
	qRecord.UpdatedAt = time.Now()

	delErr := dao.DbUpsertWorkload(ctx, qRecord)
	if delErr != nil {
		return nil, dbErr
	}

	notifyErr := workloadsClient.Get().PatchWorkload(ctx, qRecord.ID, workloadsApi.WorkloadPatchFields{
		Phase: workloadsApi.PHASE_DELETING,
	})
	if notifyErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", notifyErr.Message, notifyErr.Details)
	}

	response := qRecord.ToWorkloadV2()
	redactWorkloadV2SensitiveValues(&response)

	return &response, nil
}

// ProcessBatch is called internally from workloads service. here we only do the DB work, everything
// related to updating the phase in workloads table is done by workloads service
func (w *WorkloadsV2Impl) ProcessBatch(ctx context.Context, batch api.WorkloadsV2Batch) (*api.BatchResponse, *api.Error) {
	permittedScopes, scopesErr := authClient.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	response := api.BatchResponse{
		Id:        uuid.NewString(),
		Succeeded: make([]string, 0, len(batch.Ids)),
		Failed:    make([]api.BatchResponseFailureItem, 0, len(batch.Ids)),
	}

	switch batch.Action {
	case api.WORKLOADSV2BATCHACTION_DELETE:
		subjectId, apiErr := runaictx.ParseSubjectId(ctx)
		if apiErr != nil {
			return nil, api.NewValidationError(apiErr.Error())
		}

		recordsMap, dbErr := dao.DbListByIdsToMap(ctx, batch.Ids, permittedScopes)
		if dbErr != nil {
			return nil, dbErr
		}

		for _, id := range batch.Ids {
			qRecord, found := recordsMap[id]
			if !found {
				response.Failed = append(response.Failed, api.BatchResponseFailureItem{
					Id:      id,
					Code:    http.StatusNotFound,
					Message: fmt.Sprintf("workload with id '%s' not found", id),
				})
				continue
			}

			qRecord.DesiredPhase = string(api.DESIREDPHASE_DELETED)
			qRecord.UpdatedBy = subjectId
			qRecord.UpdatedAt = time.Now()

			if delErr := dao.DbUpsertWorkload(ctx, qRecord); delErr != nil {
				response.Failed = append(response.Failed, api.BatchResponseFailureItem{
					Id:      id,
					Code:    delErr.Code,
					Message: delErr.Message,
				})
				continue
			}

			response.Succeeded = append(response.Succeeded, id)
		}

	case api.WORKLOADSV2BATCHACTION_SUSPEND, api.WORKLOADSV2BATCHACTION_RESUME:
		runState := api.RUNSTATE_SUSPEND
		if batch.Action == api.WORKLOADSV2BATCHACTION_RESUME {
			runState = api.RUNSTATE_RESUME
		}

		for _, id := range batch.Ids {
			_, runStateErr := w.SetWorkloadRunState(ctx, id, api.WorkloadRunStateRequest{Value: runState})
			if runStateErr != nil {
				response.Failed = append(response.Failed, api.BatchResponseFailureItem{
					Id:      id,
					Code:    runStateErr.Code,
					Message: runStateErr.Message,
				})
				continue
			}

			response.Succeeded = append(response.Succeeded, id)
		}

	default:
		return nil, api.NewError(http.StatusBadRequest,
			fmt.Sprintf("unsupported batch action %q", batch.Action), nil)
	}

	return &response, nil
}

func (w *WorkloadsV2Impl) GetWorkloadV2Updates(ctx context.Context, clusterId string, syncAfter int64, limit int32,
	clusterReport api.ClusterSyncReport,
) (*api.WorkloadsV2Updates, *api.Error) {
	dbErr := dao.DbApplyClusterReport(ctx, clusterId, &clusterReport)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to apply cluster report from %s", clusterId)
	}

	notificationCtx := context.WithoutCancel(ctx)
	go notifySyncIssues(notificationCtx, clusterId, &clusterReport)

	//
	//   now that we processed the cluster issues, time to send a new bulk of updates to the agent
	//
	workloadsToSend, dbErr := dao.DbListUpdates(ctx, clusterId, syncAfter, int(limit))
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to list update for cluster %s", clusterId)
	}

	response := api.WorkloadsV2Updates{
		RequestInfo: api.RequestInfo{
			AfterUpdate: syncAfter,
		},
		SystemInfo: api.SystemInfo{
			HighestUpdate: dao.GetHighestUpdateCount(ctx),
		},
		Workloads: make([]api.WorkloadV2Update, 0, limit),
	}

	if len(workloadsToSend) == 0 {
		return &response, nil
	}

	response.ResponseInfo.LowestUpdate = workloadsToSend[0].UpdateCount
	response.ResponseInfo.HighestUpdate = workloadsToSend[0].UpdateCount

	workloadsById := make(map[string]models.WorkloadsQueueRecord)

	for _, workloadToSend := range workloadsToSend {
		response.Workloads = append(response.Workloads, workloadToClusterUpdateWithNetworkingSubmissionProfile(&workloadToSend))

		workloadsById[workloadToSend.ID] = workloadToSend

		response.ResponseInfo.HighestUpdate = max(response.ResponseInfo.HighestUpdate, workloadToSend.UpdateCount)
		response.ResponseInfo.LowestUpdate = min(response.ResponseInfo.LowestUpdate, workloadToSend.UpdateCount)
	}

	dbErr = dao.DbUpdateSentToAgent(ctx, clusterId, response, workloadsById)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to sent to cluster for cluster %s", clusterId)
	}

	return &response, nil
}

func validateCreationAuth(ctx context.Context, ancestors *orgUnitApi.Ancestors, projectId string) (*authApi.PermittedScopes, *api.Error) {
	//   get the permitted scopes from the context
	permittedScopes, scopesErr := authClient.ParsePermittedScopes(ctx)
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

func validateClusterSupportsV2Workloads(ctx context.Context, clusterId string) *api.SubmissionErrorV2 {
	capabilitiesChecker := checker.NewClusterCapabilitiesChecker(clusterId)

	clusterSupportWorkloadV2, checkerError := capabilitiesChecker.ValidateClusterCapability(ctx, checker.MinVersionOfGenericWorkloadsSupport)
	if checkerError != nil {
		return NewClusterError(checkerError.Error())
	}

	if !clusterSupportWorkloadV2 {
		return NewClusterError(fmt.Sprintf("This API is incompatible with the version of cluster '%s'", clusterId))
	}

	return nil
}

func notifyWorkloadCreated(ctx context.Context, qRecord *models.WorkloadsQueueRecord) {
	notification := workloadCreationNotificationFromQueueRecord(qRecord)
	sc := config.Get[*ourConfig.Server]()

	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.MaxInterval = time.Duration(sc.RetryMaxIntervalSec) * time.Second
	exponentialBackoff.MaxElapsedTime = time.Duration(sc.RetryMaxIntervalSec) * time.Second

	var err *workloadsApi.Error

	// we need to preserve the token, because the workloads service extracts the tenant ID from it
	ctxWithJwtToken := context.WithValue(ctx, workloadsApi.ContextAccessToken, ctx.Value(middlewares.ContextJwtToken))

	ticker := backoff.NewTicker(exponentialBackoff)
	defer ticker.Stop()
	for range ticker.C {
		_, err = workloadsClient.Get().CreateWorkload(ctxWithJwtToken, notification)
		if err == nil {
			return
		}
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to notify workloads service: workload %s (%s) workload type %s error %v",
		notification.Id, notification.Name, notification.Type, err)
}

func workloadCreationNotificationFromQueueRecord(qRecord *models.WorkloadsQueueRecord) workloadsApi.WorkloadCreationFields {
	workloadType := qRecord.ResourceKind
	if nativeType, ok := nativeWorkloadTypeFromV1Cast(qRecord.TranslatedFromV1, qRecord.RawManifest); ok {
		workloadType = nativeType
	}
	var distributedFramework *string
	if workloadType == string(workloadsApi.WORKLOADTYPE_DISTRIBUTED) {
		if framework := distributedFrameworkFromKind(qRecord.ResourceKind); framework != "" {
			distributedFramework = convutils.Ref(framework)
		}
	}

	return workloadsApi.WorkloadCreationFields{
		Type:        workloadType,
		Group:       convutils.Ref(qRecord.ResourceGroup),
		Version:     convutils.Ref(qRecord.ResourceVersion),
		Name:        qRecord.Name,
		Id:          qRecord.ID,
		SubmittedBy: convutils.Ref(qRecord.CreatedBy),

		ClusterId:      qRecord.ClusterId,
		ProjectId:      qRecord.ProjectId,
		ProjectName:    convutils.Ref(qRecord.ProjectName),
		DepartmentId:   convutils.Ref(qRecord.DepartmentId),
		DepartmentName: convutils.Ref(qRecord.DepartmentName),

		CreatedAt:            qRecord.CreatedAt,
		FullySaasWorkload:    convutils.Ref(true),
		Category:             qRecord.Category,
		DistributedFramework: distributedFramework,
		SourceApi:            *workloadsApi.NewNullableSourceApi(convutils.Ref(workloadsApi.SOURCEAPI_WORKLOADS_V2)),
		Preemptibility:       *workloadsApi.NewNullableString(qRecord.Preemptibility),
	}
}

func notifySyncIssues(ctx context.Context, clusterId string, report *api.ClusterSyncReport) {
	//
	//   we also need to send the report to the workloads service to update his status
	//
	clusterReportNotif := workloadsApi.ClusterSyncReport{
		Updates: make([]workloadsApi.ClusterSyncReportUpdate, 0),
	}

	for _, report := range report.Updates {
		clusterReportNotif.Updates = append(clusterReportNotif.Updates, workloadsApi.ClusterSyncReportUpdate{
			Id:                convutils.Ref(report.Id),
			UpdateCount:       convutils.Ref(report.UpdateCount),
			UpdateAttemptTime: convutils.Ref(report.UpdateAttemptTime),
			Message:           *workloadsApi.NewNullableString(report.Message),
		})
	}

	sc := config.Get[*ourConfig.Server]()

	exponentialBackoff := backoff.NewExponentialBackOff()
	exponentialBackoff.MaxInterval = time.Duration(sc.RetryMaxIntervalSec) * time.Second
	exponentialBackoff.MaxElapsedTime = time.Duration(sc.RetryMaxIntervalSec) * time.Second

	var err *workloadsApi.Error

	// we need to preserve the token, because the workloads service extracts the tenant ID from it
	ctxWithJwtToken := context.WithValue(ctx, workloadsApi.ContextAccessToken, ctx.Value(middlewares.ContextJwtToken))

	ticker := backoff.NewTicker(exponentialBackoff)
	defer ticker.Stop()
	for range ticker.C {
		err := workloadsClient.Get().ProcessSyncReport(ctxWithJwtToken, clusterId, clusterReportNotif)
		if err == nil {
			return
		}
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to notify workloads sync issues of cluster %s: %v", clusterId, err)
}

func (w *WorkloadsV2Impl) inferProjectFromClusterAndNamespace(ctx context.Context, clusterId string, manifest map[string]interface{}) (*orgUnitApi.ProjectMetadata, *orgUnitApi.Error) {
	resourceInfo, err := w.k8sUtil.GetResourceInfo(manifest)
	if err != nil {
		return nil, orgUnitApi.NewError(http.StatusBadRequest, "failed to extract resource info from manifest")
	}

	namespace := resourceInfo.Namespace
	if namespace == "" {
		return nil, orgUnitApi.NewError(http.StatusBadRequest, "namespace must be specified in the manifest when specifying clusterId instead of projectId")
	}

	projects, orgUnitErr := orgUnitClient.Get().GetProjectsMetadataByCluster(ctx, clusterId)
	if orgUnitErr != nil {
		return nil, orgUnitErr
	}

	for _, project := range projects {
		projectNamespace := ""
		if convutils.IsStringHasValue(project.Namespace.Get()) {
			projectNamespace = *project.Namespace.Get()
		} else {
			projectNamespace = fmt.Sprintf("%s%s", RunaiNamespacePrefix, project.Name)
		}

		if projectNamespace == namespace {
			return &project, nil
		}
	}

	return nil, orgUnitApi.NewError(http.StatusNotFound, fmt.Sprintf("no project found for cluster %s with namespace %s", clusterId, namespace))
}

// UpdateWorkloadNetworkingConfig updates only the networking configuration of a workload record.
func (w *WorkloadsV2Impl) UpdateWorkloadNetworkingConfig(ctx context.Context, workloadId string, netCfg *[]api.NetworkingEndpoint) *api.Error {
	permittedScopes, _ := authClient.ParsePermittedScopes(ctx)
	qRecord, dbErr := dao.DbGetWorkload(ctx, workloadId, permittedScopes)
	if dbErr != nil {
		return dbErr
	}

	if networkingErr := validateWorkloadV2NetworkingForCluster(ctx, qRecord.ClusterId, qRecord.Name, netCfg, networkingSubmissionProfileForRecord(qRecord)); networkingErr != nil {
		return SubmissionErrorToApiError(networkingErr)
	}

	if qRecord.Configuration == nil {
		qRecord.Configuration = &models.WorkloadConfiguration{}
	}
	qRecord.Configuration.Networking = cloneNetworkingConfig(netCfg)
	if usesNetworkingOwnershipContract(qRecord.Configuration) {
		resourceInfo, resourceErr := w.k8sUtil.GetResourceInfo(qRecord.EffectiveManifest)
		if resourceErr != nil {
			return api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to extract resource info: %v", resourceErr), nil)
		}
		workloadType, gvkErr := w.getWorkloadTypeFromGVK(ctx, resourceInfo, qRecord.ClusterId)
		if gvkErr != nil {
			return SubmissionErrorToApiError(gvkErr)
		}
		if bindingErr := applyNetworkingOwnershipBinding(ctx, qRecord, workloadType); bindingErr != nil {
			return SubmissionErrorToApiError(NewConfigurationError(fmt.Sprintf("failed to apply networking ownership binding: %v", bindingErr)))
		}
	}

	return dao.DbUpsertWorkload(ctx, qRecord)
}

func validateWorkloadV2NetworkingForCluster(ctx context.Context, clusterId string, workloadName string, netCfg *[]api.NetworkingEndpoint, profile api.NetworkingSubmissionProfile) *api.SubmissionErrorV2 {
	if netCfg == nil {
		return nil
	}
	if err := validateNetworkingEndpoints(*netCfg); err != nil {
		return err
	}

	capChecker := checker.NewClusterCapabilitiesChecker(clusterId)
	useWorkloadV2Networking, capErr := capChecker.ValidateClusterCapability(ctx, checker.MinVersionNIMIngressViaWorkloadV2)
	if capErr != nil {
		return NewConfigurationError(fmt.Sprintf("cluster capability check: %v", capErr))
	}
	if !useWorkloadV2Networking {
		minVersion := checker.MinVersionForSupporting(checker.MinVersionNIMIngressViaWorkloadV2)
		return NewConfigurationError(fmt.Sprintf("metadata.configuration.networking is supported for cluster version %s and above", minVersion))
	}

	// Preserve the existing native/NIM compatibility contract. Only caller-supplied
	// networking requires the newer cluster-side ownership enforcement.
	if len(*netCfg) > 0 && profile != api.NETWORKINGSUBMISSIONPROFILE_GENERATED_NATIVE && profile != api.NETWORKINGSUBMISSIONPROFILE_GENERATED_NIM {
		supportsOwnership, capErr := capChecker.ValidateClusterCapability(ctx, checker.MinVersionNetworkingOwnershipEnforcement)
		if capErr != nil {
			return NewConfigurationError(fmt.Sprintf("cluster capability check: %v", capErr))
		}
		if !supportsOwnership {
			minVersion := checker.MinVersionForSupporting(checker.MinVersionNetworkingOwnershipEnforcement)
			return NewConfigurationError(fmt.Sprintf("direct-v2 metadata.configuration.networking requires cluster version %s and above for networking ownership validation", minVersion))
		}
	}

	supportsNetworkingEndpoints, capErr := capChecker.ValidateClusterCapability(ctx, checker.MinVersionWorkloadV2NetworkingEndpoints)
	if capErr != nil {
		return NewConfigurationError(fmt.Sprintf("cluster capability check: %v", capErr))
	}
	if supportsNetworkingEndpoints {
		return nil
	}

	networkingConfig := models.NetworkingConfiguration(*netCfg)
	if err := networkingConfig.AgentNetworkingCompatibilityError(workloadName); err != nil {
		minVersion := checker.MinVersionForSupporting(checker.MinVersionWorkloadV2NetworkingEndpoints)
		return NewConfigurationError(fmt.Sprintf("metadata.configuration.networking includes values that require cluster version %s and above: %v", minVersion, err))
	}

	return nil
}

// validateNetworkingEndpoints rejects malformed endpoint structures before queueing.
// Generated decoding allows null array entries, and runaihttp's item-count helpers
// do not enforce bounds on pointer-backed slices.
func validateNetworkingEndpoints(endpoints []api.NetworkingEndpoint) *api.SubmissionErrorV2 {
	for i, endpoint := range endpoints {
		field := fmt.Sprintf("metadata.configuration.networking[%d]", i)
		hasService := endpoint.Services != nil && len(*endpoint.Services) > 0
		access := endpoint.ExternalAccess
		hasExposure := access != nil && (access.Ingress != nil || access.Route != nil || access.HttpRoute != nil)
		if !hasService && !hasExposure {
			return NewConfigurationError(field + " must contain a service or externalAccess with an Ingress, Route, or HTTPRoute")
		}
		if endpoint.Services == nil {
			continue
		}
		if len(*endpoint.Services) > 1 {
			return NewConfigurationError(field + ".services must contain at most one service")
		}
		for j, service := range *endpoint.Services {
			switch {
			case service.Spec == nil && service.Ref == nil:
				return NewConfigurationError(fmt.Sprintf("%s.services[%d] must contain spec or ref", field, j))
			case service.Spec != nil && service.Ref != nil:
				return NewConfigurationError(fmt.Sprintf("%s.services[%d] must not contain both spec and ref", field, j))
			}
		}
	}
	return nil
}

func cloneNetworkingConfig(net *[]api.NetworkingEndpoint) *models.NetworkingConfiguration {
	if net == nil {
		return nil
	}
	out := make(models.NetworkingConfiguration, len(*net))
	for i, endpoint := range *net {
		out[i] = cloneNetworkingEndpoint(endpoint)
	}
	return &out
}

func cloneNetworkingEndpoint(endpoint api.NetworkingEndpoint) api.NetworkingEndpoint {
	return api.NetworkingEndpoint{
		ExternalAccess: cloneExternalAccess(endpoint.ExternalAccess),
		Services:       cloneNetworkingServices(endpoint.Services),
	}
}

func cloneExternalAccess(access *api.ExternalAccess) *api.ExternalAccess {
	if access == nil {
		return nil
	}
	return &api.ExternalAccess{
		Ingress:   cloneInterfaceMapPtr(access.Ingress),
		Route:     cloneInterfaceMapPtr(access.Route),
		HttpRoute: cloneInterfaceMapPtr(access.HttpRoute),
	}
}

func cloneNetworkingServices(services *[]api.NetworkingService) *[]api.NetworkingService {
	if services == nil {
		return nil
	}
	out := make([]api.NetworkingService, len(*services))
	for i, service := range *services {
		out[i] = api.NetworkingService{
			Spec:          cloneServiceSpec(service.Spec),
			Ref:           clonePtr(service.Ref),
			Authorization: cloneWorkloadv2Authorization(service.Authorization),
		}
	}
	return &out
}

func cloneServiceSpec(spec *api.ServiceSpec) *api.ServiceSpec {
	if spec == nil {
		return nil
	}
	out := &api.ServiceSpec{
		Name:     spec.Name,
		Type:     clonePtr(spec.Type),
		Selector: cloneStringMapPtr(spec.Selector),
		Labels:   maps.Clone(spec.Labels),
		Ports:    make([]api.ServicePort, len(spec.Ports)),
	}
	for i, port := range spec.Ports {
		out.Ports[i] = api.ServicePort{
			Name:       clonePtr(port.Name),
			Port:       port.Port,
			TargetPort: clonePtr(port.TargetPort),
			NodePort:   clonePtr(port.NodePort),
			Protocol:   clonePtr(port.Protocol),
		}
	}
	return out
}

func cloneWorkloadv2Authorization(auth *api.Workloadv2Authorization) *api.Workloadv2Authorization {
	if auth == nil {
		return nil
	}
	return &api.Workloadv2Authorization{
		AuthorizationType: clonePtr(auth.AuthorizationType),
		AuthorizedUsers:   cloneStringSlicePtr(auth.AuthorizedUsers),
		AuthorizedGroups:  cloneStringSlicePtr(auth.AuthorizedGroups),
		ForceJsonResponse: clonePtr(auth.ForceJsonResponse),
	}
}

func clonePtr[T any](in *T) *T {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

func cloneStringSlicePtr(in *[]string) *[]string {
	if in == nil {
		return nil
	}
	out := append([]string(nil), (*in)...)
	return &out
}

func cloneStringMapPtr(in *map[string]string) *map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string, len(*in))
	for key, value := range *in {
		out[key] = value
	}
	return &out
}

func cloneInterfaceMapPtr(in *map[string]interface{}) *map[string]interface{} {
	if in == nil {
		return nil
	}
	out := cloneInterfaceMap(*in)
	return &out
}

func cloneInterfaceMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for key, value := range in {
		out[key] = cloneInterfaceValue(value)
	}
	return out
}

func cloneInterfaceSlice(in []interface{}) []interface{} {
	out := make([]interface{}, len(in))
	for i, value := range in {
		out[i] = cloneInterfaceValue(value)
	}
	return out
}

func cloneInterfaceValue(in interface{}) interface{} {
	switch value := in.(type) {
	case map[string]interface{}:
		return cloneInterfaceMap(value)
	case []interface{}:
		return cloneInterfaceSlice(value)
	default:
		return value
	}
}

// getComponentDefinition fetches the root ComponentDefinition for a GVK with a short TTL cache.
func (w *WorkloadsV2Impl) getComponentDefinition(ctx context.Context, group, version, kind string) (*kartav1alpha1.ComponentDefinition, *workloadsApi.Error) {
	key := fmt.Sprintf("%s/%s/%s", group, version, kind)
	cached, err := cache.Get[componentDefCacheEntry](ctx, key)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get component definition from cache for %s: %v", key, err)
	}
	if cached != nil {
		return cached.Def, nil
	}
	def, apiErr := fetchComponentDefinition(ctx, workloadsClient.Get(), group, version, kind)
	if apiErr != nil {
		return nil, apiErr
	}
	if err := cache.Set[componentDefCacheEntry](ctx, key, componentDefCacheEntry{Def: def}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to cache component definition for %s: %v", key, err)
	}
	return def, nil
}

func fetchComponentDefinition(ctx context.Context, client workloadsClient.ClientInterface, group, version, kind string) (*kartav1alpha1.ComponentDefinition, *workloadsApi.Error) {
	resp, apiErr := client.ListWorkloadTypes(ctx, getFilters(group, kind))
	if apiErr != nil {
		return nil, apiErr
	}
	for _, wt := range resp.Types {
		for _, k := range wt.GetKartas() {
			if k.GetVersion() == version {
				return extractComponentDefinition(k.GetKarta())
			}
		}
	}
	return nil, nil
}

func extractComponentDefinition(kartaMap map[string]interface{}) (*kartav1alpha1.ComponentDefinition, *workloadsApi.Error) {
	b, err := json.Marshal(kartaMap)
	if err != nil {
		return nil, workloadsApi.NewError(500, fmt.Sprintf("failed to marshal karta: %v", err))
	}
	var karta kartav1alpha1.Karta
	if err := json.Unmarshal(b, &karta); err != nil {
		return nil, workloadsApi.NewError(500, fmt.Sprintf("failed to unmarshal karta: %v", err))
	}
	componentDef := karta.Spec.StructureDefinition.RootComponent
	return &componentDef, nil
}

// getActualPhase resolves the workload's actual (cluster-reported) phase from the workloads service.
// It returns an empty phase when the phase can't be determined (workloads service unavailable or the
// workload not yet known to it), in which case callers fall back to the stored DesiredPhase.
func (w *WorkloadsV2Impl) getActualPhase(ctx context.Context, workloadId string) workloadsApi.Phase {
	workload, err := workloadsClient.Get().GetWorkload(ctx, workloadId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf(
			"failed to fetch actual phase for workload %s from workloads service, falling back to desired phase: %v", workloadId, err.Message)
		return ""
	}
	if workload == nil {
		return ""
	}
	return workload.Phase
}

// unstoppablePhases lists the actual (cluster-reported) phases in which a suspend request is
// rejected. NOTE: this map must stay identical to the duplicate maintained in the workloads service
// (pkg/services_crud/change_desired_phase.go). The duplication is intentional: the workloads service
// is slated for removal, after which this copy becomes the single source of truth.
var unstoppablePhases = map[workloadsApi.Phase]bool{
	workloadsApi.PHASE_CREATING:    true,
	workloadsApi.PHASE_STOPPING:    true,
	workloadsApi.PHASE_DELETING:    true,
	workloadsApi.PHASE_STOPPED:     true,
	workloadsApi.PHASE_FAILED:      true,
	workloadsApi.PHASE_COMPLETED:   true,
	workloadsApi.PHASE_TERMINATING: true,
}

// validateSuspendAllowed gates suspend on the actual cluster phase when it is known, and falls back to
// the stored DesiredPhase otherwise. The actual-phase branch is intentionally kept identical to the
// workloads service (pkg/services_crud/change_desired_phase.go).
func validateSuspendAllowed(qRecord *models.WorkloadsQueueRecord, actualPhase workloadsApi.Phase) *api.Error {
	if actualPhase != "" {
		if unstoppablePhases[actualPhase] {
			return api.NewError(http.StatusBadRequest,
				fmt.Sprintf("suspend request is not permitted for %s workloads", actualPhase), nil)
		}
		return nil
	}
	if api.DesiredPhase(qRecord.DesiredPhase) != api.DESIREDPHASE_RUNNING {
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("cannot suspend workload in desired phase %q; only Running workloads can be suspended", qRecord.DesiredPhase), nil)
	}
	return nil
}

// validateResumeAllowed gates resume on the actual cluster phase when it is known, and falls back to
// the stored DesiredPhase otherwise. This allows resuming a workload that was suspended cluster-side
// (e.g. by the runtime-rules), where DesiredPhase would still read Running. The actual-phase
// branch is intentionally kept identical to the workloads service.
func validateResumeAllowed(qRecord *models.WorkloadsQueueRecord, actualPhase workloadsApi.Phase) *api.Error {
	if actualPhase != "" {
		if actualPhase != workloadsApi.PHASE_STOPPED {
			return api.NewError(http.StatusBadRequest,
				fmt.Sprintf("resume request is permitted only for stopped workloads (currently the workload is in %s phase)", actualPhase), nil)
		}
		return nil
	}
	if api.DesiredPhase(qRecord.DesiredPhase) != api.DESIREDPHASE_STOPPED {
		return api.NewError(http.StatusBadRequest,
			fmt.Sprintf("cannot resume workload in desired phase %q; only Stopped workloads can be resumed", qRecord.DesiredPhase), nil)
	}
	return nil
}

// SetWorkloadRunState implements PATCH /api/v2/workloads/{id}/runState.
func (w *WorkloadsV2Impl) SetWorkloadRunState(ctx context.Context, workloadId string, request api.WorkloadRunStateRequest) (*api.WorkloadV2, *api.Error) {
	permittedScopes, _ := authClient.ParsePermittedScopes(ctx)

	subjectId, apiErr := runaictx.ParseSubjectId(ctx)
	if apiErr != nil {
		return nil, api.NewValidationError(apiErr.Error())
	}

	qRecord, dbErr := dao.DbGetWorkload(ctx, workloadId, permittedScopes)
	if dbErr != nil {
		return nil, dbErr
	}

	// The stored DesiredPhase only reflects run-state changes made through this API. It does not
	// capture cluster-driven transitions (e.g. a workload suspended by the runtime-rules or one
	// that has failed). We therefore consult the actual phase from the workloads service and gate
	// suspend/resume on it, matching V1 behavior. When the actual phase is unavailable we fall back
	// to DesiredPhase so the operation still works when the workloads service can't be reached.
	// Temporary solution: actualPhase is fetched from the workloads service directly.
	// This should be replaced once get logic is fully handled within the workloads-manager.
	actualPhase := w.getActualPhase(ctx, workloadId)

	switch request.Value {
	case api.RUNSTATE_SUSPEND:
		if apiErr := validateSuspendAllowed(qRecord, actualPhase); apiErr != nil {
			return nil, apiErr
		}
	case api.RUNSTATE_RESUME:
		if apiErr := validateResumeAllowed(qRecord, actualPhase); apiErr != nil {
			return nil, apiErr
		}
	default:
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("unknown run state %q", request.Value), nil)
	}

	capChecker := checker.NewClusterCapabilitiesChecker(qRecord.ClusterId)
	clusterSupported, capErr := capChecker.ValidateClusterCapability(ctx, checker.MinVersionOfAnyWorkloadSuspendResume)
	if capErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("cluster capability check: %v", capErr), nil)
	}
	if !clusterSupported {
		return nil, api.NewError(http.StatusBadRequest,
			"suspend/resume is supported for cluster version 2.26 and above", nil)
	}

	componentDef, clientErr := w.getComponentDefinition(ctx,
		qRecord.ResourceGroup, qRecord.ResourceVersion, qRecord.ResourceKind)
	if clientErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to fetch component definition: %s", clientErr.Message), nil)
	}
	if componentDef == nil || componentDef.SuspendDefinition == nil {
		return nil, api.NewError(http.StatusBadRequest,
			"workload type does not support suspend/resume", nil)
	}

	suspend := request.Value == api.RUNSTATE_SUSPEND
	updatedRaw, patchErr := patchManifestSuspendActions(ctx, qRecord.RawManifest, *componentDef, suspend)
	if patchErr != nil {
		return nil, patchErr
	}
	updatedEffective, patchErr := patchManifestSuspendActions(ctx, qRecord.EffectiveManifest, *componentDef, suspend)
	if patchErr != nil {
		return nil, patchErr
	}

	if request.Value == api.RUNSTATE_SUSPEND {
		qRecord.DesiredPhase = string(api.DESIREDPHASE_STOPPED)
	} else {
		qRecord.DesiredPhase = string(api.DESIREDPHASE_RUNNING)
	}

	qRecord.RawManifest = updatedRaw
	qRecord.EffectiveManifest = updatedEffective
	qRecord.UpdatedAt = time.Now().Truncate(time.Microsecond)
	qRecord.UpdatedBy = subjectId

	dbErr = dao.DbUpdateWorkloadOptimistic(ctx, qRecord, qRecord.UpdateCount)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to persist run-state change for workload %s: %v", workloadId, dbErr.Message)
		return nil, dbErr
	}

	responsePayload := qRecord.ToWorkloadV2()
	redactWorkloadV2SensitiveValues(&responsePayload)
	return &responsePayload, nil
}

// patchManifestSuspendActions applies karta suspend/resume actions to a manifest clone via the karta accessor.
func patchManifestSuspendActions(ctx context.Context, manifest dbmodels.JsonBMap, componentDef kartav1alpha1.ComponentDefinition, suspend bool) (dbmodels.JsonBMap, *api.Error) {
	if manifest == nil {
		return manifest, nil
	}

	acc := kartaresource.NewAccessor(execution.NewDefaultRunner(manifest.Clone().AsMAp()))
	var err error
	if suspend {
		err = acc.ApplySuspendActions(ctx, componentDef)
	} else {
		err = acc.ApplyResumeActions(ctx, componentDef)
	}
	if err != nil {
		var defNotFound kartaresource.DefinitionNotFoundError
		if errors.As(err, &defNotFound) {
			return nil, api.NewError(http.StatusBadRequest, "workload type does not support suspend/resume", nil)
		}
		return nil, api.NewError(http.StatusBadRequest,
			fmt.Sprintf("failed to apply suspend/resume patch: %v", err), nil)
	}

	patched, err := acc.GetObject()
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("failed to read patched manifest: %v", err), nil)
	}
	return dbmodels.JsonBMap(patched), nil
}
