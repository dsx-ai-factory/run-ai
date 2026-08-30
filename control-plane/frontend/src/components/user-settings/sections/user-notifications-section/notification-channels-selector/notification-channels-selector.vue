<template>
  <div class="q-mb-lg">
    <div class="q-py-sm">Send me system notifications via:</div>
    <div class="row items-center q-py-sm q-gutter-md">
      <div v-for="option in options" :key="option.label" class="row items-center">
        <q-checkbox
          :model-value="option.isSelected"
          size="md"
          dense
          :label="option.label"
          :disable="option.isDisabled"
          @update:model-value="onOptionToggle(option.id)"
        />
        <q-tooltip
          v-if="option.isDisabled && option.disabledTooltipText"
          anchor="bottom middle"
          self="top middle"
          :offset="[0, 10]"
          max-width="355px"
        >
          {{ option.disabledTooltipText }}
        </q-tooltip>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { NotificationChannelOption } from "./notification-channels-selector.types";

export default defineComponent({
  name: "notification-channels-selector",
  emits: ["channel-toggled"],
  props: {
    options: {
      type: Array as PropType<NotificationChannelOption[]>,
      required: true,
    },
  },
  methods: {
    onOptionToggle(id: string): void {
      this.$emit("channel-toggled", id);
    },
  },
});
</script>
