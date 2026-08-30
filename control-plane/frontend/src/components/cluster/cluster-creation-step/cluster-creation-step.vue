<template>
  <runai-expansion-wrapper>
    <cluster-name-section
      :model-value="clusterFormModel.name"
      @update:model-value="updateClusterName"
      :is-cluster-edit-mode="isClusterEditMode"
    />

    <cluster-version-section
      :model-value="clusterFormModel.version"
      @update:model-value="updateVersion"
      :allowed-versions="versions"
    />

    <cluster-prerequisites-section
      v-if="clusterFormModel.version"
      :is-higher-cluster-version="isHigherClusterVersion"
      :is-multi-tenant="isMultiTenant"
      :prerequisites="clusterFormModel.prerequisites"
      @update:model-value="updatePrerequisites"
    />
  </runai-expansion-wrapper>
</template>
<script lang="ts">
// cmps
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { ClusterNameSection } from "@/components/cluster/cluster-name-section";
import { ClusterVersionSection } from "@/components/cluster/cluster-version-section";
import { ClusterPrerequisitesSection } from "@/components/cluster/cluster-prerequisites-section";
// utils
import { isNewerVersion } from "@/utils/version.util";
// type
import {
  EClusterDistIds,
  EClusterLocations,
  type IClusterInstalltionModel,
  type IPrerequisites,
  type IK8sDistributionCardData,
  clusterK8SDistributionOptions,
  clusterNewInstallerThreshold,
} from "@/models/cluster.model";
import type { PropType } from "vue";
import { defineComponent } from "vue";

export default defineComponent({
  components: {
    ClusterNameSection,
    RunaiExpansionWrapper,
    ClusterVersionSection,
    ClusterPrerequisitesSection,
  },
  emits: ["update-form-model"],
  props: {
    isMultiTenant: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isClusterEditMode: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    clusterFormModel: {
      type: Object as PropType<IClusterInstalltionModel>,
      required: true,
    },
    versions: {
      type: Array as PropType<string[]>,
      required: true,
    },
    isHigherClusterVersion: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      k8sDistributions: clusterK8SDistributionOptions as IK8sDistributionCardData[],
      clusterForm: null as HTMLFormElement | null,
    };
  },
  methods: {
    updateClusterName(name: string): void {
      const newModel: IClusterInstalltionModel = { ...this.clusterFormModel, name: name };
      this.updateFormModel(newModel);
    },
    updateVersion(val: string): void {
      const newModel: IClusterInstalltionModel = {
        ...this.clusterFormModel,
        version: val,
        prerequisites: this.resetModelPrerequisites(val),
      };
      this.updateFormModel(newModel);
    },
    updatePrerequisites(prerequisites: IPrerequisites): void {
      const newModel: IClusterInstalltionModel = { ...this.clusterFormModel, prerequisites };
      this.updateFormModel(newModel);
    },
    resetModelPrerequisites(selectedVersion: string): IPrerequisites {
      let dist: string;
      if (selectedVersion) {
        dist = isNewerVersion(selectedVersion, clusterNewInstallerThreshold) ? EClusterDistIds.OPEN_SHIFT : "";
      } else {
        dist = "";
      }
      return {
        location: this.isMultiTenant ? EClusterLocations.REMOTE : EClusterLocations.CONTROL_PLANE,
        clusterUrl: this.clusterFormModel?.prerequisites?.clusterUrl ?? "",
        k8sDistribution: dist,
      };
    },
    updateFormModel(newModel: IClusterInstalltionModel): void {
      this.$emit("update-form-model", newModel);
    },
  },
});
</script>
<style lang="scss"></style>
