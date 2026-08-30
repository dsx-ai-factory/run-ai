<template>
  <runai-expansion-item :label="expansionLabel" default-opened prevent-top-padding-when-open disable-closing>
    <template #subheader>
      {{ selectedScope?.path || "none" }}
    </template>
    <runai-section title="Set under which department to create this project">
      <scope-input
        v-if="selectedScope"
        :readonly="!isNewProject || readonly"
        :selected-scope="selectedScope"
        tick-strategy="leaf"
        @update-selected="selectedScope = $event"
        :hidden-scope-types="hiddenScopeTypes"
        :clearable="false"
        :title="scopeInputTitle"
        :placeholder-search="placeHolderSearch"
        :field-tooltip="!isNewProject ? projectCannotBeModifiedTooltip : ''"
      />
    </runai-section>
  </runai-expansion-item>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { ScopeInput } from "@/components/rbac/common/scope-input";
import type { IOrgTreeNodeId } from "@/models/org-tree.model";
import { ScopeType } from "@/swagger-models/authorization-client";
import { orgTreeService } from "@/services/control-plane/rbac/org-tree.service/org-tree.service";
import { useAuthStore } from "@/stores/auth.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import { RunaiSection } from "@/components/common/runai-section";
import { EOrgUnitFormTooltip } from "@/models/org-unit-forms.model";

export default defineComponent({
  name: "project-scope-section",
  components: { RunaiSection, ScopeInput, RunaiExpansionItem },
  emits: ["update-scope"],
  props: {
    departmentId: {
      type: [String, null] as PropType<string | null>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    isNewProject: {
      type: Boolean,
      required: true,
    },
    readonly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      selectedScope: { id: "", type: ScopeType.Cluster } as IOrgTreeNodeId,
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      authStore: useAuthStore(),
      hiddenScopeTypes: {
        [ScopeType.System]: false,
        [ScopeType.Tenant]: false,
        [ScopeType.Cluster]: false,
        [ScopeType.Department]: false,
        [ScopeType.Project]: true,
      } as Record<ScopeType, boolean>,
    };
  },
  created() {
    this.initSelectedScope();
  },
  computed: {
    expansionLabel(): string {
      return "Department";
    },
    scopeInputTitle(): string {
      return "Departments";
    },
    placeHolderSearch(): string {
      return `Search department`;
    },
    projectCannotBeModifiedTooltip(): string {
      return EOrgUnitFormTooltip.ProjectCannotBeModified;
    },
    orgUnitList(): OrgUnit[] {
      return this.authStore.orgUnitList;
    },
  },
  methods: {
    initSelectedScope(): void {
      const ancestorId = this.departmentId ? this.departmentId : this.clusterId;
      const ancestorScopeType = ScopeType.Department;
      const selectedScope = orgTreeService.getOrgTreeNodeId(this.orgUnitList, ancestorId, ancestorScopeType);

      if (selectedScope !== null) {
        this.selectedScope = selectedScope;
      }
    },
    updateSelectedScope(): void {
      const scope = {
        clusterId: this.getClusterId(),
        departmentId: this.getDepartmentId(),
      };
      this.$emit("update-scope", scope);
    },
    getClusterId(): string | null {
      return orgTreeService.getParentClusterUuid(this.orgUnitList, this.selectedScope?.id, this.selectedScope?.type);
    },
    getDepartmentId(): string | undefined {
      return this.selectedScope?.id;
    },
  },
  watch: {
    selectedScope: {
      handler(): void {
        if (!this.isNewProject) return;
        this.updateSelectedScope();
      },
    },
  },
});
</script>

<style scoped lang="scss"></style>
