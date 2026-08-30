package externalworkloads

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWorkloadRunnerLabel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WorkloadRunner Label Suite")
}

const (
	testJobName   = "my-job"
	testNamespace = "default"
	testWlrName   = "my-wlr"
)

func newObjectWithOwnerRefs(ownerRefs ...metav1.OwnerReference) *metav1.PartialObjectMetadata {
	return &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name:            testJobName,
			Namespace:       testNamespace,
			OwnerReferences: ownerRefs,
		},
	}
}

func newExternalWorkload() *v1alpha1.ExternalWorkload {
	return &v1alpha1.ExternalWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{},
		},
	}
}

func wlrOwnerRef(name string) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: v1alpha1.SchemeGroupVersion.String(),
		Kind:       dromosv1alpha1.WorkloadRunnerKind,
		Name:       name,
	}
}

var _ = Describe("setWorkloadRunnerLabel", func() {
	It("does not set label when owner is not a WorkloadRunner", func() {
		object := newObjectWithOwnerRefs(metav1.OwnerReference{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       "my-cronjob",
		})
		ew := newExternalWorkload()

		setWorkloadRunnerLabel(object, ew)

		Expect(ew.Labels).NotTo(HaveKey(commonconstants.WorkloadRunnerNameLabel))
	})

	It("does not set label when there are no owner refs", func() {
		object := newObjectWithOwnerRefs()
		ew := newExternalWorkload()

		setWorkloadRunnerLabel(object, ew)

		Expect(ew.Labels).NotTo(HaveKey(commonconstants.WorkloadRunnerNameLabel))
	})

	It("sets label when owner is a WorkloadRunner", func() {
		object := newObjectWithOwnerRefs(
			metav1.OwnerReference{APIVersion: "batch/v1", Kind: "CronJob", Name: "my-cronjob"},
			wlrOwnerRef(testWlrName),
		)
		ew := newExternalWorkload()

		setWorkloadRunnerLabel(object, ew)

		Expect(ew.Labels[commonconstants.WorkloadRunnerNameLabel]).To(Equal(testWlrName))
	})
})

var _ = Describe("externalWorkloadTypeFromObjectLabels", func() {
	It("uses the native workload type for v1-converted workloads", func() {
		labels := map[string]string{
			commonconstants.ConvertedFromNativeLabel: "true",
			clustercommonconstants.LabelWorkloadKind: "Training",
		}

		Expect(externalWorkloadTypeFromObjectLabels(labels, "Job")).To(Equal("Training"))
	})

	It("keeps the fallback type for non-converted workloads", func() {
		labels := map[string]string{
			clustercommonconstants.LabelWorkloadKind: "Training",
		}

		Expect(externalWorkloadTypeFromObjectLabels(labels, "Job")).To(Equal("Job"))
	})

	It("keeps the fallback type when the native workload type is missing", func() {
		labels := map[string]string{
			commonconstants.ConvertedFromNativeLabel: "true",
		}

		Expect(externalWorkloadTypeFromObjectLabels(labels, "Job")).To(Equal("Job"))
	})
})

var _ = Describe("fillLabelsAndAnnotations", func() {
	It("copies the converted-from-native label onto the external workload", func() {
		labels := map[string]string{
			commonconstants.ConvertedFromNativeLabel: "true",
			clustercommonconstants.LabelWorkloadKind: "Training",
		}
		ew := newExternalWorkload()

		fillLabelsAndAnnotations(labels, nil, ew)

		Expect(ew.Labels).To(HaveKeyWithValue(commonconstants.ConvertedFromNativeLabel, "true"))
	})

	It("copies the logical workload name onto the external workload", func() {
		labels := map[string]string{
			clustercommonconstants.LabelWorkloadName: "my-workload",
			commonconstants.ConvertedFromNativeLabel: "true",
		}
		ew := newExternalWorkload()

		fillLabelsAndAnnotations(labels, nil, ew)

		Expect(ew.Labels).To(HaveKeyWithValue(clustercommonconstants.LabelWorkloadName, "my-workload"))
	})

	It("copies the converted-from-native label when the external workload has no labels yet", func() {
		labels := map[string]string{commonconstants.ConvertedFromNativeLabel: "true"}
		ew := &v1alpha1.ExternalWorkload{}

		fillLabelsAndAnnotations(labels, nil, ew)

		Expect(ew.Labels).To(HaveKeyWithValue(commonconstants.ConvertedFromNativeLabel, "true"))
	})

	It("does not set the converted-from-native label for natively-submitted workloads", func() {
		labels := map[string]string{clustercommonconstants.LabelWorkloadKind: "Training"}
		ew := newExternalWorkload()

		fillLabelsAndAnnotations(labels, nil, ew)

		Expect(ew.Labels).NotTo(HaveKey(commonconstants.ConvertedFromNativeLabel))
	})

	It("does not copy labels outside the allowlist", func() {
		labels := map[string]string{
			commonconstants.ConvertedFromNativeLabel: "true",
			"some/unrelated-label":                   "value",
		}
		ew := newExternalWorkload()

		fillLabelsAndAnnotations(labels, nil, ew)

		Expect(ew.Labels).NotTo(HaveKey("some/unrelated-label"))
	})
})

var _ = Describe("ExternalWorkloadHandler.updateSpecIfNeeded", func() {
	// An annotation-only change (the ewi-status handshake) must persist even when spec/labels are unchanged.
	It("persists an annotation-only change", func() {
		ctx := context.Background()
		scheme := runtime.NewScheme()
		Expect(v1alpha1.AddToScheme(scheme)).To(Succeed())

		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Name: "ew-1", Namespace: "team-a"},
			Spec:       v1alpha1.ExternalWorkloadSpec{OriginalWorkloadName: "my-job", Group: "batch", Version: "v1", Type: "Training"},
		}
		c := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(ew).Build()

		// Read back to get the stored resourceVersion, then change only the annotation.
		stored := &v1alpha1.ExternalWorkload{}
		Expect(c.Get(ctx, client.ObjectKeyFromObject(ew), stored)).To(Succeed())
		updated := stored.DeepCopy()
		updated.Annotations = map[string]string{commonconstants.EwiStatusAnnotation: "true"}

		handler := &ExternalWorkloadHandler{cachedClient: c}
		Expect(handler.updateSpecIfNeeded(ctx, updated, stored)).To(Succeed())

		after := &v1alpha1.ExternalWorkload{}
		Expect(c.Get(ctx, client.ObjectKeyFromObject(ew), after)).To(Succeed())
		Expect(after.Annotations).To(HaveKeyWithValue(commonconstants.EwiStatusAnnotation, "true"))
	})
})
