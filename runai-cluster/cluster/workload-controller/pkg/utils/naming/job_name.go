/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package naming

import (
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils"
)

func generateSuffix() string {
	uuidParts := strings.Split(utils.ActiveUuidGenerator.Get(0), "-")
	// The last part of v4 uuid is 48 bit long, which means there are 2^48 possible options
	return uuidParts[len(uuidParts)-1]
}

// GenerateJobNameByPrefix returns a job name that starts with the given prefix and ends with a random suffix,
// which is based on the last part of a v4 uuid, separated with a '-'
func GenerateJobNameByPrefix(prefix string) string {
	suffix := generateSuffix()
	jobName := fmt.Sprintf("%s-%s", prefix, suffix)

	return jobName
}
