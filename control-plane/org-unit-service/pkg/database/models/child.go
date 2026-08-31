/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

type Child struct {
	Id   int64
	Name string
}

func convertRawToChild(raw map[string]interface{}) Child {
	child := Child{}

	if name, ok := raw["name"]; ok && name != nil {
		child.Name = name.(string)
	}
	if id, ok := raw["id"]; ok && id != nil {
		child.Id = int64(id.(float64))
	}

	return child
}
