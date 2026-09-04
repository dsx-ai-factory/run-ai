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

type optionsSingleSelector struct {
	options  []Options
	selected int
}

func NewOptionsSingleSelect(options []Options) *optionsSingleSelector {
	selector := &optionsSingleSelector{
		options:  options,
		selected: -1,
	}
	for i, val := range options {
		if val.IsSelected() {
			selector.selected = i
			break
		}
	}
	return selector
}

func (o *optionsSingleSelector) render(cursor int) string {
	var sb strings.Builder
	for i, option := range o.options {
		radioText := "( )"
		if i == o.selected {
			radioText = "(•)"
		}

		if i == cursor {
			line := fmt.Sprintf("%s %s", radioText, option.GetUiLabel())
			sb.WriteString(styles.AppStyles.Selected.Render(line) + "\n")
		} else {
			radio := styles.AppStyles.Checkbox.Render(radioText)
			if i == o.selected {
				radio = styles.AppStyles.CheckboxSelected.Render(radioText)
			}
			style := styles.AppStyles.Normal
			sb.WriteString(fmt.Sprintf("%s %s\n", radio, style.Render(option.GetUiLabel())))
		}
	}
	return sb.String()
}

func (o *optionsSingleSelector) selectItem(nextIndex int) {
	if o.selected != -1 {
		o.options[o.selected].SetSelected(false)
	}
	o.selected = nextIndex
	o.options[nextIndex].SetSelected(true)
}

func (o *optionsSingleSelector) isSelected(index int) bool {
	return o.selected == index
}
