/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { metricUtil } from "@/utils/metric.util";
import { MetricsType, type MetricsResponse } from "@/swagger-models/cluster-service-client";
import type { INodePoolSummaryResources } from "../node-pool-wizard-summary/node-pool-wizard-summary.model";
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
  return orgUnitService.sumDepartmentNodePoolResources(scope.poolId, schedulingFilters(scope));
}
export async function loadChildQuotaSum(scope: SchedulingScope, departmentId: string): Promise<number> {
  const filters = [...schedulingFilters(scope), filterUtil.getEqualsFilterString("parentId", departmentId)];
  let offset: number | undefined = 0;
  let total = 0;
  do {
    const response = await orgUnitService.getProjectNodePoolResources(filters, "orgUnitName", "asc", offset, 100);
    for (const row of response.nodePoolResources) {
      const value = row.resources.gpu?.deserved;
      if (value == null || !Number.isFinite(value) || value < 0)
        throw new Error("Project quota totals are unavailable.");
      total += value;
    }
    offset = response.next;
  } while (offset != null);
  return total;
}
export function telemetryResources(response: MetricsResponse): INodePoolSummaryResources {
  const values = metricUtil.mapMeasurementsToMap(response)[MetricsType.TotalGpu];
  const raw = values?.at(-1)?.value;
  const value = raw == null || raw === "" ? undefined : Number(raw);
  return { gpus: value != null && Number.isFinite(value) ? value : undefined };
}
export async function loadSchedulingResources(pool: { clusterId: string; name: string }) {
  const end = new Date();
  return telemetryResources(
    await clusterService.getNodepoolMetrics(
      pool.clusterId,
      pool.name,
      new Date(end.getTime() - 60000).toISOString(),
      end.toISOString(),
      [MetricsType.TotalGpu],
      1,
    ),
  );
}
