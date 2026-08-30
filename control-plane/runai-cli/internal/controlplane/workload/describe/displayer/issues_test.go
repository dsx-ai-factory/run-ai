package displayer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

var _ = Describe("IssueBuilder", func() {
	It("should build an Issue correctly", func() {
		issue := displayer.NewIssueBuilder().
			Type("Compute").
			Kind("Pod").
			Name("issue1").
			Description("Test issue description").
			Build()

		Expect(issue.Type).To(Equal("Compute"))
		Expect(issue.Kind).To(Equal("Pod"))
		Expect(issue.Name).To(Equal("issue1"))
		Expect(issue.Description).To(Equal("Test issue description"))
	})
})

var _ = Describe("IssuesBuilder", func() {
	It("should build a collection of issues", func() {
		issue1 := displayer.NewIssueBuilder().Name("issue1").Build()
		issue2 := displayer.NewIssueBuilder().Name("issue2").Build()

		issues := displayer.NewIssuesBuilder().
			Add(issue1).
			Add(issue2).
			Build()

		Expect(issues).To(HaveLen(2))
		Expect(issues[0].Name).To(Equal("issue1"))
		Expect(issues[1].Name).To(Equal("issue2"))
	})

	It("should set entire collection with Set method", func() {
		issue1 := displayer.NewIssueBuilder().Name("issue1").Build()
		issue2 := displayer.NewIssueBuilder().Name("issue2").Build()

		issues := displayer.NewIssuesBuilder().
			Set([]displayer.Issue{issue1}).
			Add(issue2).
			Build()

		Expect(issues).To(HaveLen(2))
	})
})
