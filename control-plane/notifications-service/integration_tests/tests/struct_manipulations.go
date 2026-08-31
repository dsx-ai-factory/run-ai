/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func StructToMap(obj interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}
	err := mapstructure.Decode(obj, &newMap)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return newMap
}
