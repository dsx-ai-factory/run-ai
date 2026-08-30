import { computed, type Ref } from "vue";
//model
import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import type { EAdvancedExportEntity } from "@/composables/helpers/use-export-file.mapper";
import { type GetRecordsFunction, useFileExportConfigMapper } from "@/composables/helpers/use-export-file.mapper";
//svc
import { filterService } from "@/services/filter.service/filter.service";
import { fileService } from "@/services/file.service";
//store
import { useClusterStore } from "@/stores/cluster.store";
import { ErrorAlert } from "@/utils/error-alert.util";
import { Notify, useQuasar } from "quasar";
import { alertUtil } from "@/utils/alert.util";

export function useExportFileComposable<T>(
  filterBy: Ref<IFilterBy>,
  columns: Ref<ITableColumn[]>,
  entity: EAdvancedExportEntity,
  getRecords: GetRecordsFunction<T>,
  getCount: (filter: IPaginationFilter) => Promise<number>,
  filterByMapper?: (filterBy: IFilterBy) => IPaginationFilter,
) {
  const exportConfig = useFileExportConfigMapper[entity];
  const clusterStore = useClusterStore();
  const q = useQuasar();

  const clusterList = computed(() => clusterStore.clusterList);
  const filteredColumns = computed(() =>
    columns.value.filter((col: ITableColumn) => filterBy.value.displayedColumns?.includes(col.name)),
  );
  const filters = computed(() => {
    return filterService.mapColumnsFilterToFilterParams(
      filterBy.value,
      exportConfig.getRecordsValidFields,
      exportConfig.isAdvancedSearch,
    );
  });

  const exportCsv = async () => {
    const dismiss = q.notify(alertUtil.getLoading(`Downloading`));

    try {
      await fileService.csvExporter(
        getRecords,
        getCount,
        filters.value,
        filteredColumns.value,
        exportConfig.fileName,
        clusterList.value,
        filterByMapper,
      );
    } catch (error: unknown) {
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedToExportMessage(entity),
      });
      Notify.create(errorAlert.getNotification(error));
    } finally {
      dismiss();
    }
  };

  return { exportCsv };
}
