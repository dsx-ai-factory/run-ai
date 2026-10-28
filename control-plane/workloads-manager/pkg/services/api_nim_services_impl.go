/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	nimv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/nim_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/workloadutils"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	clustersClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/clusters"
	orgUnitClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/middleware"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nimServiceKind       = "NIMService"
	nimServiceAPIVersion = "apps.nvidia.com/v1alpha1"
	LWSNotAvailableMsg   = "NIMService can't be created because LeaderWorkerSet (LWS) is not installed on the selected cluster. Contact your administrator."
)

type NimServicesImpl struct {
	workloadsImpl *WorkloadsV2Impl
}

func NewNimServicesImpl(workloadsImpl *WorkloadsV2Impl) *NimServicesImpl {
	return &NimServicesImpl{
		workloadsImpl: workloadsImpl,
	}
}

func (nsi *NimServicesImpl) CreateNimService(ctx context.Context, request api.NimServiceCreateRequest) (*api.NimServiceResponse, *api.SubmissionErrorV2) {
	// validate LWS installation if multiNode is requested
	if err := nsi.validateLwsInstallation(ctx, &request); err != nil {
		return nil, ToSubmissionError(err)
	}

	policyRules, policyDefaults, _, policyErr := GetPolicyOfProject(ctx, WorkloadTypeNimService, request.Metadata.ProjectId)
	if policyErr != nil {
		return nil, ToSubmissionError(policyErr)
	}

	mutatedSpec, mutateErr := MutateWorkload[api.NimServiceSpec](&request.Spec, nil /*template spec when supported*/, policyDefaults)
	if mutateErr != nil {
		return nil, mutateErr
	}

	// Even though the prepare for mutation is primarily needed for mutation function, we also need to call it before validateSpecAgainstPolicy.
	// - Transforms itemized fields from policy format (with .instances and .attributes sections) to workload format (arrays)
	// - Adds policy keys (nicknames) for itemized instances without keys
	// - Adds field name as a dummy instance for merging purposes
	// This transformation is critical because compliance validation (itemKeys function) expects both
	// spec and defaults to have the same structure when comparing keys for canAdd/locked validation.
	policy_compliance.PrepareDefaultsForMutation(policyDefaults)

	// Validate policy compliance on the mutated workload
	_, complianceErr := validateMutatedSpecVsPolicy(policyRules, policyDefaults, mutatedSpec)
	if complianceErr != nil {
		return nil, complianceErr
	}

	// Keep an independent copy before merging defaults into nested request fields.
	var submittedSpec api.NimServiceSpec
	if err := convutils.CopySpecs(&submittedSpec, request.Spec); err != nil {
		return nil, NewSubmissionError(http.StatusInternalServerError, "failed to copy the submitted NIM spec", err)
	}
	if request.Spec.Compute != nil {
		compute := *request.Spec.Compute
		compute.GpuPortionRequest, compute.GpuPortionLimit = nil, nil
		request.Spec.Compute = &compute
	}
	copyErr := convutils.CopySpecs(&request.Spec, mutatedSpec)
	if copyErr != nil {
		return nil, NewSubmissionError(http.StatusInternalServerError, "failed to prepare the workload spec for submission", copyErr)
	}

	// exposedUrl and auth can only be set when exposeExternally=true.
	if err := validateNimServingPortUrlRequiresExposure(request.Spec.ServingPort); err != nil {
		return nil, ToSubmissionError(err)
	}
	if err := validateNimServingPortAuthRequiresExposure(request.Spec.ServingPort); err != nil {
		return nil, ToSubmissionError(err)
	}
	if err := validateNimServingPortAuthFieldsConsistency(request.Spec.ServingPort); err != nil {
		return nil, ToSubmissionError(err)
	}

	manifest, netInfo, err := nsi.composeNimServiceManifest(ctx, &request.Metadata, request.Spec, false)
	if err != nil {
		return nil, ToSubmissionError(err)
	}

	workloadRequest := api.WorkloadV2CreateRequest{
		Metadata: api.WorkloadV2MetadataCreateParams{
			Name:                 request.Metadata.Name,
			UseGivenNameAsPrefix: false,
			ProjectId:            request.Metadata.ProjectId,
			Priority:             request.Spec.PriorityClass,
			Category:             request.Spec.Category,
			Preemptibility:       request.Spec.Preemptibility,
		},
		Manifest: manifest,
	}

	if netInfo.UseWorkloadV2Networking {
		netCfg, err := nimServingPortToNetworkingConfig(request.Metadata.Name, request.Spec.ServingPort, netInfo)
		if err != nil {
			return nil, ToSubmissionError(err)
		}
		if netCfg != nil {
			workloadRequest.Metadata.Configuration.Networking = netCfg
		}
	}

	// Persist the selected mode without changing NIM's original request in the DB for other defaults.
	if request.Spec.Compute != nil && request.Spec.Compute.GpuComputeMode != nil {
		compute := api.NimServiceSpecAllOfCompute{}
		if submittedSpec.Compute != nil {
			compute = *submittedSpec.Compute
		}
		compute.GpuComputeMode = request.Spec.Compute.GpuComputeMode
		submittedSpec.Compute = &compute
	}
	ctx, ctxErr := nsi.updateContextWithMergedSpec(ctx, &request.Metadata, submittedSpec)
	if ctxErr != nil {
		return nil, NewSubmissionError(http.StatusInternalServerError, "failed to save NIM request with resolved GPU mode", ctxErr)
	}
	workloadV2Response, submissionErrors := nsi.workloadsImpl.CreateWorkloadV2(ctx, workloadRequest, false)
	if len(submissionErrors) > 0 {
		return nil, &submissionErrors[0]
	}

	response := &api.NimServiceResponse{
		Metadata:     workloadV2Response.Metadata,
		DesiredPhase: workloadV2Response.DesiredPhase,
		Spec:         legacyCompatibleNimSpec(request.Spec),
	}
	redactNimServiceResponseSensitiveValues(response)

	return response, nil
}

func (nsi *NimServicesImpl) composeNimServiceManifest(
	ctx context.Context, metadata *api.NimServiceMetadataCreateParams,
	spec api.NimServiceSpec, forceOldFlowIngress bool,
) (map[string]interface{}, nimServiceNetworkingInfo, *api.Error) {
	nimService, netInfo, apiErr := nsi.composeNimService(ctx, metadata, spec, forceOldFlowIngress)
	if apiErr != nil {
		return nil, nimServiceNetworkingInfo{}, apiErr
	}

	// convert nimService into a manifest
	var manifest map[string]interface{}
	err := convutils.CopySpecs(&manifest, nimService)
	if err != nil {
		return nil, nimServiceNetworkingInfo{}, api.NewError(http.StatusInternalServerError, "failed to copy NIMService to manifest", err)
	}

	removeEmptyExposeRouter(manifest)

	return manifest, netInfo, nil
}

// removeEmptyExposeRouter drops the empty spec.expose.router that the NIM operator's
// value-type Router field always serializes (json omitempty does not omit empty structs).
// NIMService CRDs that predate the OSS-KServe build (e.g. k8s-nim-operator v3.0.0) do not
// declare expose.router, and the cluster-side server-side apply rejects the undeclared
// field, so the NIMService is never created.
func removeEmptyExposeRouter(manifest map[string]interface{}) {
	spec, ok := manifest["spec"].(map[string]interface{})
	if !ok {
		return
	}
	expose, ok := spec["expose"].(map[string]interface{})
	if !ok {
		return
	}
	if router, ok := expose["router"].(map[string]interface{}); ok && len(router) == 0 {
		delete(expose, "router")
	}
}

func (nsi *NimServicesImpl) composeNimService(
	ctx context.Context, metadata *api.NimServiceMetadataCreateParams,
	spec api.NimServiceSpec, forceOldFlowIngress bool,
) (*nimv1alpha1.NIMService, nimServiceNetworkingInfo, *api.Error) {
	nimService := &nimv1alpha1.NIMService{}

	// get project
	project, orgUnitErr := orgUnitClient.Get().GetProjectById(ctx, metadata.ProjectId)
	if orgUnitErr != nil {
		return nil, nimServiceNetworkingInfo{}, api.NewError(int(orgUnitErr.Code), orgUnitErr.Message, nil)
	}

	// handle TypeMeta
	nimService.TypeMeta = metav1.TypeMeta{
		Kind:       nimServiceKind,
		APIVersion: nimServiceAPIVersion,
	}

	cluster, getErr := clustersClient.Get().GetClusterByUuid(ctx, project.ClusterId)
	if getErr != nil {
		return nil, nimServiceNetworkingInfo{}, api.NewError(int(getErr.Code), fmt.Sprintf("failed to get cluster info: %v", getErr), nil)
	}

	if authErr := validateNimServingPortAuthSupportedForClusterVersion(ctx, cluster.GetVersion(), spec.ServingPort); authErr != nil {
		return nil, nimServiceNetworkingInfo{}, authErr
	}

	var netInfo nimServiceNetworkingInfo
	var gatewayAPIEnabled *bool
	if cluster.Status.IsSet() && cluster.Status.Get().Config.IsSet() {
		cfg := cluster.Status.Get().Config.Get()
		if v := cfg.IngressClass.Get(); v != nil {
			netInfo.IngressClass = *v
		}
		netInfo.IngressEnabled = cfg.IngressEnabled.Get()
		if n := cfg.GatewayName.Get(); n != nil && *n != "" {
			netInfo.GatewayName = *n
		}
		if ns := cfg.GatewayNamespace.Get(); ns != nil && *ns != "" {
			netInfo.GatewayNamespace = *ns
		}
		gatewayAPIEnabled = cfg.GatewayAPIEnabled.Get()
	}
	if cluster.Status.IsSet() && cluster.Status.Get().HasPlatform() {
		p := cluster.Status.Get().GetPlatform()
		netInfo.UseOpenShiftRoute = strings.EqualFold(p.GetType(), "openshift")
	}
	gatewayAPIDependencyAvailable := false
	if cluster.Status.IsSet() && cluster.Status.Get().HasDependencies() {
		deps := cluster.Status.Get().GetDependencies()
		if gwDep, ok := deps.Optional["gateway-api"]; ok && gwDep.Available {
			gatewayAPIDependencyAvailable = true
		}
	}
	netInfo.UseGatewayAPI = effectiveNimGatewayAPIEnabled(gatewayAPIEnabled, gatewayAPIDependencyAvailable, netInfo.GatewayName, netInfo.GatewayNamespace)

	// handle Spec
	name := nsi.getNimServiceName(ctx, *metadata)
	metadata.Name = name

	requester, parseErr := runaictx.ParseSubjectId(ctx)
	if parseErr != nil {
		return nil, nimServiceNetworkingInfo{}, api.NewError(http.StatusInternalServerError, "failed to identify the subject ID", parseErr)
	}

	nimComposer := nim_workloads.NewNimServiceComposer(name, requester, cluster, project.Name, composer.OptionsForNimService())
	nimComposer.SetForceOldFlowIngress(forceOldFlowIngress)

	var generalizedSpec openapi.NimServiceSpec
	if err := convutils.CopySpecs(&generalizedSpec, spec); err != nil {
		return nil, nimServiceNetworkingInfo{}, api.NewError(http.StatusInternalServerError, "failed to copy specs", err)
	}

	if generalizedSpec.ServingPort.IsSet() {
		u, err := nimComposer.GetAutoExposureUrl(generalizedSpec.ServingPort.Get())
		if err != nil {
			return nil, nimServiceNetworkingInfo{}, api.NewError(http.StatusBadRequest, fmt.Sprintf("failed to compute exposure URL: %v", err), err)
		}
		netInfo.AutoExposureUrl = u
	}

	err := nimComposer.Compose(ctx, &generalizedSpec, &nimService.Spec)()
	if err != nil {
		return nil, nimServiceNetworkingInfo{}, api.NewError(int(err.Code), "failed to compose NIMService spec", err)
	}

	// Read back the networking decision made by the composer during composition.
	netInfo.UseWorkloadV2Networking = nimComposer.UsedNewFlowIngress()

	return nimService, netInfo, nil
}

func (nsi *NimServicesImpl) validateLwsInstallation(ctx context.Context, request *api.NimServiceCreateRequest) *api.Error {
	// Check if multi-node is requested
	if request.Spec.MultiNode == nil {
		// No multi-node requested, skip validation
		return nil
	}

	// Get project to retrieve cluster ID
	project, orgUnitErr := orgUnitClient.Get().GetProjectById(ctx, request.Metadata.ProjectId)
	if orgUnitErr != nil {
		return api.NewError(int(orgUnitErr.Code), orgUnitErr.Message, nil)
	}

	err := workloadutils.CheckLWSAvailable(ctx, project.ClusterId, LWSNotAvailableMsg)
	if err != nil {
		return api.NewError(int(err.Code), err.Message, nil)
	}

	return nil
}

func (nsi *NimServicesImpl) getNimServiceName(ctx context.Context, metadata api.NimServiceMetadataCreateParams) string {
	return nsi.workloadsImpl.GetWorkloadName(ctx, &api.WorkloadV2MetadataCreateParams{
		Name:                 metadata.Name,
		UseGivenNameAsPrefix: metadata.UseGivenNameAsPrefix,
	}, nil)
}

func (nsi *NimServicesImpl) GetNimService(ctx context.Context, id string) (*api.NimServiceResponse, *api.Error) {
	// Check if this workload was created by nim-services route
	// We need to access the database record to get InputRequestPath and InputRequestBody
	// For now, let's get the database record directly
	_, response, apiErr := nsi.getNimServiceInternal(ctx, id)
	redactNimServiceResponseSensitiveValues(response)
	return response, apiErr
}

func (nsi *NimServicesImpl) getNimServiceInternal(ctx context.Context, id string) (*models.WorkloadsQueueRecord, *api.NimServiceResponse, *api.Error) {
	permittedScopes, scopesErr := authClient.ParsePermittedScopes(ctx)
	if scopesErr != nil {
		permittedScopes = nil
	}

	qRecord, dbErr := dao.DbGetWorkload(ctx, id, permittedScopes)
	if dbErr != nil {
		return nil, nil, dbErr
	}

	// Check if this workload was created via nim-services route
	if qRecord.InputRequestPath != NimServicesAPIPath {
		return nil, nil, api.NewError(http.StatusBadRequest, "workload was not created via nim-services route", nil)
	}

	workloadV2Response := qRecord.ToGetResponse()

	// Parse the original request body to get the NimServiceCreateRequest
	var originalRequest api.NimServiceCreateRequest
	if qRecord.InputRequestBody != nil {
		// JsonBMap is a wrapper for map[string]interface{}, so we need to marshal it back to JSON first
		requestBodyBytes, marshalErr := json.Marshal(map[string]interface{}(qRecord.InputRequestBody))
		if marshalErr != nil {
			return nil, nil, api.NewError(http.StatusInternalServerError, "failed to marshal request body", marshalErr)
		}

		if unmarshalErr := json.Unmarshal(requestBodyBytes, &originalRequest); unmarshalErr != nil {
			return nil, nil, api.NewError(http.StatusInternalServerError, "failed to unmarshal original request", unmarshalErr)
		}
	} else {
		return nil, nil, api.NewError(http.StatusInternalServerError, "original request body not found", nil)
	}

	nimService := &api.NimServiceResponse{
		Metadata:     workloadV2Response.Workload.Metadata,
		DesiredPhase: workloadV2Response.Workload.DesiredPhase,
		Spec:         legacyCompatibleNimSpec(originalRequest.Spec),
	}

	return qRecord, nimService, nil
}

func ToSubmissionError(err *api.Error) *api.SubmissionErrorV2 {
	return &api.SubmissionErrorV2{
		Code:    err.Code,
		Message: err.Message,
		Details: err.Details,
	}
}

func NewSubmissionError(code int, message string, details error) *api.SubmissionErrorV2 {
	result := api.SubmissionErrorV2{
		Code:    int32(code),
		Message: message,
	}
	if details != nil {
		result.Details = details.Error()
	}

	return &result
}

func isOldFlowNimService(manifest map[string]interface{}) bool {
	spec, ok := manifest["spec"].(map[string]interface{})
	if !ok {
		return false
	}
	expose, ok := spec["expose"].(map[string]interface{})
	if !ok {
		return false
	}
	ingress, ok := expose["ingress"].(map[string]interface{})
	if !ok {
		return false
	}
	enabled, ok := ingress["enabled"].(bool)
	return ok && enabled
}

func updateRequestHasAuthFields(spec api.NimServiceSpec) bool {
	return spec.ServingPort != nil && hasNimServingPortAuth(spec.ServingPort)
}

func (nsi *NimServicesImpl) UpdateNimServiceSpec(ctx context.Context, id string, request api.NimServiceUpdateRequest) (*api.NimServiceResponse, *api.Error) {
	qRecord, existingNimService, apiErr := nsi.getNimServiceInternal(ctx, id)
	if apiErr != nil {
		return nil, apiErr
	}

	oldFlow := isOldFlowNimService(qRecord.RawManifest)

	if oldFlow && updateRequestHasAuthFields(request.Spec) {
		return nil, api.NewError(http.StatusBadRequest,
			"cannot set authorization fields on a NIM service that was created with ingress managed by the NIM operator; delete and recreate the NIM service to use authorization",
			nil)
	}

	// Validate that every requested field-to-remove is a known removable field.
	// Without this, an unknown name (e.g. a typo like "multiNodexx") would be a
	// silent no-op during the merge and the update would wrongly succeed with 200.
	for _, field := range request.FieldsToRemove {
		if !field.IsValid() {
			return nil, api.NewError(http.StatusBadRequest,
				fmt.Sprintf("invalid fieldsToRemove value %q; allowed values: %v", string(field), api.AllowedNimServiceRemovableFieldsEnumValues),
				nil)
		}
	}

	fieldsToRemoveStrings := removableFieldsToStrings(request.FieldsToRemove)

	mergedSpec, err := nsi.mergeUpdateSpecIntoFullSpec(ctx, request.Spec, existingNimService.Spec, fieldsToRemoveStrings)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, "failed to merge NIMService spec for update", err)
	}

	if err := validateNimServingPortUrlRequiresExposure(mergedSpec.ServingPort); err != nil {
		return nil, err
	}
	if err := validateNimServingPortAuthRequiresExposure(mergedSpec.ServingPort); err != nil {
		return nil, err
	}
	if err := validateNimServingPortAuthFieldsConsistency(mergedSpec.ServingPort); err != nil {
		return nil, err
	}

	metadata := &api.NimServiceMetadataCreateParams{
		Name:      existingNimService.Metadata.Name,
		ProjectId: existingNimService.Metadata.ProjectId,
	}
	manifest, netInfo, apiErr := nsi.composeNimServiceManifest(ctx, metadata, mergedSpec, oldFlow)
	if apiErr != nil {
		return nil, apiErr
	}

	ctx, ctxErr := nsi.updateContextWithMergedSpec(ctx, metadata, mergedSpec)
	if ctxErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to update context with merged spec", ctxErr)
	}

	workloadRequest := api.WorkloadV2UpdateRequest{
		Metadata: api.WorkloadV2MetadataUpdateParams{
			Priority: mergedSpec.PriorityClass,
			Category: mergedSpec.Category,
		},
		Manifest: manifest,
	}

	var modifiers []RecordModifier
	if netInfo.UseWorkloadV2Networking {
		netCfg, err := nimServingPortToNetworkingConfig(existingNimService.Metadata.Name, mergedSpec.ServingPort, netInfo)
		if err != nil {
			return nil, err
		}
		modifiers = append(modifiers, func(record *models.WorkloadsQueueRecord) {
			if record.Configuration == nil {
				record.Configuration = &models.WorkloadConfiguration{}
			}
			record.Configuration.Networking = cloneNetworkingConfig(netCfg)
		})
	}

	workloadV2Response, apiErr := nsi.workloadsImpl.UpdateWorkloadV2(ctx, id, workloadRequest, modifiers...)
	if apiErr != nil {
		return nil, apiErr
	}

	nimServiceResponse := &api.NimServiceResponse{
		Metadata:     workloadV2Response.Metadata,
		DesiredPhase: workloadV2Response.DesiredPhase,
		Spec:         legacyCompatibleNimSpec(mergedSpec),
	}
	redactNimServiceResponseSensitiveValues(nimServiceResponse)

	return nimServiceResponse, nil
}

// mergeUpdateSpecIntoFullSpec merges the update spec into the full spec.
// the update spec is a partial spec, so we need to merge it into the full spec.
// the fields are merged at the root level, i.e. not recursively.
// e.g. if compute field is updated, the full spec's compute field is overwritten with the update spec's compute field.
// If the partial spec contains an empty object or empty array (not nil), the corresponding field is deleted from the full spec.
// fieldsToRemove contains field names to explicitly remove from the full spec.
func (nsi *NimServicesImpl) mergeUpdateSpecIntoFullSpec(ctx context.Context, updateSpec api.NimServiceSpec, fullSpec api.NimServiceSpec, fieldsToRemove []string) (api.NimServiceSpec, error) {
	if err := restoreRedactedNimSensitiveValues(&updateSpec, fullSpec); err != nil {
		return api.NimServiceSpec{}, err
	}

	// Convert full spec to map via JSON
	var fullSpecMap map[string]interface{}
	if err := convutils.CopySpecs(&fullSpecMap, fullSpec); err != nil {
		return api.NimServiceSpec{}, err
	}

	// First, remove explicitly specified fields
	for _, field := range fieldsToRemove {
		delete(fullSpecMap, field)
	}

	// Convert partial spec to map via JSON
	var partialSpecMap map[string]interface{}
	if err := convutils.CopySpecs(&partialSpecMap, updateSpec); err != nil {
		return api.NimServiceSpec{}, err
	}

	if err := gpucompat.ValidateMemoryRequestInput(partialSpecMap); err != nil {
		return api.NimServiceSpec{}, err
	}

	for _, source := range []*map[string]interface{}{&fullSpecMap, &partialSpecMap} {
		canonical, err := gpucompat.NormalizeLegacyGPUFields(*source)
		if err != nil {
			return api.NimServiceSpec{}, err
		}
		*source = canonical
	}

	// Merge at root level: overwrite full spec fields with partial spec fields
	// If a field in the partial spec is an empty object or empty array, delete it from the full spec
	for key, value := range partialSpecMap {
		if isEmptyMapOrSlice(value) {
			delete(fullSpecMap, key)
		} else {
			fullSpecMap[key] = value
		}
	}

	// Convert merged map back to api.NimServiceSpec
	var mergedSpec api.NimServiceSpec
	if err := convutils.CopySpecs(&mergedSpec, fullSpecMap); err != nil {
		return api.NimServiceSpec{}, err
	}

	return mergedSpec, nil
}

// isEmptyMapOrSlice checks if a value is an empty map or empty slice.
// It returns false for nil values, only true for intentionally empty collections.
func isEmptyMapOrSlice(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case map[string]interface{}:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// removableFieldsToStrings converts a slice of NimServiceRemovableFields to a slice of strings.
func removableFieldsToStrings(fields []api.NimServiceRemovableFields) []string {
	result := make([]string, len(fields))
	for i, f := range fields {
		result[i] = string(f)
	}
	return result
}

func (nsi *NimServicesImpl) updateContextWithMergedSpec(ctx context.Context, metadata *api.NimServiceMetadataCreateParams, mergedSpec api.NimServiceSpec) (context.Context, error) {
	createRequest := api.NimServiceCreateRequest{
		Metadata: *metadata,
		Spec:     nimSpecForDB(mergedSpec),
	}
	body, err := json.Marshal(createRequest)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, middleware.RequestBodyKey, body), nil
}
