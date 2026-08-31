<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="data-origin-section" label="Data origin" default-opened should-validate>
    <runai-section title="Set the project where the data is located">
      <scope-input
        :selected-scope="selectedScope"
        :forced-disabled-scopes="disabledScopes"
        :allowed-scopes="allowedScopes"
        @update-selected="updateScope"
        @clear="updateScope"
        :readonly="disable"
      >
      </scope-input>
      <q-tooltip v-if="disable" target=".scope-input">
        This can’t be modified when editing an existing data volume
      </q-tooltip>
    </runai-section>
    <runai-section title="Set a PVC from which to create the data volume">
      <div class="row">
        <runai-select
          ref="pvcName"
          label="PVC name"
          :options="pvcs"
          aid="origin-pvc-select"
          :model-value="modelValue.originPvcName"
          class="col-6 origin-pvc-select"
          :rules="[validationRules.validateNameSelected]"
          :loading="loadingPvcs"
          emit-value
          no-option-text="No PVCs found"
          @update:model-value="updatePvcName"
          :disable="disable"
        />
        <q-tooltip v-if="disable" target=".origin-pvc-select">
          This can’t be modified when editing an existing data volume
        </q-tooltip>
      </div>
    </runai-section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiSection } from "@/components/common/runai-section";
import { ScopeInput } from "@/components/rbac/common/scope-input";
import { RunaiSelect } from "@/components/common/runai-select";
// Models
import { type PermittedScopes, ScopeType } from "@/swagger-models/authorization-client";
import type { IOrgTreeNodeId } from "@/models/org-tree.model";
// Utils
import { validationRules } from "@/common/validation-rules";
import type { ISelectOption } from "@/models/global.model";

export interface IDataOriginModel {
  projectId: string;
  originPvcName: string;
}

export default defineComponent({
  name: "data-origin-section",
  components: {
    RunaiExpansionItem,
    RunaiSection,
    ScopeInput,
    RunaiSelect,
  },
  emits: {
    "update:model-value": (payload: IDataOriginModel) => payload,
  },
  props: {
    modelValue: {
      type: Object as PropType<IDataOriginModel>,
      required: true,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
      required: true,
    },
    loadingPvcs: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    pvcs: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    disable: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      disabledScopes: {
        [ScopeType.System]: true,
        [ScopeType.Tenant]: true,
        [ScopeType.Cluster]: true,
        [ScopeType.Department]: true,
        [ScopeType.Project]: false,
      },
      validationRules,
    };
  },
  computed: {
    selectedScope(): IOrgTreeNodeId {
      return {
        id: this.modelValue.projectId,
        type: ScopeType.Project,
      };
    },
  },
  methods: {
    updateScope(scope: IOrgTreeNodeId | undefined): void {
      this.$emit("update:model-value", {
        ...this.modelValue,
        projectId: scope?.id || "",
      });
    },
    updatePvcName(pvcName: string): void {
      this.$emit("update:model-value", {
        ...this.modelValue,
        originPvcName: pvcName,
      });
    },
  },
});
</script>
