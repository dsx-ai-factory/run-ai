<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="summary-wrapper text-body2">
    <div v-if="isScopePage" class="row items-center no-wrap">
      <runai-ellipsis-text-with-tooltip
        class="summary-col text-item text-b-subtitle2"
        :text="`Subject: ${formattedSubject}`"
      />
      <runai-ellipsis-text-with-tooltip
        class="summary-col text-item text-b-subtitle2"
        :text="`Role: ${accessRule.roleName}`"
      />
    </div>
    <div v-else class="row items-center no-wrap">
      <runai-ellipsis-text-with-tooltip
        class="summary-col text-item text-b-subtitle2"
        :text="`Role: ${accessRule.roleName}`"
      />
      <runai-ellipsis-text-with-tooltip class="summary-col text-item text-b-subtitle2" :text="`Scope: ${scope}`" />
      <runai-ellipsis-text-with-tooltip
        class="summary-col text-item text-b-subtitle2"
        :text="`Subject: ${formattedSubject}`"
      />
    </div>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { EAccessRuleModalPage } from "@/models/access-rule.model";
import { ALL_AUTHENTICATED_USERS_LABEL } from "@/models/access-rule.model";
import { accessRuleUtil } from "@/utils/rbac.util/access-rule.util";
import type { AccessRule } from "@/swagger-models/authorization-client";
import { SubjectTypeExtended } from "@/swagger-models/authorization-client";
import { RunaiEllipsisTextWithTooltip } from "@/components/common/runai-ellipsis-text-with-tooltip";

export default defineComponent({
  name: "access-rule-summary",
  components: { RunaiEllipsisTextWithTooltip },
  props: {
    accessRule: {
      type: Object as PropType<AccessRule>,
      required: true,
    },
    page: {
      type: String as PropType<EAccessRuleModalPage>,
      required: true,
    },
  },
  computed: {
    isScopePage(): boolean {
      return accessRuleUtil.isScopePage(this.page);
    },
    scope(): string {
      return this.accessRule.scopeName
        ? `${this.accessRule.scopeType} - ${this.accessRule.scopeName}`
        : this.accessRule.scopeType;
    },
    formattedSubject(): string {
      switch (this.accessRule.subjectType) {
        case SubjectTypeExtended.AllAuthenticatedUsers:
          return ALL_AUTHENTICATED_USERS_LABEL;
        case SubjectTypeExtended.Group:
          return `Group - ${this.accessRule.subjectId}`;
        case SubjectTypeExtended.ServiceAccount:
        case SubjectTypeExtended.App:
          return `Service account - ${this.accessRule.subjectId}`;
        default:
          return this.accessRule.subjectId;
      }
    },
  },
});
</script>

<style scoped lang="scss">
.summary-wrapper {
  padding-right: 32px;
  .text-item {
    text-overflow: ellipsis;
    margin-top: 5px;
  }
  .summary-col {
    flex: 1 1 0;
    min-width: 0;
    padding-right: 12px;
  }
}
</style>
