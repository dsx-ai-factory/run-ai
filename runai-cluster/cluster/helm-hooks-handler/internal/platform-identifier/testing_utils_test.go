/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package platform_identifier

import (
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	assert2 "github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

type testUtils struct {
	k8s    client.Client
	assert *assert2.Assertions
}

func newTestingUtils(t *testing.T) *testUtils {
	// If anthing fails in this function, it should call t.Error
	return &testUtils{
		k8s:    kubeClient(),
		assert: assert2.New(t),
	}
}

func kubeClient() client.Client {
	s := scheme_init.Scheme()
	return fake.NewClientBuilder().WithScheme(s).Build()
}

func customResourceDefinition(name string) *extensions.CustomResourceDefinition {
	return &extensions.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func node(name string, labels map[string]string) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}
}
