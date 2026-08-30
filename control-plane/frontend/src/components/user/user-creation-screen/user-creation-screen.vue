<template>
  <p>Enter the user's email address</p>
  <runai-email-validation
    :model-value="modelValue"
    :rules="[isEmpty, isValidEmail, isEmailNotTaken]"
    auto-focus
    @update:model-value="$emit('update:model-value', $event)"
  ></runai-email-validation>
  <div class="q-mt-lg text-b-subtitle2 name-section-label">
    Enter the user's first and last name
    <span class="optional-label">(optional)</span>
  </div>
  <div class="row q-col-gutter-md q-mt-xs">
    <div class="col-6">
      <q-input
        aid="user-creation-first-name"
        :model-value="firstName"
        label="First name"
        stack-label
        no-error-icon
        @update:model-value="$emit('update:first-name', String($event ?? '').trim())"
      />
    </div>
    <div class="col-6">
      <q-input
        aid="user-creation-last-name"
        :model-value="lastName"
        label="Last name"
        stack-label
        no-error-icon
        @update:model-value="$emit('update:last-name', String($event ?? '').trim())"
      />
    </div>
  </div>
  <q-checkbox
    :model-value="sendInvitation"
    @update:model-value="$emit('update:send-invitation', $event)"
    :disable="!isSmtpConfigured"
    label="Send an email invitation to the new user to sign in and get started."
    class="q-mt-md"
  >
    <q-tooltip v-if="!isSmtpConfigured" max-width="300px" anchor="top right" self="center middle">
      To enable email invitations, first go to General settings and configure email notifications.
    </q-tooltip>
  </q-checkbox>
</template>
<script lang="ts">
import { errorMessages } from "@/common/error-message.constant";
import { isNotEmpty, isValidEmail } from "@/common/form.validators";
import { RunaiEmailValidation } from "@/components/common/runai-email-validation";

import type { PropType } from "vue";
import userService from "@/services/control-plane/user.service/user.service";
export default {
  components: { RunaiEmailValidation },
  emits: ["update:model-value", "update:send-invitation", "update:first-name", "update:last-name"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    sendInvitation: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isSmtpConfigured: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    firstName: {
      type: String as PropType<string>,
      default: "",
    },
    lastName: {
      type: String as PropType<string>,
      default: "",
    },
  },
  methods: {
    isEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.EMAIL_NOT_EMPTY;
    },
    isValidEmail(val: string): boolean | string {
      return isValidEmail(val) || errorMessages.EMAIL_NOT_VALID;
    },
    async isEmailNotTaken(userEmail: string): Promise<boolean | string> {
      const isUserExists = await userService.isUserExist(userEmail);
      return !isUserExists || errorMessages.EMAIL_ALREADY_EXIST;
    },
  },
};
</script>
<style lang="scss" scoped>
.optional-label {
  color: var(--b-text-sub);
  font-style: italic;
  margin-left: 4px;
}
</style>
