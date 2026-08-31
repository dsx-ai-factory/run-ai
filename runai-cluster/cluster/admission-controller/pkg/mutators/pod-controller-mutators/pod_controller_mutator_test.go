/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_controller_mutators

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/magiconair/properties/assert"

	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/kube-client/mock"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/tests"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/types"
	"github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/utils"

	"k8s.io/api/admission/v1beta1"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMutateAllSuccess(t *testing.T) {
	admRequest := v1beta1.AdmissionRequest{}
	admRequest.UserInfo = authv1.UserInfo{}
	admRequest.UserInfo.Groups = []string{"authz", "elon.musk@run.ai", "kubernetes"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	namespace := v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"runai/queue": testProjectName,
			},
		},
	}

	affinitySlice := []string{"aaa", "bbb", "ccc"}
	project := projectWithNodeAffinity(false, affinitySlice)

	m := mock.NewMockClient(ctrl)
	m.EXPECT().GetNamespace(testProjectName).Return(&namespace, nil)
	m.EXPECT().GetProject(testProjectName).Return(&project, nil)
	m.EXPECT().GetCachedReader().Return(fakeCachedReader()).AnyTimes()

	podController := utils.CreatePodController(testProjectName)
	mutatingHandler := createMutator(testProjectName)
	mutatingHandler.client = m
	err := mutatingHandler.ControllerMutateAll(podController, &admRequest)

	assert.Equal(t, err, nil)
	assert.Equal(t, len(podController.Annotations), 0)
}

func TestNonRunaiScheduler(t *testing.T) {
	admRequest := v1beta1.AdmissionRequest{}
	admRequest.UserInfo = authv1.UserInfo{}
	admRequest.UserInfo.Groups = []string{"authz", "elon.musk@run.ai"}

	podController := utils.CreatePodController(testProjectName)
	podController.Spec.Template.Spec.SchedulerName = "aluf-hashnitzel"

	mutator := createMutator("nil")
	assert.Equal(t, mutator.shouldMutateController(podController), false)
}

func createMutator(namespaceName string) *PodControllerMutator {
	namespace := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	scheme := tests.InitScheme()
	k8sReader := fakerunaiclient.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).Build()
	handlerParams := &types.MutatingParams{}
	k8sClient := tests.NewClient(k8sReader)
	return NewPodControllerMutator(k8sClient, handlerParams)
}
