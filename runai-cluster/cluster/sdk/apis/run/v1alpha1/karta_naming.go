/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1alpha1

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	kartaNamePrefix               = "krt"
	kartaNameMaxLen               = 62
	kartaNameSuffixIfTrailingDash = "a"
)

// GenerateKartaNameByGvk returns the canonical Karta CR name for a GVK.
// Both agent and karta-migrate must call this so the cluster ends up with one
// canonical Karta per GVK, regardless of which side created it.
func GenerateKartaNameByGvk(group, version, kind string) (string, error) {
	if version == "" || kind == "" {
		return "", fmt.Errorf("karta name requires non-empty version and kind (got group=%q version=%q kind=%q)", group, version, kind)
	}
	raw := fmt.Sprintf("%s-%s-%s-%s", kartaNamePrefix, version, group, kind)
	return convertToK8sResourceName(raw), nil
}

func convertToK8sResourceName(name string) string {
	result := name
	if len(result) > kartaNameMaxLen {
		result = result[:kartaNameMaxLen]
	}
	result = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' {
			return r
		}
		return '-'
	}, result)
	if len(result) > 0 && result[len(result)-1] == '-' {
		result += kartaNameSuffixIfTrailingDash
	}
	return strings.ToLower(result)
}
