<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td class="source-name-col">
    {{ sourceToDisplay }}
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import type { IPolicyViewProperty } from "@/utils/policy.util/policy-view.util";
import type { ITableColumn } from "@/models/table.model";
import { Scope } from "@/swagger-models/assets-service-client";
import { useAuthStore } from "@/stores/auth.store";
import { capitalizeString } from "@/utils/string.util/string.util";

const NO_SOURCE_NAME = "-";

export default defineComponent({
  name: "source-name-col",
  props: {
    row: {
      type: Object as PropType<IPolicyViewProperty>,
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
    };
  },
  computed: {
    sourceToDisplay(): string {
      const scope: Scope = this.row.sourceOfRule?.scope as Scope;

      if (scope === Scope.System) return capitalizeString(Scope.System);
      if (!this.sourceName()) return NO_SOURCE_NAME;

      const capitalizedScope: string = capitalizeString(scope || "scope");
      return `${capitalizedScope} = ${this.sourceName()}`;
    },
  },
  methods: {
    sourceName(): string {
      const { scope, projectId, departmentId, clusterId } = this.row.sourceOfRule || {};
      let scopeId: string;

      switch (scope) {
        case Scope.Tenant:
          return this.authStore.tenant.name;
        case Scope.Cluster:
          scopeId = clusterId || "";
          break;
        case Scope.Department:
          scopeId = departmentId || "";
          break;
        case Scope.Project:
          scopeId = projectId?.toString() || "";
          break;
        default:
          return "";
      }
      return this.authStore.getOrgUnitByScopeAndId(scope, scopeId)?.name || "";
    },
  },
});
</script>
