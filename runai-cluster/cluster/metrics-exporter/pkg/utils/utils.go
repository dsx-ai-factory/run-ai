/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"strconv"
)

func ParseFloatFromAnnotation(annotations map[string]string, annotationKey string) (float64, error) {
	valString, found := annotations[annotationKey]
	if !found {
		return 0, fmt.Errorf("could not find %s annotation", annotationKey)
	}
	valFloat, err := strconv.ParseFloat(valString, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse annotation %s value='%s', error: %w",
			annotationKey, valString, err)
	}

	return valFloat, nil
}

func BoolFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
