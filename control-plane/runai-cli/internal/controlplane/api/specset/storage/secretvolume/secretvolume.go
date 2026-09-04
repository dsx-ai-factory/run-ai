/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secretvolume

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

func NewList(list []string) ([]openapi.SecretInstance, error) {
	var objList = make([]openapi.SecretInstance, 0)
	for _, host := range list {
		obj, err := New(host)
		if err != nil {
			return nil, err
		}
		objList = append(objList, *obj)
	}
	return objList, nil
}

func New(att string) (*openapi.SecretInstance, error) {
	sv := NewParser(att)
	err := sv.Parse()
	if err != nil {
		return nil, err
	}
	obj, err := create(sv)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func create(sv *parser.AttributeMapper) (*openapi.SecretInstance, error) {
	err := parser.CheckMandatoryFields(sv, MountPath, SecretName)
	if err != nil {
		return nil, err
	}

	path := sv.Get(MountPath)
	secret := sv.Get(SecretName)
	si := &openapi.SecretInstance{
		Secret:    *openapi.NewNullableString(secret),
		MountPath: *openapi.NewNullableString(path),
	}

	defaultMode := sv.Get(DefaultMode)
	if defaultMode != nil && *defaultMode != "" {
		si.SetDefaultMode(*defaultMode)
	}

	return si, nil
}
