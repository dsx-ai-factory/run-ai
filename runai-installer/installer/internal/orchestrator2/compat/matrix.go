/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package compat holds the version compatibility matrix that pins each
// prerequisite/feature to a known-good version per major.minor Run:ai release.
// Plan resolution consults it to fill any dependency version the spec leaves
// unset. An unrecognized Run:ai version is not fatal: Resolve falls back to the
// latest known release's locks and reports that it did so, so the caller can
// surface a warning.
package compat

import (
	"sort"
	"strings"
)

// Component identifies a prerequisite/feature whose version is locked per
// Run:ai release.
type Component string

const (
	Prometheus      Component = "prometheus"
	GPUOperator     Component = "gpuOperator"
	GPUOperatorFake Component = "gpuOperatorFake"
	NetworkOperator Component = "networkOperator"
	Knative         Component = "knative"
	Kubeflow        Component = "kubeflow"
	MPIOperator     Component = "mpiOperator"
	LWS             Component = "lws"
	NIMOperator     Component = "nimOperator"
	Grove           Component = "grove"
	Dynamo          Component = "dynamo"
	// NFD is the OpenShift Node Feature Discovery operator, a dependency of the
	// GPU and network operators on OpenShift. Its lock is the OLM subscription
	// channel (e.g. "stable") rather than a chart version, since it installs via
	// OperatorHub. The NVIDIA certified operators' (GPU, network, NIM) OpenShift
	// channels are derived from their versions, so they need no separate lock.
	NFD Component = "nfd"
	// SRIOVOperator is the OpenShift SR-IOV Network Operator, a network-operator
	// dependency installed via OperatorHub. Its lock is an OLM subscription
	// channel.
	SRIOVOperator Component = "sriovOperator"
	// Serverless is the OLM subscription channel of the Red Hat OpenShift
	// Serverless operator — Knative on OpenShift. Red Hat publishes every release
	// into a single "stable" channel, so the channel is fixed and the exact
	// version is pinned separately via ServerlessVersion (the Knative lock above
	// is the upstream Helm operator version used off OpenShift).
	Serverless Component = "serverless"
	// ServerlessVersion is the Red Hat OpenShift Serverless operator version that
	// pins the Subscription's startingCSV (serverless-operator.<version>), so the
	// OperatorHub install lands on a known version rather than the channel head.
	// It is a Red Hat operator version (serverless-operator.vX.Y.Z), unrelated to
	// the upstream Knative version used off OpenShift.
	ServerlessVersion Component = "serverlessVersion"
	// LWSOperator is the OLM subscription channel of the Red Hat Leader Worker
	// Set operator — LWS on OpenShift (the LWS lock above is the upstream Helm
	// chart version used off OpenShift).
	LWSOperator Component = "lwsOperator"
	// CertManagerOperator is the OLM subscription channel of the cert-manager
	// Operator for Red Hat OpenShift, which the Leader Worker Set operator
	// requires for its webhook certificates.
	CertManagerOperator Component = "certManagerOperator"
)

// Locks maps each component to its pinned version for a given Run:ai release.
type Locks map[Component]string

// latest must always point at the newest key in matrix.
const latest = "3.0"

// matrix is keyed by Run:ai major.minor. Only the newest row is fully
// populated; older rows are filled in as their real locks are confirmed.
var matrix = map[string]Locks{
	"3.0": {
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
	},
	"2.26": {
		Prometheus:          "76.2.0",
		GPUOperator:         "v25.3.0",
		GPUOperatorFake:     "0.0.62",
		NetworkOperator:     "v25.10.0",
		Knative:             "1.18.2",
		Kubeflow:            "v1.8.1",
		MPIOperator:         "v0.6.0",
		LWS:                 "0.7.0",
		NIMOperator:         "3.1.1",
		Grove:               "v0.1.0-alpha.10-rc1",
		Dynamo:              "1.2.1",
		NFD:                 "stable",
		SRIOVOperator:       "stable",
		Serverless:          "stable",
		ServerlessVersion:   "v1.36.1",
		LWSOperator:         "stable-v1.0",
		CertManagerOperator: "stable-v1",
	},
	"2.25": {
		Prometheus:          "76.2.0",
		GPUOperator:         "v25.3.0",
		GPUOperatorFake:     "0.0.62",
		NetworkOperator:     "v25.10.0",
		Knative:             "1.18.2",
		Kubeflow:            "v1.8.1",
		MPIOperator:         "v0.6.0",
		LWS:                 "0.7.0",
		NIMOperator:         "3.1.1",
		Grove:               "v0.1.0-alpha.10-rc1",
		Dynamo:              "1.2.1",
		NFD:                 "stable",
		SRIOVOperator:       "stable",
		Serverless:          "stable",
		ServerlessVersion:   "v1.36.1",
		LWSOperator:         "stable-v1.0",
		CertManagerOperator: "stable-v1",
	},
}

// LatestVersion is the newest Run:ai major.minor the matrix knows about; it is
// the release whose locks are used when an unrecognized version falls back.
func LatestVersion() string { return latest }

// SupportedVersions returns the Run:ai major.minor releases the matrix knows
// about, sorted, so callers can surface them (e.g. in a validation error).
func SupportedVersions() []string {
	out := make([]string, 0, len(matrix))
	for v := range matrix {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// Resolve returns the locks for a Run:ai version and whether it fell back to
// the latest known release because the version was unrecognized.
func Resolve(runaiVersion string) (Locks, bool) {
	if locks, ok := matrix[majorMinor(runaiVersion)]; ok {
		return locks, false
	}
	return matrix[latest], true
}

// VersionFor returns specVersion when set, otherwise the matrix lock for the
// component. Precedence: explicit spec value > matrix lock.
func (l Locks) VersionFor(c Component, specVersion string) string {
	if specVersion != "" {
		return specVersion
	}
	return l[c]
}

// majorMinor reduces a version like "2.26.1" or "v2.26.0" to "2.26".
func majorMinor(version string) string {
	v := strings.TrimPrefix(version, "v")
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return v
	}
	return parts[0] + "." + parts[1]
}
