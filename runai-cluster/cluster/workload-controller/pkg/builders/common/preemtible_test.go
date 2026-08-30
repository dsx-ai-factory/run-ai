package common

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
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
		// Start with an empty fake client for each test
		c = fake.NewClientBuilder().WithScheme(scheme).Build()
	})

	addPriorityClass := func(pc *v1.PriorityClass) {
		c = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(pc).Build()
	}

	It("returns true if no priority class, relevant kind, and specPreemptible is true", func() {
		isPreemptible, err := IsPreemptible(ctx, nil, "", true, constants.KindInteractiveWorkload)
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeTrue())
	})

	It("returns false if no priority class, relevant kind, and specPreemptible is false", func() {
		isPreemptible, err := IsPreemptible(ctx, nil, "", false, constants.KindInteractiveWorkload)
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeFalse())
	})

	It("returns false if no priority class and irrelevant kind", func() {
		isPreemptible, err := IsPreemptible(ctx, nil, "", true, constants.KindTrainingWorkload)
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeFalse())
	})

	It("delegates to priority.IsPreemptible if priority class is set (preemptible)", func() {
		priorityClass := &v1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{Name: "test-preemptible"},
			Value:      50, // less than threshold
		}
		addPriorityClass(priorityClass)
		isPreemptible, err := IsPreemptible(ctx, c, "test-preemptible", false, constants.KindTrainingWorkload)
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeTrue())
	})

	It("delegates to priority.IsPreemptible if priority class is set (not preemptible)", func() {
		priorityClass := &v1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{Name: "test-nonpreemptible"},
			Value:      150, // greater than threshold
		}
		addPriorityClass(priorityClass)
		isPreemptible, err := IsPreemptible(ctx, c, "test-nonpreemptible", false, constants.KindTrainingWorkload)
		Expect(err).ToNot(HaveOccurred())
		Expect(isPreemptible).To(BeFalse())
	})

	It("returns error if priority class lookup fails", func() {
		// Use a fake client with no objects, so lookup fails
		_, err := IsPreemptible(ctx, c, "nonexistent", false, constants.KindTrainingWorkload)
		Expect(err).To(HaveOccurred())
	})
})
