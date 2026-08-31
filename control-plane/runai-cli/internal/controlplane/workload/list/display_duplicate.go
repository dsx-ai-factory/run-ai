/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"
	"k8s.io/utils/ptr"
)

const (
	DuplicateWorkloadSubject = "Duplicate Workloads Table"
	DefaultFrameworkRunai    = "Runai"
)

type DuplicateOutput struct {
	names     []string
	workloads []manipulate.Brief
}

func NewDuplicateOutput(names []string, workloads []manipulate.Brief) *DuplicateOutput {
	return &DuplicateOutput{names: names, workloads: workloads}
}

func (wd *DuplicateOutput) Subject() string {
	return DuplicateWorkloadSubject
}

func (wd *DuplicateOutput) Columns() []table.Column {
	return []table.Column{
		{Title: "Name", Width: manipulate.DefaultWidth},
		{Title: "Type", Width: manipulate.DefaultWidth},
		{Title: "Framework", Width: manipulate.DefaultWidth},
		{Title: "ID", Width: manipulate.DefaultWidth},
	}
}

func (wd *DuplicateOutput) Data() [][]string {
	var (
		data [][]string
	)

	for _, dup := range wd.names {
		for _, wl := range wd.workloads {
			if dup == wl.Name {
				var row []string
				row = append(row, wl.Name)
				row = append(row, wl.Type)
				row = append(row, ptr.Deref(wl.DistributedFramework, DefaultFrameworkRunai))
				row = append(row, wl.Id)
				data = append(data, row)
			}
		}
	}

	return data
}
