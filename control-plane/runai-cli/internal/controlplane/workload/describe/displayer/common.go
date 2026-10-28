/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"time"

	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
)

var (
	SectionsSeparator = "\n\n"
)

const (
	// Date format
	DateFormat = time.RFC3339

	// Shared column widths and padding
	GeneralAndRuntimeColumnPadding = 35
	MnnvlPaddingOffset             = 8
	LabelsValuesExtraPadding       = 3
	MaxDisplayLabelsCount          = 100
	MaxDisplayEnvVarsCount         = 100
)

type DisplayOptions struct {
	ShowGeneral       bool
	ShowCompute       bool
	ShowStorage       bool
	ShowPods          bool
	ShowEvents        bool
	ShowNetworks      bool
	IncludeContainers bool
}

type DisplayField struct {
	Name  string
	Value interface{}
}

type Section struct {
	Displayer uitable.Constructor
}

type AfterTableDisplayer interface {
	DisplayAfterTable()
}

type BeforeTableDisplayer interface {
	DisplayBeforeTable(indent string)
}
