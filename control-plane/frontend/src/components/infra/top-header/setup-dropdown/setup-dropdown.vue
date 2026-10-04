<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <span>
    <q-btn
      class="setup-dropdown runai-icon-btn--default"
      flat
      round
      icon="ph-bold ph-folder-simple-dotted"
      no-icon-animation
      aid="setup-dropdown"
    >
      <q-menu
        :class="setupDropdownMenuLayout"
        class="setup-dropdown-menu-container"
        anchor="bottom middle"
        self="top middle"
      >
        <q-list>
          <q-item>
            <q-item-section>
              <q-item-label>Setup</q-item-label>
            </q-item-section>
          </q-item>

          <q-item
            :clickable="allowToOpenGeneralSettings"
            v-close-popup
            @click="itemClicked(routes.settings)"
            :class="{ disabled: !allowToOpenGeneralSettings }"
            aid="general-settings-btn"
          >
            <q-item-section thumbnail class="icon-container">
              <q-icon size="16px" name="ph-bold ph-gear" />
            </q-item-section>
            <q-item-section>
              <q-item-label>General</q-item-label>
            </q-item-section>
          </q-item>

          <q-item
            :clickable="allowOpenUsers"
            v-close-popup
            @click="itemClicked(routes.users)"
            :class="{ disabled: !allowOpenUsers }"
            aid="users-btn"
          >
            <q-item-section thumbnail class="icon-container">
              <q-icon size="16px" name="ph-bold ph-user" />
            </q-item-section>
            <q-item-section>
              <q-item-label>Users</q-item-label>
            </q-item-section>
          </q-item>

          <q-item
            :clickable="allowOpenApplications"
            v-close-popup
            @click="itemClicked(routes.applications)"
            :class="{ disabled: !allowOpenApplications }"
            aid="applications-btn"
          >
            <q-item-section thumbnail class="icon-container">
              <q-icon size="16px" name="ph-bold ph-game-controller" />
            </q-item-section>
            <q-item-section>
              <q-item-label>Applications</q-item-label>
            </q-item-section>
          </q-item>

          <q-item
            :clickable="allowOpenAccessRules"
            v-close-popup
            @click="itemClicked(routes.accessRules)"
            :class="{ disabled: !allowOpenAccessRules }"
            aid="access-rules-and-roles-btn"
          >
            <q-item-section thumbnail class="icon-container">
              <q-icon size="16px" name="runai-filled-access-rule" />
            </q-item-section>
            <q-item-section>
              <q-item-label>Access rules & Roles</q-item-label>
            </q-item-section>
          </q-item>

          <q-item
            :clickable="allowPoliciesRules"
            v-close-popup
            @click="itemClicked(routes.policies)"
            :class="{ disabled: !allowPoliciesRules }"
            aid="policies-btn"
            v-if="showPolicies"
          >
            <q-item-section thumbnail class="icon-container">
              <q-icon size="16px" name="ph-bold ph-clipboard-text" />
            </q-item-section>
            <q-item-section>
              <q-item-label>Workload policies</q-item-label>
            </q-item-section>
          </q-item>
        </q-list>
      </q-menu>
    </q-btn>
    <q-tooltip class="q-pa-sm tooltip-dark">
      <div>Tools & Settings</div>
    </q-tooltip>
  </span>
</template>

<script lang="ts">
import { defineComponent } from "vue";

import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";
import { useAuthStore } from "@/stores/auth.store";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { USER_ROUTES_NAMES } from "@/router/users.routes/users.routes.names";
import { SERVICE_ACCOUNTS_ROUTE_NAMES } from "@/router/service-accounts.routes/service-accounts.routes.names";
import { ROLE_ROUTE_NAMES } from "@/router/role.routes/role.routes.names";
import { POLICIES_ROUTE_NAMES } from "@/router/policy.routes/policy.routes.names";
import { useSettingStore } from "@/stores/setting.store";

export default defineComponent({
  data() {
    return {
      permissionStore: usePermissionStore(),
      clusterStore: useClusterStore(),
      authStore: useAuthStore(),
      settingStore: useSettingStore(),
      routes: {
        settings: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
        users: USER_ROUTES_NAMES.USER_INDEX,
        applications: SERVICE_ACCOUNTS_ROUTE_NAMES.SERVICE_ACCOUNTS_INDEX,
        accessRules: ROLE_ROUTE_NAMES.ACCESS_RULES_INDEX,
        policies: POLICIES_ROUTE_NAMES.POLICIES_INDEX,
      } as Record<string, string>,
    };
  },
  computed: {
    setupDropdownMenuLayout(): string {
      return "list-menu-layout";
    },
    allowToOpenGeneralSettings(): boolean {
      return this.permissionStore.hasSomePermittedActions(
        [ResourceType.SecuritySettings, ResourceType.Settings],
        [Action.Update, Action.Create],
      );
    },
    allowOpenUsers(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Users, Action.Read);
    },
    allowOpenApplications(): boolean {
      return this.permissionStore.hasPermission([ResourceType.ServiceAccount, ResourceType.Apps], Action.Read);
    },
    allowOpenAccessRules(): boolean {
      return this.permissionStore.hasPermission(ResourceType.AccessRules, Action.Read);
    },
    allowPoliciesRules(): boolean {
      return this.permissionStore.hasPermission(ResourceType.Policies, Action.Read);
    },
    showPolicies(): boolean {
      return this.clusterStore.clusterList.length > 0 && this.settingStore.isPolicyManagerEnabled;
    },
  },
  methods: {
    itemClicked(name: string): void {
      this.$router.push({ name });
    },
  },
});
</script>

<style lang="scss">
.setup-dropdown-menu-container {
  background: $dark-gray !important;
  color: $white;

  .setup-grid-menu-layout {
    display: grid;
    grid-template-columns: 1fr 10px 1fr;
    gap: 10px;
  }

  .icon-container {
    color: $white;
    padding-left: 16px;
  }

  .side-icon-container {
    color: $white;
    padding-right: 16px !important;
    padding-left: 0 !important;
  }
}
</style>
