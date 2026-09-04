<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="yaml-submission-setting">
    <toggle-setting :setting="setting" emit-only @update-setting="onUpdateSetting" />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { ToggleSetting } from "@/components/settings/toggle-setting";
// Models
import type { Setting } from "@/models/setting.model";
import type { YamlSubmissionSetting } from "@/swagger-models/workloads-manager-client";

export default defineComponent({
  name: "yaml-submission-setting",
  components: {
    ToggleSetting,
  },
  emits: ["update-value"],
  props: {
    yamlSubmissionSetting: {
      type: Object as PropType<YamlSubmissionSetting>,
      required: true,
    },
  },
  computed: {
    setting(): Setting {
      return {
        key: "yaml-submission",
        value: this.yamlSubmissionSetting.enabled ? "true" : "false",
        label: "Submit supported workload types via YAML",
        description:
          "Enable to submit a range of workload types using a standard Kubernetes YAML. For more information, see the <a href='https://run-ai-docs.nvidia.com/self-hosted/workloads-in-nvidia-run-ai/submit-via-yaml' target='_blank'>Submit Supported Workload Types via YAML guide</a>.",
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
