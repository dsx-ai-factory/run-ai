<template>
  <!-- TODO: Need to implement the policy here -->
  <policy-non-compliance-card-wrapper
    :compliance-data="undefined"
    :field="data.meta.kind"
    :show-disabled-info="undefined"
    :asset-name="data.meta.name"
  >
    <runai-select-card
      class="credential-card"
      :value="isSelected"
      :disabled="disabled"
      @clicked="$emit('card-clicked')"
      :aid="getAid(data.meta.name)"
    >
      <q-tooltip v-if="data.meta.description" max-width="300px" anchor="top middle" self="bottom middle">{{
        data.meta.description
      }}</q-tooltip>
      <div class="workload-card-content">
        <section class="top-card-section">
          <div class="credential-logo">
            <q-icon name="ph-bold ph-key" size="42px" color="grey-8" />
          </div>
        </section>
        <div class="credential-card-title q-mb-md" :class="{ 'text-primary': isSelected }">
          <div class="col text-subtitle1 text-weight-medium ellipsis">
            {{ data.meta.name }}
          </div>
        </div>
        <div class="credential-card-details">
          <span class="text-weight-medium q-mr-xs">Type:</span>
          {{ data.meta.kind }}
        </div>
      </div>
    </runai-select-card>
  </policy-non-compliance-card-wrapper>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmp
import { RunaiSelectCard } from "@/components/common/runai-select-card";

// models
import type { CredentialsListResponseEntry } from "@/swagger-models/assets-service-client";

// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";
import { policyNonComplianceCardWrapper } from "../../policy-non-compliance-card-wrapper";

export default defineComponent({
  name: "credential-card",
  components: {
    RunaiSelectCard,
    policyNonComplianceCardWrapper,
  },
  emits: ["card-clicked"],
  props: {
    data: {
      type: Object as PropType<CredentialsListResponseEntry>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showDisabledInfo: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  methods: {
    getAid(name: string): string {
      return `${toLowerCase(spaceToDash(name))}-card`;
    },
  },
});
</script>

<style lang="scss" scoped>
.credential-card {
  .top-card-section {
    display: flex;
    justify-content: space-between;
  }

  .credential-card-details {
    line-break: anywhere;
  }
}
</style>
