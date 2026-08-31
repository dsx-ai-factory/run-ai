/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/tests"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testNamespaceName = "runai-proj-1"
	testProjectName   = "proj-1"
)

type ExpectedPatchResults struct {
	ExpectedAnnotations           map[string]string
	ExpectedConfigMapEnvs         map[string]string // Name to configMapKeyRef key
	ExpectedValueEnvs             map[string]string
	ExpectedNumOfConfigMapVolumes int
	ExpectedHostPathVolumes       map[string]string      // HostPathName to Type
	ExpectedVolumeMounts          map[string]VolumeMount // MountPath to VolumeMount
}

type VolumeMount struct {
	Name      string
	SubPath   string
	ReadOnly  bool
	MountPath string
}

func createMutator(namespaceName string, coreAnalytics bool) *PodMutator {
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	scheme := tests.InitScheme()
	k8sReader := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).Build()
	handlerParams := &types.MutatingParams{
		CoreAnalytics: coreAnalytics,
	}
	k8sClient := tests.NewClient(k8sReader)
	return NewPodMutator(k8sClient, handlerParams, nil)
}
