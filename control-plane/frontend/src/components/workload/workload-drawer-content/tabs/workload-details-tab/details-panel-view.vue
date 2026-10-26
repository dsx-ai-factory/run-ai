<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-default-loader v-if="loading" />

  <runai-drawer-no-data
    v-else-if="!sections.length"
    :messages="emptyMessages"
    icon-name="ph-bold ph-file"
    icon-class="text-icon-default"
    wrapper-padding-class="q-px-md q-pb-xl"
  />

  <runai-details-panel v-else :key="selectedSpec || 'default'" :sections="sections">
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
import type { IDetailsPanelSectionDef } from "@/models/details-panel.model";
import type {
  IGeneralSectionData,
  IEnvironmentSectionData,
  INetworkSectionData,
  IComputeSectionData,
  IStorageSectionData,
  ICliCommandSectionData,
} from "@/models/workload-details.model";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { RunaiDetailsPanel } from "@/components/common/runai-details-panel";
import { RunaiDrawerNoData } from "@/components/common/runai-drawer-no-data";
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

defineOptions({ name: "details-panel-view" });

// Pure renderer over the accordion-section model — it has no idea whether the sections came from
// a workload or a workload element. workload-details-tab / element-details-tab each build that
// model from their own data source (via workloadDetailsUtil / workload-element-details.util) and
// hand it here. Owns the loading + no-data states itself, since both callers need identical
// handling for them.
withDefaults(
  defineProps<{
    sections: IDetailsPanelSectionDef[];
    loading?: boolean;
    selectedSpec?: string;
    showSpecSelect?: boolean;
    specOptions?: ISelectOption[];
    /** Empty-state copy — callers own the wording, this component owns the layout. */
    emptyMessages?: string[];
  }>(),
  {
    loading: false,
    selectedSpec: "",
    showSpecSelect: false,
    specOptions: () => [],
    emptyMessages: () => ["There are no details yet."],
  },
);

const emit = defineEmits<{
  "update:selectedSpec": [value: string];
}>();

function onSelectedSpecUpdate(value: string): void {
  emit("update:selectedSpec", value);
}
</script>

<style lang="scss" scoped>
.spec-select {
  min-width: 140px;
}
</style>
