<template>
  <section class="card-compliance-wrapper">
    <div>
      <slot />
    </div>
    <q-btn
      v-if="showDisabledInfo"
      class="policy-non-compliance-button"
      :aid="`${field}-${assetName}-compliance-btn`"
      @click="showInfo"
      icon="ph-bold ph-clipboard-text"
      flat
      round
      size="12px"
      color="primary"
    />
  </section>
  <compliance-info-modal
    v-if="openInfoModal && complianceData && showDisabledInfo"
    :field="field"
    :title="`${titleEntityText} Cannot be Selected`"
    :subtitle="`This ${subtitleEntityText} doesn't comply with the following rules your administrator set as part of the policy:`"
    :policy="currentPolicy"
    :distributed-replica-type="replicaType"
    :compliance-info="complianceData"
    @close="openInfoModal = false"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmp
import { ComplianceInfoModal } from "@/components/policy/compliance-info-modal";
import { AssetKind, type ComplianceInfo } from "@/swagger-models/assets-service-client";
import type { ReplicaType } from "@/swagger-models/workloads-client";

export default defineComponent({
  components: {
    ComplianceInfoModal,
  },
  emits: ["card-clicked"],
  inject: {
    policy: {
      from: "policy",
    },
    distributedReplicaType: {
      from: "distributedReplicaType",
      default: undefined,
    },
  },
  props: {
    complianceData: {
      type: Object as PropType<ComplianceInfo>,
      required: false,
    },
    field: {
      type: String as PropType<AssetKind>,
      required: true,
    },
    showDisabledInfo: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    assetName: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      openInfoModal: false as boolean,
    };
  },
  computed: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    currentPolicy(): any {
      return this.policy;
    },
    replicaType(): ReplicaType | undefined {
      return this.distributedReplicaType as ReplicaType | undefined;
    },
    titleEntityText(): string {
      switch (this.field) {
        case AssetKind.Compute:
          return "Compute Resource";
        case AssetKind.Environment:
          return "Environment";
        case AssetKind.S3:
        case AssetKind.Pvc:
        case AssetKind.Git:
        case AssetKind.HostPath:
        case AssetKind.Nfs:
          return "Data Source";
        default:
          return "";
      }
    },
    subtitleEntityText(): string {
      switch (this.field) {
        case AssetKind.Compute:
          return "compute resource";
        case AssetKind.Environment:
          return "environment";
        case AssetKind.S3:
        case AssetKind.Pvc:
        case AssetKind.Git:
        case AssetKind.HostPath:
        case AssetKind.Nfs:
          return "data source";
        default:
          return "";
      }
    },
  },
  methods: {
    showInfo(): void {
      this.openInfoModal = true;
    },
  },
});
</script>

<style lang="scss" scoped>
.card-compliance-wrapper {
  position: relative;
  .policy-non-compliance-button {
    position: absolute;
    top: 15px;
    right: 15px;
  }
}
</style>
