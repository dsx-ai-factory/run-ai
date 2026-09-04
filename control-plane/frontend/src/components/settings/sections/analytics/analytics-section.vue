<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="settings-container" label="Analytics" :subheader="sectionSubHeader">
    <template v-if="analyticsSectionSettings.advancedGpuMetrics">
      <div class="row items-center q-mb-md">
        <toggle-setting :setting="analyticsSectionSettings.advancedGpuMetrics" />
      </div>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import ToggleSetting from "@/components/settings/toggle-setting/toggle-setting.vue";
import type { IAnalyticsSectionSettings } from "@/models/setting.model";
import { settingsUtil } from "@/utils/settings.util";

export default defineComponent({
  name: "analytics-section",
  components: {
    ToggleSetting,
    RunaiExpansionItem,
  },
  props: {
    analyticsSectionSettings: {
      type: Object as PropType<IAnalyticsSectionSettings>,
      required: true,
    },
  },
  computed: {
    sectionSubHeader(): string {
      return settingsUtil.getToggleSectionCollapsedTitle(this.analyticsSectionSettings);
    },
  },
});
</script>
