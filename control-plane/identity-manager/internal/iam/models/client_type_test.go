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
		{name: "access key", clientType: ClientTypeUserApp, want: "user-application", wantOk: true},
		{name: "service account", clientType: ClientTypeTenantApp, want: "tenant-application", wantOk: true},
		{name: "never renamed", clientType: ClientTypeWorkloadsAuth},
		{name: "already legacy", clientType: LegacyClientTypeUserApp},
		{name: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := LegacyClientType(tc.clientType)
			assert.Equal(t, tc.wantOk, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsUserAppClientType(t *testing.T) {
	for _, tc := range []struct {
		name       string
		clientType string
		want       bool
	}{
		{name: "current value", clientType: ClientTypeUserApp, want: true},
		{name: "legacy value", clientType: LegacyClientTypeUserApp, want: true},
		{name: "service account", clientType: ClientTypeTenantApp},
		{name: "legacy service account", clientType: LegacyClientTypeTenantApp},
		{name: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, IsUserAppClientType(tc.clientType))
		})
	}
}
