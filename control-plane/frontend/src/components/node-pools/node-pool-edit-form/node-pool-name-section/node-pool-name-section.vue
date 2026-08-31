<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item label="Node pool name" default-opened :section-invalid="sectionInvalid" :loading="loading">
    <template #subheader>
      {{ expansionSubHeader }}
    </template>
    <runai-section>
      <div class="row">
        <runai-name-validation
          class="col-7"
          aid="node-pool-name-input"
          stack-label
          :model-value="nodePoolName"
          @update:model-value="$emit('update:node-pool-name', $event)"
          :rules="[nameNotEmpty, isValidNameFormat, isNameUnique]"
          :disable="disabled"
        />
      </div>
      <q-tooltip anchor="center left" v-if="disabled">The node pool name cannot be modified</q-tooltip>
    </runai-section>
  </runai-expansion-item>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiNameValidation } from "@/components/common/runai-name-validation";
import { isFirstLetterIsLowerCase, isLowerCaseAndNumbers, isNotEmpty } from "@/common/form.validators";
import { errorMessages } from "@/common/error-message.constant";
import { ErrorAlert } from "@/utils/error-alert.util";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import { RunaiSection } from "@/components/common/runai-section";

export default defineComponent({
  name: "node-pool-name-section",
  components: { RunaiSection, RunaiNameValidation, RunaiExpansionItem },
  emits: ["update:node-pool-name", "is-section-invalid"],
  props: {
    nodePoolName: {
      type: String as PropType<string>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      isNodePoolNameUnique: true as boolean,
    };
  },
  computed: {
    expansionSubHeader(): string {
      return this.nodePoolName || "None";
    },
    sectionInvalid(): boolean {
      return !this.nodePoolName || !this.isNodePoolNameUnique;
    },
  },
  methods: {
    nameNotEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.NAME_NOT_EMPTY;
    },
    isValidNameFormat(val: string): boolean | string {
      return (isLowerCaseAndNumbers(val) && isFirstLetterIsLowerCase(val)) || errorMessages.VALID_FORMAT;
    },
    async isNameUnique(name: string): Promise<boolean | string> {
      try {
        const isNameUnique = await clusterService.isNodePoolNameUnique(this.clusterId, name);
        this.isNodePoolNameUnique = isNameUnique;
        if (!isNameUnique) {
          return errorMessages.NAME_ALREADY_EXIST;
        }
        return true;
      } catch (error: unknown) {
        const errorAlert = new ErrorAlert({
          generalMessage: ErrorAlert.failedToValidateMessage("node pool name"),
        });
        this.$q.notify(errorAlert.getNotification(error));
        console.error(error);
        return false;
      }
    },
  },
  watch: {
    sectionInvalid: {
      handler(isSectionInvalid: boolean): void {
        this.$emit("is-section-invalid", isSectionInvalid);
      },
      immediate: true,
    },
  },
});
</script>
