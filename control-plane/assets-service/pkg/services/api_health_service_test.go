/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Health API Service", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
		testing.NewRandomsBasedAuthMockClient(randoms)
	})

	Context("calling health endpoint", func() {
		It("should return 'I'm alive!'", func() {
			request, err := http.NewRequest("GET", "/", nil)
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			// Validate the response
			Expect(response.Code).To(Equal(http.StatusOK))
			Expect(response.Body.String()).To(ContainSubstring("I'm alive!"))
		})
	})
})
