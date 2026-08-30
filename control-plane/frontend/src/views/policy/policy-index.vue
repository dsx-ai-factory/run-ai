<template>
  <section class="policy-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :columns="columns"
          :filters="filterBy"
          primary-btn-label="new policy"
          :primary-btn-disable="disablePrimaryBtn"
          @filters-changed="setFilterBy"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          @create-entity="createPolicy"
          @refresh-data="refreshList"
          support-refresh
          hide-export-csv
        >
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper :display-tooltip="disableEdit" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Policies, action: Action.Update }"
                  btn-action="edit"
                  aid="edit-policy-btn"
                  @click="editPolicy"
                  :disable="disableEdit"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableView" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Project, action: Action.Read }"
                  aid="view-project-policy-btn"
                  btn-action="viewPolicy"
                  @click="viewPolicy"
                  :disable="disableView"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="noActionPermissionText">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Policies, action: Action.Delete }"
                  btn-action="delete"
                  aid="delete-policy-btn"
                  @click="openDeleteModal"
                  :disable="disableDelete"
                />
              </runai-tooltip-wrapper>
            </section>
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          class="policy-table"
          :rows="policies"
          :columns="columns"
          :loading="loadingTableData"
          :filter-by="filterBy"
          :top-row="lastCreatedPolicy"
          :loading-error="loadingError"
          :get-row-key="getRowKey"
          v-model:selected="selectedRows"
          sticky-columns
          @update-filters="setFilterBy"
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedPolicy"
              :disable-create-btn="!canCreatePolicy"
              :show-error="loadingError"
              :filter-by="filterBy"
              entity-name="policy"
              ph-icon-name="ph-bold ph-clipboard-text"
              icon-size="44"
              @create-new="createPolicy"
              @clear-filters="clearPolicyFilters"
              ignore-cluster-filter
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      aid="delete-policy-modal"
      v-if="isDeleteModalOpen && selectedPolicy"
      :entity-name="selectedPolicyName"
      entity-type="policy"
      @cancel="isDeleteModalOpen = false"
      @delete="removePolicy"
      :deleting="deleting"
    />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { DOCS } from "@/constants/docs";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { RunaiDeleteModal } from "@/components/common";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// stores
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { usePolicyStore, type IPolicy } from "@/stores/policy.store";
// models
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { PolicyType } from "@/swagger-models/policy-service-client";

// utils
import { alertUtil } from "@/utils/alert.util";

// constants
import { allPolicyColumnsMap, policyIndexColumns } from "@/table-models/policy.table-model";
import { HttpErrorResponse } from "@/models/http-response.model";
import { POLICIES_ROUTE_NAMES } from "@/router/policy.routes/policy.routes.names";
import { useTableData } from "@/composables/use-table-data.composable";
import { EIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { ENotificationBarVariant } from "@/components/common/notification-alert-bar";
import { useClusterStore } from "@/stores/cluster.store";

// services
import { filterService } from "@/services/filter.service/filter.service";

export default defineComponent({
  components: {
    RunaiTableWrapper,
    RunaiPageActions,
    RunaiActionButton,
    RunaiTooltipWrapper,
    RunaiTable,
    RunaiTableNoData,
    RunaiDeleteModal,
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
      policyStore: usePolicyStore(),
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      selectedRows: [] as IPolicy[],
      isDeleteModalOpen: false as boolean,
      deleting: false as boolean,
      ENotificationBarVariant,
      DOCS,
    };
  },
  setup() {
    const columns = policyIndexColumns;

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
    } = useTableData<IPolicy>(usePolicyStore().loadPolicies, columns, EIndexPages.POLICY);

    initTableFilter();

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
  created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    disablePrimaryBtn(): boolean {
      if (this.loadingTableData) return true;
      return !this.canCreatePolicy;
    },
    canCreatePolicy(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Policies, Action.Create);
    },
    canDeletePolicy(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Policies, Action.Delete);
    },
    canEditPolicy(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Policies, Action.Update);
    },
    disableDelete(): boolean {
      return !this.canDeletePolicy;
    },
    disableEdit(): boolean {
      return !this.canEditPolicy || this.isDistributedInferencePolicy;
    },
    disableView(): boolean {
      return (
        !this.permissionStore.hasPermission(ResourceType.Policies, Action.Read) || this.isDistributedInferencePolicy
      );
    },
    isDistributedInferencePolicy(): boolean {
      return this.selectedPolicy?.type === PolicyType.DistributedInference;
    },
    noActionPermissionText(): string {
      if (this.isDistributedInferencePolicy) {
        return "This action isn't supported for distributed inference policies. To perform it, use the API.";
      }
      return "Action for the selected policy is not authorized";
    },
    policies(): IPolicy[] {
      return this.tableData;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    selectedPolicy(): IPolicy | undefined {
      return this.selectedRows[0];
    },
    lastCreatedPolicy(): IPolicy | null {
      return this.lastCreatedEntity;
    },
    selectedPolicyName(): string {
      return this.selectedPolicy?.meta?.name || "";
    },
    selectedPolicyId(): number | null {
      return this.selectedPolicy?.meta?.projectId || null;
    },
  },
  methods: {
    getRowKey(policy: IPolicy): string {
      return policy.meta?.id || "";
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    clearPolicyFilters(): void {
      const defaultFilters = filterService.getDefaultFilters("policy", policyIndexColumns);
      defaultFilters.columnFilters = defaultFilters.columnFilters || [];

      defaultFilters.columnFilters.push({
        name: allPolicyColumnsMap.cluster.name,
        term: this.filterBy.clusterUuid || "",
        field: allPolicyColumnsMap.cluster.field,
      });

      this.clearFilters(defaultFilters);
    },
    createPolicy(): void {
      const query = filterService.getClusterUuidQueryFromColumnFilter(this.filterBy);
      this.$router.push({ name: POLICIES_ROUTE_NAMES.POLICY_NEW, query });
    },
    editPolicy(): void {
      if (!this.selectedPolicy) return;
      this.$router.push({ name: POLICIES_ROUTE_NAMES.POLICY_EDIT, params: { id: this.selectedPolicy.policyId } });
    },
    viewPolicy(): void {
      this.$router.push({ name: POLICIES_ROUTE_NAMES.POLICY_VIEW, params: { id: this.selectedPolicy?.policyId } });
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    getErrorMsgIfAccessDenied(baseMsg: string, error: string): string {
      const errStr = error.toString().toLowerCase();
      if (errStr.includes("access denied") || errStr.includes("not allowed")) {
        return `${baseMsg}: Access Denied`;
      }
      return baseMsg;
    },
    notifyMsgWithError(baseMsg: string, error: string): void {
      const errorStr = this.getErrorMsgIfAccessDenied(baseMsg, error);
      this.$q.notify(alertUtil.getError(errorStr));
    },
    async removePolicy(): Promise<void> {
      if (!this.selectedPolicy) return;
      try {
        this.deleting = true;
        await this.policyStore.remove(this.selectedPolicy);
        this.clearLastCreated();
        this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Policy ${this.selectedPolicyName} deleted`));
      } catch (error: unknown) {
        if (error instanceof HttpErrorResponse) {
          console.error(error.serialize());
          this.$q.notify(alertUtil.getError(error.message));
        } else {
          console.error(error);
          this.notifyMsgWithError("Failed to delete policy", error as string);
        }
      } finally {
        this.resetSelectedRows();
        this.isDeleteModalOpen = false;
        this.deleting = false;
      }
    },
  },
});
</script>
