<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="role-review column items-center q-pt-md">
    <runai-form-wrapper class="role-edit-form" :form-state="roleFormState">
      <runai-expansion-item label="Role name" default-opened>
        <template #subheader>
          <div>{{ role.name }}</div>
        </template>
        <div class="row items-center role-name">
          <q-input class="col" :model-value="role.name" disable />

          <role-stage-badge :role-name="role.name || ''" />
        </div>
      </runai-expansion-item>
      <runai-expansion-item label="Permissions" default-opened>
        <template #subheader>
          <div>Authorized entities ({{ role.permissions.length }})</div>
        </template>
        <div class="q-px-xs q-py-md">
          <span>The actions that the role assignee is authorized to perform for each of the following entities</span>
        </div>
        <q-separator />
        <role-permissions v-if="role.permissions" :role="role" />
      </runai-expansion-item>
      <div class="row">
        <div class="buttons q-ml-auto q-pt-sm">
          <q-btn class="q-mx-sm" label="Close" color="primary" @click="close" aid="close-button" />
        </div>
      </div>
    </runai-form-wrapper>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// store
import { useRoleStore } from "@/stores/role.store";
import { useAppStore } from "@/stores/app.store";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";

import RolePermissions from "@/views/rbac/roles/role-permissions.vue";
// models
import { type Role } from "@/swagger-models/authorization-client";
import { RoleStageBadge } from "@/components/common/role-stage-badge";

export default defineComponent({
  components: { RoleStageBadge, RolePermissions, RunaiExpansionItem, RunaiFormWrapper },
  data() {
    return {
      appStore: useAppStore(),
      role: {} as Role,
      roleStore: useRoleStore(),
      roleFormState: {
        name: "",
        description: "",
        permissions: [],
      },
    };
  },
  async created() {
    if (this.roleStore.roleList.length === 0) {
      await this.roleStore.loadRoles();
    }
    this.role = this.roleStore.getRoleById(Number(this.$route.params.id));
    this.appStore.setPageLoading(false);
  },
  methods: {
    close() {
      this.$router.back();
    },
  },
});
</script>
<style lang="scss">
.role-review {
  .runai-expansion-header-content {
    padding-left: 4px;
  }
}
</style>
<style scoped lang="scss">
.role-review {
  .role-name {
    width: 33%;
  }
  .permissions-table {
    width: 100%;
    border-collapse: collapse;
    border-spacing: 0;
    td {
      border-bottom: 1px solid $black-12;
      padding: 8px;
    }
  }
  .buttons {
    display: flex;
    flex-direction: row;
    align-items: center;
  }
}
</style>
