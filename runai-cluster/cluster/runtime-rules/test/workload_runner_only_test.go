package test

import (
	"context"
	"reflect"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	pkg "github.com/run-ai/runai/runai-cluster/cluster/runtime-rules"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/prometheus"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/queuesManager"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/validation"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

var _ = Describe("runtime-rules uses WorkloadRunner resources only", Serial, func() {
	newClient := func(functions interceptor.Funcs) client.WithWatch {
		ns := BuildNamespace(testNamespaceName, map[string]string{RunaiQueueLabel: testProjectName})
		projects := BuildProjectFixtures(testProjectName, testNamespaceName, 5, 5)
		return fakerunaiclient.NewClientBuilder().
			WithScheme(initSchemaWithRunner()).
			WithObjects(append(projects, ns)...).
			WithInterceptorFuncs(functions).
			Build()
	}

	newRuntimeRules := func(k8sClient client.WithWatch) *pkg.RuntimeRules {
		return pkg.NewRuntimeRules(k8sClient, &queuesManager.QueuesManager{}, prometheus.GetQuerierClient(false, ""),
			record.NewFakeRecorder(100), []int{queuesManager.UnknownEvent, queuesManager.DeactivateEvent}, validation.NewWorkloadValidator())
	}

	It("lists and watches WorkloadRunners without accessing legacy workload resources", func() {
		var mu sync.Mutex
		listCalls := map[string]int{}
		watchCalls := map[string]int{}
		resourceType := func(list client.ObjectList) string {
			return reflect.TypeOf(list).Elem().Name()
		}

		k8sClient := newClient(interceptor.Funcs{
			List: func(ctx context.Context, next client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
				mu.Lock()
				listCalls[resourceType(list)]++
				mu.Unlock()
				return next.List(ctx, list, opts...)
			},
			Watch: func(ctx context.Context, next client.WithWatch, list client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
				mu.Lock()
				watchCalls[resourceType(list)]++
				mu.Unlock()
				return next.Watch(ctx, list, opts...)
			},
		})

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go newRuntimeRules(k8sClient).Start(ctx)

		Eventually(func() int {
			mu.Lock()
			defer mu.Unlock()
			return watchCalls["WorkloadRunnerList"]
		}, time.Second).Should(BeNumerically(">=", 2))

		mu.Lock()
		defer mu.Unlock()
		Expect(listCalls["WorkloadRunnerList"]).To(BeNumerically(">=", 1))
		for _, legacyListType := range []string{"TrainingWorkloadList", "InteractiveWorkloadList", "DistributedWorkloadList"} {
			Expect(listCalls[legacyListType]).To(BeZero(), "unexpected List for %s", legacyListType)
			Expect(watchCalls[legacyListType]).To(BeZero(), "unexpected Watch for %s", legacyListType)
		}
	})
})
