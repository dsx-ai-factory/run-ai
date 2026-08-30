package kartaparser

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/utils/ptr"
)

var _ = Describe("TestRidHandler", Ordered, func() {
	Describe("Other funcs tests", func() {
		It("sumReplicas doesn't fail on nil", func() {
			sumReplicas(nil, nil)
			// if we got here, it means it didn't panic
			sumReplicas(nil, ptr.To(int32(5)))

			source := ptr.To(int32(5))
			sumReplicas(&source, nil)
			Expect(*source).To(BeEquivalentTo(int32(5)), "Expected source to remain unchanged when adding nil")
			sumReplicas(&source, ptr.To(int32(2)))
			Expect(*source).To(BeEquivalentTo(int32(7)), "Expected source to be 5 + 2 = 7")
		})
	})
})
