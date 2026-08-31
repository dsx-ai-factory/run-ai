<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="assignment-index">
    <legacy-access-rule-transition-modal
      v-if="isTransitionModalOpen"
      :available-roles="roles"
      @close="handleTransitionModalClose"
    />
    <access-rule-deletion-modal
      v-if="isDeleteModalOpen && selectedRows.length > 0"
      :selected-access-rules="selectedRows"
      :deletable-access-rules-ids="deletableAccessRulesIds"
      @cancel="isDeleteModalOpen = false"
      @deletion-success="handleDeletionSuccess"
    />
    <new-access-rule-modal
      v-if="isNewAccessRuleModalOpen"
      @close="closeAccessRuleModal"
      :is-sso="isSSO"
      :roles="roles"
    />
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:banner>
        <legacy-access-rule-banner
          v-if="hasAccessRulesWithLegacyRoles"
          @transition-click="isTransitionModalOpen = true"
        />
      </template>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          :columns="columns"
          primary-btn-label="New Access rule"
          :selected-rows-amount="selectedRowsAmount"
          @filters-changed="setFilterBy"
          @selected-actions-close="resetSelectedRows"
          @create-entity="isNewAccessRuleModalOpen = true"
          @export-csv="exportCsv"
          advanced-filters
          hide-cluster-filter
        >
          <template v-slot:selected-rows-actions>
            <runai-tooltip-wrapper :display-tooltip="isDeleteBtnDisabled" :tooltip-text="deleteBtnTooltip">
              <runai-action-button
                v-permission="{ resourceType: ResourceType.AccessRules, action: Action.Delete }"
                btn-action="delete"
                :disable="isDeleteBtnDisabled"
                @click="isDeleteModalOpen = true"
              >
              </runai-action-button>
            </runai-tooltip-wrapper>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :top-row="lastCreatedEntity"
          :rows="tableData"
          :columns="columns"
          :loading="loadingTableData"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          @update-filters="setFilterBy"
          :get-row-key="getRowKey"
          sticky-columns
          @role-clicked="onRoleClicked"
          is-server-side-pagination
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25, 50, 100]"
          :bordered="false"
          selection="multiple"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedEntity"
              :filter-by="filterBy"
              entity-name="access rule"
              @clear-filters="clearFilters"
              ph-icon-name="runai-outlined-access-rule"
              icon-size="44"
              @create-new="isNewAccessRuleModalOpen = true"
              :disable-create-btn="!canCreateAccessRules"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>
  </section>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";

// models
import {
  accessRulesIndexColumns,
  allAccessRulesColumnsMap,
  type IAccessRuleTable,
} from "@/table-models/access-rule.table-model";
import {
  type AccessRule,
  AccessRulesBatchAction,
  type AccessRulesBatchCreateResponse,
  type RoleV1,
} from "@/swagger-models/authorization-client";
import { ECustomCell } from "@/models/table.model";
import type { ITableColumn } from "@/models/table.model";

// stores
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useRoleStore } from "@/stores/role.store";

// services
import { rolesService } from "@/services/control-plane/rbac/roles.service/roles.service";

// utils
import { alertUtil } from "@/utils/alert.util";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";

import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { useAuthStore } from "@/stores/auth.store";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { useExportFileComposable } from "@/composables/use-export-file-composable/use-export-file-composable";
import { EAdvancedExportEntity } from "@/composables/helpers/use-export-file.mapper";
import type { IPaginationFilter } from "@/models/filter.model";
import { AccessRuleDeletionModal } from "@/components/access-rules/access-rule-deletion-modal/";
import { useClusterStore } from "@/stores/cluster.store";
import { accessRuleUtil } from "@/utils/rbac.util/access-rule.util";
import { BadRequestError } from "@/models/http-response.model";
import NewAccessRuleModal from "@/components/rbac/access-rule/new-access-rule-modal/new-access-rule-modal.vue";
import { LegacyAccessRuleTransitionModal } from "@/components/rbac/access-rule/legacy-access-rule-transition-modal";
import { LegacyAccessRuleBanner } from "@/components/rbac/access-rule/legacy-access-rule-banner";
import { useLegacyAccessRuleStore } from "@/stores/legacy-access-rule.store";

export default defineComponent({
  components: {
    LegacyAccessRuleBanner,
    NewAccessRuleModal,
    LegacyAccessRuleTransitionModal,
    RunaiTableNoData,
    RunaiActionButton,
    RunaiPageActions,
    RunaiTable,
    RunaiTableWrapper,
    AccessRuleDeletionModal,
    RunaiTooltipWrapper,
  },
  data() {
    return {
      appStore: useAppStore(),
      roleStore: useRoleStore(),
      permissionStore: usePermissionStore(),
      authStore: useAuthStore(),
      legacyAccessRuleStore: useLegacyAccessRuleStore(),
      deleting: false as boolean,
      loading: false as boolean,
      isDeleteModalOpen: false as boolean,
      isNewAccessRuleModalOpen: false as boolean,
      isTransitionModalOpen: false as boolean,
      roles: [] as Array<RoleV1>,
      deletableAccessRulesIds: [] as number[],
      lastSysAdminDeletionError: false,
    };
  },
  setup() {
    const permissionStore = usePermissionStore();
    const clusterStore = useClusterStore();
    const hasRolesReadPermission = computed(() => {
      return permissionStore.hasPermission(ResourceType.Roles, Action.Read);
    });
    const getCustomCellConfig = (column: ITableColumn) => {
      return {
        ...column.customCellConfig,
        [ECustomCell.ROLE_WITH_BADGE_COL]: {
          ...column.customCellConfig?.[ECustomCell.ROLE_WITH_BADGE_COL],
          isTextOnly: false,
        },
      };
    };
    const columns = computed(() => {
      return accessRulesIndexColumns.map((column: ITableColumn) => {
        if (column.name === allAccessRulesColumnsMap.role.name && hasRolesReadPermission) {
          return {
            ...column,
            customCellConfig: getCustomCellConfig(column),
          };
        }
        return column;
      });
    });

    const clusterList = computed(() => clusterStore.clusterList);
    const accessRulesRows = computed(() => {
      return accessRuleUtil.addStatusExplanationToAccessRules(tableData.value, clusterList.value);
    });
    const {
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      initTableFilter,
      updateFilterByKey,
      removeRow,
      selectedRows,
      resetSelectedRows,
      selectedRow,
      selectedRowsAmount,
    } = useTableDataAdvancedFilters<IAccessRuleTable>(
      accessRuleService.list,
      columns,
      EAdvancedIndexPages.ACCESS_RULE,
      (id: string) => accessRuleService.getAccessRule(parseInt(id)),
    );
    initTableFilter();

    const { exportCsv } = useExportFileComposable<IAccessRuleTable>(
      filterBy,
      columns,
      EAdvancedExportEntity.AccessRules,
      accessRuleService.getAccessRulesV2,
      (filters: IPaginationFilter) => accessRuleService.getAccessRulesCount(filters.filterBy, filters.search),
    );

    return {
      setFilterBy,
      clearFilters,
      tableData: accessRulesRows,
      filterBy,
      loadingTableData,
      loadingError,
      columns,
      updateFilterByKey,
      hasRolesReadPermission,
      selectedRows,
      resetSelectedRows,
      selectedRow,
      lastCreatedEntity,
      removeRow,
      refreshList,
      selectedRowsAmount,
      exportCsv,
    };
  },
  async created() {
    // setPageLoading must run even if loadLegacyAccessRules rejects
    try {
      if (this.canDeleteAccessRules) {
        await this.legacyAccessRuleStore.loadLegacyAccessRules();
      }
    } finally {
      this.appStore.setPageLoading(false);
    }
    if (this.hasRolesReadPermission) {
      await this.roleStore.loadRoles();
    }

    this.roles = this.getPermittedRoles();
  },
  computed: {
    hasAccessRulesWithLegacyRoles(): boolean {
      return this.legacyAccessRuleStore.hasLegacyAccessRules && this.canDeleteAccessRules;
    },
    canDeleteAccessRules(): boolean {
      return this.permissionStore.hasPermission(ResourceType.AccessRules, Action.Delete);
    },
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    isSSO(): boolean {
      return this.authStore.isSSO;
    },
    canCreateAccessRules(): boolean {
      return this.permissionStore.hasPermission(ResourceType.AccessRules, Action.Create);
    },
    selectedAccessRulesIds(): Array<number> {
      return this.selectedRows.map((row: AccessRule) => row.id);
    },
    isDeleteBtnDisabled(): boolean {
      return this.deletableAccessRulesIds.length === 0 || this.loading || this.lastSysAdminDeletionError;
    },
    deleteBtnTooltip(): string {
      if (this.loading) return "";
      if (this.lastSysAdminDeletionError) {
        return "To avoid losing access, at least one system-wide administrator access rule must be kept";
      }
      const suffix = this.selectedRows.length > 1 ? "s" : "";
      if (this.isDeleteBtnDisabled) {
        return `You aren't authorized to delete the selected access rule${suffix}`;
      }
      return `Delete access rule${suffix}`;
    },
  },
  methods: {
    getRowKey(row: AccessRule): number {
      return row.id;
    },
    handleDeletionSuccess(subjectIds: number[]): void {
      if (!this.selectedRow) return;

      if (subjectIds.length === 1) {
        this.$q.notify(alertUtil.getSuccess(`Access rule for ${this.selectedRow.subjectId} deleted`));
        this.removeRow(this.selectedRow);
      } else {
        this.$q.notify(alertUtil.getSuccess(`${subjectIds.length} access rules deleted`));
        this.refreshList(true);
      }
      this.isDeleteModalOpen = false;
      this.resetSelectedRows();
    },
    onRoleClicked(accessRule: AccessRule): void {
      this.$router.push({ name: ROLE_ROUTE_NAMES.ACCESS_RULES_ROLE_VIEW, params: { id: accessRule.roleId } });
    },
    async handleTransitionModalClose(): Promise<void> {
      this.isTransitionModalOpen = false;
      await this.legacyAccessRuleStore.loadLegacyAccessRules();
    },
    getPermittedRoles(): Array<RoleV1> {
      return this.roleStore.roleList.filter(
        (role: RoleV1) =>
          role.effectiveEnabled && rolesService.isPermittedRole(role, this.permissionStore.permissionsList),
      );
    },
    closeAccessRuleModal(createdAccessRules?: AccessRulesBatchCreateResponse): void {
      this.isNewAccessRuleModalOpen = false;
      if (!createdAccessRules) return;

      this.notifyCreatedAccessRules(createdAccessRules);
      void this.refreshList(true, createdAccessRules.succeeded?.[0]?.id?.toString());
    },
    notifyCreatedAccessRules(createdAccessRules: AccessRulesBatchCreateResponse): void {
      const succeeded = createdAccessRules.succeeded || [];
      const totalCount = succeeded.length + (createdAccessRules.failed?.length || 0);

      let message: string;
      if (succeeded.length === 1 && totalCount === 1) {
        message = `Access rule for ${succeeded[0].subjectId} created`;
      } else if (succeeded.length === totalCount) {
        message = `${succeeded.length} access rule(s) created`;
      } else {
        message = `${succeeded.length}/${totalCount} access rule(s) created`;
      }

      this.$q.notify(alertUtil.getSuccess(message));
    },
    async validateDeletableAccessRulesIds() {
      try {
        this.loading = true;
        const validationResponse = await accessRuleService.accessRulesBatch({
          action: AccessRulesBatchAction.ValidateDelete,
          ids: this.selectedAccessRulesIds.map(String),
        });
        if (validationResponse.succeeded) {
          this.deletableAccessRulesIds = validationResponse.succeeded.map((id: string) => parseInt(id));
        } else {
          this.deletableAccessRulesIds = [];
        }

        this.lastSysAdminDeletionError =
          Array.isArray(validationResponse.failed) &&
          validationResponse.failed.some((failure) => failure.code === BadRequestError.statusCode);
      } catch (error) {
        this.deletableAccessRulesIds = [];
      } finally {
        this.loading = false;
      }
    },
  },
  watch: {
    selectedRows: {
      handler: async function () {
        if (this.selectedRows.length === 0) {
          this.deletableAccessRulesIds = [];
          return;
        }
        await this.validateDeletableAccessRulesIds();
      },
    },
  },
});
</script>
