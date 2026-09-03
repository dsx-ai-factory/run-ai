<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:banner>
        <section class="notification-alert-container" v-if="isSSOEnabled">
          <notification-alert-bar
            :variant="ENotificationBarVariant.Info"
            message="*SSO (Single Sign-On) users are managed by the identity provider and will appear once they've signed in to Run:ai.* In the meantime, you can give them access via"
          >
            <template v-slot:inline-actions>
              <a href="javascript:void(0)" @click="goToAccessRules()">Access rules.</a>
            </template>
          </notification-alert-bar>
        </section>
      </template>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          :columns="columns"
          :primary-btn-label="newUserLabel"
          :selected-rows-amount="selectedRowsAmount"
          @create-entity="openCreateModal"
          @filters-changed="updateFilterBy"
          @selected-actions-close="resetSelectedRows"
          @export-csv="exportCsv"
          advanced-filters
          hide-cluster-filter
        >
          <template v-slot:selected-rows-actions>
            <runai-action-button
              v-if="selectedUser.isLocal"
              v-permission="{ resourceType: ResourceType.Users, action: Action.Update }"
              btn-action="edit"
              @click="openEditModal"
            />
            <runai-action-button
              v-permission="{ resourceType: ResourceType.AccessRules, action: Action.Read }"
              btn-action="assign"
              @click="openAccessRuleModal"
            />
            <runai-action-button
              v-if="selectedUser.isLocal"
              v-permission="{ resourceType: ResourceType.Users, action: Action.Create }"
              btn-action="reset"
              @click="openCreateModal"
            />
            <runai-action-button
              :disable="isSelectedUserIsMe"
              v-if="selectedUser.isLocal"
              v-permission="{ resourceType: ResourceType.Users, action: Action.Delete }"
              btn-action="delete"
              @click="openDeleteModal"
            >
              <q-tooltip v-if="isSelectedUserIsMe" anchor="bottom end" max-width="300px"
                >The selected user is the one you're currently signed in as and therefore can't be deleted.</q-tooltip
              >
            </runai-action-button>
            <runai-action-button
              :disable="isSelectedUserIsMe"
              v-if="!selectedUser.isLocal"
              v-permission="{ resourceType: ResourceType.Users, action: Action.Delete }"
              btn-action="remove"
              @click="openDeleteSSoModal"
            >
              <q-tooltip v-if="isSelectedUserIsMe" anchor="bottom end" max-width="300px"
                >The selected user is the one you're currently signed in as and therefore can't be removed.</q-tooltip
              >
            </runai-action-button>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="tableData"
          :columns="columns"
          :loading="loadingTableData"
          :get-row-key="getRowKey"
          :filter-by="filterBy"
          v-model:selected="selectedUsers"
          :top-row="lastCreatedUser"
          @update-filters="updateFilterBy"
          @update:selected="updateSelectedUsers"
          @access-rules-clicked="displayAccessRuleTableModal"
          sticky-columns
          :bordered="false"
          is-server-side-pagination
          :rows-per-page-options="[5, 7, 10, 15, 20, 25, 50, 100]"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedUser"
              :filter-by="filterBy"
              entity-name="user"
              ph-icon-name="ph-bold ph-user"
              icon-size="44"
              @create-new="openCreateModal"
              @clear-filters="clearFilters"
              :disable-create-btn="!canCreateUsers"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedUser"
      :entity-name="selectedUser.username"
      entity-type="user"
      @cancel="isDeleteModalOpen = false"
      @delete="removeUser"
      :deleting="deleting"
    >
      <template v-slot:description>
        <div>The user will be deleted along with its access rules and service accounts.</div>
        <div>This can't be undone.</div>
      </template>
    </runai-delete-modal>

    <runai-delete-modal
      v-if="isDeleteSSOModalOpen && selectedUser"
      :entity-name="selectedUser.username"
      entity-type="user"
      @cancel="isDeleteSSOModalOpen = false"
      @delete="removeUser"
      :deleting="deleting"
      :title="customTitleDeleteSSO"
      delete-button-label="remove"
      header-title="Remove"
    >
      <template v-slot:description>
        <div>The user will be removed along with their access rules and service accounts.</div>
      </template>
    </runai-delete-modal>

    <user-creation-modal
      v-if="isCreationModalOpen"
      :action-type="modalActionType"
      :model-value="isCreationModalOpen"
      :selected-user="selectedUsers[0]"
      @close="isCreationModalOpen = false"
      @load-users="loadUsers"
      :filter-by="filterBy"
    />

    <user-edit-modal
      v-if="isEditModalOpen && selectedUser"
      :model-value="isEditModalOpen"
      :user="selectedUser"
      @close="isEditModalOpen = false"
      @saved="onUserSaved"
    />

    <access-rule-table-modal
      v-if="isAccessRuleTableModalOpen"
      @close="closeAccessRuleTableModal"
      :modal-options="accessRuleTableModalOptions"
    />
    <access-rule-management-modal
      v-if="isAccessRuleManagementModalOpen"
      @close="isAccessRuleManagementModalOpen = false"
      :modal-options="accessRuleManagementModalOptions"
    />
  </section>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";
import { useAppStore } from "@/stores/app.store";
import type { ITableColumn } from "@/models/table.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";
import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
// Switching to swagger `User` type for selected rows to match modal prop typing
import { allUsersColumnsMap, usersDependentColumns, usersIndexColumns } from "@/table-models/user.table-model";
import { UserCreationModal } from "@/components/user/user-creation-modal";
import { UserEditModal } from "@/components/user/user-edit-modal";
import type { TActionTypes } from "@/components/user/action-type.model";
import { ActionTypeEnum } from "@/components/user/action-type.model";
import { RunaiDeleteModal, RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import type { IAccessRuleTableModalOptions, IAccessRuleManagementModalOptions } from "@/models/access-rule.model";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
import { Action, ResourceType, SubjectType } from "@/swagger-models/authorization-client";
import {
  subjectRolesMiniTableColumns,
  subjectRolesMiniTableColumnsWithSSO,
} from "@/table-models/access-rule.table-model";
import { AccessRuleTableModal } from "@/components/rbac/access-rule/access-rule-table-modal";
import { useSettingStore } from "@/stores/setting.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { NotificationAlertBar, ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { useAuthStore } from "@/stores/auth.store";
// removed legacy local fetch helpers in favor of server-side pagination in modal
import userService from "@/services/control-plane/user.service/user.service";
import type { User } from "@/swagger-models/identity-manager-client";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { useExportFileComposable } from "@/composables/use-export-file-composable/use-export-file-composable";
import { EAdvancedExportEntity } from "@/composables/helpers/use-export-file.mapper";
import { ErrorAlert } from "@/utils/error-alert.util";
import { AccessRuleManagementModal } from "@/components/rbac/access-rule/access-rule-management-modal";

export default defineComponent({
  components: {
    AccessRuleManagementModal,
    NotificationAlertBar,
    AccessRuleTableModal,
    RunaiActionButton,
    RunaiPageActions,
    RunaiTableWrapper,
    UserCreationModal,
    UserEditModal,
    RunaiDeleteModal,
    RunaiTable,
    RunaiTableNoData,
  },

  data() {
    return {
      appStore: useAppStore(),
      permissionStore: usePermissionStore(),
      authStore: useAuthStore(),
      deleting: false as boolean,
      selectedUsers: [] as Array<User>,
      isDeleteModalOpen: false as boolean,
      isDeleteSSOModalOpen: false as boolean,
      isCreationModalOpen: false as boolean,
      isEditModalOpen: false as boolean,
      isAccessRuleManagementModalOpen: false as boolean,
      isAccessRuleTableModalOpen: false as boolean,
      accessRuleTableModalOptions: {
        header: "",
        columns: subjectRolesMiniTableColumns,
      } as IAccessRuleTableModalOptions,
      accessRuleManagementModalOptions: {
        page: EAccessRuleModalPage.User,
        subjectId: "",
        subjectType: SubjectType.User,
      } as IAccessRuleManagementModalOptions,
    };
  },

  setup() {
    const settingStore = useSettingStore();
    const authStore = useAuthStore();
    const permissionStore = usePermissionStore();

    const getUsers = async (filterBy: IFilterBy): Promise<User[]> => {
      const userResponse: Array<User> = await userService.listUsers(filterBy, !authStore.isSSO);
      const countUsers: number = await userService.countUsers(filterBy, !authStore.isSSO);
      updateFilterByKey("rowsNumber", countUsers);
      return userResponse;
    };

    const getUserById = async (userId: string): Promise<User> => {
      return await userService.getUserById(userId);
    };

    const columns = computed<ITableColumn[]>(() => {
      return usersIndexColumns.filter((col: ITableColumn) => {
        if (usersDependentColumns.accessRules.has(col.name)) {
          return permissionStore.hasPermission(ResourceType.Users, Action.Read);
        }
        return true;
      });
    });

    const {
      setFilterBy,
      clearFilters,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      initTableFilter,
      updateFilterByKey,
      refreshList,
      lastCreatedEntity,
      clearLastCreated,
    } = useTableDataAdvancedFilters<User>(getUsers, columns, EAdvancedIndexPages.USER, getUserById);
    initTableFilter();

    const { exportCsv } = useExportFileComposable<User>(
      filterBy,
      columns,
      EAdvancedExportEntity.Users,
      userService.getUsers,
      (filter: IPaginationFilter) => userService.countUsersClean(filter.filterBy, filter.search),
    );

    return {
      setFilterBy,
      clearFilters,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      columns,
      settingStore,
      updateFilterByKey,
      refreshList,
      lastCreatedEntity,
      clearLastCreated,
      exportCsv,
    };
  },

  created() {
    this.appStore.setPageLoading(false);
  },

  computed: {
    lastCreatedUser(): User | null {
      return this.lastCreatedEntity;
    },
    ENotificationBarVariant() {
      return ENotificationBarVariant;
    },
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    selectedRowsAmount(): number {
      return this.selectedUsers.length;
    },
    selectedUser(): User {
      return this.selectedUsers[0];
    },
    modalActionType(): TActionTypes {
      return !this.selectedUser ? ActionTypeEnum.CREATE : ActionTypeEnum.RESET;
    },
    canCreateUsers(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Users, Action.Create);
    },
    isSSOEnabled(): boolean {
      return this.authStore.isSSO;
    },
    newUserLabel(): string {
      return this.isSSOEnabled ? "new local user" : "new user";
    },
    isSelectedUserIsMe(): boolean {
      return this.selectedUser?.username === this.authStore.currentUser.username;
    },
    customTitleDeleteSSO(): string {
      return `Remove user: ${this.selectedUser.username}?`;
    },
  },
  methods: {
    getRowKey(row: User): string {
      return row.username;
    },
    getColumns(): Array<ITableColumn> {
      return this.isSSOEnabled
        ? usersIndexColumns
        : usersIndexColumns.filter((col: ITableColumn) => col.name !== allUsersColumnsMap.type.name);
    },

    updateFilterBy(filterBy: IFilterBy): void {
      this.setFilterBy(filterBy);
    },

    updateSelectedUsers(selectedUsers: Array<User>): void {
      this.selectedUsers = selectedUsers;
    },

    resetSelectedRows(): void {
      this.selectedUsers = [];
    },

    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },

    openDeleteSSoModal(): void {
      this.isDeleteSSOModalOpen = true;
    },

    async removeUser(): Promise<void> {
      if (!this.selectedUsers || !this.selectedUser.id) return;
      const deletePhrase = this.selectedUser.isLocal ? "deleted" : "removed";
      try {
        this.deleting = true;
        await userService.remove(this.selectedUser.id);
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`User ${this.selectedUser.username} ${deletePhrase}`));
      } catch (e: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: `User ${this.selectedUser.username} couldn't be ${deletePhrase}`,
          403: `You're not authorized to delete the selected user`,
        });
        this.$q.notify(errorAlert.getNotification(e));
      } finally {
        this.resetSelectedRows();
        this.clearLastCreated();
        this.isDeleteModalOpen = false;
        this.isDeleteSSOModalOpen = false;
        this.deleting = false;
      }
    },

    openCreateModal(): void {
      this.isCreationModalOpen = true;
    },
    openEditModal(): void {
      this.isEditModalOpen = true;
    },
    async onUserSaved(): Promise<void> {
      this.isEditModalOpen = false;
      await this.refreshList();
      this.selectedUsers = this.selectedUsers.map((u) => this.tableData.find((t) => t.id === u.id) || u);
    },
    loadUsers(newCreatedUser: User): void {
      const userId = newCreatedUser?.id ? newCreatedUser.id : undefined;
      this.refreshList(false, userId);
    },
    async displayAccessRuleTableModal(user: User): Promise<void> {
      if (this.isSSOEnabled) {
        this.accessRuleTableModalOptions.columns = subjectRolesMiniTableColumnsWithSSO;
      } else {
        this.accessRuleTableModalOptions.columns = subjectRolesMiniTableColumns;
      }
      this.accessRuleTableModalOptions.header = `Access Rules for ${user.username}`;
      this.accessRuleTableModalOptions.subjectType = SubjectType.User;
      this.accessRuleTableModalOptions.subjectId = user.username;
      this.isAccessRuleTableModalOpen = true;
    },
    closeAccessRuleTableModal(): void {
      this.isAccessRuleTableModalOpen = false;
      this.accessRuleTableModalOptions.subjectId = undefined;
      this.accessRuleTableModalOptions.subjectType = undefined;
      this.accessRuleTableModalOptions.header = "";
    },
    openAccessRuleModal(): void {
      if (!this.selectedUser) return;
      this.accessRuleManagementModalOptions.subjectId = this.selectedUser.username;
      this.accessRuleManagementModalOptions.userGroups = this.selectedUser.groups || [];
      this.isAccessRuleManagementModalOpen = true;
    },
    async closeAccessRuleManagementModal(): Promise<void> {
      this.isAccessRuleManagementModalOpen = false;
      await this.refreshList();
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));
      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.User, this.tableData, columns);
    },
    goToAccessRules(): void {
      this.$router.push({ name: ROLE_ROUTE_NAMES.ACCESS_RULES_INDEX });
    },
  },
});
</script>
