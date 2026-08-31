/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package telemetry_utils

import (
	"errors"
	"fmt"
	"strings"
)

func GetModelGroupBy(groupBy []string, modelFieldToDbFieldMap map[string][]string) ([][]string, error) {
	if len(groupBy) > 2 {
		return nil, errors.New("number of groups can be up to 2")
	}
	var modelGroupBy [][]string
	for _, field := range groupBy {
		modelField, ok := modelFieldToDbFieldMap[field]
		if !ok || modelField == nil {
			var availableGroups []string
			for key := range modelFieldToDbFieldMap {
				availableGroups = append(availableGroups, key)
			}
			return nil, fmt.Errorf("group %s is invalid. the group should be only a part of the following: %s", field, strings.Join(availableGroups, ", "))
		}
		modelGroupBy = append(modelGroupBy, modelField)
	}
	return modelGroupBy, nil
}
