<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="inference-server-section"
    aid="inference-server-section"
    label="Inference server"
    :section-invalid="sectionInvalid"
    :default-opened="isSelectable"
    :disable-opening="!isSelectable"
    :disable-closing="isSelectable"
    hide-expend-icon
  >
    <template #subheader>
      <div class="row items-center justify-between q-pr-sm" aid="inference-server-section-summary">
        <span>{{ summary }}</span>
      </div>
    </template>

    <runai-radio-options
      title="Set the inference server"
      :model-value="modelValue"
      :options="serverOptions"
      @update:model-value="onSelectionChanged"
      tooltip-width="273px"
      title-tooltip-width="273px"
    />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";

// models
import type { ISelectOption } from "@/models/global.model";
import { EServerType } from "@/models/inference.model";
import type { TInferenceServerSelection } from "@/models/inference.model";

// stores
import { useClusterStore } from "@/stores/cluster.store";
import { useSettingStore } from "@/stores/setting.store";

import { MIN_CLUSTER_VERSION_FOR_NIM } from "@/common/version.constant";
import { allWorkloadsIconsMap } from "@/common/icons.constant";
import { SettingKeys } from "@/models/setting.model";

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiRadioOptions,
  },
  emits: ["update-section-model"],
  props: {
    modelValue: {
      type: String as PropType<TInferenceServerSelection>,
      required: true,
    },
    isSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      settingStore: useSettingStore(),
      serverOptions: [] as Array<ISelectOption>,
    };
  },
  created() {
    this.buildServerOptions();
  },
  computed: {
    sectionInvalid(): boolean {
      return false;
    },
    summary(): string {
      switch (this.modelValue) {
        case EServerType.VLLM:
          return EServerType.VLLM;
        case EServerType.TGI:
          return EServerType.TGI;
        case EServerType.Nim:
          return "NVIDIA NIM";
        default:
          return "Standard";
      }
    },
  },
  methods: {
    buildServerOptions(): void {
      const options: Array<ISelectOption> = [];
      const isNimSupported =
        this.clusterStore.isClusterVersionSufficient(this.clusterId, MIN_CLUSTER_VERSION_FOR_NIM) &&
        this.settingStore.isFeatureEnabled(SettingKeys.EnableNIM);
      if (isNimSupported) {
        options.push({
          icon: allWorkloadsIconsMap.NvidiaNim,
          value: EServerType.Nim,
          label: "NVIDIA NIM",
          aid: "nim-option",
        });
      }
      if (this.settingStore.isFeatureEnabled(SettingKeys.EnableModelCatalog)) {
        options.push(
          { aid: "vllm-option", icon: allWorkloadsIconsMap.VLLM, value: EServerType.VLLM, label: EServerType.VLLM },
          { aid: "tgi-option", icon: allWorkloadsIconsMap.TGI, value: EServerType.TGI, label: EServerType.TGI },
        );
      }
      options.push({
        aid: "custom-option",
        value: EServerType.Custom,
        label: "Custom",
        toolTip:
          "Use an inference server not listed here, such as SGLang. This requires a custom container image and manual server configuration.",
      });
      this.serverOptions = options;
    },
    onSelectionChanged(selection: TInferenceServerSelection): void {
      this.$emit("update-section-model", selection);
    },
  },
});
</script>
<style lang="scss" scoped></style>
