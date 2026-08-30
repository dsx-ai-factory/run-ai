package pod

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaipgconstants "github.com/kai-scheduler/KAI-scheduler/pkg/podgrouper/podgrouper/plugins/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/pod-group-assigner/pkg/config"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("PodMutator project name label", func() {
	BeforeEach(func() {
		DeferCleanup(config.SetForTest(enforcementConfig()))
	})

	It("labels the pod with the project of its namespace", func() {
		mutator := newPodMutatorForTest(namespaceWithProject(projectNamespace, projectName))
		pod := newPod(nil, nil)

		mutator.mutateProjectNameLabel(context.Background(), pod, projectNamespace)

		Expect(pod.Labels).To(HaveKeyWithValue(kaipgconstants.ProjectLabelKey, projectName))
	})

	It("adds the label to a pod that already has other labels", func() {
		mutator := newPodMutatorForTest(namespaceWithProject(projectNamespace, projectName))
		pod := newPod(map[string]string{"app": "trainer"}, nil)

		mutator.mutateProjectNameLabel(context.Background(), pod, projectNamespace)

		Expect(pod.Labels).To(HaveKeyWithValue(kaipgconstants.ProjectLabelKey, projectName))
		Expect(pod.Labels).To(HaveKeyWithValue("app", "trainer"))
	})

	It("keeps an explicit project label set by the submitter", func() {
		mutator := newPodMutatorForTest(namespaceWithProject(projectNamespace, projectName))
		pod := newPod(map[string]string{kaipgconstants.ProjectLabelKey: "a-different-project"}, nil)

		mutator.mutateProjectNameLabel(context.Background(), pod, projectNamespace)

		Expect(pod.Labels).To(HaveKeyWithValue(kaipgconstants.ProjectLabelKey, "a-different-project"))
	})

	It("does nothing when the namespace has no project label", func() {
		mutator := newPodMutatorForTest(
			&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: projectNamespace}},
		)
		pod := newPod(nil, nil)

		mutator.mutateProjectNameLabel(context.Background(), pod, projectNamespace)

		Expect(pod.Labels).NotTo(HaveKey(kaipgconstants.ProjectLabelKey))
	})

	It("does nothing when the namespace cannot be read", func() {
		mutator := newPodMutatorForTest()
		pod := newPod(nil, nil)

		mutator.mutateProjectNameLabel(context.Background(), pod, projectNamespace)

		Expect(pod.Labels).NotTo(HaveKey(kaipgconstants.ProjectLabelKey))
	})
})
