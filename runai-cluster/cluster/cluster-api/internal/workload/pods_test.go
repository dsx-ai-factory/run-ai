/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	iwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	workloadWithBackendUUID = &Workload{
		Name:        "test-workload",
		Type:        api.Workspace,
		BackendUUID: "test-uuid",
	}
	workloadWithoutBackendUUID = &Workload{
		Name: "test-workload",
		Type: api.Training,
	}
	externalWorkload = &Workload{
		Name:        "test-workload",
		Type:        api.External,
		BackendUUID: "original-workload-uid",
		ClusterUUID: "test-uuid",
		Original: &v1alpha1.ExternalWorkload{
			Spec: v1alpha1.ExternalWorkloadSpec{
				OriginalWorkloadUID: "original-workload-uid",
			},
		},
	}
)

var _ = Describe("Pods", func() {
	var (
		ctx        context.Context
		podGetter  *Pods
		fakeClient client.Client
	)
	scheme := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(scheme))   // core
	utilruntime.Must(iwv1.AddToScheme(scheme))     // interactive workload
	utilruntime.Must(twv1.AddToScheme(scheme))     // training workload
	utilruntime.Must(v1alpha1.AddToScheme(scheme)) // training workload
	BeforeEach(func() {
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		ctx = context.Background()
		podGetter = NewPodsSelector(fakeClient)
	})

	Describe("GetPod", func() {
		DescribeTable("Pods should pass the request to the next handler",
			func(workload *Workload, createFunc func(workload *Workload, k8sClient client.Client) error, preferredPod string) {
				err := createFunc(workload, fakeClient)
				Expect(err).NotTo(HaveOccurred())

				pod, err := podGetter.GetPod(ctx, workload, preferredPod)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod).To(BeAssignableToTypeOf(&corev1.Pod{}))
			},
			Entry("workload with backend uuid - with preferred", workloadWithBackendUUID, CreateResourceForWorkload, "test-workload"),
			Entry("workload with backend uuid - without preferred", workloadWithBackendUUID, CreateResourceForWorkload, ""),
			Entry("workload without backend uuid - with preferred", workloadWithoutBackendUUID, CreateResourceForWorkloadBackward, "test-workload"),
			Entry("workload without backend uuid - without preferred", workloadWithoutBackendUUID, CreateResourceForWorkloadBackward, ""),
			Entry("external workload - with preferred", externalWorkload, CreateResourceForWorkload, "test-workload"),
			Entry("external workload - without preferred", externalWorkload, CreateResourceForWorkload, ""),
		)
	})

	Describe("getPod", func() {
		BeforeEach(func() {
			err := CreateResourceForWorkload(workloadWithBackendUUID, fakeClient)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("when the preferredPod is not empty", func() {
			It("should return the preferred pod", func() {
				pod, err := podGetter.getPod(ctx, workloadWithBackendUUID, workloadWithBackendUUID.Name, workloadWithBackendUUID.BackendUUID.String())
				Expect(err).NotTo(HaveOccurred())
				Expect(pod).To(BeAssignableToTypeOf(&corev1.Pod{}))
			})
			It("should return err preferred pod not found", func() {
				_, err := podGetter.getPod(ctx, workloadWithBackendUUID, "preferredPod", workloadWithBackendUUID.BackendUUID.String())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the preferredPod is empty", func() {
			It("should return the first pod of the workload", func() {
				pod, err := podGetter.getPod(ctx, workloadWithBackendUUID, "", workloadWithBackendUUID.BackendUUID.String())
				Expect(err).NotTo(HaveOccurred())
				Expect(pod).To(BeAssignableToTypeOf(&corev1.Pod{}))
			})
		})

		Context("when the workload has no pods", func() {
			It("should return NotReadyError for DistributedInferenceJob with no pods", func() {
				workload := &Workload{
					Name:        "test-dinf",
					Namespace:   "test-namespace",
					JobType:     api.DistributedInferenceJob,
					BackendUUID: "test-dinf-uuid",
					ClusterUUID: "test-dinf-cluster-uuid",
				}

				_, err := podGetter.GetPod(ctx, workload, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("pod not found"))
				Expect(err).To(BeAssignableToTypeOf(&api.NotReadyError{}))
			})

			It("should return NotReadyError for InferenceJob with no pods", func() {
				workload := &Workload{
					Name:        "test-inf",
					Namespace:   "test-namespace",
					JobType:     api.InferenceJob,
					BackendUUID: "test-inf-uuid",
					ClusterUUID: "test-inf-cluster-uuid",
				}

				_, err := podGetter.GetPod(ctx, workload, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("pod not found"))
				Expect(err).To(BeAssignableToTypeOf(&api.NotReadyError{}))
			})
		})
	})
})

func CreateResourceForWorkloadBackward(workload *Workload, k8sClient client.Client) error {
	workloadResource, err := GetK8sResource(workload.Type)
	if err != nil {
		return err
	}

	workloadResource.SetName(workload.Name)
	workloadResource.SetNamespace(workload.Namespace)

	err = k8sClient.Create(context.Background(), workloadResource)
	if err != nil {
		return err
	}

	workload.ClusterUUID = "test-uuid"
	backendUUID := workload.ClusterUUID
	if workload.BackendUUID != "" {
		backendUUID = workload.BackendUUID
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workload.Name,
			Namespace: workload.Namespace,
			Labels: map[string]string{
				constants.LabelWorkloadKind:             workload.Type.ToK8sResource(),
				constants.LabelWorkloadName:             workload.Name,
				commonconstants.TopOwnerWorkloadIdLabel: backendUUID.String(),
				commonconstants.ClusterTopOwnerUidLabel: workload.ClusterUUID.String(),
				commonconstants.TopOwnerUidLabel:        backendUUID.String(),
			},
		},
	}

	err = controllerruntime.SetControllerReference(workloadResource, pod, k8sClient.Scheme())
	if err != nil {
		return err
	}

	return k8sClient.Create(context.Background(), pod)
}

func CreateResourceForWorkload(workload *Workload, k8sClient client.Client) error {
	if workload.ClusterUUID == "" {
		workload.ClusterUUID = "test-uuid"
	}
	backendUUID := workload.ClusterUUID
	if workload.BackendUUID != "" {
		backendUUID = workload.BackendUUID
	}

	// For external workloads, use the OriginalWorkloadUID as the cluster UID for the label
	clusterUID := workload.ClusterUUID.String()
	if workload.Type == api.External && workload.Original != nil {
		if external, ok := workload.Original.(*v1alpha1.ExternalWorkload); ok {
			clusterUID = string(external.Spec.OriginalWorkloadUID)
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workload.Name,
			Namespace: workload.Namespace,
			Labels: map[string]string{
				commonconstants.TopOwnerWorkloadIdLabel: backendUUID.String(),
				commonconstants.ClusterTopOwnerUidLabel: clusterUID,
				commonconstants.TopOwnerUidLabel:        backendUUID.String(),
			},
		},
	}

	err := k8sClient.Create(context.Background(), pod)
	if err != nil {
		return err
	}

	workloadResource, err := GetK8sResource(workload.Type)
	if err != nil {
		return err
	}

	workloadResource.SetName(workload.Name)
	workloadResource.SetNamespace(workload.Namespace)

	err = controllerruntime.SetControllerReference(workloadResource, pod, k8sClient.Scheme())
	if err != nil {
		return err
	}

	return k8sClient.Create(context.Background(), workloadResource)
}
