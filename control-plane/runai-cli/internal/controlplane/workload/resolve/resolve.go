/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package resolve resolves the workload identifiers used for type-specific cluster API
// operations (logs, exec, attach, port-forward).
//
// # Background
//
// When a submission is redirected through workloads-manager (for example under
// ANY_WORKLOAD_SUBMISSION), the workload is represented in the cluster by a
// WorkloadRunner-managed resource rather than by the native CR the CLI would normally address
// (TrainingWorkload, InteractiveWorkload, PytorchJob, ...). The backing resource can be a
// batch/v1 Job, an InteractiveWorkload CR, a TrainingWorkload CR, or any other kind produced
// by the WorkloadRunner - tenants may disable batch/v1 Job creation and keep the CRD-backed
// variants, so the resolver must not assume a particular backing kind. Type-specific
// cluster operations that address the workload by its native CR name therefore fail to locate
// the backing object even though the workload and its pods are healthy.
//
// The cluster API exposes a generic external/external/{name} path (see
// internal/cluster/url) that resolves the workload via workloads-manager regardless of the
// backing kind. This package looks up the workload in the control-plane; if it has a backing
// external resource, callers are switched onto the external/external path, otherwise the
// caller-provided native identifiers are kept.
package resolve

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
)

const workloadListLimit = 10

// ClusterWorkload adjusts the workload identifiers used for cluster API calls so that any
// submission redirected through workloads-manager is addressed via its backing external
// resource, and native submissions keep addressing their native CR.
//
// The control-plane workload record's ExternalWorkloadName field is the source of truth:
//
//   - Set: the submission is redirected through workloads-manager (regardless of whether the
//     backing kind is a batch/v1 Job, an InteractiveWorkload/TrainingWorkload CR, or any
//     other WorkloadRunner-managed resource). The cluster API is addressed via
//     external/external/{externalName}.
//   - Unset: the submission is native and the caller-provided identifiers are kept.
//
// Resolution is applied to every framework/type except NimService, which has its own
// external cluster API routing that does not depend on workloads-manager (see
// api.Framework.ClusterAPIIdentifiers and the URL builder). Applying resolution broadly means
// distributed and inference workloads are also handled correctly when they move behind
// workloads-manager.
func ClusterWorkload(instance *config.Instance, txn *transaction.Txn, workloadType api.WorkloadType,
	framework api.Framework, project, name string) (api.WorkloadType, api.Framework, string, error) {
	if !shouldResolve(framework) {
		return workloadType, framework, name, nil
	}

	cpURL := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterID := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()

	externalName, err := ExternalWorkloadName(txn, cpURL, token, clusterID, project, name)
	if err != nil {
		return workloadType, framework, name, err
	}
	if externalName == "" {
		// Native submission: keep the caller-provided identifiers so the cluster URL is
		// built against the native CR (TrainingWorkload, InteractiveWorkload, PytorchJob,
		// etc.).
		return workloadType, framework, name, nil
	}

	// Redirected submission: address the WorkloadRunner-managed external resource. Setting
	// the framework to UndefinedFramework causes the cluster URL builder and status client
	// to route the request through the external/external path (regardless of the backing
	// resource kind). The workloadType is preserved so callers that inspect it after
	// resolution still see the original (semantically correct) type.
	return workloadType, api.UndefinedFramework, externalName, nil
}

// shouldResolve reports whether the given framework participates in the workloads-manager
// redirection flow. Every framework does except:
//   - NimService: has its own external cluster API routing (ew-nimservice-<name> under
//     external/external) that is unrelated to workloads-manager.
//   - UndefinedFramework: signals that the caller has already resolved the workload (either
//     via this package or via the generic `runai workload` command's initializer) and the
//     cluster URL will be routed through external/external. Skipping resolution here makes
//     ClusterWorkload safely idempotent.
func shouldResolve(framework api.Framework) bool {
	return framework != api.NimService && framework != api.UndefinedFramework
}

// ExternalWorkloadName looks up the workload by name in the given project and returns the name
// of its backing external resource, if any. It returns an empty string when the workload is a
// native submission with no external backing.
func ExternalWorkloadName(txn *transaction.Txn, cpURL, token, clusterID, project, name string) (string, error) {
	if _, err := projectservice.GetProjectId(txn, cpURL, token, clusterID, project); err != nil {
		return "", err
	}

	workloadClient, err := client.NewWorkloadServiceClient(cpURL, token)
	if err != nil {
		return "", err
	}

	fb := filter.NewBuilder().
		WithClusterID(clusterID).
		WithProject(project).
		AppendName(name)

	workloadList, err := workloadClient.List(txn.GetCtx(), workloadListLimit, 0, fb.StringList(), false)
	if err != nil {
		return "", err
	}
	if len(workloadList.Workloads) == 0 {
		return "", fmt.Errorf("workload %s not found in project %s", name, project)
	}

	workload, err := workloadClient.GetWorkload(txn.GetCtx(), workloadList.Workloads[0].Id)
	if err != nil {
		return "", err
	}
	if workload.ExternalWorkloadName != nil {
		return *workload.ExternalWorkloadName, nil
	}
	return "", nil
}
