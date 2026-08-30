<template>
  <div class="nim-model-and-access">
    <section class="model-name-section q-mb-lg">
      <nim-model-selector
        :model-name="modelValue.modelName"
        @nim-model-name-changed="onModelNameChanged"
        @is-valid="$emit('is-name-valid', $event)"
      />
    </section>

    <template v-if="modelValue.modelName">
      <section class="model-profile-section">
        <nim-profile-selector
          :profile-name-or-hash="modelValue.profileNameOrHash"
          @nim-profile-changed="onProfileChanged"
          @is-valid="$emit('is-profile-valid', $event)"
          :is-profile-required="isProfileRequired"
        />
      </section>

      <section class="access-model-section q-mt-lg">
        <nim-access-model
          :model-value="modelValue.accessModel"
          :cluster-id="clusterId"
          @update:model-value="onAccessModelChanged"
          :is-access-required="isAccessRequired"
          :scope-model="scopeModel"
          :show-create-new-personal-credentials-btn="showCreateNewPersonalCredentialsBtn"
          :show-personal-credentials-option="showPersonalCredentialsOption"
          :supporting-ngc-api-key="supportingNgcApiKey"
          :supports-docker-registry="supportsDockerRegistry"
        />
      </section>
    </template>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { NimModelSelector } from "../nim-model-selector";
import { NimProfileSelector } from "../nim-profile-selector";
import { NimAccessModel } from "../nim-access-model";
import type { INimModelAndAccessModel } from "./nim-model-and-access.model";
import type { INimAccessModelModel } from "../nim-access-model/nim-access-model.model";
import type { IScopeModel } from "@/models/global.model";

export default defineComponent({
  name: "nim-model-and-access",
  components: {
    NimModelSelector,
    NimProfileSelector,
    NimAccessModel,
  },
  emits: {
    "update:model-value": (value: INimModelAndAccessModel) => value,
    "is-name-valid": (value: boolean) => typeof value === "boolean",
    "is-profile-valid": (value: boolean) => typeof value === "boolean",
  },
  props: {
    modelValue: {
      type: Object as PropType<INimModelAndAccessModel>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
    scopeModel: {
      type: Object as PropType<IScopeModel>,
      required: true,
    },
    isProfileRequired: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    isAccessRequired: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showCreateNewPersonalCredentialsBtn: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    showPersonalCredentialsOption: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    supportingNgcApiKey: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    supportsDockerRegistry: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  methods: {
    onModelNameChanged(modelName: string): void {
      const previousModelName = this.modelValue.modelName ?? "";
      let profileNameOrHash = "";
      if (modelName) {
        profileNameOrHash = modelName === previousModelName ? this.modelValue.profileNameOrHash : "";
      }
      this.$emit("update:model-value", {
        modelName,
        profileNameOrHash,
        accessModel: this.modelValue.accessModel,
      });
    },
    onProfileChanged(profileNameOrHash: string): void {
      this.$emit("update:model-value", {
        modelName: this.modelValue.modelName,
        profileNameOrHash,
        accessModel: this.modelValue.accessModel,
      });
    },
    onAccessModelChanged(accessModel: INimAccessModelModel): void {
      this.$emit("update:model-value", {
        modelName: this.modelValue.modelName,
        profileNameOrHash: this.modelValue.profileNameOrHash,
        accessModel,
      });
    },
  },
});
</script>
