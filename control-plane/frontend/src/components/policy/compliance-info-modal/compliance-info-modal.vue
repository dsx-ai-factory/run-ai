<template>
  <runai-base-dialog
    aid="compliance-info-modal"
    class="compliance-info-modal"
    :model-value="true"
    @close="$emit('close')"
    size="width-lg"
    no-padding
  >
    <template #header>{{ title }}</template>
    <template #body>
      <div class="compliance-info-modal-content">
        <div class="subtitle" aid="compliance-info-modal-subtitle">
          {{ subtitle }}
        </div>
        <q-separator />
        <policy-details-table
          :bordered="false"
          :compliance-data="{
            complianceInfo,
            kind: field,
            distributedReplicaType,
          }"
          :policy-data="policyData"
          :policy-table="policyGroup"
          :is-distributed="!!distributedReplicaType"
        />
      </div>
    </template>
    <template #footer>
      <q-btn label="Close" color="primary" @click="$emit('close')" aid="compute-policy-info-modal-close-button" />
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import { PolicyDetailsTable } from "../policy-details-table";
// models
import { type ComplianceInfo, AssetKind } from "@/swagger-models/assets-service-client";

import {
  EPolicyGroup,
  getDistributedPolicyViewDetails,
  getPolicyViewDetails,
  type IPolicyViewProperty,
} from "@/utils/policy.util/policy-view.util";
import type { ReplicaType } from "@/swagger-models/workloads-client";

export default defineComponent({
  components: {
    RunaiBaseDialog,
    PolicyDetailsTable,
  },
  emits: ["close"],
  props: {
    complianceInfo: {
      type: Object as PropType<ComplianceInfo>,
      required: true,
    },
    policy: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      type: Object as PropType<any>,
      required: true,
    },
    field: {
      type: String as PropType<AssetKind>,
      required: true,
    },
    title: {
      type: String as PropType<string>,
      required: true,
    },
    subtitle: {
      type: String as PropType<string>,
      required: true,
    },
    distributedReplicaType: {
      type: String as PropType<ReplicaType>,
      required: false,
    },
  },
  computed: {
    policyData(): IPolicyViewProperty[] {
      const policyDetails = this.distributedReplicaType
        ? getDistributedPolicyViewDetails(this.policy.effective || {})
        : getPolicyViewDetails(this.policy.effective);
      return Object.values(policyDetails[this.policyGroup] || {});
    },
    policyGroup(): EPolicyGroup {
      switch (this.field) {
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
.compliance-info-modal-content {
  padding: 15px;
  .subtitle {
    margin-bottom: 15px;
    font-size: 12px;
  }
}
</style>
