<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <runai-table-wrapper :filters-object="filterBy" aid="report-index">
      <template v-slot:actions>
        <report-page-actions
          :selected-report="selectedRow"
          :selected-rows-amount="selectedRowsAmount"
          :filter-by="filterBy"
          :columns="columns"
          @filters-changed="setFilterBy"
          @selected-actions-close="resetSelectedRows"
          @action="handleReportAction"
        />
      </template>
      <template v-slot:table>
        <runai-table
          :rows="tableData"
          :columns="columns"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          :loading="loadingTableData"
          @update-filters="setFilterBy"
          @clear-filters="clearFilters"
          :top-row="lastCreatedEntity"
          :get-row-key="getRowKey"
          sticky-columns
          :bordered="false"
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25, 50, 100]"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedEntity"
              :filter-by="filterBy"
              entity-name="report"
              ph-icon-name="ph-bold ph-table"
              icon-size="44"
              :show-error="loadingError"
              @clear-filters="clearFilters"
            >
              <template #primary-btn>
                <create-report-dropdown @create="handleReportAction(EReportAction.Create)" flat />
              </template>
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>
    <report-creation-modal
      v-if="isCreateReportModalOpen"
      @close="isCreateReportModalOpen = false"
      @report-created="onReportCreated"
    />
    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedRow"
      entity-type="report"
      :entity-name="selectedRow.name"
      :deleting="deleting"
      @cancel="isDeleteModalOpen = false"
      @delete="deleteReport"
    />
  </section>
</template>
<script lang="ts">
import { computed, defineComponent, ref } from "vue";
import { useQuasar } from "quasar";
// store
import { useAppStore } from "@/stores/app.store";
// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper/";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiTable } from "@/components/common/runai-table/";
import { ReportPageActions } from "@/components/report/report-page-actions/";
// model
import { type IReportTable, reportIndexColumns } from "@/table-models/report.table-model";
import type { IFilterBy } from "@/models/filter.model";
import { EReportAction } from "@/models/report.model";
import { ReportCreationModal } from "@/components/report/report-creation-modal";
import { RunaiDeleteModal } from "@/components/common/runai-delete-modal/";
// service
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
//util
import { alertUtil } from "@/utils/alert.util";
import { ErrorAlert } from "@/utils/error-alert.util";
// other
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import CreateReportDropdown from "@/components/report/create-report-dropdown/create-report-dropdown.vue";
import DOCS from "@/constants/docs";

export default defineComponent({
  name: "report-index",
  components: {
    CreateReportDropdown,
    RunaiDeleteModal,
    ReportCreationModal,
    ReportPageActions,
    RunaiTable,
    RunaiTableNoData,
    RunaiTableWrapper,
  },
  setup() {
    const appStore = useAppStore();
    const q = useQuasar();
    const isCreateReportModalOpen = ref<boolean>(false);
    const isDeleteModalOpen = ref<boolean>(false);
    const deleting = ref<boolean>(false);

    /*** api ***/
    const listReports = async (filterBy: IFilterBy) => {
      return await orgUnitService.listReports(filterBy);
    };
    const getReportById = async (id: string) => await orgUnitService.getReportById(id);

    const downloadReport = async (report: IReportTable) => {
      const dismiss = q.notify(alertUtil.getLoading(`Downloading`));
      try {
        await orgUnitService.downloadReport(report.id);
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: DOCS.SUPPORT.URL
            ? `Report ${report.name} couldn't be downloaded. Try again. If the issue persists, contact ${DOCS.SUPPORT.LABEL}.`
            : `Report ${report.name} couldn't be downloaded. Try again.`,
        });
        q.notify(errorAlert.getNotification(error));
      } finally {
        dismiss();
      }
    };

    const deleteReport = async () => {
      if (!selectedRow.value) return;
      const reportName = selectedRow.value.name;
      deleting.value = true;
      try {
        await orgUnitService.deleteReport(selectedRow.value.id);
        removeRow(selectedRow.value);
        q.notify(alertUtil.getSuccess(`Report ${reportName} deleted`));
        await refreshList();
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: DOCS.SUPPORT.URL
            ? `Report ${reportName} couldn't be deleted. Try again. If the issue persists, contact ${DOCS.SUPPORT.LABEL}.`
            : `Report ${reportName} couldn't be deleted. Try again.`,
        });
        q.notify(errorAlert.getNotification(error));
      } finally {
        deleting.value = false;
        isDeleteModalOpen.value = false;
        resetSelectedRows();
      }
    };

    const columns = computed(() => reportIndexColumns);

    const getRowKey = (row: IReportTable) => row.id;

    const onReportCreated = (report: IReportTable) => {
      refreshList(true, report.id);
      isCreateReportModalOpen.value = false;
    };

    const handleReportAction = (action: EReportAction) => {
      switch (action) {
        case EReportAction.Create:
          isCreateReportModalOpen.value = true;
          break;
        case EReportAction.Delete:
          isDeleteModalOpen.value = true;
          break;
        case EReportAction.Download:
          if (selectedRow.value) downloadReport(selectedRow.value);
          break;
      }
    };

    const {
      setFilterBy,
      clearFilters,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      selectedRows,
      resetSelectedRows,
      selectedRowsAmount,
      initTableFilter,
      refreshList,
      selectedRow,
      removeRow,
    } = useTableDataAdvancedFilters<IReportTable>(listReports, columns, EAdvancedIndexPages.REPORT, getReportById);

    return {
      appStore,
      filterBy,
      setFilterBy,
      columns,
      selectedRows,
      selectedRow,
      resetSelectedRows,
      selectedRowsAmount,
      tableData,
      loadingTableData,
      clearFilters,
      lastCreatedEntity,
      loadingError,
      getRowKey,
      isCreateReportModalOpen,
      isDeleteModalOpen,
      deleting,
      deleteReport,
      initTableFilter,
      onReportCreated,
      handleReportAction,
      EReportAction,
    };
  },
  created() {
    this.appStore.setPageLoading(false);
    this.initTableFilter();
  },
});
</script>

<style scoped lang="scss"></style>
