package completion

import (
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type ResourceType string

const (
	Project       ResourceType = "project"
	Department    ResourceType = "department"
	Workload      ResourceType = "workload"
	Priorities    ResourceType = "priorities"
	ExistingPvc   ResourceType = "existing-pvc"
	Node          ResourceType = "node"
	AIApplication ResourceType = "application"
)

func (t ResourceType) String() string {
	return string(t)
}

type CompletionFn func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

func GetArgCompletionFn(
	resourceType ResourceType,
	initInstanceFn func(*cobra.Command) error,
	optionsListerFn func(*cobra.Command) ([]string, error),
) CompletionFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// prevent duplicate arg values
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		flagFn := GetFlagCompletionFn(resourceType, initInstanceFn, optionsListerFn)
		return flagFn(cmd, args, toComplete)
	}
}

func GetFlagCompletionFn(
	resourceType ResourceType,
	initInstanceFn func(*cobra.Command) error,
	optionsListerFn func(*cobra.Command) ([]string, error),
) CompletionFn {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		err := initInstanceFn(cmd)
		if err != nil {
			logError(err, resourceType)
			return nil, cobra.ShellCompDirectiveError
		}

		options, err := optionsListerFn(cmd)
		if err != nil {
			logError(err, resourceType)
			return nil, cobra.ShellCompDirectiveError
		}

		var shouldComplete = len(toComplete) > 0
		filtered := make([]string, 0)
		for _, option := range options {
			// if got a prefix to complete, make sure the option starts with it
			if shouldComplete && !strings.HasPrefix(option, toComplete) {
				continue
			}

			filtered = append(filtered, option)
		}

		if resourceType == ExistingPvc {
			return filtered, cobra.ShellCompDirectiveNoSpace
		}

		return filtered, cobra.ShellCompDirectiveNoFileComp
	}
}

func logError(err error, resourceType ResourceType) {
	logger := config.GetInstance().GetLogger()
	logger.Debug(consts.FailedToAutoComplete, zap.String("resource", resourceType.String()), zap.String("error", err.Error()))
}
