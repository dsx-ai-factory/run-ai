/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestSumResources(t *testing.T) {
	type args struct {
		left  v1.ResourceList
		right v1.ResourceList
	}
	tests := []struct {
		name string
		args args
		want v1.ResourceList
	}{
		{
			"Add same single resource",
			args{
				v1.ResourceList{"r1": resource.MustParse("1")},
				v1.ResourceList{"r1": resource.MustParse("1")},
			},
			v1.ResourceList{"r1": resource.MustParse("2")},
		},
		{
			"Add whole to fraction",
			args{
				v1.ResourceList{"r1": resource.MustParse("1")},
				v1.ResourceList{"r1": resource.MustParse("500m")},
			},
			v1.ResourceList{"r1": resource.MustParse("1500m")},
		},
		{
			"Add multiple resources",
			args{
				v1.ResourceList{"r1": resource.MustParse("1"), "r2": resource.MustParse("2")},
				v1.ResourceList{"r1": resource.MustParse("1"), "r3": resource.MustParse("3")},
			},
			v1.ResourceList{"r1": resource.MustParse("2"), "r2": resource.MustParse("2"),
				"r3": resource.MustParse("3")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SumResources(tt.args.left, tt.args.right)
			err := EqualResourceLists(got, tt.want)
			if err != nil {
				t.Errorf("SumResources() = \nGot:\n%v\nWant:\n%v, err:\n%v", got, tt.want, err)
			}
		})
	}
}

func Test_GetExceedingResources(t *testing.T) {
	type test struct {
		name              string
		request           v1.ResourceList
		capacity          v1.ResourceList
		expectedExceeding v1.ResourceList
	}

	tests := []test{
		{
			name: "no exceeding resource",
			request: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("1"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("2"),
			},
			expectedExceeding: v1.ResourceList{},
		},
		{
			name: "exceeding resource",
			request: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("3"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("2"),
			},
			expectedExceeding: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("1"),
			},
		},
		{
			name: "no exceeding resource - exact match",
			request: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("2"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("2"),
			},
			expectedExceeding: v1.ResourceList{},
		},
		{
			name: "exceeding resources - multiple resources",
			request: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("3"),
				v1.ResourceMemory: resource.MustParse("4Gi"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("2"),
				v1.ResourceMemory: resource.MustParse("3Gi"),
			},
			expectedExceeding: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
		},
		{
			name: "no exceeding resources - multiple resources",
			request: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("2"),
				v1.ResourceMemory: resource.MustParse("2Gi"),
			},
			expectedExceeding: v1.ResourceList{},
		},
		{
			name: "exceeding resources - multiple resources - one resource exact match",
			request: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("2Gi"),
			},
			expectedExceeding: v1.ResourceList{},
		},
		{
			name: "exceeding resources - multiple resources - one resource not in capacity",
			request: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("1"),
			},
			expectedExceeding: v1.ResourceList{
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
		},
		{
			name: "exceeding resources - unlimited",
			request: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("1"),
			},
			capacity: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("-1"),
			},
			expectedExceeding: v1.ResourceList{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExceedingResources(tt.request, tt.capacity)
			err := EqualResourceLists(got, tt.expectedExceeding)
			if err != nil {
				t.Errorf("GetExceedingResources() = \nGot:\n%v\nWant:\n%v, err:\n%v", got, tt.expectedExceeding, err)
			}
		})
	}
}

func TestEqualResourceLists(t *testing.T) {
	type args struct {
		got  v1.ResourceList
		want v1.ResourceList
	}
	tests := []struct {
		name        string
		args        args
		expectEqual bool
	}{
		{
			"Equal resources",
			args{
				v1.ResourceList{"r1": resource.MustParse("1"), "r2": resource.MustParse("2")},
				v1.ResourceList{"r1": resource.MustParse("1"), "r2": resource.MustParse("2")},
			},
			true,
		},
		{
			"Different resources",
			args{
				v1.ResourceList{"r1": resource.MustParse("1"), "r2": resource.MustParse("2")},
				v1.ResourceList{"r1": resource.MustParse("1"), "r2": resource.MustParse("3")},
			},
			false,
		},
		{
			"Different keys",
			args{
				v1.ResourceList{"r1": resource.MustParse("1"), "r2": resource.MustParse("2")},
				v1.ResourceList{"r1": resource.MustParse("1"), "r3": resource.MustParse("2")},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := EqualResourceLists(tt.args.got, tt.args.want)
			if tt.expectEqual {
				if err != nil {
					t.Errorf("Expected equal, got error:\n%v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected not equal, got nil")
				}
			}
		})
	}
}

func TestAddRunaiAnnotationsRequestedResources_GpuFractionScaling(t *testing.T) {
	tests := []struct {
		name           string
		annotations    map[string]string
		wantGpuMilli   int64
		wantGpuPresent bool
	}{
		{
			name:           "no annotations",
			annotations:    map[string]string{},
			wantGpuPresent: false,
		},
		{
			name: "gpu-fraction only, no num-devices annotation",
			annotations: map[string]string{
				constants.AnnotationGpuFraction: "0.5",
			},
			wantGpuMilli:   500,
			wantGpuPresent: true,
		},
		{
			name: "gpu-fraction with num-devices=1 (single device, unscaled)",
			annotations: map[string]string{
				constants.AnnotationGpuFraction:        "0.5",
				constants.AnnotationGpuFractionDevices: "1",
			},
			wantGpuMilli:   500,
			wantGpuPresent: true,
		},
		{
			name: "gpu-fraction 0.5 with num-devices=3 (scaled: 0.5 * 3 = 1.5)",
			annotations: map[string]string{
				constants.AnnotationGpuFraction:        "0.5",
				constants.AnnotationGpuFractionDevices: "3",
			},
			wantGpuMilli:   1500,
			wantGpuPresent: true,
		},
		{
			name: "gpu-fraction 0.7 with num-devices=2 (scaled: 0.7 * 2 = 1.4)",
			annotations: map[string]string{
				constants.AnnotationGpuFraction:        "0.7",
				constants.AnnotationGpuFractionDevices: "2",
			},
			wantGpuMilli:   1400,
			wantGpuPresent: true,
		},
		{
			name: "gpu-fraction with num-devices=0 (treated as single device, unscaled)",
			annotations: map[string]string{
				constants.AnnotationGpuFraction:        "0.5",
				constants.AnnotationGpuFractionDevices: "0",
			},
			wantGpuMilli:   500,
			wantGpuPresent: true,
		},
		{
			name: "gpu-fraction with malformed num-devices annotation falls back to unscaled",
			annotations: map[string]string{
				constants.AnnotationGpuFraction:        "0.5",
				constants.AnnotationGpuFractionDevices: "not-a-number",
			},
			wantGpuMilli:   500,
			wantGpuPresent: true,
		},
		{
			name: "gpu-fraction with empty num-devices annotation falls back to unscaled",
			annotations: map[string]string{
				constants.AnnotationGpuFraction:        "0.5",
				constants.AnnotationGpuFractionDevices: "",
			},
			wantGpuMilli:   500,
			wantGpuPresent: true,
		},
		{
			name: "num-devices without gpu-fraction is ignored",
			annotations: map[string]string{
				constants.AnnotationGpuFractionDevices: "3",
			},
			wantGpuPresent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := v1.ResourceList{}
			limits := v1.ResourceList{}

			AddRunaiAnnotationsRequestedResources(tt.annotations, requests, limits)

			gotReq, reqPresent := requests[constants.NvidiaGpuResourceName]
			gotLim, limPresent := limits[constants.NvidiaGpuResourceName]

			if reqPresent != tt.wantGpuPresent || limPresent != tt.wantGpuPresent {
				t.Fatalf("nvidia.com/gpu presence mismatch: requests=%v, limits=%v, want=%v",
					reqPresent, limPresent, tt.wantGpuPresent)
			}
			if !tt.wantGpuPresent {
				return
			}
			if gotReq.MilliValue() != tt.wantGpuMilli {
				t.Errorf("request nvidia.com/gpu = %dm, want %dm", gotReq.MilliValue(), tt.wantGpuMilli)
			}
			if gotLim.MilliValue() != tt.wantGpuMilli {
				t.Errorf("limit nvidia.com/gpu = %dm, want %dm", gotLim.MilliValue(), tt.wantGpuMilli)
			}
		})
	}
}
