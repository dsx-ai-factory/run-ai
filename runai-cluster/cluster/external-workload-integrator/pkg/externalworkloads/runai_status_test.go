package externalworkloads

import (
	"context"
	"encoding/json"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/runaistatus"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type nativeInferenceStatusHandler struct{}

func (nativeInferenceStatusHandler) Parse(_ context.Context, _ client.Object,
	externalWorkload *v1alpha1.ExternalWorkload,
) error {
	externalWorkload.Spec.Type = inferenceWorkloadType
	externalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWPendingPhase
	externalWorkload.Status.ObservedStatus.Message = "native status"
	return nil
}

func (nativeInferenceStatusHandler) GetObject(_ context.Context, _, _ string) (client.Object, error) {
	return nil, nil
}

var _ = Describe("ExternalWorkloadHandler.applyRunaiStatus", func() {
	const namespace = "team-a"

	newScheme := func() *runtime.Scheme {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		Expect(dromosv1alpha1.AddToScheme(scheme)).To(Succeed())
		return scheme
	}

	newRunner := func(name string) *dromosv1alpha1.WorkloadRunner {
		return &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	}

	newEW := func(workloadType, runnerName string) *v1alpha1.ExternalWorkload {
		return &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Labels:    map[string]string{commonconstants.WorkloadRunnerNameLabel: runnerName},
			},
			Spec: v1alpha1.ExternalWorkloadSpec{Type: workloadType},
		}
	}

	It("publishes authoritative inference phase while preserving native status", func() {
		scheme := newScheme()
		runner := newRunner("inference-runner")
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name: "model-00001-pod", Namespace: namespace,
			Labels: map[string]string{"serving.knative.dev/revision": "model-00001"},
		}, Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
		}}
		client := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner, pod).Build()
		handler := &ExternalWorkloadHandler{cachedClient: client}
		service := &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "model", Namespace: namespace}}
		service.Status.LatestCreatedRevisionName = "model-00001"
		service.Status.LatestReadyRevisionName = "model-00001"
		ew := newEW(inferenceWorkloadType, runner.Name)
		ew.Status.ObservedStatus.K8sStatus = "Ready"

		handler.applyRunaiStatus(context.Background(), service, ew)

		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWRunningPhase))
		Expect(ew.Status.ObservedStatus.Message).To(BeEmpty())
		Expect(ew.Status.ObservedStatus.K8sStatus).To(Equal("Ready"))
		Expect(ew.Annotations).To(HaveKeyWithValue(commonconstants.EwiStatusAnnotation, "true"))
		Expect(additionalFields(ew)).To(SatisfyAll(
			HaveKeyWithValue("latestCreatedRevisionName", "model-00001"),
			HaveKeyWithValue("latestReadyRevisionName", "model-00001"),
		))
	})

	It("publishes authoritative distributed inference phase and replica counts", func() {
		scheme := newScheme()
		runner := newRunner("dinf-runner")
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{
			Name: "model-0", Namespace: namespace,
			Labels: map[string]string{
				"leaderworkerset.sigs.k8s.io/name":         "model",
				"leaderworkerset.sigs.k8s.io/group-index":  "0",
				"leaderworkerset.sigs.k8s.io/worker-index": "0",
			},
		}, Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
		}}
		client := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner, pod).Build()
		handler := &ExternalWorkloadHandler{cachedClient: client}
		lws := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "leaderworkerset.x-k8s.io/v1",
			"kind":       "LeaderWorkerSet",
			"metadata":   map[string]any{"name": "model", "namespace": namespace},
			"spec": map[string]any{
				"replicas": int64(1), "leaderWorkerTemplate": map[string]any{"size": int64(1)},
			},
		}}
		lws.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet",
		})
		ew := newEW(distributedInferenceWorkloadType, runner.Name)

		handler.applyRunaiStatus(context.Background(), lws, ew)

		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWRunningPhase))
		Expect(ew.Status.ObservedStatus.Message).To(Equal("Replicas: 1; Ready: 1"))
		Expect(ew.Status.ObservedStatus.K8sStatus).To(Equal("Ready"))
		Expect(ew.Annotations).To(HaveKeyWithValue(commonconstants.EwiStatusAnnotation, "true"))
		Expect(additionalFields(ew)).To(SatisfyAll(
			HaveKeyWithValue("statusReplicas", float64(1)),
			HaveKeyWithValue("readyReplicas", float64(1)),
		))
	})

	It("rejects an LWS replica count above the int32 maximum", func() {
		scheme := newScheme()
		runner := newRunner("dinf-runner")
		client := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(runner).Build()
		handler := &ExternalWorkloadHandler{cachedClient: client}
		lws := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "leaderworkerset.x-k8s.io/v1",
			"kind":       "LeaderWorkerSet",
			"metadata":   map[string]any{"name": "model", "namespace": namespace},
			"spec": map[string]any{
				"replicas":             int64(math.MaxInt32) + 1,
				"leaderWorkerTemplate": map[string]any{"size": int64(1)},
			},
		}}
		lws.SetGroupVersionKind(schema.GroupVersionKind{
			Group: "leaderworkerset.x-k8s.io", Version: "v1", Kind: "LeaderWorkerSet",
		})
		ew := newEW(distributedInferenceWorkloadType, runner.Name)

		err := handler.applyDistributedInferenceStatus(context.Background(), lws, runner, ew)

		Expect(err).To(MatchError("LWS replicas 2147483648 are outside the int32 range"))
	})

	It("omits zero replica counts and clears stale owned values", func() {
		staleFields := `{"statusReplicas":2,"readyReplicas":1,"keep":"value"}`
		ew := &v1alpha1.ExternalWorkload{
			Spec: v1alpha1.ExternalWorkloadSpec{AdditionalFields: &staleFields},
		}

		runaistatus.SetAdditionalFields(ew,
			[]string{statusReplicasField, readyReplicasField},
			map[string]any{statusReplicasField: int32(0), readyReplicasField: int32(0)})

		Expect(additionalFields(ew)).To(Equal(map[string]any{"keep": "value"}))
	})

	It("clears the authority annotation and keeps native status when the WLR is unavailable", func() {
		scheme := newScheme()
		client := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
		handler := &ExternalWorkloadHandler{cachedClient: client}
		service := &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "model", Namespace: namespace}}
		ew := newEW(inferenceWorkloadType, "missing-runner")
		ew.Annotations = map[string]string{commonconstants.EwiStatusAnnotation: "true"}
		ew.Status.ObservedStatus.Phase = v1alpha1.ExWPendingPhase
		ew.Status.ObservedStatus.Message = "native status"
		staleFields := `{"latestCreatedRevisionName":"old","latestReadyRevisionName":"old","keep":"value"}`
		ew.Spec.AdditionalFields = &staleFields

		handler.applyRunaiStatus(context.Background(), service, ew)

		Expect(ew.Annotations).NotTo(HaveKey(commonconstants.EwiStatusAnnotation))
		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWPendingPhase))
		Expect(ew.Status.ObservedStatus.Message).To(Equal("native status"))
		Expect(additionalFields(ew)).To(Equal(map[string]any{"keep": "value"}))
	})

	It("keeps newly parsed native status when the WLR is unavailable", func() {
		scheme := newScheme()
		cachedClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
		gvk := schema.GroupVersionKind{Group: "serving.knative.dev", Version: "v1", Kind: "Service"}
		handler := &ExternalWorkloadHandler{
			cachedClient: cachedClient,
			supportedGVKsHandlers: map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler{
				gvk: nativeInferenceStatusHandler{},
			},
		}
		service := &servingv1.Service{ObjectMeta: metav1.ObjectMeta{Name: "model", Namespace: namespace}}
		service.SetGroupVersionKind(gvk)
		ew := newEW(inferenceWorkloadType, "missing-runner")
		ew.Annotations = map[string]string{commonconstants.EwiStatusAnnotation: "true"}
		ew.Status.ObservedStatus.Phase = v1alpha1.ExWRunningPhase
		ew.Status.ObservedStatus.Message = "previous EWI status"
		staleFields := `{"latestCreatedRevisionName":"old","latestReadyRevisionName":"old","keep":"value"}`
		ew.Spec.AdditionalFields = &staleFields

		Expect(handler.parseObjectToExternalWorkload(context.Background(), service, ew)).To(Succeed())

		Expect(ew.Annotations).NotTo(HaveKey(commonconstants.EwiStatusAnnotation))
		Expect(ew.Status.ObservedStatus.Phase).To(Equal(v1alpha1.ExWPendingPhase))
		Expect(ew.Status.ObservedStatus.Message).To(Equal("native status"))
		Expect(additionalFields(ew)).To(Equal(map[string]any{"keep": "value"}))
	})
})

func additionalFields(ew *v1alpha1.ExternalWorkload) map[string]any {
	GinkgoHelper()
	result := make(map[string]any)
	Expect(ew.Spec.AdditionalFields).NotTo(BeNil())
	Expect(json.Unmarshal([]byte(*ew.Spec.AdditionalFields), &result)).To(Succeed())
	return result
}
