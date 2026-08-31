/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reconcilers_test

import (
	"testing"

	kaiv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// Test Globals
var scheme *runtime.Scheme

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv2.AddToScheme(scheme)).Should(Succeed())
	Expect(monitoringv1.AddToScheme(scheme)).Should(Succeed())

	config.SetForTest(test.RunaiConfigForTests())
})

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunTest(t, "Reconciler Tests", "../../../bin/test/results/reconciler_test_results.xml")
}
