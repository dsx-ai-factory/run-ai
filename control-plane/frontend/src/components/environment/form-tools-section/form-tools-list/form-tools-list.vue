<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section title="Set the connection for your tool(s)" class="form-tools-list">
    <template v-if="tools.length">
      <tool-box
        v-for="tool in tools"
        :key="tool.id"
        :form-type="formType"
        :tool="tool"
        :tool-names="toolNames"
        :tools="tools"
        :removeable="!sectionDisabled"
        :disable="sectionDisabled"
        :cluster-id="clusterId"
        @removed="removeTool(tool.id)"
        @update-tool="updateTool($event, tool.id)"
        :access-disabled="accessDisabled"
        :tool-disabled-info="toolsDisabledInfo"
        :policy-rules="policyRules"
        :policy-defaults="policyDefaults"
      ></tool-box>
    </template>
    <span>
      <q-btn
        :disable="sectionDisabled"
        flat
        aid="add-tool-button"
        color="primary"
        class="btn-icon-sm"
        icon="ph-bold ph-plus"
        label="tool"
        @click="addTool"
      ></q-btn>
      <q-tooltip v-if="sectionDisabled && toolsDisabledInfo" max-width="350px"> {{ toolsDisabledInfo }} </q-tooltip>
    </span>
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Services
import { environmentService } from "@/services/control-plane/environment.service/environment.service";
// Cmps
import type { EToolBoxFormType } from "@/components/environment/form-tools-section/tool-box";
import { ToolBox } from "@/components/environment/form-tools-section/tool-box";
// Models
import type { IUIConnection } from "@/models/environment.model";
import { RunaiSection } from "@/components/common/runai-section";
import type { ExposedUrlsRules, ExposedUrlsDefaults } from "@/swagger-models/policy-service-client";

export default defineComponent({
  name: "form-tools-list",
  components: {
    RunaiSection,
    ToolBox,
  },
  emits: ["update"],
  props: {
    tools: {
      type: Array as PropType<Array<IUIConnection>>,
      required: true,
      default: () => [],
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    formType: {
      type: String as PropType<EToolBoxFormType>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    accessDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    toolsDisabledInfo: {
      type: String as PropType<string>,
      required: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<ExposedUrlsRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<ExposedUrlsDefaults | null>,
      required: false,
    },
  },
  computed: {
    toolNames(): Array<string> {
      return this.tools.map((tool: IUIConnection): string => tool.name);
    },
    validSection(): boolean {
      return environmentService.validateUIConnections(this.tools);
    },
  },
  methods: {
    addTool(): void {
      const newTool: IUIConnection = environmentService.getConnectionModel();
      this.updateTools([...this.tools, newTool]);
    },
    removeTool(toolId: string): void {
      this.updateTools(this.tools.filter((tool: IUIConnection) => tool.id !== toolId));
    },
    updateTool(newTool: IUIConnection, toolId: string): void {
      if (!this.tools.find((tool: IUIConnection) => tool.id === toolId)) {
        this.updateTools([...this.tools, newTool]);
      } else {
        this.updateTools(this.tools.map((tool: IUIConnection) => (tool.id !== toolId ? tool : newTool)));
      }
    },
    updateTools(toolsToUpdate: IUIConnection[]): void {
      this.$emit("update", toolsToUpdate);
    },
  },
});
</script>
