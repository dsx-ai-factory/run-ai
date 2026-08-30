<template>
  <runai-details-panel :sections="sections">
    <template #section-general="{ section, searchTerm }">
      <pvc-general-info-section :data="section.data as IPvcGeneralSectionData" :search-term="searchTerm" />
    </template>

    <template #section-request="{ section, searchTerm }">
      <pvc-request-section :data="section.data as IPvcRequestSectionData" :search-term="searchTerm" />
    </template>

    <template #section-storage-class="{ section, searchTerm }">
      <pvc-storage-class-section :data="section.data as IPvcStorageClassSectionData" :search-term="searchTerm" />
    </template>
  </runai-details-panel>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiDetailsPanel } from "@/components/common/runai-details-panel";
import { PvcGeneralInfoSection, PvcRequestSection, PvcStorageClassSection } from "./sections";
import { buildPvcSections } from "@/utils/pvc-details.util";
import { useAuthStore } from "@/stores/auth.store";
import type { DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import type { StorageClassSpec } from "@/swagger-models/k8s-objects-tracker-client";
import type {
  IPvcGeneralSectionData,
  IPvcRequestSectionData,
  IPvcStorageClassSectionData,
} from "@/models/pvc-details.model";

defineOptions({ name: "pvc-details-panel-view" });

const props = defineProps<{
  pvc: DatasourceListResponseEntry;
  storageClassSpec: StorageClassSpec | null;
}>();

const authStore = useAuthStore();

const sections = computed(() =>
  buildPvcSections({
    pvc: props.pvc,
    storageClassSpec: props.storageClassSpec,
    tenant: authStore.tenant,
    orgUnitList: authStore.orgUnitList,
  }),
);
</script>
