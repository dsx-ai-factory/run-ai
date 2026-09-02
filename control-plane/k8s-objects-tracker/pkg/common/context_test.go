/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

var _ = Describe("context test", func() {
	Context("no tenant id in context", func() {
		It("should be rejected", func() {
			_, err := TenantIdFromContext(context.Background())
			Expect(err).NotTo(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusForbidden))
		})
	})

	Context("malformed tenant id in context", func() {
		It("should be rejected", func() {
			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "abb")
			_, err := TenantIdFromContext(ctx)
			Expect(err).NotTo(BeNil())
			Expect(int(err.Code)).To(Equal(http.StatusForbidden))
		})
	})
})
