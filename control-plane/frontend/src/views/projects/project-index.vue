<template>
  <section>
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          primary-btn-label="new project"
          :filters="filterBy"
          :columns="columns"
          :selected-rows-amount="selectedRowsAmount"
          @filters-changed="updateFilters"
          @create-entity="createProject"
          @selected-actions-close="resetSelectedRows"
          @export-csv="exportCsv"
          advanced-filters
        >
          <template v-slot:selected-rows-actions>
            <runai-action-button
              v-permission="{ resourceType: ResourceType.Project, action: Action.Update }"
              aid="edit-project-btn"
              btn-action="edit"
              @click="handleEditAction"
            >
              <runai-select-menu
                :list="availableEditSteps"
                :loading="loadingNodepoolCapabilities"
                @selected-value="editProject"
              />
            </runai-action-button>
            <runai-tooltip-wrapper
              :display-tooltip="disablePolicyButton"
              tooltip-text="There is no policy applied for the selected project"
            >
              <runai-action-button
                v-if="showPolicyButton"
                v-permission="{ resourceType: ResourceType.Project, action: Action.Read }"
                aid="view-project-policy-btn"
                btn-action="viewPolicy"
                :disable="disablePolicyButton"
              >
                <policy-dropdown
                  @policy-selected="viewPolicy"
                  entity-name="project"
                  :existing-policies="policyMenuOptions"
                  :additional-options="gvkPolicyMenuOptions"
                />
              </runai-action-button>
            </runai-tooltip-wrapper>
            <runai-action-button
              v-permission="{ resourceType: ResourceType.Project, action: Action.Delete }"
              aid="delete-project-btn"
              btn-action="delete"
              @click="openDeleteModal"
            />
          </template>
        </runai-page-actions>
      </template>

      <template v-slot:table>
        <runai-table
          :columns="columns"
          :rows="projectsTableData"
          :loading="loadingTableData"
          :filter-by="filterBy"
          v-model:selected="selectedProjects"
          @update-filters="updateFilters"
          @node-pools-clicked="displayProjectNodePools"
          @access-rules-clicked="displayAccessRuleTableModal"
          @workloads-clicked="displayWorkloadsModal"
          @avg-resource-timeframe-changed="onAvgResourceTimeframeChanged"
          @status-clicked="onProjectStatusClicked"
          :get-row-key="getRowKey"
          :top-row="lastCreatedProject"
          sticky-columns
          is-server-side-pagination
          :rows-per-page-options="[1, 5, 7, 10, 15, 20, 25, 50, 100]"
          :bordered="false"
          override-table-header
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loadingTableData && !lastCreatedProject"
              :filter-by="filterBy"
              entity-name="project"
              ph-icon-name="ph-bold ph-suitcase"
              icon-size="44"
              @clear-filters="clearFilters"
              @create-new="createProject"
              :disable-create-btn="!canCreateProjects"
              :show-error="loadingError"
              :custom-message="getCustomMessage"
              ignore-cluster-filter
            >
              <template v-if="!canCreateProjects" v-slot:secondary-text>
                To be assigned to a project contact your administrator.</template
              >
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>
    <node-pools-modal v-if="isNodePoolsModalOpen" :modal-options="nodePoolsModalOptions" @close="closeNodePoolsModal" />
    <access-rule-table-modal
      v-if="isAccessRuleTableModalOpen"
      @close="closeAccessRuleTableModal"
      :modal-options="accessRuleTableModalOptions"
    />
    <project-delete-modal
      v-if="isDeleteModalOpen && selectedProject"
      :project="selectedProject"
      @delete="(deleteType?: DeletionTypes) => selectedProject && onProjectDelete(selectedProject, deleteType)"
      @close="isDeleteModalOpen = false"
      :deleting="isDeleting"
    />
    <access-rule-management-modal
      v-if="isAccessRuleManagementModalOpen"
      @close="isAccessRuleManagementModalOpen = false"
      @access-rules-created="refreshList"
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
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
//components
import { NodePoolsModal } from "@/components/node-pools/node-pools-modal";
import { PolicyDropdown } from "@/components/policy/policy-dropdown";
import { ProjectDeleteModal } from "@/components/project/project-delete-modal";
import { AccessRuleManagementModal } from "@/components/rbac/access-rule/access-rule-management-modal";
import { AccessRuleTableModal } from "@/components/rbac/access-rule/access-rule-table-modal/";
import { WorkloadListModal } from "@/components/workload/workload-list-modal";
import { EAdvancedExportEntity } from "@/composables/helpers/use-export-file.mapper";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import { useExportFileComposable } from "@/composables/use-export-file-composable/use-export-file-composable";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import type { IAccessRuleManagementModalOptions, IAccessRuleTableModalOptions } from "@/models/access-rule.model";
import { EAccessRuleModalPage } from "@/models/access-rule.model";
//models
import {
  EColumnFilterType,
  EFilterOperator,
  type IFilterBy,
  type IFreeTextFilterModel,
  type IPaginationFilter,
} from "@/models/filter.model";
import {
  AVG_RESOURCE_TIMEFRAME_FILTER_KEY,
  EOrgUnitOverTimeColumnName,
  EOrgUnitOverTimeValue,
} from "@/models/org-unit.model";
import type { IPolicyMap } from "@/models/policy.model";
import { type INodePoolModalOptions, type IProjectTable, type IProjectTableFilterBy } from "@/models/project.model";
import { EQuotaEntity } from "@/models/resource.model";
import { ETableFilters, type ITableColumn } from "@/models/table.model";
import { EWorkloadModalEntity, type IWorkloadListModalOptions } from "@/models/workload.model";
import { POLICIES_ROUTE_NAMES } from "@/router/policy.routes/policy.routes.names";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { policyService } from "@/services/control-plane/policy.service/policy.service";
import { gvkPolicyService } from "@/services/control-plane/gvk-policy.service/gvk-policy.service";
import type { IPolicySelectOptions } from "@/components/policy/policy-dropdown/policy-dropdown.vue";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";
import { filterService } from "@/services/filter.service/filter.service";
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useNodePoolStore } from "@/stores/node-pool.store";
import { usePermissionStore } from "@/stores/permissions.store";
//store
import { useSettingStore } from "@/stores/setting.store";
import {
  Action,
  ResourceType,
  ScopeType,
  type AccessRule,
  AccessRulesSortFilterFields,
} from "@/swagger-models/authorization-client";
import { DeletionTypes } from "@/swagger-models/org-unit-service-client";
import { ProjectFilterSortFields, type Project } from "@/swagger-models/org-unit-service-client";
import {
  PolicyType,
  type Policy as GvkPolicy,
  type DistributedInferencePolicyV2,
  type DistributedPolicyDefaultsV2,
  type DistributedPolicyRulesV2,
  type DistributedPolicyV2,
  type InferencePolicyDefaultsAndRulesV2Defaults,
  type InferencePolicyDefaultsAndRulesV2Rules,
  type InferencePolicyV2,
  type TrainingPolicyDefaultsAndRulesV2Defaults,
  type TrainingPolicyDefaultsAndRulesV2Rules,
  type TrainingPolicyV2,
  type WorkspacePolicyDefaultsAndRulesV2Defaults,
  type WorkspacePolicyDefaultsAndRulesV2Rules,
  type WorkspacePolicyV2,
} from "@/swagger-models/policy-service-client";
import { WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import { accessRulesMiniTableColumns } from "@/table-models/access-rule.table-model";
import { EProjectColumnLabel, projectDependentColumns, projectIndexColumns } from "@/table-models/project.table-model";
import { workloadListModalColumns } from "@/table-models/workload.table-model";
//constants
import { alertUtil } from "@/utils/alert.util";
import { orgUnitUtil } from "@/utils/org-unit.util";
import { policyUtil, type IPolicyScopeProperties } from "@/utils/policy.util/policy.util";
import { enrichScopeEntityWithAccessRules } from "@/utils/rbac.util/access-rule.util/access-rule.util";
import { computed, defineComponent, ref, type DefineComponent } from "vue";
import { useRoute } from "vue-router";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { PROJECT_FORM_WIZARD_STEPS } from "@/models/org-unit-forms.model";
import type { ISelectOption } from "@/models/global.model";
import { filterOrgUnitSteps, stepsToSelectOptions } from "@/utils/wizard-steps.util";
import { RunaiSelectMenu } from "@/components/common/runai-select-menu";
import { useNodepoolSchedulingCapabilities } from "@/composables/org-unit/use-nodepool-scheduling-capabilities.composable";

const DEFAULT_WORKLOADS_MODAL_OPTIONS: IWorkloadListModalOptions = {
  entityFilter: "",
  entityName: "",
  filterName: WorkloadSortFilterFields.ProjectName,
  entityType: EWorkloadModalEntity.Project,
  columns: workloadListModalColumns,
  clusterId: "",
};

const ProjectsIndex = defineComponent({
  name: "projects-index",
  components: {
    AccessRuleManagementModal,
    RunaiTooltipWrapper,
    WorkloadListModal,
    AccessRuleTableModal,
    RunaiTableNoData,
    RunaiTable,
    ProjectDeleteModal,
    NodePoolsModal,
    RunaiPageActions,
    RunaiTableWrapper,
    RunaiActionButton,
    PolicyDropdown,
    RunaiSelectMenu,
  },
  data() {
    return {
      appStore: useAppStore(),
      isDeleting: false as boolean,
      isDeleteModalOpen: false as boolean,
      isNodePoolsModalOpen: false as boolean,
      isAccessRuleTableModalOpen: false as boolean,
      isAccessRuleManagementModalOpen: false as boolean,
      isWorkloadsModalOpen: false as boolean,
      selectedProjects: [] as Array<IProjectTable>,
      accessRuleTableModalOptions: {
        header: "",
        columns: accessRulesMiniTableColumns,
        scopeType: ScopeType.Project,
      } as IAccessRuleTableModalOptions,
      nodePoolsModalOptions: {
        resources: [],
        nodePoolQuotaStatuses: [],
        entityType: EQuotaEntity.project,
        header: "",
      } as INodePoolModalOptions,
      accessRuleManagementModalOptions: {
        page: EAccessRuleModalPage.Project,
        scopeType: ScopeType.Project,
        scopeName: "",
        scopeId: "",
      } as IAccessRuleManagementModalOptions,
      currentTimeoutId: 0 as number,
      loadingPolicies: false as boolean,
      selectedProjectPolicies: null as IPolicyMap | null,
      selectedProjectGvkPolicies: [] as GvkPolicy[],
      workloadsModalOptions: {
        ...DEFAULT_WORKLOADS_MODAL_OPTIONS,
      } as IWorkloadListModalOptions,
      loadedPoliciesMapByProjectId: {} as Record<string, IPolicyMap>,
      loadedGvkPoliciesByProjectId: {} as Record<string, GvkPolicy[]>,
    };
  },
  setup() {
    const settingStore = useSettingStore();
    const clusterStore = useClusterStore();
    const permissionStore = usePermissionStore();
    const nodePoolStore = useNodePoolStore();
    const route = useRoute();

    const isOnlyDefaultNodePool = ref(false);

    // Nodepool scheduling capabilities for step-based editing
    const {
      loading: loadingNodepoolCapabilities,
      capabilitiesByClusterId: nodepoolCapabilitiesByClusterId,
      loadCapabilities: loadNodepoolCapabilities,
    } = useNodepoolSchedulingCapabilities();

    const columns = computed<ITableColumn[]>(() => {
      return projectIndexColumns.filter((col: ITableColumn) => {
        if (projectDependentColumns.cpu.has(col.name)) {
          return settingStore.isCPUResourcesQuotaEnabled;
        }
        if (projectDependentColumns.accessRules.has(col.name)) {
          return permissionStore.hasPermission(ResourceType.AccessRules, Action.Read);
        }
        if (projectDependentColumns.nodePools.has(col.name)) {
          return !isOnlyDefaultNodePool.value;
        }
        if (projectDependentColumns.cluster.has(col.name)) {
          return clusterStore.isMultiCluster;
        }
        return true;
      });
    });

    const canReadAccessRules = computed(() => permissionStore.hasPermission(ResourceType.AccessRules, Action.Read));

    const checkIfOnlyDefaultNodePool = async (): Promise<void> => {
      //we want to check if there is only default node pool only if we not all clusters selected.
      if (!clusterId.value) return;
      try {
        isOnlyDefaultNodePool.value = await nodePoolStore.isOnlyDefaultNodePoolByClusterId(clusterId.value);
      } catch (e) {
        isOnlyDefaultNodePool.value = false;
      }
    };

    const getProjects = async (filterBy: IProjectTableFilterBy): Promise<IProjectTable[]> => {
      return orgUnitService.listProjects(filterBy);
    };

    const getProject = async (id: string): Promise<IProjectTable> => {
      return (await orgUnitService.getProject(id)) as IProjectTable;
    };

    const countProjects = async (): Promise<void> => {
      const filters: IPaginationFilter = orgUnitUtil.getProjectListFilters(filterBy.value as IProjectTableFilterBy);
      const count = await orgUnitService.countProjects(filters);
      updateFilterByKey("rowsNumber", count);
    };

    const getProjectsWithAdditionalData = async (
      filterBy: IProjectTableFilterBy,
      softLoad = false,
    ): Promise<IProjectTable[]> => {
      let projects = tableData.value;
      if (!softLoad) {
        projects = await getProjects(filterBy);
      }
      try {
        if (canReadAccessRules.value) {
          const filters = [filterUtil.getEqualsFilterString(AccessRulesSortFilterFields.ScopeType, ScopeType.Project)];
          const scopeAccessRules = await accessRuleService.getAllAccessRules(filters);

          return enrichScopeEntityWithAccessRules(projects, scopeAccessRules) as IProjectTable[];
        }
      } catch (e) {
        console.error("Failed to enrich projects with additional data", e);
      }
      return projects as IProjectTable[];
    };

    const _setFilter = (filterField: string, filterTerm: string, filterLabel: string): void => {
      const filteredClusterId = route.query[ProjectFilterSortFields.ClusterId] as string;
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
      filterService.setColumnAdvancedFilter(filterBy.value, filter, ETableFilters.PROJECT);
    };

    const setProjectNameFilter = (): void => {
      _setFilter(
        ProjectFilterSortFields.Name,
        route.query[ProjectFilterSortFields.Name] as string,
        EProjectColumnLabel.ProjectName,
      );
    };

    const setDepartmentNameFilter = (): void => {
      _setFilter(
        ProjectFilterSortFields.ParentName,
        route.query[ProjectFilterSortFields.ParentName] as string,
        EProjectColumnLabel.DepartmentName,
      );
    };

    const initCustomColumnsHeader = () => {
      if (!(filterBy.value as IProjectTableFilterBy)[AVG_RESOURCE_TIMEFRAME_FILTER_KEY]) {
        (filterBy.value as IProjectTableFilterBy)[AVG_RESOURCE_TIMEFRAME_FILTER_KEY] =
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

    const {
      setFilterBy,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      lastCreatedEntity,
      clearLastCreated,
      initTableFilter,
      updateCellContent,
      removeRow,
      updateFilterByKey,
      clusterId,
    } = useTableDataAdvancedFilters<IProjectTable>(
      getProjects,
      columns,
      EAdvancedIndexPages.PROJECT,
      getProject,
      getProjectsWithAdditionalData,
    );
    if (route.query[ProjectFilterSortFields.Name]) {
      setProjectNameFilter();
    }
    if (route.query[ProjectFilterSortFields.ParentName]) {
      setDepartmentNameFilter();
    }
    initTableFilter();

    initCustomColumnsHeader();

    checkIfOnlyDefaultNodePool();
    countProjects();

    const { exportCsv } = useExportFileComposable<Project>(
      filterBy,
      columns,
      EAdvancedExportEntity.Projects,
      orgUnitService.getProjects,
      orgUnitService.countProjects,
      orgUnitUtil.getProjectListFilters,
    );

    return {
      lastCreatedEntity,
      clearLastCreated,
      setFilterBy,
      updateFilterByKey,
      clearFilters,
      refreshList,
      tableData,
      filterBy,
      loadingTableData,
      loadingError,
      columns,
      updateCellContent,
      removeRow,
      settingStore,
      clusterStore,
      permissionStore,
      nodePoolStore,
      clusterId,
      countProjects,
      checkIfOnlyDefaultNodePool,
      onAvgResourceTimeframeChanged,
      exportCsv,
      // Nodepool capabilities
      loadingNodepoolCapabilities,
      nodepoolCapabilitiesByClusterId,
      loadNodepoolCapabilities,
    };
  },
  created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    DeletionTypes(): typeof DeletionTypes {
      return DeletionTypes;
    },
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    selectedRowsAmount(): number {
      return this.selectedProjects.length;
    },
    canCreateProjects(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Project, Action.Create);
    },
    selectedProject(): IProjectTable | null {
      if (this.selectedProjects.length > 0) {
        return this.selectedProjects[0];
      }
      return null;
    },
    projectsTableData(): IProjectTable[] {
      const avgResourceTimeframe = (this.filterBy as IProjectTableFilterBy)?.[AVG_RESOURCE_TIMEFRAME_FILTER_KEY];

      const getAverageGpuUtilization = (project: IProjectTable) =>
        project.overtimeData &&
        orgUnitUtil.getAverageGpuUtilizationDataByTimeframe(project.overtimeData, avgResourceTimeframe);

      const getAverageGpuMemoryUtilization = (project: IProjectTable) =>
        project.overtimeData &&
        orgUnitUtil.getAverageGpuMemoryUtilizationDataByTimeframe(project.overtimeData, avgResourceTimeframe);

      const getAverageGpuAllocation = (project: IProjectTable) =>
        project.overtimeData &&
        orgUnitUtil.getAverageGpuAllocationDataByTimeframe(project.overtimeData, avgResourceTimeframe);

      return this.tableData.map((project: IProjectTable) => {
        return {
          ...project,
          averageGpuUtilization: getAverageGpuUtilization(project),
          averageGpuAllocation: getAverageGpuAllocation(project),
          averageGpuMemoryUtilization: getAverageGpuMemoryUtilization(project),
        };
      });
    },
    getCustomMessage(): string {
      if (this.canCreateProjects) {
        return "You don't have any projects yet.";
      }
      return "You are not assigned to any project yet.";
    },
    disablePolicyButton(): boolean {
      return this.loadingPolicies || (!this.policyMenuOptions.length && !this.gvkPolicyMenuOptions.length);
    },
    hasInteractivePolicy(): boolean {
      if (!this.selectedProjectPolicies) return false;
      const rules: WorkspacePolicyDefaultsAndRulesV2Rules =
        this.selectedProjectPolicies[PolicyType.Workspace]?.effective?.rules || {};
      const defaults: WorkspacePolicyDefaultsAndRulesV2Defaults =
        this.selectedProjectPolicies[PolicyType.Workspace]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
    hasTrainingPolicy(): boolean {
      if (!this.selectedProjectPolicies) return false;
      const rules: TrainingPolicyDefaultsAndRulesV2Rules =
        this.selectedProjectPolicies[PolicyType.Training]?.effective?.rules || {};
      const defaults: TrainingPolicyDefaultsAndRulesV2Defaults =
        this.selectedProjectPolicies[PolicyType.Training]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
    hasDistributedPolicy(): boolean {
      if (!this.selectedProjectPolicies) return false;
      const rules: DistributedPolicyRulesV2 =
        this.selectedProjectPolicies[PolicyType.Distributed]?.effective?.rules || {};
      const defaults: DistributedPolicyDefaultsV2 =
        this.selectedProjectPolicies[PolicyType.Distributed]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
    hasInferencePolicy(): boolean {
      if (!this.selectedProjectPolicies) return false;
      const rules: InferencePolicyDefaultsAndRulesV2Rules =
        this.selectedProjectPolicies[PolicyType.Inference]?.effective?.rules || {};
      const defaults: InferencePolicyDefaultsAndRulesV2Defaults =
        this.selectedProjectPolicies[PolicyType.Inference]?.effective?.defaults || {};
      return !!Object.keys(rules).length || !!Object.keys(defaults).length;
    },
    policyMenuOptions(): Array<PolicyType> {
      const policyOptions: Array<PolicyType> = [];
      if (!this.selectedProjectPolicies) return policyOptions;

      this.hasInteractivePolicy && policyOptions.push(PolicyType.Workspace);
      this.hasTrainingPolicy && policyOptions.push(PolicyType.Training);
      this.hasDistributedPolicy && policyOptions.push(PolicyType.Distributed);
      this.hasInferencePolicy && policyOptions.push(PolicyType.Inference);

      return policyOptions;
    },
    gvkPolicyMenuOptions(): IPolicySelectOptions[] {
      // Any-workload (GVK) policies defined for this project — appended to the effective-policy dropdown.
      return this.selectedProjectGvkPolicies.map((policy: GvkPolicy) => ({
        label: policy.workloadType.kind,
        value: policy.id,
        toolTip: `This policy defines ${policy.workloadType.kind} submission. Its rules derive from both the project-level policy and higher-level policies.`,
      }));
    },
    showPolicyButton(): boolean {
      return useSettingStore().isPolicyManagerEnabled;
    },
    lastCreatedProject(): IProjectTable | null {
      return this.lastCreatedEntity;
    },
    isMultiCluster(): boolean {
      return this.clusterStore.isMultiCluster;
    },
    availableEditSteps(): ISelectOption[] {
      if (!this.selectedProject?.clusterId) return [];

      const filteredSteps = filterOrgUnitSteps(PROJECT_FORM_WIZARD_STEPS, {
        nodepoolCapabilities: this.nodepoolCapabilitiesByClusterId[this.selectedProject.clusterId],
        hasPermission: (r, a) => this.permissionStore.hasPermission(r, a),
      });

      return stepsToSelectOptions(filteredSteps);
    },
  },
  methods: {
    updateFilters(filters: IProjectTableFilterBy): void {
      this.setFilterBy(filters);
      this.countProjects();
      this.checkIfOnlyDefaultNodePool();
    },
    async loadSelectedProjectPolicy(): Promise<void> {
      if (!this.selectedProject) {
        this.selectedProjectPolicies = null;
        this.selectedProjectGvkPolicies = [];
        return;
      }
      // Disable the dropdown and drop the previous project's menus for the whole switch window, so a
      // click can't open the prior project's policy while the next fetch is still in flight.
      this.loadingPolicies = true;
      this.selectedProjectPolicies = null;
      this.selectedProjectGvkPolicies = [];

      try {
        // GVK policies are listed and cached independently of the native-policy cache below.
        await this.loadSelectedProjectGvkPolicies();
        const alreadyLoaded = this.loadedPoliciesMapByProjectId[this.selectedProject.id];
        if (alreadyLoaded) {
          this.selectedProjectPolicies = alreadyLoaded;
          return;
        }
        const scopeProperties: IPolicyScopeProperties = {
          projectId: Number(this.selectedProject.id),
          scope: ScopeType.Project,
        };
        await Promise.all([
          policyService.loadPolicyByMetaAndType(PolicyType.Workspace, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.Training, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.Distributed, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.Inference, scopeProperties),
          policyService.loadPolicyByMetaAndType(PolicyType.DistributedInference, scopeProperties),
        ]).then((res) => {
          this.selectedProjectPolicies = {
            [PolicyType.Workspace]: res[0] as WorkspacePolicyV2,
            [PolicyType.Training]: res[1] as TrainingPolicyV2,
            [PolicyType.Distributed]: res[2] as DistributedPolicyV2,
            [PolicyType.Inference]: res[3] as InferencePolicyV2,
            [PolicyType.DistributedInference]: res[4] as DistributedInferencePolicyV2,
          };
          this.selectedProject?.id &&
            (this.loadedPoliciesMapByProjectId[this.selectedProject.id] = this.selectedProjectPolicies);
        });
      } catch (err: unknown) {
        console.error("Failed to load project policies", err);
      } finally {
        this.loadingPolicies = false;
      }
    },
    async loadSelectedProjectGvkPolicies(): Promise<void> {
      const projectId = this.selectedProject?.id;
      if (!projectId) {
        this.selectedProjectGvkPolicies = [];
        return;
      }
      const cached = this.loadedGvkPoliciesByProjectId[projectId];
      if (cached) {
        this.selectedProjectGvkPolicies = cached;
        return;
      }
      try {
        const { items } = await gvkPolicyService.listGvkPolicies();
        const forProject = (items ?? []).filter(
          (policy: GvkPolicy) => policy.scope === ScopeType.Project && String(policy.projectId) === String(projectId),
        );
        this.selectedProjectGvkPolicies = forProject;
        this.loadedGvkPoliciesByProjectId[projectId] = forProject;
      } catch (err: unknown) {
        // Best-effort: never let a GVK list failure block the native effective-policy dropdown.
        console.error("Failed to load any-workload (GVK) project policies", err);
        this.selectedProjectGvkPolicies = [];
      }
    },
    viewPolicy(value: PolicyType | string): void {
      if (!this.selectedProject?.id) return;
      const policyId = policyUtil.resolveSelectedPolicyId(value, ScopeType.Project, this.selectedProject.id);
      this.$router.push({
        name: POLICIES_ROUTE_NAMES.POLICY_VIEW,
        params: { id: policyId },
        query: { previousRoute: PROJECT_ROUTE_NAMES.PROJECT_INDEX },
      });
    },
    resetSelectedRows(): void {
      this.selectedProjects = [];
    },
    createProject(): void {
      const clusterId = this.isMultiCluster ? this.filterBy.clusterUuid : this.clusterStore.selectedClusterId;

      this.$router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
        query: { clusterId },
      });
    },
    editProject(stepName?: string): void {
      if (!this.selectedProject) return;

      const query: Record<string, string> = {
        clusterId: this.selectedProject.clusterId,
      };
      if (stepName) {
        query.stepName = stepName;
      }

      this.$router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_EDIT,
        params: { id: this.selectedProject.id },
        query,
      });
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    async onProjectDelete(project: IProjectTable, deleteType?: DeletionTypes, force?: boolean): Promise<void> {
      this.isDeleting = true;
      try {
        await orgUnitService.deleteProject(project.id, deleteType, force);
        this.clearLastCreated();
        await this.refreshList();

        await useAuthStore().loadUserOrgUnits();
      } catch (e) {
        console.error(e);
        this.$q.notify(alertUtil.getError(`Project ${project.name} couldn't be deleted`));
      } finally {
        this.resetSelectedRows();
        this.isDeleteModalOpen = false;
        this.isDeleting = false;
      }
    },
    displayProjectNodePools(project: IProjectTable): void {
      this.isNodePoolsModalOpen = true;
      this.nodePoolsModalOptions.header = `Node Pools Associated with Project ${project.name}`;
      this.nodePoolsModalOptions.resources = project.resources;
      this.nodePoolsModalOptions.nodePoolQuotaStatuses = project.status?.nodePoolQuotaStatuses || [];
      this.nodePoolsModalOptions.defaultNodePools = project.effective?.defaultNodePools || [];
    },
    displayAccessRuleTableModal(project: IProjectTable): void {
      this.accessRuleTableModalOptions.header = `Subjects Authorized for Project ${project.name}`;
      this.accessRuleTableModalOptions.scopeId = project.id.toString();
      this.isAccessRuleTableModalOpen = true;
    },
    closeAccessRuleTableModal(): void {
      this.isAccessRuleTableModalOpen = false;
      this.accessRuleTableModalOptions.header = "";
      this.accessRuleTableModalOptions.scopeId = undefined;
    },
    closeNodePoolsModal(): void {
      this.isNodePoolsModalOpen = false;
      this.resetNodePoolsModalOptions();
    },
    resetNodePoolsModalOptions(): void {
      this.nodePoolsModalOptions = {
        resources: [],
        entityType: EQuotaEntity.project,
        header: "",
        nodePoolQuotaStatuses: [],
        defaultNodePools: [],
      };
    },
    getRowKey(row: IProjectTable): string {
      return row.id;
    },

    openAccessRuleModal(): void {
      if (!this.selectedProject) return;
      this.accessRuleManagementModalOptions.scopeName = this.selectedProject?.name;
      this.accessRuleManagementModalOptions.scopeId = this.selectedProject.id.toString();
      this.isAccessRuleManagementModalOpen = true;
    },
    onAccessRuleCreated(accessRule: AccessRule): void {
      if (!this.selectedProject?.id) return;
      const roles = this.tableData.find((project: IProjectTable) => project.id === this.selectedProject?.id)?.rolesNames;
      this.updateCellContent(this.selectedProject, "rolesNames", [accessRule.subjectType, ...(roles || [])]);
    },
    onAccessRuleDeleted(accessRuleIndex: number): void {
      if (!this.selectedProject?.id) return;
      const roles = this.tableData.find((project: IProjectTable) => project.id === this.selectedProject?.id)?.rolesNames;
      const rolesWithoutDeleted = roles?.filter((role: string, index: number) => index !== accessRuleIndex);
      this.updateCellContent(this.selectedProject, "rolesNames", rolesWithoutDeleted);
    },
    async displayWorkloadsModal(project: IProjectTable): Promise<void> {
      this.isWorkloadsModalOpen = true;
      this.workloadsModalOptions.clusterId = project.clusterId;
      this.workloadsModalOptions.entityFilter = `${WorkloadSortFilterFields.ProjectId}${EFilterOperator.Equals}${project.id}`;
      this.workloadsModalOptions.entityName = project.name;
    },
    resetWorkloadsModalOptions(): void {
      this.workloadsModalOptions = { ...DEFAULT_WORKLOADS_MODAL_OPTIONS };
    },
    onProjectStatusClicked(project: IProjectTable): void {
      this.onProjectDelete(project, undefined, true);
    },
    handleEditAction(): void {
      if (!this.selectedProject?.clusterId) return;
      this.loadNodepoolCapabilities(this.selectedProject.clusterId);
    },
  },
  watch: {
    selectedProject: {
      handler(newVal: IProjectTable | null): void {
        if (!newVal) return;
        this.loadSelectedProjectPolicy();
      },
    },
    isWorkloadsModalOpen(isOpen: boolean): void {
      if (!isOpen) {
        this.resetWorkloadsModalOptions();
      }
    },
  },
});

// Type assertion to prevent TS7056 while maintaining type safety
export default ProjectsIndex as typeof ProjectsIndex & DefineComponent;
</script>

<style lang="scss">
ul.custom-dots {
  padding-left: 0;
  li {
    list-style: inside;
    margin: 0 5px;
  }
}
</style>
