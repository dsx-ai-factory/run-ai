/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"
)

func TestPodMutator_IsIdempotent(t *testing.T) {
	type fields struct {
		params    *types.MutatingParams
		namespace *v1.Namespace
		project   *unstructured.Unstructured
		nodePool  *unstructured.Unstructured
		draParser *dra.DRAParser
	}
	type args struct {
		pod     *v1.Pod
		request *v1beta1.AdmissionRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "autoscaler - build",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
						Labels: map[string]string{
							"priorityClassName": "build",
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "autoscaler - train",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
						Labels: map[string]string{
							"priorityClassName": "train",
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "inject job envVars - FromPod",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pod",
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "inject job envVars - FromPodOwner",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "my-pod",
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "job-1",
								UID:  "12345",
							},
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "GKE - fraction",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics:       false,
					SpecificClusterType: "gke",
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: testNamespaceName,
						Annotations: map[string]string{
							types.RunaiGpuFraction: "0.1",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "job-1",
								UID:  "12345",
							},
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "GKE - non fraction",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics:       false,
					SpecificClusterType: "gke",
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "my-pod",
						Namespace:   testNamespaceName,
						Annotations: map[string]string{},
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "job-1",
								UID:  "12345",
							},
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "OKE - fraction",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics:       false,
					SpecificClusterType: "oke",
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: testNamespaceName,
						Annotations: map[string]string{
							types.RunaiGpuFraction: "0.1",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "job-1",
								UID:  "12345",
							},
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "OKE - non fraction",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics:       false,
					SpecificClusterType: "oke",
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "my-pod",
						Namespace:   testNamespaceName,
						Annotations: map[string]string{},
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "job-1",
								UID:  "12345",
							},
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "Project label",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNamespaceName,
						Labels: map[string]string{
							"runai/queue": "p2",
						},
					},
				},
				project: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name": "p2",
						},
					},
				},
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-pod",
						Namespace: testNamespaceName,
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "resource mutator with DRA claims",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
				draParser: createDRAParser(testNamespaceName),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespaceName,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{},
								Name:      "my-container",
							},
						},
						ResourceClaims: []v1.PodResourceClaim{
							{
								Name:              "gpu-claim",
								ResourceClaimName: ptr.To("my-gpu-claim"),
							},
						},
						SchedulerName: constants.RunaiSchedulerName,
					},
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "resource mutator with DRA claims and traditional GPU",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
				draParser: createDRAParser(testNamespaceName),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespaceName,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Requests: map[v1.ResourceName]resource.Quantity{
										v1.ResourceCPU:    resource.MustParse("2000m"),
										v1.ResourceMemory: resource.MustParse("4"),
										types.GpuResource: resource.MustParse("1"),
									},
									Limits: v1.ResourceList{
										v1.ResourceCPU:    resource.MustParse("3000m"),
										v1.ResourceMemory: resource.MustParse("5"),
										types.GpuResource: resource.MustParse("1"),
									},
								},
								Name: "my-container",
							},
						},
						ResourceClaims: []v1.PodResourceClaim{
							{
								Name:              "gpu-claim",
								ResourceClaimName: ptr.To("my-gpu-claim"),
							},
						},
						SchedulerName: constants.RunaiSchedulerName,
					},
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "resource mutator with DRA claims nil parser",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespaceName,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{},
								Name:      "my-container",
							},
						},
						ResourceClaims: []v1.PodResourceClaim{
							{
								Name:              "gpu-claim",
								ResourceClaimName: ptr.To("my-gpu-claim"),
							},
						},
						SchedulerName: constants.RunaiSchedulerName,
					},
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "node-pool affinity",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
				nodePool: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"metadata": map[string]interface{}{
							"name": "poolA",
						},
						"spec": map[string]interface{}{
							"labelKey":   getFieldFromName("poolA", "Key"),
							"labelValue": getFieldFromName("poolA", "Value"),
						},
						"status": map[string]interface{}{
							"phase": "Ready",
						},
					},
				},
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{},
						Labels: map[string]string{
							patch.NodePoolNameLabel: "poolA",
						},
					},
					Spec: defaultTestPodSpec(),
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "resource mutator",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespaceName,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Resources: v1.ResourceRequirements{
									Limits: v1.ResourceList{
										types.GpuResource: resource.MustParse("1"),
									},
								},
								Name: "my-container",
							},
						},
						SchedulerName: constants.RunaiSchedulerName,
					},
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "resource mutator no gpu",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: defaultTestNamespace(),
				project:   defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespaceName,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "my-container",
							},
						},
						SchedulerName: constants.RunaiSchedulerName,
					},
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
		{
			name: "enforce runai scheduler",
			fields: fields{
				params: &types.MutatingParams{
					CoreAnalytics: false,
				},
				namespace: &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: testNamespaceName,
						Annotations: map[string]string{
							constants.EnforceSchedulerAnnotationName: "true",
						},
						Labels: map[string]string{
							"runai/queue": testProjectName,
						},
					},
				},
				project: defaultTestProject(),
			},
			args: args{
				pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: testNamespaceName,
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name: "my-container",
							},
						},
					},
				},
				request: &v1beta1.AdmissionRequest{
					Namespace: testNamespaceName,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			pm := createTestMutator(ctrl, tt.fields.params, tt.fields.namespace, tt.fields.project, tt.fields.nodePool)
			if tt.fields.draParser != nil {
				pm.draParser = tt.fields.draParser
			}

			err := pm.MutatePod(tt.args.pod, tt.args.request)
			if err != nil {
				t.Errorf("first MutatePod received an error: %v", err)
			}
			t.Logf("First mutation for test %s", tt.name)
			podAfterFirstMutation := tt.args.pod.DeepCopyObject()

			err = pm.MutatePod(tt.args.pod, tt.args.request)
			if err != nil {
				t.Errorf("second MutatePod received an error: %v", err)
			}
			t.Logf("Second mutation for test %s", tt.name)

			assert.Equal(t, tt.args.pod, podAfterFirstMutation)
		})
	}
}

func createTestMutator(ctrl *gomock.Controller, params *types.MutatingParams,
	namespace *v1.Namespace, project *unstructured.Unstructured, nodePool *unstructured.Unstructured) *PodMutator {
	pm := createMutator(namespace.GetName(), false)
	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(namespace.GetName()).AnyTimes().Return(namespace, nil)
	m.EXPECT().GetRunaiConfigMap().AnyTimes().Return(&v1.ConfigMap{
		Data: map[string]string{
			CpuDefaultRequestGpuFactor:    "2",
			CpuDefaultLimitGpuFactor:      "3",
			MemoryDefaultRequestGpuFactor: "4",
			MemoryDefaultLimitGpuFactor:   "5",
		},
	}, nil)
	m.EXPECT().GetProject(project.GetName()).AnyTimes().Return(project, nil)
	m.EXPECT().GetCachedReader().AnyTimes().Return(fakeCachedReader())
	if nodePool != nil {
		m.EXPECT().GetNodePool(nodePool.GetName()).AnyTimes().Return(nodePool, nil)
	}
	pm.client = m
	pm.params = params
	return pm
}

func defaultTestNamespace() *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespaceName,
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}
}

func defaultTestProject() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": testProjectName,
			},
		},
	}
}

func defaultTestPodSpec() v1.PodSpec {
	return v1.PodSpec{
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: map[v1.ResourceName]resource.Quantity{
						v1.ResourceCPU:    resource.MustParse("2000m"),
						v1.ResourceMemory: resource.MustParse("4"),
						types.GpuResource: resource.MustParse("1"),
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("3000m"),
						v1.ResourceMemory: resource.MustParse("5"),
						types.GpuResource: resource.MustParse("1"),
					},
				},
				Name: "my-container",
			},
		},
		SchedulerName: constants.RunaiSchedulerName,
	}
}
