<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="credential-index">
    <runai-table-wrapper :filters-object="{}" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          @filters-changed="setFilterBy"
          :columns="columns"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          primary-btn-label="credential"
          @export-csv="exportTableAsCsv"
          @refresh-data="refreshList"
          support-refresh
        >
          <template v-slot:primary-btn>
            <div class="q-mr-md">
              <q-btn
                color="primary"
                class="btn-icon-sm runai-icon-btn--default"
                icon="ph-bold ph-plus"
                aid="primary-cta-btn"
                label="new credential"
                :disable="disablePrimaryBtn"
                v-permission="{ resourceType: ResourceType.Credentials, action: Action.Create }"
              >
                <credential-dropdown @credential-selected="createCredential" />
              </q-btn>
            </div>
          </template>
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper :display-tooltip="disableRename" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  aid="rename-credentials-btn"
                  v-permission="{ resourceType: ResourceType.Credentials, action: Action.Update }"
                  btn-action="edit"
                  @click="openRenameModal"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper
                :display-tooltip="disableDelete"
                :tooltip-text="deleteBtnTooltip"
                disable-break-line
              >
                <runai-action-button
                  aid="delete-credential-btn"
                  btn-action="delete"
                  @click="openDeleteModal"
                  :disable="disableDelete"
                  v-permission="{ resourceType: ResourceType.Credentials, action: Action.Delete }"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :columns="columns"
          :rows="credentials"
          :loading="isLoading"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          @update-filters="setFilterBy"
          @data-source-clicked="displayDataSources"
          @environments-clicked="displayEnvironments"
          :top-row="lastCreatedCredential"
          :get-row-key="getRowKey"
          sticky-columns
          :bordered="false"
        >
          <template v-slot:no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedCredential"
              :filter-by="filterBy"
              entity-name="credential"
              ph-icon-name="ph-bold ph-key"
              icon-size="44"
              :show-error="loadingError"
              @clear-filters="clearFilters"
              :show-create-btn="!disablePrimaryBtn"
              ignore-cluster-filter
            >
              <template v-slot:primary-btn>
                <q-btn
                  color="primary"
                  flat
                  class="btn-icon-sm runai-icon-btn--default"
                  icon="ph-bold ph-plus"
                  aid="empty-state-add-entity-btn"
                  label="new credential"
                  v-permission="{ resourceType: ResourceType.Credentials, action: Action.Create }"
                >
                  <credential-dropdown @credential-selected="createCredential" />
                </q-btn>
              </template>
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedRows[0]"
      :entity-name="selectedCredentialName"
      :deleting="isDeleting"
      entity-type="credential"
      @cancel="isDeleteModalOpen = false"
      @delete="removeCredential"
    />
    <data-source-list-modal
      v-if="openDataSourceModal"
      @close="closeDataSourceModal"
      :modal-options="dataSourceModalOptions"
    />

    <environment-list-modal
      v-if="openEnvironmentModal"
      @close="closeEnvironmentModal"
      :modal-options="environmentModalOptions"
    />

    <runai-rename-modal
      v-if="isRenameModalOpen"
      :asset-type="selectedCredentialKind"
      :scope-model="selectedScopeModel"
      :model-value="isRenameModalOpen"
      :existing-name="selectedCredentialName"
      :existing-description="selectedCredentialDescription"
      :renaming="renaming"
      :show-name-input="false"
      main-action="Edit"
      submit-label="Save"
      @close="isRenameModalOpen = false"
      @rename="renameCredential"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { DataSourceListModal } from "@/components/data-source/data-source-list-modal";
import { CredentialDropdown } from "@/components/credential/credential-dropdown";
import { RunaiRenameModal } from "@/components/common/runai-rename-modal";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { EnvironmentListModal } from "@/components/environment/environment-list-modal";

import { RunaiDeleteModal } from "@/components/common";
// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useCredentialStore } from "@/stores/credential.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useSettingStore } from "@/stores/setting.store";

//util
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";
// types
import type { TCredentialKinds, TCredentialUpdateRequests } from "@/models/credential.model";
import { credentialIndexColumns } from "@/table-models/credential.table-model";
import type { ITableColumn } from "@/models/table.model";
import {
  Scope,
  AssetKind,
  type AccessKeyUpdateRequest,
  type CredentialsListResponseEntry,
  type DatasourceRef,
  type PasswordUpdateRequest,
  type DockerRegistryUpdateRequest,
  type AssetUpdateRequest,
  type GenericSecretUpdateRequest,
  type AssetRef,
  type NgcApiKeyUpdateRequest,
} from "@/swagger-models/assets-service-client";
import type { IAssetsFilter } from "@/models/filter.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import type { IScopeModel } from "@/models/global.model";
import type { IDataSourceModalOptions } from "@/models/data-source.model";
import type { IEnvironmentModalOptions } from "@/models/environment.model";
// composable
import { useTableData } from "@/composables/use-table-data.composable";
import { EIndexPages } from "@/composables/helpers/use-table-data.mapper";

// services
import { filterService } from "@/services/filter.service/filter.service";

export default defineComponent({
  components: {
    RunaiTableWrapper,
    RunaiPageActions,
    RunaiTooltipWrapper,
    RunaiActionButton,
    DataSourceListModal,
    EnvironmentListModal,
    CredentialDropdown,
    RunaiRenameModal,
    RunaiTable,
    RunaiTableNoData,
    RunaiDeleteModal,
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      appStore: useAppStore(),
      credentialStore: useCredentialStore(),
      authStore: useAuthStore(),
      settingsStore: useSettingStore(),
      selectedRows: [] as Array<CredentialsListResponseEntry>,
      isDeleteModalOpen: false as boolean,
      isDeleting: false as boolean,
      isRenameModalOpen: false as boolean,
      renaming: false as boolean,
      openDataSourceModal: false as boolean,
      openEnvironmentModal: false as boolean,
      dataSourceModalOptions: { dataSourceIds: [], header: "" } as IDataSourceModalOptions,
      environmentModalOptions: { environmentIds: [], header: "" } as IEnvironmentModalOptions,
    };
  },
  setup() {
    const columns = credentialIndexColumns;

    const {
      initTableFilter,
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      clearLastCreated,
      lastCreatedEntity,
    } = useTableData<CredentialsListResponseEntry>(
      useCredentialStore().loadCredentials,
      columns,
      EIndexPages.CREDENTIAL,
    );

    const indexSpecificFilter: IAssetsFilter = {
      usageInfo: true,
      statusInfo: true,
    };

    initTableFilter(indexSpecificFilter);

    return {
      lastCreatedEntity,
      clearLastCreated,
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      columns,
    };
  },
  async created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    selectedCredentialKind(): TCredentialKinds {
      return this.selectedRows[0].meta.kind as TCredentialKinds;
    },
    canCreateCredential(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Credentials, Action.Create);
    },
    credentials(): Array<CredentialsListResponseEntry> {
      return this.tableData;
    },
    isLoading(): boolean {
      return this.loadingTableData;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    canDeleteCredential(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Credentials, Action.Delete);
    },
    disableDelete(): boolean {
      if (!this.canDeleteCredential) return true;
      return (
        !!this.selectedRows[0]?.usedBy?.assets?.datasources?.length ||
        !!this.selectedRows[0]?.usedBy?.assets?.environment
      );
    },
    disableRename(): boolean {
      return !this.permissionStore.hasPermission(ResourceType.Credentials, Action.Update);
    },
    deleteBtnTooltip(): string {
      if (!this.canDeleteCredential) {
        return this.noActionPermissionText;
      } else {
        return "The selected credential is being used by at least one data source or environment and therefore can't be deleted.";
      }
    },
    noActionPermissionText(): string {
      return "Action for the selected credential is not authorized";
    },
    selectedCredentialName(): string {
      return this.selectedRows[0]?.meta.name || "";
    },
    selectedCredentialDescription(): string {
      return this.selectedRows[0]?.meta.description || "";
    },
    lastCreatedCredential(): CredentialsListResponseEntry | null {
      return this.lastCreatedEntity;
    },
    disablePrimaryBtn(): boolean {
      if (this.loadingTableData) return true;
      return !this.canCreateCredential;
    },
    selectedScopeModel(): IScopeModel {
      if (!this.selectedRows[0]) return { scope: Scope.Tenant };
      return {
        scope: this.selectedRows[0].meta.scope,
        projectId: this.selectedRows[0].meta.projectId,
        departmentId: this.selectedRows[0].meta.departmentId,
        clusterId: this.selectedRows[0].meta.clusterId,
      };
    },
  },
  methods: {
    getRowKey(credential: CredentialsListResponseEntry): string {
      return credential.meta.id;
    },
    openRenameModal(): void {
      this.isRenameModalOpen = true;
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    createCredential(pageName: string): void {
      const query = filterService.getClusterUuidQueryFromColumnFilter(this.filterBy);
      this.$router.push({ name: pageName, query });
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    displayDataSources(credential: CredentialsListResponseEntry): void {
      if (!credential.usedBy?.assets?.datasources) return;

      const datasources: Array<DatasourceRef> = credential.usedBy?.assets.datasources;

      this.dataSourceModalOptions.header = `Data Sources Associated with Credential ${credential.meta.name}`;
      this.dataSourceModalOptions.dataSourceIds = datasources.map((dataSource: DatasourceRef) => dataSource.id);
      this.openDataSourceModal = Boolean(credential);
    },
    displayEnvironments(credential: CredentialsListResponseEntry): void {
      if (!credential.usedBy?.assets?.environments) return;

      const environments: Array<AssetRef> = credential.usedBy?.assets.environments;

      this.environmentModalOptions.header = `Environments Associated with Credential ${credential.meta.name}`;
      this.environmentModalOptions.environmentIds = environments.map((env: AssetRef) => env.id);
      this.openEnvironmentModal = Boolean(credential);
    },
    async removeCredential(): Promise<void> {
      if (!this.selectedRows.length) return;
      try {
        this.isDeleting = true;
        await this.credentialStore.remove(
          this.selectedRows[0].meta.id,
          this.selectedRows[0].meta.kind as TCredentialKinds,
        );
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Credential ${this.selectedCredentialName} deleted`));
      } catch (e: unknown) {
        this.$q.notify(alertUtil.getError(`Credential ${this.selectedCredentialName} couldn't be deleted`));
      } finally {
        this.resetSelectedRows();
        this.isDeleteModalOpen = false;
        this.isDeleting = false;
      }
    },
    async renameCredential(changes: AssetUpdateRequest): Promise<void> {
      if (!this.selectedRows.length) return;
      try {
        this.renaming = true;
        const credentialToUpdate = {
          meta: {
            // TODO: uncomment when rename is possbiel
            // name: changes.name,
            description: changes.description,
          },
        } as TCredentialUpdateRequests;
        await this.updateCredential(credentialToUpdate);
        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Credential ${this.selectedCredentialName} updated`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.$q.notify(alertUtil.getError(`Credential ${this.selectedCredentialName} couldn't be updated`));
        }
      } finally {
        this.resetSelectedRows();
        this.isRenameModalOpen = false;
        this.renaming = false;
      }
    },
    async updateCredential(credentialToUpdate: TCredentialUpdateRequests): Promise<void> {
      const selectedCredId: string = this.selectedRows[0].meta.id;

      switch (this.selectedCredentialKind) {
        case AssetKind.AccessKey:
          return this.credentialStore.updateAccessKey(selectedCredId, credentialToUpdate as AccessKeyUpdateRequest);
        case AssetKind.Password:
          return this.credentialStore.updateUserpass(selectedCredId, credentialToUpdate as PasswordUpdateRequest);
        case AssetKind.DockerRegistry:
          return this.credentialStore.updateDockerRegistry(
            selectedCredId,
            credentialToUpdate as DockerRegistryUpdateRequest,
          );
        case AssetKind.GenericSecret:
          return this.credentialStore.updateGenericSecret(
            selectedCredId,
            credentialToUpdate as GenericSecretUpdateRequest,
          );
        case AssetKind.NgcApiKey:
          return this.credentialStore.updateNgcApiKey(selectedCredId, credentialToUpdate as NgcApiKeyUpdateRequest);
      }
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));

      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.Credential, this.credentials, columns);
    },
    closeDataSourceModal(): void {
      this.openDataSourceModal = false;
      this.resetDataSourceModalOptions();
    },
    resetDataSourceModalOptions(): void {
      this.dataSourceModalOptions = { dataSourceIds: [], header: "" };
    },
    closeEnvironmentModal(): void {
      this.openEnvironmentModal = false;
      this.resetEnvironmentModalOptions();
    },
    resetEnvironmentModalOptions(): void {
      this.environmentModalOptions = { environmentIds: [], header: "" };
    },
  },
});
</script>
<style lang="scss" scoped>
.disabled-tool-tip {
  white-space: nowrap;
}
</style>
