/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"bytes"
	"encoding/json"
)

func IndentJSONString(prefix string, data []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, data, prefix, "\t")
	if err == nil {
		return out.Bytes()
	}
	return data
}
