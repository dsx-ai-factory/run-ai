/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gvk_compliance"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/database/models"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

// ApplyPolicy loads the policy for the workload and runs its rules against the
// workload's effective manifest. Validation errors are accumulated (one per
// rule at most per cycle) up to maxErrors, allowing the caller to surface
// multiple issues in a single submission response. A fatal error (e.g. policy
// load failure) is still returned as a single-element slice.
func ApplyPolicy(ctx context.Context, workload *models.WorkloadsQueueRecord, maxErrors int) []*api.Error {
	// Workloads translated from v1 native types already had the original native
	// policy applied on the v1 path, so skip the v2 gvk policy here.
	if workload.TranslatedFromV1 {
		return nil
	}

	gvk := policyApi.GVK{
		Kind:    workload.ResourceKind,
		Group:   *policyApi.NewNullableString(convutils.Ref(workload.ResourceGroup)),
		Version: workload.ResourceVersion,
	}

	// get the policy if there is one for this workload
	pol, err := policies.Get().Client().GetGvkPolicyOfProject(ctx, workload.ProjectId, gvk)
	if err != nil {
		if err.Code == http.StatusNotFound {
			return nil
		}
		logging.Errorf("failed to load policy of project %s: %s/%s/%s: %v", workload.ProjectId,
			convutils.Deref(gvk.Group.Get(), ""), gvk.Kind, gvk.Version, err)
		return []*api.Error{api.NewError(int(err.Code), err.Message, nil)}
	}

	// resolve all the rules according to the breakdown of the policy
	// broken rules will be logged and excluded from the set of rules that needs to run
	rules := gvk_compliance.ResolveRules(pol)

	// run all the rules which were resolved
	manifest := map[string]any(workload.EffectiveManifest)

	// Pass spec selectors to enable dependency path creation for SET/DEFAULT actions
	var specSelectors map[string]policyApi.BreakdownSpecSelector
	if bd := pol.Breakdown.Get(); bd != nil && bd.SpecSelectors != nil {
		specSelectors = *bd.SpecSelectors
	}

	errs := gvk_compliance.RunRulesWithDependencies(manifest, rules, specSelectors, maxErrors, gvk_compliance.AllCycles)
	if len(errs) == 0 {
		return nil
	}

	errResp := make([]*api.Error, 0, len(errs))
	for _, err := range errs {
		errResp = append(errResp, api.NewError(int(err.Code), err.Message, nil))
	}
	return errResp
}
