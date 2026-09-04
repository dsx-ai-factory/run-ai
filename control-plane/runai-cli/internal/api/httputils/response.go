/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package httputils

import (
	"fmt"
	"net/http"

	clierrors "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/errors"
)

const (
	FailedApiCall = "API call failed, status code error: %s"
)

func HandleResponse[T any](model *T, response *http.Response, err error) (*T, error) {
	if err != nil {
		return nil, clierrors.HandleResponseError(response, err)
	}

	if response == nil {
		return nil, fmt.Errorf("response is nil")
	}

	defer func() {
		if response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if !CheckResponseStatusCodeOK(response) {
		return nil, fmt.Errorf(FailedApiCall, response.Status)
	}

	return model, nil
}

func CheckResponseStatusCodeOK(response *http.Response) bool {
	return response.StatusCode >= http.StatusOK && response.StatusCode <= http.StatusIMUsed
}
