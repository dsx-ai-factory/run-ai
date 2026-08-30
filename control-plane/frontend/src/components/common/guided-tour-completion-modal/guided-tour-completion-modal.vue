<template>
  <runai-base-dialog
    :model-value="open"
    @close="$emit('close')"
    size="width-sm"
    :hide-x="true"
    class="guided-tour-modal"
  >
    <template #header>Guide completed!</template>

    <template #body>
      You're ready to go!
      <br />
      To revisit the guide later on, go to Help - > Guided steps
    </template>

    <template #footer>
      <q-btn
        label="Got it"
        color="primary"
        @click="$emit('close')"
        aid="guided-tour-completion-leave-btn"
        class="q-px-lg"
      />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

export default defineComponent({
  name: "GuidedTourCompletionModal",
  components: {
    RunaiBaseDialog,
  },
  emits: ["close"],
  props: {
    open: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>
