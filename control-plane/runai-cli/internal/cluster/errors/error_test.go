/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterErrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/errors"
)

var _ = Describe("errors", func() {
	Describe("WorkloadNotReadyError", func() {
		It("creates a new WorkloadNotReadyError with the provided message", func() {
			message := "not ready"
			err := clusterErrors.NewNotReadyError(message)

			Expect(err).To(BeAssignableToTypeOf(&clusterErrors.WorkloadNotReadyError{}))
			Expect(err.Error()).To(ContainSubstring(message))
		})
	})

	Describe("IgnoreNotReadyError", func() {
		It("returns nil if the error is a WorkloadNotReadyError", func() {
			err := clusterErrors.NewNotReadyError("not ready")
			result := clusterErrors.IgnoreNotReadyError(err)

			Expect(result).To(BeNil())
		})

		It("returns the original error if it is not a WorkloadNotReadyError", func() {
			err := errors.New("some error")
			result := clusterErrors.IgnoreNotReadyError(err)

			Expect(result).To(Equal(err))
		})
	})

	Describe("HandleError", func() {
		var (
			response *http.Response
			err      error
		)

		BeforeEach(func() {
			response = nil
			err = nil
		})

		It("returns a generic error if the response is nil", func() {
			err = errors.New("some error")
			result := clusterErrors.HandleError(response, err)

			Expect(result).To(HaveOccurred())
			Expect(result.Error()).To(ContainSubstring(
				fmt.Errorf("got error while trying to stream workload: %w", err).Error()))
		})

		It("returns a WorkloadNotReadyError if the response status code is StatusPreconditionFailed", func() {
			errorMsg := "not ready"
			response = &http.Response{
				StatusCode:    http.StatusPreconditionFailed,
				Body:          io.NopCloser(strings.NewReader(errorMsg)),
				ContentLength: int64(len(errorMsg)),
			}
			result := clusterErrors.HandleError(response, err)

			Expect(result).To(HaveOccurred())
			Expect(result).To(BeAssignableToTypeOf(&clusterErrors.WorkloadNotReadyError{}))
			Expect(result.Error()).To(ContainSubstring(
				fmt.Sprintf("workload is not ready to stream: %v", errorMsg)))
		})

		It("returns a not found error if the response status code is StatusNotFound", func() {
			errorMsg := "not found"
			response = &http.Response{
				StatusCode:    http.StatusNotFound,
				Body:          io.NopCloser(strings.NewReader(errorMsg)),
				ContentLength: int64(len(errorMsg)),
			}
			result := clusterErrors.HandleError(response, err)

			Expect(result).To(HaveOccurred())
			Expect(result.Error()).To(ContainSubstring(fmt.Sprintf("one of requested resource was not found: %v", errorMsg)))
		})

		It("returns a generic error if the response status code is not handled", func() {
			errorMsg := "internal server error"
			response = &http.Response{
				StatusCode:    http.StatusInternalServerError,
				Body:          io.NopCloser(strings.NewReader(errorMsg)),
				ContentLength: int64(len(errorMsg)),
			}
			result := clusterErrors.HandleError(response, err)

			Expect(result).To(HaveOccurred())
			Expect(result.Error()).
				To(ContainSubstring(
					fmt.Errorf("got error while trying to stream workload: %w",
						fmt.Errorf("%s", errorMsg)).Error()))
		})
	})
})
