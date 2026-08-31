/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hostpath

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Path             = "path"
	Mount            = "mount"
	ReadWrite        = "readwrite"
	MountPropagation = "mount-propagation"

	ParserName = "HostPathParser"
)

var (
	HostPath = []string{Path, Mount, MountPropagation, ReadWrite}
)

// path=PATH,mount=MOUNT,mount-propagation=None|HostToContainer,readwrite

func NewParser(cmv string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(cmv, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](HostPath...)))
}
