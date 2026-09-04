/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/project-controller/pkg/handlers"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Queue naming", func() {

	Describe("GetQueueName - suggested name from the spec, else a fallback", func() {
		It("uses the queue spec name when it is set, for any nodepool", func() {
			Expect(handlers.GetQueueName(
				kaiv1alpha1.QueueConfig{Name: "explicit-queue", Nodepool: "gpu-pool"}, "dep-a"),
			).To(Equal("explicit-queue"))

			Expect(handlers.GetQueueName(
				kaiv1alpha1.QueueConfig{Name: "explicit-queue", Nodepool: config.Get().DefaultNodepoolName}, "dep-a"),
			).To(Equal("explicit-queue"))
		})

		It("falls back to <owner>-<nodepool> when the spec name is empty", func() {
			Expect(handlers.GetQueueName(
				kaiv1alpha1.QueueConfig{Name: "", Nodepool: "gpu-pool"}, "dep-a"),
			).To(Equal("dep-a-gpu-pool"))
		})

		It("falls back to <owner> for the default nodepool when the spec name is empty", func() {
			Expect(handlers.GetQueueName(
				kaiv1alpha1.QueueConfig{Name: "", Nodepool: config.Get().DefaultNodepoolName}, "dep-a"),
			).To(Equal("dep-a"))
		})
	})

	Describe("collision - a random suffix is appended when the suggested name is taken", func() {
		It("keeps the suggested name as a prefix and appends a suffix when another resource owns it", func() {
			k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			handler := handlers.NewDepartmentHandler(k8sClient)

			// A department whose default-nodepool queue has no explicit name, so its suggested
			// name is the fallback "<department-name>" = "dep-a".
			department := &kaiv1alpha1.Department{
				ObjectMeta: metav1.ObjectMeta{Name: "dep-a", UID: "uid-dep-a"},
				Spec: kaiv1alpha1.DepartmentSpec{
					Queues: []kaiv1alpha1.QueueConfig{{
						Nodepool:  config.Get().DefaultNodepoolName,
						Resources: &kaiv1alpha1.QueueResourcesConfig{GPU: kaiv1alpha1.SystemResource{Deserved: 1}},
					}},
				},
			}
			suggestedName := handlers.GetQueueName(department.Spec.Queues[0], department.Name)
			Expect(suggestedName).To(Equal("dep-a"))

			// A queue with that exact name already exists, owned by a DIFFERENT department.
			takenQueue := &kaiv2.Queue{
				ObjectMeta: metav1.ObjectMeta{
					Name:   suggestedName,
					Labels: map[string]string{config.Get().QueueDepartmentNameLabelKey: "other-dep"},
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: kaiv1alpha1.GroupVersion.Identifier(),
						Kind:       common.DepartmentKind,
						Name:       "other-dep",
						UID:        "uid-other-dep",
					}},
				},
			}
			Expect(k8sClient.Create(context.Background(), takenQueue)).To(Succeed())

			// When
			Expect(handler.Handle(context.Background(), department)).To(Succeed())

			// Then - dep-a's queue was created with the suggested name as a prefix plus a random
			// suffix (not the exact suggested name, which is taken by other-dep), and is owned by dep-a.
			created, err := getQueueForDepartmentByLabel(k8sClient, department.Name, config.Get().DefaultNodepoolName)
			Expect(err).ToNot(HaveOccurred())
			Expect(created.Name).ToNot(Equal(suggestedName))
			Expect(created.Name).To(HavePrefix(suggestedName + "-"))
			Expect(isOwnedByDepartment(created.OwnerReferences)).To(BeTrue())
			Expect(created.OwnerReferences[0].UID).To(Equal(department.UID))

			// The pre-existing queue is left untouched (still owned by other-dep).
			original, err := getQueue(k8sClient, suggestedName)
			Expect(err).ToNot(HaveOccurred())
			Expect(original.OwnerReferences[0].Name).To(Equal("other-dep"))
		})
	})
})
