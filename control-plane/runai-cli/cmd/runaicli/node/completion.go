package node

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/spf13/cobra"
)

func GetNodeNamesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetFlagCompletionFn(completion.Node, initNodeInstance, getNodeNames)
	return fn(cmd, args, toComplete)
}

func getNodeNames(_ *cobra.Command) ([]string, error) {
	nodesHandler, err := executeNodeList()
	if err != nil {
		return nil, err
	}
	nodes := nodesHandler.GetList()
	names := make([]string, 0)
	for _, brief := range nodes {
		names = append(names, brief.Name)
	}
	return names, nil
}
