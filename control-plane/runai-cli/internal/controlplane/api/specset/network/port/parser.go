/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package port

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
)

const (
	ServiceType = "service-type"
	Container   = "container"
	External    = "external"

	ParserName = "PortParser"
)

var (
	Port = []string{ServiceType, Container, External}
)

/*
Attribute structure:
service-type=NodePort,container=80,external=8080

service-type - mandatory, currently has one value (NodePort).
container - mandatory, int32.
external - optional, set external port, if not set, random value will be generated
*/

func NewParser(port string) *parser.AttributeMapper {
	return parser.NewAttributeMapper(port, parser.NewLooseMapParser(ParserName, mapset.NewSet[string](Port...)))
}
