<template>
  <q-input
    class="runai-name-validation"
    aid="runai-name-validation"
    :model-value="modelValue"
    :placeholder="placeholder"
    debounce="300"
    :counter="counter"
    :maxlength="maxlength"
    @update:model-value="$emit('update:modelValue', $event)"
    :rules="[...rules]"
    no-error-icon
    ref="nameValidate"
    :reactive-rules="reactiveRules"
    :autofocus="autoFocus"
    :input-class="inputClass"
    :label="label"
    :stack-label="stackLabel"
    :readonly="readOnly"
  ></q-input>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// model
import type { ValidationRule } from "quasar";

export default defineComponent({
  emits: ["update:modelValue", "invalid"],
  props: {
    modelValue: {
      type: String as PropType<string>,
      required: true,
    },
    placeholder: {
      type: String as PropType<string>,
    },
    rules: {
      type: Array as PropType<ValidationRule[]>,
      default: () => [],
    },
    autoFocus: {
      type: Boolean,
      default: false,
    },
    inputClass: {
      type: String as PropType<string>,
      required: false,
    },
    label: {
      type: String as PropType<string>,
      required: false,
      default: "Name",
    },
    stackLabel: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    reactiveRules: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    counter: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    maxlength: {
      type: Number as PropType<number>,
      default: 40,
    },
  },
  data() {
    return {
      nameValidateEl: null as HTMLFormElement | null,
    };
  },
  mounted() {
    this.nameValidateEl = this.$refs["nameValidate"] as HTMLFormElement;

    if (this.$route.query.fromCopyId && this.autoFocus) {
      this.$nextTick(async () => {
        await this.nameValidateEl?.validate();
        this.nameValidateEl?.focus();
      });
    }
  },
  computed: {
    hasError(): undefined | boolean {
      return this.nameValidateEl?.hasError;
    },
  },
  watch: {
    hasError(val: boolean): void {
      this.$emit("invalid", val);
    },
  },
});
</script>
