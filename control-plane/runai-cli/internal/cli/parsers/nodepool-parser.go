/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

/*
example:
nodepool=nodepool-name,gpu-deserved=10,gpu-limit=2,priority=Low
*/

const (
	Nodepool    = "nodepool"
	GpuDeserved = "gpu-deserved"
	GpuLimit    = "gpu-limit"
	Priority    = "priority"

	ParserNameNew = "NewNodepoolParser"
)

var (
	NewProjectNodepoolMap = []string{Nodepool, GpuDeserved, GpuLimit, Priority}
)

func NewNodepoolParser(resources string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(resources, parser.NewLooseMapParser(ParserNameNew, mapset.NewSet[string](NewProjectNodepoolMap...)))
}
