<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog :model-value="modelValue" @close="$emit('close')" size="dynamic" no-padding>
    <template #header>{{ modalOptions.header }}</template>
    <template #body>
      <runai-list-modal-search v-model="tableFilter.searchTerm" placeholder="Search node pools" />
      <q-separator />
      <runai-table
        :rows="nodePools"
        :columns="columns"
        :loading="modalOptions.loading || false"
        :filter-by="tableFilter"
        :get-row-key="getRowKey"
        @update-filters="updateFilter"
        modal-view
        disable-selection
      >
        <template #no-data>
          <runai-table-no-data
            v-if="!modalOptions.loading"
            :filter-by="tableFilter"
            entity-name="node pool"
            :show-create-btn="false"
            :show-filter-icon-and-button="false"
          />
        </template>
      </runai-table>
    </template>
    <template #footer-left v-if="modalOptions.showRedirectBtn">
      <q-btn flat :ripple="false" color="primary" class="btn-link" :label="redirectLabel" @click="$emit('redirect')" />
    </template>
    <template #footer>
      <q-btn aid="close-node-pools-modal-btn" label="Close" color="primary" @click="$emit('close')" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiListModalSearch } from "@/components/common";

// services
import { filterService } from "@/services/filter.service/filter.service";

// models
import type { IFilterBy } from "@/models/filter.model";
import type { INodePoolModalOptions, INodePoolsResourcesRow } from "@/models/project.model";
import {
  allOrgUnitNodePoolsModalColumnsMap,
  orgUnitNodePoolsModalTableColumns,
  orgUnitNodePoolsModalDependentColumns,
} from "@/table-models/node-pool.table-model";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";
import type { ITableColumn } from "@/models/table.model";
import { useSettingStore } from "@/stores/setting.store";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { EQuotaEntity } from "@/models/resource.model";
import type { QuotaStatusResource, Resources } from "@/swagger-models/org-unit-service-client";
import { resourceUtil } from "@/utils/resource.util";
import { nodePoolUtil } from "@/utils/node-pool.util/node-pool.util";
import { orgUnitUtil } from "@/utils/org-unit.util";

export default defineComponent({
  components: { RunaiTableNoData, RunaiTable, RunaiBaseDialog, RunaiListModalSearch },
  emits: ["close", "redirect"],
  props: {
    modalOptions: {
      type: Object as PropType<INodePoolModalOptions>,
      required: true,
    },
  },
  data() {
    return {
      modelValue: true as boolean,
      tableFilter: {} as IFilterBy,
      settingStore: useSettingStore(),
    };
  },
  created() {
    this.loadFilters();
  },
  computed: {
    resources(): Resources[] {
      return orgUnitUtil.initResourcesOverQuotaWeight(this.modalOptions.resources, this.isOverQuotaPriorityEnabled);
    },
    isProjectEntity(): boolean {
      return this.modalOptions.entityType === EQuotaEntity.project;
    },
    columns(): Array<ITableColumn> {
      return orgUnitNodePoolsModalTableColumns
        .filter((col: ITableColumn) => {
          if (orgUnitNodePoolsModalDependentColumns.cpu.has(col.name)) {
            return this.settingStore.isCPUResourcesQuotaEnabled;
          }
          if (orgUnitNodePoolsModalDependentColumns.priorityOrder.has(col.name)) {
            return this.nodePools.length > 1;
          }
          if (orgUnitNodePoolsModalDependentColumns.project.has(col.name)) {
            return this.isProjectEntity;
          }
          if (orgUnitNodePoolsModalDependentColumns.department.has(col.name)) {
            return !this.isProjectEntity;
          }
          return true;
        })
        .map((col: ITableColumn) => {
          if (col.name === allOrgUnitNodePoolsModalColumnsMap.overQuotaWeight.name) {
            return {
              ...col,
              label: this.isOverQuotaPriorityEnabled ? "Over quota weight" : "Over quota",
            };
          }
          return col;
        });
    },
    isOverQuotaPriorityEnabled(): boolean {
      return this.settingStore.isOverQuotaPriorityEnabled;
    },
    nodePools(): INodePoolsResourcesRow[] {
      const allocateByNodePoolNameMap = this.modalOptions.nodePoolQuotaStatuses.reduce(
        (acc, quotaStatus) => {
          if (quotaStatus.nodePoolName && quotaStatus.allocated) {
            acc[quotaStatus.nodePoolName] = quotaStatus.allocated;
          }
          return acc;
        },
        {} as Record<string, QuotaStatusResource>,
      );

      return (
        this.resources.map((resource: Resources) => {
          const findQuotaStatus = (key: keyof QuotaStatusResource) => {
            const nodePoolName = resource?.nodePool?.name;
            if (!nodePoolName) return 0;
            return allocateByNodePoolNameMap[nodePoolName]?.[key] || 0;
          };

          const nodePoolRow: INodePoolsResourcesRow = {
            ...resource,
            allocatedGpus: findQuotaStatus("gpu"),
            allocatedCpu: findQuotaStatus("cpu"),
            allocatedMemory: findQuotaStatus("memory"),
          };

          if (this.modalOptions.defaultNodePools) {
            nodePoolRow.priorityOrder = nodePoolUtil.getNodePoolPriorityByName(
              this.modalOptions.defaultNodePools || [],
              resource?.nodePool?.name || "",
            );
          }

          nodePoolRow.overQuotaWeightLabel = resourceUtil.getNodePoolResourceOverQuotaWeightDisplayedValue(
            resource.gpu,
            this.isOverQuotaPriorityEnabled,
          );

          return nodePoolRow;
        }) || []
      );
    },
    redirectLabel(): string {
      return `Go to ${this.modalOptions.entityType}s to view more details`;
    },
  },
  methods: {
    loadFilters(): void {
      const defaultFilters: IFilterBy = filterService.getDefaultFilters(
        allOrgUnitNodePoolsModalColumnsMap.priorityOrder.name,
        orgUnitNodePoolsModalTableColumns,
      );
      this.tableFilter = filterService.getEmptyFilterByModel(defaultFilters);
    },

    updateFilter(filter: IFilterBy): void {
      this.tableFilter = filter;
    },
    getRowKey(row: INodePoolsResourcesRow): string {
      return row.nodePool?.id || "";
    },
  },
});
</script>

<style scoped></style>
