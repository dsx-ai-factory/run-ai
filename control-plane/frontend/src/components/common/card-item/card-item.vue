<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-card
    flat
    class="card-item text-body2"
    :class="{ 'tight-spacing': tightSpacing, 'content-center': contentCentered }"
  >
    <div class="card-wrapper">
      <div class="card-close" v-if="showDeleteButton">
        <q-btn
          v-if="permissionConfig"
          v-permission="permissionConfig"
          :disable="disabledDeleteButton"
          @click="$emit('delete')"
          flat
          size="10px"
          round
          class="runai-icon-btn--default"
          aid="card-item-delete-button"
          :icon="deleteIcon"
          aria-label="Delete"
        />
        <q-btn
          v-else
          :disable="disabledDeleteButton"
          @click="$emit('delete')"
          flat
          size="10px"
          round
          class="runai-icon-btn--default"
          aid="card-item-delete-button"
          :loading="deleting"
          :icon="deleteIcon"
          aria-label="Delete"
        />
        <q-tooltip v-if="deleteButtonTooltip" max-width="400px" anchor="bottom end">
          {{ deleteButtonTooltip }}
        </q-tooltip>
      </div>
      <div class="card-summary">
        <slot name="summary"></slot>
      </div>
      <div v-if="$slots.content" class="card-content">
        <slot name="content"></slot>
      </div>
    </div>
    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer"></slot>
    </div>
  </q-card>
</template>

<script lang="ts">
import type { Action, ResourceType } from "@/swagger-models/authorization-client";
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  name: "card-item",
  emits: ["delete"],
  props: {
    // Delete button configuration
    showDeleteButton: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disabledDeleteButton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    deleteIcon: {
      type: String as PropType<string>,
      default: "ph-bold ph-trash",
    },
    deleteButtonTooltip: {
      type: String as PropType<string>,
      default: "",
    },
    // Permission configuration for delete button
    permissionConfig: {
      type: Object as PropType<{ resourceType: ResourceType; action: Action }>,
      default: null,
    },
    // Styling options
    minHeight: {
      type: String as PropType<string>,
      default: "50px",
    },
    contentHeight: {
      type: String as PropType<string>,
      default: "280px",
    },
    collapsed: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    marginTop: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    // Spacing control
    tightSpacing: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    deleting: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    contentCentered: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>

<style lang="scss" scoped>
.card-item {
  border: 1px solid $black-12;
  border-radius: 0;
  min-height: v-bind(minHeight);
  // padding: 10px;

  // Tight spacing for list-like layouts (like access rules)
  &.tight-spacing {
    border-bottom-width: 0.5px;
    margin-bottom: -1px;

    &:last-child {
      border-bottom-width: 1px;
      margin-bottom: 0;
    }
  }

  // Regular spacing for card-like layouts
  &:not(.tight-spacing) {
    margin-bottom: 8px;
  }

  &.margin-top {
    margin-top: 15px;
  }

  .card-wrapper {
    position: relative;
    padding: 10px;

    .card-close {
      top: 50%;
      transform: translateY(-50%);
      right: 10px;
      cursor: pointer;
      position: absolute;
      z-index: 150;
      color: $black-54;
    }

    .card-content {
      transition: height 0.3s ease;
    }

    &.collapsed .card-content {
      height: v-bind(minHeight);
    }

    &:not(.collapsed) .card-content {
      height: v-bind(contentHeight);
    }
  }

  .card-footer {
    border-top: 1px solid $black-12;
    display: flex;
    justify-content: space-around;
    align-items: center;
    height: 50px;

    .error-field {
      min-width: 100px;
    }
  }
}
</style>
