package cmd_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

func TestSearchBaseCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SearchBaseCommand Suite")
}

var _ = Describe("SearchBaseCommand", func() {
	var (
		searchBaseCmd *cmd.SearchBaseCommand
		instance      *config.Instance
		cmdHelper     helper.Cmd
		mockTexts     *SearchableCommandTexts
	)

	BeforeEach(func() {
		instance = &config.Instance{}
		cmdHelper = helper.Cmd{}
		mockTexts = &SearchableCommandTexts{}
		searchBaseCmd = cmd.NewSearchBaseCommand(
			api.MPI,
			cmd.NewEmptyCobraCommand,
			"test",
			distributed.Delete,
			distributed.GetByNameAndProj,
			api.Distributed,
			instance,
			mockTexts,
			cmdHelper,
		)
	})

	Describe("NewSearchBaseCommand", func() {
		It("should create a new SearchBaseCommand with correct properties", func() {
			Expect(searchBaseCmd.Framework).To(Equal(api.MPI))
			Expect(searchBaseCmd.GetTxnFn).ToNot(BeNil())
			Expect(searchBaseCmd.Instance).To(Equal(instance))
			Expect(searchBaseCmd.OperationType).To(Equal("test"))
			Expect(searchBaseCmd.OperationFunc).To(BeAssignableToTypeOf(distributed.Delete))
			Expect(searchBaseCmd.GetWorkloadFn).To(BeAssignableToTypeOf(distributed.GetByNameAndProj))
			Expect(searchBaseCmd.WorkloadType).To(Equal(api.Distributed))
			Expect(searchBaseCmd.Texts).To(Equal(mockTexts))
		})
	})

	Describe("GetCommand", func() {
		It("should return a cobra.Command with correct properties", func() {
			command := searchBaseCmd.GetCommand()
			Expect(command).To(BeAssignableToTypeOf(&cobra.Command{}))
			Expect(command.Use).To(Equal("test [WORKLOAD_NAME]"))
			Expect(command.Short).To(Equal(mockTexts.GetShort()))
			Expect(command.PreRunE).ToNot(BeNil())
			Expect(command.RunE).ToNot(BeNil())
		})
	})

	Describe("RegisterFlags", func() {
		It("should register flags without error", func() {
			err := searchBaseCmd.RegisterFlags()
			Expect(err).ToNot(HaveOccurred())
		})
	})

})

// Mock implementation of SearchableCommandTexts for testing
type SearchableCommandTexts struct{}

func (m *SearchableCommandTexts) GetShort() string            { return "Mock short description" }
func (m *SearchableCommandTexts) GetLong() string             { return "Mock long description" }
func (m *SearchableCommandTexts) GetExample() string          { return "Mock example" }
func (m *SearchableCommandTexts) GetAliases() []string        { return []string{"mock-alias"} }
func (m *SearchableCommandTexts) GetBaseErrorMsg() string     { return "Mock base error message" }
func (m *SearchableCommandTexts) GetSuccessMsg(string) string { return "Mock success message" }
func (m *SearchableCommandTexts) GetNoResultsErrorMsg() string {
	return "Mock no results error message"
}
func (m *SearchableCommandTexts) GetSearchableErrorMsg() string {
	return "Mock searchable error message"
}
