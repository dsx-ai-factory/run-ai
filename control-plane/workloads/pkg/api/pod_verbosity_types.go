/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
)

type Verbosity int32

const (
	VerbosityBrief Verbosity = iota
	VerbosityFull
)

const DefaultVerbosity = VerbosityBrief

// ParsePodVerbosity parses a verbosity string into a Verbosity object.
// Returns an error if the string is not a valid verbosity.
func ParsePodVerbosity(verbosity api.PodVerbosity) (Verbosity, *api.Error) {
	switch verbosity {
	case api.PODVERBOSITY_BRIEF:
		return VerbosityBrief, nil
	case api.PODVERBOSITY_FULL:
		return VerbosityFull, nil
	case "":
		return DefaultVerbosity, nil
	default:
		return -1, api.NewError(http.StatusBadRequest, fmt.Sprintf("Invalid verbosity: %s", verbosity), nil)
	}
}
