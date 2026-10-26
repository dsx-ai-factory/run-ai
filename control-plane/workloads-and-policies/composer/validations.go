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

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
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
