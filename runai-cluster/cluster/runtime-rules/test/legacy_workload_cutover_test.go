/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"os"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	pkg "github.com/run-ai/runai/runai-cluster/cluster/runtime-rules"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/util"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/validation"
	dw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	tw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func setClusterVersion(version string) func() {
	previous, wasSet := os.LookupEnv("CLUSTER_VERSION")
	Expect(os.Setenv("CLUSTER_VERSION", version)).To(Succeed())
	return func() {
		if wasSet {
			Expect(os.Setenv("CLUSTER_VERSION", previous)).To(Succeed())
			return
		}
		Expect(os.Unsetenv("CLUSTER_VERSION")).To(Succeed())
	}
}

var _ = Describe("runtime-rules legacy workload cutover", Serial, func() {
	newClient := func(functions interceptor.Funcs) client.WithWatch {
		ns := BuildNamespace(testNamespaceName, map[string]string{RunaiQueueLabel: testProjectName})
		projects := BuildProjectFixtures(testProjectName, testNamespaceName, 5, 5)
		return fakerunaiclient.NewClientBuilder().
			WithScheme(initSchemaWithRunner()).
			WithObjects(append(projects, ns)...).
			WithInterceptorFuncs(functions).
			Build()
	}

	newRuntimeRules := func(k8sClient client.WithWatch, disableWorkloadRunnerRules bool) *pkg.RuntimeRules {
		return pkg.NewRuntimeRules(k8sClient, &queuesManager.QueuesManager{}, prometheus.GetQuerierClient(false, ""),
			record.NewFakeRecorder(100), []int{queuesManager.UnknownEvent, queuesManager.DeactivateEvent}, validation.NewWorkloadValidator(), disableWorkloadRunnerRules)
	}

	DescribeTable("always selects WorkloadRunner actions",
		func(version string, expectWorkloadRunner bool) {
			restoreVersion := setClusterVersion(version)
			defer restoreVersion()

			runtimeRules := newRuntimeRules(nil, true)

			if expectWorkloadRunner {
				Expect(runtimeRules.SupportedJobKinds).To(HaveKey(constants.KindWorkloadRunner))
				return
			}
			Expect(runtimeRules.SupportedJobKinds).NotTo(HaveKey(constants.KindWorkloadRunner))
		},
		Entry("below the cutover", "2.99.0", true),
		Entry("at the cutover", "3.0.0", true),
		Entry("after the cutover", "3.1.0", true),
	)

	It("rejects a missing cluster version", func() {
		previous, wasSet := os.LookupEnv("CLUSTER_VERSION")
		Expect(os.Unsetenv("CLUSTER_VERSION")).To(Succeed())
		defer func() {
			if wasSet {
				Expect(os.Setenv("CLUSTER_VERSION", previous)).To(Succeed())
			}
		}()

		Expect(pkg.ValidateClusterVersion()).To(MatchError("env var CLUSTER_VERSION is not set"))
	})

	It("rejects an invalid cluster version", func() {
		restoreVersion := setClusterVersion("not-a-version")
		defer restoreVersion()

		Expect(pkg.ValidateClusterVersion()).To(MatchError(ContainSubstring(`invalid CLUSTER_VERSION value "not-a-version"`)))
	})

	It("does not enumerate legacy workloads on 3.0 while it still enumerates WorkloadRunners when the pre-3 disable flag is set", func() {
		restoreVersion := setClusterVersion("3.0.0")
		defer restoreVersion()

		var mu sync.Mutex
		listCalls := map[string]int{}
		k8sClient := newClient(interceptor.Funcs{
			List: func(ctx context.Context, next client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				mu.Lock()
				switch list.(type) {
				case *tw.TrainingWorkloadList:
					listCalls[constants.KindTrainingWorkload]++
				case *iw.InteractiveWorkloadList:
					listCalls[constants.KindInteractiveWorkload]++
				case *dw.DistributedWorkloadList:
					listCalls[constants.KindDistributedWorkload]++
				case *dromosv1alpha1.WorkloadRunnerList:
					listCalls[constants.KindWorkloadRunner]++
				}
				mu.Unlock()
				return next.List(ctx, list, opts...)
			},
		})

		workloadCache := util.WorkloadCache{}
		runtimeRules := newRuntimeRules(k8sClient, true)
		Expect(runtimeRules.SupportedJobKinds).To(HaveKey(constants.KindWorkloadRunner))
		runtimeRules.InitQueueManager(context.Background(), &workloadCache)

		Expect(listCalls[constants.KindTrainingWorkload]).To(BeZero())
		Expect(listCalls[constants.KindInteractiveWorkload]).To(BeZero())
		Expect(listCalls[constants.KindDistributedWorkload]).To(BeZero())
		Expect(listCalls[constants.KindWorkloadRunner]).To(Equal(1))
	})

	It("does not enumerate legacy workloads below 3.0 and always enumerates WorkloadRunners", func() {
		restoreVersion := setClusterVersion("2.26.0")
		defer restoreVersion()

		var mu sync.Mutex
		listCalls := map[string]int{}
		k8sClient := newClient(interceptor.Funcs{
			List: func(ctx context.Context, next client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				mu.Lock()
				switch list.(type) {
				case *tw.TrainingWorkloadList:
					listCalls[constants.KindTrainingWorkload]++
				case *iw.InteractiveWorkloadList:
					listCalls[constants.KindInteractiveWorkload]++
				case *dw.DistributedWorkloadList:
					listCalls[constants.KindDistributedWorkload]++
				case *dromosv1alpha1.WorkloadRunnerList:
					listCalls[constants.KindWorkloadRunner]++
				}
				mu.Unlock()
				return next.List(ctx, list, opts...)
			},
		})

		workloadCache := util.WorkloadCache{}
		runtimeRules := newRuntimeRules(k8sClient, true)
		Expect(runtimeRules.SupportedJobKinds).To(HaveKey(constants.KindWorkloadRunner))
		runtimeRules.InitQueueManager(context.Background(), &workloadCache)

		Expect(listCalls[constants.KindTrainingWorkload]).To(BeZero())
		Expect(listCalls[constants.KindInteractiveWorkload]).To(BeZero())
		Expect(listCalls[constants.KindDistributedWorkload]).To(BeZero())
		Expect(listCalls[constants.KindWorkloadRunner]).To(Equal(1))
	})

	It("does not start legacy lifecycle or auto-delete watches on 3.0 while WorkloadRunner watches remain active despite the pre-3 disable flag", func() {
		restoreVersion := setClusterVersion("3.0.0")
		defer restoreVersion()

		var mu sync.Mutex
		watchCalls := map[string]int{}
		k8sClient := newClient(interceptor.Funcs{
			Watch: func(ctx context.Context, next client.WithWatch, list client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
				mu.Lock()
				switch list.(type) {
				case *tw.TrainingWorkloadList:
					watchCalls[constants.KindTrainingWorkload]++
				case *iw.InteractiveWorkloadList:
					watchCalls[constants.KindInteractiveWorkload]++
				case *dw.DistributedWorkloadList:
					watchCalls[constants.KindDistributedWorkload]++
				case *dromosv1alpha1.WorkloadRunnerList:
					watchCalls[constants.KindWorkloadRunner]++
				}
				mu.Unlock()
				return next.Watch(ctx, list, opts...)
			},
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go newRuntimeRules(k8sClient, true).Start(ctx)

		Eventually(func() int {
			mu.Lock()
			defer mu.Unlock()
			return watchCalls[constants.KindWorkloadRunner]
		}, time.Second).Should(BeNumerically(">=", 2))

		mu.Lock()
		defer mu.Unlock()
		Expect(watchCalls[constants.KindTrainingWorkload]).To(BeZero())
		Expect(watchCalls[constants.KindInteractiveWorkload]).To(BeZero())
		Expect(watchCalls[constants.KindDistributedWorkload]).To(BeZero())
	})
})
