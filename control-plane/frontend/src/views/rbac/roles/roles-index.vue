<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          :columns="columns"
          :show-primary-btn="false"
          :selected-rows-amount="selectedRowsAmount"
          @filters-changed="updateFilterBy"
          @selected-actions-close="resetSelectedRows"
          @export-csv="exportTableAsCsv"
          hide-cluster-filter
        >
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="roles"
          :columns="columns"
          :get-row-key="getRowKey"
          :filter-by="filterBy"
          :loading="loading"
          @update-filters="updateFilterBy"
          sticky-columns
          @role-clicked="onRoleClicked"
          disable-selection
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading"
              :filter-by="filterBy"
              entity-name="roles"
              @clear-filters="clearFilters"
              :show-create-btn="false"
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// enums
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";

// store
import { useAppStore } from "@/stores/app.store";
import { useRoleStore } from "@/stores/role.store";

import type { ITableColumn } from "@/models/table.model";
// models
import { ETableExportCsvFilesNames, ETableFilters } from "@/models/table.model";
import { allRoleColumnsMap, roleIndexColumns } from "@/table-models/roles.table-model";
import type { IFilterBy } from "@/models/filter.model";

// services
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";
import { filterService } from "@/services/filter.service/filter.service";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";
import { usePermissionStore } from "@/stores/permissions.store";
import type { Role } from "@/swagger-models/authorization-client";

export default defineComponent({
  components: {
    RunaiTable,
    RunaiTableNoData,
    RunaiPageActions,
    RunaiTableWrapper,
  },
  data() {
    return {
      appStore: useAppStore(),
      roleStore: useRoleStore(),
      permissionStore: usePermissionStore(),
      columns: roleIndexColumns as Array<ITableColumn>,
      filterBy: {} as IFilterBy,
      loading: false as boolean,
      loadingError: false as boolean,
      selectedRoles: [] as Array<Role>,
      isAccessRuleTableModalOpen: false as boolean,
    };
  },
  async created() {
    this.appStore.setPageLoading(false);
    this.loadFilters();
    await this.loadRoles();
  },
  computed: {
    roles(): Array<Role> {
      return this.roleStore.roleList;
    },
    selectedRowsAmount(): number {
      return this.selectedRoles.length;
    },
  },
  methods: {
    getRowKey(row: Role): number {
      return row.id;
    },
    async loadRoles(): Promise<void> {
      try {
        this.loading = true;
        await this.roleStore.loadRoles(this.filterBy);
      } catch (e) {
        this.loadingError = true;
        this.$q.notify(alertUtil.getError("Failed to load roles"));
        console.error(e, "Failed to load roles");
      } finally {
        this.loading = false;
      }
    },
    clearFilters(): void {
      this.updateFilterBy({
        ...this.filterBy,
        columnFilters: [],
        searchTerm: "",
      });
    },
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(allRoleColumnsMap.role.name, this.columns, true);

      const searchParamsFilters: IFilterBy = filterService.loadFilters(
        window.location,
        ETableFilters.ROLE,
        defaultFilters,
      );
      this.updateFilterBy(searchParamsFilters, null, false);
    },
    updateFilterBy(filterBy: IFilterBy, keyChanged: null | string = null, forceLoad = true): void {
      this.filterBy = filterBy;
      filterService.saveFilters(ETableFilters.ROLE, filterBy);
      if (!forceLoad || keyChanged === "displayedColumns") return;

      this.loadRoles();
    },
    resetSelectedRows(): void {
      this.selectedRoles = [];
    },
    onRoleClicked(role: Role): void {
      this.$router.push({ name: ROLE_ROUTE_NAMES.ROLE_VIEW, params: { id: role.id } });
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));

      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.Role, this.roles, columns);
    },
  },
});
</script>
