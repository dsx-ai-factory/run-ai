/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package finalizers_test

import (
	"context"

	controller_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/controller-utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/finalizers"
	fakeClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	testWorkload constants.WorkloadType = "test-workload"
	namespace                           = "test-namespace"
	name                                = "test-name"
)

var _ = Describe("WorkloadRemoveFinalizer", func() {
	var (
		k8sClient    client.Client
		workloadType constants.WorkloadType
		convertor    mocks.Convertor
		wrf          *finalizers.WorkloadRemoveFinalizer
		job          *jobs.JobWorkload
		workload     *v2alpha1.InteractiveWorkload
	)

	BeforeEach(func() {
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)
		workloadType = testWorkload
		workload = &v2alpha1.InteractiveWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		controller_utils.AddFinalizer(workload, constants.FinalizerName)

		k8sClient = fakeClient.InitFakeClientWithObjectNoStatus(scheme, workload)
		convertor = mocks.Convertor{}
		wrf = finalizers.NewWorkloadRemoveFinalizer(k8sClient, workloadType, &convertor)
		job = &jobs.JobWorkload{}
	})

	Context("RemoveFinalizer", func() {
		Context("workload exits in cluster", func() {
			BeforeEach(func() {
			})
		})
		It("should remove finalizer successfully", func() {
			convertor.On("Convert", job).Return(workload)
			err := wrf.RemoveFinalizer(context.Background(), job)
			Expect(err).To(BeNil())
		})

		It("should fail to remove finalizer", func() {
			workload.ResourceVersion = "wrong-version"
			convertor.On("Convert", job).Return(workload)
			err := wrf.RemoveFinalizer(context.Background(), job)
			Expect(err).ToNot(BeNil())
		})
	})
})
