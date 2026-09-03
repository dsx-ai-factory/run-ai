<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="department-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          :filters="filterBy"
          :columns="columns"
          primary-btn-label="New department"
          :primary-btn-disable="disablePrimaryBtn"
          :primary-tooltip-text="primaryBtnTooltip"
          @filters-changed="updateFilters"
          :selected-rows-amount="selectedRowsAmount"
          @selected-actions-close="resetSelectedRows"
          @create-entity="createNewDepartment"
          @export-csv="exportCsv"
          advanced-filters
        >
          <template v-slot:selected-rows-actions>
            <department-page-actions
              :selected-department="selectedDepartment"
              :cluster-id="clusterId"
              :available-edit-steps="availableEditSteps"
              :available-review-steps="availableReviewSteps"
              :loading-steps="loadingNodepoolCapabilities"
              @open-access-rule-modal="openAccessRuleModal"
              @open-delete-modal="openDeleteModal"
              @edit-department="editDepartment"
              @handle-edit-action="handleEditAction"
              @view-department="viewDepartment"
              @handle-review-action="handleReviewAction"
            />
          </template>
        </runai-page-actions>
      </template>
      <template v-slot:table>
        <runai-table
          :rows="departmentsTableData"
          :columns="columns"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          :loading="loadingTableData"
          @projects-clicked="displayProjects"
          @access-rules-clicked="displayAccessRuleTableModal"
          @workloads-clicked="displayWorkloadsModal"
          @node-pools-clicked="displayDepartmentNodePools"
          @update-filters="updateFilters"
          @avg-resource-timeframe-changed="onAvgResourceTimeframeChanged"
          :top-row="lastCreatedDepartment"
          :get-row-key="getRowKey"
          sticky-columns
          is-server-side-pagination
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25, 50, 100]"
          :bordered="false"
          override-table-header
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedDepartment"
              :filter-by="filterBy"
              entity-name="department"
              ph-icon-name="ph-bold ph-garage"
              icon-size="44"
              :show-error="loadingError"
              @clear-filters="clearFilters"
              @create-new="createNewDepartment"
              :disable-create-btn="!canCreateDepartment"
            />
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <project-modal v-if="openProjectModal" @close="closeProjectModal" :modal-options="projectModalOptions" />
    <access-rule-table-modal
      v-if="isAccessRuleTableModalOpen"
      @close="closeAccessRuleTableModal"
      :modal-options="accessRuleTableModalOptions"
    />
    <node-pools-modal :modal-options="nodePoolsModalOptions" v-if="isNodePoolsModalOpen" @close="closeNodePoolsModal" />
    <delete-department-modal
      v-if="isDeleteModalOpen && selectedDepartment"
      :is-department-deleting="deleting"
      :current-department="selectedDepartment"
      @cancel="isDeleteModalOpen = false"
      @delete="deleteDepartment"
    />
    <access-rule-management-modal
      v-if="isAccessRuleManagementModalOpen"
      @close="isAccessRuleManagementModalOpen = false"
      @access-rules-created="() => refreshList()"
      @access-rule-deleted="onAccessRuleDeleted"
      :modal-options="accessRuleManagementModalOptions"
    />
    <workload-list-modal
      v-if="isWorkloadsModalOpen"
      :modal-options="workloadsModalOptions"
      @close="isWorkloadsModalOpen = false"
    />
  </section>
</template>

<script lang="ts">
import { computed, defineComponent, ref } from "vue";

// cmps
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { ProjectModal } from "@/components/project/project-modal";
import { DeleteDepartmentModal } from "@/components/department/delete-department-modal";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { AccessRuleTableModal } from "@/components/rbac/access-rule/access-rule-table-modal";

// stores
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAppStore } from "@/stores/app.store";
// models
import type { IDepartmentTable, IDepartmentTableFilterBy } from "@/models/department.model";
import {
  departmentDependentColumns,
  departmentIndexColumns,
  EDepartmentColumnLabel,
} from "@/table-models/department.table-model";
import { type ITableColumn } from "@/models/table.model";
import { EFilterOperator, type IFilterBy, type IPaginationFilter } from "@/models/filter.model";
import type { IAccessRuleTableModalOptions } from "@/models/access-rule.model";
import { EAccessRuleModalPage, type IAccessRuleManagementModalOptions } from "@/models/access-rule.model";
import { accessRulesMiniTableColumns } from "@/table-models/access-rule.table-model";
// services
import { alertUtil } from "@/utils/alert.util";

// route
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import NodePoolsModal from "@/components/node-pools/node-pools-modal/node-pools-modal.vue";
import { EProjectModalEntity, type INodePoolModalOptions, type IProjectsModalOptions } from "@/models/project.model";
import { EQuotaEntity } from "@/models/resource.model";
import {
  type AccessRule,
  AccessRulesSortFilterFields,
  Action,
  ResourceType,
  ScopeType,
} from "@/swagger-models/authorization-client";
import { usePermissionStore } from "@/stores/permissions.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { WorkloadListModal } from "@/components/workload/workload-list-modal";
import { EWorkloadModalEntity, type IWorkloadListModalOptions } from "@/models/workload.model";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { workloadListModalColumns } from "@/table-models/workload.table-model";
import { useAuthStore } from "@/stores/auth.store";
import { enrichScopeEntityWithAccessRules } from "@/utils/rbac.util/access-rule.util/access-rule.util";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { ErrorAlert } from "@/utils/error-alert.util";
import {
  AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
  EOrgUnitOverTimeColumnName,
  EOrgUnitOverTimeValue,
} from "@/models/org-unit.model";
import { orgUnitUtil } from "@/utils/org-unit.util";
import DepartmentPageActions from "@/components/department/department-page-actions/department-page-actions.vue";
import { useExportFileComposable } from "@/composables/use-export-file-composable/use-export-file-composable";
import { DepartmentFilterSortFields, Verbosity } from "@/swagger-models/org-unit-service-client";
import { EAdvancedExportEntity } from "@/composables/helpers/use-export-file.mapper";
import { AccessRuleManagementModal } from "@/components/rbac/access-rule/access-rule-management-modal";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { useRoute } from "vue-router";
import { EColumnFilterType, type IFreeTextFilterModel } from "@/models/filter.model";
import { ETableFilters } from "@/models/table.model";
import { filterService } from "@/services/filter.service/filter.service";
import type { ISelectOption } from "@/models/global.model";
import { filterOrgUnitSteps, stepsToSelectOptions } from "@/utils/wizard-steps.util/wizard-steps.util";
import { DEPARTMENT_FORM_WIZARD_STEPS } from "@/models/org-unit-forms.model";
import { useNodepoolSchedulingCapabilities } from "@/composables/org-unit/use-nodepool-scheduling-capabilities.composable";

const DEFAULT_WORKLOADS_MODAL_OPTIONS: IWorkloadListModalOptions = {
  entityFilter: "",
  entityName: "",
  filterName: WorkloadSortFilterFields.DepartmentName,
  entityType: EWorkloadModalEntity.Department,
  columns: workloadListModalColumns,
  clusterId: "",
};
const DEFAULT_PROJECT_MODAL_OPTIONS: IProjectsModalOptions = {
  clusterId: "",
  entityName: "",
  entityType: EProjectModalEntity.Department,
  parentId: "",
};
export default defineComponent({
  components: {
    AccessRuleManagementModal,
    DepartmentPageActions,
    WorkloadListModal,
    NodePoolsModal,
    AccessRuleTableModal,
    RunaiPageActions,
    ProjectModal,
    DeleteDepartmentModal,
    RunaiTableWrapper,
    RunaiTable,
    RunaiTableNoData,
  },
  data() {
    return {
      selectedRows: [] as Array<IDepartmentTable>,
      openProjectModal: false as boolean,
      projectModalOptions: {
        ...DEFAULT_PROJECT_MODAL_OPTIONS,
      } as IProjectsModalOptions,
      isDeleteModalOpen: false as boolean,
      isAccessRuleTableModalOpen: false as boolean,
      isNodePoolsModalOpen: false as boolean,
      deleting: false as boolean,
      isAccessRuleManagementModalOpen: false as boolean,
      isWorkloadsModalOpen: false as boolean,
      permissionStore: usePermissionStore(),
      accessRuleTableModalOptions: {
        scopeType: ScopeType.Department,
        header: "",
        columns: accessRulesMiniTableColumns,
      } as IAccessRuleTableModalOptions,
      nodePoolsModalOptions: {
        header: "",
        entityType: EQuotaEntity.department,
        resources: [],
        nodePoolQuotaStatuses: [],
      } as INodePoolModalOptions,
      accessRuleManagementModalOptions: {
        page: EAccessRuleModalPage.Department,
        scopeType: ScopeType.Department,
        scopeName: "",
        scopeId: "",
      } as IAccessRuleManagementModalOptions,
      workloadsModalOptions: {
        ...DEFAULT_WORKLOADS_MODAL_OPTIONS,
      } as IWorkloadListModalOptions,
    };
  },
  setup() {
    const clusterStore = useClusterStore();
    const appStore = useAppStore();
    const permissionStore = usePermissionStore();
    const settingStore = useSettingStore();
    const nodePoolStore = useNodePoolStore();
    const route = useRoute();
    const isOnlyDefaultNodePool = ref(false);

    // Nodepool scheduling capabilities for step-based editing
    const {
      loading: loadingNodepoolCapabilities,
      capabilitiesByClusterId: nodepoolCapabilitiesByClusterId,
      loadCapabilities: loadNodepoolCapabilities,
    } = useNodepoolSchedulingCapabilities();

    const checkIfOnlyDefaultNodePool = async (): Promise<void> => {
      //we want to check if there is only default node pool only if we not all clusters selected.
      if (!clusterId.value) return;
      isOnlyDefaultNodePool.value = await nodePoolStore.isOnlyDefaultNodePoolByClusterId(clusterId.value);
    };

    const getDepartments = async (filterBy: IDepartmentTableFilterBy): Promise<IDepartmentTable[]> => {
      return orgUnitService.listDepartments(filterBy);
    };

    const getDepartment = async (id: string): Promise<IDepartmentTable> => {
      return (await orgUnitService.getDepartment(id)) as IDepartmentTable;
    };

    const canReadAccessRules = computed(() => permissionStore.hasPermission(ResourceType.AccessRules, Action.Read));

    const countDepartments = async (): Promise<void> => {
      const filters: IPaginationFilter = orgUnitUtil.getDepartmentListFilters(
        filterBy.value as IDepartmentTableFilterBy,
      );
      const count = await orgUnitService.countDepartments(filters.filterBy || []);
      updateFilterByKey("rowsNumber", count);
    };

    const getDepartmentsWithAdditionalData = async (
      filterBy: IDepartmentTableFilterBy,
      softLoad = false,
    ): Promise<IDepartmentTable[]> => {
      let departments = tableData.value;
      if (!softLoad) {
        departments = await getDepartments(filterBy);
      }
      try {
        if (canReadAccessRules.value) {
          const filters = [
            filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.ScopeType, ScopeType.Department),
          ];
          const scopeAccessRules = await accessRuleService.getAllAccessRules(filters);

          departments = enrichScopeEntityWithAccessRules(departments, scopeAccessRules) as IDepartmentTable[];
        }

        return departments;
      } catch (e) {
        console.error("Failed to enrich departments with additional data", e);
        return departments;
      }
    };

    const _setFilter = (filterField: string, filterTerm: string, filterLabel: string): void => {
      const filteredClusterId = route.query[DepartmentFilterSortFields.ClusterId] as string;
      const filter: IFreeTextFilterModel = {
        name: filterField,
        term: filterTerm,
        value: EFilterOperator.Equals,
        label: filterLabel,
        type: EColumnFilterType.FreeText,
      };
      if (filteredClusterId) {
        clusterStore.setSelectedClusterById(filteredClusterId);
        filterBy.value.clusterUuid = filteredClusterId;
      }
      filterService.setColumnAdvancedFilter(filterBy.value, filter, ETableFilters.DEPARTMENT);
    };

    const setDepartmentNameFilter = (): void => {
      _setFilter(
        DepartmentFilterSortFields.Name,
        route.query[DepartmentFilterSortFields.Name] as string,
        EDepartmentColumnLabel.Name,
      );
    };

    const initCustomColumnsHeader = () => {
      if (!(filterBy.value as IDepartmentTableFilterBy)[AVG_RESOURCE_TIMEFRAME_FILTER_KEY]) {
        (filterBy.value as IDepartmentTableFilterBy)[AVG_RESOURCE_TIMEFRAME_FILTER_KEY] =
          EOrgUnitOverTimeValue.TwentyFourHours;
      }
    };

    const onAvgResourceTimeframeChanged = (timeframe: string) => {
      updateFilterByKey(AVG_RESOURCE_TIMEFRAME_FILTER_KEY as keyof IFilterBy, timeframe);
      const shouldRefresh = filterBy.value?.advancedFilters?.some(
        (filter) =>
          filter.name === EOrgUnitOverTimeColumnName.AverageGpuUtilization ||
          filter.name === EOrgUnitOverTimeColumnName.AverageGpuAllocation ||
          filter.name === EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization,
      );
      if (shouldRefresh) {
        refreshList();
      }
    };

    const columns = computed<ITableColumn[]>(() => {
      return departmentIndexColumns.filter((col) => {
        if (departmentDependentColumns.cpu.includes(col.name)) {
          return settingStore.isCPUResourcesQuotaEnabled;
        }
        if (departmentDependentColumns.accessRules.includes(col.name)) {
          return permissionStore.hasPermission(ResourceType.AccessRules, Action.Read);
        }
        if (departmentDependentColumns.nodePools.includes(col.name)) {
          return !isOnlyDefaultNodePool.value;
        }
        if (departmentDependentColumns.cluster.includes(col.name)) {
          return clusterStore.isMultiCluster;
        }

        return true;
      });
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
      updateCellContent,
      removeRow,
      updateFilterByKey,
      clusterId,
    } = useTableDataAdvancedFilters<IDepartmentTable>(
      getDepartments,
      columns,
      EAdvancedIndexPages.DEPARTMENT,
      getDepartment,
      getDepartmentsWithAdditionalData,
    );
    if (route.query[DepartmentFilterSortFields.Name]) {
      setDepartmentNameFilter();
    }
    initTableFilter();

    initCustomColumnsHeader();
    checkIfOnlyDefaultNodePool();
    countDepartments();

    const { exportCsv } = useExportFileComposable(
      filterBy,
      columns,
      EAdvancedExportEntity.Departments,
      (filterBy, sortBy, sortOrder, offset, limit) =>
        orgUnitService.getDepartments(filterBy, sortBy, Verbosity.Verbose, sortOrder, offset, limit),
      (filters: IPaginationFilter) => orgUnitService.countDepartments(filters.filterBy),
      orgUnitUtil.getDepartmentListFilters,
    );

    return {
      lastCreatedEntity,
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      columns,
      updateCellContent,
      removeRow,
      clusterStore,
      settingStore,
      clusterId,
      checkIfOnlyDefaultNodePool,
      countDepartments,
      onAvgResourceTimeframeChanged,
      appStore,
      exportCsv,
      // Nodepool capabilities
      loadingNodepoolCapabilities,
      nodepoolCapabilitiesByClusterId,
      loadNodepoolCapabilities,
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
    primaryBtnTooltip(): string | undefined {
      return undefined;
    },
    disablePrimaryBtn(): boolean {
      if (this.loadingTableData) return true;
      return !this.canCreateDepartment;
    },
    canCreateDepartment(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Department, Action.Create);
    },
    selectedRowsAmount(): number {
      return this.selectedDepartment ? 1 : 0;
    },
    lastCreatedDepartment(): IDepartmentTable | null {
      return this.lastCreatedEntity;
    },

    departmentsTableData(): IDepartmentTable[] {
      const avgResourceTimeframe = (this.filterBy as IDepartmentTableFilterBy)?.[AVG_RESOURCE_TIMEFRAME_FILTER_KEY];

      const getAverageGpuUtilization = (department: IDepartmentTable) =>
        department.overtimeData &&
        orgUnitUtil.getAverageGpuUtilizationDataByTimeframe(department.overtimeData, avgResourceTimeframe);

      const getAverageGpuMemoryUtilization = (department: IDepartmentTable) =>
        department.overtimeData &&
        orgUnitUtil.getAverageGpuMemoryUtilizationDataByTimeframe(department.overtimeData, avgResourceTimeframe);

      const getAverageGpuAllocation = (department: IDepartmentTable) =>
        department.overtimeData &&
        orgUnitUtil.getAverageGpuAllocationDataByTimeframe(department.overtimeData, avgResourceTimeframe);

      return this.tableData.map((department: IDepartmentTable) => {
        return {
          ...department,
          averageGpuUtilization: getAverageGpuUtilization(department),
          averageGpuAllocation: getAverageGpuAllocation(department),
          averageGpuMemoryUtilization: getAverageGpuMemoryUtilization(department),
        };
      });
    },
    selectedDepartment(): IDepartmentTable | null {
      return this.selectedRows[0] || null;
    },
    isMultiCluster(): boolean {
      return this.clusterStore.isMultiCluster;
    },
    availableEditSteps(): ISelectOption[] {
      if (!this.selectedDepartment?.clusterId) return [];

      const filteredSteps = filterOrgUnitSteps(DEPARTMENT_FORM_WIZARD_STEPS, {
        nodepoolCapabilities: this.nodepoolCapabilitiesByClusterId[this.selectedDepartment.clusterId],
        hasPermission: (r, a) => this.permissionStore.hasPermission(r, a),
      });

      return stepsToSelectOptions(filteredSteps);
    },
    availableReviewSteps(): ISelectOption[] {
      if (!this.selectedDepartment?.clusterId) return [];

      const filteredSteps = filterOrgUnitSteps(DEPARTMENT_FORM_WIZARD_STEPS, {
        nodepoolCapabilities: this.nodepoolCapabilitiesByClusterId[this.selectedDepartment.clusterId],
        hasPermission: (r, a) => this.permissionStore.hasPermission(r, a),
      });

      return stepsToSelectOptions(filteredSteps);
    },
  },
  methods: {
    updateFilters(filters: IDepartmentTableFilterBy): void {
      this.setFilterBy(filters);
      this.countDepartments();
      this.checkIfOnlyDefaultNodePool();
    },
    getRowKey(department: IDepartmentTable): number {
      return Number(department.id);
    },

    displayProjects(department: IDepartmentTable): void {
      this.openProjectModal = Boolean(department);
      this.projectModalOptions.parentId = department.id;
      this.projectModalOptions.entityName = department.name;
      this.projectModalOptions.clusterId = department.clusterId;
    },
    closeProjectModal(): void {
      this.openProjectModal = false;
      this.resetProjectModalOptions();
    },
    resetProjectModalOptions(): void {
      this.projectModalOptions = { ...DEFAULT_PROJECT_MODAL_OPTIONS };
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    createNewDepartment(): void {
      const clusterId = this.isMultiCluster ? this.filterBy.clusterUuid : this.clusterStore.selectedClusterId;

      this.$router.push({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_NEW,
        query: { clusterId },
      });
    },
    async deleteDepartment(): Promise<void> {
      const departmentId = this.selectedDepartment?.id;
      const departmentName = this.selectedDepartment?.name;
      if (this.selectedDepartment === null || !departmentId || !departmentName) return;
      try {
        this.deleting = true;
        await orgUnitService.deleteDepartment(departmentId);
        this.isDeleteModalOpen = false;
        this.removeRow(this.selectedDepartment);
        await useAuthStore().loadUserOrgUnits();
        this.$q.notify(alertUtil.getSuccess(`Department ${departmentName} deleted`));
      } catch (err) {
        console.error(err, `Could not remove department with id ${departmentId}`);
        this.$q.notify(alertUtil.getError(`Department ${departmentName} couldn't be deleted `));
      } finally {
        this.resetSelectedRows();
        this.deleting = false;
        this.isDeleteModalOpen = false;
      }
    },
    displayAccessRuleTableModal(department: IDepartmentTable): void {
      this.accessRuleTableModalOptions.header = `Subjects Authorized for department ${department.name}`;
      this.accessRuleTableModalOptions.scopeId = department.id;
      this.isAccessRuleTableModalOpen = true;
    },
    closeAccessRuleTableModal(): void {
      this.isAccessRuleTableModalOpen = false;
      this.accessRuleTableModalOptions.scopeId = undefined;
      this.accessRuleTableModalOptions.header = "";
    },
    async displayDepartmentNodePools(department: IDepartmentTable): Promise<void> {
      try {
        this.nodePoolsModalOptions.loading = true;
        this.isNodePoolsModalOpen = true;
        const fullDepartment = await orgUnitService.getDepartment(department.id);
        this.nodePoolsModalOptions.header = `Node Pools Associated with department ${fullDepartment.name}`;
        this.nodePoolsModalOptions.resources = fullDepartment.resources;
        this.nodePoolsModalOptions.nodePoolQuotaStatuses = fullDepartment.status?.nodePoolQuotaStatuses || [];
        this.nodePoolsModalOptions.defaultNodePools = fullDepartment.defaultNodePools || [];
      } catch (error) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedLoadingMessage("node pools"),
        });
        this.$q.notify(errorAlert.getNotification(error));
      } finally {
        this.nodePoolsModalOptions.loading = false;
      }
    },
    closeNodePoolsModal(): void {
      this.isNodePoolsModalOpen = false;
      this.resetNodePoolsModalOptions();
    },
    resetNodePoolsModalOptions(): void {
      this.nodePoolsModalOptions = {
        resources: [],
        entityType: EQuotaEntity.department,
        header: "",
        nodePoolQuotaStatuses: [],
      };
    },
    openAccessRuleModal(): void {
      if (!this.selectedDepartment) return;
      this.accessRuleManagementModalOptions.scopeName = this.selectedDepartment?.name;
      this.accessRuleManagementModalOptions.scopeId = this.selectedDepartment?.id?.toString();
      this.isAccessRuleManagementModalOpen = true;
    },
    closeAccessRuleManagementModal(): void {
      this.refreshList();
      this.isAccessRuleManagementModalOpen = false;
    },
    async displayWorkloadsModal(department: IDepartmentTable): Promise<void> {
      this.isWorkloadsModalOpen = true;
      this.workloadsModalOptions.clusterId = department.clusterId;
      this.workloadsModalOptions.entityFilter = `${WorkloadSortFilterFields.DepartmentId}${EFilterOperator.Equals}${department.id}`;
      this.workloadsModalOptions.entityName = department.name;
    },
    resetWorkloadsModalOptions(): void {
      this.workloadsModalOptions = { ...DEFAULT_WORKLOADS_MODAL_OPTIONS };
    },
    onAccessRuleCreated(accessRule: AccessRule): void {
      if (!this.selectedDepartment?.id) return;
      const rolesNames = this.tableData.find(
        (department: IDepartmentTable) => department.id === this.selectedDepartment?.id,
      )?.rolesNames;
      this.updateCellContent(this.selectedDepartment, "rolesNames", [accessRule.subjectType, ...(rolesNames || [])]);
    },
    onAccessRuleDeleted(accessRuleIndex: number): void {
      if (!this.selectedDepartment?.id) return;
      const rolesNames = this.tableData.find(
        (department: IDepartmentTable) => department.id === this.selectedDepartment?.id,
      )?.rolesNames;
      const rolesWithoutDeleted = rolesNames?.filter((roleName: string, index: number) => index !== accessRuleIndex);
      this.updateCellContent(this.selectedDepartment, "rolesNames", rolesWithoutDeleted);
    },
    handleEditAction(): void {
      if (!this.selectedDepartment?.clusterId) return;
      this.loadNodepoolCapabilities(this.selectedDepartment.clusterId);
    },
    handleReviewAction(): void {
      if (!this.selectedDepartment?.clusterId) return;
      this.loadNodepoolCapabilities(this.selectedDepartment.clusterId);
    },
    viewDepartment(stepName?: string): void {
      if (!this.selectedDepartment) return;

      const query: Record<string, string> = {
        clusterId: this.selectedDepartment.clusterId,
      };
      if (stepName) {
        query.stepName = stepName;
      }

      this.$router.push({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_VIEW,
        params: { id: this.selectedDepartment.id },
        query,
      });
    },
    editDepartment(stepName?: string): void {
      if (!this.selectedDepartment) return;

      const query: Record<string, string> = {
        clusterId: this.selectedDepartment.clusterId,
      };
      if (stepName) {
        query.stepName = stepName;
      }

      this.$router.push({
        name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_EDIT,
        params: { id: this.selectedDepartment.id },
        query,
      });
    },
  },
  watch: {
    isWorkloadsModalOpen(isOpen: boolean): void {
      if (!isOpen) {
        this.resetWorkloadsModalOptions();
      }
    },
  },
});
</script>
