/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
)

var _ = Describe("error_utils", func() {
	Context("error raised", func() {
		It("should return the correct response", func() {
			response := NewErrorResponse(api.NewError(http.StatusBadRequest, "incorrect request", nil))
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.(api.Error).Message).To(ContainSubstring("incorrect request"))
		})
	})
})
