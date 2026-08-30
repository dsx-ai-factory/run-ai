package datasource_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/datasource"
)

var _ = Describe("Datasource List Command", func() {
	Describe("Command Structure", func() {
		It("should have correct aliases", func() {
			Expect(datasource.List.Aliases).To(ContainElement("ls"))
		})

		It("should accept no arguments", func() {
			Expect(datasource.List.Args).NotTo(BeNil())
		})

		It("should have correct use string", func() {
			Expect(datasource.List.Use).To(Equal("list"))
		})
	})

	Describe("Flags", func() {
		It("should have project flag with shorthand", func() {
			flag := datasource.List.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have all flag with shorthand", func() {
			flag := datasource.List.Flags().Lookup("all")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("A"))
		})

		It("should have type flag", func() {
			flag := datasource.List.Flags().Lookup("type")
			Expect(flag).NotTo(BeNil())
		})

		It("should have type flag as optional (not required)", func() {
			flag := datasource.List.Flags().Lookup("type")
			Expect(flag).NotTo(BeNil())
			// Check that the flag does NOT have required annotation
			annotations := flag.Annotations
			Expect(annotations).NotTo(HaveKey("cobra_annotation_bash_completion_one_required_flag"))
		})

		It("should allow parsing type flag", func() {
			err := datasource.List.Flags().Parse([]string{"--type", "Pvc"})
			Expect(err).To(BeNil())
		})

		It("should allow running without type flag", func() {
			err := datasource.List.Flags().Parse([]string{"-A"})
			Expect(err).To(BeNil())
		})
	})
})
