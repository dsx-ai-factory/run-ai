/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package podgroup

import (
	"context"
	"testing"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testNamespace = "test-namespace"
	testPodGroup  = "test-pod-group"
	testQueueName = "test-queue"
)

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := kaiv2.AddToScheme(scheme); err != nil {
		t.Fatalf("add kaiv2 to scheme: %v", err)
	}
	if err := kaiv2alpha2.AddToScheme(scheme); err != nil {
		t.Fatalf("add kaiv2alpha2 to scheme: %v", err)
	}
	return scheme
}

func reconcileRequest() ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Name: testPodGroup, Namespace: testNamespace}}
}

// A pod group whose project cannot be resolved must surface the failure as an error
// so that the controller's exponential rate limiter backs the retries off. Returning
// a nil error with a fixed RequeueAfter bypasses the limiter and retries forever.
func TestReconcileFailureBacksOff(t *testing.T) {
	podGroup := &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testPodGroup,
			Namespace: testNamespace,
			Annotations: map[string]string{
				commonconstants.PodGroupRequestedMigDevicesAnnotation: `{"1g.5gb":1}`,
			},
		},
		Spec: kaiv2alpha2.PodGroupSpec{Queue: testQueueName},
	}
	// No queue exists, so the project lookup cannot succeed.
	k8sClient := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(podGroup).Build()

	result, err := NewPodGroupController(k8sClient).Reconcile(context.Background(), reconcileRequest())

	if err == nil {
		t.Fatal("expected an error so that the rate limiter handles the retry")
	}
	if result.RequeueAfter != 0 {
		t.Errorf("got RequeueAfter %v, want 0 so the rate limiter decides the delay", result.RequeueAfter)
	}
}

// Since 2.25 nothing writes the MIG annotation, so the overwhelming majority of pod
// groups have nothing to report. That must be a no-op, not an error and not a requeue.
func TestReconcileWithoutMigAnnotationIsNoOp(t *testing.T) {
	podGroup := &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{Name: testPodGroup, Namespace: testNamespace},
		Spec:       kaiv2alpha2.PodGroupSpec{Queue: testQueueName},
	}
	k8sClient := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(podGroup).Build()

	result, err := NewPodGroupController(k8sClient).Reconcile(context.Background(), reconcileRequest())

	if err != nil {
		t.Fatalf("unexpected error for a pod group without the MIG annotation: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Errorf("got RequeueAfter %v, want 0", result.RequeueAfter)
	}
}

func TestHasMigAnnotation(t *testing.T) {
	for _, tc := range []struct {
		name        string
		annotations map[string]string
		want        bool
	}{
		{name: "no annotations", want: false},
		{name: "unrelated annotation", annotations: map[string]string{"other": "value"}, want: false},
		{
			name:        "mig annotation present",
			annotations: map[string]string{commonconstants.PodGroupRequestedMigDevicesAnnotation: `{"1g.5gb":1}`},
			want:        true,
		},
		{
			name:        "empty mig annotation",
			annotations: map[string]string{commonconstants.PodGroupRequestedMigDevicesAnnotation: "{}"},
			want:        true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			podGroup := &kaiv2alpha2.PodGroup{ObjectMeta: metav1.ObjectMeta{Annotations: tc.annotations}}
			if got := hasMigAnnotation(podGroup); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
