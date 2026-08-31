/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config/attributes"
	"k8s.io/utils/ptr"
)

type Report struct {
	Type     *attributes.ReportType `json:"type"`
	Location *string                `json:"location"`
	Files    *int                   `json:"files"`
	Size     *int                   `json:"size"`
	Age      *int                   `json:"age"`
}

func NewReport(reportType attributes.ReportType, location string, files, size, age int) (*Report, error) {
	switch reportType {
	case attributes.None, attributes.Logger:
		return &Report{Type: &reportType}, nil
	case attributes.Local:
		return &Report{Type: &reportType, Location: &location, Files: &files, Size: &size, Age: &age}, nil
	default:
		return nil, fmt.Errorf("no such reporter type: %s", reportType)
	}
}

type Txn struct {
	Enabled *bool   `json:"enabled"`
	Report  *Report `json:"report"`
}

func NewTxn() *Txn {
	return &Txn{}
}

func (t *Txn) ReportEnabled() bool {
	return ptr.Deref[bool](t.Enabled, false)
}

func (t *Txn) ReportType() attributes.ReportType {
	if t.Report != nil {
		return ptr.Deref[attributes.ReportType](t.Report.Type, attributes.None)
	}
	return attributes.None
}

func (t *Txn) GetLocation() string {
	if t.Report != nil {
		return ptr.Deref[string](t.Report.Location, "")
	}
	return ""
}

func (t *Txn) GetSize() int {
	if t.Report != nil {
		return ptr.Deref[int](t.Report.Size, 0)
	}
	return 0
}

func (t *Txn) GetFiles() int {
	if t.Report != nil {
		return ptr.Deref[int](t.Report.Files, 0)
	}
	return 0
}

func (t *Txn) GetMaxAge() int {
	if t.Report != nil {
		return ptr.Deref[int](t.Report.Age, 0)
	}
	return 0
}
