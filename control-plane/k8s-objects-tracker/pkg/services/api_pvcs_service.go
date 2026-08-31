/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
)

type PvcsService struct{}

func NewPVCService() *PvcsService {
	return &PvcsService{}
}

func (ps *PvcsService) CreateClusterWidePvcV1(ctx context.Context, clusterId string, pvcName string, _ api.CreateClusterWidePvcV1Request) (api.ImplResponse, error) {
	return ps.CreateClusterWidePvcV2(ctx, clusterId, pvcName, existingPvc)
}

func (ps *PvcsService) CreateClusterWidePvcV2(ctx context.Context, clusterId string, pvcName string, request api.PvcCreationRequest) (api.ImplResponse, error) {
	pvc, err := ps.CreatePvc(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, pvcName, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *pvc), nil
}

func (ps *PvcsService) CreateDepartmentWidePvcV1(ctx context.Context, clusterId string, departmentId string, pvcName string, _ api.CreateClusterWidePvcV1Request) (api.ImplResponse, error) {
	return ps.CreateDepartmentWidePvcV2(ctx, clusterId, departmentId, pvcName, existingPvc)
}

func (ps *PvcsService) CreateDepartmentWidePvcV2(ctx context.Context, clusterId string, departmentId string, pvcName string, request api.PvcCreationRequest) (api.ImplResponse, error) {
	return ps.CreateDepartmentsWidePvcV2(ctx, clusterId, departmentId, pvcName, request)
}

func (ps *PvcsService) CreateDepartmentsWidePvcV2(ctx context.Context, clusterId string, departmentId string, pvcName string, request api.PvcCreationRequest) (api.ImplResponse, error) {
	pvc, err := ps.CreatePvc(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, pvcName, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *pvc), nil
}

func (ps *PvcsService) CreateProjectPvcV1(ctx context.Context, clusterId string, namespace string, pvcName string, _ api.CreateClusterWidePvcV1Request) (api.ImplResponse, error) {
	return ps.CreateProjectPvcV2(ctx, clusterId, namespace, pvcName, existingPvc)
}

func (ps *PvcsService) CreateProjectPvcV2(ctx context.Context, clusterId string, namespace string, pvcName string, request api.PvcCreationRequest) (api.ImplResponse, error) {
	return ps.CreateProjectsPvcV2(ctx, clusterId, namespace, pvcName, request)
}

func (ps *PvcsService) CreateProjectsPvcV2(ctx context.Context, clusterId string, namespace string, pvcName string, request api.PvcCreationRequest) (api.ImplResponse, error) {
	pvc, err := ps.CreatePvc(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, pvcName, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *pvc), nil
}

func (ps *PvcsService) SyncClusterWidePvcNamesV1(ctx context.Context, clusterId string, pvcNames api.PvcNames) (api.ImplResponse, error) {
	return ps.SyncClusterWidePvcsV2(ctx, clusterId, existingPvcsFromNames(pvcNames))
}

func (ps *PvcsService) SyncClusterWidePvcsV2(ctx context.Context, clusterId string, request api.Pvcs) (api.ImplResponse, error) {
	pvcs, err := ps.SyncPvcs(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *pvcs), nil
}

func (ps *PvcsService) SyncDepartmentWidePvcNamesV1(ctx context.Context, clusterId string, departmentId string, pvcNames api.PvcNames) (api.ImplResponse, error) {
	return ps.SyncDepartmentWidePvcsV2(ctx, clusterId, departmentId, existingPvcsFromNames(pvcNames))
}

func (ps *PvcsService) SyncDepartmentWidePvcsV2(ctx context.Context, clusterId string, departmentId string, request api.Pvcs) (api.ImplResponse, error) {
	return ps.SyncDepartmentsWidePvcsV2(ctx, clusterId, departmentId, request)
}

func (ps *PvcsService) SyncDepartmentsWidePvcsV2(ctx context.Context, clusterId string, departmentId string, request api.Pvcs) (api.ImplResponse, error) {
	pvcs, err := ps.SyncPvcs(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *pvcs), nil
}

func (ps *PvcsService) SyncProjectPvcNamesV1(ctx context.Context, clusterId string, projectId string, pvcNames api.PvcNames) (api.ImplResponse, error) {
	return ps.SyncProjectPvcsV2(ctx, clusterId, projectId, existingPvcsFromNames(pvcNames))
}

func (ps *PvcsService) SyncProjectPvcsV2(ctx context.Context, clusterId string, namespace string, request api.Pvcs) (api.ImplResponse, error) {
	return ps.SyncProjectsPvcsV2(ctx, clusterId, namespace, request)
}

func (ps *PvcsService) SyncProjectsPvcsV2(ctx context.Context, clusterId string, namespace string, request api.Pvcs) (api.ImplResponse, error) {
	pvcs, err := ps.SyncPvcs(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, request)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, *pvcs), nil
}

func (ps *PvcsService) DeleteClusterWidePvcV1(ctx context.Context, clusterId string, pvcName string) (api.ImplResponse, error) {
	return ps.DeleteClusterWidePvcV2(ctx, clusterId, pvcName)
}

func (ps *PvcsService) DeleteClusterWidePvcV2(ctx context.Context, clusterId string, pvcName string) (api.ImplResponse, error) {
	err := ps.DeletePvc(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, pvcName)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (ps *PvcsService) DeleteDepartmentWidePvcV1(ctx context.Context, clusterId string, departmentId string, pvcName string) (api.ImplResponse, error) {
	return ps.DeleteDepartmentWidePvcV2(ctx, clusterId, departmentId, pvcName)
}

func (ps *PvcsService) DeleteDepartmentWidePvcV2(ctx context.Context, clusterId string, departmentId string, pvcName string) (api.ImplResponse, error) {
	return ps.DeleteDepartmentsWidePvcV2(ctx, clusterId, departmentId, pvcName)
}

func (ps *PvcsService) DeleteDepartmentsWidePvcV2(ctx context.Context, clusterId string, departmentId string, pvcName string) (api.ImplResponse, error) {
	err := ps.DeletePvc(ctx, clusterId, authorizationClient.SCOPETYPE_DEPARTMENT, departmentId, pvcName)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (ps *PvcsService) DeleteProjectPvcV1(ctx context.Context, clusterId string, namespace string, pvcName string) (api.ImplResponse, error) {
	return ps.DeleteProjectPvcV2(ctx, clusterId, namespace, pvcName)
}

func (ps *PvcsService) DeleteProjectPvcV2(ctx context.Context, clusterId string, namespace string, pvcName string) (api.ImplResponse, error) {
	return ps.DeleteProjectsPvcV2(ctx, clusterId, namespace, pvcName)
}

func (ps *PvcsService) DeleteProjectsPvcV2(ctx context.Context, clusterId string, namespace string, pvcName string) (api.ImplResponse, error) {
	err := ps.DeletePvc(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, pvcName)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return NoContentResponse(), nil
}

func (ps *PvcsService) ObsoleteCreatePvc(ctx context.Context, clusterId string, namespace string, request api.SinglePvc) (api.ImplResponse, error) {
	var err *api.Error

	if namespace == NamespaceRunai {
		_, err = ps.CreatePvc(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, request.PvcName, existingPvc)
	} else {
		_, err = ps.CreatePvc(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, request.PvcName, existingPvc)
	}
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, request), nil
}

func (ps *PvcsService) ObsoleteDeletePvc(ctx context.Context, clusterId string, namespace string, pvcName string) (api.ImplResponse, error) {
	var err *api.Error

	if namespace == NamespaceRunai {
		err = ps.DeletePvc(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, pvcName)
	} else {
		err = ps.DeletePvc(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, pvcName)
	}
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, api.SinglePvc{
		PvcName: pvcName,
	}), nil
}

func (ps *PvcsService) ObsoleteSyncPvc(ctx context.Context, clusterId string, namespace string, pvcs []api.SinglePvc) (api.ImplResponse, error) {
	newRequest := api.Pvcs{
		Pvcs: make([]api.Pvc, 0, len(pvcs)),
	}

	for _, pvc := range pvcs {
		newRequest.Pvcs = append(newRequest.Pvcs, api.Pvc{
			Name:   pvc.PvcName,
			Origin: &originExisting,
		})
	}

	var err *api.Error

	if namespace == NamespaceRunai {
		_, err = ps.SyncPvcs(ctx, clusterId, authorizationClient.SCOPETYPE_CLUSTER, clusterId, newRequest)
	} else {
		_, err = ps.SyncPvcs(ctx, clusterId, authorizationClient.SCOPETYPE_PROJECT, namespace, newRequest)
	}
	if err != nil {
		return NewErrorResponse(err), nil
	}

	return api.Response(http.StatusOK, pvcs), nil
}

var (
	originExisting = api.ORIGIN_EXISTING
	originAssets   = api.ORIGIN_ASSETS //nolint:unused
)

var existingPvc = api.PvcCreationRequest{
	Origin: &originExisting,
}

func existingPvcsFromNames(pvcNames api.PvcNames) api.Pvcs {
	pvcs := api.Pvcs{
		Pvcs: make([]api.Pvc, 0, len(pvcNames.Pvcs)),
	}

	for _, pvcName := range pvcNames.Pvcs {
		pvcs.Pvcs = append(pvcs.Pvcs, api.Pvc{
			Name:   pvcName,
			Origin: &originExisting,
		})
	}

	return pvcs
}
