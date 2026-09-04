/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmv

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Name        = "name"
	Path        = "path"
	SubPath     = "subpath"
	DefaultMode = "default-mode"

	ParserName = "ConfigMapVolumeParser"
)

var (
	ConfigMapVolume = []string{Name, Path, SubPath, DefaultMode}
)

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](ConfigMapVolume...)))
}
