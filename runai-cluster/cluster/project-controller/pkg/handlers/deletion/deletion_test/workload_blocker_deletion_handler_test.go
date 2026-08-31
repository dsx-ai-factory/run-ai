/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion_test

import (
	"context"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Workloads Deletion Handler", func() {
	var (
		handler   ConfigurableBlocker
		client    client.Client
		project   kaiv1alpha1.Project
		namespace v1.Namespace
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		namespace = *TestNamespace.DeepCopy()

		project.Finalizers = append(project.Finalizers, config.FinalizerName())
		client = fake.NewClientBuilder().WithScheme(scheme).Build()
		handler = blockerForDisplayName(client, "Workloads")
	})

	It("Finalizes a Project - Sanity", func() {
		// Given
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		Expect(client.Create(context.TODO(), &namespace)).To(Succeed())

		// Sanity
		Expect(handler.GetNamespace(TestNamespace.Name)).ToNot(BeZero())

		//When
		conditions, err := handler.OnDelete(&project)
		Expect(err).To(BeNil())

		// Then
		By("ProjectConditions are correct", func() {
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.ProjectConditionType("WorkloadsReady")))
			Expect(conditions[0].Status).To(Equal(v1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))
		})
	})
	It("Finalizes a Project - Externals blocking", func() {
		// Given
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		Expect(client.Create(context.TODO(), &namespace)).To(Succeed())
		externals := *TestExternalWorkloads.DeepCopy()
		for _, external := range externals.Items {
			Expect(client.Create(context.TODO(), &external)).To(Succeed())
		}

		// Sanity
		Expect(handler.GetNamespace(TestNamespace.Name)).ToNot(BeZero())

		//When
		conditions, err := handler.OnDelete(&project)
		Expect(err).To(Not(BeNil()))

		// Then
		By("ProjectConditions are correct", func() {
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.ProjectConditionType("WorkloadsReady")))
			Expect(conditions[0].Status).To(Equal(v1.ConditionFalse))
			Expect(conditions[0].Reason).To(Equal(string(WorkloadsDeletionHandlerFailed)))
			Expect(conditions[0].Message).To(Equal("The project couldn't be deleted because the following needs to be deleted first:\nExternalWorkload ew-1\n"))
		})

		for _, external := range externals.Items {
			Expect(client.Delete(context.TODO(), &external)).To(Succeed())
		}
	})
	It("Finalizes a Project - DistributedInferenceWorkload blocking", func() {
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		Expect(client.Create(context.TODO(), &namespace)).To(Succeed())
		workload := legacyWorkload(schema.GroupVersionKind{Group: "run.ai", Version: "v1alpha1", Kind: "DistributedInferenceWorkload"}, "distributed-inference-workload-1", namespace.Name)
		Expect(client.Create(context.TODO(), workload)).To(Succeed())

		conditions, err := handler.OnDelete(&project)

		Expect(err).To(HaveOccurred())
		Expect(conditions).To(HaveLen(1))
		Expect(conditions[0].Status).To(Equal(v1.ConditionFalse))
		Expect(conditions[0].Message).To(ContainSubstring("DistributedInferenceWorkload distributed-inference-workload-1"))
	})
	It("Finalizes a Project - RunaiJob blocking", func() {
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		Expect(client.Create(context.TODO(), &namespace)).To(Succeed())
		job := legacyWorkload(schema.GroupVersionKind{Group: "run.ai", Version: "v1", Kind: "RunaiJob"}, "runai-job-1", namespace.Name)
		Expect(client.Create(context.TODO(), job)).To(Succeed())

		conditions, err := handler.OnDelete(&project)

		Expect(err).To(HaveOccurred())
		Expect(conditions).To(HaveLen(1))
		Expect(conditions[0].Status).To(Equal(v1.ConditionFalse))
		Expect(conditions[0].Message).To(ContainSubstring("RunaiJob runai-job-1"))
	})
})

func legacyWorkload(gvk schema.GroupVersionKind, name, namespace string) *unstructured.Unstructured {
	workload := &unstructured.Unstructured{}
	workload.SetGroupVersionKind(gvk)
	workload.SetName(name)
	workload.SetNamespace(namespace)
	return workload
}
