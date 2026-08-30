package displayer

// Builder
type IssueBuilder struct {
	issue Issue
}

func NewIssueBuilder() *IssueBuilder {
	return &IssueBuilder{issue: Issue{}}
}

func (b *IssueBuilder) Type(t string) *IssueBuilder {
	b.issue.Type = t
	return b
}

func (b *IssueBuilder) Kind(kind string) *IssueBuilder {
	b.issue.Kind = kind
	return b
}

func (b *IssueBuilder) Name(name string) *IssueBuilder {
	b.issue.Name = name
	return b
}

func (b *IssueBuilder) Description(description string) *IssueBuilder {
	b.issue.Description = description
	return b
}

func (b *IssueBuilder) Build() Issue {
	return b.issue
}
