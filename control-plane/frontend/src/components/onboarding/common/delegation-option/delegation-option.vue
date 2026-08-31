<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="delegation-option">
    <!-- Delegation Checkbox -->
    <div class="delegation-section q-mb-lg">
      <q-checkbox
        :model-value="isDelegated"
        @update:model-value="handleDelegationChange"
        :label="checkboxLabel"
        aid="delegate-checkbox"
        class="delegation-checkbox q-mb-lg"
        :disable="readOnly"
      />
      <transition name="delegation-description-transition">
        <div v-if="isDelegated" class="delegation-description q-mt-lg q-pl-sm">
          {{ descriptionText }}
        </div>
      </transition>
    </div>

    <!-- Delegation Email Input -->
    <div v-if="isDelegated" class="delegation-email-input-container q-pl-sm">
      <q-input
        :model-value="delegateEmail"
        @update:model-value="handleEmailChange"
        label="User email"
        placeholder="Enter an email"
        aid="delegate-email"
        class="delegate-email-input"
        :rules="[isValidEmail]"
        no-error-icon
        stack-label
        :disable="readOnly"
      />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { formValidators } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "delegation-option",
  emits: {
    "update:is-delegated": (value: boolean) => typeof value === "boolean",
    "update:delegate-email": (email: string) => typeof email === "string",
  },
  props: {
    isDelegated: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    delegateEmail: {
      type: String as PropType<string>,
      default: "",
    },
    checkboxLabel: {
      type: String as PropType<string>,
      default: "Let someone else take it from here",
    },
    descriptionText: {
      type: String as PropType<string>,
      default: "Provide the email of the user responsible, and we'll send them an invitation to continue.",
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  setup(_, { emit }) {
    const handleDelegationChange = (value: boolean) => {
      emit("update:is-delegated", value);
      // Clear email when disabling delegation
      if (!value) {
        emit("update:delegate-email", "");
      }
    };

    const handleEmailChange = (value: string | number | null) => {
      emit("update:delegate-email", String(value || ""));
    };

    const isValidEmail = (value: string) => {
      return formValidators.isValidEmail(value) || errorMessages.EMAIL_NOT_VALID;
    };

    return {
      handleDelegationChange,
      handleEmailChange,
      isValidEmail,
    };
  },
});
</script>

<style scoped lang="scss">
.delegation-option {
  width: 100%;
}

.delegation-checkbox {
  font-size: 14px;
  color: $black-87;
  transition: color 0.2s ease;
}

.delegation-description {
  font-size: 14px;
  line-height: 1.4;
  color: $black-54;
}

// Delegation styles are now in the delegation-option component

// Transition for delegation description
.delegation-description-transition-enter-active {
  transition: all 0.35s ease-out;
}

.delegation-description-transition-leave-active {
  transition: all 0.3s ease-in;
}

.delegation-description-transition-enter-from {
  opacity: 0;
  transform: translateY(-10px);
  max-height: 0;
  overflow: hidden;
}

.delegation-description-transition-leave-to {
  opacity: 0;
  transform: translateY(-5px);
  max-height: 0;
  overflow: hidden;
}

.delegation-description-transition-enter-to,
.delegation-description-transition-leave-from {
  opacity: 1;
  transform: translateY(0);
  max-height: 100px;
  overflow: visible;
}

.delegation-email-input-container {
  max-width: 445px;
}
</style>
