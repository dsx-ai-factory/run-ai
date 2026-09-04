<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    v-bind="$props"
    should-validate
    :body-no-padding="bodyNoPadding"
    :prevent-top-padding-when-open="preventTopPaddingWhenOpen"
  >
    <template #subheader>
      <slot name="subheader" />
    </template>
    <template #content-header v-if="showContentHeader">
      <section class="text-black-54 q-px-md q-py-sm row justify-between items-center" aid="header-content">
        <div class="row items-center">
          <q-btn
            :disable="sectionDisabled"
            aid="load-existing-templates"
            flat
            padding="9px"
            @click="$emit('load-section-templates')"
          >
            <q-icon name="ph-bold ph-download-simple" size="20px" />
            <q-tooltip>Load from existing setup</q-tooltip>
          </q-btn>
          <q-separator class="q-ml-sm" vertical />
        </div>
        <div v-if="selectedName" class="row items-center">
          <template v-if="supportLinking">
            <q-separator class="q-mr-sm" vertical />
            <q-btn :disable="sectionDisabled" aid="link-btn" flat padding="9px" @click="$emit('toggle-linked')">
              <q-icon
                :name="isLinked ? 'ph-bold ph-link-simple-horizontal' : 'ph-bold ph-link-simple-horizontal-break'"
                :color="isLinked ? 'primary' : undefined"
                size="20px"
              />
              <q-tooltip>
                <div>
                  Link with <span class="text-bold">{{ selectedName }}</span
                  >.
                </div>
                <div>Changes made to the setup will be applied to this template.</div>
              </q-tooltip>
            </q-btn>
          </template>
          <q-separator class="q-mx-sm" vertical />

          <q-btn
            :disable="sectionDisabled || isLinked"
            flat
            class="q-mr-sm"
            aid="reset-btn"
            no-caps
            @click="$emit('reset')"
          >
            <q-icon name="ph-bold ph-arrow-counter-clockwise" size="20px" class="q-mr-sm" />
            <span class="q-mr-xs">Reset to</span>
            <q-item-label aid="selected-name" lines="1" style="max-width: 40ch">{{ selectedName }}</q-item-label>
            <q-tooltip v-if="isLinked">To reset, first unlink with setup</q-tooltip>
          </q-btn>
          <q-btn :disable="sectionDisabled" aid="clear-btn" flat padding="9px" @click="$emit('clear')">
            <q-icon name="ph-bold ph-x" size="20px" />
            <q-tooltip>Clear setup</q-tooltip>
          </q-btn>
        </div>
      </section>
    </template>
    <slot />
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";

export default defineComponent({
  name: "hybrid-section-wrapper",
  components: {
    RunaiExpansionItem,
  },
  emits: ["load-section-templates", "reset", "clear", "toggle-linked"],
  props: {
    selectedName: {
      type: String as PropType<string>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
    },
    subheader: {
      type: String as PropType<string>,
      required: false,
    },
    defaultOpened: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hideExpendIcon: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    showSubHeader: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    showContentHeader: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disableOpening: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    sectionInvalid: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    bodyNoPadding: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportLinking: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isLinked: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    preventTopPaddingWhenOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>
