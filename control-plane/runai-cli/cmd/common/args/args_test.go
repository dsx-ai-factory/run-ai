package args_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/spf13/cobra"
)

func TestLogs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Args Suite")
}

var _ = Describe("Args Test", func() {
	var (
		cmd *cobra.Command
	)

	Context("GetNameFromArgs", func() {
		It("empty args", func() {
			value := args.GetNameFromArgs(&[]string{})
			Expect(value).To(Equal(""))
		})
		It("list of 1", func() {
			value := args.GetNameFromArgs(&[]string{"bla"})
			Expect(value).To(Equal("bla"))
		})
		It("list of 2", func() {
			value := args.GetNameFromArgs(&[]string{"bla", "bla2"})
			Expect(value).To(Equal("bla"))
		})
	})

	Context("ExtractNameFromArgs", func() {
		BeforeEach(func() {
			cmd = &cobra.Command{Use: "use"}
		})

		It("empty args", func() {
			var cmdArgs []string
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal(""))
		})

		It("command with another argument - second arg is the name - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "bla")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal("bla"))
			Expect(len(cmdArgs)).To(Equal(0))
		})

		It("command with flag - name is empty - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "-a bla")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal(""))
			Expect(len(cmdArgs)).To(Equal(0))
		})

		It("command with name and a flag - name is first arg - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "bla", "-a bla")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal("bla"))
			Expect(len(cmdArgs)).To(Equal(0))
		})

		It("command with flag at first and arg later - error missing delimiter", func() {
			cmdArgs := setOsArgsGetArgs("use", "-a bla", "bla1")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("[bla1]"))
			Expect(value).To(Equal(""))
		})

		It("command with flag at first and arg after delimiter - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "-a bla", "--", "bla1")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal(""))
			Expect(cmdArgs).To(Equal([]string{"bla1"}))
		})

		It("command with flag at first and arg after delimiter with command and arg - name is empty - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "-a bla1", "--", "use bla2")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal(""))
			Expect(cmdArgs).To(Equal([]string{"use bla2"}))
		})

		It("command with flag at first and arg after delimiter with command and arg - name is empty - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "bla1", "-a bla2", "--", "use bla3")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal("bla1"))
			Expect(cmdArgs).To(Equal([]string{"use bla3"}))
		})

		It("command with flag at first and arg after delimiter with command and arg - name is empty - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "bla1", "-a bla2", "anotherarg", "--", "use bla3")
			_, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("found args prior to args marker (--): [anotherarg]\nPlease check --help examples for how to user args and commands"))
		})

		It("command with flag at first and arg after delimiter with command and arg - name is empty - ok", func() {
			cmdArgs := setOsArgsGetArgs("use", "bla1", "-a bla2", "--", "use bla3", "-a yada", "--test this")
			value, err := args.ValidateArgsAndExtractName(cmd, &cmdArgs)
			Expect(err).ToNot(HaveOccurred())
			Expect(value).To(Equal("bla1"))
			Expect(cmdArgs).To(Equal([]string{"use bla3", "-a yada", "--test this"}))
		})
	})

	Context("BothAreSet", func() {
		It("args does not contain use", func() {
			Expect(args.BothAreSet(true, true).Error()).To(Equal("both prefix and name are set, only one can be used"))
			Expect(args.BothAreSet(false, true)).To(BeNil())
			Expect(args.BothAreSet(true, false)).To(BeNil())
			Expect(args.BothAreSet(false, false)).To(BeNil())
		})
	})
})

func setOsArgsGetArgs(argsString ...string) []string {
	if len(argsString) == 0 {
		return []string{}
	}

	var (
		foundDelimieter bool
		osArgsList      []string
		cmdArgsList     []string
	)

	for i := range argsString {
		if i == 0 {
			osArgsList = append(osArgsList, argsString[i])
			continue
		}

		if foundDelimieter {
			cmdArgsList = append(cmdArgsList, argsString[i])
			osArgsList = append(osArgsList, argsString[i])
			continue
		}

		if argsString[i] == "--" {
			foundDelimieter = true
		}

		if !strings.HasPrefix(argsString[i], "-") {
			cmdArgsList = append(cmdArgsList, argsString[i])
		}
		osArgsList = append(osArgsList, argsString[i])
	}

	os.Args = osArgsList
	if len(os.Args) == 0 {
		return []string{}
	}
	return cmdArgsList
}
