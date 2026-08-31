/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package status_info

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestResponseStatusInfo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status info suite")
}

// create a fake k8s client which holds a set of policies
func NewFakeClient(t *testing.T, initObjs ...runtime.Object) client.WithWatch {
	return fake.NewClientBuilder().WithRuntimeObjects(initObjs...).Build()
}
