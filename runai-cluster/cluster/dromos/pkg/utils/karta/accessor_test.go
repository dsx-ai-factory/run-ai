/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GetAllComponents", func() {
	var (
		ctrl     *gomock.Controller
		accessor *kartaresource.MockComponentAccessor
		karta    *kartav1alpha1.Karta
		factory  *kartaresource.ComponentFactory
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		accessor = kartaresource.NewMockComponentAccessor(ctrl)

		karta = &kartav1alpha1.Karta{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-karta",
			},
			Spec: kartav1alpha1.KartaSpec{
				StructureDefinition: kartav1alpha1.StructureDefinition{
					RootComponent: kartav1alpha1.ComponentDefinition{
						Name: "root",
						Kind: &kartav1alpha1.GroupVersionKind{
							Group:   "test.io",
							Version: "v1",
							Kind:    "TestKind",
						},
					},
					ChildComponents: []kartav1alpha1.ComponentDefinition{
						{
							Name: "child1",
							Kind: &kartav1alpha1.GroupVersionKind{
								Group:   "test.io",
								Version: "v1",
								Kind:    "ChildKind",
							},
						},
						{
							Name: "child2",
							Kind: &kartav1alpha1.GroupVersionKind{
								Group:   "test.io",
								Version: "v1",
								Kind:    "ChildKind",
							},
						},
					},
				},
			},
		}

		factory = kartaresource.NewComponentFactory(karta, accessor)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("returns children and root component combined", func() {
		result, err := GetAllComponents(factory)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(HaveLen(3))
	})
})

var _ = Describe("UpdateComponent", func() {
	var (
		ctx       context.Context
		ctrl      *gomock.Controller
		component *MockComponent
	)

	BeforeEach(func() {
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())
		component = NewMockComponent(ctrl)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("calls UpdateFragmentedPodSpec when instance has FragmentedPodSpec", func() {
		fragmentedSpec := kartaresource.FragmentedPodSpec{
			SchedulerName: "test-scheduler",
		}
		instances := map[string]kartaresource.ExtractedInstance{
			"instance-1": {
				FragmentedPodSpec: &fragmentedSpec,
			},
		}

		expectedSpecs := map[string]kartaresource.FragmentedPodSpec{
			"instance-1": fragmentedSpec,
		}
		component.EXPECT().UpdateFragmentedPodSpec(ctx, expectedSpecs).Return(nil)

		err := UpdateComponent(ctx, component, instances)

		Expect(err).NotTo(HaveOccurred())
	})

	It("calls UpdatePodSpec when instance has PodSpec", func() {
		podSpec := corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test-container"},
			},
		}
		instances := map[string]kartaresource.ExtractedInstance{
			"instance-1": {
				PodSpec: &podSpec,
			},
		}

		expectedSpecs := map[string]corev1.PodSpec{
			"instance-1": podSpec,
		}
		component.EXPECT().UpdatePodSpec(ctx, expectedSpecs).Return(nil)

		err := UpdateComponent(ctx, component, instances)

		Expect(err).NotTo(HaveOccurred())
	})

	It("calls UpdatePodTemplateSpec when instance has PodTemplateSpec", func() {
		podTemplateSpec := corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "test-container"},
				},
			},
		}
		instances := map[string]kartaresource.ExtractedInstance{
			"instance-1": {
				PodTemplateSpec: &podTemplateSpec,
			},
		}

		expectedSpecs := map[string]corev1.PodTemplateSpec{
			"instance-1": podTemplateSpec,
		}
		component.EXPECT().UpdatePodTemplateSpec(ctx, expectedSpecs).Return(nil)

		err := UpdateComponent(ctx, component, instances)

		Expect(err).NotTo(HaveOccurred())
	})

	It("returns error when UpdateFragmentedPodSpec fails", func() {
		fragmentedSpec := kartaresource.FragmentedPodSpec{}
		instances := map[string]kartaresource.ExtractedInstance{
			"instance-1": {
				FragmentedPodSpec: &fragmentedSpec,
			},
		}
		expectedErr := errors.New("update error")
		component.EXPECT().UpdateFragmentedPodSpec(ctx, gomock.Any()).Return(expectedErr)

		err := UpdateComponent(ctx, component, instances)

		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(expectedErr))
	})

	It("returns error when UpdatePodSpec fails", func() {
		podSpec := corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "test-container"},
			},
		}
		instances := map[string]kartaresource.ExtractedInstance{
			"instance-1": {
				PodSpec: &podSpec,
			},
		}
		expectedErr := errors.New("update error")
		component.EXPECT().UpdatePodSpec(ctx, gomock.Any()).Return(expectedErr)

		err := UpdateComponent(ctx, component, instances)

		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(expectedErr))
	})

	It("returns error when UpdatePodTemplateSpec fails", func() {
		podTemplateSpec := corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "test-container"},
				},
			},
		}
		instances := map[string]kartaresource.ExtractedInstance{
			"instance-1": {
				PodTemplateSpec: &podTemplateSpec,
			},
		}
		expectedErr := errors.New("update error")
		component.EXPECT().UpdatePodTemplateSpec(ctx, gomock.Any()).Return(expectedErr)

		err := UpdateComponent(ctx, component, instances)

		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(expectedErr))
	})
})

var _ = Describe("HasGPU", func() {
	DescribeTable("for different Karta pod definition types",
		func(testFunc func(kartaresource.ExtractedInstance), container corev1.Container) {
			instance := newExtractedInstanceWithPodTemplate(container)
			testFunc(instance)

			instance = newExtractedInstanceWithPodSpec(container)
			testFunc(instance)

			instance = newExtractedInstanceWithFragmentedSpec(container)
			testFunc(instance)
		},
		Entry("returns true when container has GPU in requests",
			func(instance kartaresource.ExtractedInstance) {
				result := HasGPU(instance)
				Expect(result).To(BeTrue())
			},
			newContainer("test", map[string]string{"nvidia.com/gpu": "1"}, nil),
		),
		Entry("returns true when container has GPU in limits",
			func(instance kartaresource.ExtractedInstance) {
				result := HasGPU(instance)
				Expect(result).To(BeTrue())
			},
			newContainer("test", nil, map[string]string{"nvidia.com/gpu": "2"}),
		),
		Entry("returns false when container without GPU resources",
			func(instance kartaresource.ExtractedInstance) {
				result := HasGPU(instance)
				Expect(result).To(BeFalse())
			},
			newContainer("test", map[string]string{"cpu": "1", "memory": "1Gi"}, nil),
		),
	)
})

var _ = Describe("CanStampPodLabels", func() {
	DescribeTable("by pod shape",
		func(instance kartaresource.ExtractedInstance, want bool) {
			Expect(CanStampPodLabels(instance)).To(Equal(want))
		},
		Entry("PodTemplateSpec can be stamped", newExtractedInstanceWithPodTemplate(), true),
		Entry("FragmentedPodSpec can be stamped", newExtractedInstanceWithFragmentedSpec(), true),
		Entry("bare PodSpec cannot be stamped", newExtractedInstanceWithPodSpec(), false),
		Entry("empty instance cannot be stamped", kartaresource.ExtractedInstance{}, false),
	)
})

func newContainer(name string, requests map[string]string, limits map[string]string) corev1.Container {
	container := corev1.Container{
		Name: name,
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{},
			Limits:   corev1.ResourceList{},
		},
	}

	for key, value := range requests {
		container.Resources.Requests[corev1.ResourceName(key)] = resource.MustParse(value)
	}

	for key, value := range limits {
		container.Resources.Limits[corev1.ResourceName(key)] = resource.MustParse(value)
	}

	return container
}

func newExtractedInstanceWithPodTemplate(containers ...corev1.Container) kartaresource.ExtractedInstance {
	return kartaresource.ExtractedInstance{
		PodTemplateSpec: &corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: containers,
			},
		},
	}
}

func newExtractedInstanceWithPodSpec(containers ...corev1.Container) kartaresource.ExtractedInstance {
	return kartaresource.ExtractedInstance{
		PodSpec: &corev1.PodSpec{
			Containers: containers,
		},
	}
}

func newExtractedInstanceWithFragmentedSpec(containers ...corev1.Container) kartaresource.ExtractedInstance {
	return kartaresource.ExtractedInstance{
		FragmentedPodSpec: &kartaresource.FragmentedPodSpec{
			Containers: containers,
		},
	}
}
