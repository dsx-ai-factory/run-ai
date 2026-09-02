<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="user-settings-container"
    label="User details"
    :subheader="subheader"
    default-opened
    body-no-padding
  >
    <q-card-section>
      <user-profile-avatar :email="userEmail" :first-name="firstName" :last-name="lastName" />
      <q-separator />
    </q-card-section>
    <user-profile-access-rules-table @access-rules-fetched="updateAccessRules" />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";

// stores
import { useAuthStore } from "@/stores/auth.store";
import { UserProfileAvatar } from "@/components/user-settings/sections/user-profile-section/user-profile-avatar";
import { UserProfileAccessRulesTable } from "@/components/user-settings/sections/user-profile-section/user-profile-access-rules-table";
import type { MeAccessRule } from "@/swagger-models/authorization-client";
import { previewStrings } from "@/utils/common.util";

export default defineComponent({
  name: "user-profile-section",
  components: {
    UserProfileAccessRulesTable,
    UserProfileAvatar,
    RunaiExpansionItem,
  },
  data() {
    return {
      authStore: useAuthStore(),
      accessRules: [] as MeAccessRule[],
    };
  },
  computed: {
    userEmail(): string {
      return this.authStore.currentUser.username || "";
    },
    firstName(): string {
      return this.authStore.currentUser.firstName || "";
    },
    lastName(): string {
      return this.authStore.currentUser.lastName || "";
    },
    accessRulesList(): string {
      return previewStrings(this.accessRules.map((rule) => rule.roleName));
    },
    subheader(): string {
      return `${this.userEmail} / Role: ${this.accessRulesList}`;
    },
  },
  methods: {
    updateAccessRules(accessRules: MeAccessRule[]): void {
      this.accessRules = accessRules;
    },
  },
});
</script>
<style lang="scss" scoped>
.user-settings-container {
  margin: auto;
  width: 800px;
}
</style>
