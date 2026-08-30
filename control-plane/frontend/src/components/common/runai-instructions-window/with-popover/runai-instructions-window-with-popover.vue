<template>
  <runai-small-popover
    @hide="onHide"
    :no-padding="true"
    custom-width="351px"
    ref="popover"
    max-height="50vh"
    :show-separator="false"
  >
    <template #header>
      <strong class="text-b-main">{{ editInstructionText ? instructionsTitleEdit : instructionsTitleView }}</strong>
    </template>

    <template #body>
      <div class="q-pa-md">
        <div v-if="editInstructionText">
          <div class="row q-mb-lg">Add instructions on the environment variable for the workload creator to follow.</div>
          <div class="small-popover-body">
            <q-input
              type="textarea"
              counter
              maxlength="250"
              aid="environment-variable-insturctions"
              no-error-icon
              v-model="instructionText"
              rows="4"
              outlined
            />
          </div>
        </div>

        <div v-else>
          <div class="row description-text">
            {{ environmentVariableDescription }}
          </div>
        </div>
      </div>
    </template>
    <template #footer>
      <section>
        <q-btn color="primary" flat :label="selectButtonCaption" @click="onClose" />
      </section>
    </template>
  </runai-small-popover>
</template>

<script lang="ts">
// stores
import { defineComponent, type PropType } from "vue";

// models

// cmps
import { RunaiSmallPopover } from "@/components/common/runai-small-popover";

// services
export type RunaiSmallPopoverType = InstanceType<typeof RunaiSmallPopover>;

export default defineComponent({
  components: { RunaiSmallPopover },
  emits: ["update-instruction-text", "hide"],
  props: {
    environmentVariableDescription: {
      type: String as PropType<string | null>,
    },
    editInstructionText: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    instructionsTitleEdit: {
      type: String as PropType<string>,
      default: "Add instructions",
    },
    instructionsTitleView: {
      type: String as PropType<string>,
      default: "Instructions",
    },
  },
  data() {
    return {
      instructionText: "" as string,
    };
  },
  created() {
    this.instructionText = this.environmentVariableDescription || "";
  },
  computed: {
    selectButtonCaption(): string {
      return this.editInstructionText ? "Save" : "Close";
    },
  },
  methods: {
    onClose(): void {
      if (this.editInstructionText) {
        this.$emit("update-instruction-text", this.instructionText);
      }
      this.hide();
    },
    hide(): void {
      const popover = this.$refs.popover as RunaiSmallPopoverType;
      popover.hide();
    },
    onHide(): void {
      this.instructionText = this.environmentVariableDescription || "";
      this.$emit("hide");
    },
  },
});
</script>

<style lang="scss">
.small-popover-body {
  max-height: calc(50vh - 184px);
}

.description-text {
  white-space: pre-line;
  word-wrap: break-word;
  overflow-wrap: anywhere;
  width: 100%;
}
</style>
