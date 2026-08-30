package attributes

import (
	"errors"

	"github.com/spf13/cobra"
)

type EnableFlag string

const (
	EnableTrue  EnableFlag = "enabled"
	EnableFalse EnableFlag = "disabled"

	ErrNoMatchEnableFlagMsg = "enable flag can be one of this values: enabled|disabled"
)

func (ef *EnableFlag) String() string {
	return string(*ef)
}

// Type is only used in help text
func (ef *EnableFlag) Type() string {
	return "enable"
}

// Set must have pointer receiver so it doesn't change the value of a copy
func (ef *EnableFlag) Set(v string) error {
	switch v {
	case string(EnableTrue), string(EnableFalse):
		*ef = EnableFlag(v)
		return nil
	default:
		return errors.New(ErrNoMatchEnableFlagMsg)
	}
}

func (ef *EnableFlag) IsEnabled() bool {
	return ef.String() == string(EnableTrue)
}

// EnableFlagCompletion auto completion support for log level
func EnableFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{
		"enabled\tenable interactive mode",
		"disabled\tdisable interactive mode",
	}, cobra.ShellCompDirectiveNoFileComp
}
