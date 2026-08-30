<template>
  <section class="project-modal">
    <runai-base-dialog :model-value="modelValue" @close="$emit('close')" :size="modalSize" no-padding>
      <template #header> {{ modalHeader }}</template>

      <template #body>
        <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search projects" />
        <q-separator />
        <runai-table
          disable-selection
          :columns="columns"
          :rows="projects"
          :loading="loading"
          :filter-by="tableFilter"
          flat
          modal-view
          is-server-side-pagination
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25]"
          @update-filters="updateFilterBy"
        >
          <template #no-data>
            <runai-table-no-data
              :filter-by="tableFilter"
              entity-name="project"
              :show-create-btn="false"
              :show-filter-icon-and-button="false"
            />
          </template>
        </runai-table>
      </template>
      <template #footer-left>
        <q-btn
          flat
          :ripple="false"
          color="primary"
          class="btn-link"
          label="Go to projects to view more details"
          @click="redirectToProjects"
        />
      </template>
      <template #footer>
        <q-btn aid="close-projects-modal-btn" label="Close" color="primary" @click="$emit('close')" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
// services
import { filterService } from "@/services/filter.service/filter.service";
// stores
import { useSettingStore } from "@/stores/setting.store";
// models
import type { IProjectsModalOptions } from "@/models/project.model";
import { projectDependentColumns, projectMiniTableColumns } from "@/table-models/project.table-model";
import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import { type ITableColumn } from "@/models/table.model";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { type Project, ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { ErrorAlert } from "@/utils/error-alert.util";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";

export default defineComponent({
  components: {
    RunaiTableNoData,
    RunaiTable,
    RunaiBaseDialog,
    RunaiListModalSearch,
  },
  emits: ["close"],
  props: {
    modalOptions: {
      type: Object as PropType<IProjectsModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      modelValue: true as boolean,
      tableFilter: {} as IFilterBy,
      loading: false as boolean,
      projects: [] as Project[],
    };
  },
  created() {
    this.loadFilters();
  },
  computed: {
    columns(): Array<ITableColumn> {
      return projectMiniTableColumns.filter((col: ITableColumn) => {
        if (projectDependentColumns.cpu.has(col.name)) {
          return this.settingStore.isCPUResourcesQuotaEnabled;
        }

        return true;
      });
    },
    isCPUResourcesQuotaEnabled(): boolean {
      return this.settingStore.isCPUResourcesQuotaEnabled;
    },
    modalSize(): "dynamic" | "width-md" {
      return this.isCPUResourcesQuotaEnabled ? "dynamic" : "width-md";
    },
    modalHeader(): string {
      return `Projects Associated with ${this.modalOptions.entityType} ${this.modalOptions.entityName}`;
    },
  },
  methods: {
    async updateFilterBy(filterBy: IFilterBy): Promise<void> {
      this.tableFilter = filterBy;
      try {
        this.loading = true;
        await this.loadProjectsCount();
        await this.loadProjects();
      } catch (error: unknown) {
        this.handleLoadProjectsError(error);
      } finally {
        this.loading = false;
      }
    },
    loadFilters(): void {
      const defaultFilters = filterService.getEmptyFilterByModel({
        sortBy: ProjectFilterSortFields.Name,
        rowsPerPage: 10,
      });
      this.updateFilterBy(defaultFilters);
    },
    async loadProjects(): Promise<void> {
      try {
        this.loading = true;
        this.projects = await this.getProjects();
      } catch (e) {
        this.handleLoadProjectsError(e);
      } finally {
        this.loading = false;
      }
    },
    mapFiltersToParams(): IPaginationFilter {
      return filterService.mapColumnsFilterToFilterParams(
        this.tableFilter,
        Object.values(ProjectFilterSortFields),
        true,
      );
    },
    getFilters(): IPaginationFilter {
      const paginationFilters = this.mapFiltersToParams();
      const filterBy = [
        `${filterUtil.getEqualsFilterString(ProjectFilterSortFields.ClusterId, this.modalOptions.clusterId)}`,
        `${filterUtil.getEqualsFilterString(ProjectFilterSortFields.ParentId, this.modalOptions.parentId)}`,
      ];
      return {
        ...paginationFilters,
        filterBy,
      };
    },
    async getProjects(): Promise<Project[]> {
      const filters = this.getFilters();
      return await orgUnitService.getProjects(
        filters.filterBy,
        filters.sortBy as ProjectFilterSortFields,
        filters.sortOrder,
        filters.offset,
        filters.limit,
        filters.search,
      );
    },
    async loadProjectsCount(): Promise<void> {
      this.tableFilter.rowsNumber = await orgUnitService.countProjects(this.getFilters());
    },
    handleLoadProjectsError(error: unknown): void {
      console.error(error);
      const errorAlert = new ErrorAlert({
        generalMessage: ErrorAlert.failedLoadingMessage("projects"),
      });
      this.$q.notify(errorAlert.getNotification(error));
    },
    redirectToProjects(): void {
      this.$router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
        query: {
          [ProjectFilterSortFields.ParentName]: this.modalOptions.entityName,
          [ProjectFilterSortFields.ClusterId]: this.modalOptions.clusterId,
        },
      });
    },
  },
});
</script>
