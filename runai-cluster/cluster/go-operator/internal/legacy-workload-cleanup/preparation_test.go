package legacycleanup

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	ktesting "k8s.io/client-go/testing"
)

var _ = Describe("Idempotent preparation", func() {
	It("patches only the allowlisted legacy workload finalizer", func() {
		descriptor := legacyWorkloadResources[0]
		workload := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		workload.SetResourceVersion("7")
		workload.SetFinalizers([]string{"foreign.example/finalizer", legacyWorkloadFinalizer})
		client := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
			map[schema.GroupVersionResource]string{descriptor.GVR: "TrainingWorkloadList"}, workload)

		Expect(removeObjectFinalizer(context.Background(), client, descriptor.GVR, workload, legacyWorkloadFinalizer)).To(Succeed())
		updated, err := client.Resource(descriptor.GVR).Namespace(workload.GetNamespace()).Get(
			context.Background(), workload.GetName(), metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.GetFinalizers()).To(Equal([]string{"foreign.example/finalizer"}))
		patchAction, ok := client.Actions()[0].(ktesting.PatchAction)
		Expect(ok).To(BeTrue())
		Expect(string(patchAction.GetPatch())).To(ContainSubstring("/metadata/uid"))
	})

	It("does not issue a patch when the finalizer is already absent", func() {
		descriptor := legacyWorkloadResources[0]
		workload := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		workload.SetFinalizers([]string{"foreign.example/finalizer"})
		client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), workload)

		Expect(removeObjectFinalizer(context.Background(), client, descriptor.GVR, workload, legacyWorkloadFinalizer)).To(Succeed())
		Expect(client.Actions()).To(HaveLen(0))
	})

	It("treats a workload that disappears before the patch as prepared", func() {
		descriptor := legacyWorkloadResources[0]
		workload := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		workload.SetResourceVersion("7")
		workload.SetFinalizers([]string{legacyWorkloadFinalizer})
		client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())

		Expect(removeObjectFinalizer(
			context.Background(), client, descriptor.GVR, workload, legacyWorkloadFinalizer,
		)).To(Succeed())
	})

	It("retries a finalizer patch after an optimistic-concurrency conflict", func() {
		descriptor := legacyWorkloadResources[0]
		workload := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		workload.SetResourceVersion("7")
		workload.SetFinalizers([]string{legacyWorkloadFinalizer})
		client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), workload)
		patches := 0
		client.PrependReactor("patch", descriptor.GVR.Resource, func(ktesting.Action) (bool, runtime.Object, error) {
			patches++
			if patches == 1 {
				return true, nil, apierrors.NewConflict(
					schema.GroupResource{Group: descriptor.GVR.Group, Resource: descriptor.GVR.Resource},
					workload.GetName(), errors.New("conflict"))
			}
			return false, nil, nil
		})

		Expect(removeObjectFinalizer(context.Background(), client, descriptor.GVR, workload, legacyWorkloadFinalizer)).To(Succeed())
		Expect(patches).To(Equal(2))
	})

	It("removes only the allowlisted finalizer from a terminating workload without waiting", func() {
		descriptor := legacyWorkloadResources[0]
		workload := newObject("run.ai/v2alpha1", "TrainingWorkload", "project-a", "training", "training-uid")
		workload.SetResourceVersion("7")
		now := metav1.Now()
		workload.SetDeletionTimestamp(&now)
		workload.SetFinalizers([]string{"foreign.example/finalizer", legacyWorkloadFinalizer})
		client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), workload)

		Expect(prepareWorkload(context.Background(), client, descriptor, workload)).To(Succeed())
		updated, err := client.Resource(descriptor.GVR).Namespace(workload.GetNamespace()).Get(
			context.Background(), workload.GetName(), metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(updated.GetFinalizers()).To(Equal([]string{"foreign.example/finalizer"}))
		Expect(client.Actions()).To(HaveLen(2))
		Expect(client.Actions()[0].GetVerb()).To(Equal("patch"))
	})
})
