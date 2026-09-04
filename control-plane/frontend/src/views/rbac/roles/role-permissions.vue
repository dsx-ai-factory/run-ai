<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="role-permissions">
    <table class="permissions-table">
      <thead>
        <tr>
          <td class="permission-title">
            <span>Entity</span>
            <runai-tooltip-wrapper
              tooltip-text="An entity is a system-managed object that can be viewed, edited, created or deleted by a user based on their assigned role and scope"
              max-width="300px"
              ><q-icon class="q-pl-xs text-icon-default" name="ph-bold ph-question"
            /></runai-tooltip-wrapper>
          </td>
          <td class="permission-title" v-for="action in Actions" :key="action.id">
            <span>{{ action.title }}</span>
            <runai-tooltip-wrapper :tooltip-text="action.description" :max-width="action.tooltipWidth"
              ><q-icon class="q-pl-xs text-icon-default" name="ph-bold ph-question"
            /></runai-tooltip-wrapper>
          </td>
        </tr>
      </thead>
      <tbody>
        <tr v-for="permission in groupAndSortPermissions" :key="permission.resourceType">
          <td :aid="`${permission.resourceType}`">{{ permission.displayName }}</td>
          <td v-for="action in Actions" :key="action.id">
            <q-checkbox
              size="md"
              :model-value="permission.actions.includes(action.id)"
              disable
              :aid="getActionAid(permission.displayName, action.id)"
            />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { ResourceType, type Permission, type RoleV1 } from "@/swagger-models/authorization-client";
import { RunaiTooltipWrapper } from "@/components/common/runai-tooltip-wrapper";
import { Actions } from "@/models/roles.model";
import { excludedResourcesTypes } from "@/models/access-rule.model";
import { stringUtil } from "@/utils/string.util";
import { accessRuleUtil } from "@/utils/rbac.util/access-rule.util";

export default defineComponent({
  components: { RunaiTooltipWrapper },
  props: {
    role: {
      type: Object as PropType<RoleV1>,
      required: true,
    },
  },
  computed: {
    Actions() {
      return Actions;
    },
    permissions(): Permission[] {
      return this.role.permissions;
    },
    groupAndSortPermissions(): Permission[] {
      const allPermissions = [...this.nonDataSourcePermissions, this.dataSourcePermissions, this.pvcPermission].filter(
        Boolean,
      ) as Permission[];

      const filtered = this.filterExcludedResources(allPermissions);
      return accessRuleUtil.sortAndGroupPermissionsByGroupId(filtered);
    },
    nonDataSourcePermissions(): Permission[] {
      return this.permissions.filter(
        (p) => p.resourceType !== ResourceType.PvcAssets && !this.dataSources.has(p.resourceType),
      );
    },
    pvcPermission(): Permission | null {
      const pvcPerm = this.permissions.find((p) => p.resourceType === ResourceType.PvcAssets);
      return pvcPerm ? { ...pvcPerm, displayName: "PVC" } : null;
    },
    dataSources(): Set<string> {
      const dataSourceIdentifier = this.dataSourcePermissions?.displayName;
      const dataSourceResourcesTypes = this.permissions
        .filter((p) => p.displayName === dataSourceIdentifier && p.resourceType !== ResourceType.PvcAssets)
        .map((p) => p.resourceType);

      return new Set(dataSourceResourcesTypes);
    },
    dataSourcePermissions(): Permission | null {
      const dataSourceResMap = {
        [ResourceType.HostPathAssets]: ResourceType.HostPathAssets,
        [ResourceType.CmVolumeAssets]: ResourceType.CmVolumeAssets,
        [ResourceType.S3Assets]: ResourceType.S3Assets,
        [ResourceType.NfsAssets]: ResourceType.NfsAssets,
        [ResourceType.GitAssets]: ResourceType.GitAssets,
        [ResourceType.SecretVolumeAssets]: ResourceType.SecretVolumeAssets,
      };

      return this.permissions.find((p) => dataSourceResMap[p.resourceType as keyof typeof dataSourceResMap]) || null;
    },
  },

  methods: {
    filterExcludedResources(permissions: Permission[]): Permission[] {
      return permissions.filter((p: Permission) => {
        return !excludedResourcesTypes.has(p.resourceType);
      });
    },
    getActionAid(displayName: string, action: string): string {
      return `action-${stringUtil.slug(displayName)}-${action}`;
    },
  },
});
</script>
<style scoped lang="scss">
.role-permissions {
  .permissions-table {
    width: 100%;
    border-collapse: collapse;
    border-spacing: 0;
    td {
      border-bottom: 1px solid $black-12; /* add bottom border */
      padding: 8px;
    }
    tbody tr:last-child td {
      border-bottom: none;
    }
  }
  .permission-title {
    padding: 18px 8px 18px 8px !important;
  }
  // Review checkboxes are read-only — suppress hover indications/states on disabled checkboxes
  :deep(.q-checkbox.disabled) {
    pointer-events: none;
  }
}
</style>
