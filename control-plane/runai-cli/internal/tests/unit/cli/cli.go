/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/spf13/cobra"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/pflag"
)

const (
	IgnoreFolder = "ignore_folder"
	ConfigJson   = "config.json"

	ClusterOneUuid   = "uuid-cluster-1"
	ClusterOneName   = "name-cluster-1"
	ClusterOneDomain = "http://localhost.example.com/"
	ClusterTwoUuid   = "uuid-cluster-2"
	ClusterTwoName   = "name-cluster-2"

	ErrorConfigFileAlreadyExist = "runai-cli/ignore_folder/%s/config.json\" Already Exists\n"
)

func RunCommandWithContext(ctx context.Context, command string, env map[string]string) (string, error) {
	config.GetInstance().Reload()
	settings.GetInstance().Reload()

	// One buffer for both streams keeps the chronological interleaving legacy callers expect.
	actual := new(bytes.Buffer)
	err := executeCommand(ctx, strings.Split(command, " "), env, actual, actual)
	return actual.String(), err
}

// RunCommandSeparate is like RunCommandWithContext but captures stdout and stderr
// independently, so tests can assert which stream a message went to.
func RunCommandSeparate(ctx context.Context, command string, env map[string]string) (stdout, stderr string, err error) {
	config.GetInstance().Reload()
	settings.GetInstance().Reload()

	outBuf, errBuf := new(bytes.Buffer), new(bytes.Buffer)
	err = executeCommand(ctx, strings.Split(command, " "), env, outBuf, errBuf)
	return outBuf.String(), errBuf.String(), err
}

func RunCommand(command string, env map[string]string) (string, error) {
	return RunCommandWithContext(context.TODO(), command, env)
}

func RunCommandStringSplit(ctx context.Context, command []string, env map[string]string) (string, error) {
	actual := new(bytes.Buffer)
	err := executeCommand(ctx, command, env, actual, actual)
	return actual.String(), err
}

// executeCommand runs the root command with args, resetting flag state left over
// from previous tests and applying env only for the duration of the call.
func executeCommand(ctx context.Context, args []string, env map[string]string, stdout, stderr io.Writer) error {
	os.Args = append([]string{}, args...)
	cmd := runaicli.RootCmd.Root()
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// starting from root command resetting each flags for each subcommand.
	resetSubCommandFlags(cmd)
	resetCommandFlags(cmd.Commands(), args)

	closer := tempEnvironmentSetter(env)
	defer closer()

	_, err := cmd.ExecuteContextC(ctx)

	cmd.SetOut(nil)
	cmd.SetErr(nil)

	return err
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
