/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_actions

import (
	"encoding/json"
	"fmt"

	jsonpatch "github.com/evanphx/json-patch"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

// Patches RunaiConfig using JSON Merge Patch (RFC 7396)
func PatchRunaiConfig(runaiConfig *runaiv1.RunaiConfig, patchBytes json.RawMessage) error {
	if runaiConfig == nil {
		return fmt.Errorf("runai config is nil")
	}

	// Handle nil patch gracefully - treat as no-op
	if patchBytes == nil {
		return nil
	}

	originalBytes, err := json.Marshal(&runaiConfig.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal original runai config: %w", err)
	}

	mergedBytes, err := jsonpatch.MergePatch(originalBytes, patchBytes)
	if err != nil {
		return fmt.Errorf("failed to apply merge patch: %w", err)
	}

	// Unmarshalling to a new spec to avoid keeping old values that were deleted in the patch
	var newSpec runaiv1.RunaiConfigSpec
	err = json.Unmarshal(mergedBytes, &newSpec)
	if err != nil {
		return fmt.Errorf("failed to unmarshal patched runai config: %w", err)
	}

	runaiConfig.Spec = newSpec

	return nil
}
