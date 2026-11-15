/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { clusterServiceApi } from "@/services/infra/client-apis/cluster-service-api/cluster-service-api";
import { httpService } from "@/services/infra/https.service/http.service";
import type { Nodepool, NodepoolUpdateFields } from "@/swagger-models/cluster-service-client";
import type { BatchPatchNodePoolResourcesRequest } from "@/swagger-models/org-unit-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { metricUtil } from "@/utils/metric.util";
import { MetricsType, type MetricsResponse } from "@/swagger-models/cluster-service-client";
import type { INodePoolSummaryResources } from "@/models/node-pool.model";
export interface SchedulingScope {
  poolId: string;
  clusterId: string;
  rank?: string;
}
export function schedulingFilters(scope: SchedulingScope) {
  const filters = [
    filterUtil.getEqualsFilterString("nodePoolId", scope.poolId),
    filterUtil.getEqualsFilterString("clusterId", scope.clusterId),
  ];
  if (scope.rank) filters.push(filterUtil.getEqualsFilterString("rank", scope.rank));
  return filters;
}
export const loadSchedulingPool = (poolId: string) => clusterService.getNodepool(poolId);
export async function loadDepartmentPage(
  params: SchedulingScope & { search?: string; offset?: number; limit?: number },
) {
  const response = await orgUnitService.getDepartmentNodePoolResources(
    schedulingFilters(params),
    "orgUnitName",
    "asc",
    params.offset ?? 0,
    params.limit ?? 50,
    params.search,
  );
  return response.nodePoolResources;
}
export function loadSchedulingTotals(scope: SchedulingScope) {
  // The sum endpoint takes pool scope as nodepoolId, and rejects nodePoolId filters.
  return orgUnitService.sumDepartmentNodePoolResources(scope.poolId, schedulingFilters(scope).slice(1));
}
// Sum APIs use the saved capability's weight semantics. Read all scoped pages to
// obtain both denominators while previewing an unsaved mode change.
export async function loadSchedulingDepartmentSnapshot(scope: SchedulingScope, rawWeightsAvailable = true) {
  const totals: Record<
    string,
    {
      rawWeightsSum?: number;
      automaticWeightsSum: number;
      finiteQuotaSum: number;
      unlimitedQuotaCount: number;
      deservedSum: number;
    }
  > = {
    all: { rawWeightsSum: 0, automaticWeightsSum: 0, finiteQuotaSum: 0, unlimitedQuotaCount: 0, deservedSum: 0 },
  };
  const rows: Awaited<ReturnType<typeof loadDepartmentPage>> = [];
  let offset: number | undefined = 0;
  do {
    const response = await orgUnitService.getDepartmentNodePoolResources(
      schedulingFilters(scope),
      "orgUnitName",
      "asc",
      offset,
      100,
    );
    rows.push(...response.nodePoolResources);
    for (const row of response.nodePoolResources) {
      for (const key of ["all", row.rank]) {
        const total =
          totals[key] ??
          (totals[key] = {
            rawWeightsSum: 0,
            automaticWeightsSum: 0,
            finiteQuotaSum: 0,
            unlimitedQuotaCount: 0,
            deservedSum: 0,
          });
        const raw = row.resources.gpu?.overQuotaWeight;
        if (raw == null || !Number.isFinite(raw) || raw < 0) total.rawWeightsSum = undefined;
        else if (total.rawWeightsSum !== undefined) total.rawWeightsSum += raw;
        const quota = row.resources.gpu?.deserved;
        // Matches backend SumDepartmentNodePoolResources: unlimited/null have effective weight 0.
        if (quota != null && Number.isFinite(quota) && quota >= 0) {
          total.automaticWeightsSum += quota;
          total.finiteQuotaSum += quota;
        } else total.unlimitedQuotaCount++;
      }
    }
    if (response.next != null && response.next <= offset)
      throw new Error("Department weight pagination did not advance.");
    offset = response.next;
  } while (offset != null);
  if (!rawWeightsAvailable) for (const total of Object.values(totals)) total.rawWeightsSum = undefined;
  for (const total of Object.values(totals)) total.deservedSum = total.unlimitedQuotaCount ? -1 : total.finiteQuotaSum;
  return { totals, rows };
}
export async function loadSchedulingWeightTotals(scope: SchedulingScope, rawWeightsAvailable = true) {
  const snapshot = await loadSchedulingDepartmentSnapshot(scope, rawWeightsAvailable);
  return Object.fromEntries(
    Object.entries(snapshot.totals).map(([key, { rawWeightsSum, automaticWeightsSum }]) => [
      key,
      { rawWeightsSum, automaticWeightsSum },
    ]),
  );
}
export async function loadChildQuotaSum(scope: SchedulingScope, departmentId: string): Promise<number> {
  const department = await orgUnitService.getDepartment(departmentId);
  if (department.clusterId !== scope.clusterId) throw new Error("Department does not belong to this cluster.");
  const value = department.projectsAggregatedResources?.find((resource) => resource.nodePool.id === scope.poolId)?.gpu
    ?.deserved;
  // The verbose department response omits aggregates when its complete child list is empty.
  if (value == null && Array.isArray(department.children) && department.children.length === 0) return 0;
  if (value == null || !Number.isFinite(value) || value < 0) throw new Error("Project quota totals are unavailable.");
  return value;
}
export function telemetryResources(response: MetricsResponse): INodePoolSummaryResources {
  const values = metricUtil.mapMeasurementsToMap(response)[MetricsType.TotalGpu];
  const raw = values?.at(-1)?.value;
  const value = raw == null || raw === "" ? undefined : Number(raw);
  return { gpus: value != null && Number.isFinite(value) ? value : undefined };
}
export async function loadSchedulingResources(
  pool: Pick<Nodepool, "clusterId" | "name" | "phase" | "status" | "gpuTypes">,
): Promise<INodePoolSummaryResources> {
  // Empty is an authoritative pool state, even before metrics report zeros.
  if (pool.phase === "Empty") return { gpus: 0, nodes: 0 };
  const inventory: INodePoolSummaryResources = { nodes: pool.status?.nodes?.length };
  const types = pool.gpuTypes;
  // Inventory is reported by the cluster before a newly created pool has metric samples.
  if (types?.length && types.every((type) => Number.isFinite(type.totalDevices) && type.totalDevices >= 0)) {
    inventory.gpus = types.reduce((sum, type) => sum + type.totalDevices, 0);
    inventory.gpuTypes = [...new Set(types.map((type) => type.type))];
    if (
      types.every(
        (type) => type.gpuMemoryBytes != null && Number.isFinite(type.gpuMemoryBytes) && type.gpuMemoryBytes >= 0,
      )
    )
      inventory.gpuMemoryBytes = types.reduce((sum, type) => sum + type.totalDevices * (type.gpuMemoryBytes ?? 0), 0);
    return inventory;
  }
  const end = new Date();
  return {
    ...inventory,
    ...telemetryResources(
      await clusterService.getNodepoolMetrics(
        pool.clusterId,
        pool.name,
        new Date(end.getTime() - 60000).toISOString(),
        end.toISOString(),
        [MetricsType.TotalGpu],
        1,
      ),
    ),
  };
}

// PATCH only the scheduling fields owned by this step; PUT would reapply defaults.
export async function patchSchedulingPool(id: string, payload: NodepoolUpdateFields) {
  try {
    return (await clusterServiceApi.nodePoolsApi.patchNodepool(id, payload)).data;
  } catch (error) {
    throw httpService.handleHttpError(error);
  }
}
export const patchSchedulingDepartments = (request: BatchPatchNodePoolResourcesRequest) =>
  orgUnitService.batchPatchDepartmentNodePoolResources(request);
