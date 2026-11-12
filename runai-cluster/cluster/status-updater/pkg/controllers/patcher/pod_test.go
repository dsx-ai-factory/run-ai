/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patcher

import (
	"context"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestShouldUpdatePodAnnotations(t *testing.T) {
	type args struct {
		pod        *v1.Pod
		podMetrics metadata.PodMetadata
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Pod no annotations",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
					},
				},
				metadata.PodMetadata{},
			},
			true,
		},
		{
			"Pod already updated",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							PodStatusAnnotationName:             "Pending",
							PodAllocatedGPUsAnnotationName:      "0",
							PodAllocatedGpuMemoryAnnotationName: "0",
							PodAllocatedMigGPUsAnnotationName:   "0",
						},
					},
				},
				metadata.PodMetadata{
					Status:    "Pending",
					GPUs:      0,
					GpuMemory: 0,
					MigGPUs:   0,
				},
			},
			false,
		},
		{
			"Pod has some annotations updated",
			args{
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							PodStatusAnnotationName:             "Pending",
							PodAllocatedGPUsAnnotationName:      "0",
							PodAllocatedGpuMemoryAnnotationName: "0",
							PodAllocatedMigGPUsAnnotationName:   "0",
						},
					},
				},
				metadata.PodMetadata{
					Status:    "Running",
					GPUs:      0,
					GpuMemory: 0,
					MigGPUs:   0,
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldUpdatePodAnnotations(tt.args.pod, tt.args.podMetrics); got != tt.want {
				t.Errorf("ShouldUpdatePodAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdatePodAnnotations(t *testing.T) {
	type args struct {
		ctx        context.Context
		pod        *v1.Pod
		podMetrics metadata.PodMetadata
		kubeClient client.Client
	}
	tests := []struct {
		name                string
		args                args
		wantErr             bool
		expectedAnnotations map[string]string
	}{
		{
			"Update annotations on an empty pod",
			args{
				context.TODO(),
				&v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "t1",
						Namespace: "namespaceT1",
					},
				},
				metadata.PodMetadata{
					Status:    "Pending",
					GPUs:      0,
					GpuMemory: 0,
					MigGPUs:   0,
				},
				fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).Build(),
			},
			false,
			map[string]string{
				PodStatusAnnotationName:             "Pending",
				PodAllocatedGPUsAnnotationName:      "0",
				PodAllocatedGpuMemoryAnnotationName: "0",
				PodAllocatedMigGPUsAnnotationName:   "0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.args.kubeClient.Create(tt.args.ctx, tt.args.pod)
			if err != nil {
				t.Fatal(err)
			}

			if err = UpdatePodAnnotations(tt.args.ctx, tt.args.pod, tt.args.podMetrics,
				tt.args.kubeClient); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePodAnnotations() error = %v, wantErr %v", err, tt.wantErr)
			}

			patchedPod := &v1.Pod{}
			err = tt.args.kubeClient.Get(tt.args.ctx,
				client.ObjectKey{Namespace: tt.args.pod.Namespace, Name: tt.args.pod.Name},
				patchedPod)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, patchedPod.Annotations, tt.expectedAnnotations)
		})
	}
}

func Test_transformToPodAnnotations(t *testing.T) {
	type args struct {
		podMetrics metadata.PodMetadata
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			"Basic metadata annotations",
			args{
				metadata.PodMetadata{
					Status:    "Pending",
					GPUs:      0,
					GpuMemory: 0,
					MigGPUs:   0,
				},
			},
			map[string]string{
				PodStatusAnnotationName:             "Pending",
				PodAllocatedGPUsAnnotationName:      "0",
				PodAllocatedGpuMemoryAnnotationName: "0",
				PodAllocatedMigGPUsAnnotationName:   "0",
			},
		},
		{
			"Basic metadata annotations - fractions",
			args{
				metadata.PodMetadata{
					Status:    "Running",
					GPUs:      0.5001,
					GpuMemory: 0,
					MigGPUs:   0,
				},
			},
			map[string]string{
				PodStatusAnnotationName:             "Running",
				PodAllocatedGPUsAnnotationName:      "0.5001",
				PodAllocatedGpuMemoryAnnotationName: "0",
				PodAllocatedMigGPUsAnnotationName:   "0",
			},
		},
		{
			"Basic metadata annotations - fraction memory",
			args{
				metadata.PodMetadata{
					Status:    "Running",
					GPUs:      0.1,
					GpuMemory: 1000,
					MigGPUs:   0,
				},
			},
			map[string]string{
				PodStatusAnnotationName:             "Running",
				PodAllocatedGPUsAnnotationName:      "0.1",
				PodAllocatedGpuMemoryAnnotationName: "1000",
				PodAllocatedMigGPUsAnnotationName:   "0",
			},
		},
		{
			"Basic metadata annotations - migs",
			args{
				metadata.PodMetadata{
					Status:    "Running",
					GPUs:      0,
					GpuMemory: 0,
					MigGPUs:   0.275,
				},
			},
			map[string]string{
				PodStatusAnnotationName:             "Running",
				PodAllocatedGPUsAnnotationName:      "0",
				PodAllocatedGpuMemoryAnnotationName: "0",
				PodAllocatedMigGPUsAnnotationName:   "0.275",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transformToPodAnnotations(tt.args.podMetrics); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transformToPodAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestUpdatePodStatusAnnotation checks that only the status annotation is patched and that other
// annotations on the pod are kept as they are.
func TestUpdatePodStatusAnnotation(t *testing.T) {
	ctx := context.TODO()
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "t1",
			Namespace:   "namespaceT1",
			Annotations: map[string]string{PodStatusAnnotationName: "Pending", "other": "value"},
		},
	}
	kubeClient := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(pod.DeepCopy()).Build()

	err := UpdatePodStatusAnnotation(ctx, pod, "Running", kubeClient)
	assert.NoError(t, err)

	patchedPod := &v1.Pod{}
	err = kubeClient.Get(ctx, client.ObjectKeyFromObject(pod), patchedPod)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{PodStatusAnnotationName: "Running", "other": "value"},
		patchedPod.Annotations)
}
