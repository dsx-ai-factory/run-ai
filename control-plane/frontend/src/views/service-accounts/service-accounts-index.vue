<template>
  <section class="service-accounts-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :columns="columns"
          :filters="filterBy"
          primary-btn-label="NEW SERVICE ACCOUNT"
          primary-tooltip-text="You don't have permission to create a service account."
          @filters-changed="setFilterBy"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          @create-entity="openCreateModal"
          @export-csv="exportTableAsCsv"
          hide-cluster-filter
        >
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-action-button
                v-permission="{ resourceType: ResourceType.AccessRules, action: Action.Read }"
                btn-action="assign"
                @click="openAccessRuleModal"
              />
              <runai-action-button
                v-permission="{ resourceType: [ResourceType.ServiceAccount, ResourceType.Apps], action: Action.Create }"
                btn-action="regenerate"
                aid="regenerate-service-account-btn"
                @click="openCreateModal"
              />
              <runai-tooltip-wrapper
                :display-tooltip="!canDeleteServiceAccount"
                tooltip-text="You're not authorized to delete the selected service account."
              >
                <runai-action-button
                  v-permission="{
                    resourceType: [ResourceType.ServiceAccount, ResourceType.Apps],
                    action: Action.Delete,
                  }"
                  btn-action="delete"
                  aid="delete-service-account-btn"
                  @click="openDeleteModal"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="tableData"
          :columns="columns"
          :get-row-key="getRowKey"
          :filter-by="filterBy"
          :loading="loadingTableData"
          :top-row="lastCreatedEntity"
          v-model:selected="selectedRows"
          @access-rules-clicked="displayAccessRuleTableModal"
          @update-filters="setFilterBy"
          :bordered="false"
          sticky-columns
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedEntity"
              :show-error="loadingError"
              :filter-by="filterBy"
              entity-name="service account"
              ph-icon-name="ph-bold ph-squares-four"
              icon-size="44"
              @create-new="openCreateModal"
              @clear-filters="clearFilters"
            >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedServiceAccount"
      :entity-name="selectedServiceAccount.name"
      entity-type="service account"
      @cancel="isDeleteModalOpen = false"
      @delete="removeServiceAccount"
      :deleting="deleting"
    >
      <template v-slot:description>
        <div>The service account will be deleted along with its access rules.</div>
        <div>This can't be undone.</div>
      </template>
    </runai-delete-modal>

    <credential-creation-modal
      v-if="isCreationModalOpen"
      entity-type="service-account"
      :action-type="modalActionType"
      :model-value="isCreationModalOpen"
      :credential="selectedServiceAccount"
      @close="isCreationModalOpen = false"
      @load-credential="refreshWithLastCreated"
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
import { defineComponent } from "vue";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { CredentialCreationModal } from "@/components/common/credential-creation-modal";
import { RunaiDeleteModal, RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// stores
import { useAppStore } from "@/stores/app.store";

// models
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";

// services
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";

// constants
import { serviceAccountsIndexColumns } from "@/table-models/service-account.table-model";

// types
import type { TActionTypes } from "@/components/common/credential-creation-modal";
import {
  EAccessRuleModalPage,
  type IAccessRuleManagementModalOptions,
  type IAccessRuleRecords,
  type IAccessRuleTableModalOptions,
} from "@/models/access-rule.model";
import { subjectRolesMiniTableColumns } from "@/table-models/access-rule.table-model";
import { useSettingStore } from "@/stores/setting.store";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { AccessRuleTableModal } from "@/components/rbac/access-rule/access-rule-table-modal";
import { Action, ResourceType, SubjectType } from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";
import { useTableData } from "@/composables/use-table-data.composable";
import { EIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { serviceAccountsService } from "@/services/control-plane/service-accounts.service/service-accounts.service";
import { enrichServiceAccountsWithRoles } from "@/utils/rbac.util/access-rule.util/access-rule.util";
import { useQuasar } from "quasar";
import type { ServiceAccount, AccessKey } from "@/swagger-models/identity-manager-client";
import { ErrorAlert } from "@/utils/error-alert.util";
import { AccessRuleManagementModal } from "@/components/rbac/access-rule/access-rule-management-modal";

export default defineComponent({
  components: {
    AccessRuleManagementModal,
    RunaiTableWrapper,
    RunaiPageActions,
    RunaiActionButton,
    RunaiTooltipWrapper,
    CredentialCreationModal,
    RunaiDeleteModal,
    RunaiTable,
    RunaiTableNoData,
    AccessRuleTableModal,
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      appStore: useAppStore(),
      settingStore: useSettingStore(),
      selectedRows: [] as Array<ServiceAccount>,
      isDeleteModalOpen: false as boolean,
      isCreationModalOpen: false as boolean,
      isAccessRuleManagementModalOpen: false as boolean,
      deleting: false as boolean,
      renaming: false as boolean,
      isAccessRuleTableModalOpen: false as boolean,
      accessRuleTableModalOptions: {
        header: "",
        columns: subjectRolesMiniTableColumns,
      } as IAccessRuleTableModalOptions,
      accessRuleManagementModalOptions: {
        page: EAccessRuleModalPage.Application,
        subjectId: "",
        subjectType: SubjectType.ServiceAccount,
      } as IAccessRuleManagementModalOptions,
    };
  },
  setup() {
    const $q = useQuasar();
    const loadServiceAccounts = async (filterBy: IFilterBy = {}): Promise<Array<ServiceAccount>> => {
      const serviceAccounts = await serviceAccountsService.list(filterBy);
      let serviceAccountsAccessRules: IAccessRuleRecords = { displayRecords: 0, totalRecords: 0, accessRules: [] };

      if (serviceAccounts?.length > 0) {
        try {
          serviceAccountsAccessRules = await accessRuleService.getAccessRules({
            subjectType: SubjectType.ServiceAccount,
          });
          return enrichServiceAccountsWithRoles(serviceAccounts, serviceAccountsAccessRules);
        } catch (e) {
          console.error(e);
          $q.notify(alertUtil.getError(`Cannot enrich service accounts with roles`));
          return serviceAccounts;
        }
      }
      return [];
    };
    const {
      initTableFilter,
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      clearLastCreated,
    } = useTableData<ServiceAccount>(loadServiceAccounts, serviceAccountsIndexColumns, EIndexPages.APPLICATION);
    initTableFilter();
    return {
      setFilterBy,
      lastCreatedEntity,
      clearLastCreated,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
    };
  },
  created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    columns(): Array<ITableColumn> {
      return serviceAccountsIndexColumns.filter((col: ITableColumn) => col.display);
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    Action(): typeof Action {
      return Action;
    },
    canDeleteServiceAccount(): boolean {
      return this.permissionStore.hasPermission([ResourceType.ServiceAccount, ResourceType.Apps], Action.Delete);
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    selectedServiceAccount(): AccessKey | undefined {
      return this.selectedRows[0];
    },
    modalActionType(): TActionTypes {
      const isServiceAccountSelected = !!this.selectedServiceAccount;
      return !isServiceAccountSelected ? "create" : "regenerate";
    },
  },
  methods: {
    getRowKey(row: ServiceAccount): string {
      return row.id;
    },
    async refreshWithLastCreated(serviceAccountId: string): Promise<void> {
      await this.refreshList(true, serviceAccountId);
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    openCreateModal(): void {
      this.isCreationModalOpen = true;
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    openAccessRuleModal(): void {
      if (!this.selectedServiceAccount) return;
      this.accessRuleManagementModalOptions.subjectId = this.selectedServiceAccount?.name;
      this.isAccessRuleManagementModalOpen = true;
    },
    async removeServiceAccount(): Promise<void> {
      if (!this.selectedServiceAccount) return;
      try {
        this.deleting = true;
        await serviceAccountsService.remove(this.selectedServiceAccount.id);
        this.$q.notify(alertUtil.getSuccess(`Service account ${this.selectedServiceAccount.name} deleted`));
        this.resetSelectedRows();
      } catch (e: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedDeleteMessage(`Service account ${this.selectedServiceAccount?.name || ""}`),
          403: `You're not authorized to delete the selected service account`,
        });
        this.$q.notify(errorAlert.getNotification(e));
      } finally {
        this.isDeleteModalOpen = false;
        this.deleting = false;
        this.selectedRowsAmount = 0;
        this.clearLastCreated();
        await this.refreshList(true);
      }
    },
    async displayAccessRuleTableModal(serviceAccount: ServiceAccount): Promise<void> {
      this.accessRuleTableModalOptions.header = `Access Rules for ${serviceAccount.name}`;
      this.accessRuleTableModalOptions.subjectType = SubjectType.ServiceAccount;
      this.accessRuleTableModalOptions.subjectId = serviceAccount.name;
      this.isAccessRuleTableModalOpen = true;
    },
    closeAccessRuleTableModal(): void {
      this.isAccessRuleTableModalOpen = false;
      this.accessRuleTableModalOptions.subjectId = undefined;
      this.accessRuleTableModalOptions.subjectType = undefined;
      this.accessRuleTableModalOptions.header = "";
      this.refreshList(true, this.lastCreatedEntity?.id);
    },
    closeAccessRuleManagementModal(): void {
      this.isAccessRuleManagementModalOpen = false;
      this.refreshList(false, this.lastCreatedEntity?.id);
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));
      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.ServiceAccount, this.tableData, columns);
    },
  },
});
</script>
