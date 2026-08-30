package cmd_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
)

var _ = Describe("BaseCommandTexts", func() {
	var (
		baseCmd *cmd.BaseCommandTexts
		tokens  []cmd.TokenPair
		aliases []string
	)

	BeforeEach(func() {
		tokens = []cmd.TokenPair{
			{Old: cmd.WorkloadTypeReplacer, New: "job"},
			{Old: cmd.FrameworkReplacer, New: "pytorch"},
		}
		aliases = []string{"j", "jobs"}

		baseCmd = cmd.NewBaseCommandTexts(
			"Short {{.WorkloadType}} description for {{.Framework}}",
			"Long {{.WorkloadType}} description for {{.Framework}}",
			"Example of {{.WorkloadType}} with {{.Framework}}",
			"Error in {{.WorkloadType}} with {{.Framework}}",
			"Successfully created %s",
			aliases,
			tokens,
		)
	})

	Context("token replacement", func() {
		It("should replace all tokens in short description", func() {
			result := baseCmd.GetShort()
			Expect(result).To(Equal("Short job description for pytorch"))
		})

		It("should replace all tokens in long description", func() {
			result := baseCmd.GetLong()
			Expect(result).To(Equal("Long job description for pytorch"))
		})

		It("should replace all tokens in example", func() {
			result := baseCmd.GetExample()
			Expect(result).To(Equal("Example of job with pytorch"))
		})

		It("should replace all tokens in error message", func() {
			result := baseCmd.GetBaseErrorMsg()
			Expect(result).To(Equal("Error in job with pytorch"))
		})

		It("should format success message with workload name", func() {
			result := baseCmd.GetSuccessMsg("test-workload")
			Expect(result).To(Equal("Successfully created %s"))
		})

		It("should return configured aliases", func() {
			result := baseCmd.GetAliases()
			Expect(result).To(Equal([]string{"j", "jobs"}))
		})
	})

	Context("empty template handling", func() {
		BeforeEach(func() {
			baseCmd = cmd.NewBaseCommandTexts(
				"",
				"",
				"",
				"",
				"",
				nil,
				tokens,
			)
		})

		It("should return empty string for empty templates", func() {
			Expect(baseCmd.GetShort()).To(Equal(""))
			Expect(baseCmd.GetLong()).To(Equal(""))
			Expect(baseCmd.GetExample()).To(Equal(""))
			Expect(baseCmd.GetBaseErrorMsg()).To(Equal(""))
			Expect(baseCmd.GetSuccessMsg("test-workload")).To(Equal(""))
		})

		It("should return nil for empty aliases", func() {
			Expect(baseCmd.GetAliases()).To(BeNil())
		})
	})

	Context("following messages", func() {
		It("should correctly format following workload message", func() {
			tokens = []cmd.TokenPair{
				{Old: cmd.WorkloadTypeReplacer, New: "job"},
				{Old: cmd.FrameworkReplacer, New: "pytorch"},
			}
			result := cmd.NewBaseCommandTexts(
				"",
				"",
				"",
				"",
				cmd.FollowingWorkload,
				nil,
				tokens,
			).GetSuccessMsg("")
			Expect(result).To(Equal("To track the workload's status, run 'runai job pytorch describe '\n"))
		})

		It("should correctly format following frameworkless message", func() {
			tokens = []cmd.TokenPair{
				{Old: cmd.WorkloadTypeReplacer, New: "job"},
			}
			result := cmd.NewBaseCommandTexts(
				"",
				"",
				"",
				"",
				cmd.FollowingFrameworkLess,
				nil,
				tokens,
			).GetSuccessMsg("")
			Expect(result).To(Equal("To track the workload's status, run 'runai job describe '\n"))
		})
	})
})

var _ = Describe("BaseSearchableCommandTexts", func() {
	var (
		searchableCmd *cmd.BaseSearchableCommandTexts
		tokens        []cmd.TokenPair
		aliases       []string
	)

	BeforeEach(func() {
		tokens = []cmd.TokenPair{
			{Old: cmd.WorkloadTypeReplacer, New: "job"},
			{Old: cmd.FrameworkReplacer, New: "pytorch"},
		}
		aliases = []string{"search", "find"}

		searchableCmd = cmd.NewBaseSearchableCommandTexts(
			"Short {{.WorkloadType}} description",
			"Long {{.WorkloadType}} description",
			"Example {{.WorkloadType}}",
			"Error in {{.WorkloadType}}",
			"Success with {{.WorkloadType}}",
			"No results for {{.WorkloadType}}",
			"Search {{.WorkloadType}} with {{.Framework}}",
			tokens,
			aliases,
		)
	})

	Context("searchable specific functionality", func() {
		It("should replace tokens in no results message", func() {
			result := searchableCmd.GetNoResultsErrorMsg()
			Expect(result).To(Equal("No results for job"))
		})

		It("should replace tokens in searchable error message", func() {
			result := searchableCmd.GetSearchableErrorMsg()
			Expect(result).To(Equal("Search job with pytorch"))
		})

		It("should return configured aliases", func() {
			result := searchableCmd.GetAliases()
			Expect(result).To(Equal([]string{"search", "find"}))
		})
	})

	Context("empty template handling", func() {
		BeforeEach(func() {
			searchableCmd = cmd.NewBaseSearchableCommandTexts(
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				tokens,
				nil,
			)
		})

		It("should handle empty searchable templates", func() {
			Expect(searchableCmd.GetNoResultsErrorMsg()).To(Equal(""))
			Expect(searchableCmd.GetSearchableErrorMsg()).To(Equal(""))
		})
	})
})

var _ = Describe("TokenPair Utils", func() {
	Context("ToList function", func() {
		It("should convert token pairs to replacement list", func() {
			pairs := []cmd.TokenPair{
				{Old: "{{.Test1}}", New: "value1"},
				{Old: "{{.Test2}}", New: "value2"},
			}

			result := cmd.ToList(pairs)
			expected := []string{"{{.Test1}}", "value1", "{{.Test2}}", "value2"}
			Expect(result).To(Equal(expected))
		})

		It("should handle empty token pairs", func() {
			var pairs []cmd.TokenPair
			result := cmd.ToList(pairs)
			Expect(result).To(BeEmpty())
		})
	})
})

var _ = Describe("Inference Pod Description Functions", func() {
	Describe("GetInferencePodDescription", func() {
		It("should return the correct description for Inference", func() {
			Expect(cmd.GetInferencePodDescription()).To(Equal("If no pod is specified, a pod from the latest ready revision will be used."))
		})
	})

	Describe("GetDistributedInferencePodDescription", func() {
		It("should return the correct description for DistributedInference", func() {
			Expect(cmd.GetDistributedInferencePodDescription()).To(Equal("If no pod is specified, a leader pod will be used."))
		})
	})

	Describe("GetNimPodDescription", func() {
		It("should return the correct description for NIM", func() {
			Expect(cmd.GetNimPodDescription()).To(Equal("If no pod is specified, a ready pod from the workload is used."))
		})
	})
})

var _ = Describe("Token Replacement Helper Functions", func() {
	Describe("GetFrameworkWorkloadReplacementTokens", func() {
		It("should create basic replacement tokens for Training workload with PyTorch", func() {
			tokens := cmd.GetFrameworkWorkloadReplacementTokens(api.Training, api.PyTorch)

			baseCmd := cmd.NewBaseCommandTexts(
				"Type: {{.Type}}, Command: {{.CommandType}}, Workload: {{.WorkloadType}}, Framework: {{.Framework}}, Example: {{.ExampleType}}, Following: {{.Following}}",
				"", "", "", "", nil, tokens,
			)

			result := baseCmd.GetShort()
			expectedText := "Type: pytorch training, Command: training pytorch, Workload: training, Framework: pytorch, Example: pytorch, Following: To track the workload's status, run 'runai training pytorch describe {{.WorkloadName}}'\n"
			Expect(result).To(Equal(expectedText))
		})

		It("should handle Workspace workload with standard framework", func() {
			tokens := cmd.GetFrameworkWorkloadReplacementTokens(api.Workspace, api.Runai)

			baseCmd := cmd.NewBaseCommandTexts(
				"Type: {{.Type}}, Command: {{.CommandType}}",
				"", "", "", "", nil, tokens,
			)

			result := baseCmd.GetShort()
			Expect(result).To(Equal("Type: standard workspace, Command: workspace standard"))
		})

		It("should handle all framework types with Training workload", func() {
			frameworks := []api.Framework{
				api.MPI,
				api.PyTorch,
				api.TF,
				api.XGBoost,
				api.Runai,
			}

			for _, framework := range frameworks {
				tokens := cmd.GetFrameworkWorkloadReplacementTokens(api.Training, framework)
				baseCmd := cmd.NewBaseCommandTexts("{{.Type}}", "", "", "", "", nil, tokens)
				result := baseCmd.GetShort()
				expectedFramework := framework.CLIType()
				Expect(result).To(Equal(fmt.Sprintf("%s training", expectedFramework)))
			}
		})

		It("should handle all framework types with Inference workload", func() {

			tokens := cmd.GetFrameworkWorkloadReplacementTokens(api.Inference, api.Knative)
			baseCmd := cmd.NewBaseCommandTexts("{{.Type}}", "", "", "", "", nil, tokens)
			result := baseCmd.GetShort()
			expectedFramework := api.Knative.CLIType()
			Expect(result).To(Equal(fmt.Sprintf("%s inference", expectedFramework)))

		})

		It("should handle extra token pairs with different workload types", func() {
			extraTokens := []cmd.TokenPair{
				{Old: "{{.Custom}}", New: "custom-value"},
			}

			testCases := []struct {
				workload  api.WorkloadType
				framework api.Framework
				expected  string
			}{
				{api.Training, api.PyTorch, "pytorch training with custom-value"},
				{api.Workspace, api.Runai, "standard workspace with custom-value"},
				{api.Inference, api.Knative, "knative inference with custom-value"},
			}

			for _, tc := range testCases {
				tokens := cmd.GetFrameworkWorkloadReplacementTokens(tc.workload, tc.framework, extraTokens...)
				baseCmd := cmd.NewBaseCommandTexts("{{.Type}} with {{.Custom}}", "", "", "", "", nil, tokens)
				result := baseCmd.GetShort()
				Expect(result).To(Equal(tc.expected))
			}
		})
	})

	Describe("GetWorkloadOnlyReplacementTokens", func() {
		It("should create basic replacement tokens for Training workload", func() {
			tokens := cmd.GetWorkloadOnlyReplacementTokens(api.Training)

			baseCmd := cmd.NewBaseCommandTexts(
				"Type: {{.Type}}, Command: {{.CommandType}}, Workload: {{.WorkloadType}}, Example: {{.ExampleType}}, Following: {{.Following}}",
				"", "", "", "", nil, tokens,
			)

			result := baseCmd.GetShort()
			expectedText := "Type: training, Command: training, Workload: training, Example: training, Following: To track the workload's status, run 'runai training describe {{.WorkloadName}}'\n"
			Expect(result).To(Equal(expectedText))
		})

		It("should handle Workspace workload", func() {
			tokens := cmd.GetWorkloadOnlyReplacementTokens(api.Workspace)

			baseCmd := cmd.NewBaseCommandTexts(
				"Type: {{.Type}}, Command: {{.CommandType}}",
				"", "", "", "", nil, tokens,
			)

			result := baseCmd.GetShort()
			Expect(result).To(Equal("Type: workspace, Command: workspace"))
		})

		It("should handle Inference workload", func() {
			tokens := cmd.GetWorkloadOnlyReplacementTokens(api.Inference)

			baseCmd := cmd.NewBaseCommandTexts(
				"Type: {{.Type}}, Command: {{.CommandType}}",
				"", "", "", "", nil, tokens,
			)

			result := baseCmd.GetShort()
			Expect(result).To(Equal("Type: inference, Command: inference"))
		})

		It("should handle Distributed workload type", func() {
			tokens := cmd.GetWorkloadOnlyReplacementTokens(api.Distributed)

			baseCmd := cmd.NewBaseCommandTexts(
				"Type: {{.Type}}, Command: {{.CommandType}}",
				"", "", "", "", nil, tokens,
			)

			result := baseCmd.GetShort()
			// Note: Distributed workload type uses training CLI type as per WorkloadType.CLIType()
			Expect(result).To(Equal("Type: training, Command: training"))
		})

		It("should handle extra token pairs with different workload types", func() {
			extraTokens := []cmd.TokenPair{
				{Old: "{{.Environment}}", New: "production"},
			}

			testCases := []struct {
				workload api.WorkloadType
				expected string
			}{
				{api.Training, "training in production"},
				{api.Workspace, "workspace in production"},
				{api.Inference, "inference in production"},
				{api.Distributed, "training in production"}, // Uses training CLI type
			}

			for _, tc := range testCases {
				tokens := cmd.GetWorkloadOnlyReplacementTokens(tc.workload, extraTokens...)
				baseCmd := cmd.NewBaseCommandTexts("{{.Type}} in {{.Environment}}", "", "", "", "", nil, tokens)
				result := baseCmd.GetShort()
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("should preserve unmatched tokens", func() {
			tokens := cmd.GetWorkloadOnlyReplacementTokens(api.Training)

			baseCmd := cmd.NewBaseCommandTexts(
				"{{.Type}} with {{.UnknownToken}}",
				"", "", "", "", nil, tokens,
			)

			result := baseCmd.GetShort()
			Expect(result).To(Equal("training with {{.UnknownToken}}"))
		})
	})
})
