package project

import (
	"bytes"
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Project Tests")
}

var _ = Describe("Project Test", func() {
	const (
		TestFilePath = "internal/tests/config_samples/"
	)

	var (
		err      error
		homePath string
		actual   *bytes.Buffer
	)

	BeforeEach(func() {
		homePath, err = os.Getwd()
		if err != nil {
			Fail("failed to get pwd")
		}
		homePath = strings.SplitAfter(homePath, "runai-cli/")[0]
		homePath += TestFilePath

		actual = new(bytes.Buffer)
		Cmd.SetOut(actual)
		Cmd.SetErr(actual)
	})

	Context("Project Command", func() {
		It("help", func() {
			Cmd.SetArgs([]string{"--help"})
			err := Cmd.Execute()
			Expect(err).To(BeNil())

			Expect(actual.String()).To(HavePrefix("project management\n\n"))
		})

		It("arg help", func() {
			Cmd.SetArgs([]string{"--interactive=disabled"})
			err := Cmd.Execute()
			Expect(err).To(BeNil())

			Expect(actual.String()).To(HavePrefix("project management\n\n"))
		})
	})
})
