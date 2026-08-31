/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"encoding/json"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func GetSpecByInferenceName(wsName string) (*openapi.InferenceSpecSpec, error) {
	value, ok := specMap[wsName]
	if !ok {
		return nil, fmt.Errorf("inference name: %s is not in map", wsName)
	}
	var byteArr []byte
	byteArr, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	wlSpec := openapi.InferenceSpecSpec{}

	err = json.Unmarshal(byteArr, &wlSpec)
	if err != nil {
		return nil, err
	}

	return &wlSpec, nil
}
