/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { defineStore } from "pinia";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import { ERoleStage, ROLE_NAME_API_MAP, ROLE_STAGE_CONFIG } from "@/models/roles.model";
import { accessRuleService } from "@/services/control-plane/rbac/access-rule.service/access-rule.service";

export const useLegacyAccessRuleStore = defineStore("LegacyAccessRule", {
  state: () => ({
    legacyRows: [] as ILegacyRoleTransitionRow[],
    loading: false as boolean,
  }),
  getters: {
    hasLegacyAccessRules(): boolean {
      return this.legacyRows.length > 0;
    },
  },
  actions: {
    async loadLegacyAccessRules(): Promise<void> {
      this.loading = true;

      try {
        const legacyRoles = Object.entries(ROLE_STAGE_CONFIG).filter(([, config]) => config.stage === ERoleStage.Legacy);

        const results = await Promise.all(
          legacyRoles.map(async ([roleName, config]) => {
            const [count, rules] = await Promise.all([
              accessRuleService.getAccessRulesCount([`roleName==${roleName}`]),
              accessRuleService.getAccessRulesV2([`roleName==${roleName}`], undefined, undefined, 0, 1),
            ]);
            return {
              id: roleName,
              legacyRole: roleName,
              roleId: rules[0]?.roleId,
              accessRuleCount: count,
              newRole: config.replacement?.replacementRole
                ? (ROLE_NAME_API_MAP[config.replacement.replacementRole] ?? config.replacement.replacementRole)
                : undefined,
            };
          }),
        );

        this.legacyRows = results.filter(
          (r) => r.accessRuleCount > 0 && r.roleId !== undefined,
        ) as ILegacyRoleTransitionRow[];
      } finally {
        this.loading = false;
      }
    },
  },
});
