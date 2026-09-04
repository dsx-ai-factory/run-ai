<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section aid="report-filters-section" class="q-pt-lg">
    <runai-section-title
      title="Set the report segmentation and filters"
      width="100%"
      tooltip-text="Data will be collected and aggregated based on the segment"
    />
    <div class="report-filters-wrapper">
      <div class="q-pa-sm">
        <runai-advanced-filters
          :filter-by="filters"
          :filter-options="filterOptions"
          :apply-cluster-filter-globally="false"
          @update-filter-by="updateAdvancedFilters"
        />
      </div>
      <q-separator />

      <report-group-by-select :group-by="groupBy" @update:group-by="$emit('update:group-by', $event)" />
    </div>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import RunaiSectionTitle from "@/components/common/runai-section-title/runai-section-title.vue";
import RunaiAdvancedFilters from "@/components/common/runai-advanced-filters/runai-advanced-filters.vue";
import ReportGroupBySelect from "@/components/report/report-creation-modal/report-creation-form/report-filters-section/report-group-by-select/report-group-by-select.vue";
//svc
import { filterService } from "@/services/filter.service/filter.service";
//model
import { EColumnFilterType, EFilterOperator, type IAdvancedFilterOption, type IFilterBy } from "@/models/filter.model";
import type { ReportGroupByOptions } from "@/swagger-models/org-unit-service-client";

export enum EReportFilterByFields {
  Project = "projectName",
  Department = "departmentName",
}

export default defineComponent({
  name: "report-filters-section",
  components: { ReportGroupBySelect, RunaiAdvancedFilters, RunaiSectionTitle },
  emits: ["update:filter-by", "update:group-by"],
  props: {
    filterBy: {
      type: [Array, null] as PropType<string[] | null>,
      required: false,
    },
    groupBy: {
      type: [String, null] as PropType<ReportGroupByOptions | null>,
      required: false,
    },
  },
  data: function () {
    return {
      filters: {} as IFilterBy,
      filterOptions: [
        {
          field: () => "",
          label: "Project",
          name: EReportFilterByFields.Project,
          type: EColumnFilterType.FreeText,
          excludeOperators: [EFilterOperator.DoesNotContain, EFilterOperator.NotEquals, EFilterOperator.Contains],
        },
      ] as IAdvancedFilterOption[],
    };
  },
  created() {
    this.initFilterOptions();
  },
  methods: {
    initFilterOptions(): void {
      this.filterOptions.push({
        field: () => "",
        label: "Department",
        name: EReportFilterByFields.Department,
        type: EColumnFilterType.FreeText,
        excludeOperators: [EFilterOperator.DoesNotContain, EFilterOperator.NotEquals, EFilterOperator.Contains],
      });
    },
    updateAdvancedFilters(filterBy: IFilterBy) {
      const parsedFilters = filterService.mapColumnsFilterToFilterParams(filterBy, [
        ...Object.values(EReportFilterByFields),
      ]);
      this.filters = filterBy;
      this.$emit("update:filter-by", parsedFilters.filterBy);
    },
  },
});
</script>

<style scoped lang="scss">
.report-filters-wrapper {
  border: 1px solid $black-12;
  border-radius: 4px;
}
</style>
