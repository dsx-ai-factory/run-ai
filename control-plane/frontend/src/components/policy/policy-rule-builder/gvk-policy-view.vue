<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-table
    class="gvk-policy-view"
    aid="gvk-rules-table"
    :rows="rows"
    :columns="columns"
    :loading="false"
    hide-pagination
    :filter-by="{ rowsPerPage: 0 }"
    no-shadow
    sticky-header
    :style="{ maxHeight }"
    disable-selection
    bordered
  />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiTable } from "@/components/common/runai-table";
import { gvkPolicyUtil } from "@/utils/gvk-policy.util";
import { EEnforcementType, type IEnforcementDraft, type IGvkPolicyViewRule } from "@/models/gvk-policy.model";
import { ENFORCEMENT_META } from "./policy-rule-builder.constant";
import { gvkPreviewPolicyColumns } from "@/table-models/policy.table-model";
import { ECustomCell, type ITableColumn } from "@/models/table.model";
import type { SourceOfRule } from "@/swagger-models/policy-service-client";
import type { OrgUnit } from "@/swagger-models/org-unit-service-client";

export interface Props {
  // Effective rules plus saved local rules blocked by an inherited rule.
  rules: IGvkPolicyViewRule[];
  orgUnits?: OrgUnit[];
  tenantName?: string;
  // sticky-header needs a max-height to scroll within.
  maxHeight?: string;
}
const props = withDefaults(defineProps<Props>(), { maxHeight: "600px", orgUnits: () => [], tenantName: "" });

interface IGvkRuleRow {
  field: string;
  specSelector: string;
  itemSelector: string;
  rule: string;
  allowOverride: boolean;
  sourceOfRule: SourceOfRule | null;
  inactiveTooltip: string;
}

const sourceColumn: ITableColumn = {
  name: "sourceOfRule",
  label: "Source",
  field: (row) => row.sourceOfRule,
  align: "left",
  display: true,
  customCell: ECustomCell.SOURCE_NAME_COL,
};

function rangeText(min: string, max: string, step?: string): string {
  const parts: string[] = [];
  if (min) parts.push(`min ${min}`);
  if (max) parts.push(`max ${max}`);
  if (step) parts.push(`step ${step}`);
  return parts.join(", ");
}

function enforcementText(enf: IEnforcementDraft): string {
  const label = enf.type ? ENFORCEMENT_META[enf.type].label : "";
  switch (enf.type) {
    case EEnforcementType.Set:
    case EEnforcementType.Default: {
      const value =
        enf.value ||
        Object.entries(enf.setParams)
          .map(([key, val]) => `${key}: ${val}`)
          .join(", ");
      return value ? `${label}: ${value}` : label;
    }
    case EEnforcementType.Required:
      return enf.value === "false" ? "Not required" : "Required";
    case EEnforcementType.Blocked:
      return enf.value === "false" ? "Not blocked" : "Blocked";
    case EEnforcementType.Match:
      return enf.value ? `${label}: ${enf.value}` : label;
    case EEnforcementType.Options:
      return `${label}: ${enf.options.join(", ")}`;
    case EEnforcementType.Length:
    case EEnforcementType.Elements:
    case EEnforcementType.Items:
      return `${label}: ${rangeText(enf.min, enf.max)}`;
    case EEnforcementType.Range:
    case EEnforcementType.QuantityRange:
      return `${label}: ${rangeText(enf.min, enf.max, enf.step)}`;
    default:
      return label;
  }
}

function escapeHtml(value: string): string {
  return value.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}

function enforcementsToHtml(enforcements: IEnforcementDraft[]): string {
  return enforcements
    .filter((enf) => enf.type)
    .map((enf) => `• ${escapeHtml(enforcementText(enf))}`)
    .join("<br/>");
}

const rows = computed<IGvkRuleRow[]>(() =>
  props.rules.map((rule) => {
    const draft = gvkPolicyUtil.apiRuleToDraft(rule);
    const params = Object.entries(draft.selectorParams)
      .map(([key, val]) => `${key}: ${val}`)
      .join(", ");
    const specSelector = draft.specSelector ? (params ? `${draft.specSelector} (${params})` : draft.specSelector) : "";
    return {
      field: draft.field ?? "",
      specSelector,
      itemSelector: draft.itemSelector ?? "",
      rule: enforcementsToHtml(draft.enforcements),
      allowOverride: rule.allowOverride ?? false,
      sourceOfRule: rule.sourceOfRule,
      inactiveTooltip: gvkPolicyUtil.inactiveRuleTooltip(rule.blockedBy, props.orgUnits, props.tenantName),
    };
  }),
);

const columns = computed<ITableColumn[]>(() => {
  const hasItemSelector = rows.value.some((row) => row.itemSelector);
  const baseColumns = hasItemSelector
    ? gvkPreviewPolicyColumns
    : gvkPreviewPolicyColumns.filter((col) => col.name !== "itemSelector");
  const overrideColumn = baseColumns.find((column) => column.name === "allowOverride");
  const ruleColumns = baseColumns.filter((column) => column.name !== "allowOverride");
  const result = overrideColumn ? [...ruleColumns, sourceColumn, overrideColumn] : [...ruleColumns, sourceColumn];
  if (rows.value.some((row) => row.inactiveTooltip)) {
    result.push({
      name: "inactiveWarning",
      label: "",
      field: () => "",
      align: "right",
      style: "width: 48px",
      display: true,
      cellIcon: (row: IGvkRuleRow) =>
        row.inactiveTooltip
          ? {
              name: "ph ph-warning-circle",
              color: "warning",
              size: "20px",
              tooltip: row.inactiveTooltip,
            }
          : undefined,
    });
  }
  return result;
});
</script>

<style lang="scss" scoped>
.gvk-policy-view {
  :deep([aid="inactiveWarning"] .q-icon) {
    color: var(--icon-warning);
  }
  :deep(.html-col) {
    white-space: normal;
    word-break: break-word;
  }
  // Field/selector cells render inside Quasar's `.row.no-wrap`; let that flex row wrap.
  :deep(.default-col) {
    .no-wrap {
      flex-wrap: wrap;
    }
    span {
      white-space: normal;
      word-break: break-word;
    }
  }
}
</style>
