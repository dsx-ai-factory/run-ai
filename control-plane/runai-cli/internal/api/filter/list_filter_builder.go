/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
)

type Builder struct {
	project           string
	clusterId         string
	workloadTypes     []string
	workloadStatus    string
	workloadFramework api.Framework
	nameList          []string
	createdAt         string
	source            string
}

func NewBuilder() *Builder {
	return &Builder{workloadFramework: api.UndefinedFramework}
}

func (fb *Builder) Build() []Filter {
	var filters []Filter
	if fb.project != "" {
		filters = append(filters, *NewProjectFilter(fb.project))
	}
	if fb.clusterId != "" {
		filters = append(filters, *NewClusterIDFilter(fb.clusterId))
	}
	if len(fb.workloadTypes) > 0 {
		filters = append(filters, *NewWorkloadTypeFilter(strings.Join(fb.workloadTypes, "||")))
	}
	if fb.workloadStatus != "" {
		filters = append(filters, *NewWorkloadStatusFilter(fb.workloadStatus))
	}
	if fb.workloadFramework != api.UndefinedFramework {
		filters = append(filters, GetFrameworkFilters(fb.workloadFramework)...)
	}
	if len(fb.nameList) > 0 {
		filters = append(filters, *NameFilter(strings.Join(fb.nameList, "||")))
	}
	if fb.createdAt != "" {
		filters = append(filters, *NewCreatedAtFilter(fb.createdAt))
	}
	if fb.source != "" {
		filters = append(filters, *NewSourceFilter(fb.source))
	}
	return filters
}

func (fb *Builder) WithSource(source string) *Builder {
	fb.source = source
	return fb
}

func (fb *Builder) WithProject(project string) *Builder {
	fb.project = project
	return fb
}

func (fb *Builder) WithClusterID(clusterId string) *Builder {
	fb.clusterId = clusterId
	return fb
}

func (fb *Builder) WithWorkloadType(workloadType string) *Builder {
	// Only add if not already in the list
	for _, t := range fb.workloadTypes {
		if t == workloadType {
			return fb
		}
	}
	fb.workloadTypes = append(fb.workloadTypes, workloadType)
	return fb
}

func (fb *Builder) WithOnlyWorkloadType(workloadType string) *Builder {
	fb.workloadTypes = []string{workloadType}
	return fb
}

func (fb *Builder) WithWorkloadStatus(workloadStatus string) *Builder {
	fb.workloadStatus = workloadStatus
	return fb
}

func (fb *Builder) AppendName(name string) *Builder {
	fb.nameList = append(fb.nameList, name)
	return fb
}

func (fb *Builder) WithWorkloadFramework(framework api.Framework) *Builder {
	fb.workloadFramework = framework
	return fb
}

func (fb *Builder) WithCreatedAt(time time.Time) *Builder {
	if !time.IsZero() {
		fb.createdAt = time.Format(structs.TimeType)
	} else {
		fb.createdAt = ""
	}
	return fb
}

func (fb *Builder) ClearNames() *Builder {
	fb.nameList = []string{}
	return fb
}

func (fb *Builder) StringList() []string {
	var strList []string
	if fb == nil {
		return strList
	}
	for _, filter := range fb.Build() {
		strList = append(strList, filter.Get())
	}
	return strList
}
