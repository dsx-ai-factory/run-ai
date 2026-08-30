<template>
  <stat-widget-wrapper :options="widgetOptions" :loading="displayLoading" :error="error">
    <template #data>
      <div class="stat-row">
        <span v-if="isOrgUnitMode">{{ quotaFormatted }}</span>
        <span v-else>{{ quotaFormatted }}/{{ totalFormatted }}</span>
        <span v-if="!isOrgUnitMode" class="percentage">{{ percentage }}</span>
      </div>
    </template>
  </stat-widget-wrapper>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from "vue";
import type { IWidgetWrapperOptions } from "@/models/chart.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { StatWidgetWrapper } from "@/components/dashboard-v2/widgets/common/widget-wrapper/stat-widget-wrapper";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { OrgUnitTelemetryType } from "@/swagger-models/org-unit-service-client";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { NodeTelemetryType } from "@/swagger-models/cluster-service-client";
import { widgetUtil } from "@/utils/widget.util";
import { formatUtil } from "@/utils/format.util/format.util";
import { intervalUtil } from "@/utils/interval.util";
import { EIntervalLabels } from "@/models/interval.model";
import { EOrgUnitMetricSourceType, type IOrgUnitMetricSource } from "@/models/org-unit.model";

export interface Props {
  clusterId: string;
  nodePoolName?: string;
  orgUnitMetricSource?: IOrgUnitMetricSource | null;
}

const props = withDefaults(defineProps<Props>(), {
  orgUnitMetricSource: null,
});

const quota = ref(0);
const total = ref(0);
const displayLoading = ref(true);
const error = ref(false);

const isOrgUnitMode = computed(() => !!props.orgUnitMetricSource);

const quotaFormatted = computed(() => formatUtil.numberFormat(quota.value, true));
const totalFormatted = computed(() => formatUtil.numberFormat(total.value, true));

const percentage = computed(() => {
  if (total.value === 0) return "";
  return `${Math.round((quota.value / total.value) * 100)}%`;
});

const widgetOptions = computed<IWidgetWrapperOptions>(() => {
  if (isOrgUnitMode.value) {
    return {
      title: "GPU quota",
      tooltipText: "Quota: The number of GPUs allocated to the selected department or project",
    };
  }
  return {
    title: "GPU quota / total",
    tooltipText:
      "Quota: The number of GPUs allocated to the subordinate departments or projects<br>Total: The number of GPUs in the cluster",
  };
});

function getNodepoolQuota(resources: Array<{ nodePool?: { name?: string } | null; gpu: { deserved: number } }>): number {
  const entry = resources.find((r) => r.nodePool?.name === props.nodePoolName);
  return entry?.gpu.deserved ?? 0;
}

async function fetchOrgUnitQuota(): Promise<number> {
  const source = props.orgUnitMetricSource;
  if (!source) return 0;

  if (source.type === EOrgUnitMetricSourceType.Project) {
    const project = await orgUnitService.getProject(source.id);
    if (props.nodePoolName) return getNodepoolQuota(project.resources);
    return project.totalResources?.gpuQuota ?? 0;
  }
  if (source.type === EOrgUnitMetricSourceType.Department) {
    const department = await orgUnitService.getDepartment(source.id);
    if (props.nodePoolName) return getNodepoolQuota(department.resources);
    return department.totalResources?.gpuQuota ?? 0;
  }
  return 0;
}

async function fetchData(): Promise<void> {
  try {
    error.value = false;

    if (isOrgUnitMode.value) {
      quota.value = await fetchOrgUnitQuota();
    } else {
      const [quotaResponse, totalResponse] = await Promise.all([
        orgUnitService.getProjectsTelemetry(OrgUnitTelemetryType.GpuQuota, props.clusterId, props.nodePoolName),
        clusterService.getNodeTelemetry(NodeTelemetryType.TotalGpus, props.clusterId, props.nodePoolName),
      ]);
      quota.value = widgetUtil.sumAllTelemetryValues(quotaResponse.values);
      total.value = widgetUtil.sumAllTelemetryValues(totalResponse.values);
    }
  } catch (err: unknown) {
    error.value = true;
    if (err instanceof HttpErrorResponse) {
      console.error(err.message);
    } else {
      console.error("Error fetching data", err);
    }
  } finally {
    displayLoading.value = false;
  }
}

watch(
  () => [props.clusterId, props.nodePoolName, props.orgUnitMetricSource],
  async () => {
    displayLoading.value = true;
    await fetchData();
  },
  { deep: true },
);

fetchData();
intervalUtil.startInterval(EIntervalLabels.QuotaTotalGpuWidget, fetchData);

onUnmounted(() => {
  intervalUtil.stopInterval(EIntervalLabels.QuotaTotalGpuWidget);
});
</script>

<style scoped lang="scss">
.stat-row {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}
.percentage {
  font-size: 0.5em;
}
</style>
