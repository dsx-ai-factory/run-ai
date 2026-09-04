/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

/*
example:
claimname=CLAIM_NAME,storageclass=STORAGE_CLASS,size=SIZE,path=PATH,accessmode-rwo,accessmode-rom,accessmode-rwm,ro,ephemeral
*/

const (
	ClaimName     = "claimname"
	StorageClass  = "storageclass"
	Size          = "size"
	Path          = "path"
	AccessModeRwo = "accessmode-rwo"
	AccessModeRom = "accessmode-rom"
	AccessModeRwm = "accessmode-rwm"
	ReadOnly      = "ro"
	Ephemeral     = "ephemeral"

	ParserNameNew      = "NewPvcParser"
	ParserNameExisting = "ExistingPvc"
)

var (
	NewPvcMap = []string{ClaimName, StorageClass, Size, Path,
		AccessModeRwo, AccessModeRom, AccessModeRwm,
		ReadOnly, Ephemeral}

	ExistingPvcMap = []string{ClaimName, Path}
)

func NewParser(pvc string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(pvc, parser.NewLooseMapParser(ParserNameNew, mapset.NewSet[string](NewPvcMap...)))
}

func ExistingParser(pvc string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(pvc, parser.NewStrictMapParser(ParserNameExisting, mapset.NewSet[string](ExistingPvcMap...)))
}
