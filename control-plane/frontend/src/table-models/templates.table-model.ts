/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { CLUSTER_COLUMN_FILTER_NAME, EColumnFilterType, type IEnumBasedFilterSelectOption } from "@/models/filter.model";
import { autocompleteUtil } from "@/utils/autocomplete.util";
import { ECustomCell, USER_DISPLAY_NAME_COL_DEFAULTS, type ITableColumn } from "@/models/table.model";
import { EWorkloadNodeArchitecture } from "@/models/workload.model";
import { ScopeType, WorkloadType, type SupersetTemplate } from "@/swagger-models/workloads-client";
import { dateUtil } from "@/utils/date.util";

type TemplatesColumnsNames =
  | "name"
  | "description"
  | "scopeType"
  | "image"
  | "gpuComputeRequest"
  | "workloadType"
  | "workloadArchitecture"
  | "createdBy"
  | "createdAt"
  | "lastUpdated"
  | "cluster";

// Because Distributed is not treated as an actual type of workload in the UI we hide it as a filter option
const templateTypeFilterOptions: IEnumBasedFilterSelectOption[] = Object.keys(WorkloadType).map((key) => ({
  label: key,
  value: key,
  hidden: key === WorkloadType.Distributed || key === WorkloadType.DistributedInference,
}));

export const allTemplateColumnsMap: Record<TemplatesColumnsNames, ITableColumn> = {
  name: {
    name: "name",
    label: "Template",
    field: (row: SupersetTemplate) => row.meta.name,
    sortable: true,
    align: "left",
    display: true,
    customCell: ECustomCell.NAME_COL,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.templates("name", (t) => t.meta.name),
    },
  },
  description: {
    name: "description",
    label: "Description",
    field: (row: SupersetTemplate) => row.meta.description,
    format: (val: string | null | undefined) => val || "-",
    sortable: false,
    align: "left",
    display: false,
    customCell: ECustomCell.DESCRIPTION_COL,
    sanitizeCell: true,
    hideFilter: true,
  },
  scopeType: {
    name: "scopeType",
    label: "Scope",
    field: (row: SupersetTemplate) => row.meta.scopeType,
    sortable: false,
    align: "left",
    display: true,
    customCell: ECustomCell.SCOPE_COL,
    hideFilter: true,
  },
  image: {
    name: "image",
    label: "Image",
    field: (row: SupersetTemplate) => row?.spec?.image || row?.combinedSpec?.image || "-",
    sortable: false,
    align: "left",
    display: true,
    hideFilter: true,
  },
  gpuComputeRequest: {
    name: "gpuComputeRequest",
    label: "GPU compute request",
    field: (row: SupersetTemplate) => row,
    sortable: false,
    align: "left",
    display: true,
    hideFilter: true,
    format: (sst: SupersetTemplate) => {
      const workersGpuDevices = sst.spec?.compute?.gpuDevicesRequest || sst.combinedSpec?.compute?.gpuDevicesRequest;
      const workersGpuPortion = sst.spec?.compute?.gpuPortionRequest || sst.combinedSpec?.compute?.gpuPortionRequest;

      const workersResourcesWithoutWorkersCount = (workersGpuDevices || 0) * (workersGpuPortion || 1);

      // if its not distributed just return the workers resources
      if (sst.meta?.workloadType !== WorkloadType.Distributed) {
        return workersResourcesWithoutWorkersCount.toFixed(2);
      }

      const numOfWorkers = sst.spec?.numWorkers || sst.combinedSpec?.numWorkers || 0;

      const masterGpuDevices =
        sst.master?.spec?.compute?.gpuDevicesRequest || sst.combinedMasterSpec?.compute?.gpuDevicesRequest;
      const masterGpuPortion =
        sst.master?.spec?.compute?.gpuPortionRequest || sst.combinedMasterSpec?.compute?.gpuPortionRequest;

      const totalMasterResources = (masterGpuDevices || 0) * (masterGpuPortion || 1);

      const totalWorkerResources = workersResourcesWithoutWorkersCount * numOfWorkers;

      // if its sameAsWorker we don't get a masterSpec therefore we need to calculate the resources based on the workers resources
      if (sst.masterSpecSameAsWorker) {
        return (totalWorkerResources + workersResourcesWithoutWorkersCount).toFixed(2);
      }

      // calculate the total resources based on the workers and master resources
      return (totalWorkerResources + totalMasterResources).toFixed(2);
    },
  },
  workloadType: {
    name: "workloadType",
    label: "Workload type",
    field: (row: SupersetTemplate) => row.meta.workloadType,
    sortable: true,
    align: "left",
    display: true,
    format: (workloadType: WorkloadType) => {
      return workloadType === WorkloadType.Distributed ? WorkloadType.Training : workloadType;
    },
    filter: {
      type: EColumnFilterType.EnumBased,
      selectOptions: templateTypeFilterOptions,
    },
  },
  workloadArchitecture: {
    name: "workloadArchitecture",
    label: "Workload architecture",
    field: (row: SupersetTemplate) => row.spec?.distributedFramework || row.combinedSpec?.distributedFramework,
    sortable: false,
    align: "left",
    display: true,
    format: (distributedFramework: string | undefined) =>
      distributedFramework ? EWorkloadNodeArchitecture.Distributed : EWorkloadNodeArchitecture.Standard,
    hideFilter: true,
  },
  createdBy: {
    name: "createdBy",
    label: "Created by",
    field: (row: SupersetTemplate) => row.meta.createdBy,
    sortable: true,
    align: "left",
    display: true,
    ...USER_DISPLAY_NAME_COL_DEFAULTS,
    filter: {
      type: EColumnFilterType.FreeText,
      autocompleteOptionsFn: autocompleteUtil.templates("createdBy", (t) => t.meta.createdBy),
    },
  },
  createdAt: {
    name: "createdAt",
    label: "Creation time",
    field: (row: SupersetTemplate) => row.meta.createdAt,
    sortable: true,
    align: "left",
    format: (createdAt: string) => createdAt && dateUtil.dateAndTimeFormat(new Date(createdAt)),
    filter: {
      type: EColumnFilterType.Date,
    },
  },
  lastUpdated: {
    name: "updatedAt",
    label: "Last updated",
    field: (row: SupersetTemplate) => row.meta.updatedAt,
    hideFilter: true,
    sortable: true,
    align: "left",
    display: true,
    filter: {
      type: EColumnFilterType.Date,
    },
    format: (updatedAt: string) => updatedAt && dateUtil.dateAndTimeFormat(new Date(updatedAt)),
  },
  cluster: {
    name: CLUSTER_COLUMN_FILTER_NAME,
    label: "Cluster",
    field: (row: SupersetTemplate) => row.meta.clusterId,
    sortable: true,
    align: "left",
    customCell: ECustomCell.CLUSTER_ID_TO_NAME_COL,
    hideFilter: true,
    ignoreFilter: (row: SupersetTemplate) => {
      // ignoring this column filter if the scope is system or tenant (as clusterId is undefined)
      return !row.meta.clusterId && (row.meta.scopeType === ScopeType.System || row.meta.scopeType === ScopeType.Tenant);
    },
  },
};

export const allTemplateColumns: Array<ITableColumn> = [
  allTemplateColumnsMap.name,
  allTemplateColumnsMap.description,
  allTemplateColumnsMap.scopeType,
  allTemplateColumnsMap.image,
  allTemplateColumnsMap.gpuComputeRequest,
  allTemplateColumnsMap.workloadType,
  allTemplateColumnsMap.workloadArchitecture,
  allTemplateColumnsMap.createdBy,
  allTemplateColumnsMap.createdAt,
  allTemplateColumnsMap.lastUpdated,
  allTemplateColumnsMap.cluster,
];

export const templateIndexColumns: Array<ITableColumn> = [
  { ...allTemplateColumnsMap.name, display: true, mandatory: true },
  { ...allTemplateColumnsMap.description, display: false },
  { ...allTemplateColumnsMap.scopeType, display: true },
  { ...allTemplateColumnsMap.image, display: true },
  { ...allTemplateColumnsMap.gpuComputeRequest, display: true },
  { ...allTemplateColumnsMap.workloadType, display: true },
  { ...allTemplateColumnsMap.workloadArchitecture, display: false },
  { ...allTemplateColumnsMap.createdBy, display: true },
  { ...allTemplateColumnsMap.createdAt, display: true },
  { ...allTemplateColumnsMap.lastUpdated, display: false },
  { ...allTemplateColumnsMap.cluster, display: true },
];
