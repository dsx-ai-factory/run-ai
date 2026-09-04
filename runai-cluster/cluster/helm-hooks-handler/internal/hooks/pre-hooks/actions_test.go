/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"testing"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAddPrivilegedPodSecurityStandardToRunaiNamespace(t *testing.T) {
	runaiNS := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.DefaultNamespace,
		},
	}

	k8s := fake.NewClientBuilder().WithObjects(runaiNS).Build()

	err := addPrivilegedPodSecurityStandardToRunaiNamespace(k8s)
	if err != nil {
		t.Errorf("failed setting privileged pod security label: %v", err)
	}

	ns := &v1.Namespace{}
	err = k8s.Get(context.TODO(), client.ObjectKeyFromObject(runaiNS), ns)
	if err != nil {
		t.Errorf("could not get namesapce: %v", err)
	}

	if len(ns.Labels) == 0 {
		t.Errorf("expected privileged pod security standard label: %v", err)
	}
}
