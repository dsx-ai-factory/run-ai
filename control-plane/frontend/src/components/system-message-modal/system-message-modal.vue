<template>
  <runai-base-dialog :model-value="true" class="system-message-modal" size="width-sm" @close="$emit('close', dismiss)">
    <template #header>
      System notification
      <span v-if="messageCreationTime" class="text-subtitle1">({{ messageCreationTime }})</span>
    </template>

    <template #body>
      <div class="modal-body" v-dompurify-html="message.msg"></div>
    </template>

    <template #footer>
      <q-btn label="got it" color="primary" @click="$emit('close', dismiss)" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiBaseDialog } from "../common/runai-base-dialog";
import type { IAdminMessage } from "@/models/setting.model";
import { dateUtil } from "@/utils/date.util";

export default defineComponent({
  name: "system-message-modal",
  components: {
    RunaiBaseDialog,
  },
  emits: ["close"],
  props: {
    message: {
      type: Object as PropType<IAdminMessage>,
      required: true,
    },
  },
  data() {
    return {
      dismiss: false,
    };
  },
  computed: {
    messageCreationTime(): string {
      if (!this.message.createdAt) return "";
      return dateUtil.dateAndTimeFormat(new Date(this.message.createdAt));
    },
  },
});
</script>

<style lang="scss">
.system-message-modal {
  .modal-body {
    min-height: 200px;
    word-break: break-word;
  }
}
</style>
