/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"flag"
	"testing"

	"github.com/rs/zerolog"
	"go.uber.org/zap/zapcore"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	pgaconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/controllers/podgroup"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/scheme"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.
//
// The suite runs entirely against a fake (in-memory) controller-runtime client
// rather than envtest. There is no controller manager and no cache firing
// events, so Reconcile is invoked manually via the reconcile() helper at each
// point where, in production, a watch would have triggered it. This makes the
// suite a plain `go test` package (see the Makefile) and fully deterministic.
const (
	suite = "PodGrouper Tests"

	RunaiProjectLabel = "project"
	RunaiQueueLabel   = "runai/queue"

	testNodePoolLabelKey    = "test/node-pool"
	testDefaultNodePoolName = "default"
	testUnexistingNodePool  = "test-unexisting-node-pool"
)

var (
	// k8sClient, cachedClient and basicK8sClient all point at the same fake
	// client; the three names are kept so the existing helpers read naturally.
	k8sClient      client.Client
	cachedClient   client.Client
	basicK8sClient client.Client
	apiCtx         context.Context
	reconciler     *podgroup.PodGroupReconciler
	// The run.ai/v2 Project kind is only ever written by these tests, never read by
	// the service, so it is registered here rather than in the production scheme.
	scheme = testScheme()
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, suite)
}

func initLogging() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	opts := zap.Options{
		Development:     true,
		StacktraceLevel: zapcore.LevelEnabler(zapcore.FatalLevel),
		DestWriter:      GinkgoWriter,
		TimeEncoder:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
}

func testSuiteSetup() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	initLogging()

	pgaconfig.SetForTest(pgaconfig.PodGroupAssignerConfig{
		NodePoolLabelKey:           testNodePoolLabelKey,
		QueueLabelKey:              RunaiQueueLabel,
		NamespaceProjectLabelKey:   RunaiQueueLabel,
		ProjectLabelKey:            RunaiProjectLabel,
		UnexistingNodepoolSentinel: testUnexistingNodePool,
		DefaultNodepoolName:        testDefaultNodePoolName,
	})

	apiCtx = context.Background()

	// Status subresources must be declared so that spec-level Update() calls
	// (made by the assigner) don't clobber status, and the scheduler mock's
	// Status().Update() on PodGroups/Pods persists. The pod field index mirrors
	// the one the manager installs in production, so the assigner's pod lookup
	// by pod-group works against the fake client.
	fakeClient := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&kaiv2alpha2.PodGroup{}, &corev1.Pod{}, &v1alpha1.NodePool{}).
		WithIndex(&corev1.Pod{}, common.PodByPodGroupIndexerName, podgroup.PodByPodGroupIndexer).
		Build()

	k8sClient = fakeClient
	cachedClient = fakeClient
	basicK8sClient = fakeClient
}

// controllerSetup (re)creates the reconciler, resetting the assigner's per-UID
// round-robin cache between the suite's Describe-level setups.
func controllerSetup() {
	reconciler = podgroup.NewPodGroupReconciler(k8sClient)
}

// reconcile drives a single manual reconciliation of the given pod group,
// standing in for the event the controller manager's watch would have
// delivered in the envtest-based suite. Errors are intentionally swallowed:
// in production Reconcile requeues on error (e.g. a pod group with no pods),
// and tests assert the resulting object state rather than the return value.
func reconcile(pg types.NamespacedName) {
	_, _ = reconciler.Reconcile(apiCtx, ctrl.Request{NamespacedName: pg})
}

var _ = BeforeSuite(testSuiteSetup)

// testScheme returns the production scheme; no extra kinds are needed.
func testScheme() *runtime.Scheme {
	return scheme_init.Scheme()
}
