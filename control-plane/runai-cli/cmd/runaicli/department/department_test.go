package department

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Department Tests")
}

var _ = Describe("Department Test", func() {
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

	Context("Department Command", func() {
		It("help", func() {
			Cmd.SetArgs([]string{"--help"})
			err := Cmd.Execute()
			Expect(err).To(BeNil())

			Expect(actual.String()).To(HavePrefix(fmt.Sprintf("%s\n\n", consts.DepartmentShortMsg)))
		})

		It("arg help", func() {
			Cmd.SetArgs([]string{"--interactive=disabled"})
			err := Cmd.Execute()
			Expect(err).To(BeNil())

			Expect(actual.String()).To(HavePrefix(fmt.Sprintf("%s\n\n", consts.DepartmentShortMsg)))
		})
	})
})
