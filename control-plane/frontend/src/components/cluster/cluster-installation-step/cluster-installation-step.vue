<template>
  <runai-form-wrapper class="cluster-form" :form-state="{}">
    <runai-expansion-item class="name-section" label="Cluster name" :subheader="nameSummary"></runai-expansion-item>
    <runai-expansion-item class="version" label="Run:ai version" :subheader="versionSummary"></runai-expansion-item>
    <runai-expansion-item class="settings" label="Settings" :subheader="prerequisitesSummary"></runai-expansion-item>
    <cluster-installation-section
      :created-cluster="createdCluster"
      :selected-installation-model="clusterFormModel"
      :show-certificate-instructions="showCertificateInstructions"
      :is-higher-cluster-version="isHigherClusterVersion"
      :is-multi-tenant="isMultiTenant"
      :versions-map="versionsMap"
    />
  </runai-form-wrapper>
</template>
<script lang="ts">
// cmps
import { ClusterInstallationSection } from "../cluster-installation-section";
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
// type
import {
  EClusterDistIds,
  EClusterLocations,
  clusterK8SDistributionOptions,
  clusterLocationText,
} from "@/models/cluster.model";
import type { IClusterInstalltionModel } from "@/models/cluster.model";
import type { PropType } from "vue";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";

export default {
  components: {
    RunaiFormWrapper,
    ClusterInstallationSection,
    RunaiExpansionItem,
  },
  emits: ["update-name"],
  props: {
    isMultiTenant: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    createdCluster: {
      type: [Object, null] as PropType<ClustersMinimalData | null>,
      required: true,
    },
    clusterFormModel: {
      type: Object as PropType<IClusterInstalltionModel>,
      required: true,
    },
    versionsMap: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
    isHigherClusterVersion: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  computed: {
    nameSummary(): string {
      return this.clusterFormModel.name || "None";
    },
    versionSummary(): string {
      return this.clusterFormModel.version || "None";
    },
    prerequisitesSummary(): string {
      const isShowLocation: boolean =
        !(
          this.isHigherClusterVersion &&
          this.clusterFormModel.prerequisites.k8sDistribution === EClusterDistIds.OPEN_SHIFT
        ) && !this.isMultiTenant;
      let selectedDistName;
      let clusterLocation = "";
      if (this.clusterFormModel.prerequisites.k8sDistribution === "other") {
        selectedDistName = "Other";
      } else {
        selectedDistName =
          clusterK8SDistributionOptions.find((dist) => dist.id === this.clusterFormModel.prerequisites.k8sDistribution)
            ?.title || "";
      }
      const locationText = clusterLocationText(
        this.clusterFormModel.prerequisites.location === "remote" || this.isMultiTenant,
      );
      if (isShowLocation) {
        clusterLocation = ` ${locationText}`;
      }
      if (this.isHigherClusterVersion) {
        return `${locationText}`;
      } else {
        return selectedDistName
          ? `Kubernetes distribution: ${selectedDistName} ${", " + clusterLocation || ""}`
          : "None";
      }
    },
    showCertificateInstructions(): boolean {
      if (
        this.isHigherClusterVersion &&
        this.clusterFormModel.prerequisites.k8sDistribution === EClusterDistIds.OPEN_SHIFT
      )
        return false;
      return this.isMultiTenant || this.clusterFormModel.prerequisites.location === EClusterLocations.REMOTE;
    },
  },
};
</script>
<style lang="scss"></style>
