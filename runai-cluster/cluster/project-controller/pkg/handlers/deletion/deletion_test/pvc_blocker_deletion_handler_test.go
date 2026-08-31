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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("PVC's Deletion Handler", func() {
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
		handler = blockerForDisplayName(client, "Pvcs")
	})

	It("Finalizes a Project - PVC blocking", func() {
		// Given
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		Expect(client.Create(context.TODO(), &namespace)).To(Succeed())
		pvcs := *TestPVCsProjectScope.DeepCopy()
		for _, pvc := range pvcs.Items {
			Expect(client.Create(context.TODO(), &pvc)).To(Succeed())
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
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.ProjectConditionType("PvcsReady")))
			Expect(conditions[0].Status).To(Equal(v1.ConditionFalse))
			Expect(conditions[0].Reason).To(Equal(string(PvcsDeletionHandlerFailed)))
			Expect(conditions[0].Message).To(Equal("The project couldn't be deleted because the following needs to be deleted first:\nPersistentVolumeClaim non-runai-pvc-project-scope +1 more\n"))
		})

		for _, pvc := range pvcs.Items {
			Expect(client.Delete(context.TODO(), &pvc)).To(Succeed())
		}
	})
})
