<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="cluster-installation-substep">
    <!-- CLUSTER CREATION ERROR -->
    <cluster-creation-error v-if="clusterData.installationErrorType === InstallationErrorType.ClusterCreation" />

    <!-- INSTALLATION INSTRUCTIONS ERROR -->
    <installation-instructions-error
      v-else-if="clusterData.installationErrorType === InstallationErrorType.InstallationInstructions"
      @close="handleClose"
    />

    <!-- SUCCESS STATE -->
    <onboarding-content-wrapper v-else>
      <template #title>Install NVIDIA Run:ai on your cluster</template>
      <template #subtitle>
        Review the
        <a
          v-if="DOCS.INSTALLATION.PREREQUISITES"
          aid="onboarding-system-requirements-link"
          :href="DOCS.INSTALLATION.PREREQUISITES"
          target="_blank"
        >
          system
        </a>
        <span v-else>system</span>
        and
        <a
          v-if="DOCS.INSTALLATION.NETWORK_REQUIREMENTS"
          aid="onboarding-network-requirements-link"
          :href="DOCS.INSTALLATION.NETWORK_REQUIREMENTS"
          target="_blank"
        >
          network
        </a>
        <span v-else>network</span>
        requirements.
        <br />

        Then run the command below in your Kubernetes cluster to
        <a :href="DOCS.INSTALLATION.HELM_INSTALL" target="_blank">install</a> NVIDIA Run:ai.
      </template>

      <template #content>
        <div v-if="clusterData.installationCommand" class="cluster-installation-command">
          <runai-copyable-code-block :code="clusterData.installationCommand" class="q-mb-xl" />

          <cluster-connection-status
            v-if="clusterData.clusterId"
            :is-cluster-connected="clusterData.isClusterConnected"
            :cluster-uuid="clusterData.clusterId"
            @update:is-cluster-connected="$emit('update:is-cluster-connected', $event)"
          />
        </div>
      </template>
    </onboarding-content-wrapper>
  </div>
</template>

<script lang="ts">
import { type PropType } from "vue";
import { defineComponent } from "vue";
import OnboardingContentWrapper from "@/components/onboarding/common/onboarding-content-wrapper";
import { DOCS } from "@/constants/docs";
import { RunaiCopyableCodeBlock } from "@/components/common/runai-copyable-code-block";
import { ClusterConnectionStatus } from "./cluster-connection-status";
import { ClusterCreationError } from "./errors-state/cluster-creation-error";
import { InstallationInstructionsError } from "./errors-state/installation-instructions-error";
import { type ClusterData, InstallationErrorType } from "@/components/onboarding/models/admin-onboarding.model";

export default defineComponent({
  name: "cluster-installation-substep",
  components: {
    RunaiCopyableCodeBlock,
    OnboardingContentWrapper,
    ClusterConnectionStatus,
    ClusterCreationError,
    InstallationInstructionsError,
  },
  emits: {
    "update:is-cluster-connected": (isConnected: boolean) => typeof isConnected === "boolean",
    close: () => true,
  },
  props: {
    clusterData: {
      type: Object as PropType<ClusterData>,
      required: true,
    },
  },
  setup(_, { emit }) {
    const handleClose = () => {
      emit("close");
    };

    return {
      DOCS,
      handleClose,
      InstallationErrorType,
    };
  },
});
</script>

<style scoped lang="scss">
.cluster-installation-command {
  max-width: 900px;
}
</style>
