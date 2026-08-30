<template>
  <section class="department-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <section class="predefined-date">
          <runai-datepicker-with-predefined
            @changed="updateDates"
            :range-of-dates="rangeOfDatesCallback"
            :predefined-options="extendedDefaultPredefinedOptions"
            info-bar-message="Select a period of up to 30 days within the last 90 days"
          />
        </section>
        <runai-page-actions
          :show-primary-btn="false"
          :show-search="false"
          :columns="auditIndexColumns"
          :filters="filterBy"
          @filters-changed="updateFilterBy"
          filter-input-label="Equals"
          :action-menu-items="actionMenuItems"
          @export-csv="exportTableAsCsv"
          advanced-filters
          hide-cluster-filter
        >
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="tableData"
          :columns="auditIndexColumns"
          :filter-by="filterBy"
          :loading="loadingTableData"
          @update-filters="updateFilterBy"
          sticky-columns
          disable-selection
          :bordered="false"
          is-server-side-pagination
          :rows-per-page-options="[5, 7, 10, 15, 20, 25, 50, 100]"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData"
              :filter-by="filterBy"
              entity-name="audit"
              ph-icon-name="ph-bold ph-list-dashes"
              icon-size="44"
              :show-create-btn="false"
              @clear-filters="clearFilters"
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>
  </section>
</template>

<script lang="ts">
import { computed, defineComponent, ref } from "vue";
import { useAppStore } from "@/stores/app.store";
import { auditIndexColumns } from "@/table-models/audit.table-model";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiDatepickerWithPredefined } from "@/components/common/runai-datepicker-with-predefined";
import type { IFilterBy } from "@/models/filter.model";
import fileDownload from "js-file-download";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// services
import { dateUtil, TimeUnit } from "@/utils/date.util";
import type { ITableColumn } from "@/models/table.model";
import type { ActionMenuItem } from "@/components/common/runai-page-actions/model";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { auditService } from "@/services/control-plane/audit.service/audit.service";
import type { AuditLog, GetAuditLogs200Response } from "@/swagger-models/audit-service-client";
import { alertUtil } from "@/utils/alert.util";
import { extendedDefaultPredefinedOptions } from "@/utils/date.util/date.util";

const AUDIT_LOG_DOWNLOAD_LIMIT = 100_000;

export default defineComponent({
  components: {
    RunaiPageActions,
    RunaiDatepickerWithPredefined,
    RunaiTableWrapper,
    RunaiTable,
    RunaiTableNoData,
  },
  data() {
    return {
      appStore: useAppStore(),
      isExportingCSV: false,
      isExportingJSON: false,
      actionMenuItems: [
        {
          label: "Download as JSON",
          action: () => this.exportTableAsJson(),
        },
      ] as ActionMenuItem[],
      auditIndexColumns,
    };
  },
  setup() {
    const clusterStore = useClusterStore();
    const settingStore = useSettingStore();
    const today = new Date();
    const startDate = ref(dateUtil.convertDateToISO(dateUtil.adjustDateBy(TimeUnit.day, today, -7)));
    const endDate = ref(dateUtil.convertDateToISO(today));

    const getAudits = async (filterBy: IFilterBy): Promise<AuditLog[]> => {
      const { audit_logs = [], total = 0 }: GetAuditLogs200Response = await auditService.listAudits(
        filterBy,
        startDate.value,
        endDate.value,
      );
      updateFilterByKey("rowsNumber", total);
      return audit_logs;
    };

    const columns = computed<ITableColumn[]>(() => auditIndexColumns);

    const {
      setFilterBy,
      clearFilters,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      initTableFilter,
      updateFilterByKey,
    } = useTableDataAdvancedFilters<AuditLog>(getAudits, columns, EAdvancedIndexPages.AUDIT);
    initTableFilter();

    return {
      setFilterBy,
      clearFilters,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      columns,
      clusterStore,
      settingStore,
      updateFilterByKey,
      startDate,
      endDate,
    };
  },
  async created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    extendedDefaultPredefinedOptions() {
      return extendedDefaultPredefinedOptions;
    },
  },
  methods: {
    rangeOfDatesCallback(date: string): boolean {
      const dateToCheck = new Date(date);
      if (Number.isNaN(dateToCheck.getTime())) return false;

      dateToCheck.setHours(0, 0, 0, 0);
      const back90Days = dateUtil.adjustDateBy(TimeUnit.day, new Date(), -90);
      back90Days.setHours(0, 0, 0, 0);

      const today = new Date();
      today.setHours(23, 59, 59, 999);

      return dateToCheck >= back90Days && dateToCheck <= today;
    },
    updateFilterBy(filterBy: IFilterBy): void {
      this.setFilterBy(filterBy);
    },
    async updateDates(dates: [Date, Date]) {
      this.startDate = dateUtil.convertDateToISO(dates[0]);
      this.endDate = dateUtil.convertDateToISO(dates[1]);
      await this.setFilterBy(this.filterBy);
    },
    async exportTableAsCsv(): Promise<void> {
      if (this.isExportingCSV) return;
      this.isExportingCSV = true;
      try {
        const blob = await auditService.downloadAudits(
          this.filterBy,
          this.startDate,
          this.endDate,
          0,
          AUDIT_LOG_DOWNLOAD_LIMIT,
          "CSV",
        );
        fileDownload(blob, "audit-events.csv");
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError("Failed to download audit logs"));
        this.appStore.setFallback(true);
      } finally {
        this.isExportingCSV = false;
      }
    },
    async exportTableAsJson(): Promise<void> {
      if (this.isExportingJSON) return;
      this.isExportingJSON = true;
      try {
        const blob = await auditService.downloadAudits(
          this.filterBy,
          this.startDate,
          this.endDate,
          0,
          AUDIT_LOG_DOWNLOAD_LIMIT,
          "JSON",
        );
        fileDownload(blob, "audit-events.json");
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError("Failed to download audit logs"));
        this.appStore.setFallback(true);
      } finally {
        this.isExportingJSON = false;
      }
    },
  },
});
</script>
<style lang="scss" scoped>
.predefined-date {
  padding: 10px 20px 10px 0;
  border-bottom: $black-12 solid 1px;
  background-color: white;
  overflow: hidden;
  position: relative;
}
</style>
