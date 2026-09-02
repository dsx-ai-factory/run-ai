/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version

import (
	"fmt"
	"runtime"
)

// Version information set by link flags during build. We fall back to these sane
// default values when we build outside the Makefile context (e.g. go build or go test).
var (
	buildDate    = "1970-01-01T00:00:00Z" // output from `date -u +'%Y-%m-%dT%H:%M:%SZ'`
	gitCommit    = "devel-git-commit"     // output from `git rev-parse HEAD`
	version      = "0.0.1-local"
	controlPlane = "0.10.0"
)

// Version contains Arena version information
type Version struct {
	Version             string
	ControlPlaneVersion string
	BuildDate           string
	GitCommit           string
	GoVersion           string
	Compiler            string
	Platform            string
}

func (v Version) String() string {
	return v.Version
}

// GetVersion returns the version information
func GetVersion() Version {
	return Version{
		Version:             version,
		ControlPlaneVersion: controlPlane,
		BuildDate:           buildDate,
		GitCommit:           gitCommit,
		GoVersion:           runtime.Version(),
		Compiler:            runtime.Compiler,
		Platform:            fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
	}
}
