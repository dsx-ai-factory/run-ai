<template>
  <runai-page-actions
    :columns="columns"
    :filters="filterBy"
    @filters-changed="$emit('filters-changed', $event)"
    @selected-actions-close="$emit('selected-actions-close')"
    primary-btn-label="New report"
    :selected-rows-amount="selectedRowsAmount"
    advanced-filters
    hide-export-csv
    hide-cluster-filter
  >
    <template v-slot:primary-btn> <create-report-dropdown @create="$emit('action', reportAction.Create)" /></template>
    <template #selected-rows-actions>
      <!-- Download CSV -->
      <runai-tooltip-wrapper :display-tooltip="isDownloadBtnDisabled" :tooltip-text="downloadBtnTooltip">
        <runai-action-button
          btn-action="downloadCsv"
          :disable="isDownloadBtnDisabled"
          aid="download-report-btn"
          @click="$emit('action', reportAction.Download)"
        />
      </runai-tooltip-wrapper>

      <!-- Delete -->
      <runai-tooltip-wrapper :display-tooltip="isDeleteBtnDisabled" :tooltip-text="deleteBtnTooltip">
        <runai-action-button
          v-permission="{ resourceType: resourceType.Project, action: action.Delete }"
          btn-action="delete"
          aid="delete-report-btn"
          @click="$emit('action', reportAction.Delete)"
        />
      </runai-tooltip-wrapper>
    </template>
  </runai-page-actions>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
//model
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import type { ISelectOption } from "@/models/global.model";
import { ReportStatus } from "@/swagger-models/org-unit-service-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { EReportAction, EReportActionTooltip, EReportType } from "@/models/report.model";
//store
import { usePermissionStore } from "@/stores/permissions.store";
import type { IReportTable } from "@/table-models/report.table-model";
import CreateReportDropdown from "@/components/report/create-report-dropdown/create-report-dropdown.vue";

export default defineComponent({
  name: "report-page-actions",
  components: { CreateReportDropdown, RunaiTooltipWrapper, RunaiActionButton, RunaiPageActions },
  emits: ["filters-changed", "selected-actions-close", "action"],
  props: {
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
    columns: {
      type: Array as PropType<ITableColumn[]>,
      required: true,
    },
    selectedRowsAmount: {
      type: Number as PropType<number>,
      required: true,
    },
    selectedReport: {
      type: [Object, null] as PropType<IReportTable | null>,
      required: false,
    },
  },
  computed: {
    reportAction(): typeof EReportAction {
      return EReportAction;
    },
    action(): typeof Action {
      return Action;
    },
    resourceType(): typeof ResourceType {
      return ResourceType;
    },
    newReportOptions(): ISelectOption[] {
      return [{ label: "Consumption", value: EReportType.Consumption }];
    },
    hasDeletePermission(): boolean {
      return usePermissionStore().hasDeletePermission(ResourceType.Project);
    },
    isDeleteBtnDisabled(): boolean {
      return !this.hasDeletePermission;
    },
    isDownloadBtnDisabled(): boolean {
      if (!this.selectedReport) return true;
      return this.selectedReport.status !== ReportStatus.Ready;
    },
    /*** tooltips ***/
    downloadBtnTooltip(): string {
      return this.isDownloadBtnDisabled && this.selectedReport
        ? "To download, the selected report must be in status Ready."
        : "";
    },
    deleteBtnTooltip(): string {
      return this.isDeleteBtnDisabled ? EReportActionTooltip.NoDeletePermission : "";
    },
  },
});
</script>

<style scoped lang="scss"></style>
