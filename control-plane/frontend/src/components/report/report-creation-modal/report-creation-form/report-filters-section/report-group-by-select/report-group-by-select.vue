<template>
  <div class="row items-center justify-between no-wrap">
    <div class="q-py-sm q-px-md report-group-by-wrapper">
      <runai-select
        :options="groupByOptions"
        :model-value="groupBy"
        label="Segment by"
        outlined
        @update:model-value="$emit('update:group-by', $event.value)"
      />
    </div>
    <q-separator vertical />
    <template v-for="(column, index) in exportedDocumentColumns" :key="column">
      <div class="q-px-sm text-bold">
        {{ column }}
      </div>
      <q-separator vertical v-if="index !== exportedDocumentColumns.length - 1" />
      <div v-else></div>
    </template>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//store
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
//cmps
import RunaiSelect from "@/components/common/runai-select/runai-select.vue";
//svc
import { ReportGroupByOptions } from "@/swagger-models/org-unit-service-client";
//model
import { hiddenExportedColumnsByReportGroupBy, ReportExportedDocumentColumns } from "@/models/report.model";

export default defineComponent({
  name: "report-group-by-select",
  components: { RunaiSelect },
  emits: ["update:group-by"],
  props: {
    groupBy: {
      type: [String, null] as PropType<ReportGroupByOptions | null>,
      required: false,
    },
  },
  data: function () {
    return {
      settingsStore: useSettingStore(),
      clusterStore: useClusterStore(),
    };
  },
  computed: {
    exportedDocumentColumns(): typeof ReportExportedDocumentColumns {
      if (!this.groupBy) return ReportExportedDocumentColumns;
      return ReportExportedDocumentColumns.filter(
        (column) => !hiddenExportedColumnsByReportGroupBy[this.groupBy as ReportGroupByOptions].includes(column),
      );
    },
    groupByOptions() {
      return [
        { label: "Node pool", value: ReportGroupByOptions.Nodepool },
        { label: "Project", value: ReportGroupByOptions.Project },
        { label: "Department", value: ReportGroupByOptions.Department },
        { label: "Cluster", value: ReportGroupByOptions.Cluster },
      ].filter((option) => {
        if (option.value === ReportGroupByOptions.Cluster) {
          return this.clusterStore.isMultiCluster;
        }
        return true;
      });
    },
  },
});
</script>

<style scoped lang="scss">
.report-group-by-wrapper {
  min-width: 180px;
}
</style>
