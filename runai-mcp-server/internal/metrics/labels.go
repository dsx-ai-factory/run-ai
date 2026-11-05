/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxLabelLen = 64

// normalizeLabel truncates method strings that exceed the Prometheus label length
// limit, appending a short hash so distinct long names remain distinguishable.
func normalizeLabel(v string) string {
	if len(v) <= maxLabelLen {
		return v
	}
	// Zero pad the hash: %x drops leading zeros, so slicing its output to a
	// fixed width panics for the roughly one in sixteen inputs whose hash starts
	// with a zero nibble.
	h := fmt.Sprintf("%08x", fnv1a(v))
	return v[:48] + "~" + h
}

func fnv1a(s string) uint32 {
	var h uint32 = 2166136261
	for i := 0; i < len(s); i++ {
		h ^= uint32(s[i])
		h *= 16777619
	}
	return h
}

// toolLabel returns the tool name if it is in knownTools, otherwise "unknown".
// This bounds the cardinality of the tool_name label to the set of registered tools.
func toolLabel(params mcp.Params, knownTools map[string]bool) string {
	var name string
	switch p := params.(type) {
	case *mcp.CallToolParams:
		name = p.Name
	case *mcp.CallToolParamsRaw:
		name = p.Name
	}
	if knownTools[name] {
		return name
	}
	return "unknown"
}
