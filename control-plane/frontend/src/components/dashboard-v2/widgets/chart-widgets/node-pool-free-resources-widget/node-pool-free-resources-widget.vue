<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <chart-widget-wrapper
    :options="wrapperOptions"
    :breadcrumbs="breadcrumbs"
    @link-clicked="redirectToNodes"
    @breadcrumb-clicked="drillUp"
    @export-csv="exportCsv"
    :error="freeGpusError || totalGpusError"
    :loading="displayLoading"
    :empty="isChartEmpty"
  >
    <node-pool-free-gpu-chart
      class="node-pool-free-gpu-chart"
      v-if="isScatterChart && !displayLoading"
      :node-pool-free-gpus-telemetry="nodePoolFreeGpusTelemetry"
      @scatter-point-click="handleNodePoolSelection"
    />
    <node-free-gpu-chart
      v-else-if="!displayLoading"
      class="node-free-gpu-chart"
      :node-pool-free-gpus-telemetry="nodePoolFreeGpusTelemetry"
      :node-pool-name="selectedNodePoolName"
      :node-pool-color="selectedNodePoolColor"
      :node-total-gpu-devices-telemetry="nodeTotalGpuDevicesTelemetry"
    />
  </chart-widget-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { ChartWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/chart-widget-wrapper";
import NodeFreeGpuChart from "@/components/dashboard-v2/widgets/chart-widgets/node-pool-free-resources-widget/node-free-gpu-chart/node-free-gpu-chart.vue";
import NodePoolFreeGpuChart from "@/components/dashboard-v2/widgets/chart-widgets/node-pool-free-resources-widget/node-pool-free-gpu-chart/node-pool-free-gpu-chart.vue";
//routes
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
//service
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
//model
import {
  NodeTelemetryType,
  type TelemetryResponse,
  type TelemetryResponseValuesInner,
} from "@/swagger-models/cluster-service-client";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { CHART_COLOR_PALETTE } from "@/models/chart.model";
import { EWidgetEntity, type IWidgetWrapperOptions, EWidgetEmptyMessages } from "@/models/chart.model";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
//store
import { useClusterStore } from "@/stores/cluster.store";
import type { IFilterBy } from "@/models/filter.model";
import { widgetUtil } from "@/utils/widget.util";
import { dashboardUtil } from "@/utils/dashboard.util";
import { filterService } from "@/services/filter.service/filter.service";
import { ETableFilters } from "@/models/table.model";
import { nodeColumnName, nodeIndexColumns } from "@/table-models/node.table-model";

export default defineComponent({
  name: "node-pool-free-resources-widget",
  components: { NodePoolFreeGpuChart, NodeFreeGpuChart, ChartWidgetWrapper },
  props: {
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    nodePoolName: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      isScatterChart: true,
      breadcrumbs: [] as string[],
      clusterStore: useClusterStore(),
      wrapperOptions: {
        title: "Free nodes by node pool",
        linkText: "All nodes",
        entityType: EWidgetEntity.Nodepool,
        tooltipText:
          "The number of nodes that are free<br/>, to ensure large workloads are<br/> being scheduled efficiently",
        emptyMessage: EWidgetEmptyMessages.NoNodes,
      } as IWidgetWrapperOptions,
      freeGpusError: false as boolean,
      totalGpusError: false as boolean,
      displayLoading: true as boolean,
      isChartEmpty: false as boolean,
      nodePoolFreeGpusTelemetry: [] as TelemetryResponseValuesInner[],
      nodeTotalGpuDevicesTelemetry: [] as TelemetryResponseValuesInner[],
      selectedNodePoolName: "" as string,
      selectedNodePoolColor: null as CHART_COLOR_PALETTE | null,
    };
  },
  async created() {
    const clusterName = this.clusterStore.clusterNameByUuid(this.clusterId);
    if (clusterName) {
      this.breadcrumbs.push(clusterName);
    }
    await this.loadInitialData();
  },
  methods: {
    async loadInitialData(): Promise<void> {
      await this.loadFreeGpusByNodePool();
      await this.loadTotalGpusDevices();
    },
    async loadTotalGpusDevices(): Promise<void> {
      try {
        const response: TelemetryResponse = await clusterService.getNodeTelemetry(
          NodeTelemetryType.TotalGpus,
          this.clusterId,
          this.nodePoolName,
          [ENodeTelemetryGroupBy.Node],
        );

        this.nodeTotalGpuDevicesTelemetry = response.values || [];
        this.totalGpusError = false;
        if (!this.isScatterChart) {
          this.isChartEmpty = this.nodeTotalGpuDevicesTelemetry.length === 0;
        }
      } catch (error: unknown) {
        this.totalGpusError = true;
        this.handleError(error);
      } finally {
        this.displayLoading = false;
      }
    },
    async loadFreeGpusByNodePool(): Promise<void> {
      try {
        const response: TelemetryResponse = await clusterService.getNodeTelemetry(
          NodeTelemetryType.FreeGpus,
          this.clusterId,
          this.nodePoolName,
          [ENodeTelemetryGroupBy.Nodepool, ENodeTelemetryGroupBy.Node],
        );

        this.nodePoolFreeGpusTelemetry = widgetUtil.sortByNodePool(response.values || []);
        if (this.isScatterChart) {
          this.isChartEmpty = this.nodePoolFreeGpusTelemetry.length === 0;
        }

        this.freeGpusError = false;
      } catch (error: unknown) {
        this.freeGpusError = true;
        this.handleError(error);
      }
    },
    async exportCsv(): Promise<void> {
      try {
        const telemetryType = this.isScatterChart ? NodeTelemetryType.FreeGpus : NodeTelemetryType.TotalGpus;
        const groupBy = this.isScatterChart
          ? [ENodeTelemetryGroupBy.Nodepool, ENodeTelemetryGroupBy.Node]
          : [ENodeTelemetryGroupBy.Node];
        await clusterService.getNodeTelemetryCsv(telemetryType, this.clusterId, this.nodePoolName, groupBy);
      } catch (error: unknown) {
        this.$q.notify(dashboardUtil.getCsvErrorMessage());
        this.handleError(error);
      }
    },
    handleNodePoolSelection(nodePoolName: string, nodePoolColor: CHART_COLOR_PALETTE | null): void {
      this.breadcrumbs.push(nodePoolName);
      this.selectedNodePoolName = nodePoolName;
      this.selectedNodePoolColor = nodePoolColor;
      this.isScatterChart = false;
      this.refreshChartLayout();
    },
    drillUp(index: number): void {
      if (index >= 0 && index < this.breadcrumbs.length) {
        this.breadcrumbs = this.breadcrumbs.slice(0, index + 1);
        this.isScatterChart = true;
        this.refreshChartLayout();
      }
    },
    refreshChartLayout(): void {
      this.$nextTick(() => {
        window.dispatchEvent(new Event("resize"));
      });
    },
    handleError(error: unknown): void {
      if (error instanceof HttpErrorResponse) {
        console.error(error.message);
      } else {
        console.error("Error fetching data", error);
      }
    },
    redirectToNodes(): void {
      this.setColumnFilter();
      this.$router.push({
        name: NODE_ROUTE_NAMES.NODE_INDEX,
      });
    },
    getNodesFilterBy(): IFilterBy {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(nodeColumnName.status, nodeIndexColumns);
      return filterService.loadFilters(window.location, ETableFilters.NODE, defaultFilters);
    },
    setColumnFilter(): void {
      const nodesFilterBy: IFilterBy = this.getNodesFilterBy();
      if (this.nodePoolName) {
        this.setNodePoolColumnFilter(nodesFilterBy);
      } else {
        this.removeNodePoolFilter(nodesFilterBy);
      }
    },
    setNodePoolColumnFilter(nodesFilterBy: IFilterBy): void {
      filterService.setColumnFilter(
        nodesFilterBy,
        this.nodePoolName as string,
        nodeColumnName.nodePool,
        ETableFilters.NODE,
      );
    },
    removeNodePoolFilter(nodesFilterBy: IFilterBy): void {
      filterService.removeColumnFilter(nodesFilterBy, nodeColumnName.nodePool, ETableFilters.NODE);
    },
  },
  watch: {
    nodePoolName: {
      handler(): void {
        this.displayLoading = true;
        this.loadInitialData();
      },
    },
    clusterId: {
      handler(): void {
        this.displayLoading = true;
        this.loadInitialData();
      },
    },
  },
});
</script>

<style lang="scss">
.node-pool-free-gpu-chart,
.node-free-gpu-chart {
  width: 100%;
  height: 240px;
}
</style>
