/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deletion_test

import (
	"context"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
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

var _ = Describe("Queue Deletion Handler", func() {

	var (
		handler   QueueDeletionHandler
		client    client.Client
		project   kaiv1alpha1.Project
		namespace v1.Namespace
		queue     kaiv2.Queue
	)

	BeforeEach(func() {
		project = *TestProject.DeepCopy()
		namespace = *TestNamespace.DeepCopy()
		queue = *TestQueue.DeepCopy()

		project.Finalizers = append(project.Finalizers, config.FinalizerName())
		client = fake.NewClientBuilder().WithScheme(scheme).Build()
		handler = NewQueueDeletionHandler(client)
	})

	It("Finalizes a Project", func() {

		// Given
		Expect(client.Create(context.TODO(), &project)).To(Succeed())
		Expect(client.Create(context.TODO(), &namespace)).To(Succeed())
		Expect(client.Create(context.TODO(), &queue)).To(Succeed())

		// Sanity
		Expect(handler.GetQueue(queue.Name)).ToNot(BeZero())
		Expect(handler.GetNamespace(TestNamespace.Name)).ToNot(BeZero())

		//When
		conditions, err := handler.OnDelete(&project)
		Expect(err).To(BeNil())

		// Then
		By("ProjectConditions are correct", func() {
			Expect(conditions).To(HaveLen(1))
			Expect(conditions[0]).ToNot(BeNil())
			Expect(conditions[0].Type).To(Equal(kaiv1alpha1.QueuesReady))
			Expect(conditions[0].Status).To(Equal(v1.ConditionTrue))
			Expect(conditions[0].Reason).To(Equal(""))
			Expect(conditions[0].Message).To(Equal(""))
		})

		By("Deleting the owned Queue", func() {
			for _, queueSpec := range project.Spec.Queues {
				actualQueue, err := handler.GetQueue(queueSpec.Name)
				Expect(err).To(HaveOccurred())
				Expect(actualQueue).To(BeZero())
			}
		})
	})
})
