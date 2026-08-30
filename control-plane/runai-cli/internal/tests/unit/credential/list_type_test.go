package credential

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

const (
	listTypeTestPath     = "credential_list_type"
	invalidTypeSubstring = "invalid credential type"
	emptyURL             = "empty url"
)

var listTypeEnvMap map[string]string

func listCmdArgs(typeVal string) []string {
	return []string{"credential", "list", "-A", "--type", typeVal}
}

var _ = Describe("credential list --type", Ordered, func() {
	BeforeAll(func() {
		homePath, err := cli.GetHomePath()
		Expect(err).ToNot(HaveOccurred())
		configPath := cli.GetConfigPath(homePath, listTypeTestPath)
		listTypeEnvMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)
	})

	Describe("validation", func() {
		It("rejects type that is not in the allowed list", func() {
			args := listCmdArgs("invalidType")
			_, err := cli.RunCommandStringSplit(context.Background(), args, listTypeEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Or(
				ContainSubstring(invalidTypeSubstring),
				ContainSubstring(emptyURL),
			))
		})

		It("rejects type value that is only whitespace", func() {
			args := []string{"credential", "list", "-A", "--type", "   "}
			_, err := cli.RunCommandStringSplit(context.Background(), args, listTypeEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Or(
				ContainSubstring(invalidTypeSubstring),
				ContainSubstring(emptyURL),
			))
		})
	})

	Describe("valid types", func() {
		It("accepts genericSecret", func() {
			args := listCmdArgs("genericSecret")
			_, err := cli.RunCommandStringSplit(context.Background(), args, listTypeEnvMap)
			Expect(err).To(Or(BeNil(), WithTransform(func(e error) string { return e.Error() }, ContainSubstring(emptyURL))))
		})

		It("accepts dockerRegistry", func() {
			args := listCmdArgs("dockerRegistry")
			_, err := cli.RunCommandStringSplit(context.Background(), args, listTypeEnvMap)
			Expect(err).To(Or(BeNil(), WithTransform(func(e error) string { return e.Error() }, ContainSubstring(emptyURL))))
		})

		It("accepts ngcApiKey", func() {
			args := listCmdArgs("ngcApiKey")
			_, err := cli.RunCommandStringSplit(context.Background(), args, listTypeEnvMap)
			Expect(err).To(Or(BeNil(), WithTransform(func(e error) string { return e.Error() }, ContainSubstring(emptyURL))))
		})

		It("accepts password", func() {
			args := listCmdArgs("password")
			_, err := cli.RunCommandStringSplit(context.Background(), args, listTypeEnvMap)
			Expect(err).To(Or(BeNil(), WithTransform(func(e error) string { return e.Error() }, ContainSubstring(emptyURL))))
		})

		It("accepts accessKey", func() {
			args := listCmdArgs("accessKey")
			_, err := cli.RunCommandStringSplit(context.Background(), args, listTypeEnvMap)
			Expect(err).To(Or(BeNil(), WithTransform(func(e error) string { return e.Error() }, ContainSubstring(emptyURL))))
		})
	})
})
