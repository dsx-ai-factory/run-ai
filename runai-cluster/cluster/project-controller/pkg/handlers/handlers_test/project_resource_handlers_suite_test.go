/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers_test

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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Test Globals
var scheme *runtime.Scheme

// populateGVK backfills the TypeMeta (apiVersion/kind) on an object using the
// test scheme. The controller-runtime fake client strips TypeMeta from objects
// returned by Get/Create, so test assertions that compare TypeMeta against
// objects built in-memory need it restored first.
func populateGVK(obj client.Object) {
	gvks, _, err := scheme.ObjectKinds(obj)
	Expect(err).ToNot(HaveOccurred())
	Expect(gvks).ToNot(BeEmpty())
	obj.GetObjectKind().SetGroupVersionKind(gvks[0])
}

var _ = BeforeSuite(func() {
	scheme = runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv2.AddToScheme(scheme)).Should(Succeed())
	Expect(monitoringv1.AddToScheme(scheme)).Should(Succeed())

	// Set up the package-level config with runai values so production code
	// (which reads via config.Get()) sees the same vocabulary the test data
	// assumes.
	config.SetForTest(test.RunaiConfigForTests())
})

func TestResourceHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	test.RunTest(t, "Project Resource Handlers Tests", "../../../../../bin/test/results/project_resource_handlers_test_results.xml")
}
