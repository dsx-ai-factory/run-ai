/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package data

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testQueueName    = "test-queue"
	testProjectName  = "test-project"
	testPodGroupName = "test-pod-group"
	testNamespace    = "test-namespace"
)

func TestPodGroupData(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "workload-exporter PodGroup Data Test Suite")
}

var dataScheme *runtime.Scheme

var _ = BeforeSuite(func() {
	dataScheme = runtime.NewScheme()
	Expect(kaiv2.AddToScheme(dataScheme)).Should(Succeed())
	Expect(kaiv2alpha2.AddToScheme(dataScheme)).Should(Succeed())
})

func buildQueue(labels map[string]string) *kaiv2.Queue {
	return &kaiv2.Queue{
		ObjectMeta: metav1.ObjectMeta{Name: testQueueName, Labels: labels},
	}
}

func buildPodGroupForQueue() *kaiv2alpha2.PodGroup {
	return &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{Name: testPodGroupName, Namespace: testNamespace},
		Spec:       kaiv2alpha2.PodGroupSpec{Queue: testQueueName},
	}
}

func buildClient(objects ...client.Object) client.Reader {
	return fake.NewClientBuilder().WithScheme(dataScheme).WithObjects(objects...).Build()
}

var _ = Describe("GetProjectNameForPodGroup", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("when the queue carries the project label", func() {
		It("resolves the project name", func() {
			queue := buildQueue(map[string]string{constants.RunaiProjectLabel: testProjectName})

			projectName, err := GetProjectNameForPodGroup(ctx, buildPodGroupForQueue(), buildClient(queue))

			Expect(err).To(BeNil())
			Expect(projectName).To(Equal(testProjectName))
		})
	})

	// runai/queue is written on the project namespace, never on the queue, so a queue
	// carrying only it must not resolve to a project.
	Context("when the queue carries only the namespace queue label", func() {
		It("returns an error", func() {
			queue := buildQueue(map[string]string{constants.RunaiQueueLabel: testProjectName})

			projectName, err := GetProjectNameForPodGroup(ctx, buildPodGroupForQueue(), buildClient(queue))

			Expect(err).To(HaveOccurred())
			Expect(projectName).To(BeEmpty())
		})
	})

	Context("when the queue has no labels", func() {
		It("returns an error", func() {
			queue := buildQueue(map[string]string{})

			projectName, err := GetProjectNameForPodGroup(ctx, buildPodGroupForQueue(), buildClient(queue))

			Expect(err).To(HaveOccurred())
			Expect(projectName).To(BeEmpty())
		})
	})

	Context("when the queue does not exist", func() {
		It("returns an error", func() {
			projectName, err := GetProjectNameForPodGroup(ctx, buildPodGroupForQueue(), buildClient())

			Expect(err).To(HaveOccurred())
			Expect(projectName).To(BeEmpty())
		})
	})
})
