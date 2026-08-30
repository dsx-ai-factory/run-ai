package attributes

import (
	"errors"

	"github.com/spf13/cobra"
)

type ReportType string

const (
	None   ReportType = "none"
	Logger ReportType = "logger"
	Local  ReportType = "local"

	ErrNoMatchReportTypeMsg = "reporter type can be one of this values: none|logger|local"
)

func (rt *ReportType) String() string {
	return string(*rt)
}

func (rt *ReportType) Type() string {
	return "reporter"
}

func (rt *ReportType) Set(v string) error {
	switch v {
	case string(None), string(Logger), string(Local):
		*rt = ReportType(v)
		return nil
	default:
		return errors.New(ErrNoMatchReportTypeMsg)
	}
}

func (rt *ReportType) IsEnabled() bool {
	return rt.String() == string(EnableTrue)
}

// ReportTypeCompletion auto-completion support for reporter type
func ReportTypeCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"none\tno report",
		"logger\tonly console output",
		"local\tfile based output",
	}, cobra.ShellCompDirectiveNoFileComp
}
