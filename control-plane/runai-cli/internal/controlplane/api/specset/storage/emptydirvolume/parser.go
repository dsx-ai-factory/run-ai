/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package emptydirvolume

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Name      = "name"
	Path      = "path"
	Medium    = "medium"
	SizeLimit = "size-limit"

	ParserName = "EmptyDirVolumeParser"
)

var (
	EmptyDirVolume = []string{Name, Path, Medium, SizeLimit}
)

func NewParser(emptydirvolume string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(emptydirvolume, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](EmptyDirVolume...)))
}
