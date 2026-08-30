<template>
  <runai-expansion-item
    label="Volume"
    :subheader="summary"
    :section-invalid="sectionInvalid"
    :should-validate="shouldAutoValidate"
    aid="volume-section"
  >
    <q-linear-progress animation-speed="100" indeterminate v-if="loading" />

    <template v-else>
      <div class="row items-center q-mb-md">
        <span class="q-mr-md">Set the volume needed for your workload</span>
        <runai-tooltip
          tooltip-text="A volume allocates storage space to your workload that is persistent across restarts."
        />
      </div>
      <volume-box
        v-for="volume in volumes"
        :key="volume.claimName"
        :volume="volume"
        :storage-classes="storageClasses"
        @update-volume="updateVolume"
        @remove-volume="removeVolume"
        :disable="sectionDisabled"
        :policy-rules="policyRules?.attributes"
        :show-volume-persistency-section="showVolumePersistencySection"
        :loading-storage-class="loadingStorageClass"
      />
      <runai-btn
        :disable="disableAddButton"
        v-if="!volumes.length"
        aid="add-volume-btn"
        icon="ph-bold ph-plus"
        label="volume"
        class="btn-icon-sm"
        flat
        color="primary"
        @click="addVolume"
      >
        <q-tooltip v-if="cannotAddDueToPolicy">{{ disableAddDueToPolicyTooltip }}</q-tooltip>
      </runai-btn>
    </template>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { RunaiTooltip } from "@/components/common/runai-tooltip";
import { RunaiBtn } from "@/components/common/runai-btn";
import { VolumeBox } from "./volume-box";
// Models
import type { IUIVolume } from "@/models/data-source.model";
import type { Pvc, PvcAccessModes } from "@/swagger-models/assets-service-client";
import type { StorageClassV2 } from "@/swagger-models/k8s-objects-tracker-client";
// Utils
import { dataSourceUtil } from "@/utils/data-source.util";
import type { PvcsRules } from "@/swagger-models/policy-service-client";
import { errorMessages } from "@/common/error-message.constant";

export default defineComponent({
  name: "volume-section",
  components: { RunaiExpansionItem, RunaiTooltip, VolumeBox, RunaiBtn },
  emits: ["add-volume", "remove-volume", "update-volumes"],
  props: {
    volumes: {
      type: Array as PropType<Array<IUIVolume>>,
      required: true,
    },
    storageClasses: {
      type: Array as PropType<Array<StorageClassV2>>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    sectionDisabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    policyRules: {
      type: [Object, null] as PropType<PvcsRules | null>,
      required: false,
    },
    policyDefaults: {
      type: [Object, null] as PropType<Pvc | null>,
      required: false,
    },
    shouldAutoValidate: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showVolumePersistencySection: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    loadingStorageClass: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  computed: {
    disableAddButton(): boolean {
      return this.sectionDisabled || this.cannotAddDueToPolicy;
    },
    cannotAddDueToPolicy(): boolean {
      return this.policyRules?.instances?.canAdd === false;
    },
    disableAddDueToPolicyTooltip(): string {
      return errorMessages.CANT_BE_MODIFIED_POLICY;
    },
    summary(): string {
      if (this.volumes[0]) {
        return this.volumes[0].ephemeral ? "Ephemeral" : "Persistent";
      } else {
        return "None";
      }
    },
    sectionInvalid(): boolean {
      if (!this.volumes.length) {
        return false;
      } else {
        return this.volumes.some((volume: IUIVolume) => !this.isVolumeValid(volume));
      }
    },
  },
  methods: {
    isVolumeValid(volume: IUIVolume): boolean {
      const { size, storageClass, accessModes } = volume.claimInfo;
      return !!size && !!storageClass && !!volume.path && this.validAccessModes(accessModes || {});
    },
    validAccessModes(accessModes: PvcAccessModes): boolean {
      return Object.values(accessModes).some((mode: boolean) => mode);
    },
    addVolume(): void {
      this.$emit("update-volumes", [...this.volumes, dataSourceUtil.getEmptyVolume(this.policyDefaults)]);
    },
    updateVolume(volume: IUIVolume): void {
      this.$emit(
        "update-volumes",
        this.volumes.map((v: IUIVolume) => (v.claimName === volume.claimName ? volume : v)),
      );
    },
    removeVolume(claimName: string): void {
      this.$emit(
        "update-volumes",
        this.volumes.filter((v: IUIVolume) => v.claimName !== claimName),
      );
    },
  },
});
</script>
