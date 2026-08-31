/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacy_runtime_runner

import (
	"context"
	"encoding/base64"
	"fmt"

	kc "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	appsv1 "k8s.io/api/apps/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Package legacy_runtime_runner is a transitional, delete-only operand that
// removes the operator-managed resources orphaned when the runtime-rules
// component was renamed. Operator-generated resources carry no
// ownerReferences, so the rename left the pre-rename Deployment, ServiceAccount,
// VPA (and the leader-election Lease on HA installs) with no owner to
// garbage-collect them.
//
// The operand is always disabled: ResourcesForOperands routes a disabled
// operand's resources to deletion, so every reconcile deletes the orphans by
// name — idempotently, and a no-op once they are gone / on fresh installs.
//
// The pre-rename names are held base64-encoded to keep the deprecated term out
// of source ahead of open-sourcing. Remove this operand once every supported
// cluster has upgraded past the rename.

//+kubebuilder:rbac:namespace=runai,groups=coordination.k8s.io,resources=leases,verbs=get;delete

const (
	name = "legacy-runtime-runner-cleanup"

	// base64 of the pre-rename component name; decodes to the orphaned
	// Deployment / ServiceAccount / VPA name.
	legacyNameB64 = "d29ya2xvYWQtb3ZlcnNlZXI="
	// base64 of "<pre-rename name>.run.ai"; decodes to the leader-election Lease
	// name created only on HA (replicas > 1) installs.
	legacyLeaseB64 = "d29ya2xvYWQtb3ZlcnNlZXIucnVuLmFp"
)

type operand struct {
	resources []operands.Resource
}

func (o *operand) Name() string { return name }

func (o *operand) Resources() []operands.Resource { return o.resources }

func (o *operand) Status(context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

// Enabled is always false: this operand exists only to delete its resources.
func (o *operand) Enabled() bool { return false }

func Operand(_ kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	namespace := runaiConfig.GetNamespace()
	legacyName := mustDecodeBase64(legacyNameB64)
	legacyLeaseName := mustDecodeBase64(legacyLeaseB64)

	return &operand{
		resources: []operands.Resource{
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: legacyName, Namespace: namespace}}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: legacyName, Namespace: namespace}}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &vpav1.VerticalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: legacyName, Namespace: namespace}}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &coordinationv1.Lease{ObjectMeta: metav1.ObjectMeta{Name: legacyLeaseName, Namespace: namespace}}
			}),
		},
	}
}

func mustDecodeBase64(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// The inputs are compile-time constants; a decode failure is a bug.
		panic(fmt.Sprintf("legacy_runtime_runner: invalid base64 constant %q: %v", encoded, err))
	}
	return string(decoded)
}
