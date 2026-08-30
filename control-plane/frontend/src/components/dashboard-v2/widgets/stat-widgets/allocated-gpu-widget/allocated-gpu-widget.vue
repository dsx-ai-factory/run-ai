<template>
  <stat-widget-wrapper :options="widgetOptions" @link-clicked="handleLinkClick" :loading="displayLoading" :error="error">
    <template #data> {{ allocatedGpuFormatted }}</template>
  </stat-widget-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//routes
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
//models
import { EWidgetEntity, type IWidgetWrapperOptions } from "@/models/chart.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { WorkloadTelemetryType } from "@/swagger-models/workloads-client";
import { NodeTelemetryType } from "@/swagger-models/cluster-service-client";
//cmps
import { StatWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/stat-widget-wrapper";
import { NODE_ROUTE_NAMES } from "@/router/node.routes/node.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import type { TelemetryResponse } from "@/swagger-models/cluster-service-client";
import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";
import { widgetUtil } from "@/utils/widget.util";
import type { IFilterBy } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import { ETableFilters } from "@/models/table.model";
import { formatUtil } from "@/utils/format.util/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { nodeColumnName, nodeIndexColumns } from "@/table-models/node.table-model";
import { usePermissionStore } from "@/stores/permissions.store";

export default defineComponent({
  name: "allocated-gpu-widget",
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
      permissionStore: usePermissionStore(),
      allocatedGpu: 0 as number,
      displayLoading: true as boolean,
      error: false as boolean,
    };
  },
  created() {
    this.getGpuAllocationTelemetry();
    this.startRefreshLoadData();
  },
  computed: {
    allocatedGpuFormatted(): string {
      return formatUtil.numberFormat(this.allocatedGpu, true);
    },
    hasReadNodesPermission(): boolean {
      return this.permissionStore.hasReadNodesPermission();
    },
    widgetOptions(): IWidgetWrapperOptions {
      // Dynamic options based on permission - determines link text and entity type
      if (this.hasReadNodesPermission) {
        return {
          linkText: "All nodes",
          title: "Allocated GPU devices",
          entityType: EWidgetEntity.Node,
        };
      }
      return {
        linkText: "All workloads",
        title: "Allocated GPU devices",
        entityType: EWidgetEntity.Workload,
      };
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.AllocateGpuWidget, this.getGpuAllocationTelemetry);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.AllocateGpuWidget);
    },
    async getGpuAllocationTelemetry(): Promise<void> {
      try {
        this.error = false;
        // Use node telemetry if user has READ Nodes permission, otherwise fall back to workloads telemetry
        if (this.hasReadNodesPermission) {
          await this.getGpuAllocationFromNodes();
        } else {
          await this.getGpuAllocationFromWorkloads();
        }
      } catch (error: unknown) {
        this.handleError(error);
      } finally {
        this.displayLoading = false;
      }
    },
    async getGpuAllocationFromNodes(): Promise<void> {
      const response: TelemetryResponse = await clusterService.getNodeTelemetry(
        NodeTelemetryType.AllocatedGpus,
        this.clusterId,
        this.nodePoolName,
      );
      this.allocatedGpu = widgetUtil.sumAllTelemetryValues(response.values);
    },
    async getGpuAllocationFromWorkloads(): Promise<void> {
      const response: TelemetryResponse = await workloadService.getWorkloadsTelemetry(
        WorkloadTelemetryType.GpuAllocation,
        this.clusterId,
        this.nodePoolName,
        undefined,
        undefined,
        [EWorkloadTelemetryGroupBy.ClusterId],
      );
      this.allocatedGpu = widgetUtil.sumAllTelemetryValues(response.values);
    },
    handleError(error: unknown): void {
      this.error = true;
      if (error instanceof HttpErrorResponse) {
        console.error(error.message);
      } else {
        console.error("Error fetching data", error);
      }
    },
    handleLinkClick(): void {
      // Redirect based on permission
      if (this.hasReadNodesPermission) {
        this.redirectToNodes();
      } else {
        this.redirectToWorkloads();
      }
    },
    redirectToNodes(): void {
      this.setColumnFilter();
      this.$router.push({ name: NODE_ROUTE_NAMES.NODE_INDEX });
    },
    redirectToWorkloads(): void {
      this.$router.push({ name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX });
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
      async handler(): Promise<void> {
        this.displayLoading = true;
        await this.getGpuAllocationTelemetry();
      },
    },
    clusterId: {
      async handler(): Promise<void> {
        this.displayLoading = true;
        await this.getGpuAllocationTelemetry();
      },
    },
  },
  unmounted() {
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss"></style>
