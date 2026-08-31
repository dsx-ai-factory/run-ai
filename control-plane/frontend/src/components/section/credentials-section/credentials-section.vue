<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    :default-opened="defaultOpen"
    label="Credentials"
    :disable-closing="disableClosing"
    :section-invalid="sectionInvalid"
    class="credentials-section"
    aid="credentials-section"
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>
    <section class="main-section">
      <div class="row q-mb-md">Enter the token to access Hugging Face</div>
      <div class="row">
        <q-input
          aid="credentials-token-input"
          :model-value="credentials.token"
          label="Token"
          stack-label
          @update:model-value="tokenChanged"
          :rules="[isEmpty, tokenFormatValid]"
          no-error-icon
          reactive-rules
          class="col-6"
        ></q-input>
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";

// models
import type { IUIWorkloadCredentialsSectionModel } from "./credentials-section.model";
import type { EWorkloadFormType } from "@/models/workload.model";

// others
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty, haggingFaceTokenValid } from "@/common/form.validators";

import { is } from "quasar";

export default defineComponent({
  components: {
    RunaiExpansionItem,
  },
  emits: ["credentials-changed", "is-section-invalid"],
  props: {
    entityType: {
      type: String as PropType<EWorkloadFormType>,
      required: true,
    },
    credentials: {
      type: Object as PropType<IUIWorkloadCredentialsSectionModel>,
      required: true,
    },
    disableClosing: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    defaultOpen: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  computed: {
    sectionInvalid(): boolean {
      return !isNotEmpty(this.credentials.token) || !haggingFaceTokenValid(this.credentials.token);
    },
    summary(): string {
      return this.credentials.token || "None";
    },
  },
  methods: {
    isEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.ENTER_A_TOKEN;
    },
    tokenFormatValid(val: string): boolean | string {
      return haggingFaceTokenValid(val) || errorMessages.ENTER_A_VALID_TOKEN;
    },
    tokenChanged(value: string | number | null): void {
      let token = "";
      if (is.number(value)) {
        token = String(value);
      } else {
        token = value || "";
      }
      this.$emit("credentials-changed", { ...this.credentials, token });
    },
  },
  watch: {
    sectionInvalid: {
      handler(newVal: boolean): void {
        this.$emit("is-section-invalid", newVal);
      },
      immediate: true,
    },
  },
});
</script>
