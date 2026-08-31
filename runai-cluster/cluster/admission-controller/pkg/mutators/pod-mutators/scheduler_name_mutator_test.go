/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"
)

func TestSchedulerNameMutator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: testNamespaceName,
		},
		Spec: v1.PodSpec{SchedulerName: "default-scheduler"},
	}

	m := mock.NewMockClient(ctrl)
	mutator := createMutator(testNamespaceName, false)
	mutator.client = m

	err := mutator.mutateSchedulerName(&pod, nil)
	assert.Nil(t, err)
	assert.Equal(t, types.RunaiScheduler, pod.Spec.SchedulerName)
}
