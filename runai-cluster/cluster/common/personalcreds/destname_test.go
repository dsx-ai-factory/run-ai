package personalcreds

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DestName", func() {
	It("is deterministic", func() {
		Expect(DestName("my-runner", "my-secret")).To(Equal("pc-my-runner-my-secret"))
		Expect(DestName("my-runner", "my-secret")).To(Equal(DestName("my-runner", "my-secret")))
	})

	It("differs per runner", func() {
		Expect(DestName("a", "s")).ToNot(Equal(DestName("b", "s")))
	})

	It("differs per source", func() {
		Expect(DestName("r", "x")).ToNot(Equal(DestName("r", "y")))
	})
})
