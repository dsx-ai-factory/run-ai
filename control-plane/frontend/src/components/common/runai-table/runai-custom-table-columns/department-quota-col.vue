<template>
  <q-td :aid="column.name" :class="colClass">
    {{ deservedGpuDisplayValue }}
    <q-tooltip v-if="tooltipText" max-width="430px" :offset="[180, 0]" anchor="top right">
      {{ tooltipText }}
    </q-tooltip></q-td
  >
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { ITableColumn } from "@/models/table.model";
import type { IDepartmentTable } from "@/models/department.model";
import { resourceUtil } from "@/utils/resource.util";
import { EResourceType } from "@/models/resource.model";
import { useSettingStore } from "@/stores/setting.store";
import type { AggregatedResources } from "@/swagger-models/org-unit-service-client";

export default defineComponent({
  name: "department-quota-col",
  props: {
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
    row: {
      type: Object as PropType<IDepartmentTable>,
      required: true,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
    };
  },
  computed: {
    deservedGpuDisplayValue(): string {
      if (this.deservedGpu < 0) {
        return "Unlimited";
      }
      return resourceUtil.getResourceDisplayValue(this.row.totalResources.gpuQuota || 0, EResourceType.GPU);
    },
    deservedGpu(): number {
      return this.row.totalResources.gpuQuota || 0;
    },
    isLimitQuotaOverSubscriptionEnabled(): boolean {
      return this.settingStore.isLimitQuotaOverSubscriptionEnabled;
    },
    colClass(): string {
      if (this.deservedGpu < 0) {
        return "";
      }
      if (this.isDepartmentExceedingProjectQuota) {
        return this.isLimitQuotaOverSubscriptionEnabled ? "negative" : "amber-14";
      }
      return "";
    },
    projectAggregatedGpus(): number {
      if (!this.row?.projectsAggregatedResources) return 0;
      return this.row.projectsAggregatedResources.reduce(
        (acc, project: AggregatedResources) => acc + (project?.gpu?.deserved || 0),
        0,
      );
    },
    isDepartmentExceedingProjectQuota(): boolean {
      return this.deservedGpu < this.projectAggregatedGpus;
    },
    tooltipText(): string {
      if (this.projectAggregatedGpus === 0) {
        return "";
      }
      return this.isDepartmentExceedingProjectQuota
        ? `The department's quota is less than the quota assigned to its subordinate projects. To remain within the limit, edit the department and enter a number above ${this.projectAggregatedGpus}.`
        : "";
    },
  },
});
</script>

<style scoped lang="scss">
.negative {
  color: $negative;
}
.amber-14 {
  color: $amber-14;
}
</style>
