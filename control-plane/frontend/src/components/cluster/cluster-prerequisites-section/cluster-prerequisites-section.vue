<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item class="settings-section" label="Settings" default-opened :section-invalid="sectionInvalid">
    <template #subheader>
      <span>{{ summaryTitle }}</span>
    </template>

    <section class="dist-selection q-mb-md" v-if="!isHigherClusterVersion">
      <runai-form-card-section
        :loading="false"
        main-message="Select the Kubernetes distribution"
        entity-name="cluster"
        :cards-list="distributionCards"
        :selected-cards-ids="selectedDistributionId"
        search-name="k8s distribution"
        @selected-card-changed="onSelectedDistribution"
        :allow-actions="false"
      />
      <div class="col-12 q-ml-sm">
        <q-field class="hidden-field" :model-value="selectedDistributionId" :rules="[isRequiredDist]"></q-field>
      </div>
    </section>
    <section class="cluster-location q-mb-lg" v-if="showLocationOptions">
      <div class="q-mb-sm">Cluster location</div>
      <runai-radio-options
        class="location-options"
        :model-value="prerequisites.location"
        :options="locationOptions"
        @update:model-value="updateClusterLocation"
        color="primary"
      />
    </section>
    <section class="k8s-url-cluster" v-if="showLocationInput">
      <div>Enter a URL for the Kubernetes cluster. It will only be accessible within the organization network.</div>
      <div class="q-my-md italic">
        For more information, see the
        <a :href="DOCS.INSTALLATION.FQDN_REQUIREMENTS" target="_blank"> Fully Qualified Domain Name (FQDN)</a>
        section
      </div>

      <q-input
        aid="cluster-url-input"
        :model-value="prerequisites.clusterUrl"
        no-error-icon
        label="Cluster URL"
        stack-label
        placeholder="Enter URL"
        :rules="[notEmpty, validURL]"
        @update:model-value="onUpdateClusterUrl"
      />
    </section>
  </runai-expansion-item>
</template>
<script lang="ts">
// types
import { TCardCmpName, type ICardListItem } from "@/components/common/runai-card-list";
import {
  clusterK8SDistributionOptions,
  type IK8sDistributionCardData,
  EClusterLocations,
  clusterLocationText,
  SAME_CLUSTER_LOCATION,
  REMOTE_CLUSTER_LOCATION,
} from "@/models/cluster.model";
import type { PropType } from "vue";
import type { IPrerequisites } from "@/models/cluster.model";
import { DOCS } from "@/constants/docs";

// cmps
import { RunaiFormCardSection } from "@/components/common";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { errorMessages } from "@/common/error-message.constant";
import { isEmpty, isNotEmpty, isValidUrl } from "@/common/form.validators";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { defineComponent } from "vue";

export interface ILocationOptions {
  label: string;
  value: EClusterLocations;
  toolTip: string;
}

export default defineComponent({
  components: {
    RunaiExpansionItem,
    RunaiFormCardSection,
    RunaiRadioOptions,
  },
  emits: ["update:model-value", "is-section-invalid"],
  props: {
    isHigherClusterVersion: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isMultiTenant: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    prerequisites: {
      type: Object as PropType<IPrerequisites>,
      required: true,
    },
  },
  data() {
    return {
      nameAlreadyExist: false as boolean,
      k8sDistributions: clusterK8SDistributionOptions as IK8sDistributionCardData[],
      locationOptions: [
        {
          label: SAME_CLUSTER_LOCATION,
          value: EClusterLocations.CONTROL_PLANE,
          toolTip:
            "The Run:ai cluster will be installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane",
        },
        {
          label: REMOTE_CLUSTER_LOCATION,
          value: EClusterLocations.REMOTE,
          toolTip:
            "The Run:ai cluster will be installed on a different Kubernetes cluster as the NVIDIA Run:ai control plane",
        },
      ] as ILocationOptions[],
      DOCS,
    };
  },
  computed: {
    sectionInvalid(): boolean {
      return (
        (!this.isHigherClusterVersion && isEmpty(this.prerequisites.k8sDistribution)) ||
        (this.isRemoteCluster && isEmpty(this.prerequisites.clusterUrl))
      );
    },
    summaryTitle(): string {
      const clusterLocation = this.getClusterLocationTitle();
      const selectedDistName = this.getClusterDistributionName();
      return this.getSummaryTitle(clusterLocation, selectedDistName);
    },
    showLocationOptions(): boolean {
      return !this.isMultiTenant;
    },
    showLocationInput(): boolean {
      return this.isMultiTenant || this.isRemoteCluster;
    },
    isRemoteCluster(): boolean {
      return this.prerequisites.location === EClusterLocations.REMOTE;
    },
    distributionCards(): Array<ICardListItem> {
      return this.k8sDistributions.map((k8sDistType) => {
        return {
          id: k8sDistType.id,
          cardName: TCardCmpName.K8S_DISTRIBUTION,
          data: k8sDistType,
        };
      });
    },
    selectedDistributionId(): string[] {
      return this.prerequisites.k8sDistribution ? [this.prerequisites.k8sDistribution] : [];
    },
  },
  methods: {
    validURL(val: string): boolean | string {
      return isValidUrl(val) || errorMessages.INVALID_URL;
    },
    notEmpty(val: string): boolean | string {
      return isNotEmpty(val) || errorMessages.URL_NOT_EMPTY;
    },
    isRequiredDist(val: string[]): boolean | string {
      return val && val.length ? true : errorMessages.SELECT_DISTRIBUTION;
    },
    updatePrerequisites(updatedPrerequisites: IPrerequisites): void {
      this.$emit("update:model-value", updatedPrerequisites);
    },
    onUpdateClusterUrl(val: string | number | null): void {
      const clusterUrl = val ? val?.toString() : "";
      this.updatePrerequisites({ ...this.prerequisites, clusterUrl });
    },
    updateClusterLocation(val: ILocationOptions["value"]): void {
      this.updatePrerequisites({ ...this.prerequisites, location: val, clusterUrl: "" });
    },
    onSelectedDistribution(selectedDistributionId: string[]): void {
      this.updatePrerequisites({ ...this.prerequisites, k8sDistribution: selectedDistributionId[0] });
    },
    getClusterDistributionName(): string {
      return this.k8sDistributions.find((dist) => dist.id === this.prerequisites.k8sDistribution)?.title || "";
    },
    getClusterLocationTitle(): string {
      let clusterLocation = "";
      if (this.showLocationOptions) {
        const location = this.locationOptions.find((dist) => dist.value === this.prerequisites.location)?.label || "";
        clusterLocation = ` Cluster location: ${location}`;
      }
      return clusterLocation;
    },
    getSummaryTitle(clusterLocation: string, selectedDistName: string): string {
      let summaryTitle = "";
      if (this.isHigherClusterVersion) {
        summaryTitle = clusterLocationText(this.isRemoteCluster);
      } else {
        summaryTitle = selectedDistName ? `Kubernetes distribution: ${selectedDistName}, ${clusterLocation}` : "None";
      }
      return summaryTitle;
    },
  },
});
</script>
<style lang="scss" scoped>
.prerequisites-section {
  .location-options,
  .other-dist-options {
    margin-left: -18px;
  }
}
</style>
