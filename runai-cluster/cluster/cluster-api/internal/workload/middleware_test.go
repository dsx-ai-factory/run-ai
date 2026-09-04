/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/workload"
	projectApi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/project"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("ExistMiddleware", func() {
	const (
		TestNamespace = "test-namespace"
		TestWorkload  = "test-workload"
		TestProject   = "test-project"
	)

	var (
		ctrl        *gomock.Controller
		locator     *workload.MockLocator
		middleware  *ExistMiddleware
		handler     http.Handler
		request     *http.Request
		recorder    *httptest.ResponseRecorder
		ctx         context.Context
		workloadObj client.Object
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		locator = workload.NewMockLocator(ctrl)
		middleware = NewExistMiddleware(locator)
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		recorder = httptest.NewRecorder()
		ctx = context.Background()
		workloadObj = &twv2.TrainingWorkload{}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("with valid workload existence", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadType(ctx, api.Training)
			ctx = api.WithWorkloadName(ctx, TestWorkload)
			ctx = projectApi.WithProject(ctx, TestProject)
			ctx = api.WithWorkloadFramework(ctx, api.RunaiFramework)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			locator.EXPECT().Locate(gomock.Any(), api.Training, TestWorkload, TestNamespace).Return(workloadObj, nil)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should pass the request to the next handler", func() {
			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should set the workload in the context", func() {
			workload := FromContext(request.Context())
			Expect(workload).NotTo(Equal(nil))
		})
	})

	Context("with missing namespace", func() {
		BeforeEach(func() {
			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring("namespace is required, but not in context"))
		})
	})

	Context("with missing workload type", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadName(ctx, TestWorkload)
			ctx = projectApi.WithProject(ctx, TestProject)
			ctx = api.WithWorkloadFramework(ctx, api.RunaiFramework)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring("workload type is required, but not in context"))
		})
	})

	Context("with missing workload framework", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadName(ctx, TestWorkload)
			ctx = projectApi.WithProject(ctx, TestProject)
			ctx = api.WithWorkloadType(ctx, api.Training)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring("workload framework is required, but not in context"))
		})
	})

	Context("with missing workload name", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadType(ctx, api.Training)
			ctx = projectApi.WithProject(ctx, TestProject)
			ctx = api.WithWorkloadFramework(ctx, api.RunaiFramework)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring("workload name is required, but not in context"))
		})
	})

	Context("with missing project", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadType(ctx, api.Training)
			ctx = api.WithWorkloadName(ctx, TestWorkload)
			ctx = api.WithWorkloadFramework(ctx, api.RunaiFramework)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring("project is required, but not in context"))
		})
	})

	Context("with workload not found", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadType(ctx, api.Training)
			ctx = api.WithWorkloadName(ctx, TestWorkload)
			ctx = projectApi.WithProject(ctx, TestProject)
			ctx = api.WithWorkloadFramework(ctx, api.RunaiFramework)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			locator.EXPECT().Locate(gomock.Any(), api.Training, TestWorkload, TestNamespace).Return(nil,
				apierrors.NewNotFound(schema.GroupResource{}, TestWorkload))

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return a not found error", func() {
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			Expect(recorder.Body.String()).To(ContainSubstring("workload not found"))
		})
	})

	Context("with workload not found - framework mismatch", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadType(ctx, api.Distributed)
			ctx = api.WithWorkloadName(ctx, TestWorkload)
			ctx = projectApi.WithProject(ctx, TestProject)
			ctx = api.WithWorkloadFramework(ctx, api.MPIFramework)

			pytorch := &dwv2.DistributedWorkload{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: dp.DistributedSpec{
					DistributedFields: field_groups.DistributedFields{JobType: constants.KindXGBoostJob},
				},
				MasterSpec: nil,
				Status:     dwv2.DistributedWorkloadStatus{},
			}

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			locator.EXPECT().Locate(gomock.Any(), api.Distributed, TestWorkload, TestNamespace).Return(pytorch, nil)

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return a not found error", func() {
			Expect(recorder.Code).To(Equal(http.StatusNotFound))
			Expect(recorder.Body.String()).To(ContainSubstring(fmt.Sprintf(FrameworkMismatchMsg, api.MPIFramework.String())))
		})
	})

	Context("with locator error", func() {
		BeforeEach(func() {
			ctx = projectApi.WithNamespace(ctx, TestNamespace)
			ctx = api.WithWorkloadType(ctx, api.Training)
			ctx = api.WithWorkloadName(ctx, TestWorkload)
			ctx = projectApi.WithProject(ctx, TestProject)
			ctx = api.WithWorkloadFramework(ctx, api.RunaiFramework)

			request, _ = http.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(ctx)

			locator.EXPECT().Locate(gomock.Any(), api.Training, TestWorkload, TestNamespace).Return(nil, errors.New("locator error"))

			middleware.Middleware(handler).ServeHTTP(recorder, request)
		})

		It("should return an internal server error", func() {
			Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			Expect(recorder.Body.String()).To(ContainSubstring("locator error"))
		})
	})

	DescribeTable("validateWorkloadTypeAndFrameworkMatch",
		func(workloadType api.WorkloadType, workloadFramework api.WorkloadFramework, matching bool) {
			err := validateWorkloadTypeAndFrameworkMatch(workloadType, workloadFramework)
			Expect(err == nil).To(Equal(matching))
		},
		Entry("", api.Workspace, api.RunaiFramework, true),
		Entry("", api.Workspace, api.MPIFramework, false),
		Entry("", api.Workspace, api.PyTorchFramework, false),
		Entry("", api.Workspace, api.TFFramework, false),
		Entry("", api.Workspace, api.XGBoostFramework, false),
		Entry("", api.Workspace, api.AnyDistributedFramework, false),
		Entry("", api.Workspace, api.JAXFramework, false),
		Entry("", api.Workspace, api.ExternalFramework, false),

		Entry("", api.Inference, api.KnativeFramework, true),
		Entry("", api.Inference, api.RunaiFramework, false),
		Entry("", api.Inference, api.MPIFramework, false),
		Entry("", api.Inference, api.PyTorchFramework, false),
		Entry("", api.Inference, api.TFFramework, false),
		Entry("", api.Inference, api.XGBoostFramework, false),
		Entry("", api.Inference, api.AnyDistributedFramework, false),
		Entry("", api.Inference, api.JAXFramework, false),
		Entry("", api.Inference, api.ExternalFramework, false),

		Entry("", api.Training, api.RunaiFramework, true),
		Entry("", api.Training, api.MPIFramework, true),
		Entry("", api.Training, api.PyTorchFramework, true),
		Entry("", api.Training, api.TFFramework, true),
		Entry("", api.Training, api.XGBoostFramework, true),
		Entry("", api.Training, api.JAXFramework, true),
		Entry("", api.Training, api.AnyDistributedFramework, true),
		Entry("", api.Training, api.ExternalFramework, false),
		Entry("", api.Training, api.KnativeFramework, false),

		Entry("", api.Distributed, api.RunaiFramework, false),
		Entry("", api.Distributed, api.MPIFramework, true),
		Entry("", api.Distributed, api.PyTorchFramework, true),
		Entry("", api.Distributed, api.TFFramework, true),
		Entry("", api.Distributed, api.XGBoostFramework, true),
		Entry("", api.Distributed, api.JAXFramework, true),
		Entry("", api.Distributed, api.AnyDistributedFramework, true),
		Entry("", api.Distributed, api.ExternalFramework, false),
		Entry("", api.Distributed, api.KnativeFramework, false),

		Entry("", api.External, api.KnativeFramework, false),
		Entry("", api.External, api.RunaiFramework, false),
		Entry("", api.External, api.MPIFramework, false),
		Entry("", api.External, api.PyTorchFramework, false),
		Entry("", api.External, api.TFFramework, false),
		Entry("", api.External, api.XGBoostFramework, false),
		Entry("", api.External, api.AnyDistributedFramework, false),
		Entry("", api.External, api.JAXFramework, false),
		Entry("", api.External, api.ExternalFramework, true),
	)

	Context("validateLocatedFramework", func() {
		It("any framework - mismatch", func() {
			w := &Workload{JobType: api.RunaiJob}
			err := validateLocatedFramework(w, api.AnyDistributedFramework)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkMismatchMsg, api.AnyDistributedFramework)))
		})

		It("any framework - match", func() {
			w := &Workload{JobType: api.MPIJob}
			err := validateLocatedFramework(w, api.AnyDistributedFramework)
			Expect(err).NotTo(HaveOccurred())
		})

		It("mismatch", func() {
			w := &Workload{JobType: api.RunaiJob}
			err := validateLocatedFramework(w, api.TFFramework)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkMismatchMsg, api.TFFramework)))
		})

		It("match", func() {
			w := &Workload{JobType: api.MPIJob}
			err := validateLocatedFramework(w, api.MPIFramework)
			Expect(err).NotTo(HaveOccurred())
		})

		It("match", func() {
			w := &Workload{JobType: api.InferenceJob}
			err := validateLocatedFramework(w, api.KnativeFramework)
			Expect(err).NotTo(HaveOccurred())
		})

		It("match", func() {
			w := &Workload{JobType: api.ExternalJob}
			err := validateLocatedFramework(w, api.ExternalFramework)
			Expect(err).NotTo(HaveOccurred())
		})

		It("mismatch", func() {
			w := &Workload{JobType: api.RunaiJob}
			err := validateLocatedFramework(w, api.ExternalFramework)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(fmt.Sprintf(FrameworkMismatchMsg, api.ExternalFrameworkName)))
		})
	})
})
