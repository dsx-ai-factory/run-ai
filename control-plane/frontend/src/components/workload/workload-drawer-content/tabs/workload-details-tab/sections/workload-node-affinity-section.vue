<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-details-panel-row label="Node affinity" class="workload-node-affinity-section">
    <div class="affinity-body column">
      <div v-if="hasRequired" class="affinity-group column">
        <div class="group-title">Required:</div>
        <div class="terms column">
          <template v-for="(term, termIndex) in requiredTerms" :key="termIndex">
            <div v-if="termIndex > 0" class="or-separator">--- OR ---</div>
            <div class="term column">
              <template v-for="(expr, exprIndex) in termExpressions(term)" :key="exprIndex">
                <div v-if="exprIndex > 0" class="conjunction">AND</div>
                <div class="expression row items-center">
                  <span v-if="exprIndex === 0" class="paren">(</span>
                  <span class="text-b-mono">{{ expr.key }}</span>
                  <span class="operator">{{ expr.operator }}</span>
                  <span v-if="expr.values" class="text-b-mono">{{ expr.values }}</span>
                  <span v-if="exprIndex === termExpressions(term).length - 1" class="paren">)</span>
                </div>
              </template>
            </div>
          </template>
        </div>
      </div>

      <div v-if="hasPreferred" class="affinity-group column">
        <div class="group-title">Preferred:</div>
        <div class="terms column">
          <template v-for="(entry, entryIndex) in preferredTerms" :key="entryIndex">
            <div
              v-for="(expr, exprIndex) in preferenceExpressions(entry)"
              :key="`${entryIndex}-${exprIndex}`"
              class="expression row items-center"
            >
              <span class="weight">Weight {{ entry.weight }}:</span>
              <span class="text-b-mono">{{ expr.key }}</span>
              <span class="operator">{{ expr.operator }}</span>
              <span v-if="expr.values" class="text-b-mono">{{ expr.values }}</span>
            </div>
          </template>
        </div>
      </div>
    </div>
  </runai-details-panel-row>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type {
  PodNodeAffinity,
  MatchExpression,
  NodeSelectorTerm,
  PodPreferredNodeSelectorTerm,
} from "@/swagger-models/workloads-client";
import { MatchExpressionOperator } from "@/swagger-models/workloads-client";
import { RunaiDetailsPanelRow } from "@/components/common/runai-details-panel";

defineOptions({ name: "workload-node-affinity-section" });

export interface Props {
  nodeAffinity: PodNodeAffinity;
}

const props = defineProps<Props>();

const operatorDisplayMap = new Map<MatchExpressionOperator, string>([
  [MatchExpressionOperator.In, "in"],
  [MatchExpressionOperator.NotIn, "not in"],
  [MatchExpressionOperator.Exists, "exists"],
  [MatchExpressionOperator.DoesNotExist, "does not exist"],
  [MatchExpressionOperator.Gt, ">"],
  [MatchExpressionOperator.Lt, "<"],
]);

interface DisplayExpression {
  key: string;
  operator: string;
  /** Already-formatted value string, or undefined for operators that take no value. */
  values?: string;
}

const requiredTerms = computed<NodeSelectorTerm[]>(() => props.nodeAffinity.required?.nodeSelectorTerms ?? []);

const preferredTerms = computed<PodPreferredNodeSelectorTerm[]>(
  () => props.nodeAffinity.preferred?.nodeSelectorTerms ?? [],
);

const hasRequired = computed(() => requiredTerms.value.length > 0);
const hasPreferred = computed(() => preferredTerms.value.length > 0);

function formatOperator(operator: MatchExpressionOperator): string {
  return operatorDisplayMap.get(operator) ?? operator;
}

function formatValues(values?: string[]): string {
  if (!values || values.length === 0) return "";
  if (values.length === 1) return values[0];
  return `[${values.join(", ")}]`;
}

function exprFromMatchExpression(expr: MatchExpression): DisplayExpression {
  const hasValues =
    expr.operator !== MatchExpressionOperator.Exists &&
    expr.operator !== MatchExpressionOperator.DoesNotExist &&
    !!expr.values?.length;
  return {
    key: expr.key ?? "",
    operator: formatOperator(expr.operator),
    values: hasValues ? formatValues(expr.values ?? undefined) : undefined,
  };
}

function expressionsFromMatchLabels(matchLabels?: Record<string, string> | null): DisplayExpression[] {
  if (!matchLabels) return [];
  return Object.entries(matchLabels).map(([key, value]) => ({ key, operator: "=", values: value }));
}

function termExpressions(term: NodeSelectorTerm): DisplayExpression[] {
  const fromExpressions = (term.matchExpressions ?? []).map(exprFromMatchExpression);
  const matchLabels = (term as NodeSelectorTerm & { matchLabels?: Record<string, string> | null }).matchLabels;
  return [...fromExpressions, ...expressionsFromMatchLabels(matchLabels)];
}

function preferenceExpressions(entry: PodPreferredNodeSelectorTerm): DisplayExpression[] {
  if (!entry.preference) return [];
  return termExpressions(entry.preference as NodeSelectorTerm);
}
</script>

<style scoped lang="scss">
.workload-node-affinity-section {
  .affinity-body {
    gap: var(--space-m);
    padding-top: 3px;
  }

  .affinity-group {
    gap: var(--space-xs);
  }

  .terms {
    gap: var(--space-xs);
  }

  .term {
    gap: var(--space-2xs);
  }

  .expression {
    gap: var(--space-xs);
    flex-wrap: wrap;
  }

  .text-b-mono {
    background: var(--bg-secondary-default, #{$body-background-color});
    border-radius: var(--space-2xs);
    padding: var(--space-3xs) var(--space-xs);
  }

  .group-title,
  .paren,
  .conjunction,
  .operator,
  .or-separator,
  .weight {
    color: var(--b-text-main);
  }
}
</style>
