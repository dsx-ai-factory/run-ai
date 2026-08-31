/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed/service"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var wsClient workload.WithSuspendableAction

func SetClient(wsClientSet workload.WithSuspendableAction) {
	wsClient = wsClientSet
}

func GetByNameAndProj(txn *transaction.Txn, domain, token, name, project, cluster string, fb *filter.Builder) (*list.Output, error) {
	fb.AppendName(name).WithProject(project).WithClusterID(cluster)
	return GetByFilter(txn, domain, token, fb)
}

func GetByFilter(txn *transaction.Txn, domain, token string, fb *filter.Builder) (*list.Output, error) {
	fb.WithWorkloadType(api.Distributed.QueryParam())
	return workload.GetWorkloadByFilter(txn, domain, token, fb)
}

func List(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
	fb.WithWorkloadType(api.Distributed.QueryParam())
	return workload.GetWorkloadPagination(txn, domain, token, pagingOptions, fb, deleted)
}

func Delete(txn *transaction.Txn, domain string, token string, uuid string) error {
	distributedService, err := getDistributedService(domain, token)
	if err != nil {
		return err
	}
	return distributedService.DeleteByUuid(txn.GetCtx(), uuid)
}

func Resume(txn *transaction.Txn, domain string, token string, uuid string) error {
	distributedService, err := getDistributedService(domain, token)
	if err != nil {
		return err
	}
	return distributedService.ResumeByUuid(txn.GetCtx(), uuid)
}

func Suspend(txn *transaction.Txn, domain string, token string, uuid string) error {
	distributedService, err := getDistributedService(domain, token)
	if err != nil {
		return err
	}
	return distributedService.SuspendByUuid(txn.GetCtx(), uuid)
}

func Submit(txn *transaction.Txn, domain, token, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	distributedService, err := getDistributedService(domain, token)
	if err != nil {
		return "", err
	}
	return distributedService.Submit(txn.GetCtx(), name, projectId, clusterId, spec)
}

func getDistributedService(domain string, token string) (*service.BackendDistributedService, error) {
	var err error
	if wsClient == nil {
		wsClient, err = client.NewDistributedClient(domain, token)
		if err != nil {
			return nil, err
		}
		return service.NewBackendDistributedService(wsClient), nil
	}
	return service.NewBackendDistributedService(wsClient), nil
}
