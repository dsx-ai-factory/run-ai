/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jobs

import (
	mpi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/jobs"
	client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	kubeCommonv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	v1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const testJobKind = "TestJob"

var _ = Describe("DistributedJobAdapter", func() {

	var (
		adapter   *DistributedJobAdapter
		k8sClient client.Client
	)

	BeforeEach(func() {
		adapter = &DistributedJobAdapter{
			Object:  &corev1.Pod{},
			jobKind: testJobKind,
			spec:    &DistributedSpec{},
		}
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)
		k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
	})

	It("returns correct job kind", func() {
		Expect(adapter.GetJobKind()).To(Equal(testJobKind))
	})

	It("returns correct spec", func() {
		Expect(adapter.GetSpec()).To(Equal(adapter.spec))
	})

	It("returns correct concurrent job", func() {
		Expect(adapter.GetConcurrentJob()).To(Equal(adapter.Object))
	})

	Context("NewPytorchJobAdapter", func() {
		It("should create a new PytorchJobAdapter without error", func() {
			pytorchJob := &v1.PyTorchJob{}

			adapter, err := NewPytorchJobAdapter(pytorchJob, k8sClient)
			Expect(adapter.status.getManualPodLabelSelectorFn).NotTo(BeNil())
			Expect(adapter.status.getManualPodLabelSelectorFn).To(BeAssignableToTypeOf(getTrainingJobManualPodLabelSelector))

			Expect(err).NotTo(HaveOccurred())
			Expect(adapter).NotTo(BeNil())
		})
	})

	Context("NewTFJobAdapter", func() {
		It("should create a new TFJobAdapter without error", func() {
			tfJob := &v1.TFJob{}

			adapter, err := NewTFJobAdapter(tfJob, k8sClient)
			Expect(adapter.status.getManualPodLabelSelectorFn).NotTo(BeNil())
			Expect(adapter.status.getManualPodLabelSelectorFn).To(BeAssignableToTypeOf(getTrainingJobManualPodLabelSelector))
			Expect(err).NotTo(HaveOccurred())
			Expect(adapter).NotTo(BeNil())
		})
	})

	Context("NewXGBoostJobAdapter", func() {
		It("should create a new XGBoostJobAdapter without error", func() {
			xgboostJob := &v1.XGBoostJob{}

			adapter, err := NewXGBoostJobAdapter(xgboostJob, k8sClient)
			Expect(adapter.status.getManualPodLabelSelectorFn).NotTo(BeNil())
			Expect(adapter.status.getManualPodLabelSelectorFn).To(BeAssignableToTypeOf(getTrainingJobManualPodLabelSelector))
			Expect(err).NotTo(HaveOccurred())
			Expect(adapter).NotTo(BeNil())
		})
	})
	Context("NewMPIJobAdapter", func() {
		It("should create a new MPIJobAdapter without error", func() {
			mpiJob := &mpi.MPIJob{}

			adapter, _ := NewMPIJobAdapter(mpiJob, k8sClient)
			Expect(adapter.status.getManualPodLabelSelectorFn).NotTo(BeNil())
			Expect(adapter.status.getManualPodLabelSelectorFn).To(BeAssignableToTypeOf(getMpiManualPodLabelSelector))
		})
	})
})

var _ = Describe("DistributedSpec", func() {
	var (
		spec *DistributedSpec
	)

	BeforeEach(func() {
		spec = &DistributedSpec{
			replicaTypeSpec: map[constants.ReplicaType]ReplicaTypeSpec{
				constants.Worker: &DistributedReplicaSpec{},
			},
		}
	})

	It("returns correct replica type specs", func() {
		Expect(spec.GetReplicaTypeSpecs()).To(Equal(spec.replicaTypeSpec))
	})

	It("returns correct replica spec", func() {
		Expect(spec.GetReplicaSpec(constants.Worker)).To(Equal(spec.replicaTypeSpec[constants.Worker]))
	})
})

var _ = Describe("DistributedReplicaSpec", func() {
	var (
		spec *DistributedReplicaSpec
	)

	BeforeEach(func() {
		spec = &DistributedReplicaSpec{
			internalReplicaTypeSpec: &kubeCommonv1.ReplicaSpec{
				Template: corev1.PodTemplateSpec{},
				Replicas: ptr.To[int32](3),
			},
		}
	})

	It("returns correct pod template spec", func() {
		Expect(spec.GetPodTemplateSpec()).To(Equal(&spec.internalReplicaTypeSpec.Template))
	})

	It("returns correct replicas", func() {
		Expect(spec.GetReplicas()).To(Equal(spec.internalReplicaTypeSpec.Replicas))
	})
})
