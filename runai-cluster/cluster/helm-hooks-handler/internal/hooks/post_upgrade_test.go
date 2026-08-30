package hooks_test

import (
	"context"
	"time"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	schemeInit "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	ra "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	hooks "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	testu "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/test-util"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	schedv2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2"
	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	post_hooks "github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/hooks/post-hooks"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("PostUpgrade Queue Version Validation", func() {
	var (
		resources *clients.Clients
		ctx       context.Context

		runaiconfigObj *unstructured.Unstructured
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient := fake.NewClientBuilder().WithScheme(schemeInit.Scheme()).Build()
		dClient := testu.NewTestDynamicClient(false, false)
		resources = &clients.Clients{
			LocalClusterClient:         k8sClient,
			LocalClusterDynamicClient:  dClient,
			RemoteClusterClient:        k8sClient,
			RemoteClusterDynamicClient: dClient,
		}

		// create the runaiconfig + its namespace
		runaiconfigObj = testu.AddYamlRunaiConf(resources.LocalClusterDynamicClient, "../test-util/test_resources/runaiconfig2_22_nodepool_controller_disabled.yaml")
	})

	It("should fail if queues are not updated in time", func() {
		// 1 queue with correct ParentQueue and 1 queue with missing ParentQueue
		queues := []*schedv2.Queue{buildQueueObject("queue1", "parent1"), buildQueueObject("queue2", "")}

		// Add a queue with missing ParentQueue (should trigger waiting)
		for _, queue := range queues {
			Expect(resources.RemoteClusterClient.Create(ctx, queue)).To(Succeed())
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
		defer cancel()

		err := post_hooks.HandleQueueVersionUpgrade(timeoutCtx, resources, 100)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("timeout reached"))

		dynamicResource := resources.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
		val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
		spec := val.Object["spec"].(map[string]interface{})
		runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
		enabled, ok := runaiNPCConf["enabled"].(bool)
		Expect(ok).To(BeTrue())
		Expect(enabled).To(BeFalse(), "nodepool-controller should be disabled when queues are not updated in time")
	})

	It("should succeed and enable nodepool-controller after all queues are updated", func() {
		queues := []*schedv2.Queue{buildQueueObject("queue1", "parent1"), buildQueueObject("queue2", "parent2")}

		// Add a queues with correct ParentQueue
		for _, queue := range queues {
			Expect(resources.RemoteClusterClient.Create(ctx, queue)).To(Succeed())
		}

		timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
		defer cancel()

		err := post_hooks.HandleQueueVersionUpgrade(timeoutCtx, resources, 100)
		Expect(err).ToNot(HaveOccurred())

		dynamicResource := resources.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
		val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
		spec := val.Object["spec"].(map[string]interface{})
		runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
		enabled, ok := runaiNPCConf["enabled"].(bool)
		Expect(ok).To(BeTrue())
		Expect(enabled).To(BeTrue(), "nodepool-controller should be enabled after all queues are updated")
	})

	It("no queues - should succeed and enable nodepool-controller", func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
		defer cancel()

		err := post_hooks.HandleQueueVersionUpgrade(timeoutCtx, resources, 100)
		Expect(err).ToNot(HaveOccurred())

		dynamicResource := resources.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
		val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
		spec := val.Object["spec"].(map[string]interface{})
		runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
		enabled, ok := runaiNPCConf["enabled"].(bool)
		Expect(ok).To(BeTrue())
		Expect(enabled).To(BeTrue(), "nodepool-controller should be enabled (no queues)")
	})

	It("post-upgrade hook should enable nodepool-controller", func() {
		timeoutCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
		defer cancel()

		// Engine-to-KAI migration disables the nodepool controller before the upgrade, so post-upgrade must re-enable it.
		hook := post_hooks.NewPostUpgradeHook(&hooks.ClusterHookParams{QueueValidationIntervalMillis: 100})
		Expect(hook.Call(timeoutCtx, resources)).To(Succeed())

		dynamicResource := resources.LocalClusterDynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
		val, _ := dynamicResource.Get(ctx, runaiconfigObj.GetName(), metav1.GetOptions{})
		spec := val.Object["spec"].(map[string]interface{})
		runaiNPCConf := spec["nodepool-controller"].(map[string]interface{})
		enabled, ok := runaiNPCConf["enabled"].(bool)
		Expect(ok).To(BeTrue())
		Expect(enabled).To(BeTrue(), "post-upgrade hook should enable nodepool-controller")
	})

	It("cleans up legacy Engine resources only on the remote cluster", func() {
		newEngineConfig := func() *unstructured.Unstructured {
			engineConfig := &unstructured.Unstructured{}
			engineConfig.SetGroupVersionKind(schema.GroupVersionKind{
				Group: "engine.run.ai", Version: "v1alpha1", Kind: "Config",
			})
			engineConfig.SetName("engine-config")
			return engineConfig
		}
		localEngineConfig := newEngineConfig()
		Expect(resources.LocalClusterClient.Create(ctx, localEngineConfig)).To(Succeed())

		remoteClient := fake.NewClientBuilder().WithScheme(schemeInit.Scheme()).WithObjects(
			&kaiv1.Config{
				ObjectMeta: metav1.ObjectMeta{Name: "kai-config"},
				Status: kaiv1.ConfigStatus{Conditions: []metav1.Condition{{
					Type:   string(kaiv1.ConditionTypeAvailable),
					Status: metav1.ConditionTrue,
				}}},
			},
			newEngineConfig(),
			&monitoringv1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{
				Name:      "nodepool-controller-accounting",
				Namespace: runaiv1.DefaultNamespace,
			}},
		).Build()
		resources.RemoteClusterClient = remoteClient

		timeoutCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		hook := post_hooks.NewPostUpgradeHook(&hooks.ClusterHookParams{
			EngineToKAIMigrationEnabled:   true,
			QueueValidationIntervalMillis: 1,
		})
		Expect(hook.Call(timeoutCtx, resources)).To(Succeed())

		Expect(remoteClient.Get(ctx, client.ObjectKey{Name: "engine-config"}, newEngineConfig())).
			To(MatchError(apierrors.IsNotFound, "be a not-found error"))
		Expect(remoteClient.Get(ctx, client.ObjectKey{
			Name: "nodepool-controller-accounting", Namespace: runaiv1.DefaultNamespace,
		}, &monitoringv1.ServiceMonitor{})).To(MatchError(apierrors.IsNotFound, "be a not-found error"))
		Expect(resources.LocalClusterClient.Get(ctx, client.ObjectKey{Name: "engine-config"}, newEngineConfig())).
			To(Succeed())
	})
})

func buildQueueObject(name, parentQueue string) *schedv2.Queue {
	return &schedv2.Queue{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{constants.RunaiProjectLabel: "test"},
		},
		Spec: schedv2.QueueSpec{
			ParentQueue: parentQueue,
		},
	}
}
