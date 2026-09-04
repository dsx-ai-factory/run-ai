<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <pvc-details-panel-view :pvc="pvc" :storage-class-spec="storageClassSpec" />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { Scope, type DatasourceListResponseEntry } from "@/swagger-models/assets-service-client";
import { tableDisplayFormatters } from "@/utils/table-display-formatters.util";
import { useAuthStore } from "@/stores/auth.store";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import type { StorageClassSpec } from "@/swagger-models/k8s-objects-tracker-client";
import PvcDetailsPanelView from "./pvc-details-panel-view.vue";

export default defineComponent({
  name: "pvc-details-tab",
  components: { PvcDetailsPanelView },
  props: {
    pvc: {
      type: Object as PropType<DatasourceListResponseEntry>,
      required: true,
    },
    storageClassSpec: {
      type: [Object, null] as PropType<StorageClassSpec | null>,
      required: true,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      emptyValue: "-",
    };
  },
  computed: {
    scopeName(): string {
      const scope = this.pvc.meta.scope;
      const scopeId = tableDisplayFormatters.getScopeId(this.pvc, this.authStore.tenant);
      const isSystemScope = (type: string) => scope === Scope.System && type === Scope.Tenant;

      return (
        this.authStore.orgUnitList.find(
          (orgUnit: OrgUnit) => orgUnit.id === scopeId && (orgUnit.type === scope || isSystemScope(orgUnit.type)),
        )?.name || ""
      );
    },
    accessModes(): string {
      const accessModesObj = this.pvc.spec.pvc?.claimInfo?.accessModes;
      if (!accessModesObj || typeof accessModesObj !== "object") return this.emptyValue;
      const enabledModes = Object.entries(accessModesObj as Record<string, boolean>)
        .filter(([, value]) => value)
        .map(([key]) => key);
      return enabledModes.length > 0 ? enabledModes.join(", ") : this.emptyValue;
    },
    creationTime(): string {
      return this.pvc.meta.createdAt
        ? dateUtil.dateFormat(new Date(this.pvc.meta.createdAt), DATE_FORMATS.STANDARD_DATE)
        : this.emptyValue;
    },
    isDefaultStorageClass(): string {
      if (this.storageClassSpec?.is_default === null || this.storageClassSpec?.is_default === undefined) return "-";
      return this.storageClassSpec.is_default.toString();
    },
  },
});
</script>

<style lang="scss" scoped>
.details-section {
  padding-block: 16px;
  border-bottom: 1px solid $black-12;

  span:not(.text-subtitle1) {
    font-size: 12px;
  }
}
</style>
