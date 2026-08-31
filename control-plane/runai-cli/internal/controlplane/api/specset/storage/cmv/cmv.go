/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmv

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func NewConfigMapVolumeList(cmvList []string) ([]openapi.ConfigMapInstance, error) {
	var cmvObjList = make([]openapi.ConfigMapInstance, 0)
	for _, cmv := range cmvList {
		cmvObj, err := NewConfigMapVolume(cmv)
		if err != nil {
			return nil, err
		}
		cmvObjList = append(cmvObjList, *cmvObj)
	}
	return cmvObjList, nil
}

func NewConfigMapVolume(cmv string) (*openapi.ConfigMapInstance, error) {
	cmvParser := NewParser(cmv)
	err := cmvParser.Parse()
	if err != nil {
		return nil, err
	}

	cm, err := createConfigMap(cmvParser)
	if err != nil {
		return nil, err
	}

	subPath := cmvParser.Get(SubPath)

	if subPath != nil {
		cm.SetSubPath(*subPath)
	}

	defaultMode := cmvParser.Get(DefaultMode)

	if defaultMode != nil && *defaultMode != "" {
		cm.SetDefaultMode(*defaultMode)
	}

	return cm, nil
}

func createConfigMap(cmvParser *parser.AttributeMapper) (*openapi.ConfigMapInstance, error) {
	err := parser.CheckMandatoryFields(cmvParser, Name, Path)
	if err != nil {
		return nil, err
	}

	name := cmvParser.Get(Name)
	path := cmvParser.Get(Path)

	return &openapi.ConfigMapInstance{
		ConfigMap: *openapi.NewNullableString(name),
		MountPath: *openapi.NewNullableString(path),
	}, nil
}
