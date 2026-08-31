/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/spf13/pflag"
)

type Container struct {
	Command                bool
	ExecuteCommand         string
	ExecuteArgs            string
	Image                  string
	ImagePullPolicy        string
	WorkingDir             string
	CreateHomeDir          bool
	Attach                 bool
	PodRunningTimeout      time.Duration
	WaitForWorkloadSubmit  time.Duration
	TTY                    bool
	Stdin                  bool
	EnvironmentVariables   []string
	PodAffinityPreferred   string
	PodAffinityRequired    string
	FieldRef               []string
	EnvSecret              []string
	EnvMyCredentials       []string
	ImagePullMyCredentials []string

	// RawEnvironment holds values from --environment flag
	RawEnvironment []string
	// EnvironmentAsset holds the environment asset name (resolved to ID before submit)
	EnvironmentAsset string
	// ComputeAsset holds the compute asset name (resolved to ID before submit)
	ComputeAsset string
	// DatasourceAssets holds raw datasource values from --datasource flag (format: type=TYPE,name=NAME or just NAME)
	RawDatasourceAssets []string
	// DatasourceAssetIds holds the resolved datasource IDs (resolved before submit)
	DatasourceAssetIds []string
	// RawMasterDatasourceAssets holds raw datasource values from --master-datasource
	RawMasterDatasourceAssets []string
	// MasterDatasourceAssetIds holds datasource IDs resolved for the master pod
	MasterDatasourceAssetIds []string

	// Distributed Training (Master)
	MasterArgs                 string
	MasterCommand              string
	MasterEnvironmentVariables []string

	// Distributed Inference (Worker)
	WorkerImage                string
	WorkerImagePullPolicy      string
	WorkerCommand              string
	WorkerArgs                 string
	WorkerWorkingDir           string
	WorkerEnvironmentVariables []string
	WorkerCreateHomeDir        bool
	WorkerFieldRef             []string
	WorkerEnvSecret            []string
}

func ContainerFlagSet(containerFlags *pflag.FlagSet, containerSpec *Container, getType spectype.Workload, framework api.Framework) {
	if getType == spectype.NimService {
		containerFlags.StringVarP(&containerSpec.Image, consts.AttributeNameImage, "i", "", consts.AttributeImageUsage)
		containerFlags.StringVar(&containerSpec.ImagePullPolicy, consts.AttributeNameImagePullPolicy, "", consts.AttributeUsageImagePullPolicy)
		containerFlags.StringArrayVarP(&containerSpec.EnvironmentVariables, consts.AttributeNameEnvironmentVariables,
			"e", []string{}, consts.AttributeUsageEnvironmentVariables)
		return
	}

	if getType != spectype.InferenceDistributed {
		containerFlags.BoolVarP(&containerSpec.Command, consts.AttributeNameCommand, "c", false, consts.AttributeUsageCommand)
	}
	containerFlags.StringVarP(&containerSpec.Image, consts.AttributeNameImage, "i", "", consts.AttributeImageUsage)
	containerFlags.StringVar(&containerSpec.ImagePullPolicy, consts.AttributeNameImagePullPolicy, "", consts.AttributeUsageImagePullPolicy)
	containerFlags.StringVar(&containerSpec.WorkingDir, consts.AttributeNameWorkingDir, "", consts.AttributeUsageWorkingDir)
	containerFlags.BoolVar(&containerSpec.CreateHomeDir, consts.AttributeNameCreateHomeDir, false, consts.AttributeUsageCreateHomeDir)
	if getType != spectype.Inference && getType != spectype.InferenceDistributed {
		containerFlags.BoolVar(&containerSpec.Attach, consts.AttributeNameSubmitAttach, false, consts.AttributeUsageSubmitAttach)
	}
	containerFlags.DurationVar(&containerSpec.PodRunningTimeout, consts.AttributeNamePodRunningTimeout, 0, consts.AttributeUsagePodRunningTimeout)
	containerFlags.DurationVar(&containerSpec.WaitForWorkloadSubmit, consts.AttributeNameWaitForWorkloadSubmit, 0, consts.AttributeUsageWaitForWorkloadSubmit)
	containerFlags.StringArrayVar(&containerSpec.FieldRef, consts.AttributeNameEnvPodFieldRef, []string{}, consts.AttributeUsageEnvPodFieldRef)
	containerFlags.StringArrayVar(&containerSpec.EnvSecret, consts.AttributeNameEnvSecret, []string{}, consts.AttributeUsageEnvSecret)
	containerFlags.StringArrayVar(&containerSpec.EnvMyCredentials, consts.AttributeNameEnvMyCredentials, []string{}, consts.AttributeUsageEnvMyCredentials)
	containerFlags.StringArrayVar(&containerSpec.ImagePullMyCredentials, consts.AttributeNameImagePullMyCredentials, []string{}, consts.AttributeUsageImagePullMyCredentials)
	if getType != spectype.Inference && getType != spectype.InferenceDistributed {
		containerFlags.BoolVarP(&containerSpec.TTY, consts.AttributeNameSubmitTTY, "t", false, consts.AttributeUsageSubmitTTY)
		containerFlags.BoolVar(&containerSpec.Stdin, consts.AttributeNameSubmitStdin, false, consts.AttributeUsageSubmitStdin)
	}
	containerFlags.StringVar(&containerSpec.PodAffinityPreferred, consts.AttributeNamePreferredPodAffinityKey, "", consts.AttributeUsagePreferredPodAffinityKey)
	containerFlags.StringVar(&containerSpec.PodAffinityRequired, consts.AttributeNameRequiredPodAffinityKey, "", consts.AttributeUsageRequiredPodAffinityKey)

	if getType == spectype.InferenceDistributed {
		// InferenceDistributed leader args and command
		containerFlags.StringVar(&containerSpec.ExecuteArgs, consts.AttributeNameLeaderArgs, "", consts.AttributeUsageLeaderArgs)
		containerFlags.StringVar(&containerSpec.ExecuteCommand, consts.AttributeNameLeaderCommand, "", consts.AttributeUsageLeaderCommand)
		// InferenceDistributed worker-specific container flags
		DistributedInferenceContainerFlagSet(containerFlags, containerSpec)
	}

	if getType == spectype.TrainingDistributed && framework != api.Jax {
		// Distributed Training master-specific container flags
		containerFlags.StringVar(&containerSpec.MasterArgs, consts.AttributeNameMasterArgs, "", consts.AttributeUsageMasterArgs)
		containerFlags.StringVar(&containerSpec.MasterCommand, consts.AttributeNameMasterCommand, "", consts.AttributeUsageMasterCommand)
		containerFlags.StringArrayVar(&containerSpec.MasterEnvironmentVariables, consts.AttributeNameMasterEnvVariable,
			[]string{}, consts.AttributeUsageMasterEnvironmentVariables)

		// Deprecated: master-environment flag
		//nolint:staticcheck // Intentionally using deprecated constant to register the deprecated flag
		containerFlags.StringArrayVar(&containerSpec.MasterEnvironmentVariables, consts.AttributeNameMasterEnvDeprecated,
			[]string{}, consts.AttributeUsageMasterEnvironmentVariables)
		//nolint:staticcheck // Intentionally using deprecated constant to register the deprecated flag
		_ = containerFlags.MarkDeprecated(consts.AttributeNameMasterEnvDeprecated, "use --master-environment-variable instead")
	}

	// Deprecated flags - should be removed in the future
	oldCommand := "old-command"
	containerFlags.BoolVar(&containerSpec.Command, oldCommand, false, consts.AttributeUsageCommand)
	_ = containerFlags.MarkDeprecated(oldCommand, "use --command instead")

	containerFlags.StringArrayVar(&containerSpec.RawEnvironment, consts.AttributeNameEnvironment, []string{}, consts.AttributeUsageEnvironmentAsset)
	containerFlags.StringArrayVarP(&containerSpec.EnvironmentVariables, consts.AttributeNameEnvironmentVariables,
		"e", []string{}, consts.AttributeUsageEnvironmentVariables)

	containerFlags.StringVar(&containerSpec.ComputeAsset, consts.AttributeNameCompute, "", consts.AttributeUsageComputeAsset)
	containerFlags.StringArrayVar(&containerSpec.RawDatasourceAssets, consts.AttributeNameDatasourceAsset, []string{}, consts.AttributeUsageDatasourceAsset)
	if getType == spectype.TrainingDistributed && framework != api.Jax {
		containerFlags.StringArrayVar(&containerSpec.RawMasterDatasourceAssets, consts.AttributeNameMasterDatasourceAsset, []string{}, consts.AttributeUsageMasterDatasourceAsset)
	}
}

// ProcessRawEnvironment processes the --environment flag values:
// - If value contains "=" (key=value format): treated as environment variable (deprecated)
// - If value does NOT contain "=": treated as environment asset name
func (c *Container) ProcessRawEnvironment(w io.Writer) error {
	var hasEnvVars bool
	for _, v := range c.RawEnvironment {
		if strings.Contains(v, "=") {
			c.EnvironmentVariables = append(c.EnvironmentVariables, v)
			hasEnvVars = true
		} else {
			c.EnvironmentAsset = v
		}
	}
	if hasEnvVars {
		_, err := fmt.Fprintln(w, consts.DeprecatedEnvironmentKeyValueMsg)
		if err != nil {
			return err
		}
	}
	return nil
}
