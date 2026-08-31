/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package testing holds the fake-client test helpers that used to live in
// builders/common.
//
// They are kept out of builders/common because registering the kai.resources
// scheme imports the kai.resources API package, which pulls KAI-scheduler's
// pkg/apis/kai/v1 and from there its scheduler internals - and there is
// controller-runtime mismatch with the control plane services.
// They must not compile any of it. Test helpers have no business in that import
// graph anyway.
package testing

import (
	"testing"

	"github.com/stretchr/testify/require"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	dinfp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferencepolicy/v1alpha1"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	dpv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	ipp1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	ipv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
)

// NewFakeClient creates a fake k8s client
func NewFakeClient(t *testing.T, initObjs ...runtime.Object) client.Client {
	err := ipv1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = ipp1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = dpv1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = runaiv2.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = kaiv1alpha1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = dinfw.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = dinfp.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	err = schedulingv1.AddToScheme(scheme.Scheme)
	require.Nil(t, err)

	initObjs = addBasicPriorityClassesToFakeClient(initObjs)

	return fake.NewClientBuilder().WithStatusSubresource(&kaiv1alpha1.NodePool{}).WithRuntimeObjects(initObjs...).Build()
}

func addBasicPriorityClassesToFakeClient(objects []runtime.Object) []runtime.Object {
	priorityClasses := []runtime.Object{
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.InteractivePreemptiblePriorityClass,
			},
			Value: 75,
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: constants.InteractivePriorityClass,
			},
			Value: 100,
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "low",
			},
			Value: 50, // doesn't matter for the purpose of the test
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "high",
			},
			Value: 125,
		},
		&schedulingv1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "medium",
			},
			Value: 75,
		},
	}

	return append(objects, priorityClasses...)
}
