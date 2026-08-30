<template>
  <q-input
    :model-value="teamData.teamName"
    @update:model-value="handleTeamNameChange"
    label="Team name"
    placeholder="Enter a name"
    aid="team-name"
    class="team-name-input"
    :rules="[isValidEntityName]"
    no-error-icon
    :maxlength="40"
    counter
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { TeamData, TeamDataKey } from "@/components/onboarding/models/admin-onboarding.model";
import { errorMessages } from "@/common/error-message.constant";
import { formValidators } from "@/common/form.validators";

export default defineComponent({
  name: "team-name-input",
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
    const handleTeamNameChange = (value: string | number | null) => {
      emit("update:team-data", "teamName", String(value || ""));
    };

    // Validation functions
    const isValidEntityName = (name: string) => {
      return formValidators.isValidEntityName(name) || errorMessages.VALID_FORMAT;
    };

    return {
      handleTeamNameChange,
      isValidEntityName,
    };
  },
});
</script>
