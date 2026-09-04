/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secretvolume

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	MountPath   = "path"
	SecretName  = "name"
	DefaultMode = "default-mode"
	ParserName  = "SecretVolumeParser"
)

var (
	SecretVolume = []string{MountPath, SecretName, DefaultMode}
)

// path=MountPath,name=SecretName

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](SecretVolume...)))
}
