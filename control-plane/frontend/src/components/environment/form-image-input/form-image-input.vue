<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="form-image-input">
    <section class="row">
      <policy-string-field
        :model-value="image"
        @update:model-value="updateModel($event)"
        label="Image URL"
        :rules="[isRequired ? isEmptyField : () => true]"
        no-error-icon
        class="col-8"
        aid="environment-image-input"
        :policy-rules="policyRules"
        :disable="disable"
      />
    </section>
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Cmps
import { PolicyStringField } from "@/components/common/policy-string-field";
// Types
import type { StringRules } from "@/swagger-models/assets-service-client";
// Stores
export default defineComponent({
  name: "form-image-input",
  components: {
    PolicyStringField,
  },
  emits: ["update"],
  props: {
    image: {
      type: [String, null] as PropType<string | null>,
      required: false,
    },
    policyRules: {
      type: [Object, null] as PropType<StringRules | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    isRequired: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
  },
  methods: {
    updateModel(val: string | number | null) {
      this.$emit("update", val);
    },
    isEmptyField(val: string): string | boolean {
      return !!val || "Enter a URL";
    },
  },
});
</script>
