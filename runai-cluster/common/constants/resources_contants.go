/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const acceleratorLabelNamePrefix = "accelerator-"

// AcceleratorResourceLabelPrefix is reserved for the operator's derived resource
// labels. The Kubernetes NodeRestriction admission plugin protects this namespace.
const AcceleratorResourceLabelPrefix = "run.ai.node-restriction.kubernetes.io/" + acceleratorLabelNamePrefix

// AcceleratorResourceNodeLabel encodes an extended resource name into a valid,
// stable label key, including names longer than the 63-character label name limit.
func AcceleratorResourceNodeLabel(resourceName string) string {
	// DNS resource domains contain no underscores, so replacing the slash
	// keeps short names distinct and makes nodegroup templates readable.
	if len(acceleratorLabelNamePrefix)+len(resourceName) <= 63 {
		return AcceleratorResourceLabelPrefix + strings.Replace(resourceName, "/", "_", 1)
	}
	digest := sha256.Sum256([]byte(resourceName))
	return AcceleratorResourceLabelPrefix + hex.EncodeToString(digest[:16])
}

const (
	GpuMemoryResourceName = "run.ai/gpu.memory"

	NvidiaGpuResourceName = "nvidia.com/gpu"
	AmdGpuResourceName    = "amd.com/gpu"
)
