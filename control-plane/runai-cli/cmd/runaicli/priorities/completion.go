package priorities

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/spf13/cobra"
)

func GetPrioritiesNamesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetFlagCompletionFn(completion.Priorities, initInstance, getPrioritiesNames)
	return fn(cmd, args, toComplete)
}

func getPrioritiesNames(_ *cobra.Command) ([]string, error) {
	output, err := getPrioritiesList()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, priority := range output.GetList() {
		names = append(names, priority.Name)
	}

	return names, nil
}
