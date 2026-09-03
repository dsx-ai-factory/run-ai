/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("headers", func() {

	someone := convutils.Ref("someone")

	Describe("Headers validation", func() {
		Context("Validate header - ok", func() {
			It("should succeed", func() {
				r := createRequest(convutils.Ref("someone"), "1001")
				Expect(ValidateAuthenticationHeaders(r)).To(Equal(true))
			})
		})

		Context("Validate header - not ok", func() {
			It("should fail", func() {
				r := createRequest(someone, "")
				Expect(ValidateAuthenticationHeaders(r)).To(Equal(false))
			})
		})
	})

	Describe("Headers loading", func() {
		Context("Missing header", func() {
			It("should fail", func() {
				r := createRequest(nil, "")
				err := loadAuthenticationHeaders(r)
				Expect(err.Error()).To(ContainSubstring("missing subject id"))
			})
		})

		Context("Empty requester", func() {
			It("should fail", func() {
				r := createRequest(convutils.Ref(""), "")
				msg := loadAuthenticationHeaders(r)
				Expect(msg.Error()).To(ContainSubstring("subject-id header should not be empty"))
			})
		})
	})
})

func createRequest(requester *string, tenantIds string) *http.Request {
	r := http.Request{}
	ctx := context.Background()
	if requester != nil {
		ctx = context.WithValue(ctx, runaictx.CfSubjectID, *requester)
	}

	if tenantIds != "" {
		for _, tenantId := range strings.Split(tenantIds, ",") {
			ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantId)
		}
	}

	return r.WithContext(ctx)
}
