package finalizers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
)

var _ = Describe("GetRemoveFinalizerFactory", func() {
	It("returns InteractiveWorkloadRemoveFinalizerFactory for InteractiveWorkload", func() {
		factory, err := GetRemoveFinalizerFactory(constants.InteractiveWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&InteractiveWorkloadRemoveFinalizerFactory{}))
	})

	It("returns TrainingWorkloadRemoveFinalizerFactory for TrainingWorkload", func() {
		factory, err := GetRemoveFinalizerFactory(constants.TrainingWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&TrainingWorkloadRemoveFinalizerFactory{}))
	})

	It("returns DistributedWorkloadRemoveFinalizerFactory for DistributedWorkload", func() {
		factory, err := GetRemoveFinalizerFactory(constants.DistributedWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&DistributedWorkloadRemoveFinalizerFactory{}))
	})

	It("returns error for unknown workload type", func() {
		_, err := GetRemoveFinalizerFactory("unknown")
		Expect(err).ToNot(BeNil())
	})
})
