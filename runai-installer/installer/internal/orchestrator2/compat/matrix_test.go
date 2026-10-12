/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compat

import (
	"reflect"
	"testing"
)

func TestThreeZeroMatrix(t *testing.T) {
	expected := Locks{
		Prometheus:          "76.2.0",
		GPUOperator:         "v26.7.1",
		GPUOperatorFake:     "0.0.62",
		NetworkOperator:     "v26.7.0",
		Knative:             "1.23.0",
		Kubeflow:            "v1.9.4",
		MPIOperator:         "v0.8.2",
		LWS:                 "0.10.0",
		NIMOperator:         "3.1.2",
		Grove:               "v0.1.0-alpha.12-rc1",
		Dynamo:              "1.5.1",
		NFD:                 "stable",
		SRIOVOperator:       "stable",
		Serverless:          "stable",
		ServerlessVersion:   "v1.36.1",
		LWSOperator:         "stable-v1.0",
		CertManagerOperator: "stable-v1",
	}

	locks, fellBack := Resolve("3.0.17")
	if fellBack {
		t.Fatal("Resolve(3.0.17) fell back")
	}
	if !reflect.DeepEqual(locks, expected) {
		t.Fatalf("Resolve(3.0.17) = %#v, want %#v", locks, expected)
	}
}

func TestLatestMatrix(t *testing.T) {
	if LatestVersion() != "3.0" {
		t.Fatalf("LatestVersion() = %q, want %q", LatestVersion(), "3.0")
	}

	locks, fellBack := Resolve("0.0.0-branch")
	if !fellBack {
		t.Fatal("Resolve(0.0.0-branch) did not fall back")
	}
	if locks[GPUOperator] != "v26.7.1" {
		t.Fatalf("Resolve(0.0.0-branch)[GPUOperator] = %q, want %q", locks[GPUOperator], "v26.7.1")
	}
}

func TestDynamoVersionResolution(t *testing.T) {
	for _, tc := range []struct {
		name         string
		runaiVersion string
		specVersion  string
		want         string
	}{
		{name: "3.0", runaiVersion: "3.0.0", want: "1.5.1"},
		{name: "3.0 prerelease", runaiVersion: "v3.0.0-rc.1", want: "1.5.1"},
		{name: "later minor", runaiVersion: "3.1.0", want: "1.5.1"},
		{name: "later major", runaiVersion: "4.0.0", want: "1.5.1"},
		{name: "development build", runaiVersion: "0.0.0-branch", want: "1.5.1"},
		{name: "2.26", runaiVersion: "2.26.1", want: "1.2.1"},
		{name: "2.25", runaiVersion: "2.25.1", want: "1.2.1"},
		{name: "explicit override", runaiVersion: "3.0.0", specVersion: "1.4.2", want: "1.4.2"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			locks, _ := Resolve(tc.runaiVersion)
			if got := locks.VersionFor(Dynamo, tc.specVersion); got != tc.want {
				t.Fatalf("Resolve(%q).VersionFor(Dynamo, %q) = %q, want %q", tc.runaiVersion, tc.specVersion, got, tc.want)
			}
		})
	}
}
