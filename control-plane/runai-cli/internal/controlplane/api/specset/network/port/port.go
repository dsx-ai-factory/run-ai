/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package port

import (
	"strconv"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

func NewPortList(portList []string) ([]openapi.Port, error) {
	var portObjList = make([]openapi.Port, 0)
	for _, port := range portList {
		portObj, err := NewPort(port)
		if err != nil {
			return nil, err
		}
		portObjList = append(portObjList, *portObj)
	}
	return portObjList, nil
}

func NewPort(port string) (*openapi.Port, error) {
	var (
		external int64
		err      error
	)

	portParser := NewParser(port)
	err = portParser.Parse()
	if err != nil {
		return nil, err
	}

	portObj, err := createPort(portParser)
	if err != nil {
		return nil, err
	}

	if portParser.Get(External) != nil {
		external, err = strconv.ParseInt(*portParser.Get(External), 10, 32)
		if err != nil {
			return nil, err
		}
		portObj.SetExternal(int32(external))
	}
	return portObj, nil
}

func createPort(gp *parser.AttributeMapper) (*openapi.Port, error) {
	var (
		container int64
		st        *openapi.PortServiceType
	)
	err := parser.CheckMandatoryFields(gp, Container, ServiceType)
	if err != nil {
		return nil, err
	}

	container, err = strconv.ParseInt(*gp.Get(Container), 10, 32)
	if err != nil {
		return nil, err
	}
	st, err = openapi.NewPortServiceTypeFromValue(*gp.Get(ServiceType))
	if err != nil {
		return nil, err
	}
	return &openapi.Port{
		Container:   *openapi.NewNullableInt32(openapi.PtrInt32(int32(container))),
		ServiceType: *openapi.NewNullablePortServiceType(st),
	}, nil
}
