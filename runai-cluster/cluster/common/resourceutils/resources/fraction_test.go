/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

func Test_getGPUFraction(t *testing.T) {
	tests := []struct {
		name    string
		pod     *v1.Pod
		want    float64
		wantErr bool
	}{
		{
			"No GPU Fraction annotation",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			0,
			true,
		},
		{
			"GPU Fraction annotation with valid value",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{constants.AnnotationGpuFraction: "0.4"},
				},
			},
			0.4,
			false,
		},
		{
			"GPU Fraction annotation with invalid value",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{constants.AnnotationGpuFraction: "abc"},
				},
			},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGPUFraction(tt.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGPUFraction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetGPUFraction() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getNumGPUFractionDevices(t *testing.T) {
	tests := []struct {
		name    string
		pod     *v1.Pod
		want    int64
		wantErr bool
	}{
		{
			"No Num GPU Fraction annotation",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			0,
			true,
		},
		{
			"Num  GPU fraction annotation with valid value",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{constants.AnnotationGpuFractionDevices: "3"},
				},
			},
			3,
			false,
		},
		{
			"Num  GPU fraction annotation with invalid value",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{constants.AnnotationGpuFractionDevices: "abc"},
				},
			},
			0,
			true,
		},
		{
			"Num  GPU fraction annotation with a non-integer value",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{constants.AnnotationGpuFractionDevices: "1.5"},
				},
			},
			0,
			true,
		},
		{
			"GPU fraction annotation without num GPU fraction annotation",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{constants.AnnotationGpuFraction: "0.1"},
				},
			},
			1,
			false,
		},
		{
			"GPU memory annotation without num GPU fraction annotation",
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{constants.AnnotationGpuMemory: "123"},
				},
			},
			1,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNumGPUFractionDevices(tt.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGPUFraction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetGPUFraction() got = %v, want %v", got, tt.want)
			}
		})
	}
}
