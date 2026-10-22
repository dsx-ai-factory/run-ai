/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	inference_endpoint_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/inference-endpoint-controller"

	advice_restart_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/advice-restart-controller"
	application_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/application-controller"
	nodepool_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/nodepool-controller"
	queue_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/queue-controller"
	runai_public "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/runai-public"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/templates"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/common/shared_data"
	gpu_fractioning_metrics "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/gpu-fractioning-metrics"
	init_ca "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/init-ca"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/knative"
	legacy_workload_cleanup "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/legacy-workload-cleanup"
	mpi_operator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/mpi-operator"

	training_operator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/training-operator"

	accessrule_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/accessrule-controller"
	admission_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/admission-controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/agent"
	anyworkload_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/anyworkload-controller"
	cluster_api "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/cluster-api"
	cluster_redis "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/cluster-redis"
	cluster_sync "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/cluster-sync"
	encryption_key "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/encryption-key"
	external_workload_integrator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/external-workload-integrator"
	kai_operator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/kai-operator"
	krm_operator "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/krm-operator"
	metrics_exporter "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/metrics-exporter"
	mps_server "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/mps-server"
	nmc_mgmt_prometheus "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/nmc-mgmt-prometheus"
	project_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/project-controller"
	prometheus "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/prometheus"
	reconciliation_trigger "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/reconciliation-trigger"
	researcher_service "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/researcher-service"
	shared_objects_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/shared-objects-controller"
	status_updater "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/status-updater"

	legacy_runtime_runner "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/legacy-runtime-runner"
	managed_nodes_config "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/managed-nodes-config"
	runtime_rules "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/runtime-rules"
	workload_exporter "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/workload-exporter"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

type ManagedOperands interface {
	GetLocalClusterEnabled() []operands.Operand
	GetLocalClusterOperands() []operands.Operand
	GetRemoteClusterEnabled() []operands.Operand
	GetRemoteClusterOperands() []operands.Operand
}

type managedOperandsImp struct {
	localClusterOperands  []operands.Operand
	remoteClusterOperands []operands.Operand
	ctx                   context.Context
	localClusterReader    kc.HybridClient
	remoteClusterReader   kc.HybridClient
	runaiConfig           *runaiv1.RunaiConfig
}

var _ ManagedOperands = &managedOperandsImp{}

// NewManagedOperands returns a list of operands managed by this controller.
// To add a new operand, implement the corrisponding package, and
// update this slice.
// Note to Dev: if you are changing anything in this func, please make sure to update
// the script in "cluster/new-service-template/hooks/post_gen_project.py" (add_app_to_go_operator_operands func)
func NewManagedOperands(ctx context.Context, localClusterClient kc.HybridClient,
	remoteClusterClient kc.HybridClient, runaiConfig *runaiv1.RunaiConfig) ManagedOperands {
	mo := &managedOperandsImp{}
	mo.ctx = ctx
	mo.localClusterReader = localClusterClient
	mo.remoteClusterReader = remoteClusterClient
	mo.runaiConfig = runaiConfig

	spec := runaiConfig.Spec

	// Used by operands that need to share data with other operands
	promSharedData := shared_data.NewSharedData()

	// When adding a new operand, the default should be to deploy it to the local cluster (management cluster).
	// The local cluster hosts operands that do not depend on hardware resources, such as:
	// - Controllers - Only communicate with the k8s api server
	// - Syncing operands - Only sync resources between a k8s api server and the control-plane

	// New operands should only be deployed to the remote cluster (tenant cluster) if they depend on hardware resources, such as:
	// - GPUs
	// - Nodes
	// - Hardware metrics exporters

	// In a single-cluster setup, operands of both clusters are deployed to the local cluster.

	// For most cases, if your operand's main resource is a Deployment, it should be deployed to the local cluster.
	// If your operand's main resource is a DaemonSet, it should be deployed to the remote cluster.
	// The below operands are deployed to the local cluster (management cluster) in a dual-cluster setup

	// TODO The ecryption-key operand is deployed to both clusters temporarily, as multiple controllers in both clusters depend on it.
	mo.addLocalCluster(encryption_key.Operand(localClusterClient, *runaiConfig))
	if mo.runaiConfig.Spec.Internal.Global.NMC.IsEnabled() {
		mo.addLocalCluster(nmc_mgmt_prometheus.Operand(localClusterClient, runaiConfig, promSharedData))
	}
	mo.addOldLocalCluster(reconciliation_trigger.Operand(), true)
	mo.addLocalCluster(agent.Operand(localClusterClient, runaiConfig))
	mo.addLocalCluster(cluster_sync.Operand(localClusterClient, runaiConfig))
	mo.addLocalCluster(external_workload_integrator.Operand(localClusterClient, runaiConfig))
	mo.addLocalCluster(metrics_exporter.Operand(localClusterClient, runaiConfig))
	mo.addLocalCluster(workload_exporter.Operand(localClusterClient, runaiConfig))
	mo.addLocalCluster(shared_objects_controller.Operand(localClusterClient, runaiConfig))
	mo.addLocalCluster(status_updater.Operand(localClusterClient, runaiConfig))
	mo.addOldLocalCluster(init_ca.Operand(), *spec.InitCA.Enabled)
	mo.addLocalCluster(nodepool_controller.Operand(localClusterClient, runaiConfig))
	mo.addLocalCluster(legacy_workload_cleanup.Operand(localClusterClient, runaiConfig))

	// The below operands are deployed to the remote cluster (tenant cluster) in a dual-cluster setup
	// TODO start migrating all operands to this
	// The controller uses its in-cluster client to manage the serving stack on the workload cluster.
	mo.addRemoteCluster(inference_endpoint_controller.Operand(remoteClusterClient, runaiConfig))
	mo.addRemoteCluster(advice_restart_controller.Operand(remoteClusterClient, runaiConfig))
	mo.addRemoteCluster(runai_public.Operand(remoteClusterClient, *runaiConfig))
	mo.addRemoteCluster(encryption_key.Operand(remoteClusterClient, *runaiConfig))
	mo.addRemoteCluster(templates.Operand(remoteClusterClient, *runaiConfig))
	mo.addRemoteCluster(prometheus.Operand(remoteClusterClient, runaiConfig, promSharedData))
	mo.addRemoteCluster(anyworkload_controller.Operand(remoteClusterClient, runaiConfig))
	// Deletes what the researcher-service retired in 2.26 left behind on upgraded clusters.
	mo.addRemoteCluster(researcher_service.Operand(remoteClusterClient, runaiConfig))
	mo.addOldRemoteCluster(kai_operator.Operand(), *spec.KAIOperator.Enabled)
	mo.addOldRemoteCluster(krm_operator.Operand(), *spec.KRMOperator.Enabled)
	mo.addRemoteCluster(runtime_rules.Operand(remoteClusterClient, runaiConfig))
	// Transitional orphan cleanup for the pre-rename runtime-rules resources.
	// Always disabled, so its resources are deleted every reconcile.
	// Remove once all supported clusters have upgraded past the rename.
	mo.addRemoteCluster(legacy_runtime_runner.Operand(remoteClusterClient, runaiConfig))
	// KRM supplies the project-controller now; registering it disabled deletes
	// the Deployment left behind.
	mo.addRemoteCluster(project_controller.Operand(remoteClusterClient, runaiConfig))
	mo.addOldRemoteCluster(accessrule_controller.Operand(), *spec.AccessRuleController.Enabled && *spec.ProjectController.CreateRoleBindings)
	mo.addOldRemoteCluster(mpi_operator.Operand(), true)
	// Inert until gpu-fractioning is installed: the PodMonitor simply selects no pods.
	mo.addOldRemoteCluster(gpu_fractioning_metrics.Operand(), true)
	mo.addOldRemoteCluster(training_operator.Operand(), true)
	mo.addOldRemoteCluster(admission_controller.Operand(), *spec.AdmissionController.Enabled)
	// mps-server is no longer configurable; registering it disabled deletes the
	// DaemonSet left behind on clusters that had opted in.
	mo.addOldRemoteCluster(mps_server.Operand(), false)
	mo.addRemoteCluster(queue_controller.Operand(remoteClusterClient, runaiConfig))
	mo.addRemoteCluster(cluster_api.Operand(remoteClusterClient, runaiConfig))
	mo.addRemoteCluster(knative.Operand(remoteClusterClient, runaiConfig))
	mo.addRemoteCluster(cluster_redis.Operand(remoteClusterClient, runaiConfig))
	mo.addOldRemoteCluster(managed_nodes_config.Operand(), *spec.NodePoolController.Enabled)
	mo.addOldRemoteCluster(application_controller.Operand(), *spec.ApplicationController.Enabled)

	return mo
}

func (m *managedOperandsImp) addLocalCluster(op operands.Operand) {
	m.localClusterOperands = append(m.localClusterOperands, op)
}

func (m *managedOperandsImp) addRemoteCluster(op operands.Operand) {
	m.remoteClusterOperands = append(m.remoteClusterOperands, op)
}

func (m *managedOperandsImp) addOldRemoteCluster(op operands.OperandOld, enabled bool) {
	operandWrapper := &oldOperandWrapper{
		ctx:            m.ctx,
		runaiConfig:    m.runaiConfig,
		OperandOld:     op,
		OperandEnabled: enabled,
		hybridClient:   m.remoteClusterReader,
	}
	m.remoteClusterOperands = append(m.remoteClusterOperands, operandWrapper)
}

func (m *managedOperandsImp) addOldLocalCluster(op operands.OperandOld, enabled bool) {
	operandWrapper := &oldOperandWrapper{
		ctx:            m.ctx,
		runaiConfig:    m.runaiConfig,
		OperandOld:     op,
		OperandEnabled: enabled,
		hybridClient:   m.localClusterReader,
	}
	m.localClusterOperands = append(m.localClusterOperands, operandWrapper)
}

func (m *managedOperandsImp) GetLocalClusterEnabled() []operands.Operand {
	return getEnabled(m.localClusterOperands)
}

func (m *managedOperandsImp) GetRemoteClusterEnabled() []operands.Operand {
	return getEnabled(m.remoteClusterOperands)
}

func (m *managedOperandsImp) GetLocalClusterOperands() []operands.Operand {
	return m.localClusterOperands
}

func (m *managedOperandsImp) GetRemoteClusterOperands() []operands.Operand {
	return m.remoteClusterOperands
}

func getEnabled(ops []operands.Operand) []operands.Operand {
	enabled := []operands.Operand{}
	for _, operand := range ops {
		if operand.Enabled() {
			enabled = append(enabled, operand)
		}
	}
	return enabled
}
