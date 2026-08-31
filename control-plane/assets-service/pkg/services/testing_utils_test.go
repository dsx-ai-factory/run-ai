/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	runaictx_http "github.com/run-ai/runai/runai-common-packages/runaictx/http"
	"net/http"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/testutils"

	. "github.com/onsi/gomega"
)

func headerWithMissingInfo(randoms *testutils.RandomValues) *http.Header {
	return &http.Header{}
}

func headerForAdmin(randoms *testutils.RandomValues) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	return runaictx_http.GetHeaders(ctx)
}

func headerForOtherAdmin(randoms *testutils.RandomValues) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, "admin-2")
	return runaictx_http.GetHeaders(ctx)
}

func headerForNonAdmin(randoms *testutils.RandomValues) *http.Header {
	ctx := context.WithValue(context.Background(), runaictx.CfSubjectID, "just@a.user")
	ctx = context.WithValue(ctx, runaictx.CfTenantId, randoms.TenantIdStr)
	return runaictx_http.GetHeaders(ctx)
}

// makePayload prepares a go object to be sent as a payload to REST call
func makePayload(payloadObject interface{}) *strings.Reader {
	payloadString, err := json.Marshal(payloadObject)
	Expect(err).To(BeNil())
	return strings.NewReader(string(payloadString))
}
