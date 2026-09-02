/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"strings"

	styles "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/tui/styles"
)

type optionsMultipleSelector struct {
	options []Options
}

func NewOptionsMultipleSelector(options []Options) *optionsMultipleSelector {
	selector := &optionsMultipleSelector{
		options: options,
	}
	return selector
}

func (o *optionsMultipleSelector) render(cursor int) string {
	var sb strings.Builder
	for i, option := range o.options {
		radioText := "[ ]"
		if option.IsSelected() {
			radioText = "[X]"
		}

		if i == cursor {
			// Selected item: Render as a highlighted block
			// We strip inner styling to allow the Selected style to apply cleanly to the whole line
			if option.IsDisabled() {
				radioText = "[ ]"
			}
			line := fmt.Sprintf("%s %s", radioText, option.GetUiLabel())
			sb.WriteString(styles.AppStyles.Selected.Render(line) + "\n")
		} else {
			// Non-selected item
			radio := styles.AppStyles.Checkbox.Render(radioText)
			if option.IsDisabled() {
				radio = styles.AppStyles.Checkbox.Render("[ ]")
			}

			style := styles.AppStyles.Normal
			if option.IsDisabled() || option.IsMandatory() {
				style = styles.AppStyles.Disabled
			}

			sb.WriteString(fmt.Sprintf("%s %s\n", radio, style.Render(option.GetUiLabel())))
		}
	}
	return sb.String()
}

func (o *optionsMultipleSelector) selectItem(nextIndex int) {
	option := o.options[nextIndex]
	existingSelected := option.IsSelected()
	if existingSelected && o.options[nextIndex].IsMandatory() {
		return
	}
	option.SetSelected(!existingSelected)
}

func (o *optionsMultipleSelector) isSelected(index int) bool {
	return o.options[index].IsSelected()
}
