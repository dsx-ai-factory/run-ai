/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"k8s.io/apimachinery/pkg/types"
)

// WorkloadCache is a map of project to workload uuid and start time
// used to prevent processing (push to the queue) the same workload twice
// there is not a concurrent safe mechanism (for now)

type WorkloadCache map[string]map[string]bool

// Add a workload to the cache
// returns error if the workload was already processed
func (wc *WorkloadCache) Add(projectName string, workloadUuid types.UID, startTime time.Time) error {
	cacheKey := GenerateCacheKey(workloadUuid, startTime.String())

	if wc.WasProcessed(projectName, cacheKey) {
		return fmt.Errorf("was already processed - not adding it to the cache (cacheKey %s)", cacheKey)
	}

	if (*wc)[projectName] == nil {
		(*wc)[projectName] = make(map[string]bool)
	}

	(*wc)[projectName][cacheKey] = true
	return nil
}

// Clean the cache for the given project
func (wc *WorkloadCache) Clean(workloadProject string) {
	(*wc)[workloadProject] = make(map[string]bool)
}

// WasProcessed returns true if the workload was Processed before
// we are using start time to make sure we don't add the same workload twice (after resume)
func (wc *WorkloadCache) WasProcessed(projectName string, cacheKey string) bool {
	if wc == nil {
		// this case should not happen - just in case
		zap.S().Warnf("workloadCache is nil, while checking if project (%s) was processed", projectName)
		return false
	}

	processedInNamespace, found := (*wc)[projectName]
	if !found || processedInNamespace == nil {
		return false
	}

	_, found = processedInNamespace[cacheKey]
	return found
}

func GenerateCacheKey(jobWorkloadUID types.UID, startTime string) string {
	return fmt.Sprintf("%s-%s", jobWorkloadUID, startTime)
}
