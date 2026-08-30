package distributed_status

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	dinfw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

var _ = Describe("LwsStatusCalculator adapter", func() {
	var (
		ctx        context.Context
		k8sClient  client.Client
		calculator *LwsStatusCalculator
		workload   *dinfw.DistributedInferenceWorkload
		lws        *lwsv1.LeaderWorkerSet
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(lwsv1.AddToScheme(scheme)).To(Succeed())
		Expect(dinfw.AddToScheme(scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
		calculator = NewLwsStatusCalculator(k8sClient)
		workload = &dinfw.DistributedInferenceWorkload{ObjectMeta: metav1.ObjectMeta{
			Name: "model", Namespace: "team-a", UID: types.UID("workload-uid"),
		}}
		lws = &lwsv1.LeaderWorkerSet{Spec: lwsv1.LeaderWorkerSetSpec{
			Replicas:             ptr.To(int32(2)),
			LeaderWorkerTemplate: lwsv1.LeaderWorkerTemplate{Size: ptr.To(int32(1))},
		}}
	})

	It("uses the LWS shape and shared calculator defaults", func() {
		phase, replicas, readyReplicas, message, err := calculator.GetStatus(ctx, workload, lws)

		Expect(err).NotTo(HaveOccurred())
		Expect(phase).To(Equal(PhasePending))
		Expect(replicas).To(Equal(int32(2)))
		Expect(readyReplicas).To(BeZero())
		Expect(message).To(Equal("Replicas: 2; Pending: 2"))
	})

	It("passes matching workload pods to the shared calculator", func() {
		lws.Spec.Replicas = ptr.To(int32(1))
		lws.Spec.LeaderWorkerTemplate.Size = ptr.To(int32(2))
		Expect(k8sClient.Create(ctx, legacyGroupPod(workload, "leader", "0", "0", true))).To(Succeed())
		Expect(k8sClient.Create(ctx, legacyGroupPod(workload, "worker", "0", "1", false))).To(Succeed())

		phase, replicas, readyReplicas, message, err := calculator.GetStatus(ctx, workload, lws)

		Expect(err).NotTo(HaveOccurred())
		Expect(phase).To(Equal(PhaseReady))
		Expect(replicas).To(Equal(int32(1)))
		Expect(readyReplicas).To(Equal(int32(1)))
		Expect(message).To(Equal("Replicas: 1; Ready: 1"))
	})

	It("does not include pods owned by another workload", func() {
		other := workload.DeepCopy()
		other.UID = types.UID("other-uid")
		Expect(k8sClient.Create(ctx, legacyGroupPod(other, "other", "0", "0", true))).To(Succeed())

		phase, _, readyReplicas, _, err := calculator.GetStatus(ctx, workload, lws)

		Expect(err).NotTo(HaveOccurred())
		Expect(phase).To(Equal(PhasePending))
		Expect(readyReplicas).To(BeZero())
	})

	It("preserves the legacy error for a complete group without a leader", func() {
		lws.Spec.Replicas = ptr.To(int32(1))
		Expect(k8sClient.Create(ctx, legacyGroupPod(workload, "worker", "0", "1", true))).To(Succeed())

		_, _, _, _, err := calculator.GetStatus(ctx, workload, lws)

		Expect(err).To(MatchError("leader pod not found"))
	})
})

func legacyGroupPod(workload *dinfw.DistributedInferenceWorkload, name, group, worker string,
	ready bool,
) *corev1.Pod {
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: name, Namespace: workload.Namespace, Labels: map[string]string{
			lwsv1.SetNameLabelKey:            builders.GetUniqueLwsName(workload),
			commonconstants.TopOwnerUidLabel: string(workload.UID), //nolint:staticcheck // Legacy pods use this label.
			lwsv1.GroupIndexLabelKey:         group,
			lwsv1.WorkerIndexLabelKey:        worker,
		},
	}, Status: corev1.PodStatus{Phase: corev1.PodRunning}}
	if ready {
		pod.Status.Conditions = []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}
	}
	return pod
}
