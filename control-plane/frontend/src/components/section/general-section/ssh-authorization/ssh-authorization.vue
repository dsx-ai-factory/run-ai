<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-section title="Set the SSH key location to ensure proper MPI communication as a non-root user">
    <policy-string-field
      :model-value="modelValue"
      @update:model-value="updateSshAuthMountPath"
      label="SSH authorization mount path"
      placeholder="e.g., /home/mpiuser/.ssh"
      :rules="[isDirectoryPath]"
      :policy-rules="policyRules"
      :disable="sectionDisabled"
      stack-label
    />
  </runai-section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiSection } from "@/components/common/runai-section";
import { PolicyStringField } from "@/components/common/policy-string-field";

// validators
import { isValidDirectoryPath } from "@/common/form.validators";

// constants
import { errorMessages } from "@/common/error-message.constant";

// Types
import type { StringRules } from "@/swagger-models/assets-service-client";

export enum ESShAuthorizationEvents {
  UpdateModelValue = "update:model-value",
}

export default defineComponent({
  name: "general-section",
  components: {
    RunaiSection,
    PolicyStringField,
  },
  emits: {
    [ESShAuthorizationEvents.UpdateModelValue]: (value: string) => value,
  },
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<StringRules | null>,
      required: false,
    },
  },
  methods: {
    isDirectoryPath(val: string): boolean | string {
      if (val === null || val === "") return true;
      return isValidDirectoryPath(val) || errorMessages.INVALID_DIRECTORY_PATH;
    },
    updateSshAuthMountPath(sshAuthMountPath: string | null): void {
      this.$emit(ESShAuthorizationEvents.UpdateModelValue, sshAuthMountPath || "");
    },
  },
});
</script>
