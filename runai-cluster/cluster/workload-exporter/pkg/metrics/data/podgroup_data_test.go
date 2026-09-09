/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package data

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testQueueName   = "test-queue"
	testProjectName = "test-project"
)

func TestGetProjectNameForPodGroup(t *testing.T) {
	for _, tc := range []struct {
		name        string
		queueLabels map[string]string
		want        string
		wantErr     bool
	}{
		{
			name:        "project label present",
			queueLabels: map[string]string{constants.RunaiProjectLabel: testProjectName},
			want:        testProjectName,
		},
		// runai/queue is written on the project namespace, never on the queue, so a
		// queue carrying only it must not resolve to a project.
		{
			name:        "only namespace queue label present",
			queueLabels: map[string]string{constants.RunaiQueueLabel: testProjectName},
			wantErr:     true,
		},
		{
			name:        "no labels",
			queueLabels: map[string]string{},
			wantErr:     true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			if err := kaiv2.AddToScheme(scheme); err != nil {
				t.Fatalf("add kaiv2 to scheme: %v", err)
			}

			queue := &kaiv2.Queue{
				ObjectMeta: metav1.ObjectMeta{Name: testQueueName, Labels: tc.queueLabels},
			}
			podGroup := &kaiv2alpha2.PodGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "test-pod-group", Namespace: "test-namespace"},
				Spec:       kaiv2alpha2.PodGroupSpec{Queue: testQueueName},
			}
			readerClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(queue).Build()

			got, err := GetProjectNameForPodGroup(context.Background(), podGroup, readerClient)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got project name %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got project name %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetProjectNameForPodGroupMissingQueue(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := kaiv2.AddToScheme(scheme); err != nil {
		t.Fatalf("add kaiv2 to scheme: %v", err)
	}

	podGroup := &kaiv2alpha2.PodGroup{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod-group", Namespace: "test-namespace"},
		Spec:       kaiv2alpha2.PodGroupSpec{Queue: testQueueName},
	}
	readerClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	if _, err := GetProjectNameForPodGroup(context.Background(), podGroup, readerClient); err == nil {
		t.Fatal("expected an error when the queue does not exist")
	}
}
