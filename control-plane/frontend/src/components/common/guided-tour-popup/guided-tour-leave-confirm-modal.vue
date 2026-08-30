<template>
  <runai-base-dialog
    :model-value="modelValue"
    @close="handleGoBack"
    size="width-sm"
    :hide-x="true"
    class="guided-tour-modal"
    aid="guided-tour-leave-confirm-modal"
  >
    <template #header>Leave the guide?</template>

    <template #body>You can always come back later via Help → Guided steps</template>

    <template #footer>
      <q-btn
        label="Go back"
        color="primary"
        flat
        @click="handleGoBack"
        aid="guided-tour-leave-cancel-btn"
        class="q-mr-sm"
      />
      <q-btn label="Leave" color="primary" @click="handleLeave" aid="guided-tour-leave-confirm-btn" class="q-px-lg" />
    </template>
  </runai-base-dialog>
</template>

<script setup lang="ts">
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  (e: "update:model-value", value: boolean): void;
  (e: "go-back"): void;
  (e: "leave"): void;
}>();

const handleGoBack = (): void => {
  emit("update:model-value", false);
  emit("go-back");
};

const handleLeave = (): void => {
  emit("update:model-value", false);
  emit("leave");
};
</script>
