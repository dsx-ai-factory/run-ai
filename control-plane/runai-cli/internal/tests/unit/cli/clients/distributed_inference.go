/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"encoding/json"
	"fmt"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var updateRequestMap = make(map[string]interface{})

func GetSpecByDistributedInferenceName(name string) (*openapi.DistributedInferenceSpec, error) {
	value, ok := specMap[name]
	if !ok {
		return nil, fmt.Errorf("distributed inference name: %s is not in map", name)
	}
	var byteArr []byte
	byteArr, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	spec := openapi.DistributedInferenceSpec{}
	err = json.Unmarshal(byteArr, &spec)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

func GetUpdateRequestByDistributedInferenceName(name string) (*openapi.UpdateRequest, error) {
	value, ok := updateRequestMap[name]
	if !ok {
		return nil, fmt.Errorf("distributed inference update request for name: %s is not in map", name)
	}
	var byteArr []byte
	byteArr, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	updateReq := openapi.UpdateRequest{}
	err = json.Unmarshal(byteArr, &updateReq)
	if err != nil {
		return nil, err
	}

	return &updateReq, nil
}
