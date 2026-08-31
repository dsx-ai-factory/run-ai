/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toleration

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	Name     = "name"
	Operator = "operator"
	Key      = "key"
	Value    = "value"
	Effect   = "effect"
	Seconds  = "seconds"

	ParserName = "TolerationParser"
)

var (
	KeyMap = []string{Name, Operator, Key, Value, Effect, Seconds}
)

/*
Attribute structure:
name=NAME,operator=Equal|Exists,key=KEY,[value=VALUE],[effect=NoSchedule|NoExecute|PreferNoSchedule],[seconds=SECONDS]

name - mandatory
operator - mandatory, enum (Equal|Exists).
key - optional string
value - optional string
effect - optional enum (NoSchedule|NoExecute|PreferNoSchedule)
seconds - optional int32
*/

func NewParser(port string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(port, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](KeyMap...)))
}
