<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item label="Projects" class="project-section q-pb-none" :disable-closing="disableClosing">
    <template #subheader>
      <span>{{ selectedProjectName }}</span>
    </template>
    <runai-form-card-section
      :loading="loading"
      :main-message="`Select the project in which your ${entityType} will run`"
      :cards-list="projectCards"
      :selected-cards-ids="selectedProjectsIds"
      search-name="projects"
      empty-message="Looks like you don't have any project for the selected cluster..."
      :can-create-entity="canCreateProject"
      entity-name="project"
      icon-name="ph-bold ph-suitcase"
      @create-new="createProject"
      @selected-card-changed="onSelectedProject"
      :sort-options="{ name: true, creation: true }"
      default-sort-option="name"
    >
      <template #no-roles>
        <div class="empty-message text-center text-italic q-pt-md">
          <div>Looks like you don't have any project for the selected cluster...</div>
          <div>To create a workload, you need to be assigned to a project.</div>
          <div>Contact your administrator to be assigned.</div>
        </div>
      </template>
    </runai-form-card-section>
    <div class="col-12" v-if="!loading">
      <q-field
        class="hidden-field"
        :class="{ 'override-hint-field': showOverrideHint }"
        :model-value="selectedProjectIdAsString"
        :rules="[isRequiredProject]"
        :hint="showOverrideHint ? 'Overrides YAML' : undefined"
      ></q-field>
    </div>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// stores
import { usePermissionStore } from "@/stores/permissions.store";

// Models
import { TCardCmpName, type ICardListItem } from "@/components/common/runai-card-list";
import { Action, ResourceType } from "@/swagger-models/authorization-client";
import { type ProjectMetadata } from "@/swagger-models/org-unit-service-client";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiFormCardSection } from "@/components/common";

// Utils
import { required } from "@/common/form.validators";

// Constants
import { errorMessages } from "@/common/error-message.constant";

// routes
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiFormCardSection,
  },
  props: {
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    projects: {
      type: Array as PropType<Array<ProjectMetadata>>,
      required: true,
    },
    selectedProjectId: {
      type: [Number, null] as PropType<number | null>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    showOverrideHint: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    entityType: {
      type: String as PropType<string>,
      default: "workload",
    },
  },
  data() {
    return {
      permissionStore: usePermissionStore(),
    };
  },
  computed: {
    projectCards(): Array<ICardListItem> {
      return this.projects.map((project: ProjectMetadata) => {
        const name = project.name;
        const quota = (project.deserved.gpu || 0).toString();
        const allocated = (project.allocated.gpu || 0).toString();
        const overquota = ((project.allocated.gpu || 0) - (project.deserved.gpu || 0)).toString();

        return {
          id: String(project.id),
          cardName: TCardCmpName.PROJECT,
          searchValues: [name, quota, allocated, overquota].filter((val) => !!val),
          data: project,
          sortInfo: {
            name: project.name,
            createdAt: project.createdAt,
          },
        };
      });
    },
    selectedProjectIdAsString(): string | null {
      return this.selectedProjectId !== null && this.selectedProjectId !== -1 ? String(this.selectedProjectId) : null;
    },
    selectedProjectsIds(): Array<string> {
      return this.selectedProjectId ? [String(this.selectedProjectId)] : [];
    },
    canCreateProject(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Project, Action.Create);
    },
    selectedProjectName(): string {
      if (this.selectedProjectId === null || this.selectedProjectId === -1) {
        return "None";
      }
      const project = this.projects.find((p) => +p.id === this.selectedProjectId);
      return project?.name || "None";
    },
  },
  methods: {
    onSelectedProject(selectedItems: Array<string>): void {
      selectedItems.length === 0
        ? this.$emit("project-changed", null)
        : this.$emit("project-changed", Number(selectedItems[0]));
    },
    isRequiredProject(val: string | null): boolean | string {
      return (val && required(val)) || errorMessages.SELECT_PROJECT;
    },
    createProject(): void {
      this.$emit("before-create-project");

      this.$router.push({
        name: PROJECT_ROUTE_NAMES.PROJECT_NEW,
        query: { clusterId: this.clusterId },
      });
    },
  },
});
</script>

<style lang="scss" scoped>
.q-field.hidden-field.override-hint-field :deep(.q-field__messages) {
  color: $warning;
}

.q-field.hidden-field.override-hint-field :deep(.q-field__label) {
  color: $warning;
}

.q-field.hidden-field.override-hint-field :deep(.q-field__control::before) {
  border-color: $warning;
}
</style>
