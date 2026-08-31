/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_isKubeflowMasterRole(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"pytorch master pod",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"project":                             "team-a",
							"run.ai/top-owner-uid":                "4212dd85-251f-422a-8694-8b6193a04305",
							"training.kubeflow.org/job-name":      "job-01e78551973b",
							"training.kubeflow.org/job-role":      "master",
							"training.kubeflow.org/operator-name": "pytorchjob-controller",
							"training.kubeflow.org/replica-index": "0",
							KubeflowReplicaTypeLabel:              KubeflowReplicaMasterPytorch,
						},
					},
				},
			},
			true,
		},
		{
			"pytorch worker pod",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"project":                             "team-a",
							"run.ai/top-owner-uid":                "4212dd85-251f-422a-8694-8b6193a04305",
							"training.kubeflow.org/job-name":      "job-01e78551973b",
							"training.kubeflow.org/operator-name": "pytorchjob-controller",
							"training.kubeflow.org/replica-index": "0",
							KubeflowReplicaTypeLabel:              "worker",
						},
					},
				},
			},
			false,
		},
		{
			"random pod",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"project": "team-a",
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isKubeflowMasterRole(tt.args.pod), "isKubeflowMasterRole(%v)", tt.args.pod)
		})
	}
}

func Test_isRayMasterPod(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"ray master pod",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/created-by": "kuberay-operator",
							"app.kubernetes.io/name":       "kuberay",
							"project":                      "team-a",
							"ray.io/cluster":               "raycluster-standard",
							"ray.io/group":                 "headgroup",
							"ray.io/identifier":            "raycluster-standard-head",
							"ray.io/is-ray-node":           "yes",
							RayPodTypeLabel:                RayMasterPodType,
						},
					},
				},
			},
			true,
		},
		{
			"ray worker pod",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app.kubernetes.io/created-by": "kuberay-operator",
							"app.kubernetes.io/name":       "kuberay",
							"project":                      "team-a",
							"ray.io/cluster":               "raycluster-standard",
							"ray.io/group":                 "small-group",
							"ray.io/identifier":            "raycluster-standard-worker",
							"ray.io/is-ray-node":           "yes",
							RayPodTypeLabel:                "worker",
						},
					},
				},
			},
			false,
		},
		{
			"random pod",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"project": "team-a",
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isRayMasterPod(tt.args.pod), "isRayMasterPod(%v)", tt.args.pod)
		})
	}
}

func Test_isMpiOperatorMaster(t *testing.T) {
	type args struct {
		pod v1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"mpi operator master pod - new version",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"controller-uid":                      "e7ca922a-c0db-4696-b31e-9a5adcaa6e3a",
							"job-name":                            "pi-launcher",
							"project":                             "team-a",
							"training.kubeflow.org/job-name":      "pi",
							MpiOperatorJobRoleLabel:               MpiOperatorMasterJobRole,
							"training.kubeflow.org/operator-name": "mpi-operator",
						},
					},
				},
			},
			true,
		},
		{
			"mpi operator worker pod - new version",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"project":                             "team-a",
							"training.kubeflow.org/job-name":      "pi",
							MpiOperatorJobRoleLabel:               "worker",
							"training.kubeflow.org/operator-name": "mpi-operator",
							"training.kubeflow.org/replica-index": "1",
						},
					},
				},
			},
			false,
		},
		{
			"mpi operator master pod - old version",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"controller-uid":         "e7ca922a-c0db-4696-b31e-9a5adcaa6e3a",
							"job-name":               "pi-launcher",
							"project":                "team-a",
							MpiOperatorRoleTypeLabel: MpiOperatorRoleTypeMaster,
						},
					},
				},
			},
			true,
		},
		{
			"mpi operator worker pod - old version",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"project":                        "team-a",
							"training.kubeflow.org/job-name": "pi",
							MpiOperatorRoleTypeLabel:         "worker",
						},
					},
				},
			},
			false,
		},
		{
			"random pod",
			args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"project": "team-a",
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, isMpiOperatorMaster(tt.args.pod), "isMpiOperatorMaster(%v)", tt.args.pod)
		})
	}
}
