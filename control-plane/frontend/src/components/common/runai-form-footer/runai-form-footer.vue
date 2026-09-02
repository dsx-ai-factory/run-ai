<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="row items-center form-footer-wrapper">
    <q-field :error="displayFormHint" class="form-hint no-padding">
      <template v-slot:error> {{ $options.errorMessage }} </template>
    </q-field>
    <div v-if="readOnly" class="buttons q-ml-auto">
      <q-btn class="q-px-lg" label="close" color="primary" @click="$emit('cancel')"></q-btn>
    </div>
    <div v-else class="buttons q-ml-auto">
      <q-btn aid="cancel-form-footer-btn" class="q-mx-sm" flat label="cancel" color="primary" @click="$emit('cancel')" />
      <q-btn
        :aid="saveBtnAid"
        class="q-px-lg"
        :label="saveBtnLabel"
        color="primary"
        :flat="saveBtnFlat"
        @click="$emit('save')"
        :loading="loading"
        :disable="loading"
        type="submit"
      ></q-btn>
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

import { errorMessages } from "@/common/error-message.constant";
const HINT_TIMEOUT = 15000;
export default defineComponent({
  name: "runai-form-footer",
  emits: ["save", "cancel"],
  errorMessage: errorMessages.FORM_INCOMPLETE,
  props: {
    isFormValid: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    saveBtnLabel: {
      type: String as PropType<string>,
      required: true,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    saveBtnAid: {
      type: String as PropType<string>,
      default: "save-form-footer-btn",
    },
    saveBtnFlat: {
      type: Boolean as PropType<boolean>,
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
    showHint(): void {
      this.displayFormHint = true;
      this.timeOutId && clearTimeout(this.timeOutId);
      this.timeOutId = setTimeout(() => (this.displayFormHint = false), HINT_TIMEOUT);
    },
    clearTimeout(): void {
      this.timeOutId && clearTimeout(this.timeOutId);
    },
  },
  watch: {
    isFormValid(valid: boolean) {
      if (valid) {
        this.displayFormHint = false;
      } else {
        this.showHint();
      }
    },
  },
  unmounted() {
    this.clearTimeout();
  },
});
</script>
<style scoped lang="scss">
.form-footer-wrapper {
  width: 100%;
}
.form-hint {
  width: 250px;
}
</style>
