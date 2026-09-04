/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/nim/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/nim/service"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

var wsClient workload.WithUpdateAction

func SetClient(wsClientSet workload.WithUpdateAction) {
	wsClient = wsClientSet
}

func GetByNameAndProj(txn *transaction.Txn, domain, token, name, project, cluster string, fb *filter.Builder) (*list.Output, error) {
	fb.AppendName(name).WithProject(project).WithClusterID(cluster)
	return GetByFilter(txn, domain, token, fb)
}

func GetByFilter(txn *transaction.Txn, domain, token string, fb *filter.Builder) (*list.Output, error) {
	fb.WithOnlyWorkloadType(api.BackendNimServiceWorkloadType)
	out, err := workload.GetWorkloadByFilter(txn, domain, token, fb)
	if err != nil || out == nil {
		return out, err
	}

	// Callers seed the filter with their own workload type, and NIM commands are
	// registered as Inference. Replace that type instead of producing the filter
	// "Inference||NIMService", which can match a standard Inference workload.
	// Also verify the response before handing its id to the NIM API.
	for _, wl := range out.GetList() {
		if wl.Type != api.BackendNimServiceWorkloadType {
			return nil, fmt.Errorf("workload %q is of type %s, not a %s workload",
				wl.Name, wl.Type, api.BackendNimServiceWorkloadType)
		}
	}

	return out, nil
}

func List(txn *transaction.Txn, domain, token string, pagingOptions *apilist.PageOptions, fb *filter.Builder, deleted bool) (*list.Output, error) {
	fb.WithWorkloadType(api.BackendNimServiceWorkloadType)
	return workload.GetWorkloadPagination(txn, domain, token, pagingOptions, fb, deleted)
}

func Delete(txn *transaction.Txn, domain string, token string, uuid string) error {
	nimClient, err := getNimClient(domain, token)
	if err != nil {
		return err
	}
	return nimClient.Delete(txn.GetCtx(), uuid)
}

func Submit(txn *transaction.Txn, domain, token, name, projectId, clusterId string, generalSpec *spec.GeneralSpec) (string, error) {
	nimClient, err := getNimClient(domain, token)
	if err != nil {
		return "", err
	}
	return nimClient.Submit(txn.GetCtx(), name, projectId, clusterId, generalSpec)
}

func Update(txn *transaction.Txn, domain, token, uuid, clusterId string, spec *spec.GeneralSpec) (string, error) {
	nimService, err := getNimService(domain, token)
	if err != nil {
		return "", err
	}
	return nimService.Update(txn.GetCtx(), uuid, clusterId, spec)
}

func getNimClient(domain string, token string) (workload.WithUpdateAction, error) {
	if wsClient != nil {
		return wsClient, nil
	}
	return client.NewNimClient(domain, token)
}

func getNimService(domain string, token string) (*service.BackendNimService, error) {
	var err error
	if wsClient == nil {
		wsClient, err = client.NewNimClient(domain, token)
		if err != nil {
			return nil, err
		}
		return service.NewBackendNimService(wsClient), nil
	}
	return service.NewBackendNimService(wsClient), nil
}

type AutoscalingInfo struct {
	Enabled     bool
	MinReplicas int32
	MaxReplicas int32
}

func GetAutoscalingInfo(txn *transaction.Txn, domain, token, uuid string) (*AutoscalingInfo, error) {
	nimClient, err := getNimClient(domain, token)
	if err != nil {
		return nil, err
	}

	backendClient, ok := nimClient.(*client.BackendNimClient)
	if !ok {
		return nil, fmt.Errorf("unexpected client type: expected *client.BackendNimClient, got %T", nimClient)
	}

	nimService, err := backendClient.Get(txn.GetCtx(), uuid)
	if err != nil {
		return nil, err
	}

	if nimService == nil {
		return nil, nil
	}

	if !nimService.Spec.HasAutoscaling() {
		return &AutoscalingInfo{Enabled: false}, nil
	}

	autoscaling := nimService.Spec.GetAutoscaling()
	info := &AutoscalingInfo{
		Enabled:     true,
		MinReplicas: autoscaling.GetMinReplicas(),
		MaxReplicas: autoscaling.GetMaxReplicas(),
	}

	return info, nil
}
