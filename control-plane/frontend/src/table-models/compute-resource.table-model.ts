/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util";
import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import { Scope, type ComputeAsset } from "@/swagger-models/assets-service-client";
import { dateUtil } from "@/utils/date.util";
import { getMemoryUnitsForUI, getCpuCoresForUI } from "@/utils/table-format.util";

export const allComputeResourceColumnsMap: Record<string, ITableColumn> = {
  computeResource: {
    name: "compute-resource",
    label: "Compute resource",
    field: (row) => row.meta.name,
    sortable: true,
    align: "left",
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.computeResources((cr) => cr.meta.name),
    },
  },
  description: {
    name: "description",
    label: "Description",
    field: (row: ComputeAsset) => row.meta.description,
    format: (val: string | null | undefined) => val || "-",
    sortable: true,
    align: "left",
    display: false,
    customCell: ECustomCell.DESCRIPTION_COL,
    sanitizeCell: true,
  },
  gpuDevicesRequest: {
    name: "gpu-devices-request",
    label: "GPU devices request per pod",
    field: (row: ComputeAsset) => row.spec.gpuDevicesRequest,
    sortable: true,
    align: "left",
    format: (gpuDevicesRequest?: number): string => {
      return typeof gpuDevicesRequest === "number" ? String(gpuDevicesRequest) : "-";
    },
  },
  gpuMemoryRequest: {
    name: "gpu-memory-request",
    label: "GPU memory request per device",
    field: (row: ComputeAsset) => row,
    sortable: true,
    align: "left",
    format: (row: ComputeAsset): string => {
      const { gpuPortionRequest, gpuMemoryRequest } = row.spec;
      return (
        (gpuMemoryRequest && getMemoryUnitsForUI(gpuMemoryRequest)) ||
        (gpuPortionRequest && `${gpuPortionRequest * 100}%`) ||
        "-"
      );
    },
  },
  gpuMemoryLimit: {
    name: "gpu-memory-limit",
    label: "GPU memory limit per device",
    field: (row: ComputeAsset) => row,
    sortable: true,
    align: "left",
    format: (row: ComputeAsset): string => {
      const { gpuPortionLimit, gpuMemoryLimit } = row.spec;
      return (
        (gpuMemoryLimit && getMemoryUnitsForUI(gpuMemoryLimit)) ||
        (gpuPortionLimit && `${gpuPortionLimit * 100}%`) ||
        "-"
      );
    },
  },
  cpuCoresRequest: {
    name: "cpu-cores-request",
    label: "CPU compute request",
    field: (row: ComputeAsset) => row.spec.cpuCoreRequest,
    sortable: true,
    align: "left",
    format: (cpuCoreRequest?: number): string => {
      if (typeof cpuCoreRequest !== "number") return "-";
      return getCpuCoresForUI(cpuCoreRequest);
    },
  },
  cpuCoresLimit: {
    name: "cpu-cores-limit",
    label: "CPU compute limit",
    field: (row: ComputeAsset) => row.spec.cpuCoreLimit,
    sortable: true,
    align: "left",
    format: (cpuCoreLimit?: number): string => {
      if (typeof cpuCoreLimit !== "number") return "-";
      return getCpuCoresForUI(cpuCoreLimit);
    },
  },
  cpuMemoryRequest: {
    name: "cpu-memory-request",
    label: "CPU memory request",
    field: (row: ComputeAsset) => getMemoryUnitsForUI(row.spec.cpuMemoryRequest || ""),
    format: (val: string) => val || "-",
    sortable: true,
    align: "left",
    sort: (a: string, b: string) => {
      let aMemory = Number(a?.split(" ")[0] || "");
      let bMemory = Number(b?.split(" ")[0] || "");
      const aUnit = a?.split(" ")[1] || "";
      const bUnit = b?.split(" ")[1] || "";
      if (aUnit === "GB") {
        aMemory *= 1000;
      }
      if (bUnit === "GB") {
        bMemory *= 1000;
      }
      return aMemory - bMemory;
    },
  },
  cpuMemoryLimit: {
    name: "cpu-memory-limit",
    label: "CPU memory limit",
    field: (row: ComputeAsset) => getMemoryUnitsForUI(row.spec.cpuMemoryLimit || ""),
    format: (val: string) => val || "-",
    sortable: true,
    align: "left",
    sort: (a: string, b: string) => {
      let aMemory = Number(a?.split(" ")[0] || "");
      let bMemory = Number(b?.split(" ")[0] || "");
      const aUnit = a?.split(" ")[1] || "";
      const bUnit = b?.split(" ")[1] || "";
      if (aUnit === "GB") {
        aMemory *= 1000;
      }
      if (bUnit === "GB") {
        bMemory *= 1000;
      }
      return aMemory - bMemory;
    },
  },
  scope: {
    name: "scope",
    label: "Scope",
    field: (row: ComputeAsset) => row.meta.scope,
    sortable: true,
    align: "left",
    format: (scope: Scope, computeResource: ComputeAsset): string | number => {
      switch (scope) {
        case Scope.Project:
          return Number(computeResource.meta.projectId);
        case Scope.Department:
          return String(computeResource.meta.departmentId);
        case Scope.Cluster:
          return String(computeResource.meta.clusterId);
        default:
          return String(computeResource.meta.tenantId);
      }
    },
    customCell: ECustomCell.SCOPE_COL,
    hideFilter: true,
  },
  createdBy: {
    name: "created-by",
    label: "Created by",
    field: (row: ComputeAsset) => row.meta.createdBy,
    sortable: true,
    align: "left",
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.computeResources((cr) => cr.meta.createdBy),
    },
  },
  createdAt: {
    name: "created-at",
    label: "Creation time",
    field: (row: ComputeAsset) => row.meta.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => dateUtil.dateAndTimeFormat(new Date(createdAt)),
  },
  lastUpdated: {
    name: "last-updated",
    label: "Last updated",
    field: (row: ComputeAsset) => row.meta.updatedAt,
    sortable: true,
    align: "left",
    display: false,
    format: (updatedAt: string) => dateUtil.dateAndTimeFormat(new Date(updatedAt)),
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row: ComputeAsset) => row.meta.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
    ignoreFilter: (row: ComputeAsset) => {
      // ignoring this column filter if the scope is system or tenant (as clusterId is undefined)
      return !row.meta.clusterId && (row.meta.scope === Scope.System || row.meta.scope === Scope.Tenant);
    },
  },
};

export const allComputeResourceColumns: Array<ITableColumn> = [
  allComputeResourceColumnsMap.computeResource,
  allComputeResourceColumnsMap.description,
  allComputeResourceColumnsMap.gpuDevicesRequest,
  allComputeResourceColumnsMap.gpuMemoryRequest,
  allComputeResourceColumnsMap.gpuMemoryLimit,
  allComputeResourceColumnsMap.cpuMemoryRequest,
  allComputeResourceColumnsMap.cpuMemoryLimit,
  allComputeResourceColumnsMap.cpuCoresRequest,
  allComputeResourceColumnsMap.cpuCoresLimit,
  allComputeResourceColumnsMap.scope,
  allComputeResourceColumnsMap.createdBy,
  allComputeResourceColumnsMap.createdAt,
  allComputeResourceColumnsMap.lastUpdated,
  allComputeResourceColumnsMap.cluster,
];

export const computeResourceIndexColumns: Array<ITableColumn> = [
  { ...allComputeResourceColumnsMap.computeResource, display: true, mandatory: true },
  { ...allComputeResourceColumnsMap.description, display: false },
  { ...allComputeResourceColumnsMap.scope, display: true },
  { ...allComputeResourceColumnsMap.gpuDevicesRequest, display: true },
  { ...allComputeResourceColumnsMap.gpuMemoryRequest, display: true },
  { ...allComputeResourceColumnsMap.gpuMemoryLimit, display: true },
  { ...allComputeResourceColumnsMap.cpuMemoryRequest, display: true },
  { ...allComputeResourceColumnsMap.cpuMemoryLimit, display: true },
  { ...allComputeResourceColumnsMap.cpuCoresRequest, display: true },
  { ...allComputeResourceColumnsMap.cpuCoresLimit, display: true },
  { ...allComputeResourceColumnsMap.createdBy, display: true },
  { ...allComputeResourceColumnsMap.createdAt, display: true },
  { ...allComputeResourceColumnsMap.lastUpdated, display: false },
  { ...allComputeResourceColumnsMap.cluster, display: true },
];
