<template>
  <section>
    <!-- Cluster Selection -->
    <runai-cluster-section
      prevent-top-padding-when-open
      disable-closing
      :cluster-id="clusterId || ''"
      entity-name="department"
      :disabled="!isNewDepartment || props.readOnly"
      @update:cluster-id="handleClusterUpdate"
    />

    <!-- Department Name & Description -->
    <department-name-section
      v-if="clusterId"
      :is-edit-mode="!!(department as Department)?.id"
      :department-name="name"
      :description="description"
      :cluster-id="clusterId"
      :read-only="props.readOnly"
      @update:department-name="handleNameUpdate"
      @update:description="handleDescriptionUpdate"
    />
  </section>
</template>

<script setup lang="ts">
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { DepartmentNameSection } from "@/components/department/department-edit-form/department-name-section";
import type { Department, DepartmentCreationRequest } from "@/swagger-models/org-unit-service-client";

// Props
export interface Props {
  department: Department | DepartmentCreationRequest;
  isNewDepartment: boolean;
  clusterId?: string | null;
  name?: string;
  description?: string;
  readOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  clusterId: null,
  name: "",
  description: "",
  readOnly: false,
});

// Emits - using update:name and update:description for v-model support
const emit = defineEmits<{
  (e: "update-scope", scope: { clusterId: string }): void;
  (e: "update:name", value: string): void;
  (e: "update:description", value: string): void;
}>();

// Methods
const handleClusterUpdate = (clusterId: string) => {
  emit("update-scope", { clusterId });
};

const handleNameUpdate = (value: string) => {
  emit("update:name", value);
};

const handleDescriptionUpdate = (value: string) => {
  emit("update:description", value);
};
</script>
