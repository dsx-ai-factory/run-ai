/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/parsers"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

const (
	typeNullableResources = "resources"
	defaultGpuDeserved    = 0
	defaultGpuLimit       = -1
	defaultNodepoolName   = "default"
)

type NullableResourcesFlag struct {
	values *[]openapi.ResourcesNullable
}

func NewNullableResourcesFlag(values *[]openapi.ResourcesNullable) *NullableResourcesFlag {
	return &NullableResourcesFlag{values: values}
}

func (rf *NullableResourcesFlag) String() string {
	if len(*rf.values) == 0 {
		return ""
	}

	var result []string
	for _, value := range *rf.values {
		var parts []string

		// Add name if available
		if value.HasNodePool() {
			nodePool := value.GetNodePool()
			if nodePool.HasName() {
				parts = append(parts, fmt.Sprintf("nodepool=%s", nodePool.GetName()))
			}
		}

		// Add priority if available
		if value.HasPriority() {
			parts = append(parts, fmt.Sprintf("priority=%s", value.GetPriority()))
		}

		// Add GPU properties if available
		if value.HasGpu() {
			gpu := value.GetGpu()
			if gpu.HasDeserved() {
				parts = append(parts, fmt.Sprintf("gpu-deserved=%.2f", gpu.GetDeserved()))
			}
			if gpu.HasLimit() {
				parts = append(parts, fmt.Sprintf("gpu-limit=%.2f", gpu.GetLimit()))
			}
		}

		result = append(result, strings.Join(parts, ","))
	}

	return strings.Join(result, " ")
}

func (rf *NullableResourcesFlag) IsSet() bool {
	return len(*rf.values) > 0
}

func errorMessagePrefix(nodepoolName string) string {
	return fmt.Sprintf("for nodepool <%s>:", nodepoolName)
}

func (rf *NullableResourcesFlag) Set(value string) error {
	am := parsers.NewNodepoolParser(value)

	err := am.Parse()
	if err != nil {
		return err
	}

	newResource := openapi.NewResourcesNullable()

	// Set NodePool if nodepool is provided
	nodepoolName := am.Get(parsers.Nodepool)
	if nodepoolName == nil {
		nodepoolName = ptr.To(defaultNodepoolName)
	}
	nodepool := openapi.NewResourcesNodePoolWithDefaults()
	nodepool.SetName(*nodepoolName)
	newResource.SetNodePool(*nodepool)

	// Set Priority if provided
	if priority := am.Get(parsers.Priority); priority != nil {
		newResource.SetPriority(*priority)
	}

	// Set GPU properties if provided
	gpuLimit := am.Get(parsers.GpuLimit)
	gpuDeserved := am.Get(parsers.GpuDeserved)
	gpu := openapi.NewResourcesNullableGpu()

	if gpuDeserved != nil {
		deserved, err := strconv.ParseFloat(*gpuDeserved, 64)
		if err != nil {
			return fmt.Errorf("%s invalid gpu-deserved value: %s", errorMessagePrefix(*nodepoolName), *gpuDeserved)
		}
		if deserved < 0 {
			return fmt.Errorf("%s invalid gpu-deserved value: [%s], must be greater than or equal to 0", errorMessagePrefix(*nodepoolName), *gpuDeserved)
		}
		gpu.SetDeserved(deserved)
	} else {
		gpu.SetDeserved(defaultGpuDeserved)
	}

	if gpuLimit != nil {
		limit, err := strconv.ParseFloat(*gpuLimit, 64)
		if err != nil {
			return fmt.Errorf("%s invalid gpu-limit value: %s", errorMessagePrefix(*nodepoolName), *gpuLimit)
		}
		if limit < -1 {
			return fmt.Errorf("%s invalid gpu-limit value: [%s], must be greater than or equal to -1", errorMessagePrefix(*nodepoolName), *gpuLimit)
		}
		gpu.SetLimit(limit)
	} else {
		gpu.SetLimit(defaultGpuLimit)
	}

	if gpu.GetLimit() >= 0 && gpu.GetLimit() < gpu.GetDeserved() {
		return fmt.Errorf("%s invalid gpu-limit (%.1f) and gpu-deserved (%.1f) values: gpu-limit should be higher or equal to gpu-deserved",
			errorMessagePrefix(*nodepoolName), gpu.GetLimit(), gpu.GetDeserved())
	}
	newResource.SetGpu(*gpu)

	*rf.values = append(*rf.values, *newResource)
	return nil
}

// Type for CLI help value
func (rf *NullableResourcesFlag) Type() string {
	return typeNullableResources
}
