package version

import (
	"os"
	"regexp"
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Version Tests")
}

const (
	TestFilePath = "version"

	VersionRegexLookupWide = `^runai-cli/([\d]+\.){2}[\d]+ sdk/([\d]+\.){2}[\d]+-local go/go[\d.].[\d.]* ` +
		`git/.* build/[\d.]*-[\d.]*-[\d.]*T[\d.]*:[\d.]*:[\d.]*Z platform/.*_.*`
	VersionRegexLookup = `^runai-cli/([\d]+\.){2}[\d]+ sdk/([\d]+\.){2}[\d]+`
)

var _ = Describe("Version Test", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)
	})

	Context("Version command validation", Ordered, func() {
		It("Version", func() {
			output, err := cli.RunCommand("version", envMap)
			Expect(err).To(BeNil())
			match, err := regexp.Match(VersionRegexLookup, []byte(output))
			Expect(err).To(BeNil())
			Expect(true).To(Equal(match))
		})

		It("Version wide", func() {
			output, err := cli.RunCommand("version --wide", envMap)
			Expect(err).To(BeNil())
			match, err := regexp.Match(VersionRegexLookupWide, []byte(output))
			Expect(err).To(BeNil())
			Expect(true).To(Equal(match))
		})

		It("Version invalid path", func() {
			output, err := cli.RunCommand("version",
				map[string]string{consts.EnvRunaiCliConfigFile: "empty.json",
					consts.EnvRunaiCliConfigPath: configPath})
			Expect(err).To(BeNil())
			match, err := regexp.Match(VersionRegexLookup, []byte(output))
			Expect(err).To(BeNil())
			Expect(true).To(Equal(match))
		})
	})
})
