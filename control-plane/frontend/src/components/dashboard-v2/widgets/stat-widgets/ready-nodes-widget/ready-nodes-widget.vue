<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <stat-widget-wrapper
    :options="wrapperOptions"
    @link-clicked="redirectToNodes"
    :loading="displayLoading"
    :error="error"
  >
    <template #data> {{ statData }}</template>
  </stat-widget-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//routes
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
//store
import { useClusterStore } from "@/stores/cluster.store";
//models
import { EWidgetEntity, type IWidgetWrapperOptions } from "@/models/chart.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import type { IFilterBy } from "@/models/filter.model";
import { ETableFilters } from "@/models/table.model";
//cmps
import { StatWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/stat-widget-wrapper";
//svc
import { filterService } from "@/services/filter.service/filter.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { NodeTelemetryType } from "@/swagger-models/cluster-service-client";
//util
import { widgetUtil } from "@/utils/widget.util";
import { formatUtil } from "@/utils/format.util/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { nodeColumnName, nodeIndexColumns } from "@/table-models/node.table-model";

export default defineComponent({
  name: "ready-nodes-widget",
  components: { StatWidgetWrapper },
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
      readyNodes: 0 as number,
      totalNodes: 0 as number,
      error: false as boolean,
      displayLoading: true as boolean,
      clusterStore: useClusterStore(),
    };
  },
  created() {
    this.loadData();
    this.startRefreshLoadData();
  },
  computed: {
    statData(): string | number {
      const format = (val: number) => formatUtil.numberFormat(val, true);

      return `${format(this.readyNodes)}/${format(this.totalNodes)}`;
    },
    wrapperOptions(): IWidgetWrapperOptions {
      const title = "Ready / Total GPU nodes";
      return {
        linkText: "All nodes",
        title: title,
        entityType: EWidgetEntity.Node,
        tooltipText:
          "Ready: The number of GPU nodes that are available for scheduling </br>Total: The number of GPU nodes in the cluster",
      };
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.ReadyNodesWidget, this.loadData);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.ReadyNodesWidget);
    },
    async loadData(): Promise<void> {
      try {
        await Promise.all([
          this.loadNodesTelemetry(NodeTelemetryType.TotalGpuNodes),
          this.loadNodesTelemetry(NodeTelemetryType.ReadyGpuNodes),
        ]);

        this.error = false;
      } catch (error: unknown) {
        this.handleError(error);
      } finally {
        this.displayLoading = false;
      }
    },
    async loadNodesTelemetry(telemetryType: NodeTelemetryType): Promise<void> {
      const response = await clusterService.getNodeTelemetry(telemetryType, this.clusterId, this.nodePoolName);
      const values = widgetUtil.sumAllTelemetryValues(response.values);
      if (telemetryType === NodeTelemetryType.TotalGpuNodes) {
        this.totalNodes = values;
      } else {
        this.readyNodes = values;
      }
    },
    handleError(error: unknown): void {
      this.error = true;
      if (error instanceof HttpErrorResponse) {
        console.error(error.message);
      } else {
        console.error("Error fetching data", error);
      }
    },
    redirectToNodes(): void {
      this.setColumnFilter();

      this.$router.push({ name: NODE_ROUTE_NAMES.NODE_INDEX });
    },
    getNodesFilterBy(): IFilterBy {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(nodeColumnName.status, nodeIndexColumns);
      return filterService.loadFilters(window.location, ETableFilters.NODE, defaultFilters);
    },
    setColumnFilter(): void {
      const nodesFilterBy: IFilterBy = this.getNodesFilterBy();

      filterService.setColumnFilter(nodesFilterBy, "Ready", nodeColumnName.status, ETableFilters.NODE);
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
        this.loadData();
      },
    },
    clusterId: {
      handler(): void {
        this.displayLoading = true;
        this.loadData();
      },
    },
  },
  unmounted() {
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss"></style>
