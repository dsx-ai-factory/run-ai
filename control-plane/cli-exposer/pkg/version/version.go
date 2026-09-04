/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/cli-exposer/pkg/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/version"
	"github.com/joho/godotenv"
	"os"
	"runtime"
)

// Version information set by link flags during build. We fall back to these sane
// default values when we build outside the Makefile context (e.g. go build or go test).
var (
	buildDate           = "1970-01-01T00:00:00Z" // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
	gitCommit           = "devel-git-commit"     // output from `git rev-parse HEAD`
	controlPlaneVersion = "0.0.1"
	versionFilePath     = "/app/version"
)

func SetVersionFilePath(path string) {
	versionFilePath = path
}

func getVersionFromFile() (string, error) {
	if _, err := os.Stat(versionFilePath); err != nil {
		return "", fmt.Errorf(consts.ErrVersionNotExist, versionFilePath)
	}

	file, err := os.Open(versionFilePath)
	if err != nil {
		return "", fmt.Errorf("file not found %w", err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	myEnv, err := godotenv.Parse(file)
	if err != nil {
		return "", fmt.Errorf("error parse version file %w", err)
	}
	version := fmt.Sprintf("%s.%s.%s", myEnv["RUNAI_CLI_MAJOR"], myEnv["RUNAI_CLI_MINOR"], myEnv["RUNAI_CLI_PATCH"])
	return version, nil
}

// GetVersion returns the version information
func GetVersion() (version.Version, error) {
	cliVersion, err := getVersionFromFile()
	if err != nil {
		return version.Version{}, err
	}

	return version.Version{
		Version:             cliVersion,
		ControlPlaneVersion: controlPlaneVersion,
		BuildDate:           buildDate,
		GitCommit:           gitCommit,
		GoVersion:           runtime.Version(),
		Compiler:            runtime.Compiler,
		Platform:            fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
	}, nil
}
