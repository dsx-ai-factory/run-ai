/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/authorization"
	orgUnit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dbtesting"
)

var _ = Describe("health test", func() {
	var randomScopes *RandomScopes

	BeforeEach(func() {
		randomScopes = NewRandomsSet()

		backendMock := NewRandomsBasedBackendMockClient(randomScopes)
		orgUnit.SetMock(backendMock)
		authorization.SetMock(backendMock)

		router = CreateRouter()

		dbtesting.CleanTables()
	})

	Context("calling health", func() {
		It("should return I am alive", func() {
			createHttpReq, err := http.NewRequest("GET", "/", nil)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("I'm alive"))
		})
	})
})
