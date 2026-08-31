/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

func ValidateSetOfValues(src *string, validValues datastructs.Set[string], fieldName string) *workloadsApi.Error {
	if src == nil {
		return nil
	}

	if !validValues.Has(*src) {
		validValuesAsString := strings.Join(validValues.Keys(), ", ")
		return workloadsApi.NewError(http.StatusBadRequest, fmt.Sprintf("%s: unsupported value '%s', valid options are: %s", fieldName, *src, validValuesAsString))
	}

	return nil
}

// IsEmpty checks if a struct has all zero values
func IsEmpty(v any) bool {
	return reflect.ValueOf(v).IsZero()
}
