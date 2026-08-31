<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card-section>
    <div class="text-weight-medium q-mb-sm">Access</div>
    <div>You have access to the following scopes based on your assigned roles</div>
  </q-card-section>
  <div class="row q-px-md q-pb-md">
    <q-input
      class="col-4"
      :model-value="filterBy.searchTerm"
      @update:model-value="updateSearch"
      clearable
      outlined
      dense
      clear-icon="ph-bold ph-x"
      debounce="300"
      placeholder="Search"
    >
      <template v-slot:prepend>
        <q-icon name="ph-bold ph-magnifying-glass" size="20px" />
      </template>
    </q-input>
  </div>
  <q-separator />
  <runai-table
    :loading="loading"
    :filter-by="filterBy"
    :columns="columns"
    :rows="accessRules"
    disable-selection
    :bordered="false"
    no-shadow
    @update-filters="updateFilter"
    :rows-per-page-options="[5, 7, 10, 15]"
  >
    <template #no-data>
      <runai-table-no-data
        v-if="!loading"
        :filter-by="filterBy"
        entity-name="access rule"
        :show-create-btn="false"
        :show-filter-icon-and-button="false"
      >
      </runai-table-no-data>
    </template>
  </runai-table>
</template>
<script lang="ts">
import { defineComponent } from "vue";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import type { MeAccessRule } from "@/swagger-models/authorization-client";
import { DEFAULT_PAGE_SIZE, type IFilterBy, type IPaginationFilter } from "@/models/filter.model";
import { RunaiTable } from "@/components/common";
import type { ITableColumn } from "@/models/table.model";
import { allMeAccessRulesColumnsMap } from "@/table-models/access-rule.table-model";
import { filterService } from "@/services/filter.service/filter.service";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { alertUtil } from "@/utils/alert.util";

export default defineComponent({
  name: "user-profile-access-rules-table",
  components: { RunaiTableNoData, RunaiTable },
  emits: {
    "access-rules-fetched": (accessRules: MeAccessRule[]) => accessRules,
  },
  data() {
    return {
      loading: true as boolean,
      isInitialLoad: true as boolean,
      accessRules: [] as MeAccessRule[],
      filterBy: {
        searchTerm: "",
        rowsNumber: DEFAULT_PAGE_SIZE,
      } as IFilterBy,
    };
  },
  computed: {
    columns(): ITableColumn[] {
      return allMeAccessRulesColumnsMap;
    },
  },
  methods: {
    updateSearch(searchTerm: string | null | number): void {
      this.filterBy.searchTerm = searchTerm as string;
      this.getAccessRules();
    },
    updateFilter(filterBy: IFilterBy): void {
      this.filterBy = filterBy;
      this.getAccessRules();
    },
    async getAccessRules(): Promise<void> {
      const filter = filterService.mapColumnsFilterToFilterParams(this.filterBy, undefined, true);
      try {
        this.loading = true;
        await this.getMeAccessRules(filter);
        await this.countMeAccessRules(filter);
        if (this.isInitialLoad) {
          this.isInitialLoad = false;
          this.$emit("access-rules-fetched", this.accessRules);
        }
      } catch (e) {
        this.handleLoadAccessRulesError(e);
      } finally {
        this.loading = false;
      }
    },
    handleLoadAccessRulesError(error: unknown): void {
      console.error("Failed to load access rules", error);
      this.$q.notify(alertUtil.getError("Failed to load access rules"));
    },
    async getMeAccessRules(filter: IPaginationFilter): Promise<void> {
      this.accessRules = await accessRuleService.getMeAccessRules(filter.limit, filter.offset, filter.search);
    },
    async countMeAccessRules(filter: IPaginationFilter): Promise<void> {
      this.filterBy.rowsNumber = await accessRuleService.countMeAccessRules(filter.search);
    },
  },
});
</script>

<style scoped lang="scss"></style>
