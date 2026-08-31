<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <!-- Department Selection -->
    <project-scope-section
      :department-id="parentId || undefined"
      :cluster-id="clusterId || ''"
      :is-new-project="isNewProject"
      @update-scope="handleScopeUpdate"
    />

    <!-- Project Name & Description -->
    <project-name-section
      v-if="clusterId"
      :project-name="name"
      :description="description"
      :namespace="computedNamespace"
      :is-new-project="isNewProject"
      @update:project-name="handleProjectNameUpdate"
      @update:description="handleDescriptionUpdate"
      @update:namespace="handleNamespaceUpdate"
      :cluster-id="clusterId"
      @is-section-invalid="handleValidationChange"
    />
  </section>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { ProjectScopeSection } from "@/components/project/project-edit-form/project-scope-section";
import { ProjectNameSection } from "@/components/project/project-edit-form/project-name-section";
import type { Project, ProjectCreationRequest, Department } from "@/swagger-models/org-unit-service-client";

// Props
export interface Props {
  project: Project | ProjectCreationRequest;
  department?: Department | null;
  isNewProject: boolean;
  clusterId?: string | null;
  parentId?: string | null;
  name?: string;
  description?: string;
  namespace?: string | null | undefined;
}

const props = withDefaults(defineProps<Props>(), {
  department: null,
  clusterId: null,
  parentId: null,
  name: "",
  description: "",
  namespace: undefined,
});

// Emits
const emit = defineEmits<{
  (e: "update-scope", scope: { clusterId: string; departmentId?: string }): void;
  (e: "update:name", value: string): void;
  (e: "update:description", value: string): void;
  (e: "update:namespace", value: string | undefined): void;
  (e: "load-department", departmentId: string): void;
  (e: "validation-change", isValid: boolean): void;
}>();

// Computed
const computedNamespace = computed(() => {
  if (props.namespace) {
    return props.namespace;
  } else {
    return (props.project as Project).status?.namespace || "";
  }
});

// Methods
const handleScopeUpdate = (scope: { clusterId: string; departmentId?: string }) => {
  emit("update-scope", scope);
};

const handleProjectNameUpdate = (value: string) => {
  emit("update:name", value);
};

const handleDescriptionUpdate = (value: string) => {
  emit("update:description", value);
};

const handleNamespaceUpdate = (value: string | undefined) => {
  emit("update:namespace", value);
};

const handleValidationChange = (isSectionInvalid: boolean) => {
  emit("validation-change", !isSectionInvalid);
};
</script>

<style lang="scss" scoped>
// Wizard step structure is handled by parent wizard-layout
</style>
