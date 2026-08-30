<template>
  <q-item
    class="workload-drawer-item q-pa-md"
    :clickable="!item.disabled"
    :focused="focused"
    :class="{ selected, focused: focused && !selected }"
    @click.stop="$emit('item-clicked', { id: item.id, name: item.name, kind: item.kind })"
  >
    <section class="icons-section row items-center q-mr-md" :class="disabledClass">
      <item-icon-with-counter :icons="icons" />
    </section>
    <q-item-section class="text-section row justify-center" :class="disabledClass">
      <q-item-label class="item-label" lines="1">{{ item.name }}</q-item-label>
      <q-item-label class="item-description" lines="2">{{ item.description || "" }}</q-item-label>
    </q-item-section>
    <section class="actions-section row">
      <q-btn
        aid="workload-drawer-item-policy-button"
        v-if="showPolicyInfoButton"
        @click.stop="onPolicyClicked"
        class="btn-icon-md"
        icon="ph-bold ph-clipboard-text"
        flat
        round
        size="12px"
        color="icon-default"
      >
        <q-tooltip>Policy is applied</q-tooltip>
      </q-btn>
      <q-btn
        v-if="item.details.length"
        aid="workload-drawer-item-details-button"
        @click="onDetailsClicked"
        icon="ph-bold ph-info"
        class="btn-icon-md"
        flat
        round
        size="12px"
        color="icon-default"
      >
        <q-tooltip anchor="bottom left" :offset="[-10, 14]">More info</q-tooltip>
      </q-btn>
    </section>
    <q-tooltip v-if="item.disabled && item.disabledInfo" anchor="bottom middle" :offset="[0, -10]">{{
      item.disabledInfo
    }}</q-tooltip>
  </q-item>
</template>

<script lang="ts">
import type { IDrawerListItem, IDrawerListItemIcon } from "@/models/workload-creation-drawer-model";
import { ItemIconWithCounter } from "../item-icon-with-counter";
import { defineComponent } from "vue";
import type { PropType } from "vue";

export default defineComponent({
  name: "workload-drawer-item",
  components: {
    ItemIconWithCounter,
  },
  emits: ["item-clicked", "policy-clicked", "details-clicked"],
  props: {
    item: {
      type: Object as PropType<IDrawerListItem>,
      required: true,
    },
    focused: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    selected: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    icons(): IDrawerListItemIcon[] {
      return this.item.icons || [];
    },
    showPolicyInfoButton(): boolean {
      return !!this.item.policy?.reason?.[0]?.field;
    },
    disabledClass(): string {
      if (this.item.disabled) return "disabled";
      return "";
    },
  },
  methods: {
    onPolicyClicked(ev: Event): void {
      ev.stopPropagation();
      this.$emit("policy-clicked", this.item.id);
    },
    onDetailsClicked(ev: Event): void {
      ev.stopPropagation();
      this.$emit("details-clicked", this.item.id);
    },
  },
});
</script>
<style lang="scss">
.q-item.workload-drawer-item {
  max-width: 372px;
  min-height: 79px;
  &.selected {
    background-color: rgba(6, 84, 254, 0.1);
  }
  &.focused {
    background-color: rgba(238, 238, 238, 0.9);
  }
  .text-section {
    .item-label {
      font-size: 16px;
      font-weight: 700;
    }
    .item-description {
      font-size: 12px;
      font-weight: 400;
    }
  }
  .actions-section {
    align-items: center;
  }
  .icons-section {
    position: relative;
    .additionals-counter {
      background-color: $body-background-color;
      position: absolute;
      top: 0;
      left: 0;
      transform: translate(50%, 50%);
    }
  }
}
</style>
