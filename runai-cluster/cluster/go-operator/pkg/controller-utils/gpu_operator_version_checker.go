/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	"context"
	"fmt"
	"strings"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"golang.org/x/mod/semver"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// This file ports the NVIDIA GPU Operator version-compatibility check from
// github.com/kai-scheduler/kai-gpu-fractioning (operator/internal/controller/dependency_checker.go),
// so Run:ai can surface the same "GPU Operator too old for GPU fractions"
// signal at the cluster level, independent of whether kai-gpu-fractioning is
// installed. Reads and comparison logic are copied verbatim from that source
// to keep the reported message and minimum-version threshold identical.

const (
	clusterPolicyVersionLabel   = "app.kubernetes.io/version"
	minimumGPUOperatorVersion   = "v26.7.1"
	clusterServiceVersionPrefix = "gpu-operator"
)

var clusterPolicyGVK = schema.GroupVersionKind{
	Group:   "nvidia.com",
	Version: "v1",
	Kind:    "ClusterPolicy",
}

var clusterServiceVersionGVK = schema.GroupVersionKind{
	Group:   "operators.coreos.com",
	Version: "v1alpha1",
	Kind:    "ClusterServiceVersion",
}

// CheckGpuOperatorVersion verifies that the cluster's NVIDIA GPU Operator
// meets the minimum version required for GPU fractions (>= v26.7.1). It reads
// the version from the ClusterPolicy CR's "app.kubernetes.io/version" label,
// falling back to the OLM ClusterServiceVersion (for OpenShift installs where
// ClusterPolicy may not be readable). If neither is found, the dependency is
// reported as available, since this checker cannot determine whether a GPU
// Operator is installed at all - that is the responsibility of other checks
// (e.g. CheckGpuOperatorDependency).
func CheckGpuOperatorVersion(ctx context.Context, reader client.Reader) runv1.PublicStatus {
	clusterPolicy, err := getClusterPolicy(ctx, reader)
	if err != nil {
		return runv1.PublicStatus{
			Available: runv1.AvailableStatus(false),
			Reason:    fmt.Sprintf("unable to read NVIDIA GPU Operator ClusterPolicy: %v", err),
		}
	}

	if clusterPolicy == nil {
		version, found, err := clusterServiceVersionGPUOperatorVersion(ctx, reader)
		if err != nil {
			return runv1.PublicStatus{
				Available: runv1.AvailableStatus(false),
				Reason:    fmt.Sprintf("unable to discover NVIDIA GPU Operator version from ClusterPolicy or ClusterServiceVersion: %v", err),
			}
		}
		if !found {
			return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
		}
		if msg := gpuOperatorVersionFailureMessage(version, "ClusterServiceVersion"); msg != "" {
			return runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: msg}
		}
		return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
	}

	version := gpuOperatorVersionFromClusterPolicy(clusterPolicy)
	if msg := gpuOperatorVersionFailureMessage(version, fmt.Sprintf("ClusterPolicy label %q", clusterPolicyVersionLabel)); msg != "" {
		return runv1.PublicStatus{Available: runv1.AvailableStatus(false), Reason: msg}
	}

	return runv1.PublicStatus{Available: runv1.AvailableStatus(true)}
}

func getClusterPolicy(ctx context.Context, reader client.Reader) (*unstructured.Unstructured, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(clusterPolicyGVK.GroupVersion().WithKind(clusterPolicyGVK.Kind + "List"))

	if err := reader.List(ctx, list); err != nil {
		if meta.IsNoMatchError(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(list.Items) == 0 {
		return nil, nil
	}
	return &list.Items[0], nil
}

func clusterServiceVersionGPUOperatorVersion(ctx context.Context, reader client.Reader) (version string, found bool, err error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(clusterServiceVersionGVK.GroupVersion().WithKind(clusterServiceVersionGVK.Kind + "List"))

	if err := reader.List(ctx, list); err != nil {
		if meta.IsNoMatchError(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("failed to list ClusterServiceVersions: %w", err)
	}

	for _, csv := range list.Items {
		if strings.HasPrefix(csv.GetName(), clusterServiceVersionPrefix) {
			version, _, _ := unstructured.NestedString(csv.Object, "spec", "version")
			return strings.TrimSpace(version), true, nil
		}
	}

	return "", false, nil
}

func gpuOperatorVersionFromClusterPolicy(clusterPolicy *unstructured.Unstructured) string {
	return strings.TrimSpace(clusterPolicy.GetLabels()[clusterPolicyVersionLabel])
}

func gpuOperatorVersionFailureMessage(rawVersion, source string) string {
	rawVersion = strings.TrimSpace(rawVersion)
	if rawVersion == "" {
		return fmt.Sprintf("NVIDIA GPU Operator version from %s is missing; minimum supported version is %s",
			source, minimumGPUOperatorVersion)
	}

	version, ok := normalizeGPUOperatorVersion(rawVersion)
	if !ok {
		return fmt.Sprintf("NVIDIA GPU Operator version %q from %s is invalid; minimum supported version is %s",
			rawVersion, source, minimumGPUOperatorVersion)
	}

	// A prerelease from the minimum supported patch train provides the same
	// GPU Operator API baseline. Compare the release cores so, for example,
	// v26.7.1-rc.1 satisfies a v26.7.1 floor while v26.7.0-rc.1 does not.
	versionCore, _ := splitSemverSuffix(version)
	if semver.Compare(versionCore, minimumGPUOperatorVersion) < 0 {
		return fmt.Sprintf("NVIDIA GPU Operator version %s is below minimum supported version %s",
			rawVersion, minimumGPUOperatorVersion)
	}

	return ""
}

func normalizeGPUOperatorVersion(version string) (string, bool) {
	normalized := strings.TrimSpace(version)
	if normalized == "" {
		return "", false
	}
	if !strings.HasPrefix(normalized, "v") {
		normalized = "v" + normalized
	}

	if semver.IsValid(normalized) {
		return semver.Canonical(normalized), true
	}

	main, suffix := splitSemverSuffix(normalized)
	if strings.Count(main, ".") == 1 {
		withPatch := main + ".0" + suffix
		if semver.IsValid(withPatch) {
			return semver.Canonical(withPatch), true
		}
	}

	return "", false
}

func splitSemverSuffix(version string) (main, suffix string) {
	idx := strings.IndexAny(version, "-+")
	if idx == -1 {
		return version, ""
	}
	return version[:idx], version[idx:]
}
