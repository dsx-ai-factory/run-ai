/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"testing"

	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeRunaiPod(labels map[string]string, annotations map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Labels: labels, Annotations: annotations},
		Spec:       corev1.PodSpec{SchedulerName: commonconstants.RunaiScheduler},
	}
}

func TestHasElementIdLabelAdded(t *testing.T) {
	tests := []struct {
		name      string
		oldLabels map[string]string
		newLabels map[string]string
		want      bool
	}{
		{
			name:      "no element labels on either",
			oldLabels: map[string]string{},
			newLabels: map[string]string{},
			want:      false,
		},
		{
			name:      "element-1-id added",
			oldLabels: map[string]string{},
			newLabels: map[string]string{"run.ai/element-1-id": "abc"},
			want:      true,
		},
		{
			name:      "element-2-id added",
			oldLabels: map[string]string{"run.ai/element-1-id": "abc"},
			newLabels: map[string]string{"run.ai/element-1-id": "abc", "run.ai/element-2-id": "def"},
			want:      true,
		},
		{
			name:      "element-1-id changed",
			oldLabels: map[string]string{"run.ai/element-1-id": "old"},
			newLabels: map[string]string{"run.ai/element-1-id": "new"},
			want:      true,
		},
		{
			name:      "element labels unchanged",
			oldLabels: map[string]string{"run.ai/element-1-id": "abc"},
			newLabels: map[string]string{"run.ai/element-1-id": "abc"},
			want:      false,
		},
		{
			name:      "element-1-id removed",
			oldLabels: map[string]string{"run.ai/element-1-id": "abc"},
			newLabels: map[string]string{},
			want:      true,
		},
		{
			name:      "element-3-id added at level 3",
			oldLabels: map[string]string{"run.ai/element-1-id": "a", "run.ai/element-2-id": "b"},
			newLabels: map[string]string{"run.ai/element-1-id": "a", "run.ai/element-2-id": "b", "run.ai/element-3-id": "c"},
			want:      true,
		},
		{
			name:      "all 3 levels unchanged",
			oldLabels: map[string]string{"run.ai/element-1-id": "a", "run.ai/element-2-id": "b", "run.ai/element-3-id": "c"},
			newLabels: map[string]string{"run.ai/element-1-id": "a", "run.ai/element-2-id": "b", "run.ai/element-3-id": "c"},
			want:      false,
		},
		{
			name:      "element-4-id does not trigger (beyond MaxStoredElementLevels)",
			oldLabels: map[string]string{},
			newLabels: map[string]string{"run.ai/element-4-id": "should-be-ignored"},
			want:      false,
		},
		{
			name:      "element-2-id changed",
			oldLabels: map[string]string{"run.ai/element-1-id": "a", "run.ai/element-2-id": "old"},
			newLabels: map[string]string{"run.ai/element-1-id": "a", "run.ai/element-2-id": "new"},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasElementIdLabelAdded(tt.oldLabels, tt.newLabels)
			if got != tt.want {
				t.Errorf("hasElementIdLabelAdded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterUpdateEvents_ElementLabels(t *testing.T) {
	tests := []struct {
		name string
		old  *corev1.Pod
		new  *corev1.Pod
		want bool
	}{
		{
			name: "element label added triggers reconcile",
			old:  makeRunaiPod(map[string]string{}, nil),
			new:  makeRunaiPod(map[string]string{"run.ai/element-1-id": "elem-uuid"}, nil),
			want: true,
		},
		{
			name: "no changes does not trigger reconcile",
			old:  makeRunaiPod(map[string]string{"run.ai/element-1-id": "elem-uuid"}, nil),
			new:  makeRunaiPod(map[string]string{"run.ai/element-1-id": "elem-uuid"}, nil),
			want: false,
		},
		{
			name: "non-runai pod does not trigger reconcile",
			old: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}},
				Spec:       corev1.PodSpec{SchedulerName: "default-scheduler"},
			},
			new: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"run.ai/element-1-id": "abc"}},
				Spec:       corev1.PodSpec{SchedulerName: "default-scheduler"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterUpdateEvents(tt.old, tt.new)
			if got != tt.want {
				t.Errorf("filterUpdateEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}
