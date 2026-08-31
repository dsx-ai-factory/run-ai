/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelWorkloadMigrationName = "label-workloads"
	PreVersionLabelFormat      = "Pre-%s"
)

type LabelWorkloads struct {
	k8sClient              client.Client
	wlToMigrate            []WorkloadList
	handler                Handler[client.Object]
	MigrateExecuteCreateFn func(int, []client.Object, func(context.Context, client.Object) error) Executor // for tests only
	PaginationLimit        int64
	migrationMaxWorkers    int
}

type WorkloadList interface {
	client.ObjectList
	GetItems() []client.Object
	GetWorkloadKind() string
}

func (lw *LabelWorkloads) String() string {
	return LabelWorkloadMigrationName
}

func NewLabelWorkloads(k8sClient client.Client, clusterVersion string, wlToMigrateOverride []WorkloadList, handlerOverride Handler[client.Object], migrationMaxWorkers int) (*LabelWorkloads, error) {
	if len(clusterVersion) == 0 {
		return nil, fmt.Errorf("could not create %s migration - cluster version is not available", LabelWorkloadMigrationName)
	}

	wlToMigrate := wlToMigrateOverride
	if wlToMigrate == nil {
		wlToMigrate = []WorkloadList{&twv2.TrainingWorkloadList{}, &iwv2a1.InteractiveWorkloadList{}, &dwv2.DistributedWorkloadList{}}
	}

	handler := handlerOverride
	if handler == nil {
		handler = NewClusterVersionLabelHandler(k8sClient, clusterVersion)
	}

	return &LabelWorkloads{
		k8sClient:           k8sClient,
		wlToMigrate:         wlToMigrate,
		handler:             handler,
		PaginationLimit:     paginationLimit,
		migrationMaxWorkers: migrationMaxWorkers,
	}, nil
}

func (lw *LabelWorkloads) RequiresWebhooks() bool {
	return true
}

func (lw *LabelWorkloads) Migrate(ctx context.Context) error {
	for _, wl := range lw.wlToMigrate {
		wlKind := wl.GetWorkloadKind()
		log.Info(fmt.Sprintf("migrating %ss", wlKind))

		p := NewParallelMigrator[client.Object](
			LabelWorkloadMigrationName,
			lw.k8sClient,
			lw.generateNextBatchGetterFn(wl),
			lw.handler,
			lw.MigrateExecuteCreateFn,
			lw.migrationMaxWorkers,
		)
		err := p.Migrate(ctx)
		if err != nil {
			return err
		}

		log.Info(fmt.Sprintf("done migrating %ss", wlKind))
	}

	return nil
}

func (lw *LabelWorkloads) generateNextBatchGetterFn(wl WorkloadList) func(string, context.Context) ([]client.Object, string, error) {
	return func(continueToken string, ctx context.Context) ([]client.Object, string, error) {
		pagination := client.ListOptions{
			Limit:    lw.PaginationLimit,
			Continue: continueToken,
		}

		err := lw.k8sClient.List(ctx, wl, &pagination)
		if err != nil {
			if meta.IsNoMatchError(err) {
				log.Info("Skipping workload label migration because the API is not served", "workloadKind", wl.GetWorkloadKind())
				return nil, "", nil
			}
			return nil, "", err
		}

		var toMigrate []client.Object
		for _, w := range wl.GetItems() {
			if !hasClusterVersionLabel(w.GetLabels()) {
				toMigrate = append(toMigrate, w)
			}
		}

		return toMigrate, wl.GetContinue(), nil
	}
}

type ClusterVersionLabelHandler struct {
	k8sClient      client.Client
	activeRawPatch client.Patch
}

func NewClusterVersionLabelHandler(k8sClient client.Client, clusterVersion string) *ClusterVersionLabelHandler {
	preVersionLabelValue := fmt.Sprintf(PreVersionLabelFormat, clusterVersion)
	clusterVersionLabelPatch := fmt.Sprintf(`{"metadata":{"labels":{"%s": "%s"}}}`, workloadCommon.CreatedAtClusterVersionLabel, preVersionLabelValue)
	return &ClusterVersionLabelHandler{
		k8sClient:      k8sClient,
		activeRawPatch: client.RawPatch(types.MergePatchType, []byte(clusterVersionLabelPatch)),
	}
}

func (lh *ClusterVersionLabelHandler) Handle(ctx context.Context, workload client.Object) error {
	err := lh.k8sClient.Patch(ctx, workload, lh.activeRawPatch)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
	}
	return nil
}

func hasClusterVersionLabel(labels map[string]string) bool {
	if labels == nil {
		return false
	}

	_, ok := labels[workloadCommon.CreatedAtClusterVersionLabel]
	return ok
}
