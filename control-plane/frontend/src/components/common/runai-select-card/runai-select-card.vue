<template>
  <label class="runai-select-card" :for="cmpId">
    <q-card class="card-content" :class="{ 'is-checked': value, 'is-disabled': disabled, 'no-padding': noPadding }">
      <slot />
      <input type="checkbox" :id="cmpId" :value="value" @click="!disabled && $emit('clicked')" />
      <q-icon class="checkmark" color="white" name="ph-bold ph-check" />
    </q-card>
  </label>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { uid } from "quasar";

export default defineComponent({
  name: "runai-select-card",
  events: ["clicked"],
  props: {
    value: {
      type: Boolean as PropType<boolean>,
      default: undefined,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    noPadding: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return {
      cmpId: uid() as string,
    };
  },
});
</script>

<style lang="scss" scoped>
.runai-select-card {
  $checkmark-size: 21px;
  $checkmark-position: 3px;
  .card-content {
    transform: scale(0.95);
    &:hover:not(.is-disabled) {
      transform: scale(1);
    }
    display: inline-block;
    position: relative;
    cursor: pointer;
    @include runaiSelectCardSize;
    transition: all 0.3s;

    &:hover:not(&.is-disabled),
    &.is-checked {
      border: 3px solid $primary;
    }

    &.is-disabled {
      opacity: 0.5;
      cursor: default;
    }

    &.is-checked {
      .checkmark {
        opacity: 1;
      }
    }

    input {
      position: absolute;
      opacity: 0;
      cursor: pointer;
      height: 0;
      width: 0;
    }

    .checkmark {
      position: absolute;
      top: calc(#{$runai-select-card-checkmark-position} * -1);
      right: calc(#{$runai-select-card-checkmark-position} * -1);
      height: $runai-select-card-checkmark-size;
      width: $runai-select-card-checkmark-size;
      background-color: $primary;
      opacity: 0;
      transition: all 0.3s;
      border-radius: 5px;
    }
  }
}
</style>
