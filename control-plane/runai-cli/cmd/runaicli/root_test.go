package runaicli

import (
	"bytes"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CLI Configuration Test", func() {
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
		RootCmd.SetOut(actual)
		RootCmd.SetErr(actual)
		RootCmd.SetArgs([]string{"--help"})
	})

	Context("Get Empty Configuration", func() {
		It("configuration should be nil", func() {
			err := RootCmd.Execute()
			Expect(err).To(BeNil())
		})
	})
})
