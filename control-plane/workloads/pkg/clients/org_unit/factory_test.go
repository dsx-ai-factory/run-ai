/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package org_unit

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	orgUnitMocks "github.com/run-ai/runai/runai-clients/org-unit-service-client/mocks"
	"github.com/run-ai/runai/runai-clients/org-unit-service-client/structs"
	cache "github.com/run-ai/runai/runai-common-packages/cache/v2"
	cacheconfig "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
)

func TestOrgUnitClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Org Unit Client Suite")
}

var _ = Describe("cacheClient.GetDepartmentName", Ordered, func() {
	ctx := context.Background()
	var mockClient *orgUnitMocks.IOrgUnitClient
	var client ClientInterface

	departmentKey := func(id string) structs.OrgUnitKey {
		return structs.OrgUnitKey{Scope: string(orgUnitApi.SCOPETYPE_DEPARTMENT), Id: id}
	}

	BeforeAll(func() {
		enabled := true
		Expect(cache.Init(&cacheconfig.Cache{
			Type:            "inmemory",
			Enabled:         &enabled,
			TtlSeconds:      3600,
			MaxStoreSizeMB:  100,
			MaxRecordSizeKB: 1,
			ServiceName:     "workloads-test",
		})).To(Succeed())
		Expect(cache.RegisterStore[departmentNameCacheValue](options.Ttl(time.Minute))).To(Succeed())
	})

	BeforeEach(func() {
		// Reset the cache between specs so each one starts from a cold cache.
		Expect(cache.ClearStore[departmentNameCacheValue]()).To(Succeed())
		mockClient = orgUnitMocks.NewIOrgUnitClient(GinkgoT())
		client = newCacheClient(mockClient)
	})

	It("fetches the org tree once and resolves multiple departments from that single fetch", func() {
		nameMap := map[structs.OrgUnitKey]string{
			departmentKey("dept-1"): "department-one",
			departmentKey("dept-2"): "department-two",
			// A non-department scope must be ignored.
			{Scope: string(orgUnitApi.SCOPETYPE_PROJECT), Id: "proj-1"}: "project-one",
		}
		// Exactly one org-tree fetch is allowed; a second call would fail this expectation.
		mockClient.On("GetOrgTreeMap", mock.Anything).Return(structs.OrgUnitMap{}, nameMap, (*orgUnitApi.Error)(nil)).Once()

		name, err := client.GetDepartmentName(ctx, "dept-1")
		Expect(err).To(BeNil())
		Expect(name).To(Equal("department-one"))

		// Served from the cache populated by the first fetch - no second org-unit call.
		name, err = client.GetDepartmentName(ctx, "dept-2")
		Expect(err).To(BeNil())
		Expect(name).To(Equal("department-two"))

		mockClient.AssertExpectations(GinkgoT())
	})

	It("returns a not-found error when the department is absent from the org tree", func() {
		nameMap := map[structs.OrgUnitKey]string{
			departmentKey("dept-1"): "department-one",
		}
		mockClient.On("GetOrgTreeMap", mock.Anything).Return(structs.OrgUnitMap{}, nameMap, (*orgUnitApi.Error)(nil)).Once()

		name, err := client.GetDepartmentName(ctx, "missing-dept")
		Expect(err).NotTo(BeNil())
		Expect(name).To(BeEmpty())
	})

	It("returns empty without calling the org-unit service for an empty department id", func() {
		name, err := client.GetDepartmentName(ctx, "")
		Expect(err).To(BeNil())
		Expect(name).To(BeEmpty())
		mockClient.AssertNotCalled(GinkgoT(), "GetOrgTreeMap", mock.Anything)
	})
})
