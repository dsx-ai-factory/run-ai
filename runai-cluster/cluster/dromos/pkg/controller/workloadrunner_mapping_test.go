package controller

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

func newJobSet() *jobsetv1alpha2.JobSet {
	return &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "jobset.x-k8s.io/v1alpha2",
			Kind:       "JobSet",
		},
	}
}

func newWorkloadRunner(name, namespace string, template runtime.RawExtension) *dromosv1alpha1.WorkloadRunner {
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: template,
		},
	}
}

var _ = Describe("mapKartaToWorkloadRunnerEvent", func() {
	var (
		ctx        context.Context
		controller *WorkloadRunnerController
	)

	BeforeEach(func() {
		ctx = context.Background()
		s := scheme.Scheme
		Expect(dromosv1alpha1.AddToScheme(s)).To(Succeed())
		Expect(kartav1alpha1.AddToScheme(s)).To(Succeed())
	})

	Context("when mapping Karta to WorkloadRunner", func() {
		It("returns reconcile request for matching WorkloadRunner", func() {
			jobset := newJobSet()
			templateBytes, err := json.Marshal(jobset)
			Expect(err).NotTo(HaveOccurred())
			runner := newWorkloadRunner("test-runner", "default", runtime.RawExtension{Raw: templateBytes})
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithObjects(runner).
				Build()

			registry := framework.NewRegistry()
			fw := framework.NewFramework(registry, nil)
			fakeRecorder := record.NewFakeRecorder(100)
			controller = NewWorkloadRunnerController(fakeClient, fw, fakeRecorder)
			karta := test.NewValidJobSetKarta()

			requests := controller.mapKartaToWorkloadRunnerEvent(ctx, karta)

			Expect(requests).To(HaveLen(1))
			Expect(requests[0]).To(Equal(reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-runner",
					Namespace: "default",
				},
			}))
		})
	})
})
