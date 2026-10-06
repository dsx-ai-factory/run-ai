/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

// calling the import here to ensure that the helper init takes place before anything else
// see inside the helper package for more info
import (
	_ "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/tests/helper"

	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("K8sObjectsHandler", func() {
	var handler *K8sObjectsHandler

	BeforeEach(func() {
		// namespaceFor does not touch any client or state on the handler,
		// so a zero-value handler is enough for this unit.
		handler = &K8sObjectsHandler{}
	})

	// metaWithProject builds an asset meta scoped to the given project id,
	// mirroring the one-value-at-a-time wrapping the generated OpenAPI
	// types force on nullable fields.
	metaWithProject := func(name string, projectID int32) assetsApi.AssetMetaForSync {
		return assetsApi.AssetMetaForSync{
			Scope:     assetsApi.SCOPE_PROJECT,
			Name:      name,
			Id:        "asset-id",
			Kind:      assetsApi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
			ProjectId: *assetsApi.NewNullableInt32(&projectID),
		}
	}

	metaWithScope := func(name string, scope assetsApi.Scope) assetsApi.AssetMetaForSync {
		return assetsApi.AssetMetaForSync{
			Scope: scope,
			Name:  name,
			Id:    "asset-id",
			Kind:  assetsApi.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
		}
	}

	Describe("namespaceFor", func() {
		It("routes a project-scoped asset to its project namespace", func() {
			update := &K8sObjectFromBackend{Meta: metaWithProject("my-rct", 42)}
			ns, err := handler.namespaceFor(update, map[string]string{"42": "runai-teama"})
			Expect(err).NotTo(HaveOccurred())
			Expect(ns).To(Equal("runai-teama"))
		})

		It("errors when a project-scoped asset is missing projectId", func() {
			update := &K8sObjectFromBackend{Meta: metaWithScope("no-project", assetsApi.SCOPE_PROJECT)}
			_, err := handler.namespaceFor(update, map[string]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing projectId"))
		})

		It("errors when the project id does not resolve to a namespace", func() {
			update := &K8sObjectFromBackend{Meta: metaWithProject("stranded", 99)}
			_, err := handler.namespaceFor(update, map[string]string{"42": "runai-teama"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("namespace not found"))
		})

		It("routes cluster-scoped assets to the runai namespace", func() {
			update := &K8sObjectFromBackend{Meta: metaWithScope("cluster-wide", assetsApi.SCOPE_CLUSTER)}
			ns, err := handler.namespaceFor(update, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(ns).To(Equal(constants.NamespaceRunAi))
		})

		It("routes department-scoped assets to the runai namespace", func() {
			update := &K8sObjectFromBackend{Meta: metaWithScope("dept-wide", assetsApi.SCOPE_DEPARTMENT)}
			ns, err := handler.namespaceFor(update, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(ns).To(Equal(constants.NamespaceRunAi))
		})

		It("explicitly rejects tenant-scoped k8s-objects", func() {
			update := &K8sObjectFromBackend{Meta: metaWithScope("tenant-wide", assetsApi.SCOPE_TENANT)}
			_, err := handler.namespaceFor(update, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("tenant-scoped k8s-objects are not supported"))
		})

		It("errors on an unknown scope rather than silently defaulting", func() {
			update := &K8sObjectFromBackend{Meta: metaWithScope("weird", assetsApi.Scope("galactic"))}
			_, err := handler.namespaceFor(update, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown scope"))
		})
	})
})
