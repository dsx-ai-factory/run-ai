/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package display

import "testing"

func TestHumanize(t *testing.T) {
	cases := map[string]string{
		"preflight-node-cpu":             "Preflight node CPU",
		"preflight-rbac":                 "Preflight RBAC",
		"install-gpu-operator":           "Install GPU operator",
		"install-control-plane-tls":      "Install control plane TLS",
		"create-nic-cluster-policy":      "Create NIC cluster policy",
		"install-sriov-network-operator": "Install SR-IOV network operator",
		"register-cluster":               "Register cluster",
		"":                               "",
	}
	for in, want := range cases {
		if got := Humanize(in); got != want {
			t.Errorf("Humanize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestOr(t *testing.T) {
	if got := Or("Custom Title", "install-lws"); got != "Custom Title" {
		t.Errorf("Or with custom = %q, want %q", got, "Custom Title")
	}
	if got := Or("", "install-lws"); got != "Install LWS" {
		t.Errorf("Or fallback = %q, want %q", got, "Install LWS")
	}
}
