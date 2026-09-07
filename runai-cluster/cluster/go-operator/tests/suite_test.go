/*
 * SPDX-FileCopyrightText: Copyright (c) 2022 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"testing"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"

	runcontrollers "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/controllers/run"
	utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8s            kc.HybridClient
	testEnv        *envtest.Environment
	apiCtx         context.Context
	mgr            manager.Manager
	managerContext context.Context
	runaiConfig    *runaiv1.RunaiConfig

	suite = "Integration Tests"
)

func skipTestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, suite)
}

var _ = BeforeSuite(func() {
	testEnv, k8s, apiCtx = utils.EnvironmentSetup(suite)

	var err error
	mgr, err = ctrl.NewManager(testEnv.Config, ctrl.Options{
		Scheme: k8s.Scheme(),
	})
	Expect(err).NotTo(HaveOccurred())

	runaiConfigController := runcontrollers.NewRunAIConfigReconciler(kc.New(k8s, k8s, nil), nil, nil, nil, 0, false)

	managerContext = ctrl.SetupSignalHandler()

	err = runaiConfigController.SetupWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()

		err := mgr.Start(managerContext)
		Expect(err).NotTo(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {
	if testEnv == nil {
		return
	}
	utils.EnvironmentTeardown(testEnv, suite)

	// Stop the controller
	cancelableContext, cancelFunc := context.WithCancel(managerContext)
	cancelFunc()

	// Wait for the controller to stop
	Eventually(func() bool {
		return cancelableContext.Err() == context.Canceled
	}, 5*utils.TimeoutUnit).Should(BeTrue(), "timed out waiting for controller to stop")
})

var _ = BeforeEach(func() {
	runaiConfig = CreateTestingRunaiConfig(k8s)
	PollForTrueCondition(k8s, runaiv1.ConditionTypeDeployed, runaiConfig)
})

var _ = AfterEach(func() {
	DeleteAndPollUntilDeleted(k8s, runaiConfig)
})
