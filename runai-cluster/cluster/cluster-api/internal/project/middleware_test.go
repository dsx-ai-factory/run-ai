/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mocks "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/workload"
)

var _ = Describe("NamespaceMiddleware", func() {
	var (
		ctrl        *gomock.Controller
		translator  *mocks.MockProjectTranslator
		middleware  *NamespaceMiddleware
		request     *http.Request
		recorder    *httptest.ResponseRecorder
		projectName string
		namespace   string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		translator = mocks.NewMockProjectTranslator(ctrl)
		middleware = NewNamespaceMiddleware(translator)

		recorder = httptest.NewRecorder()
		projectName = "test-project"
		namespace = "test-namespace"
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("with valid project name", func() {
		BeforeEach(func() {
			router := chi.NewRouter()
			router.Route("/{projectName}", func(r chi.Router) {
				r.Use(middleware.Middleware)
				r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
					Expect(NamespaceFromContext(request.Context())).To(Equal(namespace))
					Expect(FromContext(request.Context())).To(Equal(projectName))
					writer.WriteHeader(http.StatusOK)
				})
			})

			var err error
			request, err = http.NewRequest(http.MethodGet, "/"+projectName, nil)
			Expect(err).NotTo(HaveOccurred())

			translator.EXPECT().Translate(gomock.Any(), projectName).Return(namespace, nil)
			router.ServeHTTP(recorder, request)
		})

		It("should pass the request to the next handler", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Context("with missing project name", func() {
		BeforeEach(func() {
			router := chi.NewRouter()
			router.Use(middleware.Middleware)
			router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(http.StatusOK)
			})

			var err error
			request, err = http.NewRequest(http.MethodGet, "/", nil)
			Expect(err).NotTo(HaveOccurred())

			router.ServeHTTP(recorder, request)
		})

		It("should return a bad request error", func() {
			Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			Expect(recorder.Body.String()).To(ContainSubstring("project name is required"))
		})
	})

	Context("with translation error", func() {
		BeforeEach(func() {
			router := chi.NewRouter()
			router.Route("/{projectName}", func(r chi.Router) {
				r.Use(middleware.Middleware)
				r.Get("/", func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(http.StatusOK)
				})
			})

			var err error
			request, err = http.NewRequest(http.MethodGet, "/"+projectName, nil)
			Expect(err).NotTo(HaveOccurred())

			translator.EXPECT().Translate(gomock.Any(), projectName).Return("", errors.New("translation error"))

			router.ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring(failedToTranslateErrorMsg))
		})
	})
})
