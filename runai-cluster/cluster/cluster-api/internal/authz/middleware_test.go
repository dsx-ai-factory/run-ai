/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authz_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/authz"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/authz/authorizers"
	authz "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/authz/authorizers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Authz Middleware", func() {
	var (
		ctrl                 *gomock.Controller
		mockAuth             *authz.MockAuthorizer
		middleware           *Middleware
		handler              http.Handler
		request              *http.Request
		recorder             *httptest.ResponseRecorder
		ctx                  context.Context
		testWorkload         *workload.Workload
		testExternalWorkload *workload.Workload
		k8sLocator           *workload.K8sLocator
		mockClient           client.Client
		scheme               = runtime.NewScheme()
	)

	utilruntime.Must(corev1.AddToScheme(scheme))
	utilruntime.Must(twv2.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

	BeforeEach(func() {
		testWorkload = &workload.Workload{
			Type:      api.Training,
			Name:      "test-workload",
			Namespace: "test-namespace",
			Project:   "test-project",
		}
		testExternalWorkload = &workload.Workload{
			Type:      api.External,
			Name:      "test-external-workload",
			Namespace: "test-namespace",
			Project:   "test-project",
		}
		mockClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(&twv2.TrainingWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testWorkload.Name,
				Namespace: testWorkload.Namespace,
			},
		}, &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testExternalWorkload.Name,
				Namespace: testExternalWorkload.Namespace,
			},
		}).Build()
		k8sLocator = workload.NewK8sLocator(mockClient)

		ctrl = gomock.NewController(GinkgoT())
		mockAuth = authz.NewMockAuthorizer(ctrl)
		middleware = NewMiddleware(mockAuth, k8sLocator)
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		recorder = httptest.NewRecorder()
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("with valid request", func() {
		BeforeEach(func() {
			ctx = project.WithNamespace(ctx, testWorkload.Namespace)
			ctx = project.WithProject(ctx, testWorkload.Project)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{})
			ctx = api.WithWorkloadAction(ctx, api.PortForward)
			ctx = api.WithWorkloadName(ctx, testWorkload.Name)
			ctx = api.WithWorkloadType(ctx, api.Training)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			mockAuth.EXPECT().Authorize(gomock.Any(), gomock.Any()).Return(true, authorizers.ReasonAuthorized, nil)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should pass the request to the next handler", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Context("external workload with valid request", func() {
		BeforeEach(func() {
			ctx = project.WithNamespace(ctx, testExternalWorkload.Namespace)
			ctx = project.WithProject(ctx, testExternalWorkload.Project)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{})
			ctx = api.WithWorkloadAction(ctx, api.LogsDownload)
			ctx = api.WithWorkloadName(ctx, testExternalWorkload.Name)
			ctx = api.WithWorkloadType(ctx, api.External)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			mockAuth.EXPECT().Authorize(gomock.Any(), gomock.Any()).Return(true, authorizers.ReasonAuthorized, nil)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should pass the request to the next handler", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Context("with invalid request", func() {
		BeforeEach(func() {
			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring("namespace is required, but not found in context"))
		})
	})

	Context("with authorization error", func() {
		BeforeEach(func() {
			ctx = project.WithNamespace(ctx, testWorkload.Namespace)
			ctx = auth.ComposeAuthContext(ctx, &auth.Info{})
			ctx = api.WithWorkloadAction(ctx, api.PortForward)
			ctx = api.WithWorkloadName(ctx, testWorkload.Name)
			ctx = api.WithWorkloadType(ctx, api.Training)
			ctx = project.WithProject(ctx, testWorkload.Project)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			mockAuth.EXPECT().Authorize(gomock.Any(), gomock.Any()).Return(false, authorizers.ReasonUnauthorized, NewUnauthorizedError("authorization failed"))

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return a forbidden error", func() {
			Expect(recorder.Code).To(Equal(http.StatusForbidden))
		})
	})
})
