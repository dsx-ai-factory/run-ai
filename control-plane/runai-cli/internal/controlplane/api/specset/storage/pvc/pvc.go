/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/parser"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func NewPvcList(pvcList []string, exist bool) ([]openapi.PvcInstance, error) {
	var (
		opPvcList []openapi.PvcInstance
		pvcObj    *openapi.PvcInstance
		err       error
	)
	for _, pvc := range pvcList {
		if !exist {
			pvcObj, err = NewPvc(pvc)
		} else {
			pvcObj, err = ExistingPvc(pvc)
		}
		if err != nil {
			return opPvcList, err
		}
		opPvcList = append(opPvcList, *pvcObj)
	}
	return opPvcList, nil
}

func ExistingPvc(pvc string) (*openapi.PvcInstance, error) {
	am := ExistingParser(pvc)
	err := am.Parse()
	if err != nil {
		return nil, err
	}
	return createPvc(am), nil
}

func createPvc(am *parser.AttributeMapper) *openapi.PvcInstance {
	// no need to validate, this is a strict parser strategy, which enforce existing values
	cn := am.Get(ClaimName)
	path := am.Get(Path)
	pvc := &openapi.PvcInstance{
		ClaimName: *openapi.NewNullableString(cn),
		Path:      *openapi.NewNullableString(path),
	}
	pvc.SetExistingPvc(true)
	return pvc
}

func NewPvc(pvc string) (*openapi.PvcInstance, error) {
	am := NewParser(pvc)

	newPvc, err := setNewPvc(am)
	if err != nil {
		return nil, fmt.Errorf("%s, for claim: %s", err.Error(), pvc)
	}

	claimInfo := &openapi.ClaimInfo{
		Size: *openapi.NewNullableString(am.Get(Size)),
	}
	setClaimInfo(am, claimInfo, newPvc)
	return newPvc, nil
}

func setNewPvc(am *parser.AttributeMapper) (*openapi.PvcInstance, error) {
	err := am.Parse()
	if err != nil {
		return nil, err
	}

	err = parser.CheckMandatoryFields(am, ClaimName, Path, Size)
	if err != nil {
		return nil, err
	}

	cn := am.Get(ClaimName)
	path := am.Get(Path)
	newPvc := &openapi.PvcInstance{
		ClaimName: *openapi.NewNullableString(cn),
		Path:      *openapi.NewNullableString(path),
	}
	newPvc.SetExistingPvc(false)

	ro := am.Get(ReadOnly)
	if ro != nil {
		newPvc.SetReadOnly(true)
	}
	e := am.Get(Ephemeral)
	if e != nil {
		newPvc.SetEphemeral(true)
	}
	return newPvc, nil
}

func setClaimInfo(am *parser.AttributeMapper, claimInfo *openapi.ClaimInfo, newPvc *openapi.PvcInstance) {
	sc := am.Get(StorageClass)
	if sc != nil {
		claimInfo.SetStorageClass(*sc)
	}
	setAccessMode(am, claimInfo)
	newPvc.SetClaimInfo(*claimInfo)
}

func setAccessMode(am *parser.AttributeMapper, claimInfo *openapi.ClaimInfo) {
	rwo := am.Get(AccessModeRwo)
	rwm := am.Get(AccessModeRwm)
	rom := am.Get(AccessModeRom)
	if rwo != nil || rwm != nil || rom != nil {
		accessMode := openapi.PvcAccessModes{}
		if rwo != nil {
			accessMode.SetReadWriteOnce(true)
		}
		if rwm != nil {
			accessMode.SetReadWriteMany(true)
		}
		if rom != nil {
			accessMode.SetReadOnlyMany(true)
		}
		claimInfo.SetAccessModes(accessMode)
	}
}
