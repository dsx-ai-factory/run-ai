/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/submit"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/spf13/pflag"
)

func ContainerUpdateFlagSet(containerFlags *pflag.FlagSet, containerSpec *submit.Container, getType spectype.Workload) {
	if getType == spectype.NimService {
		containerFlags.StringVarP(&containerSpec.Image, consts.AttributeNameImage, "i", "", consts.AttributeImageUsage)
		containerFlags.StringVar(&containerSpec.ImagePullPolicy, consts.AttributeNameImagePullPolicy, "", consts.AttributeUsageImagePullPolicy)
		containerFlags.StringArrayVarP(&containerSpec.EnvironmentVariables, consts.AttributeNameEnvironmentVariables, "e", []string{}, consts.AttributeUsageEnvironmentVariables)
		return
	}

	containerFlags.BoolVarP(&containerSpec.Command, consts.AttributeNameCommand, "c", false, consts.AttributeUsageCommand)
	containerFlags.StringVarP(&containerSpec.Image, consts.AttributeNameImage, "i", "", consts.AttributeImageUsage)
	containerFlags.StringVar(&containerSpec.ImagePullPolicy, consts.AttributeNameImagePullPolicy, "", consts.AttributeUsageImagePullPolicy)
	containerFlags.StringVar(&containerSpec.WorkingDir, consts.AttributeNameWorkingDir, "", consts.AttributeUsageWorkingDir)
	containerFlags.BoolVar(&containerSpec.CreateHomeDir, consts.AttributeNameCreateHomeDir, false, consts.AttributeUsageCreateHomeDir)
	containerFlags.StringArrayVarP(&containerSpec.EnvironmentVariables, consts.AttributeNameEnvironmentVariables, "e", []string{}, consts.AttributeUsageEnvironmentVariables)
	// --environment flag: for environment asset name or environment variables
	containerFlags.StringArrayVar(&containerSpec.RawEnvironment, consts.AttributeNameEnvironment, []string{}, consts.AttributeUsageEnvironmentAsset)
}
