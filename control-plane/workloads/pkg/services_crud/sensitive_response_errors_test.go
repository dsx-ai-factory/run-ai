/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sensitive response errors", func() {
	It("redacts submitted token strings from public errors", func() {
		details := "invalid value synthetic-hf-value"
		responseError := workloadsApi.SubmissionErrorV2{
			Code:    400,
			Message: "failed for synthetic-hf-value",
			Details: &details,
		}

		redacted := redactSensitiveValuesFromResponse(&responseError, []string{"synthetic-hf-value"})

		Expect(redacted.Message).To(Equal("failed for [REDACTED]"))
		Expect(*redacted.Details).To(Equal("invalid value [REDACTED]"))
	})

	DescribeTable("redacts JSON-escaped token strings from public errors",
		func(value string) {
			details := "invalid value " + value
			responseError := workloadsApi.SubmissionErrorV2{
				Code:    400,
				Message: "failed for " + value,
				Details: &details,
			}

			redacted := redactSensitiveValuesFromResponse(&responseError, []string{value})

			Expect(redacted.Message).To(Equal("failed for [REDACTED]"))
			Expect(*redacted.Details).To(Equal("invalid value [REDACTED]"))
		},
		Entry("quoted values", `synthetic-"quoted"-value`),
		Entry("HTML-sensitive values", "synthetic-<html>-value"),
	)
})
