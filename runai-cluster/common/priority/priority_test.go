package priority

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("IsPreemptible", Ordered, func() {
	var (
		ctx    context.Context
		scheme *runtime.Scheme
		c      client.Client
	)

	BeforeAll(func() {
		scheme = runtime.NewScheme()
		_ = v1.AddToScheme(scheme)
	})

	BeforeEach(func() {
		ctx = context.Background()
		c = fake.NewClientBuilder().WithScheme(scheme).Build()
	})

	addPriorityClass := func(pc *v1.PriorityClass) {
		c = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(pc).Build()
	}

	It("returns false if priorityClassName is empty", func() {
		isPreemptible, err := IsPreemptible(ctx, c, "")
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeFalse())
	})

	It("returns true if PriorityClass value is below threshold", func() {
		priorityClass := &v1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{Name: "preemptible"},
			Value:      50, // below threshold
		}
		addPriorityClass(priorityClass)
		isPreemptible, err := IsPreemptible(ctx, c, "preemptible")
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeTrue())
	})

	It("returns false if PriorityClass value is above threshold", func() {
		priorityClass := &v1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{Name: "nonpreemptible"},
			Value:      150, // above threshold
		}
		addPriorityClass(priorityClass)
		isPreemptible, err := IsPreemptible(ctx, c, "nonpreemptible")
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeFalse())
	})

	It("returns error if PriorityClass is not found", func() {
		_, err := IsPreemptible(ctx, c, "doesnotexist")
		Expect(err).To(HaveOccurred())
	})
})
