/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

var _ = Describe("ParseResourceClaimTemplateToServerScheme", func() {
	buildRct := func(name, namespace string, labels map[string]string) *resourcev1.ResourceClaimTemplate {
		return &resourcev1.ResourceClaimTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    labels,
			},
		}
	}

	It("defaults an unlabelled RCT to project scope with origin=existing", func() {
		parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(
			buildRct("rct-1", "runai-ofer", nil))

		Expect(err).ToNot(HaveOccurred())
		Expect(parsed.Scope.Type).To(Equal(auth_client.SCOPETYPE_PROJECT))
		// Project scope: parseScope pulls the id from the namespace. The
		// control-plane project id override is layered on by the handler,
		// not by the parser itself.
		Expect(parsed.Scope.Id).To(Equal("runai-ofer"))
		Expect(parsed.CreationRequest.Name).To(Equal("rct-1"))
		Expect(parsed.CreationRequest.Namespace).To(Equal("runai-ofer"))
		Expect(parsed.CreationRequest.ScopeType).To(Equal(trackerclient.ResourceScopeType(auth_client.SCOPETYPE_PROJECT)))
		Expect(parsed.CreationRequest.ScopeId).To(Equal("runai-ofer"))
		Expect(parsed.CreationRequest.Origin.IsSet()).To(BeTrue())
		Expect(*parsed.CreationRequest.Origin.Get()).To(Equal(trackerclient.ORIGIN_EXISTING))
		Expect(parsed.CreationRequest.AssetId.IsSet()).To(BeFalse())
	})

	It("rewrites a cluster-scope RCT's scopeId to the cluster uuid", func() {
		parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(buildRct("rct-c", "runai", map[string]string{
			commonconstants.ClusterWideResourceLabel: "true",
		}))

		Expect(err).ToNot(HaveOccurred())
		Expect(parsed.Scope.Type).To(Equal(auth_client.SCOPETYPE_CLUSTER))
		// The parser swaps the scopeId to ClusterUuid so the control-plane
		// (scopeId == clusterId) invariant holds without the handler having
		// to post-process cluster-scoped rows.
		Expect(parsed.CreationRequest.ScopeId).To(Equal(config.Config().ClusterUuid))
	})

	It("tenant-scope label routes to cluster scope with the cluster uuid", func() {
		parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(buildRct("rct-t", "runai", map[string]string{
			commonconstants.TenantScopeLabel: "true",
		}))

		Expect(err).ToNot(HaveOccurred())
		Expect(parsed.Scope.Type).To(Equal(auth_client.SCOPETYPE_CLUSTER))
		Expect(parsed.CreationRequest.ScopeId).To(Equal(config.Config().ClusterUuid))
	})

	It("maps a department-scope RCT onto the department id from the label", func() {
		parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(buildRct("rct-d", "runai", map[string]string{
			commonconstants.DepartmentScopeLabel: "4500003",
		}))

		Expect(err).ToNot(HaveOccurred())
		Expect(parsed.Scope.Type).To(Equal(auth_client.SCOPETYPE_DEPARTMENT))
		Expect(parsed.Scope.Id).To(Equal("4500003"))
		Expect(parsed.CreationRequest.ScopeId).To(Equal("4500003"))
	})

	It("marks the origin as assets and propagates the asset id when the RCT carries the asset label", func() {
		parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(buildRct("rct-a", "runai-ofer", map[string]string{
			commonconstants.RunaiAssetLabel:   "ResourceClaimTemplate",
			commonconstants.RunaiAssetIdLabel: "asset-123",
		}))

		Expect(err).ToNot(HaveOccurred())
		Expect(parsed.CreationRequest.Origin.IsSet()).To(BeTrue())
		Expect(*parsed.CreationRequest.Origin.Get()).To(Equal(trackerclient.ORIGIN_ASSETS))
		Expect(parsed.CreationRequest.AssetId.IsSet()).To(BeTrue())
		Expect(*parsed.CreationRequest.AssetId.Get()).To(Equal("asset-123"))
	})

	It("still marks origin=assets even when the asset id label is missing (never panics)", func() {
		parsed, err := parsers.ParseResourceClaimTemplateToServerScheme(buildRct("rct-a2", "runai-ofer", map[string]string{
			commonconstants.RunaiAssetLabel: "ResourceClaimTemplate",
		}))

		Expect(err).ToNot(HaveOccurred())
		Expect(*parsed.CreationRequest.Origin.Get()).To(Equal(trackerclient.ORIGIN_ASSETS))
		// AssetId stays Null because the ControlPlane row doesn't force it.
		Expect(parsed.CreationRequest.AssetId.IsSet()).To(BeFalse())
	})

	It("rejects cluster-wide + department labels at the same time (parseScope invariant)", func() {
		_, err := parsers.ParseResourceClaimTemplateToServerScheme(buildRct("rct-bad", "runai", map[string]string{
			commonconstants.ClusterWideResourceLabel: "true",
			commonconstants.DepartmentScopeLabel:     "4500003",
		}))

		Expect(err).To(HaveOccurred())
	})
})
