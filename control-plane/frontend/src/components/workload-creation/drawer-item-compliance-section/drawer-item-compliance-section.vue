<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <overlay-drawer-section-wrapper @close="$emit('close')" :header-style="{ background: '#EEEEEE' }">
    <template #header-title>
      <q-icon name="ph-bold ph-clipboard-text" size="20.58px" color="black-54" class="q-mr-md" />
      <span class="text-weight-bold ellipsis" style="max-width: 40ch">Policy applied to {{ displayedItem?.name }}</span>
    </template>

    <template #main>
      <section class="policy-info">
        <div class="subtitle q-pa-md">
          {{ subtitle }}
        </div>
        <compliance-details
          style="max-width: 536px"
          :item="displayedItem"
          :policy-details="policyDetails"
          :distributed-replica-type="distributedReplicaType"
        />
      </section>
    </template>
  </overlay-drawer-section-wrapper>
</template>

<script lang="ts">
import type { IDrawerListItem, WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import { ComplianceDetails } from "../compliance-details";
import { OverlayDrawerSectionWrapper } from "@/components/workload-creation/overlay-drawer-section-wrapper";
import { defineComponent } from "vue";
import type { PropType } from "vue";
import type { ReplicaType } from "@/swagger-models/workloads-client";
import type { IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { workloadCreationDrawerEntityMapper } from "../workload-creation-drawer-content/content-mapper";
export default defineComponent({
  name: "drawer-item-compliance-section",
  components: {
    OverlayDrawerSectionWrapper,
    ComplianceDetails,
  },
  emits: ["close"],
  props: {
    displayedItem: {
      type: Object as PropType<IDrawerListItem>,
      required: true,
    },
    entity: {
      type: [String, null] as PropType<WorkloadCreationDrawerEntities | null>,
      required: true,
    },
    distributedReplicaType: {
      type: String as PropType<ReplicaType>,
      required: false,
    },
    policyDetails: {
      type: Object as PropType<IPolicyViewDetails>,
      required: false,
      default: () => ({}),
    },
  },
  data() {
    return {
      extenedSectionOpen: false,
    };
  },
  computed: {
    entityLabel(): string {
      if (!this.entity) return "";
      return workloadCreationDrawerEntityMapper[this.entity]?.entityLabel || "";
    },
    subtitle(): string {
      return this.displayedItem.disabled
        ? `The workload can't be created with this ${this.entityLabel} because the ${this.entityLabel} doesn’t comply with the following policy rules set by your administrator:`
        : `This ${this.entityLabel} doesn't comply with the following policy rules your administrator set. Selecting this ${this.entityLabel} for the workload will require modifying the noncompliant values.`;
    },
  },
});
</script>
<style lang="scss" scoped>
.policy-info {
  .subtitle {
    font-size: 12px;
    font-weight: 400;
  }
}
</style>
