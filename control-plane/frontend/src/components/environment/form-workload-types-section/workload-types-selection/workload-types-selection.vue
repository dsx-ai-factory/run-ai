<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="workload-types-selection">
    Select the type of workload that can use this environment
    <runai-tooltip
      tooltip-text="The environment will be available for use when a workload of the selected type is created"
      tooltip-position="right"
      width="490px"
    />
    <q-option-group
      class="workload-group"
      :options="options"
      type="checkbox"
      :model-value="selectedWorkloadTypes"
      @update:model-value="updateWorkloadTypes"
    >
      <template v-slot:label="opt">
        <div class="row items-center" :aid="`radio-option-${opt.label}`">
          <span class="q-mr-md">{{ opt.label }}</span>
          <notification-alert-bar class="flex-1" v-if="opt.informationText">
            <div class="information-bar-container-text">
              {{ opt.informationText }}
            </div>
          </notification-alert-bar>
        </div>
        <q-tooltip v-if="opt.toolTip" anchor="center right" self="center left" :max-width="opt.toolTipWidth">
          {{ opt.toolTip }}
        </q-tooltip>
      </template>
    </q-option-group>
    <div class="col-12">
      <q-field class="hidden-field" :model-value="workloadTypes" :rules="[isWorkloadTypeSelected]" />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { EWorkloadName } from "@/models/workload.model";
import { errorMessages } from "@/common/error-message.constant";
import { NotificationAlertBar } from "@/components/common/notification-alert-bar";
import type { IWorkloadTypesOption } from "./workload-types-section.model";

export default defineComponent({
  name: "workload-types-selection",
  components: {
    RunaiTooltip,
    NotificationAlertBar,
  },
  emits: ["update-workload-types"],
  props: {
    workloadTypes: {
      type: Object as PropType<Record<Exclude<EWorkloadName, "distributed-inference">, boolean>>,
      required: true,
    },
    options: {
      type: Array as PropType<Array<IWorkloadTypesOption>>,
      required: true,
    },
  },
  computed: {
    selectedWorkloadTypes(): Array<EWorkloadName> {
      return Object.entries(this.workloadTypes).reduce((acc, [key, value]) => {
        if (value) acc.push(key as EWorkloadName);
        return acc;
      }, [] as Array<EWorkloadName>);
    },
  },
  methods: {
    updateWorkloadTypes(selectedWorkloadTypes: Array<Exclude<EWorkloadName, "distributed-inference">>): void {
      const workloadTypes = selectedWorkloadTypes.reduce(
        (acc, workloadType) => {
          acc[workloadType] = true;
          return acc;
        },
        {} as Record<Exclude<EWorkloadName, "distributed-inference">, boolean>,
      );
      this.$emit("update-workload-types", workloadTypes);
    },
    isWorkloadTypeSelected(val: Record<EWorkloadName, boolean>): boolean | string {
      const workloadTypeKeys: Array<EWorkloadName> = [
        EWorkloadName.Workspace,
        EWorkloadName.Training,
        EWorkloadName.Inference,
      ];
      return workloadTypeKeys.some((key) => val[key]) ? true : errorMessages.SELECT_A_TYPE;
    },
  },
});
</script>

<style lang="scss" scoped>
:deep(.notification-alert-bar) {
  min-height: auto;
}
</style>
