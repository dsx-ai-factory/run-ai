<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <!-- eslint-disable-next-line vuejs-accessibility/click-events-have-key-events, vuejs-accessibility/no-static-element-interactions -- @click.self only closes the drawer as a mouse-only convenience when clicking the dimmed background; the drawer itself already exposes an accessible close control (see @close on the drawer in consumers, e.g. training-hybrid.vue) -->
  <section
    class="workload-form-wrapper"
    :class="{ 'light-dimmed': blurred }"
    @click.self="unblur"
    @wheel="preventScrollIfBlurred"
  >
    <runai-form-wrapper
      :form-state="formState"
      :init-dirty="initDirty"
      :greedy="greedy"
      ref="workloadFormWrapper"
      @leave-page="$emit('leave-page', $event)"
      :enter-callback="beforeSubmitting"
    >
      <slot />
    </runai-form-wrapper>

    <section class="workload-form-actions-actions-container row q-mt-md">
      <div class="error-messages-container row flex-1">
        <q-field class="col-4 form-hint flex-1" :model-value="displayFormHint" :rules="[isFormIncomplete]"></q-field>
      </div>

      <div class="navigation-btns">
        <slot name="actions">
          <runai-btn
            v-if="cancelBtnText"
            flat
            color="primary"
            :label="cancelBtnText"
            class="q-mr-sm"
            @click="$emit('canceled')"
            aid="cancel-btn"
          />
          <runai-btn
            v-if="submitBtnText"
            color="primary"
            :label="submitBtnText"
            aid="submit-btn"
            :loading="submitting"
            @click="beforeSubmitting"
          />
        </slot>
      </div>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiBtn } from "@/components/common/runai-btn";

// constants
import { errorMessages } from "@/common/error-message.constant";
import { LEAVE_REASON } from "@/components/common/runai-form-wrapper";

export default defineComponent({
  components: {
    RunaiFormWrapper,
    RunaiBtn,
  },
  emits: ["canceled", "submit", "back-clicked", "leave-page", "unblur"],
  props: {
    formState: {
      type: Object as PropType<object>,
      required: true,
    },
    /**
     * @description The text for the submit button.
     * this is optional since you can use the slot to customize the button
     */
    submitBtnText: {
      type: String as PropType<string>,
      required: false,
    },
    /**
     * @description The text for the cancel button.
     * this is optional since you can use the slot to customize the button
     */
    cancelBtnText: {
      type: String as PropType<string>,
      required: false,
    },
    submitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    initDirty: {
      type: Boolean,
      required: false,
    },
    blurred: {
      type: Boolean,
      required: false,
    },
    greedy: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      displayFormHint: false as boolean,
      timeOutId: null as ReturnType<typeof setTimeout> | null,
    };
  },
  methods: {
    isFormIncomplete(val: boolean): boolean | string {
      return !val ? true : errorMessages.FORM_INCOMPLETE;
    },
    async validate(): Promise<boolean> {
      const workloadFormWrapper: HTMLFormElement | unknown = this.$refs["workloadFormWrapper"];
      if (!workloadFormWrapper) return false;

      return await (workloadFormWrapper as HTMLFormElement).validate();
    },
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), 15000);
    },
    async beforeSubmitting(): Promise<void> {
      const success = await this.validate();
      if (!success) {
        this.showHint();
        return;
      }

      this.$emit("submit");
    },
    onBack(): void {
      (this.$refs.workloadFormWrapper as HTMLFormElement).beforeLeave();
    },
    onLeavePage(leaveReason: LEAVE_REASON): void {
      if (leaveReason === LEAVE_REASON.LEAVE_PAGE) {
        this.$emit("back-clicked");
      }
    },
    unblur() {
      this.blurred && this.$emit("unblur");
    },
    preventScrollIfBlurred(ev: Event) {
      this.blurred && ev.preventDefault();
    },
  },
  unmounted() {
    this.timeOutId && clearTimeout(this.timeOutId);
  },
});
</script>
<style lang="scss" scoped>
.workload-form-wrapper {
  .workload-form-actions-actions-container {
    width: 100%;

    display: flex;
    justify-content: space-between;
  }
}
</style>
