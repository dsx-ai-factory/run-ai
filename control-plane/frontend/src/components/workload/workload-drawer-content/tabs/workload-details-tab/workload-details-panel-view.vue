<template>
  <runai-details-panel :key="selectedSpec || 'default'" :sections="accordionSections">
    <template v-if="showSpecSelect" #controls-prepend>
      <runai-select
        aid="workload-details-spec-select"
        :model-value="selectedSpec"
        :options="specOptions"
        emit-value
        dense
        hide-bottom-space
        outlined
        class="spec-select"
        @update:model-value="onSelectedSpecUpdate"
      />
    </template>

    <template #section-general="{ section, searchTerm }">
      <workload-general-info-section :data="section.data as IGeneralSectionData" :search-term="searchTerm" />
    </template>

    <template #section-environment="{ section, searchTerm }">
      <workload-environment-section :data="section.data as IEnvironmentSectionData" :search-term="searchTerm" />
    </template>

    <template #section-network="{ section, searchTerm }">
      <workload-network-section :data="section.data as INetworkSectionData" :search-term="searchTerm" />
    </template>

    <template #section-compute="{ section, searchTerm }">
      <workload-compute-section :data="section.data as IComputeSectionData" :search-term="searchTerm" />
    </template>

    <template #section-storage="{ section, searchTerm }">
      <workload-storage-section :data="section.data as IStorageSectionData" :search-term="searchTerm" />
    </template>

    <template #section-cli-command="{ section }">
      <workload-cli-command-section :data="section.data as ICliCommandSectionData" />
    </template>
  </runai-details-panel>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { WorkloadDetailed, WorkloadNodeAffinity } from "@/swagger-models/workloads-client";
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";
import type { CrudWorkload } from "@/services/cluster/workload.service/workload.service";
import { buildSearchableSections, type WorkloadSpec } from "@/utils/workload-details.util";
import type {
  IGeneralSectionData,
  IEnvironmentSectionData,
  INetworkSectionData,
  IComputeSectionData,
  IStorageSectionData,
  ICliCommandSectionData,
} from "@/models/workload-details.model";
import { RunaiDetailsPanel } from "@/components/common/runai-details-panel";
import { RunaiSelect } from "@/components/common/runai-select";
import type { ISelectOption } from "@/models/global.model";
import {
  WorkloadGeneralInfoSection,
  WorkloadEnvironmentSection,
  WorkloadNetworkSection,
  WorkloadComputeSection,
  WorkloadStorageSection,
  WorkloadCliCommandSection,
} from "@/components/workload/workload-drawer-content/tabs/workload-details-tab/sections";

defineOptions({ name: "workload-details-panel-view" });

const props = withDefaults(
  defineProps<{
    workloadDetailed: WorkloadDetailed;
    workloadCrud?: CrudWorkload | null;
    currentSpec?: WorkloadSpec;
    clusterName?: string;
    topologyLevels?: string[];
    nodeAffinities?: WorkloadNodeAffinity[];
    selectedSpec?: string;
    showSpecSelect?: boolean;
    specOptions?: ISelectOption[];
  }>(),
  {
    workloadCrud: null,
    currentSpec: undefined,
    clusterName: "",
    topologyLevels: () => [],
    nodeAffinities: () => [],
    selectedSpec: "",
    showSpecSelect: false,
    specOptions: () => [],
  },
);

const emit = defineEmits<{
  "update:selectedSpec": [value: string];
}>();

const currentNodeAffinity = computed<WorkloadNodeAffinity | undefined>(() => {
  if (!props.nodeAffinities.length) return undefined;
  if (props.nodeAffinities.length === 1) return props.nodeAffinities[0];
  return (
    props.nodeAffinities.find((entry) => entry.specName?.toLowerCase() === props.selectedSpec?.toLowerCase()) ??
    props.nodeAffinities[0]
  );
});

const accordionSections = computed<IDetailsPanelSectionDef[]>(() =>
  buildSearchableSections({
    workloadDetailed: props.workloadDetailed,
    workloadCrud: props.workloadCrud ?? undefined,
    currentSpec: props.currentSpec,
    clusterName: props.clusterName,
    topologyLevels: props.topologyLevels,
    nodeAffinity: currentNodeAffinity.value?.nodeAffinity,
  }),
);

function onSelectedSpecUpdate(value: string): void {
  emit("update:selectedSpec", value);
}

defineExpose({ accordionSections });
</script>

<style lang="scss" scoped>
.spec-select {
  min-width: 140px;
}
</style>
