/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package snapshot

import (
	"fmt"
	"strings"

	"k8s.io/utils/ptr"
)

const (
	// DefaultNamespace is the namespace the Snapshot chart's own documentation
	// installs into.
	DefaultNamespace = "snapshot"
	// DefaultReleaseName is the Helm release name that documentation uses.
	DefaultReleaseName = "snapshot"

	// chartName is the Snapshot chart's name, needed to reproduce the chart's
	// own fullname rule.
	chartName = "snapshot"
)

// Snapshot locates an existing Snapshot installation. Snapshot is a separate
// Helm install owned by the cluster administrator; Run:ai never deploys it, it
// only looks for it.
type Snapshot struct {
	// Namespace is the namespace the Snapshot Helm release is installed in.
	// +kubebuilder:validation:Optional
	Namespace *string `json:"namespace,omitempty"`

	// ReleaseName is the Snapshot Helm release name. Run:ai derives the
	// operator Deployment and agent DaemonSet names from it the same way the
	// chart does.
	// +kubebuilder:validation:Optional
	ReleaseName *string `json:"releaseName,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (s *Snapshot) SetDefaultsWhereNeeded() {
	if s.Namespace == nil {
		s.Namespace = ptr.To(DefaultNamespace)
	}
	if s.ReleaseName == nil {
		s.ReleaseName = ptr.To(DefaultReleaseName)
	}
}

// GetNamespace returns the configured namespace, or the default.
func (s *Snapshot) GetNamespace() string {
	if s == nil || s.Namespace == nil {
		return DefaultNamespace
	}
	return *s.Namespace
}

// fullname reproduces the Snapshot chart's "snapshot.fullname" helper: the
// release name on its own when it already contains the chart name, otherwise
// the release name with the chart name appended. Truncated to 63 characters as
// the chart does, because every object name below is built on it.
func (s *Snapshot) fullname() string {
	release := DefaultReleaseName
	if s != nil && s.ReleaseName != nil {
		release = *s.ReleaseName
	}

	name := release
	if !strings.Contains(release, chartName) {
		name = fmt.Sprintf("%s-%s", release, chartName)
	}
	if len(name) > 63 {
		name = name[:63]
	}
	return strings.TrimSuffix(name, "-")
}

// OperatorDeploymentName is the name of the Snapshot operator Deployment.
func (s *Snapshot) OperatorDeploymentName() string {
	return s.fullname() + "-operator"
}

// AgentDaemonSetName is the name of the Snapshot node-agent DaemonSet.
func (s *Snapshot) AgentDaemonSetName() string {
	return s.fullname() + "-agent"
}
