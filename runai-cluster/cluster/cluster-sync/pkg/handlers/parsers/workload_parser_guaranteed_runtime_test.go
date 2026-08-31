/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetGuaranteedRuntimeTimestamp(t *testing.T) {
	tests := []struct {
		name                    string
		minGuaranteedRuntime    string
		lastStartTimestamp      string
		podNodePoolLabel        string
		expectedGuaranteedEnd   bool
		expectedDurationMinutes int
	}{
		{
			name:                    "valid guaranteed runtime calculation",
			minGuaranteedRuntime:    "30m",
			lastStartTimestamp:      time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			podNodePoolLabel:        "test-nodepool",
			expectedGuaranteedEnd:   true,
			expectedDurationMinutes: 30,
		},
		{
			name:                    "zero guaranteed runtime",
			minGuaranteedRuntime:    "0s",
			lastStartTimestamp:      time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			podNodePoolLabel:        "test-nodepool",
			expectedGuaranteedEnd:   true,
			expectedDurationMinutes: 0,
		},
		{
			name:                    "one hour guaranteed runtime",
			minGuaranteedRuntime:    "1h",
			lastStartTimestamp:      time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			podNodePoolLabel:        "test-nodepool",
			expectedGuaranteedEnd:   true,
			expectedDurationMinutes: 60,
		},
		{
			name:                    "missing last start timestamp",
			minGuaranteedRuntime:    "30m",
			lastStartTimestamp:      "",
			podNodePoolLabel:        "test-nodepool",
			expectedGuaranteedEnd:   false,
			expectedDurationMinutes: 0,
		},
		{
			name:                    "invalid last start timestamp",
			minGuaranteedRuntime:    "30m",
			lastStartTimestamp:      "invalid-timestamp",
			podNodePoolLabel:        "test-nodepool",
			expectedGuaranteedEnd:   false,
			expectedDurationMinutes: 0,
		},
		{
			name:                    "missing node pool label",
			minGuaranteedRuntime:    "30m",
			lastStartTimestamp:      time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			podNodePoolLabel:        "",
			expectedGuaranteedEnd:   false,
			expectedDurationMinutes: 0,
		},
		{
			name:                    "invalid min guaranteed runtime",
			minGuaranteedRuntime:    "invalid-duration",
			lastStartTimestamp:      time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
			podNodePoolLabel:        "test-nodepool",
			expectedGuaranteedEnd:   false,
			expectedDurationMinutes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			// Create fake client
			fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).Build()

			// Create test namespace
			namespace := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Labels: map[string]string{
						constants.RunaiQueueLabel: "test-project",
					},
				},
			}

			// Create test project
			project := &v2.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-project",
				},
				Spec: v2.ProjectSpec{
					Department: "test-department",
				},
			}

			// Create test node pool
			nodePool := &v1alpha1.NodePool{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-nodepool",
				},
				Spec: v1alpha1.NodePoolSpec{
					MinGuaranteedRuntime: tt.minGuaranteedRuntime,
				},
				Status: v1alpha1.NodePoolStatus{
					Phase: v1alpha1.NodePoolReady,
				},
			}

			// Create test pod
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "test-namespace",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "test-container",
							Image: "test-image",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			}

			// Add node pool label if specified
			if tt.podNodePoolLabel != "" {
				if pod.Labels == nil {
					pod.Labels = make(map[string]string)
				}
				pod.Labels[v1alpha1.NodePoolNameLabel] = tt.podNodePoolLabel
			}

			// Create test pod group
			podGroup := &kaiv2alpha2.PodGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-podgroup",
					Namespace: "test-namespace",
				},
			}

			// Add last start timestamp annotation if specified
			if tt.lastStartTimestamp != "" {
				if podGroup.Annotations == nil {
					podGroup.Annotations = make(map[string]string)
				}
				podGroup.Annotations[lastStartTimestampAnnotation] = tt.lastStartTimestamp
			}

			// Add objects to fake client
			err := fakeClient.Create(context.Background(), namespace)
			g.Expect(err).To(BeNil())
			err = fakeClient.Create(context.Background(), project)
			g.Expect(err).To(BeNil())
			err = fakeClient.Create(context.Background(), nodePool)
			// kai.resources twin, so this spec resolves the same runtime under both values of
			// the compile-time UseKaiNodePools flag.
			// TODO: drop the legacy nodePool once commonconstants.UseKaiNodePools=true
			err = fakeClient.Create(context.Background(), nodePoolWithMinRuntime(nodePool.Name, nodePool.Spec.MinGuaranteedRuntime))
			g.Expect(err).To(BeNil())
			err = fakeClient.Create(context.Background(), pod)
			g.Expect(err).To(BeNil())
			err = fakeClient.Create(context.Background(), podGroup)
			g.Expect(err).To(BeNil())

			// Create workload parser
			workloadParser := NewWorkloadParser(fakeClient)

			// Test the getGuaranteedRuntimeTimestamp method
			podGroups := []*kaiv2alpha2.PodGroup{podGroup}
			pods := []corev1.Pod{*pod}

			result := workloadParser.getGuaranteedRuntimeTimestamp(podGroups, &pods)

			if tt.expectedGuaranteedEnd {
				g.Expect(result).NotTo(BeNil(), "Expected guaranteed runtime timestamp to be calculated")

				// Parse the expected start time
				expectedStartTime, err := time.Parse(time.RFC3339, tt.lastStartTimestamp)
				g.Expect(err).To(BeNil())

				// Parse the guaranteed duration
				guaranteedDuration, err := time.ParseDuration(tt.minGuaranteedRuntime)
				g.Expect(err).To(BeNil())

				// Calculate expected end time
				expectedEndTime := expectedStartTime.Add(guaranteedDuration)

				// Verify the result is close to expected (within 1 second for test timing)
				g.Expect(*result).To(BeTemporally("~", expectedEndTime, time.Second),
					"Guaranteed runtime end time should match expected calculation")
			} else {
				g.Expect(result).To(BeNil(), "Expected guaranteed runtime timestamp to be nil")
			}
		})
	}
}

func TestGuaranteedRuntimeEndsAtValue(t *testing.T) {
	t.Run("exact guaranteed runtime calculation", func(t *testing.T) {
		g := NewGomegaWithT(t)

		// Create a fixed timestamp for deterministic testing
		startTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		guaranteedDuration := 45 * time.Minute
		expectedEndTime := startTime.Add(guaranteedDuration)

		// Create fake client
		fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).Build()

		// Create test namespace
		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
				Labels: map[string]string{
					constants.RunaiQueueLabel: "test-project",
				},
			},
		}

		// Create test project
		project := &v2.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-project",
			},
			Spec: v2.ProjectSpec{
				Department: "test-department",
			},
		}

		// Create test node pool with 45 minute guaranteed runtime
		nodePool := &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-nodepool",
			},
			Spec: v1alpha1.NodePoolSpec{
				MinGuaranteedRuntime: "45m",
			},
			Status: v1alpha1.NodePoolStatus{
				Phase: v1alpha1.NodePoolReady,
			},
		}

		// Create test pod
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					v1alpha1.NodePoolNameLabel: "test-nodepool",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "test-image",
					},
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}

		// Create test pod group with fixed start timestamp
		podGroup := &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-podgroup",
				Namespace: "test-namespace",
				Annotations: map[string]string{
					lastStartTimestampAnnotation: startTime.Format(time.RFC3339),
				},
			},
			Status: kaiv2alpha2.PodGroupStatus{
				Running: 1,
			},
		}

		// Add objects to fake client
		err := fakeClient.Create(context.Background(), namespace)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), project)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), nodePool)
		// kai.resources twin, so this spec resolves the same runtime under both values of
		// the compile-time UseKaiNodePools flag.
		// TODO: drop the legacy nodePool once commonconstants.UseKaiNodePools=true
		err = fakeClient.Create(context.Background(), nodePoolWithMinRuntime(nodePool.Name, nodePool.Spec.MinGuaranteedRuntime))
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), pod)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), podGroup)
		g.Expect(err).To(BeNil())

		// Create workload parser
		workloadParser := NewWorkloadParser(fakeClient)

		// Test the getGuaranteedRuntimeTimestamp method
		podGroups := []*kaiv2alpha2.PodGroup{podGroup}
		pods := []corev1.Pod{*pod}

		result := workloadParser.getGuaranteedRuntimeTimestamp(podGroups, &pods)

		// Verify the result is exactly as expected
		g.Expect(result).NotTo(BeNil(), "Expected guaranteed runtime timestamp to be calculated")
		g.Expect(*result).To(Equal(expectedEndTime),
			"Guaranteed runtime end time should be exactly start time + 45 minutes")

		// Verify the time difference is exactly 45 minutes
		actualDuration := result.Sub(startTime)
		g.Expect(actualDuration).To(Equal(guaranteedDuration),
			"Duration should be exactly 45 minutes")
	})

	t.Run("zero guaranteed runtime", func(t *testing.T) {
		g := NewGomegaWithT(t)

		startTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		expectedEndTime := startTime // Should be same as start time for zero duration

		fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).Build()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
				Labels: map[string]string{
					constants.RunaiQueueLabel: "test-project",
				},
			},
		}

		project := &v2.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-project",
			},
			Spec: v2.ProjectSpec{
				Department: "test-department",
			},
		}

		nodePool := &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-nodepool",
			},
			Spec: v1alpha1.NodePoolSpec{
				MinGuaranteedRuntime: "0s",
			},
			Status: v1alpha1.NodePoolStatus{
				Phase: v1alpha1.NodePoolReady,
			},
		}

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					v1alpha1.NodePoolNameLabel: "test-nodepool",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "test-image",
					},
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}

		podGroup := &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-podgroup",
				Namespace: "test-namespace",
				Annotations: map[string]string{
					lastStartTimestampAnnotation: startTime.Format(time.RFC3339),
				},
			},
			Status: kaiv2alpha2.PodGroupStatus{
				Running: 1,
			},
		}

		err := fakeClient.Create(context.Background(), namespace)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), project)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), nodePool)
		// kai.resources twin, so this spec resolves the same runtime under both values of
		// the compile-time UseKaiNodePools flag.
		// TODO: drop the legacy nodePool once commonconstants.UseKaiNodePools=true
		err = fakeClient.Create(context.Background(), nodePoolWithMinRuntime(nodePool.Name, nodePool.Spec.MinGuaranteedRuntime))
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), pod)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), podGroup)
		g.Expect(err).To(BeNil())

		workloadParser := NewWorkloadParser(fakeClient)
		podGroups := []*kaiv2alpha2.PodGroup{podGroup}
		pods := []corev1.Pod{*pod}

		result := workloadParser.getGuaranteedRuntimeTimestamp(podGroups, &pods)

		g.Expect(result).NotTo(BeNil(), "Expected guaranteed runtime timestamp to be calculated even for zero duration")
		g.Expect(*result).To(Equal(expectedEndTime),
			"Guaranteed runtime end time should equal start time for zero duration")
	})

	t.Run("complex duration calculation", func(t *testing.T) {
		g := NewGomegaWithT(t)

		startTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		guaranteedDuration := 2*time.Hour + 30*time.Minute + 15*time.Second
		expectedEndTime := startTime.Add(guaranteedDuration)

		fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).Build()

		namespace := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-namespace",
				Labels: map[string]string{
					constants.RunaiQueueLabel: "test-project",
				},
			},
		}

		project := &v2.Project{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-project",
			},
			Spec: v2.ProjectSpec{
				Department: "test-department",
			},
		}

		nodePool := &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-nodepool",
			},
			Spec: v1alpha1.NodePoolSpec{
				MinGuaranteedRuntime: "2h30m15s",
			},
			Status: v1alpha1.NodePoolStatus{
				Phase: v1alpha1.NodePoolReady,
			},
		}

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					v1alpha1.NodePoolNameLabel: "test-nodepool",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "test-image",
					},
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}

		podGroup := &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-podgroup",
				Namespace: "test-namespace",
				Annotations: map[string]string{
					lastStartTimestampAnnotation: startTime.Format(time.RFC3339),
				},
			},
			Status: kaiv2alpha2.PodGroupStatus{
				Running: 1,
			},
		}

		err := fakeClient.Create(context.Background(), namespace)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), project)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), nodePool)
		// kai.resources twin, so this spec resolves the same runtime under both values of
		// the compile-time UseKaiNodePools flag.
		// TODO: drop the legacy nodePool once commonconstants.UseKaiNodePools=true
		err = fakeClient.Create(context.Background(), nodePoolWithMinRuntime(nodePool.Name, nodePool.Spec.MinGuaranteedRuntime))
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), pod)
		g.Expect(err).To(BeNil())
		err = fakeClient.Create(context.Background(), podGroup)
		g.Expect(err).To(BeNil())

		workloadParser := NewWorkloadParser(fakeClient)
		podGroups := []*kaiv2alpha2.PodGroup{podGroup}
		pods := []corev1.Pod{*pod}

		result := workloadParser.getGuaranteedRuntimeTimestamp(podGroups, &pods)

		g.Expect(result).NotTo(BeNil(), "Expected guaranteed runtime timestamp to be calculated")
		g.Expect(*result).To(Equal(expectedEndTime),
			"Guaranteed runtime end time should be exactly start time + 2h30m15s")

		// Verify the time difference is exactly as expected
		actualDuration := result.Sub(startTime)
		g.Expect(actualDuration).To(Equal(guaranteedDuration),
			"Duration should be exactly 2h30m15s")
	})
}

func TestGetGuaranteedRuntimeTimestampEdgeCases(t *testing.T) {
	t.Run("multiple pod groups", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).Build()
		workloadParser := NewWorkloadParser(fakeClient)

		// Create multiple pod groups
		podGroup1 := &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-podgroup-1",
				Namespace: "test-namespace",
			},
			Status: kaiv2alpha2.PodGroupStatus{
				Running: 1,
			},
		}
		podGroup2 := &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-podgroup-2",
				Namespace: "test-namespace",
			},
			Status: kaiv2alpha2.PodGroupStatus{
				Running: 1,
			},
		}

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "test-namespace",
				Labels: map[string]string{
					v1alpha1.NodePoolNameLabel: "test-nodepool",
				},
			},
		}

		podGroups := []*kaiv2alpha2.PodGroup{podGroup1, podGroup2}
		pods := []corev1.Pod{*pod}

		result := workloadParser.getGuaranteedRuntimeTimestamp(podGroups, &pods)
		g.Expect(result).To(BeNil(), "Should return nil when multiple pod groups exist")
	})

	t.Run("no pods", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).Build()
		workloadParser := NewWorkloadParser(fakeClient)

		podGroup := &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-podgroup",
				Namespace: "test-namespace",
			},
			Status: kaiv2alpha2.PodGroupStatus{
				Running: 1,
			},
		}

		podGroups := []*kaiv2alpha2.PodGroup{podGroup}
		pods := []corev1.Pod{}

		result := workloadParser.getGuaranteedRuntimeTimestamp(podGroups, &pods)
		g.Expect(result).To(BeNil(), "Should return nil when no pods exist")
	})

	t.Run("nil pods slice", func(t *testing.T) {
		g := NewGomegaWithT(t)

		fakeClient := fake.NewClientBuilder().WithScheme(getTestScheme()).Build()
		workloadParser := NewWorkloadParser(fakeClient)

		podGroup := &kaiv2alpha2.PodGroup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-podgroup",
				Namespace: "test-namespace",
			},
			Status: kaiv2alpha2.PodGroupStatus{
				Running: 1,
			},
		}

		podGroups := []*kaiv2alpha2.PodGroup{podGroup}
		var pods *[]corev1.Pod = nil

		result := workloadParser.getGuaranteedRuntimeTimestamp(podGroups, pods)
		g.Expect(result).To(BeNil(), "Should return nil when pods slice is nil")
	})
}

// Helper function to get test scheme
func getTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = v1alpha1.AddToScheme(scheme)
	_ = kaiv1alpha1.AddToScheme(scheme)
	_ = v2.AddToScheme(scheme)
	_ = kaiv2alpha2.AddToScheme(scheme)
	return scheme
}

// nodePoolWithMinRuntime is the kai.resources twin of the legacy fixtures above. There
// the value lives under schedulingShardConfig.minRuntime, which splits it into a preempt
// and a reclaim value; the migration writes the legacy value to both.
// TODO: delete once commonconstants.UseKaiNodePools=true
func nodePoolWithMinRuntime(name, minGuaranteedRuntime string) *kaiv1alpha1.NodePool {
	nodePool := &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     kaiv1alpha1.NodePoolStatus{Phase: kaiv1alpha1.NodePoolReady},
	}
	if minGuaranteedRuntime != "" {
		nodePool.Spec.SchedulingShardConfig = &kaiv1alpha1.SchedulingShardConfig{
			MinRuntime: &kaiv1.MinRuntime{
				PreemptMinRuntime: ptr.To(minGuaranteedRuntime),
				ReclaimMinRuntime: ptr.To(minGuaranteedRuntime),
			},
		}
	}
	return nodePool
}
