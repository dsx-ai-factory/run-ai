/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package training

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/training/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/training/service"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
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
	fb.WithWorkloadType(api.Training.QueryParam())
	return workload.GetWorkloadByFilter(txn, domain, token, fb)
}

func List(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
	fb.WithWorkloadType(api.Training.QueryParam())
	return workload.GetWorkloadPagination(txn, domain, token, pagingOptions, fb, deleted)
}

func Delete(txn *transaction.Txn, domain string, token string, uuid string) error {
	trainingService, err := getTrainingService(domain, token)
	if err != nil {
		return err
	}
	return trainingService.DeleteByUuid(txn.GetCtx(), uuid)
}

func Resume(txn *transaction.Txn, domain string, token string, uuid string) error {
	trainingService, err := getTrainingService(domain, token)
	if err != nil {
		return err
	}
	return trainingService.ResumeByUuid(txn.GetCtx(), uuid)
}

func Suspend(txn *transaction.Txn, domain string, token string, uuid string) error {
	trainingService, err := getTrainingService(domain, token)
	if err != nil {
		return err
	}
	return trainingService.SuspendByUuid(txn.GetCtx(), uuid)
}

func Submit(txn *transaction.Txn, domain, token, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	trainingService, err := getTrainingService(domain, token)
	if err != nil {
		return "", err
	}
	return trainingService.Submit(txn.GetCtx(), name, projectId, clusterId, spec)
}

func getTrainingService(domain string, token string) (*service.BackendTrainingService, error) {
	var err error
	if wsClient == nil {
		wsClient, err = client.NewTrainingClient(domain, token)
		if err != nil {
			return nil, err
		}
		return service.NewBackendTrainingService(wsClient), nil
	}
	return service.NewBackendTrainingService(wsClient), nil
}
