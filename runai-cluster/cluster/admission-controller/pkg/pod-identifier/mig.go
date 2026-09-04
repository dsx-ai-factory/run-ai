/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_identifier

import (
	"strings"
)

const migResourcePrefix = "nvidia.com/mig"

func HasMigDeviceRequest(annotations map[string]string) bool {
	for annotation := range annotations {
		if strings.HasPrefix(annotation, migResourcePrefix) {
			return true
		}
	}
	return false
}
