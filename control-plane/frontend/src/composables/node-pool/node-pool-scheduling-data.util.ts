/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { Nodepool } from "@/swagger-models/cluster-service-client/index";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { metricUtil } from "@/utils/metric.util/index";
import { MetricsType, type MetricsResponse } from "@/swagger-models/cluster-service-client/index";
import type { INodePoolSummaryResources } from "@/models/node-pool.model";
import type { ISchedulingScope } from "@/models/node-pool-scheduling.model";
import { completePeerSchedulingTotals } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-totals.util";
export function schedulingFilters(scope: ISchedulingScope) {
  const filters = [
    filterUtil.getEqualsFilterString("nodePoolId", scope.poolId),
    filterUtil.getEqualsFilterString("clusterId", scope.clusterId),
  ];
  if (scope.rank) filters.push(filterUtil.getEqualsFilterString("rank", scope.rank));
  return filters;
}
// Count the entire pool, including projects and rows outside the current table filters/pages.
export async function loadNodePoolGpuLimits(scope: ISchedulingScope): Promise<boolean> {
  const filters = [
    ...schedulingFilters({ poolId: scope.poolId, clusterId: scope.clusterId }),
    filterUtil.getNotEqualsFilterString("gpuLimit", -1),
  ];
  const counts = await Promise.all([
    orgUnitService.countDepartmentNodePoolResources(filters),
    orgUnitService.countProjectsNodePoolResources(filters),
  ]);
  if (counts.some((count) => !Number.isFinite(count) || count < 0)) {
    throw new Error("Failed to check GPU allocation limits. Try again.");
  }
  return counts.some((count) => count > 0);
}
export const loadSchedulingPool = (poolId: string) => clusterService.getNodepool(poolId);
export async function loadDepartmentPage(
  params: ISchedulingScope & { search?: string; offset?: number; limit?: number },
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
export function loadSchedulingTotals(scope: ISchedulingScope) {
  // The sum endpoint takes pool scope as nodepoolId, and rejects nodePoolId filters.
  return orgUnitService.sumDepartmentNodePoolResources(scope.poolId, schedulingFilters(scope).slice(1));
}
// Sum APIs use the saved capability's weight semantics. Read all scoped pages to
// obtain both denominators while previewing an unsaved mode change.
export async function loadSchedulingDepartmentSnapshot(scope: ISchedulingScope, rawWeightsAvailable = true) {
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
    if (response.next != null && response.next <= offset)
      throw new Error("Department weight pagination did not advance.");
    offset = response.next;
  } while (offset != null);
  const totals = completePeerSchedulingTotals(
    rows.map((row) => ({
      id: row.orgUnitId,
      rank: row.rank,
      weight: row.resources.gpu?.overQuotaWeight ?? undefined,
      quota: row.resources.gpu?.deserved ?? undefined,
    })),
    { rawWeightsAvailable },
  );
  return { totals, rows };
}
export async function loadChildQuotaSum(scope: ISchedulingScope, departmentId: string): Promise<number> {
  const value = await orgUnitService.getDepartmentSubordinateGpuQuota(departmentId, {
    nodePoolId: scope.poolId,
    clusterId: scope.clusterId,
    emptyChildrenAsZero: true,
  });
  if (value == null || !Number.isFinite(value) || value < 0) throw new Error("Project quota totals are unavailable.");
  return value;
}
function telemetryResources(response: MetricsResponse): INodePoolSummaryResources {
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
