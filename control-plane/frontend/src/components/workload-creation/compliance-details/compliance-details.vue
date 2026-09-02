<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <policy-details-table
    class="border-top"
    :compliance-data="complianceData"
    :policy-data="policyData"
    :policy-table="policyGroup"
    :is-distributed="!!distributedReplicaType"
    :distributed-replica-type="distributedReplicaType"
    :bordered="false"
  />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { PolicyDetailsTable } from "@/components/policy/policy-details-table";
import type { IPolicyViewDetails, IPolicyViewProperty } from "@/utils/policy.util/policy-view.util";
import { EPolicyGroup } from "@/utils/policy.util/policy-view.util";
import { AssetKind } from "@/swagger-models/assets-service-client";
import type { IDrawerListItem } from "@/models/workload-creation-drawer-model";

export default defineComponent({
  name: "compliance-details",
  components: { PolicyDetailsTable },
  props: {
    item: {
      type: Object as PropType<IDrawerListItem>,
      required: true,
    },
    policyDetails: {
      type: Object as PropType<IPolicyViewDetails>,
      required: false,
      default: () => ({}),
    },
    distributedReplicaType: {
      type: String as PropType<string>,
      required: false,
    },
  },
  computed: {
    policyData(): IPolicyViewProperty[] {
      return Object.values(this.policyDetails[this.policyGroup] || {});
    },
    complianceData() {
      if (!this.item.policy?.reason) return undefined;
      return {
        complianceInfo: this.item.policy,
        kind: this.item.kind,
        distributedReplicaType: this.distributedReplicaType,
      };
    },
    policyGroup(): EPolicyGroup {
      switch (this.item.kind) {
        case AssetKind.ConfigMap:
        case AssetKind.HostPath:
        case AssetKind.Git:
        case AssetKind.Nfs:
        case AssetKind.S3:
        case AssetKind.Pvc:
        case AssetKind.SecretVolume:
          return EPolicyGroup.Storage;
        case AssetKind.Compute:
          return EPolicyGroup.Compute;
        case AssetKind.Environment:
          return EPolicyGroup.Environment;
        default:
          return EPolicyGroup.General;
      }
    },
  },
});
</script>
<style lang="scss" scoped>
.border-top {
  border-top: 1px solid $black-12;
}
</style>
