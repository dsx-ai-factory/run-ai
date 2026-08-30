<template>
  <section class="access-input-list">
    <div
      v-for="item in itemsData"
      class="access-input-container q-mb-md q-pa-md row justify-between items-center"
      :key="item.id"
    >
      <policy-string-field
        v-model="item.text"
        @update:model-value="onItemChanged"
        :label="inputLabel"
        stack-label
        :rules="validationRules"
        class="col-8"
        :policy-rules="policyRules"
        ref="inputRefs"
      />

      <span>
        <q-btn
          v-if="item.removable"
          class="runai-icon-btn--default text-icon-default"
          icon="ph-bold ph-x"
          flat
          round
          aid="remove-item-button"
          @click="removeItem(item.id)"
          :disable="policyRules?.canEdit === false"
        >
          <q-tooltip v-if="policyRules?.canEdit === false">{{ cantBeModifiedPolicy }}</q-tooltip>
          <q-tooltip v-else>Remove</q-tooltip>
        </q-btn>
      </span>
    </div>

    <span class="add-item-button-container">
      <q-btn
        aid="add-item-button"
        class="add-item-button btn-icon-sm"
        flat
        color="primary"
        icon="ph-bold ph-plus"
        :label="addItemBtnLabel"
        @click="addNewItem"
        :disable="policyRules?.canEdit === false"
      ></q-btn>
      <q-tooltip v-if="policyRules?.canEdit === false">{{ cantBeModifiedPolicy }}</q-tooltip>
    </span>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { ValidationRule } from "quasar";

import { makeId } from "@/utils/common.util";
import { PolicyStringField } from "@/components/common/policy-string-field";

import type { IAccessInput } from "./access-input-list.model";
import type { ArrayRules } from "@/swagger-models/policy-service-client";
import { errorMessages } from "@/common/error-message.constant";

function getEmptyItem(label: string): IAccessInput {
  return {
    id: makeId(),
    text: "",
    label,
    removable: true,
  };
}

export default defineComponent({
  name: "access-input-list",
  components: {
    PolicyStringField,
  },
  emits: ["changed"],
  props: {
    items: {
      type: Array as PropType<Array<IAccessInput>>,
      default: () => [],
    },
    inputLabel: {
      type: String as PropType<string>,
      required: true,
    },
    addItemBtnLabel: {
      type: String as PropType<string>,
      required: true,
    },
    validationRules: {
      type: Array as PropType<ValidationRule[]>,
      default: () => [],
    },
    policyRules: {
      type: [Object, null] as PropType<ArrayRules | null>,
      required: false,
    },
  },
  data() {
    return {
      itemsData: [] as IAccessInput[],
      cantBeModifiedPolicy: errorMessages.CANT_BE_MODIFIED_POLICY,
    };
  },
  mounted() {
    this.initList();
  },
  methods: {
    initList(): void {
      this.itemsData = this.items.map((item: IAccessInput) => ({
        ...item,
        id: item.id || makeId(),
      }));

      if (this.itemsData.length === 0) {
        this.itemsData.push(getEmptyItem(this.inputLabel));
        this.focusInput();
      }

      if (this.itemsData.length === 1) {
        this.itemsData[0].removable = false;
      }
    },
    focusInput(): void {
      this.$nextTick(() => {
        const lastIndex = this.itemsData.length - 1;
        const inputsRefs = this.$refs.inputRefs as HTMLInputElement[];
        inputsRefs[lastIndex]?.focus();
      });
    },
    async addNewItem(): Promise<void> {
      this.itemsData.push(getEmptyItem(this.inputLabel));
      if (this.itemsData.length > 1) {
        this.itemsData = this.itemsData.map((item: IAccessInput) => ({
          ...item,
          removable: true,
        }));
      }
      await this.focusInput();
      this.onItemChanged();
    },
    removeItem(itemId?: string): void {
      this.itemsData = this.itemsData.filter((item: IAccessInput) => item.id !== itemId);
      if (this.itemsData.length === 1) {
        this.itemsData[0].removable = false;
      }

      this.onItemChanged();
    },
    onItemChanged(): void {
      this.$emit(
        "changed",
        this.itemsData.map((item: IAccessInput) => ({
          text: item.text,
          label: item.label,
        })),
      );
    },
  },
  watch: {
    inputLabel(): void {
      this.initList();
    },
  },
});
</script>
<style lang="scss" scoped>
.access-input-list {
  width: 100%;

  .access-input-container {
    background-color: $body-background-color;
  }
}
</style>
