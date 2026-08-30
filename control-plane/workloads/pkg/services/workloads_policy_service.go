package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/policy"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	openapi "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
)

const (
	TrainingWorkload             = "TrainingWorkload"
	DistributedWorkload          = "DistributedWorkload"
	InteractiveWorkload          = "InteractiveWorkload"
	InferenceWorkload            = "InferenceWorkload"
	DistributedInferenceWorkload = "DistributedInferenceWorkload"
)

func parseWorkspacePolicyResponse(workloadPolicy *openapi.WorkspacePolicyV2, response *http.Response) *api.WorkloadTypeConfig {
	if response == nil || response.StatusCode != http.StatusOK || workloadPolicy == nil {
		return nil
	}
	result := &api.WorkloadTypeConfig{}
	if convutils.HasValue(workloadPolicy.Policy) && convutils.HasValue(workloadPolicy.Policy.Get().Defaults) {
		defaults := workloadPolicy.Policy.Get().Defaults.Get()
		result.CategoryName = convutils.Deref(defaults.Category.Get(), "")
		result.PriorityName = convutils.Deref(defaults.PriorityClass.Get(), "")
		result.Preemptibility = api.WorkloadTypePreemptibility(convutils.Deref(defaults.Preemptibility.Get(), ""))
		result.Name = InteractiveWorkload
	}
	return result
}

func parseTrainingPolicyResponse(workloadPolicy *openapi.TrainingPolicyV2, response *http.Response) *api.WorkloadTypeConfig {
	if response == nil || response.StatusCode != http.StatusOK || workloadPolicy == nil {
		return nil
	}
	result := &api.WorkloadTypeConfig{}
	if convutils.HasValue(workloadPolicy.Policy) && convutils.HasValue(workloadPolicy.Policy.Get().Defaults) {
		defaults := workloadPolicy.Policy.Get().Defaults.Get()
		result.CategoryName = convutils.Deref(defaults.Category.Get(), "")
		result.PriorityName = convutils.Deref(defaults.PriorityClass.Get(), "")
		result.Preemptibility = api.WorkloadTypePreemptibility(convutils.Deref(defaults.Preemptibility.Get(), ""))
		result.Name = TrainingWorkload
	}
	return result
}

func parseDistributedPolicyResponse(workloadPolicy *openapi.DistributedPolicyV2, response *http.Response) *api.WorkloadTypeConfig {
	if response == nil || response.StatusCode != http.StatusOK || workloadPolicy == nil {
		return nil
	}
	result := &api.WorkloadTypeConfig{}
	if convutils.HasValue(workloadPolicy.Policy) &&
		convutils.HasValue(workloadPolicy.Policy.Get().Defaults) &&
		convutils.HasValue(workloadPolicy.Policy.Get().Defaults.Get().Master) {
		defaults := workloadPolicy.Policy.Get().Defaults.Get()
		result.CategoryName = convutils.Deref(defaults.Master.Get().Category.Get(), "")
		result.PriorityName = convutils.Deref(defaults.Master.Get().PriorityClass.Get(), "")
		result.Preemptibility = api.WorkloadTypePreemptibility(convutils.Deref(defaults.Master.Get().Preemptibility.Get(), ""))
		result.Name = DistributedWorkload
	}
	return result
}

func parseInferencePolicyResponse(workloadPolicy *openapi.InferencePolicyV2, response *http.Response) *api.WorkloadTypeConfig {
	if response == nil || response.StatusCode != http.StatusOK || workloadPolicy == nil {
		return nil
	}
	result := &api.WorkloadTypeConfig{}
	if convutils.HasValue(workloadPolicy.Policy) &&
		convutils.HasValue(workloadPolicy.Policy.Get().Defaults) {
		defaults := workloadPolicy.Policy.Get().Defaults.Get()
		result.CategoryName = convutils.Deref(defaults.Category.Get(), "")
		result.PriorityName = convutils.Deref(defaults.PriorityClass.Get(), "")
		result.Preemptibility = api.WorkloadTypePreemptibility(convutils.Deref(defaults.Preemptibility.Get(), ""))
		result.Name = InferenceWorkload
	}
	return result
}

func ToSupersetDefaults(priorityName, categoryName, preemptibility string) *openapi.SupersetDefaults {
	defaults := &openapi.SupersetDefaults{}
	if categoryName != "" {
		defaults.Category.Set(convutils.Ref(categoryName))
	}
	if priorityName != "" {
		defaults.PriorityClass.Set(convutils.Ref(priorityName))
	}
	if preemptibility != "" {
		apiPreemptibility := openapi.Preemptibility(preemptibility)
		if apiPreemptibility.IsValid() {
			defaults.Preemptibility.Set(convutils.Ref(apiPreemptibility))
		}
	}
	return defaults
}

func toWorkspaceChangeRequest(tenantId, priorityName, categoryName, preemptibility string) *openapi.WorkspacePolicyChangeRequestV2 {
	defaults := &openapi.WorkspacePolicyDefaultsV2{}
	supersetDefaults := ToSupersetDefaults(priorityName, categoryName, preemptibility)
	_ = convutils.CopySpecs(defaults, supersetDefaults)
	policyInstance := openapi.NewWorkspacePolicyDefaultsAndRulesV2()
	policyInstance.Defaults.Set(defaults)
	return &openapi.WorkspacePolicyChangeRequestV2{
		Meta: &openapi.PolicyCreationFields{
			Scope: openapi.SCOPETYPE_TENANT,
			Name:  tenantId,
		},
		Policy: *openapi.NewNullableWorkspacePolicyDefaultsAndRulesV2(policyInstance),
	}
}

func toTrainingChangeRequest(tenantId, priorityName, categoryName, preemptibility string) *openapi.TrainingPolicyChangeRequestV2 {
	defaults := &openapi.TrainingPolicyDefaultsV2{}
	supersetDefaults := ToSupersetDefaults(priorityName, categoryName, preemptibility)
	_ = convutils.CopySpecs(defaults, supersetDefaults)
	policyInstance := openapi.NewTrainingPolicyDefaultsAndRulesV2()
	policyInstance.Defaults.Set(defaults)
	return &openapi.TrainingPolicyChangeRequestV2{
		Meta: &openapi.PolicyCreationFields{
			Scope: openapi.SCOPETYPE_TENANT,
			Name:  tenantId,
		},
		Policy: *openapi.NewNullableTrainingPolicyDefaultsAndRulesV2(policyInstance),
	}
}

func toInferenceChangeRequest(tenantId, priorityName, categoryName, preemptibility string) *openapi.InferencePolicyChangeRequestV2 {
	defaults := &openapi.InferencePolicyDefaultsV2{}
	supersetDefaults := ToSupersetDefaults(priorityName, categoryName, preemptibility)
	_ = convutils.CopySpecs(defaults, supersetDefaults)
	policyInstance := openapi.NewInferencePolicyDefaultsAndRulesV2()
	policyInstance.Defaults.Set(defaults)
	return &openapi.InferencePolicyChangeRequestV2{
		Meta: &openapi.PolicyCreationFields{
			Scope: openapi.SCOPETYPE_TENANT,
			Name:  tenantId,
		},
		Policy: *openapi.NewNullableInferencePolicyDefaultsAndRulesV2(policyInstance),
	}
}

func toDistributedChangeRequest(tenantId, priorityName, categoryName, preemptibility string) *openapi.DistributedPolicyChangeRequestV2 {
	master := &openapi.ReplicaDefaultsV2{}
	worker := &openapi.DistributedPolicyDefaultsV2Worker{}
	supersetDefaults := ToSupersetDefaults(priorityName, categoryName, preemptibility)
	_ = convutils.CopySpecs(master, supersetDefaults)
	_ = convutils.CopySpecs(worker, supersetDefaults)
	defaults := &openapi.DistributedPolicyDefaultsV2{
		Master: *openapi.NewNullableReplicaDefaultsV2(master),
		Worker: *openapi.NewNullableDistributedPolicyDefaultsV2Worker(worker),
	}
	policyInstance := openapi.NewDistributedPolicyDefaultsAndRulesV2()
	policyInstance.Defaults.Set(defaults)

	return &openapi.DistributedPolicyChangeRequestV2{
		Meta: &openapi.PolicyCreationFields{
			Scope: openapi.SCOPETYPE_TENANT,
			Name:  tenantId,
		},
		Policy: *openapi.NewNullableDistributedPolicyDefaultsAndRulesV2(policyInstance),
	}
}

func UpdateNativeWorkloadType(ctx context.Context, tenantId uint, updateFields dao.WorkloadTypeUpdateFieldsParsed) (*api.WorkloadTypeConfig, *api.Error) {
	var categoryName string
	var priorityName string
	var preemptibility string
	if updateFields.CategoryId != nil {
		category, err := dao.DbGetCategoryById(ctx, tenantId, updateFields.CategoryId.String())
		if err != nil {
			return nil, api.NewError(http.StatusBadRequest, err.Message, nil)
		}
		categoryName = category.Name
	}
	if updateFields.PriorityId != nil {
		priority, err := dao.GetPriorityById(ctx, strconv.Itoa(int(tenantId)), updateFields.PriorityId.String())
		if err != nil {
			return nil, api.NewError(http.StatusBadRequest, err.Message, nil)
		}
		priorityName = priority.Name

		if updateFields.Preemptibility == nil {
			preemptibility = dao.DerivePreemptibilityFromPriority(priority)
		}
	}

	if updateFields.Preemptibility != nil {
		preemptibility = *updateFields.Preemptibility
	}

	switch updateFields.WorkloadTypeName {

	case InteractiveWorkload:
		{
			changeRequest := toWorkspaceChangeRequest(strconv.Itoa(int(tenantId)), priorityName, categoryName, preemptibility)
			updatedPolicy, response, err := policy.Get().UpdateWorkspacePolicy(ctx, *changeRequest)

			if response == nil || response.StatusCode != http.StatusOK || updatedPolicy == nil {
				return nil, api.NewError(http.StatusBadRequest, "", err)
			}
			res := parseWorkspacePolicyResponse(updatedPolicy, response)
			return res, nil
		}

	case InferenceWorkload:
		{
			changeRequest := toInferenceChangeRequest(strconv.Itoa(int(tenantId)), priorityName, categoryName, preemptibility)
			updatedPolicy, response, err := policy.Get().UpdateInferencePolicy(ctx, *changeRequest)

			if response == nil || response.StatusCode != http.StatusOK || updatedPolicy == nil {
				return nil, api.NewError(http.StatusBadRequest, "", err)
			}
			res := parseInferencePolicyResponse(updatedPolicy, response)
			return res, nil
		}
	case DistributedWorkload:
		{
			changeRequest := toDistributedChangeRequest(strconv.Itoa(int(tenantId)), priorityName, categoryName, preemptibility)
			updatedPolicy, response, err := policy.Get().UpdateDistributedPolicy(ctx, *changeRequest)

			if response == nil || response.StatusCode != http.StatusOK || updatedPolicy == nil {
				return nil, api.NewError(http.StatusBadRequest, "", err)
			}
			res := parseDistributedPolicyResponse(updatedPolicy, response)
			return res, nil
		}
	case TrainingWorkload:
		{
			changeRequest := toTrainingChangeRequest(strconv.Itoa(int(tenantId)), priorityName, categoryName, preemptibility)
			updatedPolicy, response, err := policy.Get().UpdateTrainingPolicy(ctx, *changeRequest)

			if response == nil || response.StatusCode != http.StatusOK || updatedPolicy == nil {
				return nil, api.NewError(http.StatusBadRequest, "", err)
			}
			res := parseTrainingPolicyResponse(updatedPolicy, response)
			return res, nil
		}
	}
	return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("Not supported Workload type <%s>", updateFields.WorkloadTypeName), nil)
}
