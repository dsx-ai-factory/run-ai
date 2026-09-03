/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

// FakeClusterProbe is a configurable orchestrator2.ClusterProbe for tests. Set
// the version/error and the allowed-permission set; unset permissions are
// denied.
type FakeClusterProbe struct {
	Version    string
	VersionErr error

	// AllowAll grants every permission, for the common "happy path" case.
	AllowAll bool
	// Allowed maps a "verb group/resource" key to whether it is permitted.
	// Any key absent from the map is treated as denied (unless AllowAll).
	Allowed map[string]bool
	CanIErr error

	// Nodes is returned by ListNodes; empty by default so node-sizing checks
	// pass (no offenders) unless a test populates it.
	Nodes        []orchestrator2.NodeInfo
	ListNodesErr error

	// Heartbeats is returned by NodeHeartbeats; empty by default so the
	// clock-skew check passes unless a test populates it.
	Heartbeats        []orchestrator2.NodeHeartbeat
	NodeHeartbeatsErr error

	// NoDefaultStorageClass makes HasDefaultStorageClass report false; the zero
	// value (false) means the cluster has a default class so the check passes.
	NoDefaultStorageClass     bool
	HasDefaultStorageClassErr error

	// AbsentNamespaces are the namespaces NamespaceExists reports as missing;
	// any namespace not listed is reported present (the default), so a check
	// looking one up passes unless a test marks it absent.
	AbsentNamespaces   map[string]bool
	NamespaceExistsErr error

	// IngressClasses is the set of IngressClass names reported present. When
	// AllowAll is set every class is present; otherwise only listed ones are.
	IngressClasses   map[string]bool
	IngressClassErr  error
	GatewayAPIAbsent bool
	GatewayAPIErr    error
	GatewayStatusErr error
	Gateways         map[string]orchestrator2.GatewayStatus // key "namespace/name"
}

func (f *FakeClusterProbe) ServerVersion(context.Context) (string, error) {
	if f.VersionErr != nil {
		return "", f.VersionErr
	}
	return f.Version, nil
}

func (f *FakeClusterProbe) CanI(_ context.Context, ra orchestrator2.ResourceAttributes) (bool, string, error) {
	if f.CanIErr != nil {
		return false, "", f.CanIErr
	}
	if f.AllowAll {
		return true, "", nil
	}
	key := ra.Verb + " " + ra.Group + "/" + ra.Resource
	if f.Allowed[key] {
		return true, "", nil
	}
	return false, "not allowed: " + key, nil
}

func (f *FakeClusterProbe) ListNodes(context.Context) ([]orchestrator2.NodeInfo, error) {
	if f.ListNodesErr != nil {
		return nil, f.ListNodesErr
	}
	return f.Nodes, nil
}

func (f *FakeClusterProbe) NodeHeartbeats(context.Context) ([]orchestrator2.NodeHeartbeat, error) {
	if f.NodeHeartbeatsErr != nil {
		return nil, f.NodeHeartbeatsErr
	}
	return f.Heartbeats, nil
}

func (f *FakeClusterProbe) HasDefaultStorageClass(context.Context) (bool, error) {
	if f.HasDefaultStorageClassErr != nil {
		return false, f.HasDefaultStorageClassErr
	}
	return !f.NoDefaultStorageClass, nil
}

func (f *FakeClusterProbe) NamespaceExists(_ context.Context, name string) (bool, error) {
	if f.NamespaceExistsErr != nil {
		return false, f.NamespaceExistsErr
	}
	return !f.AbsentNamespaces[name], nil
}

func (f *FakeClusterProbe) IngressClassExists(_ context.Context, name string) (bool, error) {
	if f.IngressClassErr != nil {
		return false, f.IngressClassErr
	}
	if f.AllowAll {
		return true, nil
	}
	return f.IngressClasses[name], nil
}

func (f *FakeClusterProbe) GatewayAPIInstalled(_ context.Context) (bool, error) {
	if f.GatewayAPIErr != nil {
		return false, f.GatewayAPIErr
	}
	if f.AllowAll {
		return true, nil
	}
	return !f.GatewayAPIAbsent, nil
}

func (f *FakeClusterProbe) GatewayStatus(_ context.Context, namespace, name string) (orchestrator2.GatewayStatus, error) {
	if f.GatewayStatusErr != nil {
		return orchestrator2.GatewayStatus{}, f.GatewayStatusErr
	}
	if f.AllowAll {
		return orchestrator2.GatewayStatus{Found: true, Accepted: true, Programmed: true}, nil
	}
	return f.Gateways[namespace+"/"+name], nil
}
