/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	workloadsClient "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/workloads"
	k8sConstants "github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/constants"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"gopkg.in/yaml.v3"
)

func (w *WorkloadsV2Impl) validateManifest(_ context.Context, manifest map[string]interface{}, manifestBase64Encoded string) (map[string]interface{}, *api.Error) {
	if manifest != nil && manifestBase64Encoded != "" {
		return nil, api.NewValidationError("manifest and manifestBase64Encoded are mutually exclusive")
	}

	if manifest == nil && manifestBase64Encoded == "" {
		return nil, api.NewValidationError("either manifest or manifestBase64Encoded must be provided")
	}

	if manifestBase64Encoded != "" {
		manifest = make(map[string]interface{})

		decodedBytes, decodeErr := base64.StdEncoding.DecodeString(manifestBase64Encoded)
		if decodeErr != nil {
			return nil, api.NewValidationError(fmt.Sprintf("failed to decode base64 manifest: %s", decodeErr.Error()))
		}

		var yamlData interface{}
		if yamlErr := yaml.Unmarshal(decodedBytes, &yamlData); yamlErr != nil {
			return nil, api.NewValidationError(fmt.Sprintf("failed to parse YAML manifest: %s", yamlErr.Error()))
		}

		jsonBytes, jsonErr := json.Marshal(yamlData)
		if jsonErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to convert YAML to JSON", jsonErr)
		}

		if unmarshalErr := json.Unmarshal(jsonBytes, &manifest); unmarshalErr != nil {
			return nil, api.NewError(http.StatusInternalServerError, "failed to unmarshal JSON manifest", unmarshalErr)
		}
	}

	if apiErr := w.k8sUtil.ValidateK8sResource(manifest); apiErr != nil {
		return nil, apiErr
	}

	return manifest, nil
}

func (w *WorkloadsV2Impl) getWorkloadTypeFromGVK(ctx context.Context, resourceInfo *ResourceInfo, clusterId string) (*workloadsApi.WorkloadTypeConfig, *api.SubmissionErrorV2) {
	workloadType, typeErr := w.validateWorkloadTypeFromAPI(ctx, resourceInfo.Group, resourceInfo.Version, resourceInfo.Kind, clusterId)
	if typeErr != nil {
		return nil, typeErr
	}

	return workloadType, nil
}

func (w *WorkloadsV2Impl) validateWorkloadTypeFromAPI(ctx context.Context, group, version, kind, clusterId string) (*workloadsApi.WorkloadTypeConfig, *api.SubmissionErrorV2) {
	filters := getFilters(group, kind)
	response, err := workloadsClient.Get().ListWorkloadTypes(ctx, filters)
	if err != nil {
		return nil, NewGVKError(fmt.Sprintf("failed to get workload types: %v", err.Message))
	}

	var matchingType *workloadsApi.WorkloadTypeConfig
	if len(response.Types) == 1 {
		matchingType = &response.Types[0]
	}

	if matchingType == nil {
		gvkStr := fmt.Sprintf("%s/%s/%s", group, version, kind)
		if group == "" {
			gvkStr = fmt.Sprintf("%s/%s", version, kind)
		}
		return nil, NewGVKError(fmt.Sprintf("workload type not found for GVK: %s", gvkStr))
	}

	// For Run.ai created types with no kartas - we know it is ready
	hasKartas := len(matchingType.Kartas) > 0
	if hasKartas || matchingType.CreatedBy == nil || *matchingType.CreatedBy != "Run.ai" {
		// Validate status - check if workload type is ready for the specific cluster
		if err := w.validateWorkloadTypeStatus(matchingType, version, clusterId); err != nil {
			return nil, err
		}
	}

	return matchingType, nil
}

func getFilters(group string, kind string) []string {
	return client_utils.NewFilterBuilder().
		AddFilter("name", client_utils.EqualOp, kind).
		AddFilter("group", client_utils.EqualOp, group).Get()
}

func (w *WorkloadsV2Impl) validateWorkloadTypeStatus(workloadType *workloadsApi.WorkloadTypeConfig, version, clusterId string) *api.SubmissionErrorV2 {
	if len(workloadType.ClusterStatuses) == 0 {
		return NewGVKError("workload type has no cluster status information")
	}

	// Find status for the specific cluster we're submitting to
	var clusterStatus *workloadsApi.WorkloadTypeStatus
	for _, status := range workloadType.ClusterStatuses {
		if status.ClusterId != nil &&
			*status.ClusterId == clusterId &&
			status.Version != nil &&
			*status.Version == version {
			clusterStatus = &status
			break
		}
	}

	if clusterStatus == nil {
		return NewGVKError(fmt.Sprintf("workload type has no status information for cluster %s and version %s", clusterId, version))
	}

	if clusterStatus.Phase == nil {
		return NewGVKError("workload type phase is not available for this cluster")
	}

	if !w.isReadyPhase(*clusterStatus.Phase) {
		// Include phase message in error response
		errorMsg := fmt.Sprintf("workload type is not ready in cluster %s (phase: %s)", clusterId, *clusterStatus.Phase)

		// Add condition messages if available
		if len(clusterStatus.Conditions) > 0 {
			var messages []string
			for _, condition := range clusterStatus.Conditions {
				if condition.Message != nil && *condition.Message != "" {
					messages = append(messages, fmt.Sprintf("%s: %s", condition.Type, *condition.Message))
				}
			}
			if len(messages) > 0 {
				errorMsg += fmt.Sprintf(" - %s", strings.Join(messages, "; "))
			}
		}

		return NewGVKError(errorMsg)
	}

	return nil
}

func (w *WorkloadsV2Impl) isReadyPhase(phase string) bool {
	// TODO: use enum or consts when the api is defined
	if phase == "Ready" {
		return true
	}
	return false
}

func (w *WorkloadsV2Impl) applyDefaultMutations(metadata api.WorkloadV2MetadataCreateParams, workloadType *workloadsApi.WorkloadTypeConfig,
	priorityNameToValue map[string]int32,
) api.WorkloadV2MetadataCreateParams {
	// Create a copy of the metadata to avoid modifying the original
	mutated := metadata

	// Apply default priority if not set
	if (mutated.Priority == nil || *mutated.Priority == "") && workloadType.PriorityName != nil && *workloadType.PriorityName != "" {
		mutated.Priority = workloadType.PriorityName
	}

	// Apply default category if not set
	if (mutated.Category == nil || *mutated.Category == "") && workloadType.CategoryName != nil && *workloadType.CategoryName != "" {
		mutated.Category = workloadType.CategoryName
	}

	if metadata.Preemptibility == nil && mutated.Priority != nil {
		priorityValue := priorityNameToValue[*mutated.Priority]
		if priorityValue >= NonPreemptibleMinLevel {
			mutated.Preemptibility = convutils.Ref(api.PREEMPTIBILITY_NON_PREEMPTIBLE)
		} else {
			mutated.Preemptibility = convutils.Ref(api.PREEMPTIBILITY_PREEMPTIBLE)
		}
	}

	return mutated
}

func (w *WorkloadsV2Impl) editManifestWithMetadata(ctx context.Context, manifest map[string]interface{},
	metadata api.WorkloadV2MetadataCreateParams, project *orgUnitApi.Project, workloadId string,
	trustedConvertedFromNative bool,
) (map[string]interface{}, string, *api.Error) {
	modifiedManifest, err := w.k8sUtil.updateResourceMetadataWithOptions(
		ctx,
		manifest,
		workloadId,
		project,
		resourceMetadataUpdateOptions{trustedConvertedFromNative: trustedConvertedFromNative},
	)
	if err != nil {
		return nil, "", api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to add workload ID to resource: %v", err), err)
	}

	manifestMetadata, ok := modifiedManifest[k8sConstants.FieldMetadata].(map[string]interface{})
	if !ok {
		return nil, "", api.NewValidationError("manifest metadata is not a valid map")
	}

	if !shouldPreserveManifestName(modifiedManifest) {
		manifestMetadata[k8sConstants.FieldName] = metadata.Name
	}
	delete(manifestMetadata, k8sConstants.FieldGenerateName)

	labels, ok := manifestMetadata[k8sConstants.FieldLabels].(map[string]interface{})
	if !ok {
		labels = make(map[string]interface{})
		manifestMetadata[k8sConstants.FieldLabels] = labels
	}

	if convutils.IsStringHasValue(metadata.Priority) {
		labels[PriorityLabel] = *metadata.Priority
	}

	if convutils.IsStringHasValue(metadata.Category) {
		labels[CategoryLabel] = *metadata.Category
	}

	if metadata.Preemptibility != nil {
		labels[constants.LabelPreemptibilityName] = *metadata.Preemptibility
	}

	labels[ControlPlaneManagedLabel] = TrueAsString

	if w.shouldStampRestartWorkflowLabel(ctx, modifiedManifest, project.ClusterId, trustedConvertedFromNative) {
		w.stampRestartWorkflowLabel(labels, workloadId)
	}

	if err := w.applyGVKSpecificProcessing(modifiedManifest, project); err != nil {
		return nil, "", api.NewError(http.StatusBadRequest, err.Error(), err)
	}

	return modifiedManifest, workloadId, nil
}

func shouldPreserveManifestName(manifest map[string]interface{}) bool {
	if manifest["apiVersion"] != constants.LwsGroup+"/"+constants.LwsVersion || manifest["kind"] != constants.LwsKind {
		return false
	}
	metadata, ok := manifest[k8sConstants.FieldMetadata].(map[string]interface{})
	if !ok {
		return false
	}
	name, ok := metadata[k8sConstants.FieldName].(string)
	if !ok || name == "" {
		return false
	}
	labels, ok := metadata[k8sConstants.FieldLabels].(map[string]interface{})
	return ok && labels[ConvertedFromNativeLabel] == TrueAsString
}

func validateCategoryAndPriority(ctx context.Context, clusterId string, meta *api.WorkloadV2MetadataCreateParams) (map[string]int32, *api.SubmissionErrorV2) {
	priorityNameToValue := make(map[string]int32)

	available, err := workloadsClient.Get().GetAvailablePrioritiesAndCategories(ctx, &clusterId, nil, nil, nil)
	if err != nil && err.Code != http.StatusNotFound {
		return nil, ToSubmissionErrorV2(api.NewError(int(err.Code), err.Message, fmt.Errorf("%v", err.Details)))
	}

	if convutils.IsStringHasValue(meta.Category) {
		foundCategory := false
		if available != nil && len(available.Categories) > 0 {
			for _, category := range available.Categories {
				if *meta.Category == category.Name {
					foundCategory = true
					break
				}
			}
		}
		if !foundCategory {
			return nil, NewCategoryError(fmt.Sprintf("category '%s' is not available in this cluster", *meta.Category))
		}
	}

	foundPriority := false
	if available != nil && len(available.Categories) > 0 {
		for _, priority := range available.Priorities {
			if convutils.Deref(meta.Priority, "") == priority.Name {
				foundPriority = true
			}

			priorityNameToValue[priority.Name] = priority.Value
		}
	}

	if convutils.IsStringHasValue(meta.Priority) && !foundPriority {
		return nil, NewPriorityError(fmt.Sprintf("priority '%s' is not available in this cluster", *meta.Priority))
	}

	return priorityNameToValue, nil
}

// Validate that the updated manifest's GVK matches the original workload's GVK,
// to prevent changing the GVK by providing a completely different manifest than the original.
func (w *WorkloadsV2Impl) validateSameGVK(manifest map[string]interface{}, resourceGroup, resourceVersion, resourceKind string) *api.Error {
	resourceInfo, err := w.k8sUtil.GetResourceInfo(manifest)
	if err != nil {
		return api.NewValidationError(fmt.Sprintf("failed to extract resource info: %v", err))
	}

	if resourceInfo.Group != resourceGroup || resourceInfo.Version != resourceVersion || resourceInfo.Kind != resourceKind {
		return api.NewValidationError(fmt.Sprintf("manifest GVK does not match the original workload: %v", err))
	}

	return nil
}
