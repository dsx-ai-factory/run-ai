/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvlutils

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"
	"strconv"
	"strings"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	testclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("MNNVL Status", func() {
	const (
		supportedPool1   = "supported-pool-1"
		supportedPool2   = "supported-pool-2"
		unsupportedPool1 = "un-supported-pool-1"
		unsupportedPool2 = "un-supported-pool-2"
	)

	var (
		ctx       context.Context
		workload  *jobs.JobWorkload
		k8sClient k8sclient.Client
	)
	scheme := runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		workload = &jobs.JobWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindDistributedWorkload,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "w1",
				Labels: make(map[string]string),
			},
		}

		ctx = context.Background()

		// Each fixture is written to both CRDs with identical content, so these specs
		// hold under either value of the compile-time UseKaiNodePools flag.
		// TODO: drop the legacy fixtures once clusterconstants.UseKaiNodePools=true
		existingNodepools := []k8sclient.Object{
			getTestNodepool(supportedPool1, true),
			getTestNodepool(supportedPool2, true),
			getTestNodepool(unsupportedPool1, false),
			getTestNodepool(unsupportedPool2, false),
			getTestKaiNodepool(supportedPool1, true),
			getTestKaiNodepool(supportedPool2, true),
			getTestKaiNodepool(unsupportedPool1, false),
			getTestKaiNodepool(unsupportedPool2, false),
		}

		k8sClient = testclient.InitFakeClientWithObjectWithStatus(scheme,
			[]k8sclient.Object{&runv1.NodePool{}, &kaiv1alpha1.NodePool{}}, existingNodepools...)
	})

	Context("feature disabled", func() {
		It("no feature label", func() {
			status, err := GenerateStatus(ctx, k8sClient, workload)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.IsNetworkAccelerated()).To(BeFalse())
		})

		It("feature label is None", func() {
			workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.None

			status, err := GenerateStatus(ctx, k8sClient, workload)
			Expect(err).NotTo(HaveOccurred())
			Expect(status.IsNetworkAccelerated()).To(BeFalse())
		})
	})

	Context("nodepools validation", func() {
		Context("Required", func() {
			BeforeEach(func() {
				workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Required
				workload.Spec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
			})

			It("supported", func() {
				workload.Spec.NodePools = field_types.NewOptionalStringField(supportedPool1, "")

				status, err := GenerateStatus(ctx, k8sClient, workload)
				Expect(err).NotTo(HaveOccurred())
				Expect(status.IsNetworkAccelerated()).To(BeTrue())

				Expect(status.WorkerMode.IsEnabled()).To(BeTrue())
				Expect(status.MasterMode.IsEnabled()).To(BeFalse())
			})

			It("unsupported", func() {
				workload.Spec.NodePools = field_types.NewOptionalStringField(strings.Join([]string{unsupportedPool1, supportedPool1, unsupportedPool2}, " "), "")

				status, err := GenerateStatus(ctx, k8sClient, workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("some of the provided node pools do not support MNNVL capability: '%s' do not support %s in %s mode",
					strings.Join([]string{unsupportedPool1, unsupportedPool2}, ","), mnnvl.UserFacingFeatureName, mnnvl.Required)))

				Expect(status.IsNetworkAccelerated()).To(BeFalse())
			})
		})

		Context("Preferred", func() {
			BeforeEach(func() {
				workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Preferred
				workload.Spec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
			})

			It("supported", func() {
				workload.Spec.NodePools = field_types.NewOptionalStringField(supportedPool1, "")

				status, err := GenerateStatus(ctx, k8sClient, workload)
				Expect(err).NotTo(HaveOccurred())
				Expect(status.IsNetworkAccelerated()).To(BeTrue())

				Expect(status.WorkerMode.IsEnabled()).To(BeTrue())
				Expect(status.MasterMode.IsEnabled()).To(BeFalse())
			})

			It("unsupported", func() {
				workload.Spec.NodePools = field_types.NewOptionalStringField(unsupportedPool1, "")

				status, err := GenerateStatus(ctx, k8sClient, workload)
				Expect(err).NotTo(HaveOccurred())
				Expect(status.IsNetworkAccelerated()).To(BeFalse())
			})
		})
	})

	Context("specs validation", func() {
		Context("Required", func() {
			BeforeEach(func() {
				workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Required
				workload.Spec.NodePools = field_types.NewOptionalStringField(supportedPool1, "")
			})

			Context("no master", func() {
				It("supported", func() {
					workload.Spec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))

					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).NotTo(HaveOccurred())
					Expect(status.IsNetworkAccelerated()).To(BeTrue())

					Expect(status.WorkerMode.IsEnabled()).To(BeTrue())
					Expect(status.MasterMode.IsEnabled()).To(BeFalse())
				})

				It("unsupported", func() {
					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(SpecInvalidForRequiredErr))

					Expect(status.IsNetworkAccelerated()).To(BeFalse())
				})
			})

			Context("with master", func() {
				BeforeEach(func() {
					workload.MasterSpec = &jobs.JobSpec{}
				})

				It("supported", func() {
					workload.MasterSpec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))

					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).NotTo(HaveOccurred())
					Expect(status.IsNetworkAccelerated()).To(BeTrue())

					Expect(status.WorkerMode.IsEnabled()).To(BeFalse())
					Expect(status.MasterMode.IsEnabled()).To(BeTrue())
				})

				It("unsupported", func() {
					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(SpecInvalidForRequiredErr))

					Expect(status.IsNetworkAccelerated()).To(BeFalse())
				})
			})
		})

		Context("Preferred", func() {
			BeforeEach(func() {
				workload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Preferred
				workload.Spec.NodePools = field_types.NewOptionalStringField(supportedPool1, "")
			})

			Context("no master", func() {
				It("supported", func() {
					workload.Spec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))

					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).NotTo(HaveOccurred())
					Expect(status.IsNetworkAccelerated()).To(BeTrue())

					Expect(status.WorkerMode.IsEnabled()).To(BeTrue())
					Expect(status.MasterMode.IsEnabled()).To(BeFalse())
				})

				It("unsupported", func() {
					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).NotTo(HaveOccurred())
					Expect(status.IsNetworkAccelerated()).To(BeFalse())
				})
			})

			Context("with master", func() {
				BeforeEach(func() {
					workload.MasterSpec = &jobs.JobSpec{}
				})

				It("supported", func() {
					workload.MasterSpec.ResourceAllocationFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))

					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).NotTo(HaveOccurred())
					Expect(status.IsNetworkAccelerated()).To(BeTrue())

					Expect(status.WorkerMode.IsEnabled()).To(BeFalse())
					Expect(status.MasterMode.IsEnabled()).To(BeTrue())
				})

				It("unsupported", func() {
					status, err := GenerateStatus(ctx, k8sClient, workload)
					Expect(err).NotTo(HaveOccurred())
					Expect(status.IsNetworkAccelerated()).To(BeFalse())
				})
			})
		})
	})
})

// TODO: delete once clusterconstants.UseKaiNodePools=true
func getTestNodepool(name string, mnnvlSupport bool) *runv1.NodePool {
	return &runv1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: runv1.NodePoolStatus{
			GPUNetworkAccelerationDetected: mnnvlSupport,
		},
		Spec: runv1.NodePoolSpec{
			GPUNetworkAccelerationLabelKey: "gpu-network-acc",
		},
	}
}

// getTestKaiNodepool is the kai.resources twin of getTestNodepool. On that CRD the
// detection result and the label key are annotations rather than status/spec.
func getTestKaiNodepool(name string, mnnvlSupport bool) *kaiv1alpha1.NodePool {
	return &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: strconv.FormatBool(mnnvlSupport),
				kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey: "gpu-network-acc",
			},
		},
	}
}
