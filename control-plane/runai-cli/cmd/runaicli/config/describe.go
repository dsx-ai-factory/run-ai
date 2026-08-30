package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const configDescribeShortMsg = "Returns information about the configuration."

// describe Returns information about the configuration.
var describe = &cobra.Command{
	Use:     "describe",
	Short:   configDescribeShortMsg,
	Args:    cobra.NoArgs,
	Aliases: []string{"show"},

	RunE: func(cmd *cobra.Command, args []string) error {
		if ofJson {
			return generateJson(cmd)
		}

		if ofYaml {
			return generateYaml(cmd)
		}

		// should not reach this point
		return errors.New("no generate type was set")
	},
}

func generateJson(cmd *cobra.Command) error {
	configMap := viper.AllSettings()
	byteArr, err := json.MarshalIndent(configMap, "", "\t")
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, fmt.Sprintf(consts.ConfigGenerateFailed), err)
	}
	cmd.Printf("%s\n", string(byteArr))
	return nil
}

func generateYaml(cmd *cobra.Command) error {
	configMap := viper.AllSettings()
	byteArr, err := yaml.Marshal(configMap)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, fmt.Sprintf(consts.ConfigGenerateFailed), err)
	}
	cmd.Printf("%s\n", string(byteArr))
	return nil
}

func init() {
	Config.AddCommand(describe)

	describe.Flags().BoolVar(&ofJson, list.JsonStr, false, "Output structure JSON")
	describe.Flags().BoolVar(&ofYaml, list.YamlStr, false, "Output structure YAML")
	describe.MarkFlagsOneRequired(list.JsonStr, list.YamlStr)
	describe.MarkFlagsMutuallyExclusive(list.YamlStr, list.JsonStr)
}
