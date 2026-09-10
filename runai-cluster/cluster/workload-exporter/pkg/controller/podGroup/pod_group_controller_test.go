/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package podgroup

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testNamespace    = "test-namespace"
	testPodGroupName = "test-pod-group"
	testQueueName    = "test-queue"
)

func TestPodGroupController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "workload-exporter PodGroup Controller Test Suite")
}

var controllerScheme *runtime.Scheme

var _ = BeforeSuite(func() {
	controllerScheme = runtime.NewScheme()
	Expect(kaiv2.AddToScheme(controllerScheme)).Should(Succeed())
	Expect(kaiv2alpha2.AddToScheme(controllerScheme)).Should(Succeed())
})

func buildPodGroup(annotations map[string]string) *kaiv2alpha2.PodGroup {
	return &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:        testPodGroupName,
			Namespace:   testNamespace,
			Annotations: annotations,
		},
		Spec: kaiv2alpha2.PodGroupSpec{Queue: testQueueName},
	}
}

func buildClient(objects ...client.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(controllerScheme).WithObjects(objects...).Build()
}

func getReconcileRequest() ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Name: testPodGroupName, Namespace: testNamespace}}
}

var _ = Describe("PodGroupController", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("Reconcile", func() {
		// A failure must surface as an error so that the controller's exponential rate
		// limiter backs the retries off. Returning a nil error with a fixed RequeueAfter
		// bypasses the limiter and retries forever.
		Context("when the project cannot be resolved", func() {
			It("returns an error and leaves the retry delay to the rate limiter", func() {
				podGroup := buildPodGroup(map[string]string{
					commonconstants.PodGroupRequestedMigDevicesAnnotation: `{"1g.5gb":1}`,
				})
				// No queue exists, so the project lookup cannot succeed.
				controller := NewPodGroupController(buildClient(podGroup))

				result, err := controller.Reconcile(ctx, getReconcileRequest())

				Expect(err).To(HaveOccurred())
				Expect(result.RequeueAfter).To(BeZero())
			})
		})

		// Since 2.25 nothing writes the MIG annotation, so the overwhelming majority of
		// pod groups have nothing to report. That must be a no-op, not an error.
		Context("when the pod group has no MIG annotation", func() {
			It("does nothing and does not requeue", func() {
				controller := NewPodGroupController(buildClient(buildPodGroup(nil)))

				result, err := controller.Reconcile(ctx, getReconcileRequest())

				Expect(err).To(BeNil())
				Expect(result.RequeueAfter).To(BeZero())
			})
		})

		Context("when the pod group does not exist", func() {
			It("does nothing and does not requeue", func() {
				controller := NewPodGroupController(buildClient())

				result, err := controller.Reconcile(ctx, getReconcileRequest())

				Expect(err).To(BeNil())
				Expect(result.RequeueAfter).To(BeZero())
			})
		})
	})

	Describe("hasMigAnnotation", func() {
		DescribeTable("reports whether the MIG annotation is present",
			func(annotations map[string]string, expected bool) {
				Expect(hasMigAnnotation(buildPodGroup(annotations))).To(Equal(expected))
			},
			Entry("no annotations", nil, false),
			Entry("unrelated annotation", map[string]string{"other": "value"}, false),
			Entry("mig annotation present",
				map[string]string{commonconstants.PodGroupRequestedMigDevicesAnnotation: `{"1g.5gb":1}`}, true),
			// Pre-2.25 non-MIG pod groups carry the annotation with an empty map, which is
			// what drags them into the reconcile loop after an upgrade.
			Entry("empty mig annotation",
				map[string]string{commonconstants.PodGroupRequestedMigDevicesAnnotation: "{}"}, true),
		)
	})
})
