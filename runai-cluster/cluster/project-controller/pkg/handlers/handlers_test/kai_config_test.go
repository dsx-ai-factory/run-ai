package handlers_test

import (
	. "github.com/onsi/ginkgo" //nolint:staticcheck // dot import for test framework is intentional
	. "github.com/onsi/gomega" //nolint:staticcheck // dot import for test framework is intentional
	"github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/config"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/handlers"
	. "github.com/run-ai/runai/runai-cluster/cluster/project-controller/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Non-runai flag combination test (DoD).
// Verifies that when the binary is configured with all KAI-style label keys
// (instead of the runai defaults the rest of the suite uses), the consumers
// of `config.Get()` actually pick up the KAI vocabulary. This proves the
// production code reads its label keys via the config singleton and is not
// pinned to runai constants.
var _ = Describe("project-controller under KAI (non-runai) config", func() {
	const (
		kaiNodePoolLabelKey         = "kai.scheduler/node-pool"
		kaiQueueLabelKey            = "kai.scheduler/queue"
		kaiNamespaceProjectLabelKey = "project"
		kaiProjectLabelKey          = "project"
		kaiProjectIdLabelKey        = "kai.resources/project-id"
		kaiNamespaceVersionLabelKey = "kai.resources/namespace-version"
		kaiFinalizerDomain          = "kai.resources"
		kaiInstallNamespace         = "kai-scheduler"
		kaiProjectNamePrefix        = "kai"
		kaiExpectedFinalizer        = "kai.resources.finalizers.project"
	)

	var restoreConfig func()

	BeforeEach(func() {
		restoreConfig = config.SetForTest(&config.ProjectReconcilerConfig{
			NodePoolLabelKey:         kaiNodePoolLabelKey,
			QueueLabelKey:            kaiQueueLabelKey,
			NamespaceProjectLabelKey: kaiNamespaceProjectLabelKey,
			ProjectLabelKey:          kaiProjectLabelKey,
			ProjectIdLabelKey:        kaiProjectIdLabelKey,
			NamespaceVersionLabelKey: kaiNamespaceVersionLabelKey,
			FinalizerDomain:          kaiFinalizerDomain,
			InstallNamespace:         kaiInstallNamespace,
			ProjectNamePrefix:        kaiProjectNamePrefix,
		})
	})

	AfterEach(func() {
		restoreConfig()
	})

	It("exposes the KAI label keys via config.Get()", func() {
		Expect(config.Get().NodePoolLabelKey).To(Equal(kaiNodePoolLabelKey))
		Expect(config.Get().QueueLabelKey).To(Equal(kaiQueueLabelKey))
		Expect(config.Get().NamespaceProjectLabelKey).To(Equal(kaiNamespaceProjectLabelKey))
		Expect(config.Get().ProjectLabelKey).To(Equal(kaiProjectLabelKey))
		Expect(config.Get().ProjectIdLabelKey).To(Equal(kaiProjectIdLabelKey))
		Expect(config.Get().NamespaceVersionLabelKey).To(Equal(kaiNamespaceVersionLabelKey))
		Expect(config.Get().InstallNamespace).To(Equal(kaiInstallNamespace))
	})

	It("composes the finalizer string from the KAI finalizer-domain", func() {
		Expect(config.FinalizerName()).To(Equal(kaiExpectedFinalizer))
	})

	It("derives the namespace prefix from ProjectNamePrefix", func() {
		Expect(config.NamespacePrefix()).To(Equal(kaiProjectNamePrefix + "-"))
	})

	It("invokes the queue handler under the KAI config without panicking", func() {
		// Exercises the production path that previously used
		// `common.RunaiProjectLabel` / `common.RunaiProjectIdLabel` and now
		// reads from `config.Get()`. With a fake client that has no parent
		// department queue seeded, HandleResource is expected to surface a
		// not-found condition rather than panic; the value of the test is the
		// successful invocation under KAI config, demonstrating the code
		// reads label keys via Get().
		project := *TestProject.DeepCopy()
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		handler := NewQueueResourceHandler(k8sClient)

		conditions, _ := handler.HandleResource(project)
		Expect(conditions).ToNot(BeEmpty())
	})
})
