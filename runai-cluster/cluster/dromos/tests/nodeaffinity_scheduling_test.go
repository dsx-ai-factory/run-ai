package controller

import (
	"context"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/plugins/nodeaffinity"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	schedcorev1 "k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

// The merged child affinity is evaluated against fictive nodes with the
// kube-scheduler's own matcher.
var _ = Describe("NodeAffinity scheduling semantics", Serial, Ordered, func() {
	var (
		k8sManager     ctrl.Manager
		k8sClient      client.Client
		managerCtx     context.Context
		managerCancel  context.CancelFunc
		namespace      = "runai-naschedulingtest"
		controllerName = "nodeaffinity-scheduling-test-controller"
	)

	fleet := []*corev1.Node{
		fictiveNode("node-a", map[string]string{"zone": "a", "gpu-type": "a100"}),
		fictiveNode("node-b", map[string]string{"zone": "b", "gpu-type": "a100"}),
		fictiveNode("node-c", map[string]string{"zone": "a", "gpu-type": "h100"}),
		fictiveNode("node-bad", map[string]string{"zone": "b", "gpu-type": "h100"}),
	}

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		k8sManager, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(err).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		registry := framework.NewRegistry(nodeaffinity.NewPlugin(k8sClient))
		fw := framework.NewFramework(registry, nil)
		schedulingController := controller.
			NewWorkloadRunnerController(k8sClient, fw, k8sManager.GetEventRecorderFor(controllerName)).
			WithName(controllerName)
		Expect(schedulingController.SetupWithManager(managerCtx, k8sManager)).To(Succeed())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		startManager(k8sManager, managerCtx, managerCancel)

		err = k8sClient.Create(managerCtx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespace},
		})
		if !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}

		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)

		for _, node := range fleet {
			err := k8sClient.Create(managerCtx, node.DeepCopy())
			if !apierrors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		}
	})

	AfterAll(func() {
		for _, node := range fleet {
			_ = k8sClient.Delete(managerCtx, node.DeepCopy())
		}
		deleteAllTestResources(managerCtx, k8sClient)
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)
	})

	AfterEach(func() {
		deleteAllTestResources(managerCtx, k8sClient)
	})

	It("keeps every node but the excluded one schedulable", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-exclude", namespace).WithDefault().Build()
		createWorkloadRunner(managerCtx, k8sClient, "sched-exclude", namespace, jobset,
			newNodeAffinityInstructions(hostnameExclusion("node-bad")))

		child := waitForJobSet(managerCtx, k8sClient, "sched-exclude", namespace)

		Expect(matchingNodeNames(childNodeAffinity(child), fleet)).To(
			ConsistOf("node-a", "node-b", "node-c"))
	})

	It("ANDs the workload's own affinity with the exclusion", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-and", namespace).WithDefault().Build()
		setTemplateNodeAffinity(jobset, &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: "gpu-type", Operator: corev1.NodeSelectorOpIn, Values: []string{"a100"}},
					},
				}},
			},
		})
		createWorkloadRunner(managerCtx, k8sClient, "sched-and", namespace, jobset,
			newNodeAffinityInstructions(hostnameExclusion("node-a")))

		child := waitForJobSet(managerCtx, k8sClient, "sched-and", namespace)

		// a100 nodes are node-a and node-b; node-a is excluded.
		Expect(matchingNodeNames(childNodeAffinity(child), fleet)).To(ConsistOf("node-b"))
	})

	It("keeps a matches-nothing workload matching zero nodes", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-nothing", namespace).WithDefault().Build()
		setTemplateNodeAffinity(jobset, &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{}},
			},
		})
		createWorkloadRunner(managerCtx, k8sClient, "sched-nothing", namespace, jobset,
			newNodeAffinityInstructions(hostnameExclusion("node-bad")))

		child := waitForJobSet(managerCtx, k8sClient, "sched-nothing", namespace)

		Expect(matchingNodeNames(childNodeAffinity(child), fleet)).To(BeEmpty())
	})

	It("scores a duplicated preferred term once, not twice", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-prefer", namespace).WithDefault().Build()
		setTemplateNodeAffinity(jobset, &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
				Weight: 50,
				Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
					{Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"b", "a"}},
				}},
			}},
		})
		createWorkloadRunner(managerCtx, k8sClient, "sched-prefer", namespace, jobset,
			&dromosv1alpha1.Instructions{NodeAffinity: &dromosv1alpha1.NodeAffinityConfig{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
						Weight: 50,
						Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "zone", Operator: corev1.NodeSelectorOpIn, Values: []string{"a", "b"}},
						}},
					}},
				},
			}})

		child := waitForJobSet(managerCtx, k8sClient, "sched-prefer", namespace)

		Expect(preferredScore(childNodeAffinity(child), fleet[0])).To(Equal(int64(50)),
			"node-a is in zone a; a surviving duplicate would double the score to 100")
	})

	// Required filters before preferred scores: a preference for an excluded
	// node survives the merge but can never win.
	It("filters an excluded node even when the workload prefers it", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-conflict", namespace).WithDefault().Build()
		setTemplateNodeAffinity(jobset, &corev1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{{
				Weight: 100,
				Preference: corev1.NodeSelectorTerm{MatchExpressions: []corev1.NodeSelectorRequirement{
					{Key: hostnameLabel, Operator: corev1.NodeSelectorOpIn, Values: []string{"node-bad"}},
				}},
			}},
		})
		createWorkloadRunner(managerCtx, k8sClient, "sched-conflict", namespace, jobset,
			newNodeAffinityInstructions(hostnameExclusion("node-bad")))

		child := waitForJobSet(managerCtx, k8sClient, "sched-conflict", namespace)
		affinity := childNodeAffinity(child)

		Expect(matchingNodeNames(affinity, fleet)).To(ConsistOf("node-a", "node-b", "node-c"))
		Expect(preferredScore(affinity, fleet[3])).To(Equal(int64(100)),
			"the preference is kept, the exclusion just makes it unreachable")
	})

	It("matches no node when the workload requires the excluded node", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-required-conflict", namespace).WithDefault().Build()
		setTemplateNodeAffinity(jobset, &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{Key: hostnameLabel, Operator: corev1.NodeSelectorOpIn, Values: []string{"node-bad"}},
					},
				}},
			},
		})
		createWorkloadRunner(managerCtx, k8sClient, "sched-required-conflict", namespace, jobset,
			newNodeAffinityInstructions(hostnameExclusion("node-bad")))

		child := waitForJobSet(managerCtx, k8sClient, "sched-required-conflict", namespace)

		Expect(matchingNodeNames(childNodeAffinity(child), fleet)).To(BeEmpty())
	})

	It("shrinks the schedulable set when the exclusion union grows, idempotently", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-union", namespace).WithDefault().Build()
		runner := createWorkloadRunner(managerCtx, k8sClient, "sched-union", namespace, jobset,
			newNodeAffinityInstructions(hostnameExclusion("node-bad")))

		child := waitForJobSet(managerCtx, k8sClient, "sched-union", namespace)
		Expect(matchingNodeNames(childNodeAffinity(child), fleet)).To(
			ConsistOf("node-a", "node-b", "node-c"))

		union := newNodeAffinityInstructions(hostnameExclusion("node-bad", "node-b"))
		updateRunnerInstructions(managerCtx, k8sClient, runner, union)
		Eventually(func(g Gomega) {
			updated := waitForJobSet(managerCtx, k8sClient, "sched-union", namespace)
			g.Expect(matchingNodeNames(childNodeAffinity(updated), fleet)).To(
				ConsistOf("node-a", "node-c"))
		}, timeout, interval).Should(Succeed())

		updateRunnerInstructions(managerCtx, k8sClient, runner, union)
		Consistently(func(g Gomega) {
			updated := waitForJobSet(managerCtx, k8sClient, "sched-union", namespace)
			g.Expect(matchingNodeNames(childNodeAffinity(updated), fleet)).To(
				ConsistOf("node-a", "node-c"))
		}, "3s", interval).Should(Succeed())
	})

	It("restores the full schedulable set when the instruction is removed", func() {
		_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
		jobset := test.NewJobSetBuilder("sched-remove", namespace).WithDefault().Build()
		runner := createWorkloadRunner(managerCtx, k8sClient, "sched-remove", namespace, jobset,
			newNodeAffinityInstructions(hostnameExclusion("node-bad")))

		child := waitForJobSet(managerCtx, k8sClient, "sched-remove", namespace)
		Expect(matchingNodeNames(childNodeAffinity(child), fleet)).To(HaveLen(3))

		updateRunnerInstructions(managerCtx, k8sClient, runner, &dromosv1alpha1.Instructions{})
		Eventually(func(g Gomega) {
			updated := waitForJobSet(managerCtx, k8sClient, "sched-remove", namespace)
			g.Expect(matchingNodeNames(childNodeAffinity(updated), fleet)).To(
				ConsistOf("node-a", "node-b", "node-c", "node-bad"))
		}, timeout, interval).Should(Succeed())
	})
})

// fictiveNode labels the node with its hostname, which exclusions match on.
func fictiveNode(name string, labels map[string]string) *corev1.Node {
	labels[hostnameLabel] = name
	return &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
}

// matchingNodeNames returns the nodes the affinity's required selector matches;
// no required selector matches every node.
func matchingNodeNames(affinity *corev1.NodeAffinity, nodes []*corev1.Node) []string {
	names := make([]string, 0, len(nodes))
	if affinity == nil || affinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		for _, node := range nodes {
			names = append(names, node.Name)
		}
		return names
	}

	selector, err := schedcorev1.NewNodeSelector(affinity.RequiredDuringSchedulingIgnoredDuringExecution)
	Expect(err).NotTo(HaveOccurred())
	for _, node := range nodes {
		if selector.Match(node) {
			names = append(names, node.Name)
		}
	}
	return names
}

// preferredScore returns the raw weight sum the scheduler would add for node,
// before normalization.
func preferredScore(affinity *corev1.NodeAffinity, node *corev1.Node) int64 {
	Expect(affinity).NotTo(BeNil())
	terms, err := schedcorev1.NewPreferredSchedulingTerms(affinity.PreferredDuringSchedulingIgnoredDuringExecution)
	Expect(err).NotTo(HaveOccurred())
	return terms.Score(node)
}

// childNodeAffinity returns the node affinity every replicated job carries,
// asserting they all agree. A child with no affinity returns nil.
func childNodeAffinity(jobSet *jobsetv1alpha2.JobSet) *corev1.NodeAffinity {
	Expect(jobSet.Spec.ReplicatedJobs).NotTo(BeEmpty())

	var affinity *corev1.NodeAffinity
	for i, replicatedJob := range jobSet.Spec.ReplicatedJobs {
		podAffinity := replicatedJob.Template.Spec.Template.Spec.Affinity
		var current *corev1.NodeAffinity
		if podAffinity != nil {
			current = podAffinity.NodeAffinity
		}
		if i == 0 {
			affinity = current
			continue
		}
		Expect(current).To(Equal(affinity), "every replicated job must carry the same node affinity")
	}
	return affinity
}

func setTemplateNodeAffinity(jobSet *jobsetv1alpha2.JobSet, affinity *corev1.NodeAffinity) {
	for i := range jobSet.Spec.ReplicatedJobs {
		jobSet.Spec.ReplicatedJobs[i].Template.Spec.Template.Spec.Affinity = &corev1.Affinity{NodeAffinity: affinity}
	}
}

func updateRunnerInstructions(
	ctx context.Context,
	k8sClient client.Client,
	runner *dromosv1alpha1.WorkloadRunner,
	instructions *dromosv1alpha1.Instructions,
) {
	Eventually(func(g Gomega) {
		current := &dromosv1alpha1.WorkloadRunner{}
		g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(runner), current)).To(Succeed())
		current.Spec.Instructions = instructions
		g.Expect(k8sClient.Update(ctx, current)).To(Succeed())
	}, timeout, interval).Should(Succeed())
}
