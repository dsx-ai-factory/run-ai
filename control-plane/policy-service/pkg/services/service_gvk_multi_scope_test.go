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
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/run-ai/runai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
)

// ─────────────────────────────────────────────────────────────────────────────
// Multi-scope (cluster → department → project) effective + breakdown tests.
//
// These tests pin down the cross-scope invariants of the GVK policy stack:
//
//   - When an ancestor policy changes (create / update / patch-rule /
//     delete-rule / delete-policy) every descendant's cached breakdown and
//     stored effective_rules row is cleared, and the next read of a
//     descendant transparently recomputes both.
//
//   - When a descendant is created/changed, the effective + breakdown of the
//     descendant pick up every ancestor's rules, but ancestor caches are
//     left intact (only the subtree below the *changed* org unit is
//     invalidated).
//
// Each test resets the database and the org-unit / cluster / auth mocks via
// BeforeEach so individual scenarios are fully isolated. The org-unit mock
// (`BackendMockClient`) lays out the tenant as
//   tenant=1 → cluster=ClusterIdPerTenant["1"][0] → dept=111 → project=1111
// which gives us a real three-level hierarchy without any extra plumbing.
// ─────────────────────────────────────────────────────────────────────────────

var _ = Describe("GVK Policy Service - multi-scope effective + breakdown", Ordered, func() {
	router := CreateRouter()

	var backendMock *BackendMockClient

	const (
		tenantId  = int32(1)
		deptId    = "111"
		projectId = "1111"
	)

	gvk := api.Gvk{Group: convutils.Ref("apps"), Kind: "Deployment", Version: "v1"}

	// rules use disjoint fields per scope so the test can read source-of-rule
	// off each effective rule unambiguously. all three fields belong to the
	// "allContainers" specSelector which is seeded by HandleSystemMapping.
	clusterRule := api.PolicyRule{
		Field:        "imagePullPolicy",
		SpecSelector: "allContainers",
		Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
	}
	deptRule := api.PolicyRule{
		Field:        "workingDir",
		SpecSelector: "allContainers",
		Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
	}
	projectRule := api.PolicyRule{
		Field:        "image",
		SpecSelector: "allContainers",
		Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
	}

	var clusterId string

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()
		tree.Restart()

		backendMock = NewBackendMockClient(2, 1)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))
		authorization.SetMock(NewPoliciesAuthClient(backendMock))

		clusterMock := cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)

		assets.SetMock(assets.SimpleMockClientFactory{
			SimpleMockClient: assets.NewSimpleMockClient(),
		})

		clusterId = backendMock.ClusterIdPerTenant["1"][0]
	})

	// ── per-scope create-request constructors ─────────────────────────────
	newClusterPolicyReq := func(rules []api.PolicyRule) api.CreateGvkPolicyReq {
		return api.CreateGvkPolicyReq{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_CLUSTER,
			ClusterId:    convutils.Ref(clusterId),
			Rules:        rules,
		}
	}
	newDepartmentPolicyReq := func(rules []api.PolicyRule) api.CreateGvkPolicyReq {
		return api.CreateGvkPolicyReq{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_DEPARTMENT,
			DepartmentId: convutils.Ref(deptId),
			Rules:        rules,
		}
	}
	newProjectPolicyReq := func(rules []api.PolicyRule) api.CreateGvkPolicyReq {
		return api.CreateGvkPolicyReq{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_PROJECT,
			ProjectId:    convutils.Ref(projectId),
			Rules:        rules,
		}
	}

	// scope-keyed get-by-type-and-scope shorthands
	getProjectByScope := func() api.GetGvkPolicyResp {
		return getGvkPolicyByTypeAndScopeForTest(router, tenantId, api.WorkloadTypeAndScope{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_PROJECT,
			ProjectId:    convutils.Ref(projectId),
		})
	}
	getDeptByScope := func() api.GetGvkPolicyResp {
		return getGvkPolicyByTypeAndScopeForTest(router, tenantId, api.WorkloadTypeAndScope{
			WorkloadType: gvk,
			Scope:        api.SCOPETYPE_DEPARTMENT,
			DepartmentId: convutils.Ref(deptId),
		})
	}

	// ── top-down: ancestor changes invalidate descendants ──────────────────
	Context("top-down: ancestor changes invalidate descendants", func() {

		It("adding a cluster policy on top of dept+project clears their caches and the next get rebuilds the merged effective+breakdown", func() {
			// 1. start with department + project (no cluster) and verify the
			//    natural two-level merge already works.
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			projectBeforeCluster := getGvkPolicyWithBreakdownForTest(router, tenantId, projectPolicy.Id)
			expectEffectiveRuleSources(projectBeforeCluster.EffectiveRules, map[string]api.ScopeType{
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
			expectBreakdownFields(projectBeforeCluster.Breakdown, deptRule.Field, projectRule.Field)

			// priming gets so we can later assert caches were *cleared*
			getGvkPolicyWithBreakdownForTest(router, tenantId, deptPolicy.Id)
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeTrue())
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeTrue())

			// 2. add cluster policy → must invalidate the dept+project subtree.
			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeFalse(),
				"adding a cluster policy must clear the department cache")
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeFalse(),
				"adding a cluster policy must clear the project cache")

			// 3. get-by-type-and-scope of project recomputes effective + breakdown
			//    and now reflects rules from all three scopes.
			projectAfterCluster := getProjectByScope()
			expectEffectiveRuleSources(projectAfterCluster.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
			expectBreakdownFields(projectAfterCluster.Breakdown,
				clusterRule.Field, deptRule.Field, projectRule.Field)

			// 4. dept recompute via by-type-and-scope: only cluster + dept rules.
			deptAfterCluster := getDeptByScope()
			expectEffectiveRuleSources(deptAfterCluster.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
			})
			expectBreakdownFields(deptAfterCluster.Breakdown, clusterRule.Field, deptRule.Field)

			// and the cluster policy itself sees only its own rule.
			clusterFetched := getGvkPolicyWithBreakdownForTest(router, tenantId, clusterPolicy.Id)
			expectEffectiveRuleSources(clusterFetched.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
			})
			expectBreakdownFields(clusterFetched.Breakdown, clusterRule.Field)
		})

		It("updating a cluster policy's rules (PUT) clears dept+project caches and refetch shows the new rule everywhere", func() {
			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			// prime caches
			getGvkPolicyWithBreakdownForTest(router, tenantId, clusterPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, deptPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, projectPolicy.Id)
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeTrue())
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeTrue())

			// replace cluster rule with a different one
			newClusterRule := api.PolicyRule{
				Field:        "command",
				SpecSelector: "allContainers",
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}
			doPut(router, tenantId, gvkPolicyURL(clusterPolicy.Id), api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{newClusterRule},
			})

			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeFalse(),
				"updating cluster policy must clear dept cache")
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeFalse(),
				"updating cluster policy must clear project cache")

			projectAfter := getProjectByScope()
			expectEffectiveRuleSources(projectAfter.EffectiveRules, map[string]api.ScopeType{
				ruleKey(newClusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):       api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule):    api.SCOPETYPE_PROJECT,
			})
			expectBreakdownFields(projectAfter.Breakdown,
				newClusterRule.Field, deptRule.Field, projectRule.Field)

			deptAfter := getDeptByScope()
			expectEffectiveRuleSources(deptAfter.EffectiveRules, map[string]api.ScopeType{
				ruleKey(newClusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):       api.SCOPETYPE_DEPARTMENT,
			})
			expectBreakdownFields(deptAfter.Breakdown, newClusterRule.Field, deptRule.Field)
		})

		It("deleting a single rule from a cluster policy (DELETE /rules) propagates: dept+project effective and breakdown drop that field", func() {
			// give the cluster two rules so we can remove one and still keep the policy alive
			extraClusterRule := api.PolicyRule{
				Field:        "command",
				SpecSelector: "allContainers",
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}
			clusterPolicy := createGvkPolicyForTest(router, tenantId,
				newClusterPolicyReq([]api.PolicyRule{clusterRule, extraClusterRule}))
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			// prime caches via get-with-breakdown
			getGvkPolicyWithBreakdownForTest(router, tenantId, clusterPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, deptPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, projectPolicy.Id)
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeTrue())
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeTrue())

			// delete the extra rule from cluster
			doDelete(router, tenantId, gvkPolicyRulesURL(clusterPolicy.Id), api.DeleteGvkPolicyRulesReq{
				Rules: []api.PolicyRuleRef{
					{Field: extraClusterRule.Field, SpecSelector: extraClusterRule.SpecSelector},
				},
			})

			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeFalse(),
				"deleting a cluster rule must clear the dept cache")
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeFalse(),
				"deleting a cluster rule must clear the project cache")

			projectAfter := getProjectByScope()
			expectEffectiveRuleSources(projectAfter.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
			Expect(projectAfter.Breakdown).NotTo(BeNil())
			Expect(projectAfter.Breakdown.Fields).NotTo(HaveKey(extraClusterRule.Field),
				"breakdown of project should no longer include the deleted cluster rule's field")
			expectBreakdownFields(projectAfter.Breakdown,
				clusterRule.Field, deptRule.Field, projectRule.Field)
		})

		It("deleting the cluster policy entirely clears dept+project caches and refetch shows only dept+project rules", func() {
			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			// prime caches
			getGvkPolicyWithBreakdownForTest(router, tenantId, clusterPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, deptPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, projectPolicy.Id)
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeTrue())
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeTrue())

			doDelete(router, tenantId, gvkPolicyURL(clusterPolicy.Id), nil)

			Expect(cacheExistsForPolicy(ctxFor(tenantId), clusterPolicy.Id)).To(BeFalse())
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeFalse(),
				"deleting cluster must clear dept cache")
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeFalse(),
				"deleting cluster must clear project cache")

			projectAfter := getProjectByScope()
			expectEffectiveRuleSources(projectAfter.EffectiveRules, map[string]api.ScopeType{
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
			expectBreakdownFields(projectAfter.Breakdown, deptRule.Field, projectRule.Field)
			Expect(projectAfter.Breakdown.Fields).NotTo(HaveKey(clusterRule.Field),
				"deleted cluster rule's field must drop out of project breakdown")

			deptAfter := getDeptByScope()
			expectEffectiveRuleSources(deptAfter.EffectiveRules, map[string]api.ScopeType{
				ruleKey(deptRule): api.SCOPETYPE_DEPARTMENT,
			})
			expectBreakdownFields(deptAfter.Breakdown, deptRule.Field)
		})
	})

	// ── bottom-up: adding a leaf inherits ancestors, removing ancestors shrinks the leaf ──
	Context("bottom-up: descendant inherits, ancestors persist", func() {

		It("adding a project under existing cluster+dept gives the project the merged set without disturbing ancestor caches", func() {
			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))

			// prime ancestor caches so we can verify they are *not* cleared
			// by a downward-only change (project create).
			getGvkPolicyWithBreakdownForTest(router, tenantId, clusterPolicy.Id)
			getGvkPolicyWithBreakdownForTest(router, tenantId, deptPolicy.Id)
			Expect(cacheExistsForPolicy(ctxFor(tenantId), clusterPolicy.Id)).To(BeTrue())
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeTrue())

			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			// project carries all three rules with correct sources
			expectEffectiveRuleSources(projectPolicy.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})

			// breakdown follows on get-with-breakdown
			projectFetched := getGvkPolicyWithBreakdownForTest(router, tenantId, projectPolicy.Id)
			expectBreakdownFields(projectFetched.Breakdown,
				clusterRule.Field, deptRule.Field, projectRule.Field)

			// ancestors must not have been invalidated by a downward change
			Expect(cacheExistsForPolicy(ctxFor(tenantId), clusterPolicy.Id)).To(BeTrue(),
				"creating a project policy must not clear ancestor cluster cache")
			Expect(cacheExistsForPolicy(ctxFor(tenantId), deptPolicy.Id)).To(BeTrue(),
				"creating a project policy must not clear ancestor department cache")
		})

		It("deleting the department policy shrinks the project to cluster+project (effective and breakdown)", func() {
			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			// confirm full merge and prime cluster + project caches so we
			// can later assert (a) project cache was cleared, (b) cluster
			// cache was *not* cleared by the dept deletion.
			projectFetched := getGvkPolicyWithBreakdownForTest(router, tenantId, projectPolicy.Id)
			expectEffectiveRuleSources(projectFetched.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
			expectBreakdownFields(projectFetched.Breakdown,
				clusterRule.Field, deptRule.Field, projectRule.Field)

			getGvkPolicyWithBreakdownForTest(router, tenantId, clusterPolicy.Id)
			Expect(cacheExistsForPolicy(ctxFor(tenantId), clusterPolicy.Id)).To(BeTrue(),
				"sanity: cluster cache should be present after priming")
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeTrue(),
				"sanity: project cache should be present after priming")

			doDelete(router, tenantId, gvkPolicyURL(deptPolicy.Id), nil)

			// project cache must have been cleared by deleting an ancestor
			Expect(cacheExistsForPolicy(ctxFor(tenantId), projectPolicy.Id)).To(BeFalse(),
				"deleting department must clear project cache")

			projectAfter := getProjectByScope()
			expectEffectiveRuleSources(projectAfter.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
			expectBreakdownFields(projectAfter.Breakdown, clusterRule.Field, projectRule.Field)
			Expect(projectAfter.Breakdown.Fields).NotTo(HaveKey(deptRule.Field),
				"deleted dept rule's field must drop out of project breakdown")

			// cluster is unaffected by removal of one of its descendants
			Expect(cacheExistsForPolicy(ctxFor(tenantId), clusterPolicy.Id)).To(BeTrue(),
				"deleting department must not clear cluster cache")
		})

		It("deleting the cluster policy after the department is already gone leaves the project with only its own rule", func() {
			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))
			createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			doDelete(router, tenantId, gvkPolicyURL(deptPolicy.Id), nil)
			doDelete(router, tenantId, gvkPolicyURL(clusterPolicy.Id), nil)

			// after both ancestors are deleted, project carries only its own rule
			projectAfter := getProjectByScope()
			expectEffectiveRuleSources(projectAfter.EffectiveRules, map[string]api.ScopeType{
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
			expectBreakdownFields(projectAfter.Breakdown, projectRule.Field)
			Expect(projectAfter.Breakdown.Fields).NotTo(HaveKey(clusterRule.Field),
				"deleted cluster rule's field must drop out of project breakdown")
			Expect(projectAfter.Breakdown.Fields).NotTo(HaveKey(deptRule.Field),
				"deleted dept rule's field must drop out of project breakdown")
		})

		It("does not regenerate ancestor caches just because a descendant is read", func() {
			// the converse direction: getting/recomputing a descendant must
			// not magically repopulate an ancestor's cache (the ancestor
			// cache is populated only when that ancestor itself is read).
			//
			// setup: only a cluster policy exists. the project has no policy
			// of its own and the cluster cache has never been primed.
			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			Expect(cacheExistsForPolicy(ctxFor(tenantId), clusterPolicy.Id)).To(BeFalse(),
				"sanity: cluster cache must not exist before any get-with-breakdown is issued")

			// reading the descendant (project) triggers the project's own
			// recompute + breakdown caching, and writes a fresh effective
			// row for the project. but it must NOT touch the cluster's
			// breakdown cache - that one is populated only when the
			// cluster policy itself is read with breakdown=true.
			projectFetched := getProjectByScope()
			Expect(projectFetched.Rules).To(BeEmpty(),
				"project has no policy of its own, so .rules must be empty")
			expectEffectiveRuleSources(projectFetched.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
			})

			Expect(cacheExistsForPolicy(ctxFor(tenantId), clusterPolicy.Id)).To(BeFalse(),
				"reading the project must not lazily populate the ancestor cluster's breakdown cache")
		})
	})

	// ── override semantics: same rule identity at multiple scopes ──────────
	Context("override: same (field, specSelector, ...) at multiple scopes", func() {

		It("a rule defined at both cluster and project, with the same identity, is overridden by the project's version when the cluster opts in with allowOverride=true", func() {
			// cluster sets imagePullPolicy=Required:true and explicitly
			// allows descendants to override it (allowOverride=true); project
			// sets the SAME rule (same field + specSelector) but
			// Required:false. the merge must keep ONE effective entry for
			// that identity, taken from the project (descendant wins), and
			// the sourceOfRule must point to project. this is the in-place-
			// replace branch of MergeGvkRules - the only meaningful
			// uncovered branch of the multi-scope merge before this test.
			// note: since the default for allowOverride is false, the
			// cluster rule must explicitly set allowOverride=true for the
			// project's version to win.
			clusterVersion := api.PolicyRule{
				Field:         "imagePullPolicy",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(true),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}
			projectVersion := api.PolicyRule{
				Field:        "imagePullPolicy",
				SpecSelector: "allContainers",
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}},
			}

			createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterVersion}))
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectVersion}))

			// the project's own get-by-id (returned by create) effectivr has the project own rule
			// with allowOverride=nil
			Expect(projectPolicy.EffectiveRules).To(HaveLen(1),
				"override must produce exactly one effective rule per identity")
			er := projectPolicy.EffectiveRules[0]
			Expect(er.Field).To(Equal("imagePullPolicy"))
			Expect(er.SpecSelector).To(Equal("allContainers"))
			Expect(er.SourceOfRule).NotTo(BeNil())
			Expect(er.SourceOfRule.Scope).To(Equal(api.SCOPETYPE_PROJECT),
				"the descendant (project) must win the merge")
			Expect(er.Enforcements).To(HaveLen(1))
			Expect(er.Enforcements[0].Required).NotTo(BeNil())
			Expect(*er.Enforcements[0].Required).To(BeFalse(),
				"effective enforcement must be the project's value, not the cluster's")
			Expect(er.AllowOverride).To(HaveValue(BeFalse()),
				"project didn't specify allowOverride, so effective must show the default (false) - never nil")

			// and the same single-rule, project-sourced result is what
			// get-by-type-and-scope and get-effective return.
			projectByScope := getProjectByScope()
			expectEffectiveRuleSources(projectByScope.EffectiveRules, map[string]api.ScopeType{
				ruleKey(projectVersion): api.SCOPETYPE_PROJECT,
			})

			effOnly := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})
			expectEffectiveRuleSources(effOnly, map[string]api.ScopeType{
				ruleKey(projectVersion): api.SCOPETYPE_PROJECT,
			})
			Expect(effOnly[0].Enforcements[0].Required).NotTo(BeNil())
			Expect(*effOnly[0].Enforcements[0].Required).To(BeFalse(),
				"get-effective must also return the descendant's enforcement value")

			// the breakdown is also one entry: the rule appears once.
			expectBreakdownFields(projectByScope.Breakdown, "imagePullPolicy")
		})

		// department -> default value + allowOverride
		// project -> different default value + nil override
		It("a department default and a project set for the same identity: the project set replaces the department default in effective when the department opts in with allowOverride=true", func() {
			deptVersion := api.PolicyRule{
				Field:         "imagePullPolicy",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(true),
				Enforcements: []api.PolicyRuleEnforcement{
					{Default: map[string]interface{}{DefaultArgName: "IfNotPresent"}},
				},
			}
			projectVersion := api.PolicyRule{
				Field:        "imagePullPolicy",
				SpecSelector: "allContainers",
				Enforcements: []api.PolicyRuleEnforcement{
					{Set: map[string]interface{}{DefaultArgName: "Always"}},
				},
			}

			createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptVersion}))
			projectPolicy := createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectVersion}))

			// project's create response already carries the recomputed
			// effective: one rule, sourced from project, carrying the
			// project's set — the department's default has been dropped.
			Expect(projectPolicy.EffectiveRules).To(HaveLen(1),
				"override across scopes must produce exactly one effective rule per identity")
			er := projectPolicy.EffectiveRules[0]
			Expect(er.Field).To(Equal("imagePullPolicy"))
			Expect(er.SpecSelector).To(Equal("allContainers"))
			Expect(er.SourceOfRule).NotTo(BeNil())
			Expect(er.SourceOfRule.Scope).To(Equal(api.SCOPETYPE_PROJECT),
				"the project's set must win over the department's default")
			Expect(er.Enforcements).To(HaveLen(1))
			Expect(er.Enforcements[0].Set).NotTo(BeNil(),
				"effective enforcement must be the project's set")
			Expect(er.Enforcements[0].Default).To(BeNil(),
				"the department's default must not survive alongside the project's set for the same rule identity")
			Expect(er.Enforcements[0].Set[DefaultArgName]).To(Equal("Always"),
				"effective enforcement must carry the project's set value, not the department's default value")
			Expect(er.AllowOverride).To(HaveValue(BeFalse()),
				"project didn't specify allowOverride, so effective must show the default (false) - never nil")

			// get-by-type-and-scope and get-effective agree with the
			// create response.
			projectByScope := getProjectByScope()
			expectEffectiveRuleSources(projectByScope.EffectiveRules, map[string]api.ScopeType{
				ruleKey(projectVersion): api.SCOPETYPE_PROJECT,
			})
			Expect(projectByScope.EffectiveRules[0].Enforcements[0].Set).NotTo(BeNil())
			Expect(projectByScope.EffectiveRules[0].Enforcements[0].Set[DefaultArgName]).To(Equal("Always"))

			effOnly := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})
			expectEffectiveRuleSources(effOnly, map[string]api.ScopeType{
				ruleKey(projectVersion): api.SCOPETYPE_PROJECT,
			})
			Expect(effOnly[0].Enforcements[0].Set).NotTo(BeNil())
			Expect(effOnly[0].Enforcements[0].Set[DefaultArgName]).To(Equal("Always"),
				"get-effective must also return the project's set value")

			// and the department, viewed on its own, still carries the
			// default it contributed — the project's set only takes effect
			// at the project level (and below).
			deptByScope := getDeptByScope()
			expectEffectiveRuleSources(deptByScope.EffectiveRules, map[string]api.ScopeType{
				ruleKey(deptVersion): api.SCOPETYPE_DEPARTMENT,
			})
			Expect(deptByScope.EffectiveRules[0].Enforcements[0].Default).NotTo(BeNil(),
				"the department's own effective must still carry its default")
			Expect(deptByScope.EffectiveRules[0].Enforcements[0].Default[DefaultArgName]).To(Equal("IfNotPresent"))
			Expect(deptByScope.EffectiveRules[0].AllowOverride).To(HaveValue(BeTrue()),
				"department explicitly set allowOverride=true; effective must carry the same value")
		})
	})

	// ── allowOverride: parent rule controls whether descendants may replace it ──
	Context("allowOverride: parent gates whether a descendant may override", func() {

		// two rule identities shared across cluster + dept:
		//   ruleLocked: image / allContainers - cluster leaves allowOverride
		//               unset, which defaults to false (locked)
		//   ruleOpen:   imagePullPolicy / allContainers - cluster explicitly
		//               opts in with allowOverride=true
		//
		// "locked" means the cluster says "no descendant may override this
		// rule, my version wins for the whole subtree" - this is also the
		// default when allowOverride is unset. "open" means the cluster
		// explicitly lets descendants replace its version.
		ruleLockedAtCluster := api.PolicyRule{
			Field:        "image",
			SpecSelector: "allContainers",
			// no AllowOverride - defaults to false (locked)
			Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
		}
		ruleOpenAtCluster := api.PolicyRule{
			Field:         "imagePullPolicy",
			SpecSelector:  "allContainers",
			AllowOverride: convutils.Ref(true),
			Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
		}
		// the dept versions target the SAME two identities, but with the
		// opposite Required value so we can tell the source apart from
		// the enforcement value alone.
		ruleLockedAtDept := api.PolicyRule{
			Field:        "image",
			SpecSelector: "allContainers",
			Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}},
		}
		ruleOpenAtDept := api.PolicyRule{
			Field:        "imagePullPolicy",
			SpecSelector: "allContainers",
			Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}},
		}

		// findRuleByField picks the effective rule with the given field
		// out of a slice. used to assert the per-rule source / enforcement
		// without depending on slice order.
		findRuleByField := func(rules []api.EffectiveRulesEffectiveRulesInner, field string) *api.EffectiveRulesEffectiveRulesInner {
			GinkgoHelper()
			for i := range rules {
				if rules[i].Field == field {
					return &rules[i]
				}
			}
			Fail(fmt.Sprintf("no effective rule for field %q in %s", field, formatEffectiveRules(rules)))
			return nil
		}

		It("locked cluster rule stays at cluster for dept and project, open one is overridden by dept; flipping the lock to allow lets dept win for both", func() {
			// step 1: cluster declares both rules - one locked, one open.
			clusterPolicy := createGvkPolicyForTest(router, tenantId,
				newClusterPolicyReq([]api.PolicyRule{ruleLockedAtCluster, ruleOpenAtCluster}))

			// step 2: dept declares the same two identities, with opposite
			// Required values. expected merge for dept:
			//   - image (locked at cluster) -> cluster's version wins
			//   - imagePullPolicy (open)    -> dept's version wins
			createGvkPolicyForTest(router, tenantId,
				newDepartmentPolicyReq([]api.PolicyRule{ruleLockedAtDept, ruleOpenAtDept}))

			deptByScope := getDeptByScope()
			expectEffectiveRuleSources(deptByScope.EffectiveRules, map[string]api.ScopeType{
				ruleKey(ruleLockedAtDept): api.SCOPETYPE_CLUSTER,    // locked - cluster wins
				ruleKey(ruleOpenAtDept):   api.SCOPETYPE_DEPARTMENT, // open   - dept wins
			})

			lockedInDept := findRuleByField(deptByScope.EffectiveRules, "image")
			Expect(*lockedInDept.Enforcements[0].Required).To(BeTrue(),
				"locked rule must carry the cluster's enforcement value (Required=true)")
			openInDept := findRuleByField(deptByScope.EffectiveRules, "imagePullPolicy")
			Expect(*openInDept.Enforcements[0].Required).To(BeFalse(),
				"open rule must carry the dept's enforcement value (Required=false)")

			// step 3: project (no own policy) inherits the same picture as
			// dept - the merge cascades down with the dept-vs-cluster
			// decisions already baked in.
			projectByScope := getProjectByScope()
			expectEffectiveRuleSources(projectByScope.EffectiveRules, map[string]api.ScopeType{
				ruleKey(ruleLockedAtDept): api.SCOPETYPE_CLUSTER,
				ruleKey(ruleOpenAtDept):   api.SCOPETYPE_DEPARTMENT,
			})
			Expect(*findRuleByField(projectByScope.EffectiveRules, "image").Enforcements[0].Required).To(BeTrue())
			Expect(*findRuleByField(projectByScope.EffectiveRules, "imagePullPolicy").Enforcements[0].Required).To(BeFalse())

			// step 4: flip the lock on the cluster rule (allowOverride=true).
			// PUT replaces all the cluster's rules; we keep the same two
			// identities but the previously-locked one is now open.
			ruleLockedFlipped := api.PolicyRule{
				Field:         "image",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(true),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}
			doPut(router, tenantId, gvkPolicyURL(clusterPolicy.Id), api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{ruleLockedFlipped, ruleOpenAtCluster},
			})

			// step 5: get-effective for both dept and project: both rules
			// must now be sourced from dept, since neither cluster rule
			// blocks override anymore.
			deptEff := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: convutils.Ref(deptId),
			})
			expectEffectiveRuleSources(deptEff, map[string]api.ScopeType{
				ruleKey(ruleLockedAtDept): api.SCOPETYPE_DEPARTMENT,
				ruleKey(ruleOpenAtDept):   api.SCOPETYPE_DEPARTMENT,
			})
			Expect(*findRuleByField(deptEff, "image").Enforcements[0].Required).To(BeFalse(),
				"unlocked rule must now carry dept's Required=false")
			Expect(*findRuleByField(deptEff, "imagePullPolicy").Enforcements[0].Required).To(BeFalse())

			projectEff := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})
			expectEffectiveRuleSources(projectEff, map[string]api.ScopeType{
				ruleKey(ruleLockedAtDept): api.SCOPETYPE_DEPARTMENT,
				ruleKey(ruleOpenAtDept):   api.SCOPETYPE_DEPARTMENT,
			})
			Expect(*findRuleByField(projectEff, "image").Enforcements[0].Required).To(BeFalse())
			Expect(*findRuleByField(projectEff, "imagePullPolicy").Enforcements[0].Required).To(BeFalse())
		})

		// 3-level chain (cluster → dept → project) with the SAME rule
		// identity at every scope. The subtle invariant tested here is
		// that MergeGvkRules reads allowOverride off the currently-winning
		// rule at each merge step - so once dept overrides cluster,
		// dept's OWN allowOverride (not cluster's) governs whether
		// project can override in turn. This is precisely the case that
		// distinguishes the new default (unspecified = locked) from a
		// naive "top ancestor's allowOverride cascades everywhere"
		// reading of the semantics.
		It("3-level chain: intermediate scope's allowOverride gates deeper descendants; flipping levels changes the winner", func() {
			// state 0 setup:
			//   cluster.image (allowOverride=true,   Required=true)
			//   dept.image    (allowOverride unspec, Required=false)  -> defaults to locked
			//   project.image (allowOverride=true,   Required=true)   -> tries to override, but dept locks
			clusterRuleOpen := api.PolicyRule{
				Field:         "image",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(true),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}
			deptRuleUnspec := api.PolicyRule{
				Field:        "image",
				SpecSelector: "allContainers",
				// no AllowOverride - defaults to false (locked)
				Enforcements: []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}},
			}
			projectRuleOpen := api.PolicyRule{
				Field:         "image",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(true),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}

			clusterPolicy := createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRuleOpen}))
			deptPolicy := createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRuleUnspec}))
			createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRuleOpen}))

			// state 0 expectations:
			// dept-level effective: cluster is open, dept overrides -> DEPT
			// project-level effective: dept in accumulated effective has
			// no allowOverride (locked), so project's allowOverride=true
			// on its OWN rule is irrelevant - the merge reads the parent's
			// allowOverride. project's rule is dropped -> DEPT wins.
			deptEff0 := getDeptByScope().EffectiveRules
			expectEffectiveRuleSources(deptEff0, map[string]api.ScopeType{
				ruleKey(deptRuleUnspec): api.SCOPETYPE_DEPARTMENT,
			})
			Expect(*findRuleByField(deptEff0, "image").Enforcements[0].Required).To(BeFalse(),
				"dept must carry its own Required=false (dept won over cluster's Required=true)")

			projectEff0 := getProjectByScope().EffectiveRules
			expectEffectiveRuleSources(projectEff0, map[string]api.ScopeType{
				ruleKey(projectRuleOpen): api.SCOPETYPE_DEPARTMENT,
			})
			Expect(*findRuleByField(projectEff0, "image").Enforcements[0].Required).To(BeFalse(),
				"project's allowOverride=true on its own rule cannot override the dept parent (whose allowOverride is unspec/locked): dept's Required=false must survive")

			// state 1: PUT dept with allowOverride=true. now project's
			// override is unlocked and its rule wins at project level.
			// cluster is still open, so dept still wins over cluster at
			// dept level.
			deptRuleOpen := api.PolicyRule{
				Field:         "image",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(true),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}},
			}
			doPut(router, tenantId, gvkPolicyURL(deptPolicy.Id), api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{deptRuleOpen},
			})

			deptEff1 := getDeptByScope().EffectiveRules
			expectEffectiveRuleSources(deptEff1, map[string]api.ScopeType{
				ruleKey(deptRuleOpen): api.SCOPETYPE_DEPARTMENT,
			})
			Expect(*findRuleByField(deptEff1, "image").Enforcements[0].Required).To(BeFalse(),
				"cluster is still open, dept still wins at dept level")

			projectEff1 := getProjectByScope().EffectiveRules
			expectEffectiveRuleSources(projectEff1, map[string]api.ScopeType{
				ruleKey(projectRuleOpen): api.SCOPETYPE_PROJECT,
			})
			Expect(*findRuleByField(projectEff1, "image").Enforcements[0].Required).To(BeTrue(),
				"dept now allows override, so project's Required=true must win at project level")

			// state 2: PUT cluster with allowOverride explicitly false.
			// now dept cannot override cluster - cluster's rule survives
			// at dept level. and since the accumulated effective at dept
			// is cluster's rule (also locked), project cannot override
			// either. cluster wins throughout the whole subtree.
			clusterRuleLocked := api.PolicyRule{
				Field:         "image",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(false),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}
			doPut(router, tenantId, gvkPolicyURL(clusterPolicy.Id), api.UpdateGvkPolicyReq{
				Rules: []api.PolicyRule{clusterRuleLocked},
			})

			deptEff2 := getDeptByScope().EffectiveRules
			expectEffectiveRuleSources(deptEff2, map[string]api.ScopeType{
				ruleKey(deptRuleOpen): api.SCOPETYPE_CLUSTER,
			})
			Expect(*findRuleByField(deptEff2, "image").Enforcements[0].Required).To(BeTrue(),
				"cluster is now locked, so dept's Required=false is dropped and cluster's Required=true survives at dept level")

			projectEff2 := getProjectByScope().EffectiveRules
			expectEffectiveRuleSources(projectEff2, map[string]api.ScopeType{
				ruleKey(projectRuleOpen): api.SCOPETYPE_CLUSTER,
			})
			Expect(*findRuleByField(projectEff2, "image").Enforcements[0].Required).To(BeTrue(),
				"cluster is locked and its rule cascaded through dept's effective, so project's own allowOverride=true cannot help: cluster wins at project level too")
		})

		// Small pinning test: an explicit allowOverride=false on the parent
		// must behave identically to an unspecified allowOverride. Both
		// take the "keep parent" branch of MergeGvkRules but through
		// different sub-paths of the type-assertion on allowOverride
		// (explicit=false gives allowOverrideSpecified=true+allowOverride=false;
		// unspec gives allowOverrideSpecified=false), so it is worth
		// pinning this equivalence down explicitly.
		It("explicit allowOverride=false on the parent locks identically to leaving it unspecified", func() {
			clusterExplicitFalse := api.PolicyRule{
				Field:         "image",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(false),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(true)}},
			}
			deptAttemptOverride := api.PolicyRule{
				Field:         "image",
				SpecSelector:  "allContainers",
				AllowOverride: convutils.Ref(true),
				Enforcements:  []api.PolicyRuleEnforcement{{Required: convutils.Ref(false)}},
			}

			createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterExplicitFalse}))
			createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptAttemptOverride}))

			deptEff := getDeptByScope().EffectiveRules
			expectEffectiveRuleSources(deptEff, map[string]api.ScopeType{
				ruleKey(deptAttemptOverride): api.SCOPETYPE_CLUSTER,
			})
			Expect(*findRuleByField(deptEff, "image").Enforcements[0].Required).To(BeTrue(),
				"explicit allowOverride=false must lock the cluster's rule just like an unspecified one: cluster's Required=true survives even though dept tried to override")
		})
	})

	// ── descendant scope inherits even without a policy of its own ─────────
	Context("descendant scope with no policy of its own inherits ancestor rules", func() {

		It("get-by-type-and-scope of a project with only a department ancestor returns the dept's effective rules and matching breakdown", func() {
			// only a dept policy exists - the project under it has never
			// been created. by-type-and-scope must transparently materialise
			// a soft-deleted project placeholder, run the recompute, and
			// return an effective made of the dept's rules.
			createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))

			projectFetched := getProjectByScope()

			// no project policy of its own
			Expect(projectFetched.Rules).To(BeEmpty(),
				"project has no policy of its own, so .rules must be empty")

			expectEffectiveRuleSources(projectFetched.EffectiveRules, map[string]api.ScopeType{
				ruleKey(deptRule): api.SCOPETYPE_DEPARTMENT,
			})
			expectBreakdownFields(projectFetched.Breakdown, deptRule.Field)
		})

		It("get-by-type-and-scope of a project with only a cluster ancestor returns the cluster's effective rules", func() {
			// same scenario, one level higher: only a cluster policy exists,
			// no dept and no project. project recompute should still produce
			// the inherited cluster rule.
			createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))

			projectFetched := getProjectByScope()

			Expect(projectFetched.Rules).To(BeEmpty(),
				"project has no policy of its own, so .rules must be empty")
			expectEffectiveRuleSources(projectFetched.EffectiveRules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
			})
			expectBreakdownFields(projectFetched.Breakdown, clusterRule.Field)
		})
	})

	// ── get-effective-policy endpoint (POST /api/v1/policies/effective) ────
	Context("get-effective-policy endpoint", func() {

		It("returns the merged effective rules of all ancestor scopes for a project that has its own policy", func() {
			createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))
			createGvkPolicyForTest(router, tenantId, newProjectPolicyReq([]api.PolicyRule{projectRule}))

			rules := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})

			expectEffectiveRuleSources(rules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
				ruleKey(projectRule): api.SCOPETYPE_PROJECT,
			})
		})

		It("returns the inherited ancestor rules for a project that has no policy of its own", func() {
			// caller scenario: the frontend asks for the effective policy of
			// a project that has never had a policy created. the endpoint
			// must synthesize one transparently and return only the rules
			// that the ancestors contribute.
			createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))
			createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))

			rules := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})

			expectEffectiveRuleSources(rules, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
			})
		})

		It("returns an empty effective when no ancestor and no own policy exist", func() {
			// no policy at any level - the endpoint should still succeed
			// and just return an empty effective.
			rules := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})

			Expect(rules).To(BeEmpty(), "expected no effective rules when nothing is configured")
		})

		It("reflects ancestor changes on the next call (no stale cache returned)", func() {
			// the endpoint shares the recompute path with get-by-type-and-
			// scope, so adding an ancestor after a first call should show
			// up on the second call's returned effective.
			createGvkPolicyForTest(router, tenantId, newDepartmentPolicyReq([]api.PolicyRule{deptRule}))

			before := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})
			expectEffectiveRuleSources(before, map[string]api.ScopeType{
				ruleKey(deptRule): api.SCOPETYPE_DEPARTMENT,
			})

			createGvkPolicyForTest(router, tenantId, newClusterPolicyReq([]api.PolicyRule{clusterRule}))

			after := getEffectivePolicyForTest(router, tenantId, api.WorkloadTypeAndScope{
				WorkloadType: gvk,
				Scope:        api.SCOPETYPE_PROJECT,
				ProjectId:    convutils.Ref(projectId),
			})
			expectEffectiveRuleSources(after, map[string]api.ScopeType{
				ruleKey(clusterRule): api.SCOPETYPE_CLUSTER,
				ruleKey(deptRule):    api.SCOPETYPE_DEPARTMENT,
			})
		})
	})
})

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

// ruleKey builds the "<field>/<specSelector>" identity used by
// expectEffectiveRuleSources. It only needs to be deterministic per (field,
// specSelector) pair within a single test - the multi-scope tests do not
// exercise itemSelector / selectorParams variants.
func ruleKey(r api.PolicyRule) string {
	return fmt.Sprintf("%s/%s", r.Field, r.SpecSelector)
}

// expectEffectiveRuleSources asserts the effective-rules slice contains
// exactly one entry per key in `expected`, each entry's sourceOfRule.Scope
// matching the expected scope. The check is order-independent so the test
// does not depend on merge iteration order.
func expectEffectiveRuleSources(rules []api.EffectiveRulesEffectiveRulesInner, expected map[string]api.ScopeType) {
	GinkgoHelper()

	Expect(rules).To(HaveLen(len(expected)),
		"effective rule count mismatch (expected %d, got %d): %s",
		len(expected), len(rules), formatEffectiveRules(rules))

	seen := make(map[string]bool, len(rules))
	for _, er := range rules {
		key := fmt.Sprintf("%s/%s", er.Field, er.SpecSelector)
		wantScope, ok := expected[key]
		Expect(ok).To(BeTrue(),
			"unexpected effective rule %s in %s", key, formatEffectiveRules(rules))
		Expect(er.SourceOfRule).NotTo(BeNil(),
			"effective rule %s should carry a sourceOfRule", key)
		Expect(er.SourceOfRule.Scope).To(Equal(wantScope),
			"effective rule %s should come from scope %s but came from %s",
			key, wantScope, er.SourceOfRule.Scope)
		Expect(seen[key]).To(BeFalse(),
			"effective rule %s appears more than once", key)
		seen[key] = true
	}
}

// expectBreakdownFields asserts the breakdown contains exactly the listed
// field names (the keys of Breakdown.Fields). Used to verify that the
// breakdown re-computation after an ancestor change picks up / drops the
// right fields.
func expectBreakdownFields(bd *api.BreakdownBreakdown, fields ...string) {
	GinkgoHelper()

	Expect(bd).NotTo(BeNil(), "expected a breakdown but got nil")
	Expect(bd.Fields).To(HaveLen(len(fields)),
		"breakdown field count mismatch (expected %d, got %d, got fields: %v)",
		len(fields), len(bd.Fields), keys(bd.Fields))

	for _, f := range fields {
		Expect(bd.Fields).To(HaveKey(f), "breakdown missing field %q", f)
	}
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func formatEffectiveRules(rules []api.EffectiveRulesEffectiveRulesInner) string {
	out := "["
	for i, r := range rules {
		if i > 0 {
			out += ", "
		}
		src := "<nil>"
		if r.SourceOfRule != nil {
			src = string(r.SourceOfRule.Scope)
		}
		out += fmt.Sprintf("%s/%s (src=%s)", r.Field, r.SpecSelector, src)
	}
	return out + "]"
}

// doPut issues a PUT and asserts a 200 response.
func doPut(router chi.Router, tenantId int32, url string, body any) *httptest.ResponseRecorder {
	GinkgoHelper()

	req, err := http.NewRequest(http.MethodPut, url, makePayload(body))
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(tenantId)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
	Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())
	return resp
}

// doDelete issues a DELETE (with optional body for delete-rules) and asserts
// either 200 or 204 - both are valid success codes across the GVK endpoints
// we exercise.
func doDelete(router chi.Router, tenantId int32, url string, body any) *httptest.ResponseRecorder {
	GinkgoHelper()

	var req *http.Request
	var err error
	if body == nil {
		req, err = http.NewRequest(http.MethodDelete, url, nil)
	} else {
		req, err = http.NewRequest(http.MethodDelete, url, makePayload(body))
	}
	Expect(err).To(BeNil())
	req.Header = *headerForAdmin(tenantId)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req.WithContext(tenantContextForTest(tenantId)))
	Expect(resp.Code).To(BeElementOf(http.StatusOK, http.StatusNoContent), resp.Body.String())
	return resp
}

// ctxFor is a tiny re-name of tenantContextForTest so call-sites read as
// `cacheExistsForPolicy(ctxFor(tenantId), pid)` rather than re-stating the
// longer name everywhere a context is needed.
func ctxFor(tenantId int32) context.Context {
	return tenantContextForTest(tenantId)
}

// getEffectivePolicyForTest POSTs to /api/v1/policies/effective and decodes
// the response body as EffectiveRules. The endpoint returns the merged
// effective rules of the requested scope (including those inherited from
// ancestors), without the breakdown. Use this in tests that focus on the
// external "view effective rules" surface used by the frontend.
func getEffectivePolicyForTest(router chi.Router, tenantId int32, req api.WorkloadTypeAndScope) []api.EffectiveRulesEffectiveRulesInner {
	GinkgoHelper()

	httpReq, err := http.NewRequest(http.MethodPost, "/api/v1/policies/effective", makePayload(req))
	Expect(err).To(BeNil())
	httpReq.Header = *headerForAdmin(tenantId)

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httpReq.WithContext(tenantContextForTest(tenantId)))
	Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

	var result api.EffectiveRules
	Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed(), resp.Body.String())
	return result.EffectiveRules
}
