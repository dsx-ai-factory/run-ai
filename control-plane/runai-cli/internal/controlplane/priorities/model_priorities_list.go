/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package priorities

import (
	"fmt"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type Brief struct {
	Name string `json:"name"`
}

func newBriefFromClusterPriority(priority openapi.ClusterPriority) Brief {
	return Brief{priority.Name}
}

func BriefListFromResponse(prioritiesList []openapi.ClusterPriority) ([]Brief, error) {
	if prioritiesList == nil {
		return nil, fmt.Errorf("no priority suggestions available")
	}
	var briefList []Brief

	for _, priority := range prioritiesList {
		briefList = append(briefList, newBriefFromClusterPriority(priority))
	}

	return briefList, nil
}
