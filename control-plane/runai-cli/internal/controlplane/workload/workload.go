/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/batch"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	clusterApi "github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	controlApi "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/client"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

const (
	timerListTxn    = "workload_list_txn"
	TimerBulkDelete = "workload_bulk_delete"
)

var (
	batcher   batch.Batcher
	paginator list.Paginator

	ErrNilOptions = errors.New("options is nil")
)

func SetPaginator(client list.Paginator) {
	paginator = client
}

func SetBatcher(client batch.Batcher) {
	batcher = client
}

func GetWorkloadPagination(txn *transaction.Txn, domain, token string, pageOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
	var (
		workloadList []openapi.Workload
		next         *structs.NextToken
		err          error
	)
	if paginator == nil {
		paginator, err = client.NewWorkloadServiceClient(domain, token)
		if err != nil {
			return nil, err
		}
	}
	if pageOptions == nil {
		return nil, ErrNilOptions
	}

	// calculate the entire transaction time
	timer := txn.AddTimer(timerListTxn)
	defer txn.StopTimer(timer)

	workloadService := list.NewPager(paginator)
	workloadList, next, err = workloadService.GetWorkloads(txn, pageOptions, fb, deleted)
	if err != nil {
		return nil, err
	}
	return list.NewOutput(workloadList, next, deleted)
}

func GetWorkloadByFilter(txn *transaction.Txn, cpUrl, token string, fb *filter.Builder) (*list.Output, error) {
	handler, err := GetWorkloadPagination(txn, cpUrl, token, apilist.NewFilterPageOption(), fb, false)
	if err != nil {
		return nil, err
	}

	switch len(handler.GetList()) {
	case 0:
		return nil, fmt.Errorf(consts.NoWorkloadWithNameFound)
	case 1:
		return handler, nil
	default:
		return handler, fmt.Errorf(consts.ErrorMoreThanOneWithNameFound)
	}
}

func SearchClusterWorkload(ctx context.Context, reqWorkload *controlApi.Workload, controlUrl, token string) (*clusterApi.Workload, error) {
	workloadClient, err := client.NewWorkloadServiceClient(controlUrl, token)
	if err != nil {
		return nil, err
	}
	clients := describe.NewService(workloadClient, nil, nil, nil)

	workload, err := clients.GetWorkload(ctx, reqWorkload, 0, 0)
	if err != nil {
		return nil, err
	}
	workloadTypeApi, err := api.ParseWorkloadType(strings.ToLower(workload.General.Type))
	if err != nil {
		return nil, err
	}
	return clusterApi.NewWorkload(reqWorkload.Project, workloadTypeApi, reqWorkload.Name, workloadTypeApi.DefaultFramework()), nil
}

func DeleteBulkByIds(txn *transaction.Txn, domain, token string, ids []string) (*openapi.BatchResponse, error) {
	var err error

	timerID := txn.AddTimer(TimerBulkDelete)
	defer txn.StopTimer(timerID)

	if batcher == nil {
		batcher, err = client.NewWorkloadServiceClient(domain, token)
		if err != nil {
			return nil, err
		}
	}

	workloadService := batch.NewService(batcher)
	return workloadService.Delete(txn.GetCtx(), ids...)
}
