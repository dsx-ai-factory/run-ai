<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="settings-container" label="Resources" :subheader="sectionSubHeader">
    <template v-for="setting in resourceSettingsValues" :key="setting.key">
      <toggle-setting class="q-mb-md" :setting="setting" />
      <q-separator class="q-mb-md hide-if-last-child" />
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
// components
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import ToggleSetting from "@/components/settings/toggle-setting/toggle-setting.vue";

// utils
import { settingsUtil } from "@/utils/settings.util";

// models
import type { IResourceSectionSettings, Setting } from "@/models/setting.model";

export default defineComponent({
  name: "resources-section",
  components: {
    ToggleSetting,
    RunaiExpansionItem,
  },
  props: {
    resourceSettings: {
      type: Object as PropType<IResourceSectionSettings>,
      required: true,
    },
  },
  computed: {
    sectionSubHeader(): string {
      return settingsUtil.getToggleSectionCollapsedTitle(this.resourceSettings);
    },
    resourceSettingsValues(): Setting[] {
      return Object.values(this.resourceSettings).filter(
        (iResourceSectionSetting: Setting) => iResourceSectionSetting || iResourceSectionSetting === "",
      );
    },
  },
});
</script>
