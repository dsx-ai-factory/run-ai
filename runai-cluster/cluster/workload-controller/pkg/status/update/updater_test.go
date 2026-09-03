/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update_test

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/status/update"
	fakeClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/convertors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name      = "test"
	namespace = "default"
)

var _ = Describe("WorkloadUpdater", func() {
	var (
		k8sClient       client.Client
		workloadType    constants.WorkloadType
		convertor       convertors.Convertor
		workloadUpdater *update.WorkloadUpdater
		jobWorkload     *jobs.JobWorkload
		clientObject    client.Object
	)
	Context("workload exists in the cluster", func() {

		BeforeEach(func() {

			scheme := runtime.NewScheme()
			register.InitSchema(scheme)
			workloadType = constants.InteractiveWorkload
			convertor = convertors.NewInteractiveWorkloadConvertor()
			jobWorkload = workload.NewJobWorkloadBuilder().WithInteractive().WithNamespace(namespace).WithName(name).Build()

			clientObject = convertor.Convert(jobWorkload)
			k8sClient = fakeClient.InitFakeClientWithObjectWithStatus(scheme, []client.Object{&v2alpha1.InteractiveWorkload{}}, clientObject)
			workloadUpdater = update.NewWorkloadUpdater(k8sClient, workloadType, convertor, update.NewRunaiWorkloadPatcher())
		})
		workloadUpdater = update.NewWorkloadUpdater(k8sClient, workloadType, convertor, update.NewRunaiWorkloadPatcher())
		Context("workload is nil", func() {
			It("should failed to update status", func() {
				err := workloadUpdater.UpdateStatus(context.Background(), nil)
				Expect(err).ToNot(BeNil())
			})
		})
		It("should successfully update status", func() {
			jobWorkload.Status.RunningPods = 1
			jobWorkload.Status.WorkloadCommonStatus.PhaseMessage = "update me"
			err := workloadUpdater.UpdateStatus(context.Background(), jobWorkload)
			Expect(err).To(BeNil())
			err = k8sClient.Get(context.Background(), client.ObjectKey{Namespace: namespace, Name: name}, clientObject)
			Expect(err).To(BeNil())
			Expect(clientObject).ToNot(BeNil())

			iw := clientObject.(*v2alpha1.InteractiveWorkload)
			Expect(iw.Status.RunningPods).To(Equal(int32(1)))
			Expect(iw.Status.WorkloadCommonStatus.PhaseMessage).To(Equal("update me"))

		})
	})
	Context("workload does not exist in the cluster", func() {
		BeforeEach(func() {

			scheme := runtime.NewScheme()
			register.InitSchema(scheme)
			workloadType = constants.InteractiveWorkload
			convertor = convertors.NewInteractiveWorkloadConvertor()
			jobWorkload = workload.NewJobWorkloadBuilder().WithInteractive().WithNamespace(namespace).WithName(name).Build()

			clientObject = convertor.Convert(jobWorkload)
			k8sClient = fakeClient.InitFakeClientWithObjectNoStatus(scheme)
			workloadUpdater = update.NewWorkloadUpdater(k8sClient, workloadType, convertor, update.NewRunaiWorkloadPatcher())
		})
		It("updates status failed", func() {
			err := workloadUpdater.UpdateStatus(context.Background(), jobWorkload)
			Expect(err).ToNot(BeNil())
		})
	})

})
