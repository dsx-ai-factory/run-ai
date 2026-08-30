<template>
  <q-menu anchor="bottom middle" self="top middle" aid="tools-select-menu">
    <q-list>
      <q-item
        v-for="(option, idx) in options"
        :key="idx"
        clickable
        v-close-popup
        @click="$emit('value-changed', option.value)"
        class="row items-center"
        :aid="getAid(option.label)"
        :disable="option.disable"
      >
        <span class="icon-width">
          <runai-svg-icon v-if="option.icon" :name="option.icon" size="25px"></runai-svg-icon>
        </span>
        <q-item-label class="q-ml-md">{{ option.label }}</q-item-label>
        <q-tooltip v-if="option.disable">Only the workload creator can connect to this tool</q-tooltip>
      </q-item>
    </q-list>
  </q-menu>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiSvgIcon } from "@/components/common/runai-svg-icon";
import type { ISelectOption } from "@/models/global.model";
import type { IToolItem } from "@/models/workload.model";
import { toolTypeIconsMap } from "@/common/icons.constant";

import { spaceToDash, toLowerCase } from "@/utils/string.util";
import { useAuthStore } from "@/stores/auth.store";

export default defineComponent({
  components: {
    RunaiSvgIcon,
  },
  emits: ["value-changed"],
  props: {
    toolItems: {
      type: Array as PropType<IToolItem[]>,
      required: true,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
    };
  },
  computed: {
    options(): Array<ISelectOption> {
      const options: Array<ISelectOption> = [];
      (this.toolItems as Array<IToolItem>)?.forEach((tool) => {
        const disable = Array.isArray(tool.authorizedUsers) && !tool.authorizedUsers.includes(this.authStore.userEmail);
        options.push({ label: tool.toolName, value: tool.url, icon: toolTypeIconsMap[tool.toolType], disable });
      });
      return options;
    },
  },
  methods: {
    getAid(name: string): string {
      return `${toLowerCase(spaceToDash(name))}-option`;
    },
  },
});
</script>

<style lang="scss" scoped>
.icon-width {
  min-width: 25px;
}
</style>
