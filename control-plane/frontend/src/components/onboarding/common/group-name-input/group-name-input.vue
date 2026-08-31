<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-input
    :model-value="teamData.groupName"
    @update:model-value="handleGroupNameChange"
    label="Group name"
    placeholder="Enter group name"
    aid="group-name"
    class="group-name-input"
    no-error-icon
    :maxlength="40"
    :rules="[isValidGroupName]"
    hint="As recognized by the IDP"
    counter
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { TeamData, TeamDataKey } from "@/components/onboarding/models/admin-onboarding.model";
import { errorMessages } from "@/common/error-message.constant";
import { formValidators } from "@/common/form.validators";

export default defineComponent({
  name: "group-name-input",
  emits: {
    "update:team-data": (key: TeamDataKey, value: string | boolean | string[]) => !!key && value !== undefined,
  },
  props: {
    teamData: {
      type: Object as PropType<TeamData>,
      required: true,
    },
  },
  setup(_, { emit }) {
    // Event handlers
    const handleGroupNameChange = (value: string | number | null) => {
      emit("update:team-data", "groupName", String(value || ""));
    };

    // Validation functions
    const isValidGroupName = (name: string) => {
      if (!name || name.trim().length === 0) {
        return errorMessages.REQUIRED_FIELD;
      }
      return formValidators.isValidEntityName(name) || errorMessages.VALID_FORMAT;
    };

    return {
      handleGroupNameChange,
      isValidGroupName,
    };
  },
});
</script>
