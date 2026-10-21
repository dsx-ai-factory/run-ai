/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	policy_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("PoliciesHandler", func() {
	Describe("addPolicyLabels", func() {
		var (
			handler *PoliciesHandler
			labels  map[string]string
			meta    *metav1.ObjectMeta
		)

		BeforeEach(func() {
			handler = &PoliciesHandler{}
			labels = map[string]string{}
			meta = &metav1.ObjectMeta{Name: "policy", Labels: labels}
		})

		// A policy whose scope carries no id cannot be labelled. The fields are
		// nullable in the policy API, so this is a real input rather than an
		// impossible one, and it must be skipped instead of crashing the agent.
		DescribeTable("skips a scoped policy without its scope id",
			func(scope policy_openapi.ScopeType) {
				Expect(func() {
					handler.addPolicyLabels(meta, &policy_openapi.PolicyCreationFields{Name: "policy", Scope: scope})
				}).NotTo(Panic())
				Expect(labels).To(BeEmpty(), "no scope label is added without the id")
			},
			Entry("project", policy_openapi.SCOPETYPE_PROJECT),
			Entry("department", policy_openapi.SCOPETYPE_DEPARTMENT),
		)

		It("labels a project policy with its project id", func() {
			handler.addPolicyLabels(meta, &policy_openapi.PolicyCreationFields{
				Name:      "policy",
				Scope:     policy_openapi.SCOPETYPE_PROJECT,
				ProjectId: *policy_openapi.NewNullableInt32(ptr.To(int32(7))),
			})
			Expect(labels).To(HaveKeyWithValue(constants.ProjectScopeLabel, "7"))
		})

		It("labels a department policy with its department id", func() {
			handler.addPolicyLabels(meta, &policy_openapi.PolicyCreationFields{
				Name:         "policy",
				Scope:        policy_openapi.SCOPETYPE_DEPARTMENT,
				DepartmentId: *policy_openapi.NewNullableString(ptr.To("dept-1")),
			})
			Expect(labels).To(HaveKeyWithValue(constants.DepartmentScopeLabel, "dept-1"))
		})
	})

	Describe("backendToClusterWorkspacePolicies", func() {
		var handler *PoliciesHandler

		BeforeEach(func() {
			// isUnstructured keeps getObject on the fake client instead of the
			// typed createEmptyObject path, which needs a registered GVK.
			handler = &PoliciesHandler{CommonHandler: &CommonHandler{cachedClient: fakeClientWithObjects(), isUnstructured: true}}
		})

		// The counter is optional in the policy API, so a policy that arrives
		// without one must be skipped: dereferencing it used to take the agent
		// down, and the agent has no recover.
		It("skips a policy without an update counter", func() {
			policies := []policy_openapi.PolicySyncResponsePoliciesInner{{
				Meta: policy_openapi.PolicyCreationFields{Name: "policy", Scope: policy_openapi.SCOPETYPE_CLUSTER},
			}}

			var (
				updates clusterWorkloadsAndPoliciesUpdates
				err     error
			)
			Expect(func() {
				updates, err = handler.backendToClusterWorkspacePolicies(policies, nil)
			}).NotTo(Panic(), "a missing counter must not be dereferenced")
			Expect(err).NotTo(HaveOccurred())
			Expect(updates.ToDelete).To(BeEmpty())
			Expect(updates.ToUpdate).To(BeEmpty())
		})

		// Only composition reads the counter, so a deletion that arrives without
		// one must still be carried out instead of leaving the policy behind.
		It("deletes a policy without an update counter", func() {
			policies := []policy_openapi.PolicySyncResponsePoliciesInner{{
				Meta:      policy_openapi.PolicyCreationFields{Name: "policy", Scope: policy_openapi.SCOPETYPE_CLUSTER},
				IsDeleted: *policy_openapi.NewNullableBool(ptr.To(true)),
			}}

			updates, err := handler.backendToClusterWorkspacePolicies(policies, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(updates.ToDelete).To(HaveLen(1))
			Expect(updates.ToDelete[0].Name).To(Equal("policy"))
			Expect(updates.DeleteRef).To(Equal([]int{0}))
			Expect(updates.ToUpdate).To(BeEmpty())
		})

		// The guard must not skip a policy that does carry a counter.
		It("deletes a policy that carries an update counter", func() {
			policies := []policy_openapi.PolicySyncResponsePoliciesInner{{
				Meta:          policy_openapi.PolicyCreationFields{Name: "policy", Scope: policy_openapi.SCOPETYPE_CLUSTER},
				IsDeleted:     *policy_openapi.NewNullableBool(ptr.To(true)),
				UpdateCounter: ptr.To(int64(3)),
			}}

			updates, err := handler.backendToClusterWorkspacePolicies(policies, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(updates.ToDelete).To(HaveLen(1))
			Expect(updates.ToDelete[0].Name).To(Equal("policy"))
			Expect(updates.DeleteRef).To(Equal([]int{0}))
		})
	})
})
