<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    model-value
    class="data-volume-scope-modal"
    size="width-sm"
    header="Scope"
    @close="$emit('close')"
    :loading="loading"
    no-padding
  >
    <template #header> {{ dvName }} shared scopes </template>
    <template #body>
      <div class="row">
        <q-input
          :clearable="!!searchTerm"
          clear-icon="ph-bold ph-x"
          class="col-12"
          aid="data-volume-scope-search"
          v-model="searchTerm"
          placeholder="Search for a cluster, department, or project"
          :input-style="{ paddingInline: '16px' }"
        >
          <template v-slot:append>
            <q-icon name="ph-bold ph-magnifying-glass" size="20px" class="q-mr-md" />
          </template>
        </q-input>
      </div>
      <runai-org-tree :selected="selectedScopes" multiple readonly :search-term="searchTerm" />
    </template>
    <template #footer>
      <q-btn label="close" color="primary" @click="$emit('close')"></q-btn>
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
// Components
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { RunaiOrgTree } from "@/components/common/runai-org-tree";
// Models
import type { IOrgTreeNodeId } from "@/models/org-tree.model";

export default defineComponent({
  name: "data-volume-scope-modal",
  components: {
    RunaiBaseDialog,
    RunaiOrgTree,
  },
  emits: ["close"],
  props: {
    selectedScopes: {
      type: Array as PropType<IOrgTreeNodeId[]>,
      required: true,
    },
    dvName: {
      type: String as PropType<string>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      searchTerm: "",
    };
  },
});
</script>
