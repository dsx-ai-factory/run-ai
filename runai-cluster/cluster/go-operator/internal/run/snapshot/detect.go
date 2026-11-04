/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package snapshot detects an existing NVIDIA Snapshot installation.
//
// Snapshot is a separate Helm install owned by the cluster administrator.
// Run:ai never deploys it: it looks for it, reports what it found, and activates
// its own capture and restore code only when it is there. A cluster without
// Snapshot must behave exactly as it does today, so every failure path here
// resolves to "unavailable" with a reason rather than to an error that would
// hold up the rest of the reconcile.
package snapshot

import (
	"context"
	"fmt"
	"sort"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	// SnapshotJobCRDName is the CRD whose presence means the Snapshot CRDs are
	// installed. SnapshotJob rather than PodSnapshot because it is the object
	// Run:ai creates.
	SnapshotJobCRDName = "snapshotjobs.nvidia.com"

	// ComponentCRD and ComponentOperator name the sub-components reported
	// under the top-level snapshot status.
	ComponentCRD      = "snapshotjob-crd"
	ComponentOperator = "operator"
)

// Status is what Run:ai knows about the Snapshot installation after looking.
type Status struct {
	// Available is true only when the CRDs are installed and the operator has
	// a ready replica. It is what gates every other part of the integration.
	Available bool
	// Reason explains an unavailable result, and is empty when Available.
	Reason string

	// CRDInstalled and OperatorReady are the two halves of Available, reported
	// separately so an administrator can tell a missing install from a broken
	// one.
	CRDInstalled  bool
	OperatorReady bool
	// CRDReason and OperatorReason explain each half.
	CRDReason      string
	OperatorReason string

	// AgentNodeSelector is the node selector on the Snapshot agent DaemonSet's
	// pod template. Capture and restore pods can only run where the agent runs,
	// so this travels to the anyworkload-controller. Empty means the agent
	// DaemonSet sets no selector, which is to say every node.
	AgentNodeSelector map[string]string
}

// Detect looks for Snapshot and reports what it found. It never returns an
// error: a cluster without Snapshot is the expected case, not a failure.
func Detect(ctx context.Context, reader client.Reader, runaiConfig *runv1.RunaiConfig) Status {
	cfg := runaiConfig.Spec.Snapshot
	namespace := cfg.GetNamespace()

	status := Status{}

	status.CRDInstalled, status.CRDReason = crdInstalled(ctx, uncached(reader))
	status.OperatorReady, status.OperatorReason = operatorReady(ctx, reader, namespace, cfg.OperatorDeploymentName())

	switch {
	case !status.CRDInstalled:
		status.Reason = status.CRDReason
	case !status.OperatorReady:
		status.Reason = status.OperatorReason
	default:
		status.Available = true
	}

	status.AgentNodeSelector = agentNodeSelector(ctx, reader, namespace, cfg.AgentDaemonSetName())

	return status
}

// uncached returns a reader that goes straight to the API server, when the
// caller has one.
func uncached(reader client.Reader) client.Reader {
	if direct, ok := reader.(interface{ Uncached() client.Client }); ok {
		return direct.Uncached()
	}
	return reader
}

func crdInstalled(ctx context.Context, reader client.Reader) (bool, string) {
	if reader == nil {
		return false, "no client available to check for the SnapshotJob CRD"
	}

	crd := &apiextensionsv1.CustomResourceDefinition{}
	if err := reader.Get(ctx, types.NamespacedName{Name: SnapshotJobCRDName}, crd); err != nil {
		if errors.IsNotFound(err) {
			return false, "SnapshotJob CRD is not installed in the cluster"
		}
		return false, fmt.Sprintf("SnapshotJob CRD check failed: %v", err)
	}
	return true, ""
}

func operatorReady(ctx context.Context, reader client.Reader, namespace, name string) (bool, string) {
	if reader == nil {
		return false, "no client available to check the Snapshot operator deployment"
	}

	deployment := &appsv1.Deployment{}
	err := reader.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, fmt.Sprintf("Snapshot operator deployment %s/%s not found", namespace, name)
		}
		return false, fmt.Sprintf("failed to get Snapshot operator deployment %s/%s: %v", namespace, name, err)
	}

	if deployment.Status.AvailableReplicas == 0 {
		return false, "Snapshot operator has no available replicas"
	}
	return true, ""
}

// agentNodeSelector reads the node selector off the agent DaemonSet's pod
// template. A DaemonSet we cannot read yields no selector rather than an error:
// the integration still works, it just places capture and restore pods without
// the hint.
func agentNodeSelector(ctx context.Context, reader client.Reader, namespace, name string) map[string]string {
	if reader == nil {
		return nil
	}

	daemonSet := &appsv1.DaemonSet{}
	if err := reader.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, daemonSet); err != nil {
		return nil
	}
	return daemonSet.Spec.Template.Spec.NodeSelector
}

// FormatNodeSelector renders a node selector as the comma-separated key=value
// list the anyworkload-controller reads from its environment. Keys are sorted
// so that an unchanged cluster produces an unchanged deployment.
func FormatNodeSelector(selector map[string]string) string {
	if len(selector) == 0 {
		return ""
	}

	keys := make([]string, 0, len(selector))
	for key := range selector {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%s", key, selector[key]))
	}
	return strings.Join(pairs, ",")
}
