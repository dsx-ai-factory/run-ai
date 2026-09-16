<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-expansion-item
    popup
    :header-class="['runai-expansion-header', 'text-weight-medium']"
    expand-separator
    :model-value="isOpen"
    @update:model-value="onModelChanged"
    :disable="disable"
    expand-icon="ph-bold ph-caret-down"
    :expand-icon-class="['runai-expansion-icon', (disableClosing || hideExpendIcon || !hasContent) && 'display-none']"
    :class="['runai-expansion-item', preventTopPaddingWhenOpen ? 'prevent-top-padding' : '']"
    :label="label"
    @after-show="$emit('after-show', $event)"
    :aid="aidToggle"
    @before-hide="validateForm"
  >
    <template #header>
      <div class="runai-expansion-header-content row">
        <div class="runai-expansion-header-content-left row items-center">
          <slot name="header">
            <div :aid="aidLabel">{{ label }}</div>
          </slot>
        </div>
        <div
          class="flex-1 runai-expansion-header-content-right justify-between text-weight-regular"
          v-if="!isOpen || showSubHeader || isInvalid || isWarning"
        >
          <div class="flex-1 subheader" :aid="aidSubheader">
            <template v-if="!isOpen || showSubHeader">
              <slot name="subheader">{{ subheader }}</slot>
            </template>
          </div>
          <q-icon v-if="isInvalid" class="q-mr-md" name="ph-bold ph-warning-circle" color="red" size="20px"></q-icon>
          <q-icon
            v-else-if="isWarning"
            class="q-mr-md"
            name="ph-bold ph-warning-circle"
            color="warning"
            size="20px"
          ></q-icon>
        </div>
      </div>
    </template>

    <template v-if="hasContent">
      <q-separator class="runai-expansion-divider" />
      <slot name="content-header" />
      <q-card class="runai-expansion-content" flat>
        <q-card-section :class="['runai-expansion-body', { 'no-padding': bodyNoPadding }]">
          <div class="loading-section" v-if="loading && !hideLoadingSkeleton">
            <q-skeleton height="230px" square></q-skeleton>
          </div>
          <template v-else>
            <q-form
              v-if="shouldValidate"
              greedy
              ref="formRef"
              @vue:updated="updateRenderKey"
              @validation-error="updateRenderKey"
              @submit.prevent
            >
              <slot />
            </q-form>
            <slot v-else />
          </template>
        </q-card-section>
        <div v-if="$slots.footer" class="runai-expansion-footer" aid="expansion-footer">
          <slot name="footer" />
        </div>
      </q-card>
    </template>
  </q-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType, ref, type Ref } from "vue";
import type { QForm, QFormChildComponent, QField } from "quasar";
import { useFormChild } from "quasar";

export default defineComponent({
  emits: ["after-show"],
  props: {
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
    disable: {
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
    sectionWarning: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    shouldValidate: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    bodyNoPadding: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideLoadingSkeleton: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    // generally used for the first item in a form
    // so the entire UI would not go down by 15px quasar automatically adds
    preventTopPaddingWhenOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      isOpen: this.defaultOpened || this.disableClosing,
      renderKey: 1000,
    };
  },
  setup(props) {
    const formRef: Ref<QForm | null> = ref(null);

    const validate = async () => {
      /**
       * ISSUE: Sections without :should-validate="true" prop have formRef = null,
       * which causes (formRef.value as QForm)?.validate() to return undefined.
       * Quasar's parent form treats undefined as a validation failure, causing
       * the entire form to fail validation even when all actual validation rules pass.
       */
      if (!props.shouldValidate) {
        return true;
      }

      const result = await (formRef.value as QForm)?.validate();
      return result;
    };

    useFormChild({
      validate,
    });

    return { formRef };
  },
  computed: {
    hasContent(): boolean {
      return !!this.$slots.default;
    },
    aidLabel(): string {
      return `${this.label}-expansion-label`.toLowerCase();
    },
    aidToggle(): string {
      return `${this.label}-expansion-toggle`.toLowerCase();
    },
    aidSubheader(): string {
      return `${this.label}-expansion-subheader`.toLowerCase();
    },
    hasError(): boolean {
      this.renderKey; // DO NOT REMOVE THIS LINE - it's critical to trigger reactivity when new components are dynamically added/removed to the QForm
      return !!(this.$refs.formRef as QForm)?.getValidationComponents().some((c: QFormChildComponent) => {
        return (c as QField).hasError;
      });
    },
    isInvalid(): boolean {
      if (this.shouldValidate) {
        return this.hasError || this.sectionInvalid;
      }
      return this.sectionInvalid || false;
    },
    isWarning(): boolean {
      return this.sectionWarning || false;
    },
  },
  methods: {
    onModelChanged(isOpen: boolean): void {
      if (this.disableOpening) {
        this.isOpen = false;
      } else if (!this.disableClosing && this.hasContent) {
        this.isOpen = isOpen;
      }
    },
    updateRenderKey() {
      this.renderKey++;
    },
    validateForm(): void {
      if (!this.shouldValidate) return;
      (this.$refs.formRef as QForm)?.validate();
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-expansion-item {
  &.q-expansion-item--popup.q-expansion-item--collapsed {
    padding-left: 0;
    padding-right: 0;
  }

  &.prevent-top-padding.q-expansion-item--popup.q-expansion-item--expanded {
    padding-top: 0;
    padding-bottom: 15px;
  }

  .runai-expansion-header {
    .runai-expansion-header-content {
      width: 100%;

      .subheader {
        line-height: 16px;
      }

      .runai-expansion-header-content-left {
        width: 370px;
        color: var(--b-text-main);
      }
      .runai-expansion-header-content-right {
        display: flex;
        align-items: center;
        color: var(--b-text-sub);
      }
    }
  }
}
</style>

<style lang="scss">
.runai-expansion-item {
  > :first-child {
    box-shadow: var(--shadow-s);
  }
  .runai-expansion-header {
    background-color: var(--bg-primary-default);
    color: var(--b-text-main);
    height: 50px;

    .runai-expansion-icon .q-icon {
      color: var(--icon-default, #707070);
      font-size: var(--icon-size-default) !important;
    }
  }
  .runai-expansion-divider {
    background: var(--border-subtle);
  }
  .runai-expansion-content {
    background-color: var(--bg-primary-default);

    .runai-expansion-body {
      padding: var(--space-l) var(--space-m);

      &.no-padding {
        padding: 0;
      }
    }
  }
  .runai-expansion-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-m);
    padding: var(--space-m);
    border-top: 1px solid var(--border-subtle);
    background-color: var(--bg-primary-default);
  }
  .loading-section {
    width: 100%;
  }
}
</style>
