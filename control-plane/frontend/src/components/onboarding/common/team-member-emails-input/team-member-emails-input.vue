<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-select
    label="Team member(s) email"
    :placeholder="placeholder"
    v-model="memberEmails"
    use-input
    use-chips
    multiple
    hide-dropdown-icon
    no-error-icon
    input-debounce="0"
    new-value-mode="add"
    :rules="[isValidEmailList]"
    counter
    :max-values="MAX_EMAILS"
    class="member-emails-input"
    @input-value="handleInputValueChange"
    @blur="handleInputBlur"
  />
</template>

<script lang="ts">
import { computed, defineComponent, type PropType, ref, watch } from "vue";
import type { TeamData, TeamDataKey } from "@/components/onboarding/models/admin-onboarding.model";
import { errorMessages } from "@/common/error-message.constant";
import { formValidators } from "@/common/form.validators";

const MAX_EMAILS = 10;
export default defineComponent({
  name: "team-member-emails-input",
  emits: {
    "update:team-data": (key: TeamDataKey, value: string | boolean | string[]) => !!key && value !== undefined,
  },
  props: {
    teamData: {
      type: Object as PropType<TeamData>,
      required: true,
    },
  },
  setup(props, { emit }) {
    const memberEmails = ref<string[]>(props.teamData.memberEmails || []);
    const currentInputValue = ref<string>("");
    const placeholder = computed(() => (memberEmails.value.length < MAX_EMAILS ? "Enter email(s)" : undefined));

    // Event handlers
    const handleMemberEmailsChange = (value: string[]) => {
      emit("update:team-data", "memberEmails", value);
    };

    const handleInputValueChange = (value: string) => {
      currentInputValue.value = value;
    };

    const handleInputBlur = () => {
      const trimmedValue = currentInputValue.value.trim();

      // Only add if there's a value, it's a valid email, not already in the list, and we haven't reached the limit
      if (
        trimmedValue &&
        formValidators.isValidEmail(trimmedValue) &&
        !memberEmails.value.includes(trimmedValue) &&
        memberEmails.value.length < MAX_EMAILS
      ) {
        memberEmails.value = [...memberEmails.value, trimmedValue];
        currentInputValue.value = ""; // Clear the input value
      }
    };

    // Validation functions
    const isValidEmailList = (value: string[]) => {
      if (!value || value.length === 0) {
        return errorMessages.EMAIL_NOT_VALID;
      }
      for (const email of value) {
        if (!formValidators.isValidEmail(email)) {
          return errorMessages.EMAIL_NOT_VALID;
        }
      }
      return true;
    };

    watch(memberEmails, (newValue: string[]) => {
      handleMemberEmailsChange(newValue);
    });

    // Watch for external changes to team data - in case of resetting or updating
    watch(
      () => props.teamData.memberEmails,
      (newEmails) => {
        if (newEmails && JSON.stringify(newEmails) !== JSON.stringify(memberEmails.value)) {
          memberEmails.value = [...newEmails];
        }
      },
      { immediate: true },
    );

    return {
      memberEmails,
      placeholder,
      handleMemberEmailsChange,
      handleInputValueChange,
      handleInputBlur,
      isValidEmailList,
      MAX_EMAILS,
    };
  },
});
</script>

<style scoped lang="scss"></style>
