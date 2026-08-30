<template>
  <section class="command-and-args">
    <div class="q-py-md" v-if="!hideTitle">Set a command and arguments for the container running in the pod</div>
    <section class="env-var-section grid-container command-grid" v-if="displayForm">
      <section :class="commandAndArgInputSectionClass">
        <policy-string-field
          aid="command-input"
          @update:model-value="update('command', $event)"
          :model-value="commandAndArgs.command"
          label="Command"
          placeholder="e.g. start-notebook.sh"
          :rules="commandInputRules"
          stack-label
          input-class="placeholder-italic"
          :disable="disable"
          :policy-rules="commandPolicyRules"
          :dense="isViewMode"
        />
      </section>
      <section :class="commandAndArgInputSectionClass">
        <policy-string-field
          aid="args-input"
          @update:model-value="update('args', $event)"
          :model-value="commandAndArgs.args"
          label="Arguments"
          placeholder="e.g., --logtostderr --arg1=val --v=4"
          :hint="isViewMode ? '' : 'Add multiple arguments separated by spaces'"
          :rules="[isValidArgs]"
          stack-label
          input-class="placeholder-italic"
          :disable="disable"
          :policy-rules="argsPolicyRules"
          :dense="isViewMode"
        />
      </section>
      <section class="close-command-button">
        <q-btn
          v-if="showCloseButton"
          class="close-button"
          aid="close-command-and-args-button"
          icon="ph-bold ph-x"
          flat
          round
          @click="showCommandAndArgs(false)"
          :disable="disableClose"
        >
          <q-tooltip v-if="disableClose">{{ disableCloseText }}</q-tooltip>
        </q-btn>
      </section>
    </section>
    <q-btn
      v-else
      flat
      color="primary"
      aid="add-command-and-args-button"
      class="btn-icon-sm"
      icon="ph-bold ph-plus"
      label="COMMAND & ARGUMENTS"
      :disable="disable"
      @click="showCommandAndArgs(true)"
    ></q-btn>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { isEmpty, validCommandAndArgsLine } from "@/common/form.validators";
import type { IUICommandAndArgs } from "@/models/environment.model";
import { PolicyStringField } from "@/components/common/policy-string-field";
import type { StringRules } from "@/swagger-models/policy-service-client";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  components: {
    PolicyStringField,
  },
  emits: ["update"],
  props: {
    commandAndArgs: {
      type: Object as PropType<IUICommandAndArgs>,
      required: false,
      default: () => ({
        command: "",
        args: "",
      }),
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hideTitle: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    isViewMode: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    commandPolicyRules: {
      type: [Object, null] as PropType<StringRules | null>,
      required: false,
    },
    argsPolicyRules: {
      type: [Object, null] as PropType<StringRules | null>,
      required: false,
    },
  },
  data() {
    return {
      showInnerSections: false as boolean,
    };
  },
  computed: {
    displayForm(): boolean {
      if (this.isViewMode) return true;
      if (this.commandPolicyRules?.required || this.argsPolicyRules?.required) return true;
      return !(isEmpty(this.commandAndArgs?.command) && isEmpty(this.commandAndArgs?.args)) || this.showInnerSections;
    },
    commandInputRules(): undefined | ((val: string) => string | boolean)[] {
      return this.isViewMode ? undefined : [this.isValidCommand];
    },
    commandAndArgInputSectionClass(): string {
      return this.isViewMode ? "q-px-md" : "q-pa-md";
    },
    requiredByPolicy(): boolean {
      return !!this.commandPolicyRules?.required || !!this.argsPolicyRules?.required;
    },
    disableClose(): boolean {
      return (
        this.requiredByPolicy || this.commandPolicyRules?.canEdit === false || this.argsPolicyRules?.canEdit === false
      );
    },
    showCloseButton(): boolean {
      if (this.requiredByPolicy) return false;
      return !this.disable;
    },
    disableCloseText(): string {
      return errorMessages.CANT_BE_MODIFIED_POLICY;
    },
  },
  methods: {
    showCommandAndArgs(toggle: boolean): void {
      this.showInnerSections = toggle;
      if (!toggle) {
        this.clearValues();
      }
    },
    update(field: string, value: string | null | number): void {
      this.$emit("update", { ...this.commandAndArgs, [field]: value === "" ? null : value });
    },
    clearValues(): void {
      this.showInnerSections = false;
      this.$emit("update", { args: "", command: "" });
    },
    isValidCommand(val: string): string | boolean {
      if (!val) return true;
      return validCommandAndArgsLine(val) || errorMessages.ENTER_A_VALID_COMMAND;
    },
    isValidArgs(val: string): string | boolean {
      if (!val) return true;
      return validCommandAndArgsLine(val) || errorMessages.ENTER_VALID_ARGUMENTS;
    },
  },
});
</script>
<style lang="scss" scoped>
.command-and-args {
  .env-var-section {
    background-color: $body-background-color;
  }

  .close-command-button {
    color: $black-54;
    display: flex;
    justify-content: flex-end;
    align-items: flex-start;
    margin: 10px;
  }

  .close-button {
    font-size: 12px;
  }

  .close-argument-button {
    color: $black-54;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 10px;
    font-size: 12px;
  }

  .grid-container {
    display: grid;
    grid-template-columns: 92% 8%;
  }

  .command-grid {
    grid-template-rows: 50% 50%;
    grid-auto-flow: column;
  }

  .args-grid {
    grid-auto-flow: row;
  }
}
</style>
