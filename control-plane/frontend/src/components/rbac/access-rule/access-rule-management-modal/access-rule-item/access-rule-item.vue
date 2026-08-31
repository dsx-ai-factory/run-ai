<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <access-rule-expansion-item
    is-collapsed
    :is-admin-tenant-role="isAdminTenantRole"
    :is-inherited-from-group="isInheritedFromGroup"
    :is-inherited-all-auth="isInheritedAllAuth"
    :access-rule-subject-id="accessRule.subjectId"
    @delete="$emit('delete-rule')"
  >
    <template #summary>
      <access-rule-summary :page="page" :access-rule="accessRule" />
    </template>
  </access-rule-expansion-item>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
//cmps
import { AccessRuleExpansionItem } from "@/components/rbac/access-rule/access-rule-management-modal/access-rule-expansion-item/";
//model
import type { ISelectOption } from "@/models/global.model";
import type { IAccessRuleDisplayConfig } from "@/models/access-rule.model";
import type { EAccessRuleModalPage } from "@/models/access-rule.model";
import { AccessRuleSummary } from "@/components/rbac/access-rule/access-rule-management-modal/access-rule-summary/";
import { ERoleName } from "@/models/access-rule.model";
import { type AccessRule, ScopeType } from "@/swagger-models/authorization-client";
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  name: "access-rule-item",
  components: { AccessRuleSummary, AccessRuleExpansionItem },
  emits: ["save-rule", "delete-rule", "update", "is-valid"],
  props: {
    displayConfig: {
      type: Object as PropType<IAccessRuleDisplayConfig>,
      required: true,
    },
    subjectOptions: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    roleOptions: {
      type: Array as PropType<ISelectOption[]>,
      default: () => [],
    },
    accessRule: {
      type: Object as PropType<AccessRule>,
      required: true,
    },
    page: {
      type: String as PropType<EAccessRuleModalPage>,
      required: true,
    },
    isInheritedFromGroup: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isInheritedAllAuth: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
    };
  },
  computed: {
    isAdminTenantRole(): boolean {
      return (
        this.authStore.currentUser.username === this.accessRule.subjectId &&
        this.accessRule.roleName === ERoleName.SystemAdmin &&
        this.accessRule.scopeType === ScopeType.Tenant
      );
    },
  },
});
</script>

<style scoped lang="scss"></style>
