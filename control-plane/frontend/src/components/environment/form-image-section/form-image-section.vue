<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item label="Image" default-opened :section-invalid="!sectionValid">
    <template #subheader>
      <span>{{ summary }}</span>
    </template>
    <form-image-origin-section
      :model-value="imageModel.image"
      @update:model-value="updateModel('image', $event)"
      is-required
    />

    <section>
      <runai-folding-sub-section label="Image pull policy">
        <form-image-pull-policy
          :image-pull-policy="imageModel.imagePullPolicy"
          @update="updateModel('imagePullPolicy', $event)"
        />
      </runai-folding-sub-section>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiFoldingSubSection } from "@/components/common/runai-folding-sub-section";
import { FormImagePullPolicy } from "@/components/environment/form-image-pull-policy";
import { FormImageOriginSection } from "@/components/environment/form-image-origin-section";
// Types
import type { IImage } from "@/models/environment.model";
// Services
import { environmentService } from "@/services/control-plane/environment.service/environment.service";

export default defineComponent({
  name: "form-image-section",
  components: {
    RunaiExpansionItem,
    FormImagePullPolicy,
    RunaiFoldingSubSection,
    FormImageOriginSection,
  },
  emits: ["update"],
  props: {
    imageModel: {
      type: Object as PropType<IImage>,
      required: true,
    },
  },
  computed: {
    summary(): string {
      return this.imageModel.image || "None";
    },
    sectionValid(): boolean {
      return environmentService.validateImage(this.imageModel.image || "");
    },
  },
  methods: {
    updateModel(key: string, val: string | number | null) {
      this.$emit("update", { ...this.imageModel, [key]: val });
    },
  },
});
</script>
