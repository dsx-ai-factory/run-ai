/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"

	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

const ComplianceTasksInParallel = 10

var (
	policyJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedPolicyDefaultsV2{})
	masterSection     = policyJsonCracker.NameOf(&policyJsonCracker.Struct().Master)
	workerSection     = policyJsonCracker.NameOf(&policyJsonCracker.Struct().Worker)

	policyDinfJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedInferenceDefaultsV2{})
	leaderSection         = policyDinfJsonCracker.NameOf(&policyDinfJsonCracker.Struct().Leader)

	distributedJsonCracker = jsonutils.NewJsonCracker(policyApi.DistributedPolicyRulesV2{})
	sectionPerReplicaType  = map[workloadApi.ReplicaType]string{
		workloadApi.REPLICATYPE_MASTER: distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Master),
		workloadApi.REPLICATYPE_WORKER: distributedJsonCracker.NameOf(&distributedJsonCracker.Struct().Worker),
	}
)

type patchingTaskInfo struct {
	template *api.SupersetTemplate

	policyInfo      dao_crud.PolicyInfo
	effectivePolicy *policyCompliance.EffectivePolicyOfProject
}

func validatePolicy(ctx context.Context, workload *models_crud.Workload) (map[string]interface{}, *workloadApi.SubmissionErrorV2) {
	values, valuesErr := gojsonq.LoadJson(workload.Spec)
	if valuesErr != nil {
		return nil, &workloadApi.SubmissionErrorV2{
			Code:    http.StatusBadRequest,
			Message: "failed to load spec",
			Details: convutils.Ref(valuesErr.Error()),
		}
	}

	return validatePolicyHelper(ctx, workload.WorkloadType, workload.ProjectId, values)
}

func validatePolicyHelper(ctx context.Context, workloadType string, projectId string, values map[string]interface{}) (map[string]interface{}, *workloadApi.SubmissionErrorV2) {
	effectivePolicy, getErr := policyCompliance.GetEffectivePolicyOfProject(ctx, policies.Get().Client(), projectId, workloadApi.WorkloadType(workloadType))
	if getErr != nil {
		return nil, &workloadApi.SubmissionErrorV2{
			Code:    getErr.Code,
			Message: getErr.Message,
		}
	}

	validationErr := validateSpecAgainstPolicy(workloadType, effectivePolicy.RulesV2, effectivePolicy.DefaultsV2, values)
	if validationErr != nil {
		return nil, validationErr
	}

	return effectivePolicy.DefaultsV2, nil
}

// validateSpecAgainstPolicy validates that a mutated spec (workload or template) complies with policy rules.
// It checks the spec after mutations (template + policy defaults) have been applied and returns an error if violations are found.
// defaults: the original policy defaults (used for canEdit rule validation)
// values: the spec after mutations (template + policy defaults) have been applied
func validateSpecAgainstPolicy(
	workloadType string,
	rules map[string]interface{},
	defaults map[string]interface{},
	mutatedSpec map[string]interface{},
) *workloadApi.SubmissionErrorV2 {
	_, complianceErrors, err := policyCompliance.ValidateWorkloadCompliance(rules, defaults, mutatedSpec, workloadApi.WorkloadType(workloadType))
	if err != nil {
		return workloadApi.NewSubmissionErrorV2(http.StatusBadRequest, err.Error())
	}

	if len(complianceErrors) != 0 {
		return &workloadApi.SubmissionErrorV2{
			Code:             http.StatusBadRequest,
			Message:          "The workload request does not comply with the policy set by the administrator",
			ComplianceIssues: complianceErrors.ForWorkloads(),
		}
	}

	return nil
}

func patchPolicyCompliance(ctx context.Context, complianceParams *ComplianceParams, entries []*api.SupersetTemplate) *api.Error {
	if complianceParams == nil || len(entries) == 0 {
		return nil
	}

	// first thing we need to do is to get the policy of the project. even though we have information
	// cached in the policy_compliance table, we need to at least know if the policy in the table is up
	// to date with the latest update of the policy
	effectivePolicy, err := policyCompliance.GetEffectivePolicyOfProject(ctx, policies.Get().Client(), fmt.Sprintf("%d", complianceParams.projectId),
		workloadApi.WorkloadType(complianceParams.workloadType))
	if err != nil {
		if err.Code == http.StatusNotFound {
			return nil // no policy no compliance to validate
		}

		return api.NewError(int(err.Code), err.Message, nil)
	}

	// take note of all the templates we need to patch
	templateIds := make([]string, 0)
	templateToPatch := make(map[string]*api.SupersetTemplate)

	for idx, entry := range entries {
		templateIds = append(templateIds, entry.Meta.Id)
		templateToPatch[entry.Meta.Id] = entries[idx]
	}

	// see what you can find in the policy compliance table
	policyInfo := dao_crud.PolicyInfo{
		TenantId:     auth.TenantIdOf(ctx),
		WorkloadType: string(complianceParams.workloadType),
		ReplicaType:  complianceParams.replica,
		ProjectId:    complianceParams.projectId,
		PolicyTime:   effectivePolicy.UpdatedAt,
	}

	//
	//  try to load information from the policy_compliance table first. any asset found there we will not have to
	//  validate from scratch
	//
	complianceResults, dbErr := dao_crud.ListPolicyComplianceOfTemplates(ctx, templateIds, policyInfo)
	if dbErr != nil {
		return dbErr
	}

	for _, complianceResult := range complianceResults {
		patchPtr := templateToPatch[complianceResult.TemplateId]
		if patchPtr == nil {
			continue // cannot actually happen but validate it just to be on the safe side
		}

		patchPtr.ComplianceIssues = complianceResult.Reason

		delete(templateToPatch, complianceResult.TemplateId)
	}

	//
	//   if no remaining templates to validate it means that we were able to find all the information in the compliance table
	//   otherwise, we will have to calculate it
	//
	if len(templateToPatch) == 0 {
		return nil
	}

	//
	//   the rest we will do in parallel (like a thread pool) so create a wait group that will be responsible for
	//   orchestrating the validation all the remaining assets
	//
	var wg sync.WaitGroup
	wg.Add(len(templateToPatch))

	//
	//   create a channel that can buffer the number of templates that we intend to validate in parallel
	//   and create the thread pools
	//
	tasksChan := make(chan patchingTaskInfo, ComplianceTasksInParallel)

	for i := 0; i < ComplianceTasksInParallel; i++ {
		go ValidateComplianceOfTemplates(ctx, tasksChan, &wg)
	}

	//
	//   Send the tasks to the channel for validation
	//
	for _, template := range templateToPatch {
		tasksChan <- patchingTaskInfo{
			template:        template,
			policyInfo:      policyInfo,
			effectivePolicy: effectivePolicy,
		}
	}

	//
	//   Wait for all the tasks to be completed
	//
	wg.Wait()

	return nil
}

func ValidateComplianceOfTemplates(ctx context.Context, tasksChan chan patchingTaskInfo, wg *sync.WaitGroup) {
	for task := range tasksChan {
		validateComplianceOfTemplate(ctx, task)
		wg.Done()
	}
}

func validateComplianceOfTemplate(ctx context.Context, task patchingTaskInfo) {
	defaults := task.effectivePolicy.DefaultsV2
	rules := task.effectivePolicy.RulesV2
	var defaultFromOrder *map[string][]string
	var payloadAsMap map[string]interface{}

	if task.policyInfo.ReplicaType != "" {
		sectionOfReplica := sectionPerReplicaType[workloadApi.ReplicaType(task.policyInfo.ReplicaType)]
		if sectionOfReplica != "" {
			defaults = convutils.MapMe(defaults[sectionOfReplica])
			rules = convutils.MapMe(rules[sectionOfReplica])
		}
	}

	if workloadApi.ReplicaType(task.policyInfo.ReplicaType) == workloadApi.REPLICATYPE_MASTER {
		defaultFromOrder = task.effectivePolicy.DefaultFromOrder.Master
		payloadAsMap, _ = gojsonq.LoadJson(task.template.CombinedMasterSpec)
	} else {
		defaultFromOrder = task.effectivePolicy.DefaultFromOrder.Worker
		payloadAsMap, _ = gojsonq.LoadJson(task.template.CombinedSpec)
	}

	canonicalPayload, aliasErr := gpucompat.NormalizeLegacyGPUFields(payloadAsMap)
	if aliasErr != nil {
		task.template.ComplianceIssues = append(task.template.ComplianceIssues, api.ComplianceInfoReason{Details: aliasErr.Error()})
		return
	}
	payloadAsMap = canonicalPayload

	fieldsWithCantEditRule := make([]string, 0)
	complianceResults := policyCompliance.ComplianceErrors{}

	policyCompliance.ValidateCompliance("", rules, defaults, payloadAsMap,
		*defaultFromOrder, &fieldsWithCantEditRule, nil, &complianceResults)

	for _, complianceResult := range complianceResults {
		issue := api.ComplianceInfoReason{
			Field:   convutils.Ref(complianceResult.PolicyField),
			Details: complianceResult.Details,
		}

		if complianceResult.Rule != nil {
			issue.Rule = convutils.Ref(api.PolicyRuleEnum(*complianceResult.Rule))
		}

		task.template.ComplianceIssues = append(task.template.ComplianceIssues, issue)
	}

	//
	// store the validation result in the DB for next time. no need to store the compliance result
	//
	dbErr := dao_crud.UpsertPolicyComplianceOfTemplate(ctx, task.template.Meta.Id, task.policyInfo, task.template.ComplianceIssues)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to store compliance result of template %s for workload type %s in the db: %s: %v",
			task.template.Meta.Id, task.policyInfo.ReplicaType, dbErr.Message, dbErr.Details)
	}
}
