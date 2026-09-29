/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clierrors_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/ui"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Error Tests")
}

var _ = Describe("CLI Errors Test", func() {
	Context("Create HTTP error", func() {
		It("should create new HTTP error", func() {
			err := newHttpError("my status", "my message")
			Expect(err).NotTo(BeNil())
			Expect(err.Status).To(Equal("my status"))
			Expect(err.Message).To(Equal("my message"))
			Expect(err.Error()).To(Equal("status: my status, reason: my message"))
		})

		It("should unwrap HTTP error", func() {
			err := newHttpError("my status", "my message")
			Expect(err).NotTo(BeNil())
			Expect(err.Unwrap().Error()).To(Equal("status: my status, reason: my message"))
		})

		It("should handle response error", func() {
			err := clierrors.HandleResponseError(nil, nil)
			Expect(err).To(BeNil())
			err = clierrors.HandleResponseError(nil, errors.New("my error"))
			Expect(err).To(Equal(errors.New("my error")))
		})

		It("should handle response error", func() {
			response := &http.Response{
				StatusCode: 777,
				Body:       io.NopCloser(strings.NewReader("body message")),
			}
			orgUnitError := openapi.NewErrorFromResponse(response, fmt.Errorf("actual error"))
			err := clierrors.HandleResponseError(response, orgUnitError)
			var apiError *openapi.Error
			ok := errors.As(err, &apiError)
			Expect(ok).To(BeTrue())
			Expect(apiError).ToNot(BeNil())
			Expect(apiError.Code).To(Equal(int32(777)))
			Expect(apiError.Message).To(Equal("actual error"))
			Expect(apiError.Error()).To(Equal("HTTP 777: actual error"))
		})
	})
	Context("Config set errors Test", func() {
		invalidCpUri := "invalidCp"
		invalidAuthUri := "invalidAuth"
		expectedErrorMsg := "Invalid URI format. Provide a valid web address (e.g., https://www.example.com)"
		It("should return enriched value", func() {
			_, err := url.ParseRequestURI(invalidCpUri)
			mappedError := clierrors.UserEnrichment(err)
			Expect(mappedError.Error()).To(ContainSubstring(expectedErrorMsg))
		})
		It("should return the original error", func() {
			randomErr := fmt.Errorf("RandomErr")
			mappedError := clierrors.UserEnrichment(randomErr)
			Expect(mappedError).To(Equal(randomErr))
		})
		It("should check multi error format", func() {
			decisions := map[string]string{
				fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionCPUrl, ""):   invalidCpUri,
				fmt.Sprintf(consts.SelectionInputWithVisibleDefault, consts.SelectionAuthUrl, ""): invalidAuthUri,
			}
			decisionsErrors, _ := ui.CheckDecision(&decisions)
			Expect(strings.Count(decisionsErrors, expectedErrorMsg)).To(Equal(2))
			Expect(decisionsErrors).To(SatisfyAll(
				ContainSubstring(invalidCpUri),
				ContainSubstring(invalidAuthUri),
			))
		})
	})
})

func newHttpError(status, message string) *clierrors.HttpError {
	return &clierrors.HttpError{
		Status:  status,
		Message: strings.Trim(message, "\""),
	}
}
