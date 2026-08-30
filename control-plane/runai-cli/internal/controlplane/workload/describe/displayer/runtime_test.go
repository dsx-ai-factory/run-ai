package displayer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

var _ = Describe("RuntimeBuilder", func() {
	It("should build a Runtime correctly", func() {
		runtime := displayer.NewRuntimeBuilder().
			Preemptible(true).
			Images([]string{"image1", "image2"}).
			Build()

		Expect(*runtime.Preemptible).To(BeTrue())
		Expect(runtime.Images).To(Equal([]string{"image1", "image2"}))
	})

	It("should add images incrementally", func() {
		runtime := displayer.NewRuntimeBuilder().
			AddImage("image1").
			AddImage("image2").
			Build()

		Expect(runtime.Images).To(Equal([]string{"image1", "image2"}))
	})

	It("should set entire Runtime with Set method", func() {
		original := displayer.NewRuntimeBuilder().
			Preemptible(true).
			Images([]string{"original-image"}).
			Build()

		runtime := displayer.NewRuntimeBuilder().
			Set(original).
			Build()

		Expect(*runtime.Preemptible).To(BeTrue())
		Expect(runtime.Images).To(Equal([]string{"original-image"}))
	})
})
