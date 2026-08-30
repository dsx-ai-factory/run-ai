package cli

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"

	"github.com/spf13/cobra"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/pflag"
)

const (
	IgnoreFolder = "ignore_folder"
	ConfigJson   = "config.json"

	ClusterOneUuid   = "uuid-cluster-1"
	ClusterOneName   = "name-cluster-1"
	ClusterOneDomain = "http://localhost.runailabs.com/"
	ClusterTwoUuid   = "uuid-cluster-2"
	ClusterTwoName   = "name-cluster-2"

	ErrorConfigFileAlreadyExist = "runai-cli/ignore_folder/%s/config.json\" Already Exists\n"
)

func RunCommandWithContext(ctx context.Context, command string, env map[string]string) (string, error) {
	config.GetInstance().Reload()
	settings.GetInstance().Reload()

	cmdSplit := strings.Split(command, " ")
	os.Args = []string{}
	os.Args = append(os.Args, cmdSplit...)
	cmd := runaicli.RootCmd.Root()
	actual := new(bytes.Buffer)
	cmd.SetArgs(cmdSplit)
	cmd.SetOut(actual)
	cmd.SetErr(actual)

	// staring from root command resetting each flags for each subcommand.
	resetCmd := runaicli.RootCmd.Root()
	commandList := resetCmd.Commands()
	resetSubCommandFlags(resetCmd)
	resetCommandFlags(commandList, cmdSplit)

	closer := tempEnvironmentSetter(env)
	defer closer()

	_, err := cmd.ExecuteContextC(ctx)

	cmd.SetOut(nil)
	cmd.SetErr(nil)

	// TODO have a func that separate the stdout and errors
	return actual.String(), err
}

func RunCommand(command string, env map[string]string) (string, error) {
	return RunCommandWithContext(context.TODO(), command, env)
}

func RunCommandStringSplit(ctx context.Context, command []string, env map[string]string) (string, error) {
	os.Args = []string{}
	os.Args = command
	cmd := runaicli.RootCmd.Root()
	actual := new(bytes.Buffer)
	cmd.SetArgs(command)
	cmd.SetOut(actual)
	cmd.SetErr(actual)

	// staring from root command resetting each flags for each subcommand.
	resetCmd := runaicli.RootCmd.Root()
	commandList := resetCmd.Commands()
	resetSubCommandFlags(resetCmd)
	resetCommandFlags(commandList, command)

	closer := tempEnvironmentSetter(env)
	defer closer()

	_, err := cmd.ExecuteContextC(ctx)

	cmd.SetOut(nil)
	cmd.SetErr(nil)

	// TODO have a func that separate the stdout and errors
	return actual.String(), err
}

// resetCommandFlags known issue with CLI with unit tests
func resetCommandFlags(commandList []*cobra.Command, cmdSplit []string) {
	if commandList != nil {
		for _, cmdStr := range cmdSplit {
			for _, subCmd := range commandList {
				// Match on Name() or aliases rather than Use: commands with
				// positional args in their Use string (e.g. "submit
				// APPLICATION_NAME") are never matched by the full Use, and
				// invocations through an alias (e.g. "applications") never
				// match the canonical name - either way their flags would be
				// left unreset between tests.
				if cmdStr == subCmd.Name() || subCmd.HasAlias(cmdStr) {
					resetSubCommandFlags(subCmd)
					commandList = subCmd.Commands()
					break
				}
			}
		}
	}
}

func resetSubCommandFlags(subCmd *cobra.Command) {
	// Reset flag values so they don't leak between tests
	subCmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Reset the Changed status to prevent state pollution
		f.Changed = false

		switch v := f.Value.(type) {
		case pflag.SliceValue:
			_ = v.Replace([]string{})
		default:
			switch v.Type() {
			case "bool", "string", "int", "int32":
				_ = v.Set(f.DefValue)
			}
		}
	})
}

func tempEnvironmentSetter(envs map[string]string) (closer func()) {
	originalEnvs := map[string]string{}

	for name, value := range envs {
		if originalValue, ok := os.LookupEnv(name); ok {
			originalEnvs[name] = originalValue
		}
		_ = os.Setenv(name, value)
	}

	return func() {
		for name := range envs {
			origValue, has := originalEnvs[name]
			if has {
				err := os.Setenv(name, origValue)
				if err != nil {
					log.Panic("failed to set environment value: ", err)
				}
			} else {
				err := os.Unsetenv(name)
				if err != nil {
					log.Panic("failed to revert environment value: ", err)
				}
			}
		}
	}
}

func GetHomePath() (string, error) {
	homePath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if !strings.Contains(homePath, "/runai-cli/") {
		return "", fmt.Errorf("couldn't locate CLI folder")
	}
	homePath = strings.SplitAfter(homePath, "runai-cli/")[0]
	return homePath, err
}

func GetConfigPath(homePath, testFileName string) string {
	return filepath.Join(homePath, IgnoreFolder, testFileName)
}

func CreateEnvMap(configPath string) map[string]string {
	envMap := make(map[string]string)
	envMap[consts.EnvRunaiCliConfigPath] = configPath + "/"
	return envMap
}
