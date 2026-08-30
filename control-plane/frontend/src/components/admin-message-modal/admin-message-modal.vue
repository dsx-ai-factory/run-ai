<template>
  <runai-base-dialog :model-value="true" class="admin-message-modal" size="width-sm" @close="$emit('close', dismiss)">
    <template #header>
      Message from Administrator
      <span class="text-subtitle1">({{ messageCreationTime }})</span>
    </template>

    <template #body>
      <div class="modal-body" v-dompurify-html="message.msg"></div>
    </template>

    <template #footer>
      <q-btn label="got it" color="primary" @click="$emit('close', dismiss)" />
    </template>

    <template #footer-left>
      <q-checkbox v-if="message.enableDismiss" v-model="dismiss" label="Don't show again" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiBaseDialog } from "../common/runai-base-dialog";
import type { IAdminMessage } from "@/models/setting.model";
import { dateUtil } from "@/utils/date.util";

export default defineComponent({
  name: "admin-message-modal",
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
      return dateUtil.dateAndTimeFormat(new Date(this.message.createdAt));
    },
  },
});
</script>

<style lang="scss">
.admin-message-modal {
  .modal-body {
    min-height: 200px;
    word-break: break-word;
  }
}
</style>
