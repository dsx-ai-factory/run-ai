<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section :class="['runai-form-wrapper', fullWidth ? 'full-width' : 'medium-width']">
    <!-- @submit.prevent blocks the native submit fired by Enter; @keydown.enter runs the custom callback first -->
    <q-form ref="formEl" @submit.prevent @keydown.enter="handleOnEnterKey" :greedy="greedy">
      <slot />
    </q-form>

    <runai-base-dialog
      v-if="modalValue"
      :model-value="modalValue"
      size="width-sm"
      @close="choseStay"
      :on-enter-callback="choseStay"
    >
      <template #header> Leave This Page? </template>

      <template #body> You have unsaved changes. If you leave, you will lose your changes. </template>

      <template #footer>
        <q-btn label="Leave" aid="leave-page-modal-cancel-btn" flat color="primary" @click="choseLeave" />
        <q-btn label="Stay On Page" aid="leave-page-modal-stay-btn" color="primary" type="submit" @click="choseStay" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

// services
import { toString } from "@/utils/string.util";
import { eventEmitter } from "@/services/infra/event-emitter.service/event-emitter.service";

// common
import { NAVIGATION_EVENTS } from "@/common/event.constant";

// model
import { LEAVE_REASON } from "./model";

export default defineComponent({
  components: {
    RunaiBaseDialog,
  },
  emits: ["leave-page"],
  props: {
    formState: {
      type: Object as PropType<object>,
      required: true,
    },
    isPageReady: {
      type: Boolean,
      required: false,
      default: true,
    },
    fullWidth: {
      type: Boolean,
      default: false,
    },
    initDirty: {
      type: Boolean,
      default: false,
    },
    enterCallback: {
      type: Function as PropType<() => void>,
      required: false,
    },
    greedy: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      initialData: "" as string,
      isDirty: this.initDirty as boolean,
      modalValue: false as boolean,
      eventEmitterlistenerId: "" as string,
      action: null as LEAVE_REASON | null,
      formEl: null as HTMLFormElement | null,
    };
  },
  created() {
    this.eventEmitterlistenerId = eventEmitter.addListener(
      NAVIGATION_EVENTS.LEAVE_PAGE_REQUEST,
      this.beforeLeave.bind(this, LEAVE_REASON.CHANGE_ROUTE),
    );
  },
  mounted() {
    this.formEl = this.$refs.formEl as HTMLFormElement;
    window.addEventListener("beforeunload", this.beforeRefresh);
  },
  methods: {
    updateDirtyState(): void {
      const currentData: string = toString(this.formState);
      this.isDirty = this.initialData !== currentData || this.initDirty;
    },
    async beforeRefresh(event: Event): Promise<void> {
      // Cancel the event as stated by the standard.
      // Chrome requires returnValue to be set.

      if (this.isDirty) {
        event.preventDefault();
      }
    },
    beforeLeave(action: LEAVE_REASON = LEAVE_REASON.LEAVE_PAGE): void {
      this.action = action;
      if (this.isDirty) {
        this.modalValue = true;
      } else {
        this.choseLeave();
      }
    },
    choseLeave(): void {
      this.modalValue = false;
      this.sendDecision(true);
    },
    choseStay(): void {
      this.modalValue = false;
      this.sendDecision(false);
    },
    sendDecision(allowToLeave: boolean): void {
      eventEmitter.emit(NAVIGATION_EVENTS.LEAVE_PAGE_STATUS, { allowToLeave, action: this.action });
      if (allowToLeave) {
        this.$emit("leave-page", this.action);
      }
    },
    initFirstState(): void {
      this.initialData = toString(this.formState);
    },
    validate(): Promise<boolean> {
      return this.formEl && this.formEl.validate();
    },
    handleOnEnterKey(): void {
      if (this.enterCallback) {
        this.enterCallback();
      }
    },
  },
  beforeUnmount(): void {
    eventEmitter.removeListener(NAVIGATION_EVENTS.LEAVE_PAGE_REQUEST, this.eventEmitterlistenerId);

    window.removeEventListener("beforeunload", this.beforeRefresh);
  },
  watch: {
    formState: {
      handler(): void {
        this.isPageReady && this.updateDirtyState();
      },
      deep: true,
    },

    initDirty: {
      handler(newValue: boolean): void {
        this.isDirty = newValue;
      },
    },

    isPageReady: {
      handler(isReady: boolean): void {
        isReady && !this.initialData && this.initFirstState();
      },
      immediate: true,
    },
  },
});
</script>

<style lang="scss" scoped>
.medium-width {
  width: 800px;
}
.full-width {
  width: 100%;
}
</style>
