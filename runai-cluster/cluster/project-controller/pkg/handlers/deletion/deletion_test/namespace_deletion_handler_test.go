/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion_test

import (
	"context"
	"os"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers/deletion"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/test"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Namespace Deletion Handler", func() {

	var (
		handler   NamespaceDeletionHandler
		client    client.Client
		project   kaiv1alpha1.Project
		namespace v1.Namespace
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		project.Finalizers = append(project.Finalizers, config.FinalizerName())
		namespace = *TestNamespace.DeepCopy()

		client = fake.NewClientBuilder().WithScheme(scheme).Build()
		handler = NewNamespaceDeletionHandler(client, true)

		_ = client.Create(context.TODO(), &project)
		_ = client.Create(context.TODO(), &namespace)
	})

	AfterEach(func() {
		_ = os.Setenv(DeleteNamespaceFeatureFlag, "false")
	})

	It("Handles Project deletion", func() {
		By("Deleting the Project's OwnerReference from Namespace", func() {

			// Given
			Expect(handler.GetNamespace(TestNamespace.Name)).ToNot(BeZero())

			// When
			conditions, err := handler.OnDelete(&project)
			Expect(err).To(BeNil())

			// Then
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.NamespaceReady))
			Expect(conditions[0].Status).To(Equal(v1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))

			actualNamespace, err := handler.GetNamespace(TestNamespace.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(actualNamespace.Name).To(Equal(namespace.Name))
			Expect(actualNamespace.Namespace).To(Equal(namespace.Namespace))
			Expect(common.IsProjectOwner(&project, &actualNamespace)).To(BeNumerically("==", -1))
			Expect(actualNamespace.OwnerReferences).To(HaveLen(0))
		})
	})

	It("Deletes the related Namespace", func() {
		When("The feature flag is set", func() {

			// Given
			_ = os.Setenv(DeleteNamespaceFeatureFlag, "true")
			handler = NewNamespaceDeletionHandler(client, true)

			// When
			_, err := handler.OnDelete(&project)
			Expect(err).To(BeNil())

			// Then
			actualNamespace, err := handler.GetNamespace(TestNamespace.Name)
			Expect(err).To(HaveOccurred())
			Expect(actualNamespace).To(BeZero())
		})
	})

	When("CreateNamespaces flag is off", func() {
		It("Doesn't delete the Project's OwnerReference from Namespace", func() {
			Expect(client.Update(context.TODO(), &namespace)).To(Succeed())

			newHandler := NewNamespaceDeletionHandler(client, false)

			_, err := newHandler.OnDelete(&project)
			Expect(err).To(BeNil())

			actualNamespace, err := handler.GetNamespace(TestNamespace.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(common.IsProjectOwner(&project, &actualNamespace)).To(BeNumerically("==", 0))
			Expect(actualNamespace.OwnerReferences).To(HaveLen(1))
		})
	})
})
