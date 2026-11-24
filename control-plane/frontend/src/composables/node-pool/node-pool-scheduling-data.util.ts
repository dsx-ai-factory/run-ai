/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import type { Nodepool } from "@/swagger-models/cluster-service-client/index";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { metricUtil } from "@/utils/metric.util/index";
import { MetricsType, type MetricsResponse } from "@/swagger-models/cluster-service-client/index";
import type { INodePoolSummaryResources } from "@/models/node-pool.model";
import { DEPARTMENT_PAGE_SIZE, type ISchedulingScope } from "@/models/node-pool-scheduling.model";
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
export function loadSchedulingTotals(scope: ISchedulingScope) {
  // The sum endpoint takes pool scope as nodepoolId, and rejects nodePoolId filters.
  return orgUnitService.sumDepartmentNodePoolResources(scope.poolId, schedulingFilters(scope).slice(1));
}

export async function loadSchedulingDepartmentPage(scope: ISchedulingScope, offset: number) {
  const page = await orgUnitService.getDepartmentNodePoolResources(
    schedulingFilters(scope),
    "orgUnitName",
    "asc",
    offset,
    DEPARTMENT_PAGE_SIZE,
  );
  if (page.next != null && page.next <= offset) throw new Error("Department pagination did not advance.");
  return page;
}

export function schedulingSnapshotTotals(rows: DepartmentNodePoolResources[], rawWeightsAvailable: boolean) {
  return completePeerSchedulingTotals(
    rows.map((row) => ({
      id: row.orgUnitId,
      rank: row.rank,
      weight: row.resources.gpu?.overQuotaWeight ?? undefined,
      quota: row.resources.gpu?.deserved ?? undefined,
    })),
    { rawWeightsAvailable },
  );
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
  const nodes = pool.status?.nodes?.length;
  if (pool.phase === "Empty" || nodes === 0) return { gpus: 0, nodes: 0 };
  const inventory: INodePoolSummaryResources = { nodes };
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
