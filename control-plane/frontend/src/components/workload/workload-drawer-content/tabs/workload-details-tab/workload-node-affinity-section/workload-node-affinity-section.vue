<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="workload-node-affinity-section">
    <div v-if="hasRequired" class="affinity-group">
      <div class="text-b-body1 b-mb-xs">Required:</div>
      <div class="b-gap-3xs column">
        <template v-for="(term, termIndex) in requiredTerms" :key="termIndex">
          <div v-if="termIndex > 0" class="text-b-body1 b-my-xs">--- OR ---</div>
          <div class="column b-gap-3xs">
            <template v-for="(expr, exprIndex) in term.matchExpressions" :key="exprIndex">
              <div v-if="exprIndex > 0" class="text-b-body1">AND</div>
              <div class="row items-center b-gap-3xs">
                <span v-if="exprIndex === 0" class="text-b-body1">(</span>
                <span class="expression-badge text-b-mono">{{ expr.key }}</span>
                <span class="text-b-body1">{{ formatOperator(expr.operator) }}</span>
                <span v-if="hasValues(expr)" class="expression-badge text-b-mono">{{ formatValues(expr.values) }}</span>
                <span v-if="isLastExpression(term, exprIndex)" class="text-b-body1">)</span>
              </div>
            </template>
          </div>
        </template>
      </div>
    </div>

    <div v-if="hasPreferred" class="affinity-group" :class="{ 'b-mt-m': hasRequired }">
      <div class="text-b-body1 b-mb-xs">Preferred:</div>
      <div class="column b-gap-3xs">
        <template v-for="(entry, entryIndex) in preferredTerms" :key="entryIndex">
          <template v-for="(expr, exprIndex) in entry.preference?.matchExpressions" :key="exprIndex">
            <div class="row items-center b-gap-3xs">
              <span class="text-b-body1">Weight {{ entry.weight }}:</span>
              <span class="expression-badge text-b-mono">{{ expr.key }}</span>
              <span class="text-b-body1">{{ formatOperator(expr.operator) }}</span>
              <span v-if="hasValues(expr)" class="expression-badge text-b-mono">{{ formatValues(expr.values) }}</span>
            </div>
          </template>
        </template>
      </div>
    </div>
  </div>
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

const requiredTerms = computed<NodeSelectorTerm[]>(() => {
  return props.nodeAffinity.required?.nodeSelectorTerms ?? [];
});

const preferredTerms = computed<PodPreferredNodeSelectorTerm[]>(() => {
  return props.nodeAffinity.preferred?.nodeSelectorTerms ?? [];
});

const hasRequired = computed(() => requiredTerms.value.length > 0);
const hasPreferred = computed(() => preferredTerms.value.length > 0);

function formatOperator(operator: MatchExpressionOperator): string {
  return operatorDisplayMap.get(operator) ?? operator;
}

function hasValues(expr: MatchExpression): boolean {
  return (
    expr.operator !== MatchExpressionOperator.Exists &&
    expr.operator !== MatchExpressionOperator.DoesNotExist &&
    !!expr.values?.length
  );
}

function isLastExpression(term: NodeSelectorTerm, index: number): boolean {
  return index === (term.matchExpressions?.length ?? 0) - 1;
}

function formatValues(values?: string[]): string {
  if (!values || values.length === 0) return "";
  if (values.length === 1) return values[0];
  return `[${values.join(", ")}]`;
}
</script>

<style scoped lang="scss">
.expression-badge {
  background-color: var(--bg-secondary-default);
  border-radius: var(--space-2xs);
  padding: var(--space-3xs) var(--space-xs);
}
</style>
