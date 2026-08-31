<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td class="scope-col">
    <q-btn
      @click.stop="isOrgTreeModalOpen = true"
      class="btn-link no-padding"
      :ripple="false"
      flat
      color="primary"
      :label="scopeName"
      no-caps
    >
      <runai-org-tree-with-pop-over
        v-if="isOrgTreeModalOpen"
        :selected="[{ type: scopeType, id: scopeId }]"
        readonly
        @hide="isOrgTreeModalOpen = false"
      />
    </q-btn>
  </q-td>
</template>

<script lang="ts">
import type { ITableColumn } from "@/models/table.model";
import { useAuthStore } from "@/stores/auth.store";
import { Scope } from "@/swagger-models/assets-service-client";
import { defineComponent, type PropType } from "vue";
import { RunaiOrgTreeWithPopOver } from "@/components/common/runai-org-tree/with-popover";
import type { ScopeType } from "@/swagger-models/workloads-client";
import { tableDisplayFormatters } from "@/utils/table-display-formatters.util";

export default defineComponent({
  components: {
    RunaiOrgTreeWithPopOver,
  },
  props: {
    row: {
      type: Object as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      isOrgTreeModalOpen: false as boolean,
    };
  },
  computed: {
    scopeName(): string | number {
      return tableDisplayFormatters.getScopeName(this.row, this.authStore.tenant, this.authStore.orgUnitList);
    },
    scopeType(): ScopeType {
      // some of the entities have scope in meta and some have scopeType
      const scope = this.row.meta.scope || this.row.meta.scopeType;
      if (scope === Scope.System) return Scope.Tenant;
      return scope;
    },
    scopeId(): string {
      return tableDisplayFormatters.getScopeId(this.row, this.authStore.tenant);
    },
  },
});
</script>
