/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import "fmt"

const (
	UserFacingFeatureName = "GPU network acceleration (MNNVL)"
)

// Label key
const (
	// FeatureLabelKey is the label key for MNNVL support
	FeatureLabelKey = "run.ai/MNNVL"
)

// Label values
const (
	// None indicates that MNNVL feature is disabled
	None = "None"

	// Preferred indicates that MNNVL is preferred but not required
	// If at least one node pool is MNNVL-supported, the rest must support it too
	Preferred = "Preferred"

	// Required indicates that MNNVL is required
	// Workload will be required to run on MNNVL-supported nodes only
	// If any of the node pools are not MNNVL-supported, the workload will fail
	Required = "Required"
)

// Mode represents the MNNVL support mode
type Mode string

// Modes
const (
	// ModeNone disables the MNNVL feature
	ModeNone Mode = None

	// ModePreferred sets MNNVL as preferred but not required
	// This is the default mode
	ModePreferred Mode = Preferred

	// ModeRequired sets MNNVL as required
	ModeRequired Mode = Required
)

func (m Mode) String() string {
	return string(m)
}

func (m Mode) IsEnabled() bool {
	return m != ModeNone
}

// IsValidMode checks if the provided mode is valid
func IsValidMode(mode Mode) bool {
	switch mode {
	case ModeNone, ModePreferred, ModeRequired:
		return true
	default:
		return false
	}
}

// GetMode converts a string value to Mode
func GetMode(value string) (Mode, error) {
	switch value {
	case string(ModeNone):
		return ModeNone, nil
	case string(ModeRequired):
		return ModeRequired, nil
	case string(ModePreferred):
		return ModePreferred, nil

	}
	return "", fmt.Errorf("invalid %s mode: %s", UserFacingFeatureName, value)
}
