/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

/* Licensed under the Apache 2.0 License. Check apache-2.0.txt for further information */

// Package tests provides Ginkgo-based envtest integration tests for the workload-controller.
// Run with: make envtest && KUBEBUILDER_ASSETS="$(bin/setup-envtest use 1.33 -p path --arch amd64)" go test ./pkg/tests/... -v
// Or use: make envtest-go (from the repo that defines envtest-go and sets KUBEBUILDER_ASSETS).

package tests

import (
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1/testcerts"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	mpiv2beta1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	"testing"
)

func TestTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workload Controller Tests Suite")
}

var _ = BeforeSuite(func() {
	var err error
	webhookCertDir, err = testcerts.GenerateWebhookCertDir()
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func() { _ = os.RemoveAll(webhookCertDir) })
})

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(v1alpha1.AddToScheme(s))
	utilruntime.Must(v2alpha1.AddToScheme(s))
	utilruntime.Must(kubeflowv1.AddToScheme(s)) // PyTorchJob, TFJob, etc.
	utilruntime.Must(mpiv2beta1.AddToScheme(s)) // MPIJob (CRD present in config/crd)
	return s
}
