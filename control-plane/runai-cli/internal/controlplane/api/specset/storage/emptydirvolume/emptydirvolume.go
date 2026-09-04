/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package emptydirvolume

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

func NewEmptyDirVolumeList(emptydirvolumeList []string) ([]openapi.EmptyDirInstance, error) {
	var emptyDirVolumeObjList = make([]openapi.EmptyDirInstance, 0)
	for _, emptydirvolume := range emptydirvolumeList {
		emptyDirVolumeObj, err := NewEmptyDirVolume(emptydirvolume)
		if err != nil {
			return nil, err
		}
		emptyDirVolumeObjList = append(emptyDirVolumeObjList, *emptyDirVolumeObj)
	}
	return emptyDirVolumeObjList, nil
}

func NewEmptyDirVolume(emptydirvolume string) (*openapi.EmptyDirInstance, error) {
	emptyDirVolume := NewParser(emptydirvolume)
	err := emptyDirVolume.Parse()
	if err != nil {
		return nil, err
	}
	emptyDirVolumeObj, err := createEmptyDirVolume(emptyDirVolume)
	if err != nil {
		return nil, err
	}
	return emptyDirVolumeObj, nil
}

func createEmptyDirVolume(emptyDirVolume *parser.AttributeMapper) (*openapi.EmptyDirInstance, error) {
	err := parser.CheckMandatoryFields(emptyDirVolume, Name, Path, Medium, SizeLimit)
	if err != nil {
		return nil, err
	}
	name := emptyDirVolume.Get(Name)
	path := emptyDirVolume.Get(Path)
	medium := emptyDirVolume.Get(Medium)
	sizeLimit := emptyDirVolume.Get(SizeLimit)
	return &openapi.EmptyDirInstance{
		Name:      *openapi.NewNullableString(name),
		Path:      *openapi.NewNullableString(path),
		Medium:    *openapi.NewNullableString(medium),
		SizeLimit: *openapi.NewNullableString(sizeLimit),
	}, nil
}
