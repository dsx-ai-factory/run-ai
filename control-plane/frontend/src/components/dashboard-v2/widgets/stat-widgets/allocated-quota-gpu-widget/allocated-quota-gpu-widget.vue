<template>
  <stat-widget-wrapper :options="widgetOptions" :loading="displayLoading" :error="error">
    <template #data>
      <div class="stat-row">
        <span>{{ allocatedFormatted }}/{{ quotaFormatted }}</span>
        <span class="percentage" :class="percentageColorClass">{{ percentage }}</span>
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

const allocated = ref(0);
const quota = ref(0);
const displayLoading = ref(true);
const error = ref(false);

const allocatedFormatted = computed(() => formatUtil.numberFormat(allocated.value, true));
const quotaFormatted = computed(() => formatUtil.numberFormat(quota.value, true));

const percentageValue = computed(() => {
  if (quota.value === 0) return 0;
  return Math.round((allocated.value / quota.value) * 100);
});

const percentage = computed(() => {
  if (quota.value === 0) return "";
  return `${percentageValue.value}%`;
});

const HIGH_UTILIZATION_THRESHOLD = 90;
const MEDIUM_UTILIZATION_THRESHOLD = 50;

const percentageColorClass = computed(() => {
  if (percentageValue.value >= HIGH_UTILIZATION_THRESHOLD) return "feedback-success";
  if (percentageValue.value >= MEDIUM_UTILIZATION_THRESHOLD) return "feedback-warning";
  return "feedback-error";
});

const widgetOptions = computed<IWidgetWrapperOptions>(() => ({
  title: "GPU allocated / quota",
  tooltipText:
    "Allocated: The number of GPUs used by the workloads<br>Quota: The number of GPUs allocated to the subordinate departments or projects",
}));

function getNodepoolOrgUnitData(
  resources: Array<{ nodePool?: { name?: string } | null; gpu: { deserved: number } }>,
  nodePoolQuotaStatuses?: Array<{ nodePoolName: string; allocated?: { gpu?: number } }>,
): { allocated: number; quota: number } {
  const quotaEntry = resources.find((r) => r.nodePool?.name === props.nodePoolName);
  const statusEntry = nodePoolQuotaStatuses?.find((s) => s.nodePoolName === props.nodePoolName);
  return {
    allocated: statusEntry?.allocated?.gpu ?? 0,
    quota: quotaEntry?.gpu.deserved ?? 0,
  };
}

async function fetchOrgUnitData(): Promise<{ allocated: number; quota: number }> {
  const source = props.orgUnitMetricSource;
  if (!source) return { allocated: 0, quota: 0 };

  if (source.type === EOrgUnitMetricSourceType.Project) {
    const project = await orgUnitService.getProject(source.id);
    if (props.nodePoolName) return getNodepoolOrgUnitData(project.resources, project.status?.nodePoolQuotaStatuses);
    return {
      allocated: project.status?.quotaStatus?.allocated?.gpu ?? 0,
      quota: project.totalResources?.gpuQuota ?? 0,
    };
  }
  if (source.type === EOrgUnitMetricSourceType.Department) {
    const department = await orgUnitService.getDepartment(source.id);
    if (props.nodePoolName)
      return getNodepoolOrgUnitData(department.resources, department.status?.nodePoolQuotaStatuses);
    return {
      allocated: department.status?.quotaStatus?.allocated?.gpu ?? 0,
      quota: department.totalResources?.gpuQuota ?? 0,
    };
  }
  return { allocated: 0, quota: 0 };
}

async function fetchData(): Promise<void> {
  try {
    error.value = false;

    if (props.orgUnitMetricSource) {
      const data = await fetchOrgUnitData();
      allocated.value = data.allocated;
      quota.value = data.quota;
    } else {
      const [allocationResponse, quotaResponse] = await Promise.all([
        orgUnitService.getProjectsTelemetry(OrgUnitTelemetryType.GpuAllocation, props.clusterId, props.nodePoolName),
        orgUnitService.getProjectsTelemetry(OrgUnitTelemetryType.GpuQuota, props.clusterId, props.nodePoolName),
      ]);
      allocated.value = widgetUtil.sumAllTelemetryValues(allocationResponse.values);
      quota.value = widgetUtil.sumAllTelemetryValues(quotaResponse.values);
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
intervalUtil.startInterval(EIntervalLabels.AllocatedQuotaGpuWidget, fetchData);

onUnmounted(() => {
  intervalUtil.stopInterval(EIntervalLabels.AllocatedQuotaGpuWidget);
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

  &.feedback-success {
    color: var(--feedback-success);
  }
  &.feedback-warning {
    color: var(--feedback-warning);
  }
  &.feedback-error {
    color: var(--feedback-error);
  }
}
</style>
