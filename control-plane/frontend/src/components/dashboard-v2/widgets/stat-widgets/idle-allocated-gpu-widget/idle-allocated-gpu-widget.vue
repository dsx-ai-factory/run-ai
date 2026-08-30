<template>
  <stat-widget-wrapper :options="options" @link-clicked="redirectToWorkloads" :loading="displayLoading" :error="error">
    <template #data> {{ idleAllocatedGpusFormatted }}</template>
  </stat-widget-wrapper>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { StatWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/stat-widget-wrapper";
import { EWidgetEntity, type IWidgetWrapperOptions } from "@/models/chart.model";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { workloadService } from "@/services/cluster/workload.service/workload.service";
import { NodeTelemetryType, type TelemetryResponse } from "@/swagger-models/cluster-service-client";
import { WorkloadTelemetryType } from "@/swagger-models/workloads-client";
import { widgetUtil } from "@/utils/widget.util";
import { HttpErrorResponse } from "@/models/http-response.model";
import { formatUtil } from "@/utils/format.util/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";
import { EWorkloadTelemetryGroupBy } from "@/models/workload.model";
import { usePermissionStore } from "@/stores/permissions.store";

export default defineComponent({
  name: "idle-allocated-gpu-widget",
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
      options: {
        linkText: "All workloads",
        title: "Idle allocated GPU devices",
        entityType: EWidgetEntity.Workload,
        tooltipText: "Allocated GPU devices that are idle for more than 5 minutes",
      } as IWidgetWrapperOptions,
      idleAllocatedGpus: 0 as number,
      error: false as boolean,
      displayLoading: true as boolean,
    };
  },
  created() {
    this.getIdleAllocatedGpusTelemetry();
    this.startRefreshLoadData();
  },
  computed: {
    idleAllocatedGpusFormatted(): string {
      return formatUtil.numberFormat(this.idleAllocatedGpus, true);
    },
    hasReadNodesPermission(): boolean {
      return this.permissionStore.hasReadNodesPermission();
    },
  },
  methods: {
    startRefreshLoadData(): void {
      intervalUtil.startInterval(EIntervalLabels.IdleAllocatedGpuWidget, this.getIdleAllocatedGpusTelemetry);
    },
    stopRefreshLoadData(): void {
      intervalUtil.stopInterval(EIntervalLabels.IdleAllocatedGpuWidget);
    },
    async getIdleAllocatedGpusTelemetry(): Promise<void> {
      try {
        this.error = false;
        // Use node telemetry if user has READ Nodes permission, otherwise fall back to workloads telemetry
        if (this.hasReadNodesPermission) {
          await this.getIdleAllocatedGpusFromNodes();
        } else {
          await this.getIdleAllocatedGpusFromWorkloads();
        }
      } catch (error: unknown) {
        this.handleError(error);
      } finally {
        this.displayLoading = false;
      }
    },
    async getIdleAllocatedGpusFromNodes(): Promise<void> {
      const response: TelemetryResponse = await clusterService.getNodeTelemetry(
        NodeTelemetryType.IdleAllocatedGpus,
        this.clusterId,
        this.nodePoolName,
        [],
      );
      this.idleAllocatedGpus = widgetUtil.sumAllTelemetryValues(response.values);
    },
    async getIdleAllocatedGpusFromWorkloads(): Promise<void> {
      const response: TelemetryResponse = await workloadService.getWorkloadsTelemetry(
        WorkloadTelemetryType.IdleAllocatedGpus,
        this.clusterId,
        this.nodePoolName,
        undefined,
        undefined,
        [EWorkloadTelemetryGroupBy.ClusterId],
      );
      this.idleAllocatedGpus = widgetUtil.sumAllTelemetryValues(response.values);
    },
    handleError(error: unknown): void {
      this.error = true;
      if (error instanceof HttpErrorResponse) {
        console.error(error.message);
      } else {
        console.error("Error fetching data", error);
      }
    },
    redirectToWorkloads(): void {
      this.$router.push({
        name: WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX,
        query: { sortBy: "idleAllocatedGpus", descending: "true" },
      });
    },
  },
  watch: {
    nodePoolName: {
      handler(): void {
        this.displayLoading = true;
        this.getIdleAllocatedGpusTelemetry();
      },
    },
    clusterId: {
      handler(): void {
        this.displayLoading = true;
        this.getIdleAllocatedGpusTelemetry();
      },
    },
  },
  unmounted() {
    this.stopRefreshLoadData();
  },
});
</script>

<style scoped lang="scss"></style>
