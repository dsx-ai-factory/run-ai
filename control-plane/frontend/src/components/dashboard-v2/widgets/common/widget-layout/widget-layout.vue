<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="layout-wrapper">
    <template v-for="(group, index) in widgetsLayout" :key="index">
      <div class="row wrap">
        <div v-for="widget in group" :key="getWidgetKey(widget)" :class="['widget-pa-xs', getColClass(group.length)]">
          <component
            :aid="getWidgetAid(widget)"
            :is="widget"
            :cluster-id="clusterId"
            :node-pool-name="nodePoolName"
            :filter-by-dates="filterByDates"
            :has-departments="hasDepartments"
            :has-read-projects-permission="hasReadProjectsPermission"
            :has-read-departments-permission="hasReadDepartmentsPermission"
            :has-read-clusters-permission="hasReadClustersPermission"
            :org-unit-metric-source="orgUnitMetricSource"
            @date-changed="$emit('date-changed', $event)"
          />
        </div>
      </div>
    </template>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType, type DefineComponent } from "vue";
import type { EWidgetComponent } from "@/models/chart.model";
import type { IRangeDates } from "@/utils/date.util";
import type { IOrgUnitMetricSource } from "@/models/org-unit.model";
//stat widgets
import { RunningWorkloadsWidget } from "@/components/dashboard-v2/widgets/stat-widgets/running-workloads-widget";
import { PendingWorkloadsWidget } from "@/components/dashboard-v2/widgets/stat-widgets/pending-workloads-widget";
import { ReadyNodesWidget } from "@/components/dashboard-v2/widgets/stat-widgets/ready-nodes-widget/";
import { AllocatedGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/allocated-gpu-widget/";
import { ReadyGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/ready-gpu-widget/";
//chart widgets
import { WorkloadStatusWidget } from "@/components/dashboard-v2/widgets/chart-widgets/workload-status-widget";
import { WorkloadPendingTimeDistributionWidget } from "@/components/dashboard-v2/widgets/chart-widgets/workload-pending-time-distribution-widget";
import { NodePoolFreeResourcesWidget } from "@/components/dashboard-v2/widgets/chart-widgets/node-pool-free-resources-widget";
import { WorkloadByTypeWidget } from "@/components/dashboard-v2/widgets/chart-widgets/workload-by-type-widget/";
import { GpuAllocationRatioWidget } from "@/components/dashboard-v2/widgets/chart-widgets/gpu-allocation-ratio-widget/";
import { IdleAllocatedGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/idle-allocated-gpu-widget";
//time range widgets
import { OverviewTimeRangeWidgets } from "@/components/dashboard-v2/overview/overview-time-range-widgets";
//grid widgets
import { IdleGpuWorkloadWidget } from "@/components/dashboard-v2/widgets/grid-widget/idle-gpu-workload-widget";
import { OrgUnitRankingWidget } from "@/components/dashboard-v2/widgets/grid-widget/org-unit-ranking-widget";
//quota dashboard widgets
import { QuotaTotalGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/quota-total-gpu-widget";
import { AllocatedQuotaGpuWidget } from "@/components/dashboard-v2/widgets/stat-widgets/allocated-quota-gpu-widget";
import { QuotaAllocationTimeRangeWidget } from "@/components/dashboard-v2/widgets/time-range-widgets/quota-allocation-time-range-widget";
import { PendingWorkloadsByNodepoolWidget } from "@/components/dashboard-v2/widgets/chart-widgets/pending-workloads-by-nodepool-widget";
import { QuotaByNodepoolWidget } from "@/components/dashboard-v2/widgets/chart-widgets/quota-by-nodepool-widget";
import { AllocationByNodepoolWidget } from "@/components/dashboard-v2/widgets/chart-widgets/allocation-by-nodepool-widget";
import { DeptAllocationRatioWidget } from "@/components/dashboard-v2/widgets/chart-widgets/dept-allocation-ratio-widget";
import { ProjectAllocationRatioWidget } from "@/components/dashboard-v2/widgets/chart-widgets/project-allocation-ratio-widget";
//model
import { stringUtil } from "@/utils/string.util";
import { FailedWorkloadsWidget } from "@/components/dashboard-v2/widgets/stat-widgets/failed-workloads-widget";

const WidgetLayout = defineComponent({
  name: "widget-layout",
  components: {
    ReadyGpuWidget,
    AllocatedGpuWidget,
    ReadyNodesWidget,
    GpuAllocationRatioWidget,
    NodePoolFreeResourcesWidget,
    WorkloadByTypeWidget,
    WorkloadStatusWidget,
    WorkloadPendingTimeDistributionWidget,
    PendingWorkloadsWidget,
    FailedWorkloadsWidget,
    RunningWorkloadsWidget,
    IdleAllocatedGpuWidget,
    OverviewTimeRangeWidgets,
    IdleGpuWorkloadWidget,
    OrgUnitRankingWidget,
    QuotaTotalGpuWidget,
    AllocatedQuotaGpuWidget,
    QuotaAllocationTimeRangeWidget,
    PendingWorkloadsByNodepoolWidget,
    QuotaByNodepoolWidget,
    AllocationByNodepoolWidget,
    DeptAllocationRatioWidget,
    ProjectAllocationRatioWidget,
  },
  emits: ["date-changed"],
  props: {
    widgetsLayout: {
      type: Array as PropType<EWidgetComponent[][]>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    nodePoolName: {
      type: String as PropType<string>,
      required: false,
    },
    filterByDates: {
      type: Object as PropType<IRangeDates>,
      required: false,
    },
    hasDepartments: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    hasReadProjectsPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasReadDepartmentsPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hasReadClustersPermission: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    orgUnitMetricSource: {
      type: Object as PropType<IOrgUnitMetricSource | null>,
      default: null,
    },
  },
  methods: {
    getColClass(widgetCount: number): string {
      let columnSize = Math.floor(12 / widgetCount);

      if (columnSize === 1) {
        return "col";
      }
      if (columnSize === 6) {
        columnSize = 12;
      }

      return `col-md col-sm-${columnSize}`;
    },
    getWidgetAid(widgetName: string): string {
      return stringUtil.camelCaseToHyphenated(widgetName);
    },
    getWidgetKey(widgetName: string): string {
      // Include clusterId, nodePoolName, and orgUnitMetricSource in the key to force complete unmount/remount when filters change
      // This prevents chart drilldown state corruption issues
      const orgUnitKey = this.orgUnitMetricSource
        ? `${this.orgUnitMetricSource.type}-${this.orgUnitMetricSource.id}`
        : "all";
      return `${widgetName}-${this.clusterId}-${this.nodePoolName ?? "all"}-${orgUnitKey}`;
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default WidgetLayout as typeof WidgetLayout & DefineComponent;
</script>

<style scoped lang="scss">
.widget-pa-xs {
  padding: 3px 3px;
}
.layout-wrapper {
  padding: 3px;
}
</style>
