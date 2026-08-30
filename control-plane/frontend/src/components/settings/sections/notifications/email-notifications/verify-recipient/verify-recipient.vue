<template>
  <runai-section
    class="verify-recipient"
    :class="{ 'read-only': readOnly }"
    title="Set a recipient email address to verify connectivity"
    tooltip-text="Enter a recipient email to ensure end-to-end verification. This is only used for verification; no message will be sent."
    tooltip-width="400px"
  >
    <div class="row col-12">
      <setting-readonly-input class="col-5" :read-only="readOnly" :display-tooltip="displayReadOnlyTooltip">
        <template v-slot:default="{ readonly }">
          <q-input
            aid="verify-recipient-email-input"
            class="col-5"
            :model-value="modelValue"
            placeholder="e.g., your@company.com (Optional)"
            input-class="placeholder-italic"
            :rules="[isValidOptionalEmail]"
            no-error-icon
            @update:model-value="updateRecipient"
            stack-label
            label="Email address"
            ref="recipientInput"
            :disable="readonly"
          />
        </template>
      </setting-readonly-input>
    </div>
  </runai-section>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { QInput } from "quasar";
import { isEmpty, isValidEmail } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { RunaiSection } from "@/components/common/runai-section";
import SettingReadonlyInput from "@/components/settings/setting-readonly-input/setting-readonly-input.vue";

export interface Props {
  modelValue?: string;
  readOnly?: boolean;
  displayReadOnlyTooltip?: boolean;
}

withDefaults(defineProps<Props>(), {
  modelValue: "",
  readOnly: undefined,
  displayReadOnlyTooltip: true,
});

const emit = defineEmits<{
  "update:model-value": [recipient: string];
}>();

const recipientInput = ref<QInput | null>(null);

function isValidOptionalEmail(value: string): boolean | string {
  return isEmpty(value) || isValidEmail(value) || errorMessages.EMAIL_NOT_VALID;
}

function updateRecipient(recipient: string | number | null): void {
  emit("update:model-value", String(recipient ?? ""));
}

async function validate(): Promise<boolean> {
  return (await recipientInput.value?.validate()) ?? false;
}

defineExpose({ validate });
</script>

<style scoped lang="scss">
.verify-recipient.read-only :deep(.runai-section-title) {
  color: $grey-6;
}
</style>
