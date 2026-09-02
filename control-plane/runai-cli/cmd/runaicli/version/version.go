/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/version"
	"github.com/spf13/cobra"
)

var (
	wide bool
)

// Cmd represents the version command
var Cmd = &cobra.Command{
	Use:     "version",
	Short:   consts.VersionShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupConfigID,

	Run: func(cmd *cobra.Command, args []string) {
		var verString string
		ver := version.GetVersion()

		if wide {
			verString = fmt.Sprintf("runai-cli/%s sdk/%s go/%s git/%s build/%s platform/%s",
				ver.ControlPlaneVersion,
				ver.Version,
				ver.GoVersion,
				ver.GitCommit,
				ver.BuildDate,
				ver.Platform)
		} else {
			verString = fmt.Sprintf("runai-cli/%s sdk/%s",
				ver.ControlPlaneVersion,
				ver.Version)
		}

		cmd.Println(verString)
	},
}

func init() {
	Cmd.Flags().BoolVar(&wide, "wide", false, "print full version details")
}
