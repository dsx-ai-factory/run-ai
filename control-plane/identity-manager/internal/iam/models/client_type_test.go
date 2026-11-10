/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLegacyClientType(t *testing.T) {
	for _, tc := range []struct {
		name       string
		clientType string
		want       string
		wantOk     bool
	}{
		{name: "access key", clientType: ClientTypeAccessKey, want: "user-application", wantOk: true},
		{name: "service account", clientType: ClientTypeServiceAccount, want: "tenant-application", wantOk: true},
		{name: "never renamed", clientType: ClientTypeWorkloadsAuth},
		{name: "already legacy", clientType: LegacyClientTypeAccessKey},
		{name: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := LegacyClientType(tc.clientType)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsAccessKeyClientType(t *testing.T) {
	for _, tc := range []struct {
		name       string
		clientType string
		want       bool
	}{
		{name: "current value", clientType: ClientTypeAccessKey, want: true},
		{name: "legacy value", clientType: LegacyClientTypeAccessKey, want: true},
		{name: "service account", clientType: ClientTypeServiceAccount},
		{name: "legacy service account", clientType: LegacyClientTypeServiceAccount},
		{name: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsAccessKeyClientType(tc.clientType))
		})
	}
}
