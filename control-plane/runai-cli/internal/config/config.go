package config

import (
	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
)

var (
	instance = config.GetInstance()
)

func SetOutput(output *listCommon.Output) error {
	outputDefault := instance.GetConfigValue(config.SettingsOutputType)
	return output.SetDefault(outputDefault)
}

func SetProject(project *string) {
	if *project == "" {
		*project = instance.GetConfigValue(config.ClusterProjectName)
	}
}
