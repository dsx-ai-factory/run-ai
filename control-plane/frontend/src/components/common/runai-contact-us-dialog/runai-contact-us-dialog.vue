<template>
  <section class="runai-rename-modal">
    <runai-base-dialog :model-value="true" @close="$emit('close')" size="width-sm" :no-padding="true">
      <template #header>Contact Us</template>

      <template #body>
        <div class="rename-workspace-body q-px-lg q-py-lg">
          <p>Hey Run:ai,</p>
          <q-form id="contact-us-form" ref="elContactUsForm">
            <!-- eslint-disable vuejs-accessibility/no-autofocus -- this component is the dialog's content and only mounts when the dialog opens in response to a user action; focusing the input follows the WAI-ARIA dialog pattern of moving focus into a newly opened modal -->
            <q-input
              v-model="contactUsText"
              autogrow
              placeholder="Describe how we can help"
              autofocus
              :rules="[required]"
              no-error-icon
              error-message="Let us know how can we help you"
              class="contact-us"
            />
            <!-- eslint-enable vuejs-accessibility/no-autofocus -->
          </q-form>
        </div>
      </template>

      <template #footer>
        <runai-btn label="Cancel" color="primary" aid="rename-modal-cancel-button" flat @click="$emit('close')" />
        <runai-btn label="Send" color="primary" aid="rename-modal-cancel-button" @click="send" />
      </template>
    </runai-base-dialog>
  </section>
</template>

<script lang="ts">
// common
import { defineComponent } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";

// constants
import { errorMessages } from "@/common/error-message.constant";
import { supportService } from "@/services/control-plane/support.service/support.service";
import { alertUtil } from "@/utils/alert.util";
import { RunaiBtn } from "@/components/common/runai-btn";

export default defineComponent({
  components: {
    RunaiBtn,
    RunaiBaseDialog,
  },
  emits: ["close", "send"],
  data() {
    return {
      contactUsText: "" as string,
    };
  },
  methods: {
    async send() {
      const isValid: boolean = await (this.$refs.elContactUsForm as HTMLFormElement).validate();
      if (!isValid) return;
      try {
        await supportService.contactUs(this.contactUsText);
        this.$q.notify(alertUtil.getSuccess("Got it! We'll get back to you soon via email."));
      } catch (e) {
        this.$q.notify(alertUtil.getError("Message couldn't be sent. Try again in a few minutes."));
      }
      this.$emit("close");
    },
    required(val: string): boolean | string {
      return !!val || errorMessages.VALUE_NOT_EMPTY;
    },
  },
});
</script>
