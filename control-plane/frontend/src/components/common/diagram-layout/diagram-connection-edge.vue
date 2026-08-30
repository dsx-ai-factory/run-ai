<template>
  <base-edge :path="path" :style="edgeStyle" />
  <circle :cx="circleX" :cy="circleY" r="4" :fill="edgeColor" />
</template>

<script setup lang="ts">
import { computed } from "vue";
import { BaseEdge, getSmoothStepPath, Position } from "@vue-flow/core";
import type { EdgeProps } from "@vue-flow/core";

const CIRCLE_RADIUS = 4;

const props = defineProps<EdgeProps>();

const edgeColor = "var(--border-default)";

const edgeStyle = computed(() => ({
  stroke: edgeColor,
  strokeWidth: 1,
}));

const circleY = computed(() => {
  if (props.targetPosition === Position.Top) return props.targetY - (CIRCLE_RADIUS + 1);
  if (props.targetPosition === Position.Bottom) return props.targetY + (CIRCLE_RADIUS + 1);
  return props.targetY;
});

const circleX = computed(() => {
  if (props.targetPosition === Position.Left) return props.targetX - (CIRCLE_RADIUS + 1);
  if (props.targetPosition === Position.Right) return props.targetX + (CIRCLE_RADIUS + 1);
  return props.targetX;
});

const path = computed(() => {
  const gap = CIRCLE_RADIUS * 2 + 1;
  let pathTargetX = props.targetX;
  let pathTargetY = props.targetY;

  if (props.targetPosition === Position.Top) pathTargetY = props.targetY - gap;
  else if (props.targetPosition === Position.Bottom) pathTargetY = props.targetY + gap;
  else if (props.targetPosition === Position.Left) pathTargetX = props.targetX - gap;
  else if (props.targetPosition === Position.Right) pathTargetX = props.targetX + gap;

  const [edgePath] = getSmoothStepPath({
    sourceX: props.sourceX,
    sourceY: props.sourceY,
    sourcePosition: props.sourcePosition,
    targetX: pathTargetX,
    targetY: pathTargetY,
    targetPosition: props.targetPosition,
  });
  return edgePath;
});
</script>
