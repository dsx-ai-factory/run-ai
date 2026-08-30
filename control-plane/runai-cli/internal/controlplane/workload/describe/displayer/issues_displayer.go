package displayer

import (
	"github.com/charmbracelet/bubbles/table"
)

// Type definitions
type Issue struct {
	Type        string `json:"type"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Displayer
type IssuesDisplayer struct {
	issues []Issue
}

func NewIssuesSection(issues []Issue) *Section {
	return &Section{
		Displayer: &IssuesDisplayer{
			issues: issues,
		},
	}
}

func (id *IssuesDisplayer) Subject() string {
	return TitleIssues
}

func (id *IssuesDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: FieldIssueType, Width: IssueTypeColumnWidth},
		{Title: FieldResourceKind, Width: ResourceColumnWidth},
		{Title: FieldResourceName, Width: ResourceColumnWidth},
		{Title: FieldIssueDescription, Width: IssueDescriptionColumnWidth},
	}
}

func (id *IssuesDisplayer) Data() [][]string {
	var data [][]string
	for _, issue := range id.issues {
		row := []string{
			issue.Type,
			issue.Kind,
			issue.Name,
			issue.Description,
		}
		data = append(data, row)
	}
	return data
}
