/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"testing"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	sdkv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	dp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	ip "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactivepolicy/v2alpha1"
	iw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tp "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingpolicy/v2alpha1"
	tw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	kainodepool "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog"
	"go.uber.org/zap"

	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logging.Init(&logging.Config{LogLevel: zerolog.InfoLevel.String(), Pretty: false})
}

var scheme *runtime.Scheme

var _ = BeforeSuite(func() {
	initLogging()

	scheme = runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).Should(Succeed())
	Expect(v2.AddToScheme(scheme)).Should(Succeed())
	Expect(schedulingv2alpha2.AddToScheme(scheme)).Should(Succeed())
	Expect(rbacv1.AddToScheme(scheme)).Should(Succeed())
	Expect(tp.AddToScheme(scheme)).Should(Succeed())
	Expect(ip.AddToScheme(scheme)).Should(Succeed())
	Expect(dp.AddToScheme(scheme)).Should(Succeed())
	Expect(infp.AddToScheme(scheme)).Should(Succeed())
	Expect(tw.AddToScheme(scheme)).Should(Succeed())
	Expect(iw.AddToScheme(scheme)).Should(Succeed())
	Expect(dw.AddToScheme(scheme)).Should(Succeed())
	Expect(infw.AddToScheme(scheme)).Should(Succeed())
	Expect(kaiv1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(kainodepool.AddToScheme(scheme)).Should(Succeed())
	Expect(kartav1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(dromosv1alpha1.AddToScheme(scheme)).Should(Succeed())
	Expect(sdkv1alpha1.AddToScheme(scheme)).Should(Succeed())
})

func TestAgent(t *testing.T) {
	zap.S().Infof("Started service tests")
	RegisterFailHandler(Fail)
	MockEnv(t)
	RunSpecs(t, "agent service Test Suite")
}
