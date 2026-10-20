/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"fmt"
	"strings"

	uitable "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"
)

func DisplaySection(section *Section) error {
	if len(section.Displayer.Data()) == 0 {
		return nil
	}

	fmt.Printf("%s%s%s", SectionsSeparator, section.Displayer.Subject(), SectionsSeparator) //nolint:forbidigo

	if beforeDisplayer, ok := section.Displayer.(BeforeTableDisplayer); ok {
		beforeDisplayer.DisplayBeforeTable("  ")
	}

	tableDisplay := uitable.NewTableDisplay(section.Displayer)
	output, err := tableDisplay.Generate()
	if err != nil {
		return err
	}
	fmt.Println(output) //nolint:forbidigo

	if afterDisplayer, ok := section.Displayer.(AfterTableDisplayer); ok {
		afterDisplayer.DisplayAfterTable()
	}

	return nil
}

func DisplayIndentedSection(section *Section, title string) {
	if len(section.Displayer.Data()) == 0 {
		return
	}

	fmt.Printf("\n    %s\n", title) //nolint:forbidigo

	if beforeDisplayer, ok := section.Displayer.(BeforeTableDisplayer); ok {
		beforeDisplayer.DisplayBeforeTable("    ")
	}

	tableDisplay := uitable.NewTableDisplay(section.Displayer)
	output, err := tableDisplay.Generate()
	if err != nil {
		fmt.Printf("Error displaying %s: %v\n", title, err) //nolint:forbidigo
		return
	}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("    %s\n", line) //nolint:forbidigo
		}
	}
}

func DisplayFields(fields []DisplayField, indent bool, gpuAcceleration bool) {
	var columnPadding = GeneralAndRuntimeColumnPadding
	format := "%-*s %s\n"
	if indent {
		format = " %-*s %s\n"
	}

	if gpuAcceleration {
		columnPadding += MnnvlPaddingOffset
	}

	for _, field := range fields {
		if value := fmt.Sprintf("%v", field.Value); value != "" && value != "<nil>" {
			fmt.Printf(format, columnPadding, field.Name+":", value) //nolint:forbidigo
		}
	}
}

func DisplayFieldsWithSectionHeader(fields []DisplayField, sectionTitle string, indent bool) {
	if len(fields) == 0 {
		return
	}

	if sectionTitle != "" {
		fmt.Printf("%s%s%s", SectionsSeparator, sectionTitle, SectionsSeparator) //nolint:forbidigo
	}

	DisplayFields(fields, indent, false)
}
