/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/google/uuid"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	authorization_config "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cacheTesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"

	"k8s.io/utils/ptr"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)

	cacheTesting.EnableCache()
	RunSpecs(t, "Cache tests")
	t.Cleanup(func() {
		cache.ClearCache()
		cache.Close()
		config.FlushAll()
	})
}

var (
	ProjectId       = "1"
	DepartmentId    = "2"
	ClusterId       = "2dasdsa2-tr23-54rf-12424sfr"
	TenantId        = "3"
	NodepoolId      = "15"
	WorkloadId      = "2dasdsa2-tr23-54rf-12424sf1"
	DatavolumesId   = uuid.New().String()
	ClusterConfigId = "2dasdsa2-tr23-54rf-12424sfr2"
)

var _ = Describe("Org hierarchy cache tests", Ordered, func() {
	ctx := context.Background()
	authorization_config.Init()
	var orgHierarchyManager *OrgHierarchyManager

	var mockServer, datavolumesMockServer *httptest.Server
	testAncestorsOrg := &authorization_utils.Ancestors{
		TenantId:     TenantId,
		DepartmentId: &DepartmentId,
		ProjectId:    &ProjectId,
	}
	testAncestorsCompute := &authorization_utils.Ancestors{
		TenantId:    TenantId,
		ClusterUuid: &ClusterId,
	}
	testAncestorsClusterConfig := &authorization_utils.Ancestors{
		TenantId:    TenantId,
		ClusterUuid: &ClusterConfigId,
	}
	testAncestorsBoth := &authorization_utils.Ancestors{
		TenantId:     TenantId,
		ClusterUuid:  &ClusterId,
		DepartmentId: &DepartmentId,
		ProjectId:    &ProjectId,
	}
	testAncestorsDatavolumes := &authorization_utils.Ancestors{
		TenantId:     TenantId,
		ClusterUuid:  ptr.To(uuid.New().String()),
		DepartmentId: &DepartmentId,
		ProjectId:    &ProjectId,
	}

	BeforeAll(func() {
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var jsonAncestors []byte
			var ancestors *authorization_utils.Ancestors
			var resourceId string
			var resourceType api.ResourceType
			var err error
			if strings.Contains(r.URL.Path, "projects") {
				resourceType = api.RESOURCETYPE_PROJECT
				resourceId = ProjectId
				jsonAncestors, err = json.Marshal(testAncestorsOrg)
				ancestors = testAncestorsOrg
			}
			if strings.Contains(r.URL.Path, "departments") {
				resourceType = api.RESOURCETYPE_DEPARTMENT
				resourceId = DepartmentId
				jsonAncestors, err = json.Marshal(testAncestorsOrg)
				ancestors = testAncestorsOrg
			}
			if strings.Contains(r.URL.Path, "node-pools") {
				resourceType = api.RESOURCETYPE_NODEPOOLS
				resourceId = NodepoolId
				jsonAncestors, err = json.Marshal(testAncestorsCompute)
				ancestors = testAncestorsCompute
			}
			if strings.Contains(r.URL.Path, "clusters") {
				resourceType = api.RESOURCETYPE_CLUSTER_CONFIG
				resourceId = ClusterConfigId
				jsonAncestors, err = json.Marshal(testAncestorsClusterConfig)
				ancestors = testAncestorsClusterConfig
			}
			if strings.Contains(r.URL.Path, "workloads") {
				resourceType = api.RESOURCETYPE_WORKLOADS
				resourceId = WorkloadId
				jsonAncestors, err = json.Marshal(testAncestorsBoth)
				ancestors = testAncestorsBoth
			}
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			err = cache.Set[authorization_utils.Ancestors](ctx, authorization_utils.GetKey(resourceId, string(resourceType)), *ancestors)
			Expect(err).To(BeNil())
			w.Write(jsonAncestors)
		}))
		serverConfig := config.Get[*authorization_config.Server]()
		serverConfig.OrgUnitServiceUrl = mockServer.URL
		serverConfig.WorkloadsUrl = mockServer.URL
		serverConfig.ClusterServiceUrl = mockServer.URL

		datavolumesMockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var jsonAncestors []byte
			var ancestors *authorization_utils.Ancestors
			var resourceId string
			var resourceType api.ResourceType
			var err error
			// "sharedscopes" should be "converted" to "datavolumes" prior to the internal api call
			if strings.Contains(r.URL.Path, "datavolumes") && !strings.Contains(r.URL.Path, "sharedscopes") {
				resourceType = api.RESOURCETYPE_DATAVOLUMES
				resourceId = DatavolumesId
				jsonAncestors, err = json.Marshal(testAncestorsDatavolumes)
				ancestors = testAncestorsDatavolumes
			}
			Expect(err).To(BeNil())

			w.Header().Set("Content-Type", "application/json")
			err = cache.Set[authorization_utils.Ancestors](ctx, authorization_utils.GetKey(resourceId, string(resourceType)), *ancestors)
			Expect(err).To(BeNil())
			w.Write(jsonAncestors)
		}))
		serverConfig = config.Get[*authorization_config.Server]()
		serverConfig.DatavolumesServiceUrl = datavolumesMockServer.URL
		config.Set(serverConfig)

		orgHierarchyManager = GetOrgHierarchyManager()
	})

	BeforeEach(func() {
		cache.ClearCache()
	})

	AfterEach(func() {
		cache.ClearCache()
	})

	AfterAll(func() {
		if mockServer != nil {
			mockServer.Close()
		}
	})

	GetAncestorsTest := func(resourceId *string, resourceType api.ResourceType, expectedAncestors *authorization_utils.Ancestors) {
		ancestors, cacheHit, err := orgHierarchyManager.GetAncestors(ctx, *resourceId, resourceType)
		Expect(err).To(BeNil())
		Expect(cacheHit).To(BeFalse())
		Expect(ancestors).To(Equal(expectedAncestors))

		ancestors, cacheHit, err = orgHierarchyManager.GetAncestors(ctx, *resourceId, resourceType)
		Expect(err).To(BeNil())
		Expect(cacheHit).To(BeTrue())
		Expect(ancestors).To(Equal(expectedAncestors))
	}

	DescribeTable("Get ancestors", GetAncestorsTest,
		Entry("Project", &ProjectId, api.RESOURCETYPE_PROJECT, testAncestorsOrg),
		Entry("Department", &DepartmentId, api.RESOURCETYPE_DEPARTMENT, testAncestorsOrg),
		Entry("Nodepools", &NodepoolId, api.RESOURCETYPE_NODEPOOLS, testAncestorsCompute),
		Entry("ClusterConfig", &ClusterConfigId, api.RESOURCETYPE_CLUSTER_CONFIG, testAncestorsClusterConfig),
		Entry("Workloads", &WorkloadId, api.RESOURCETYPE_WORKLOADS, testAncestorsBoth),
		Entry("Datavolumes", &DatavolumesId, api.RESOURCETYPE_DATAVOLUMES, testAncestorsDatavolumes),
		Entry("Datavolumes", &DatavolumesId, api.RESOURCETYPE_DATAVOLUMES, testAncestorsDatavolumes),
	)

	It("Disable cache", func() {
		cacheTesting.DisableCache()

		ancestors, cacheHit, err := orgHierarchyManager.GetAncestors(ctx, ProjectId, api.RESOURCETYPE_PROJECT)
		Expect(err).To(BeNil())
		Expect(cacheHit).To(BeFalse())
		Expect(ancestors).To(Equal(testAncestorsOrg))

		ancestors, cacheHit, err = orgHierarchyManager.GetAncestors(ctx, ProjectId, api.RESOURCETYPE_PROJECT)
		Expect(err).To(BeNil())
		Expect(cacheHit).To(BeFalse())
		Expect(ancestors).To(Equal(testAncestorsOrg))

		cacheTesting.EnableCache()
	})
})
