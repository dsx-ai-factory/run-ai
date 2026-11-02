/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"strings"
	"testing"
)

func TestPodOwnershipSelectorCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, selector string
		reject         bool
	}{
		{"normal-app-selector", `{"matchLabels":{"app":"test"}}`, false},
		{"conflicting-owner", `{"matchLabels":{"app":"test","run.ai/top-owner-uid":"caller-owner"}}`, true},
		{"canonical-owner", `{"matchLabels":{"run.ai/top-owner-uid":"cp-uuid"}}`, false},
		{"conflicting-workload", `{"matchLabels":{"run.ai/workload-id":"caller-owner"}}`, true},
		{"copied-cluster-owner", `{"matchLabels":{"run.ai/cluster-top-owner-uid":"old-root"}}`, true},
		{"owner-in-expression", `{"matchExpressions":[{"key":"run.ai/top-owner-uid","operator":"In","values":["caller-owner"]}]}`, true},
		{"owner-not-in-expression", `{"matchExpressions":[{"key":"run.ai/top-owner-uid","operator":"NotIn","values":["caller-owner"]}]}`, false},
		{"owner-absent-expression", `{"matchExpressions":[{"key":"run.ai/top-owner-uid","operator":"DoesNotExist"}]}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest := ownershipTestManifest(t, `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"test"},"spec":{"selector":`+tc.selector+`,"template":{"metadata":{"labels":{"app":"test","run.ai/top-owner-uid":"caller-owner","run.ai/workload-id":"caller-owner","run.ai/cluster-top-owner-uid":"old-root"}},"spec":{"containers":[]}}}}`)
			_, stamped, err := ownershipTestStamp(t, manifest, ownershipLiveType(t, "krt-v1-apps-deployment"))
			if tc.reject {
				if err == nil || !strings.Contains(err.Error(), "selector") {
					t.Fatalf("expected clear selector rejection, got stamped=%v err=%v", stamped, err)
				}
			} else if err != nil || !stamped {
				t.Fatalf("compatible selector rejected: stamped=%v err=%v", stamped, err)
			}
		})
	}
}
