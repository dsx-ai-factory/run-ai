/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"fmt"
	"strings"
)

type KeyValuePair interface {
	GetName() string
	GetValue() string
}

func GetKeyValueDisplay[T KeyValuePair](items []T) string {
	itemsBuilder := strings.Builder{}
	itemsBuilder.WriteString("[")

	// add padding so the labels values will be aligned to the values columns (each label is displayed in new line)
	padding := GeneralAndRuntimeColumnPadding + LabelsValuesExtraPadding

	var displayedLabelsCount int
	if items != nil {
		// limit the number of displayed labels
		displayedLabelsCount = min(MaxDisplayLabelsCount, len(items))
		lastIndex := displayedLabelsCount - 1

		formattedLabels := make([]string, displayedLabelsCount)
		for i := 0; i <= lastIndex; i++ {
			formattedLabels[i] = fmt.Sprintf("%s:%s", items[i].GetName(), items[i].GetValue())
		}
		itemsBuilder.WriteString(strings.Join(formattedLabels, fmt.Sprintf(",%-*s", padding, "\n")))
	}

	itemsBuilder.WriteString("]")

	if len(items) > MaxDisplayLabelsCount {
		itemsBuilder.WriteString(fmt.Sprintf("%-*sDisplaying %d/%d labels", padding-1, "\n", displayedLabelsCount, len(items)))
	}

	return itemsBuilder.String()
}
