<template>
  <q-btn flat color="primary" class="btn-icon-sm" icon="ph-bold ph-plus" :label="buttonLabel" :disable="disable">
    <q-menu anchor="bottom left">
      <q-list>
        <q-item
          v-for="option in options"
          clickable
          v-close-popup
          :disable="option.disable"
          :key="option.label"
          :aid="option.value"
          @click="$emit('item-clicked', option.value)"
        >
          <q-item-section>
            <span>{{ option.label }}</span>
          </q-item-section>
        </q-item>
      </q-list>
    </q-menu>
  </q-btn>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
//models
import type { ISelectOption } from "@/models/global.model";

export default defineComponent({
  emits: ["item-clicked"],
  props: {
    buttonLabel: {
      type: String as PropType<string>,
      required: true,
    },
    options: {
      type: Array as PropType<ISelectOption[]>,
      required: true,
    },
    disable: {
      type: Boolean,
      default: false,
    },
  },
});
</script>
