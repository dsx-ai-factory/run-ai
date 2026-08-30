package department

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/spf13/cobra"
)

func GetDepartmentNamesForArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetArgCompletionFn(completion.Department, initDepartmentInstance, getDepartmentNames)
	return fn(cmd, args, toComplete)
}

func GetDepartmentNamesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetFlagCompletionFn(completion.Department, initDepartmentInstance, getDepartmentNames)
	return fn(cmd, args, toComplete)
}

func getDepartmentNames(_ *cobra.Command) ([]string, error) {
	flagOptions := ListOptions{
		limit: DefaultItems,
	}
	departmentBriefs, err := getList(flagOptions)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, brief := range departmentBriefs.GetList() {
		names = append(names, brief.Name)
	}
	return names, nil
}
