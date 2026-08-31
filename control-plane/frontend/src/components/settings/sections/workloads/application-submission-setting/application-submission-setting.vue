<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="application-submission-setting">
    <toggle-setting aid="application-submission-toggle" :setting="setting" emit-only @update-setting="onUpdateSetting" />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { ToggleSetting } from "@/components/settings/toggle-setting";
// Models
import type { Setting } from "@/models/setting.model";
import { AlphaStage } from "@/models/setting.model";
import type { IApplicationSubmissionSetting } from "@/services/cluster/workload.service/workload.service";

export default defineComponent({
  name: "application-submission-setting",
  components: {
    ToggleSetting,
  },
  emits: ["update-value"],
  props: {
    applicationSubmissionSetting: {
      type: Object as PropType<IApplicationSubmissionSetting>,
      required: true,
    },
  },
  computed: {
    setting(): Setting {
      return {
        key: "application-submission",
        value: this.applicationSubmissionSetting.enabled ? "true" : "false",
        label: "AI application submission",
        description:
          "Enable to deploy agentic AI application stacks from the NGC catalog or Helm repositories, with resource overrides for multi-component pipelines. For more information, see the <a href='https://run-ai-docs.nvidia.com/self-hosted/ai-applications/ai-applications' target='_blank' rel='noopener noreferrer'>AI Applications guide</a>.",
        stage: AlphaStage,
      };
    },
  },
  methods: {
    onUpdateSetting(_: string, value: boolean): void {
      this.$emit("update-value", value);
    },
  },
});
</script>
