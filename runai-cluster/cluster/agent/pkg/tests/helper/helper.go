/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helper

// helper package provides ensure that the encryption key is set in the environment variable
// before calling the init function of the handlers package
// this is used only for testing purposes

import (
	"os"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

func init() {
	// if there is already a key in the environment variable, do not generate a new one
	key := os.Getenv(constants.RunaiAdsEncryptionKeyEnvVar)
	if key != "" {
		return
	}
	newKey := "905c18f0839affcddca117dcac633f79"
	_ = os.Setenv(constants.RunaiAdsEncryptionKeyEnvVar, newKey)
}
