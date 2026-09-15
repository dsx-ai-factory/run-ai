<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="installation-section"
    label="Installation instructions"
    default-opened
    :section-invalid="false"
  >
    <div class="q-pb-md">
      Before installing the Run:ai cluster, refer to the
      <a v-if="DOCS.INSTALLATION.PREREQUISITES" :href="DOCS.INSTALLATION.PREREQUISITES" target="_blank">system</a>
      <span v-else>system</span>
      and
      <a v-if="DOCS.INSTALLATION.NETWORK_REQUIREMENTS" :href="DOCS.INSTALLATION.NETWORK_REQUIREMENTS" target="_blank">
        network
      </a>
      <span v-else>network</span>
      requirements and follow the instructions.
    </div>
    <div class="dashed-seperator"></div>
    <div class="q-my-md">Once all the requirements are met, run the following commands:</div>

    <!-- HELM -->
    <template v-if="showHelmInstructions">
      <installation-command-box :max-height="60">
        <template #top>Download the Run:ai Helm values file by running the following command:</template>

        <template #commands>
          curl '{{ currentURL }}/v1/k8s/clusters/{{ createdCluster?.uuid }}/installfile?cloud={{ k8sDist
          }}{{ clusterURL }}' --header 'Authorization: Bearer {{ authBearer }}' > runai-{{
            selectedInstallationModel.name
          }}.yaml
        </template>

        <template #bottom>
          <div class="italic">
            Optional: To customize the cluster installation,
            <a v-if="helmLink" :href="helmLink" target="_blank"> edit the Helm values file</a>
            <span v-else> edit the Helm values file</span>
          </div>
        </template>
      </installation-command-box>
    </template>

    <!-- RUN COMMAND -->
    <installation-command-box>
      <template #top>Run the following command:</template>
      <template #bottom>
        <div class="italic">
          For more information, see the
          <a target="_blank" :href="DOCS.INSTALLATION.HELM_INSTALL">install using Helm</a>
          guide
        </div>
      </template>

      <template #commands>
        {{ installationInstructions }}
      </template>
    </installation-command-box>
  </runai-expansion-item>
</template>
<script lang="ts">
// cmps
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { InstallationCommandBox } from "./installation-command-box";
import { clusterK8SDistributionOptions } from "@/models/cluster.model";
import type { IClusterInstalltionModel } from "@/models/cluster.model";
// types
import type { PropType } from "vue";
import { useAuthStore } from "@/stores/auth.store";
import { defineComponent } from "vue";
import { clusterService } from "@/services/control-plane/cluster.service/cluster.service";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { DOCS } from "@/constants/docs";

export default defineComponent({
  name: "cluster-installation-section",
  components: {
    RunaiExpansionItem,
    InstallationCommandBox,
  },
  props: {
    createdCluster: {
      type: [Object, null] as PropType<ClustersMinimalData | null>,
      required: true,
    },
    showCertificateInstructions: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isHigherClusterVersion: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isMultiTenant: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    selectedInstallationModel: {
      type: Object as PropType<IClusterInstalltionModel>,
      required: true,
    },
    versionsMap: {
      type: Object as PropType<Record<string, string>>,
      required: true,
    },
  },
  data() {
    return {
      authStore: useAuthStore(),
      installationInstructions: "" as string | undefined,
      DOCS,
    };
  },
  async created() {
    this.installationInstructions = await this.getInstallationInstructions();
  },
  computed: {
    helmLink(): string | undefined {
      return DOCS.INSTALLATION.CUSTOM_INSTALL;
    },
    showHelmInstructions(): boolean {
      return !this.isHigherClusterVersion;
    },
    clusterURL(): string {
      return this.selectedInstallationModel.prerequisites.clusterUrl
        ? `&clusterip=${this.selectedInstallationModel.prerequisites.clusterUrl}`
        : "";
    },
    k8sDist(): string {
      return (
        clusterK8SDistributionOptions.find(
          (dist) => dist.id === this.selectedInstallationModel.prerequisites.k8sDistribution,
        )?.value || ""
      );
    },
    currentURL(): string {
      return window.location.origin;
    },
    authBearer(): string {
      return this.authStore.accessToken;
    },
    getClusterVersion(): string {
      if (!this.selectedInstallationModel.version) return "";
      return this.versionsMap[this.selectedInstallationModel.version];
    },
  },
  methods: {
    async getInstallationInstructions(): Promise<string | undefined> {
      const response = await clusterService.getClusterInstallationInfo(
        this.createdCluster?.uuid as string,
        String(this.getClusterVersion),
        this.selectedInstallationModel.prerequisites.clusterUrl,
      );
      return response ? response.data.installationStr : undefined;
    },
  },
});
</script>
<style lang="scss">
.installation-section {
  .dev-mode-span {
    color: blue;
  }
  .dashed-seperator {
    border: 1px solid $black-12;
  }
  .list {
    li {
      display: list-item;
      list-style-type: circle;
    }
  }
}
</style>
