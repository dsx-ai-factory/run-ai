/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import "fmt"

type Verbosity int32

const (
	VerbosityMetadata Verbosity = iota
	VerbosityFull
)

const DefaultVerbosity = VerbosityFull

// ParseVerbosity parses a verbosity string into a Verbosity object.
// Returns an error if the string is not a valid verbosity.
func ParseVerbosity(verbosityStr string) (Verbosity, error) {
	switch verbosityStr {
	case "metadata":
		return VerbosityMetadata, nil
	case "full":
		return VerbosityFull, nil
	case "":
		return DefaultVerbosity, nil
	default:
		return -1, fmt.Errorf("invalid verbosity: %s", verbosityStr)
	}
}
