<template>
  <section class="extended-resources">
    <div class="q-py-md">
      <span> Set extended resource(s) </span>
    </div>
    <div class="q-pb-md more-info-container italic">
      For more information on how to set extended resources, see the
      <a :href="DOCS.KUBERNETES.TASKS.EXTENDED_RESOURCES" target="_blank">Extended resources</a>
      and
      <a :href="DOCS.KUBERNETES.REFERENCE.API.QUANTITY" target="_blank">Quantity</a>
      guides
    </div>
    <section
      v-for="extendedResource of displayedItems"
      :key="extendedResource.id"
      class="row justify-between extended-resource-box q-pa-md items-center"
      :aria-disabled="extendedResource.isLocked"
    >
      <policy-string-field
        class="col-5 input-name"
        @update:model-value="updateModel(extendedResource.id, 'resource', $event)"
        :model-value="extendedResource.origin.resource"
        label="Resource"
        placeholder="e.g., subdomain/resource-name"
        input-class="placeholder-italic"
        :rules="[isEmptyNameField, isNameValid, nameNotAdded]"
        no-error-icon
        stack-label
        :disable="extendedResource.isLocked || disable"
      />

      <policy-string-field
        class="col-5 quantity-input"
        @update:model-value="updateModel(extendedResource.id, 'quantity', $event)"
        :model-value="extendedResource.origin.quantity"
        label="Quantity"
        input-class="placeholder-italic"
        placeholder="e.g., 100"
        :rules="[isEmptyQuantityField, isValidQuantity]"
        no-error-icon
        stack-label
        :policy-rules="policyRules?.attributes?.quantity"
        :disable="extendedResource.isLocked || disable"
      />

      <q-btn
        class="close-button"
        icon="ph-bold ph-x"
        flat
        round
        @click="remove(extendedResource.id)"
        :disable="extendedResource.isLocked || disable"
      >
        <q-tooltip v-if="extendedResource.isLocked">{{ disabledByPolicyText }}</q-tooltip>
      </q-btn>
    </section>

    <q-btn
      class="add-button btn-icon-sm"
      flat
      color="primary"
      icon="ph-bold ph-plus"
      label="EXTENDED RESOURCE"
      @click="add"
      :disable="cannotAddDueToPolicy || disable"
    >
      <q-tooltip v-if="cannotAddDueToPolicy">{{ disabledByPolicyText }}</q-tooltip></q-btn
    >
    <q-tooltip anchor="center middle" v-if="tooltipText">{{ tooltipText }}</q-tooltip>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { deepCopy } from "@/utils/common.util";
import type { ExtendedResource } from "@/swagger-models/assets-service-client";
import { isExtendedResourceName, isValidExtendedResourceQuantity } from "@/utils/k8s-validation.util";
import { isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import type { ExtendedResourcesDefaults, ExtendedResourcesRules } from "@/swagger-models/policy-service-client";
import { PolicyStringField } from "@/components/common/policy-string-field";
import { useItemizedPolicy } from "@/composables/use-itemized-policy.composable";
import { ItemizedMapperEntities } from "@/composables/helpers/use-itemized-policy.mapper";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  components: {
    PolicyStringField,
  },
  emits: ["update:model-value"],
  props: {
    extendedResources: {
      type: Array as PropType<Array<ExtendedResource>>,
      required: false,
      default: () => new Array<ExtendedResource>(),
    },
    policyRules: {
      type: [Object, null] as PropType<ExtendedResourcesRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<ExtendedResourcesDefaults | null>,
      required: false,
    },
    disable: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
  },
  setup(props) {
    const { disabledByPolicyText, cannotAddDueToPolicy, addItem, removeItem, updateItem, displayedItems, fullList } =
      useItemizedPolicy(
        props.extendedResources,
        ItemizedMapperEntities.ExtendedResources,
        props.policyRules?.instances,
        props.policyDefaults?.instances,
      );

    return {
      disabledByPolicyText,
      cannotAddDueToPolicy,
      addItem,
      removeItem,
      updateItem,
      displayedItems,
      fullList,
      DOCS,
    };
  },
  methods: {
    updateModel(id: string, field: string, value: string | null | number): void {
      const eRes = this.displayedItems.find((er) => er.id === id);
      if (!eRes) return;

      const updatedERes = {
        ...eRes.origin,
        [field]: value,
      };
      this.update(id, updatedERes);
    },
    add() {
      const newERes: ExtendedResource = { resource: "", quantity: "" };
      this.addItem(newERes);
      this.$emit("update:model-value", this.fullList);
    },
    update(id: string, eRes: ExtendedResource) {
      this.updateItem(id, eRes);
      this.$emit("update:model-value", this.fullList);
    },
    remove(id: string) {
      this.removeItem(id);
      this.$emit("update:model-value", this.fullList);
    },
    isEmptyNameField(val: string): string | boolean {
      return isNotEmpty(val) || errorMessages.ENETER_RESOURCE;
    },
    isEmptyQuantityField(val: string): string | boolean {
      return isNotEmpty(val) || errorMessages.ENTER_QUANTITY;
    },
    getCopyData(): Array<ExtendedResource> {
      return deepCopy(this.extendedResources);
    },
    nameNotAdded(val: string): boolean | string {
      const matchedResNames = this.extendedResources.filter(
        (extendedRes: ExtendedResource) => extendedRes.resource === val,
      );
      return matchedResNames.length === 1 || errorMessages.RESOURCE_ALREADY_USED;
    },
    isNameValid(val: string): boolean | string {
      return isExtendedResourceName(val) || errorMessages.ENTENDED_RESOURCE_NAME_NOT_VALID;
    },
    isValidQuantity(val: string): boolean | string {
      return isValidExtendedResourceQuantity(val) || errorMessages.INVALID_QUANTITY;
    },
  },
});
</script>
<style lang="scss" scoped>
.extended-resources {
  .more-info-container {
    font-size: 14px;
  }
  .extended-resource-box {
    background-color: $body-background-color;
    margin-bottom: 15px;
  }

  .close-button {
    color: $black-54;
    font-size: 12px;
    margin-right: -5px;
  }
}
</style>
