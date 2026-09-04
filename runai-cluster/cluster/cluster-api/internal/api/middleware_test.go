/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Middleware", func() {
	var (
		middleware *Middleware
		request    *http.Request
		recorder   *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		middleware = NewMiddleware()
		recorder = httptest.NewRecorder()
	})

	Context("with valid workload type, name and framework", func() {
		BeforeEach(func() {
			router := chi.NewRouter()
			router.Route(ClusterAPIBaseUrl, func(r chi.Router) {
				r.Use(middleware.Middleware)

				r.Get("/port-forward", func(writer http.ResponseWriter, request *http.Request) {
					Expect(WorkloadTypeFromContext(request.Context())).To(Equal(Training))
					Expect(WorkloadNameFromContext(request.Context())).To(Equal("workload1"))
					Expect(WorkloadActionFromContext(request.Context())).To(Equal(PortForward))
					Expect(WorkloadFrameworkFromContext(request.Context())).To(Equal(RunaiFramework))
					writer.WriteHeader(http.StatusOK)
				})
			})

			var err error
			request, err = http.NewRequest(http.MethodGet, "/cluster-api/api/v1/test-proj/workloads/training/runai/workload1/port-forward", nil)
			Expect(err).NotTo(HaveOccurred())

			router.ServeHTTP(recorder, request)
		})

		It("should pass the request to the next handler", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Context("with invalid workload action", func() {
		BeforeEach(func() {
			router := chi.NewRouter()
			router.Route(ClusterAPIBaseUrl, func(r chi.Router) {
				r.Use(middleware.Middleware)
				r.Get("/endpoint", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			})

			var err error
			request, err = http.NewRequest(http.MethodGet, "/cluster-api/api/v1/test-proj/workloads/training/runai/workload1/endpoint", nil)
			Expect(err).NotTo(HaveOccurred())

			router.ServeHTTP(recorder, request)
		})

		It("should return a bad request error", func() {
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring("unsupported endpoint: endpoint"))
		})
	})

	Context("with invalid values", func() {
		var err error
		var router chi.Router

		BeforeEach(func() {
			router = chi.NewRouter()
			router.Route(ClusterAPIBaseUrl, func(r chi.Router) {
				r.Use(middleware.Middleware)

				r.Get("/port-forward", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			})
		})

		It("invalid type", func() {
			request, err = http.NewRequest(http.MethodGet, "/cluster-api/api/v1/test-proj/workloads/unknown/runai/workload1/port-forward", nil)
			Expect(err).NotTo(HaveOccurred())

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring("invalid workload type: unknown"))
		})

		It("invalid framework", func() {
			request, err = http.NewRequest(http.MethodGet, "/cluster-api/api/v1/test-proj/workloads/distributed/unknown/workload1/port-forward", nil)
			Expect(err).NotTo(HaveOccurred())

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring("invalid workload framework: unknown"))
		})
	})

	Context("with valid external workload type, name and framework", func() {
		var router *chi.Mux

		BeforeEach(func() {
			router = chi.NewRouter()
			router.Route(ClusterAPIBaseUrl, func(r chi.Router) {
				r.Use(middleware.Middleware)
				r.Get("/logs/download", func(writer http.ResponseWriter, request *http.Request) {
					Expect(WorkloadTypeFromContext(request.Context())).To(Equal(External))
					Expect(WorkloadNameFromContext(request.Context())).To(Equal("workload1"))
					Expect(WorkloadActionFromContext(request.Context())).To(Equal(LogsDownload))
					Expect(WorkloadFrameworkFromContext(request.Context())).To(Equal(ExternalFramework))
					writer.WriteHeader(http.StatusOK)
				})
			})
		})

		It("should pass the request to the next handler", func() {
			request, err := http.NewRequest(http.MethodGet, "/cluster-api/api/v1/test-proj/workloads/external/external/workload1/logs/download", nil)
			Expect(err).NotTo(HaveOccurred())

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should fail the request not forward to the next handler", func() {
			request, err := http.NewRequest(http.MethodGet, "/cluster-api/api/v1/test-proj/workloads/external/deployment/workload1/logs/download", nil)
			Expect(err).NotTo(HaveOccurred())

			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
