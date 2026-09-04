<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-itemized-list
    aid="generic-secret-key-value-section"
    :items-list="itemList"
    title="Set key/value pairs to store in the new secret"
    add-item-btn="KEY & VALUE"
    should-validate-value
    item-name-label="Key"
    :name-input-validation-message="errorMessages.KEY_NOT_EMPTY"
    :stack-inputs-label="false"
    prevent-delete-if-single-item
    :value-input-autogrow="true"
    :disabled-keys="disabledKeys"
    @changed="updateKeyValuePairs"
    :hide-add-item-btn="!supportMultipleItems"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import { RunaiItemizedList } from "@/components/common/runai-itemized-list";
import { errorMessages } from "@/common/error-message.constant";
import type { GenericSecretCreationSpec, KeyValuePair } from "@/swagger-models/assets-service-client";

export default defineComponent({
  name: "generic-secret-key-value-section",
  components: {
    RunaiItemizedList,
  },
  emits: ["update-secret-details"],
  props: {
    secretDetails: {
      type: Object as PropType<GenericSecretCreationSpec>,
      required: true,
    },
    supportMultipleItems: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    disabledKeys: {
      type: Array as PropType<string[]>,
      required: false,
      default: () => [],
    },
  },
  data() {
    return {
      errorMessages,
    };
  },
  computed: {
    itemList(): IItemizedListItem[] {
      if (!this.secretDetails.keyValuePairs) return [];
      return this.secretDetails.keyValuePairs.map((s: KeyValuePair): IItemizedListItem => {
        return { name: s.key || "", value: s.value || "" };
      });
    },
  },
  methods: {
    updateKeyValuePairs(updatedKeyValuePairs: Array<IItemizedListItem>): void {
      if (!updatedKeyValuePairs) return;
      this.$emit("update-secret-details", {
        ...this.secretDetails,
        keyValuePairs: updatedKeyValuePairs.map((s) => {
          return { key: s.name, value: s.value };
        }),
      });
    },
  },
});
</script>
