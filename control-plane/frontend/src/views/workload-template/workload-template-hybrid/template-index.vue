<template>
  <section class="workload-template-index">
    <runai-table-wrapper :filters-object="filterBy" sticky>
      <template v-slot:actions>
        <runai-page-actions
          primary-btn-label="new template"
          :primary-btn-disable="disableCreateBtn"
          :columns="columns"
          :filters="filterBy"
          advanced-filters
          :selected-rows-amount="selectedRowsAmount"
          @filters-changed="onFiltersChanged"
          @selected-actions-close="resetSelectedRows"
          @create-entity="createWorkloadTemplate"
          @export-csv="exportTableAsCsv"
        >
          <template v-slot:primary-btn>
            <runai-dropdown
              aid="new-template-dropdown"
              color="primary"
              icon="ph-bold ph-plus"
              class="q-mr-md"
              custom-classes="btn-icon-sm runai-icon-btn--default"
              label="New template"
              :fit="false"
              :flat="false"
              :is-disabled="disableCreateBtn"
              @on-selected="createWorkloadTemplate($event?.value)"
              :options="createTemplateOptions"
            />
          </template>
          <template v-slot:selected-rows-actions>
            <section class="row items-center">
              <runai-tooltip-wrapper :display-tooltip="disabledEdit" :tooltip-text="editBtnTooltip">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Templates, action: Action.Update }"
                  btn-action="edit"
                  @click="onEditTemplate"
                  aid="edit-template-btn"
                  :disable="disabledEdit"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disabledClone" :tooltip-text="editBtnTooltip">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Templates, action: Action.Create }"
                  btn-action="duplicate"
                  aid="copy-template-btn"
                  @click="onCloneTemplate"
                />
              </runai-tooltip-wrapper>
              <runai-tooltip-wrapper :display-tooltip="disableDelete" :tooltip-text="deleteBtnTooltip">
                <runai-action-button
                  v-permission="{ resourceType: ResourceType.Templates, action: Action.Delete }"
                  btn-action="delete"
                  aid="delete-template-btn"
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
          :rows="workloadTemplates"
          :columns="columns"
          :filter-by="filterBy"
          v-model:selected="selectedRows"
          :loading="loading"
          @update-filters="onFiltersChanged"
          :top-row="lastCreatedWorkloadTemplate"
          :get-row-key="getRowKey"
          :get-row-icon="getRowIcon"
          sticky-columns
          :bordered="false"
        >
          <template #no-data>
            <runai-table-no-data
              v-if="!loading && !lastCreatedWorkloadTemplate"
              :show-error="loadingError"
              :filter-by="filterBy"
              entity-name="template"
              icon-name="template-new"
              :disable-create-btn="!canCreateTemplate"
              @clear-filters="clearFilters"
              @create-new="createWorkloadTemplate"
              ignore-cluster-filter
            >
              <template v-slot:primary-btn="{ disabled }">
                <runai-dropdown
                  color="primary"
                  icon="ph-bold ph-plus"
                  class="q-mr-md"
                  custom-classes="btn-icon-sm runai-icon-btn--default"
                  label="New template"
                  flat
                  :is-disabled="disabled"
                  @on-selected="createWorkloadTemplate($event?.value)"
                  :options="[
                    { label: 'Workspace', value: 'workspace' },
                    { label: 'Training', value: 'training' },
                  ]"
                />
              </template>
            </runai-table-no-data>
          </template>
        </runai-table>
      </template>
    </runai-table-wrapper>

    <runai-delete-modal
      v-if="isDeleteModalOpen && selectedWorkloadTemplate"
      :deleting="isDeleting"
      :entity-name="selectedWorkloadTemplate.meta.name"
      entity-type="template"
      @cancel="isDeleteModalOpen = false"
      @delete="deleteTemplate"
    />
  </section>
</template>

<script lang="ts">
import { computed, defineComponent } from "vue";

// cmps
import { RunaiTableWrapper } from "@/components/common/runai-table-wrapper";
import { RunaiPageActions } from "@/components/common/runai-page-actions";
import { RunaiActionButton } from "@/components/common/runai-page-actions/runai-action-button";
import { RunaiDeleteModal } from "@/components/common";
import { RunaiTable } from "@/components/common";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";

// models
import { Scope } from "@/swagger-models/assets-service-client";
import type { ITableColumn } from "@/models/table.model";
import { ETableExportCsvFilesNames } from "@/models/table.model";
import type { EWorkloadType } from "@/models/workload.model";

// stores
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { useAppStore } from "@/stores/app.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";

// utils
import { tableUtil } from "@/utils/table.util";
import { alertUtil } from "@/utils/alert.util";
import { templatesUtil } from "@/utils/templates.util/templates.util";

// constants
import { Action, ResourceType, type PermittedScopes } from "@/swagger-models/authorization-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import type { IScopeModel } from "@/models/global.model";
import { EAdvancedIndexPages } from "@/composables/helpers/use-table-data.mapper";
import type { SupersetTemplate } from "@/swagger-models/workloads-client";
import { WorkloadType } from "@/swagger-models/workloads-client";
import type { TTemplateSortBy } from "@/services/control-plane/workload-template.service/workload-template.service";
import { templateIndexColumns } from "@/table-models/templates.table-model";
import type { IEnumBasedFilterModel, IFilterBy } from "@/models/filter.model";

// services
import { filterService } from "@/services/filter.service/filter.service";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { WORKLOAD_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/workload-template.routes.names";
import { RunaiDropdown } from "@/components/common/runai-dropdown";
import { workloadTemplateService } from "@/services/control-plane/workload-template.service/workload-template.service";
import { useTableDataAdvancedFilters } from "@/composables/use-table-data-advanced-filters.composable/use-table-data-advanced-filters.composable";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { distFrameworkIconsMap, toolTypeIconsMap } from "@/common/icons.constant";
import { getHuggingFaceIcon, getNimIconIfNimImage, getNvidiaImageBasedIcon } from "@/common/icon.util";
import type { DistributedFramework } from "@/swagger-models/assets-service-client";
import { INFERENCE_TEMPLATE_ROUTE_NAMES } from "@/router/workload-template.routes/inference-routes/inference-template.routes.names";
import { MIN_CLUSTER_VERSION_FOR_INFERENCE_TEMPLATE } from "@/common/version.constant";

export default defineComponent({
  components: {
    RunaiTableWrapper,
    RunaiPageActions,
    RunaiActionButton,
    RunaiDeleteModal,
    RunaiTable,
    RunaiTableNoData,
    RunaiTooltipWrapper,
    RunaiDropdown,
  },
  data() {
    return {
      workloadTemplateStore: useWorkloadTemplateStore(),
      appStore: useAppStore(),
      clusterStore: useClusterStore(),
      permissionStore: usePermissionStore(),
      authStore: useAuthStore(),
      selectedRows: [] as Array<SupersetTemplate>,
      isDeleteModalOpen: false as boolean,
      isDeleting: false as boolean,
      workloadTemplateLoading: false,
      createTemplateOptions: [
        { label: "Workspace", value: "workspace", aid: "new-template-workspace-option" },
        { label: "Training", value: "training", aid: "new-template-training-option" },
      ] as Array<{ label: string; value: string; aid: string }>,
    };
  },
  setup() {
    const columns = computed<ITableColumn[]>(() => templateIndexColumns);

    const getFlatTemplates = async (filters: IFilterBy): Promise<SupersetTemplate[]> => {
      // Because Distributed is not treated as an actual type of workload in the UI but it does in the backend we need to consider it as training when it comes to filtering
      // if the user asks to view trainings we need to show distributed templates as well
      const workloadTypeFilter = filters.advancedFilters?.find((af) => af.name === "workloadType");
      const selectedWorkloadTypeFilter = (workloadTypeFilter as IEnumBasedFilterModel)?.selectedValues as string[];
      if (selectedWorkloadTypeFilter?.includes(WorkloadType.Training)) {
        selectedWorkloadTypeFilter.push(WorkloadType.Distributed);
      }

      const { filterBy, sortBy, sortOrder, offset, limit, search } = filterService.mapColumnsFilterToFilterParams(
        filters,
        undefined,
        true,
      );

      // we should not send includeAncestors flag if no clusterId filter is present
      const includeAncestors = !!filterBy?.find((f) => f.includes("clusterId=="));

      return await workloadTemplateService.templatesList(
        offset,
        limit,
        sortOrder,
        sortBy as TTemplateSortBy,
        filterBy,
        search,
        includeAncestors,
      );
    };

    const getFlatTemplateById = async (id: string): Promise<SupersetTemplate> => {
      return await workloadTemplateService.getFlatTemplateById(id);
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
    } = useTableDataAdvancedFilters<SupersetTemplate>(
      getFlatTemplates,
      columns,
      EAdvancedIndexPages.TEMPLATES,
      getFlatTemplateById,
    );

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
    this.updateCreateTemplateOptions();
    this.appStore.setPageLoading(false);
  },
  computed: {
    Action(): typeof Action {
      return Action;
    },
    ResourceType(): typeof ResourceType {
      return ResourceType;
    },
    disableCreateBtn(): boolean {
      if (this.loading) return true;
      return !this.canCreateTemplate;
    },
    disableDelete(): boolean {
      return !this.canDeleteTemplate;
    },
    disabledEdit(): boolean {
      if (this.disableSystemTemplateEdit) return true;
      return !this.canEditTemplate;
    },
    updateTemplatePermissions(): PermittedScopes {
      return orgTreeService.getExpandedPermittedScopes(
        this.workloadTemplateStore.actionPermissionsByScopes[Action.Update] || {},
        this.authStore.orgUnitList,
      );
    },
    deleteTemplatePermissions(): PermittedScopes {
      return orgTreeService.getExpandedPermittedScopes(
        this.workloadTemplateStore.actionPermissionsByScopes[Action.Delete] || {},
        this.authStore.orgUnitList,
      );
    },
    deleteBtnTooltip(): string {
      return "You don't have permission to perform this action.";
    },
    disabledClone(): boolean {
      return !this.canCreateTemplate;
    },
    canDeleteTemplate(): boolean {
      if (!this.permissionStore.hasPermission(ResourceType.Templates, Action.Delete)) return false;
      if (!this.selectedWorkloadTemplate) return false;
      return templatesUtil.isPermittedTemplate(
        this.selectedWorkloadTemplate.meta,
        this.deleteTemplatePermissions,
        String(this.authStore.currentUser.tenantId),
      );
    },
    canEditTemplate(): boolean {
      if (!this.permissionStore.hasPermission(ResourceType.Templates, Action.Update)) return false;
      if (!this.selectedWorkloadTemplate) return false;
      return templatesUtil.isPermittedTemplate(
        this.selectedWorkloadTemplate.meta,
        this.updateTemplatePermissions,
        String(this.authStore.currentUser.tenantId),
      );
    },
    canCreateTemplate(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Templates, Action.Create);
    },
    loading(): boolean {
      return this.loadingTableData;
    },
    workloadTemplates(): Array<SupersetTemplate> {
      return this.tableData;
    },
    selectedRowsAmount(): number {
      return this.selectedRows.length;
    },
    selectedWorkloadTemplate(): SupersetTemplate | null {
      return this.selectedRows[0] || null;
    },
    selectedTemplateName(): string {
      return this.selectedRows[0]?.meta.name || "";
    },
    selectedTemplateDescription(): string {
      return this.selectedRows[0]?.meta.description || "";
    },
    lastCreatedWorkloadTemplate(): SupersetTemplate | null {
      return this.lastCreatedEntity;
    },
    noActionPermissionText(): string {
      return "Action for the selected template is not authorized";
    },
    editBtnTooltip(): string {
      return this.disableSystemTemplateEdit ? "The selected template can't be edited" : this.noActionPermissionText;
    },
    selectedScopeModel(): IScopeModel {
      if (!this.selectedRows[0]) return { scope: Scope.Tenant };
      return {
        scope: this.selectedRows[0].meta.scopeType,
        projectId: this.selectedRows[0].meta.projectId,
        departmentId: this.selectedRows[0].meta.departmentId,
        clusterId: this.selectedRows[0].meta.clusterId,
      };
    },
    disableSystemTemplateEdit(): boolean {
      return this.selectedWorkloadTemplate?.meta.scopeType == Scope.System;
    },
  },
  methods: {
    updateCreateTemplateOptions(): void {
      if (
        this.clusterStore.isClusterVersionSufficient(
          this.filterBy?.clusterUuid || "",
          MIN_CLUSTER_VERSION_FOR_INFERENCE_TEMPLATE,
        )
      ) {
        const inferenceOption = this.createTemplateOptions.find((option) => option.value === "inference");
        if (inferenceOption) return;
        this.createTemplateOptions.push({
          label: "Inference",
          value: "inference",
          aid: "new-template-inference-option",
        });
      } else {
        this.createTemplateOptions = this.createTemplateOptions.filter((option) => option.value !== "inference");
      }
    },
    getRelevantEditPage(isEdit: boolean): string {
      if (!this.selectedWorkloadTemplate) return "";

      if (this.selectedWorkloadTemplate.meta.workloadType === WorkloadType.Workspace) {
        return isEdit
          ? WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_HYBRID_EDIT
          : WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_HYBRID;
      } else if (this.selectedWorkloadTemplate.meta.workloadType === WorkloadType.Training) {
        return isEdit
          ? WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_EDIT
          : WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID;
      } else if (this.selectedWorkloadTemplate.meta.workloadType === WorkloadType.Distributed) {
        return isEdit
          ? WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED_EDIT
          : WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_HYBRID_DISTRIBUTED;
      } else if (this.selectedWorkloadTemplate.meta.workloadType === WorkloadType.Inference) {
        return isEdit
          ? INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_EDIT
          : INFERENCE_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_CUSTOM_NEW;
      }
      return "";
    },
    getRowIcon(workloadTemplate: SupersetTemplate): string {
      const defaultIcon = "environment-gray-new";

      // Check for image-based icons first (NVIDIA PyTorch, NVIDIA general, etc.)
      const image = workloadTemplate.spec?.image || workloadTemplate.combinedSpec?.image;
      const imageBasedIcon = getNvidiaImageBasedIcon(image);
      if (imageBasedIcon) {
        return imageBasedIcon;
      }

      const labels = workloadTemplate.spec?.labels || workloadTemplate.combinedSpec?.labels;

      const nimIcon = getNimIconIfNimImage(image, labels);
      if (nimIcon) return nimIcon;

      const huggingFaceIcon = getHuggingFaceIcon(workloadTemplate.meta.workloadType as EWorkloadType, image);
      if (huggingFaceIcon) return huggingFaceIcon;

      const distFramework: DistributedFramework | undefined | null =
        workloadTemplate.spec?.distributedFramework || workloadTemplate.combinedSpec?.distributedFramework;

      const connections = workloadHybridUtil.getConnectionsField(
        workloadTemplate.spec?.ports || workloadTemplate.combinedSpec?.ports || [],
        workloadTemplate.spec?.exposedUrls || workloadTemplate.combinedSpec?.exposedUrls || [],
        workloadTemplate.spec?.relatedUrls || workloadTemplate.combinedSpec?.relatedUrls || [],
      );

      if (!connections[0]?.toolType || connections[0]?.toolType.length === 0) {
        if (distFramework) {
          return distFrameworkIconsMap[distFramework];
        }
        return defaultIcon;
      }
      const toolType = connections[0].toolType;
      if (distFramework) {
        return distFrameworkIconsMap[distFramework];
      }
      return toolTypeIconsMap[toolType] || defaultIcon;
    },
    getRowKey(workloadTemplate: SupersetTemplate): string {
      return workloadTemplate.meta.id;
    },
    onCloneTemplate(): void {
      if (!this.selectedWorkloadTemplate) return;

      const query = { fromCopyId: this.selectedWorkloadTemplate.meta.id };
      const pageName = this.getRelevantEditPage(false);

      if (!pageName) return;

      this.$router.push({
        name: pageName,
        query,
      });
    },
    onEditTemplate(): void {
      const pageName = this.getRelevantEditPage(true);
      if (!pageName || !this.selectedWorkloadTemplate) return;

      this.$router.push({
        name: pageName,
        params: { id: this.selectedWorkloadTemplate.meta.id },
      });
    },
    createWorkloadTemplate(templateType: string): void {
      const query = filterService.getClusterUuidQueryFromColumnFilter(this.filterBy);
      switch (templateType) {
        case "training":
          this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_TRAINING_NEW, query });
          break;
        case "workspace":
          this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_WORKSPACE_NEW, query });
          break;
        case "inference":
          this.$router.push({ name: WORKLOAD_TEMPLATE_ROUTE_NAMES.WORKLOAD_TEMPLATE_INFERENCE_NEW, query });
          break;
      }
    },
    resetSelectedRows(): void {
      this.selectedRows = [];
    },
    openDeleteModal(): void {
      this.isDeleteModalOpen = true;
    },
    closeDeleteModal(): void {
      this.isDeleteModalOpen = false;
    },
    async deleteTemplate(): Promise<void> {
      if (!this.selectedWorkloadTemplate) return;
      try {
        this.isDeleting = true;
        await workloadTemplateService.deleteFlatTemplate(this.selectedWorkloadTemplate);

        this.clearLastCreated();
        await this.refreshList();
        this.$q.notify(alertUtil.getSuccess(`Template ${this.selectedWorkloadTemplate.meta.name} deleted`));
        this.isDeleteModalOpen = false;
        await this.workloadTemplateStore.loadWorkloadTemplates(this.filterBy);
      } catch (e: unknown) {
        console.error(e);
        this.$q.notify(alertUtil.getError(`Template ${this.selectedWorkloadTemplate.meta.name} couldn't be deleted`));
      } finally {
        this.resetSelectedRows();
        this.isDeleting = false;
      }
    },
    exportTableAsCsv(): void {
      const columns = this.columns.filter((col: ITableColumn) => this.filterBy.displayedColumns?.includes(col.name));

      tableUtil.exportTableAsCsv(ETableExportCsvFilesNames.Template, this.workloadTemplates, columns);
    },
    onFiltersChanged(filters: IFilterBy): void {
      this.setFilterBy(filters);
      this.updateCreateTemplateOptions();
    },
  },
});
</script>
