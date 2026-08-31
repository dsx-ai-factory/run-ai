/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
)

var _ = Describe("API Errors", func() {
	Describe("NotFoundError", func() {
		It("should return the correct error message", func() {
			err := api.NewNotFoundError("test error")
			Expect(err.Error()).To(Equal("not found"))
		})
	})

	Describe("NotReadyError", func() {
		It("should return the correct error message", func() {
			err := api.NewNotReadyError("test error")
			Expect(err.Error()).To(Equal("test error"))
		})
	})

	Describe("ToHttpError", func() {
		Context("when the error is a NotFoundError", func() {
			It("should return the correct HTTP status and message", func() {
				err := api.NewNotFoundError("test error")
				msg, status := api.ToHttpError(err)
				Expect(msg).To(Equal("not found"))
				Expect(status).To(Equal(http.StatusNotFound))
			})
		})

		Context("when the error is a NotReadyError", func() {
			It("should return the correct HTTP status and message", func() {
				err := api.NewNotReadyError("test error")
				msg, status := api.ToHttpError(err)
				Expect(msg).To(Equal("test error"))
				Expect(status).To(Equal(http.StatusPreconditionFailed))
			})
		})

		Context("when the error is a different type", func() {
			It("should return the correct HTTP status and message", func() {
				err := errors.New("test error")
				msg, status := api.ToHttpError(err)
				Expect(msg).To(Equal("test error"))
				Expect(status).To(Equal(http.StatusServiceUnavailable))
			})
		})
	})
})
