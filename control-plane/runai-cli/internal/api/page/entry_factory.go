/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package page

import (
	"fmt"
	"time"

	ousc "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	wsc "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type EntryAdapter interface {
	GetId() string
	GetCreatedAt() time.Time
}

func GetEntryAdapter(entry any) (EntryAdapter, error) {
	switch castedEntry := entry.(type) {
	case wsc.Workload:
		return NewWorkloadEntryAdapter(&castedEntry), nil
	case ousc.Project:
		return NewProjectEntryAdapter(&castedEntry), nil
	case wsc.AIApplication:
		return NewAIApplicationEntryAdapter(&castedEntry), nil
	default:
		return nil, fmt.Errorf("unknown pager entry type: %T", entry)
	}
}
