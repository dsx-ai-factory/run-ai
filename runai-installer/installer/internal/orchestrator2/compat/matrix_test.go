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
		Dynamo:              "1.4.2",
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
