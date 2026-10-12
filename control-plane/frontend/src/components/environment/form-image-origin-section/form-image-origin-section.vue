<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="form-image-origin-section">
    <runai-radio-options
      v-if="!policyRules && showNgcIntegration"
      :options="imageOptions"
      v-model="imageOriginValue"
      title="Set the environment image"
      :tooltip-text="tooltipText"
      :disable="disable"
      class="q-mb-md"
    />
    <runai-section :title="imageTitle" v-if="imageOriginValue === EImageOrigin.Custom">
      <form-image-input
        :image="modelValue"
        @update="$emit('update:model-value', $event)"
        :is-required="isRequired"
        :policy-rules="policyRules"
        :tooltip-text="tooltipText"
        :disable="disable"
      />
    </runai-section>
    <form-ngc-catalog-image-section
      v-else-if="imageOriginValue === EImageOrigin.NGC"
      :support-authenticated-ngc="supportAuthenticatedNgc"
      :authenticated-only="authenticatedOnly"
      :scope-id="scopeId"
      :scope-type="scopeType"
      :cluster-id="clusterId"
      :ngc-catalog-registry-id="ngcCatalogRegistryId || ''"
      @update:model-value="$emit('update:model-value', $event)"
      @credential-changed="$emit('credential-changed', $event)"
    />
    <form-ngc-private-registry-section
      v-else-if="imageOriginValue === EImageOrigin.NGC_PRIVATE_REGISTRY"
      :registries="registries"
      :scope-id="scopeId"
      :scope-type="scopeType"
      :cluster-id="clusterId"
      @update:model-value="$emit('update:model-value', $event)"
      @credential-changed="$emit('credential-changed', $event)"
    />
  </section>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { FormImageInput } from "@/components/environment/form-image-input";
import { FormNgcCatalogImageSection } from "@/components/environment/form-ngc-catalog-image-section";
import { FormNgcPrivateRegistrySection } from "@/components/environment/form-ngc-private-registry-section";
import type { ISelectOption } from "@/models/global.model";
import type { StringRules } from "@/swagger-models/assets-service-client";
import { useSettingStore } from "@/stores/setting.store";
import { ScopeType, type RegistryIntegration } from "@/swagger-models/catalog-service-client";
import { catalogService } from "@/services/control-plane/catalog.service/catalog.service";
import { RunaiSection } from "@/components/common/runai-section";

export enum EImageOrigin {
  Custom = "custom",
  NGC = "ngc",
  NGC_PRIVATE_REGISTRY = "ngc_private_registry",
}

export default defineComponent({
  name: "form-image-origin-section",
  components: {
    RunaiRadioOptions,
    FormImageInput,
    FormNgcCatalogImageSection,
    FormNgcPrivateRegistrySection,
    RunaiSection,
  },
  emits: ["update:model-value", "credential-changed"],
  props: {
    modelValue: {
      type: [String, null] as PropType<string | null>,
      required: true,
    },
    isRequired: {
      type: Boolean,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<StringRules | null>,
      required: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    disable: {
      type: Boolean,
      default: false,
    },
    supportAuthenticatedNgc: {
      type: Boolean,
      default: false,
    },
    authenticatedOnly: {
      type: Boolean,
      default: false,
    },
    scopeId: {
      type: String as PropType<string>,
      required: false,
    },
    scopeType: {
      type: String as PropType<string>,
      required: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    imageTitle: {
      type: String as PropType<string>,
      required: false,
    },
  },
  data() {
    return {
      imageOriginValue: EImageOrigin.Custom,
      EImageOrigin,
      registries: [] as RegistryIntegration[],
      ngcCatalogRegistryId: null as string | null,
    };
  },
  created() {
    this.loadRegistries();
  },
  computed: {
    showNgcIntegration(): boolean {
      return this.supportNgcCatalog || this.supportNgcPrivateRegistry;
    },
    supportNgcCatalog(): boolean {
      return useSettingStore().isNgcIntegrationEnabled;
    },
    supportNgcPrivateRegistry(): boolean {
      return this.registries.length > 0;
    },
    imageOptions(): ISelectOption[] {
      const options: ISelectOption[] = [{ label: "Custom image", value: EImageOrigin.Custom }];

      if (this.supportNgcCatalog) {
        options.push({ label: "Select from the NGC catalog", value: EImageOrigin.NGC });
      }

      if (this.supportNgcPrivateRegistry) {
        options.push({ label: "Select from the NGC private registry", value: EImageOrigin.NGC_PRIVATE_REGISTRY });
      }

      return options;
    },
  },
  methods: {
    async loadRegistries(): Promise<void> {
      try {
        const registries = await catalogService.getRegistries();
        this.registries = registries.filter((registry) => registry.scopeType !== ScopeType.System);
        this.ngcCatalogRegistryId =
          registries.find((registry) => registry.scopeType === ScopeType.System && registry.name === "NGCR")?.id || null;
      } catch (error: unknown) {
        console.error(error);
        this.$q.notify({
          message: "Failed to load registries",
        });
      }
    },
  },
});
</script>
